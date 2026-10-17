# Design

## Context

See proposal.md - Why, for motivation.

Current binding machinery (all paths verified in source):

- `classDBBindMethod` (`pkg/core/classdb.go`) resolves the target via
  `reflect.Type.MethodByName(goMethodName)`, producing a `reflect.Method`
  whose `Func` carries the receiver as `In(0)`. It validates that the
  receiver type matches the class, builds `GoMethodMetadata`, then
  registers through `classdb_register_extension_class_method` keyed by the
  class `StringName` — the registration call itself never uses an
  instance.
- `NewGDExtensionClassMethodInfoFromMethodBind` copies
  `md.MethodFlags` straight into `method_info.flags`
  (`pkg/core/method_bind.go:636`), so `METHOD_FLAG_STATIC` (value 32,
  present in `pkg/ffi/ffi_wrapper.gen.go`) propagates with no extra work.
- Both engine→Go callbacks in `pkg/core/method_bind_callback.go`
  (`GoCallback_MethodBindMethodCall` varcall at ~line 44,
  `GoCallback_MethodBindMethodPtrcall` at ~line 226) resolve the receiver
  with `ObjectClassFromGDExtensionClassInstancePtr(instPtr)` and
  `log.Panic` when the result is nil. The engine passes a NULL instance
  for static invocations, so each needs a branch.
- Dispatch (`CallWithError`, `Ptrcall`, variadic `CallSlice`) injects
  `reflect.ValueOf(inst)` as the receiver before invoking `md.Func`.
- `isPtrcallBorrowEcho(ret, args)` takes only return and arguments — no
  receiver — so it is receiver-agnostic already.
- Reference implementations: godot-cpp `bind_static_method`
  (`include/godot_cpp/core/class_db.hpp`) registers `MethodBind` objects
  with `GDEXTENSION_METHOD_FLAG_STATIC` whose call functors ignore
  `p_instance`; this proves the engine honors statics through the same
  GDExtension registration path used here.
- ABI fact for scope: `godot_headers/godot/gdextension_interface.h`
  contains no interface for registering static *variables* on a class;
  GDScript `static var` lives in the script language (an experimental
  LanguageExtension ABI that is not vendored).

## Goals / Non-Goals

**Goals:**

- A static bind is registered, introspectable (`ClassDB`), and callable
  from GDScript's class path with the same marshalling quality as instance
  binds (defaults, object args/returns, caller-error reporting).
- NULL-instance handling is precise: valid iff the bind is static; every
  existing fatal stays fatal.
- Zero behavioral change to instance-method binds — static support is
  additive at every seam.

**Non-Goals (design-level):**

- No engine-side call-legality emulation (e.g. blocking static calls made
  on an instance) — whatever the engine's `MethodBind` allows is allowed.
- No call-side generation changes for engine-class statics
  (`cmd/generate/gdclassimpl/classes.go.tmpl` "TODO: refactor for static
  instantiation" stays).
- No synthetic "class object" receiver value for Go statics — the bound
  function simply has no receiver.

## Decisions

### D1: API shape — instance as class-name carrier, function as `any`

```go
func ClassDBBindStaticMethod[T GDClass](
    inst T, gdMethodName string, fn any,
    argNames []string, defaultValues []Variant,
)
```

`inst` is used only for `GetClassName()`; the callable is passed directly
as `fn any` and reflected with `reflect.ValueOf(fn)`, replacing the
method-name lookup. Every sibling bind (`ClassDBBindMethod`,
`ClassDBBindMethodVirtual`, `ClassDBBindMethodVarargs`) takes `inst` first
and reads identically inside `RegisterClass` init blocks.

Alternatives considered:

- Bare `className string` first parameter — more honest about needing no
  instance, but breaks the family's uniform shape and allows typo'd class
  names that reflection would have caught.
- A marker-method convention on `T` (e.g. an ignored-receiver method) —
  rejected: dispatch would carry the registration-time instance as
  receiver, which is semantically wrong for a static.

The Go-internal name for logs comes from
`runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()` — no
`goMethodName` parameter.

### D2: Static mode lives in `GoMethodMetadata`; dispatch skips receiver injection

`GoMethodMetadata` gains a static notion derived from its existing
`MethodFlags` (a `IsStatic()` accessor testing `MethodFlags &
METHOD_FLAG_STATIC`, not a new duplicated bool). A sibling metadata
constructor takes a plain `reflect.Value` of a package `func` with an
argument-index base of 0 instead of the receiver-skipping base of 1; all
downstream derivation (arg types/metadata, defaults, return property
info) is index-shifted, not duplicated. `CallWithError`, `Ptrcall`, and
`CallSlice` branch on `IsStatic()` to call `md.Func` without prepending
`reflect.ValueOf(inst)`.

Alternative: synthesize a bound method value with a dummy receiver —
rejected; fighting reflection to fake a receiver adds a value that must
never be observed.

### D3: Callback branches keyed on the bind's static flag, not on instance nullness

In both callbacks, the branch condition is `bind.IsStatic()`:

- static → skip instance resolution and the `inst.GetClass()` debug log
  (log `bind.ClassName` instead); proceed with a nil `GDClass`.
- non-static → the existing nil → `log.Panic` stays exactly as-is.

Branching on the *flag* (not "if instance is nil, it's static") preserves
the null-instance crash oracle for instance-binding bugs. The
varcall's `pinner.Pin(instPtr)` moves inside the non-static branch to
avoid pinning a nil pointer.

### D4: Bind-time validation is fatal and conservative

Following the existing `NewGoCallable`/bind-panic precedent: non-callable
`fn` → `log.Panic` naming class + method; duplicate Godot method name on
the class → existing `MethodMap` guard fires unchanged (shared registry,
so instance/virtual/static shares one namespace). The static+virtual flag
combination is unreachable through the public API (separate entry points)
and is additionally rejected in the shared internal path so future
callers cannot create it.

### D5: Scope split recorded in docs

Roadmap B2 splits into B2a (this change) and B2b (static variables —
blocked on a nonexistent ABI). `docs/overview.md` gains the static-methods
usage section and its static-variables section is corrected from "NOT YET
IMPLEMENTED" to blocked-on-ABI with the reason; `docs/roadmap.md` marks
B2 methods done, B2 variables blocked.

## Risks / Trade-offs

- [Engine invokes a *static* bind with a non-null instance, or vice versa]
  → The flag-keyed branch ignores a stray instance for statics and keeps
  resolution for instance binds; engine functors are proven
  instance-agnostic by godot-cpp parity. Harness pins both call paths.
- [Receiver-index shift bug in metadata derivation silently misreads
  signatures] → Constructor validates `NumIn()` against
  `len(argNames)` (defaults included) exactly as the instance path does;
  harness covers arity mismatch error reporting on a static.
- [Variadic statics interact with `CallSlice`] → The `CallSlice` branch
  skips the receiver identically; harness includes one variadic static so
  the path is live, not just code-reviewed.
- [GDScript legality surprises (e.g. static called via an instance)]
  → Engine-side behavior; harness records what the engine actually does
  rather than asserting an assumption, so the spec stays true if Godot
  changes its mind.
- [Object-DB leak gate] → All `Variant` temporaries created in the new
  static paths follow the existing `Destroy()` discipline; the suite's
  leak gate stays a hard failure.
