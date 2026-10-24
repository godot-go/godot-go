# Design

## Context

See `proposal.md` — Why.

- The generated object-return branch is one line in
  `cmd/generate/gdclassimpl/classes.go.tmpl`: `var ret <Declared>Impl`, the
  ptrcall writes the object pointer into it, and the binding returns `&ret`
  (202 sites in `pkg/gdclassimpl/classes.gen.go`). Those 202 sites cover 72
  distinct declared return types and live only in that file; no other generated
  package emits this branch, so one unexported helper in `pkg/gdclassimpl`
  serves all of them.
- `builtin.ObjectFromInstanceBinding` already returns the wrapper bound for the
  runtime class: existing binding first, otherwise the callbacks registered for
  the object's runtime class name. The generated engine-class callbacks build
  that wrapper with `GDNativeConstructors` (`pkg/gdclassinit`).
- Ref-counted returns take a different branch,
  `NewRef<Declared>GDExtensionReturnOwner(&ret)`, which owns the reference the
  return slot transfers (`object-return-ownership`).
- Narrowing was recently pinned in the opposite direction: commit `be1fa8c`
  documents that a generated return is typed to its declared class and its
  witness `Example.TestObjectCasting` asserts `node.(Label)` is false. That
  witness and the doc section state the exact behaviour this change reverses.

## Goals / Non-Goals

**Goals**

- Make `generatedReturn.(RuntimeClass)` succeed for non-refcounted returns
  without a caller-side helper, keeping the declared static return type.
- Keep every current return path working: a return that cannot be resolved still
  yields a usable declared-class wrapper.
- Keep the machinery invisible: the helper is unexported, so no cast-shaped
  entry point reaches the library's API surface.

**Non-Goals**

- No cast entry point. Widening stays a plain Go assertion.
- No change to refcounted returns, argument decode, or binding creation.
- No per-class special cases in generated code: one helper, one template line.

## Decisions

### D1 — Resolve in the generated return branch, not in `ObjectFromInstanceBinding`

`ObjectFromInstanceBinding` cannot fix this: it is already the function that
returns the runtime-class wrapper, and the declared-class wrapper never passes
through it. The only place that holds the return-slot object pointer is the
generated method, so the template emits the resolution.

Alternatives considered:

- **`CastTo[T]` helper in the library.** Leaves a helper call at every call
  site; plain assertions keep failing. Rejected as the smaller of the two
  outcomes.
- **Change generated return types to `Object`.** Loses the declared type for
  every caller. Rejected.
- **Resolve inside the argument-decode helpers.** They already do it; returns
  simply never go through them.

### D2 — `resolveGoStructObjectReturn[T Object](declared T) T`, unexported

The helper lives in a hand-written `pkg/gdclassimpl/object_return.go`, a sibling
of the generated files: generated code can call it, the package's exported API
does not gain a symbol, and the generator only creates its own outputs so the
file survives `make generate`.

The template emits `return resolveGoStructObjectReturn[Node](&ret)`. The type
argument is explicit because inference binds `T` to `*NodeImpl` from `&ret`, and
a `*LabelImpl` never asserts to `*NodeImpl`; under inference the helper would
silently fall back on every call and the change would be a no-op.

The helper:

1. returns the zero value for a typed-nil declared wrapper, using the reflect
   check from `ObjectArgPtr` rather than `declared == nil` (an interface holding
   a nil pointer is non-nil);
2. calls `ObjectFromInstanceBinding(declared.GetGodotObjectOwner())`;
3. returns the bound wrapper when it satisfies `T`;
4. falls back to `declared` when resolution fails or the bound wrapper does not
   satisfy `T`.

Requiring the assertion to `T` inside the helper is what keeps the fallback
safe: a bound wrapper that is not an instance of the declared class (a binding
whose class model disagrees with the engine's) never replaces a working
declared-class wrapper.

The engine-free half takes the resolved wrapper and error
(`resolveObjectReturnFrom(declared T, bound Object, err error) T`), so the four
branches unit-test with fake wrappers the way `objectFromBindingPtr` does.

### D3 — Only the non-refcounted branch resolves

The `IsRefcountedClassName` branch keeps `NewRef<X>GDExtensionReturnOwner`.
Returning the binding wrapper there would drop the reference the return slot
transferred: the wrapper is retained by the binding, which keeps nothing alive
on its own, so the object could be freed while Go holds it, or leak if the
transferred reference is never released. The two branches stay disjoint.

### D4 — The witness and the docs flip with the behaviour

`Example.TestObjectCasting` currently asserts the first clause
`node.(Label)` is false. That clause becomes true; the clause after the
explicit `ObjectFromInstanceBinding` stays as coverage for wrappers that did not
come from a generated return. `docs/overview.md`'s "Narrowing to a subclass"
section is rewritten to say narrowing a generated return is the assertion
alone, and to keep re-resolution only for wrappers whose binding was not
resolved.

### D5 — A null return keeps today's shape

The engine can write a null object pointer into the return slot. The resolution
then fails with `engine object is null`, and the fallback returns the wrapper
holding the null owner — exactly what the method returns today. No extra
template branch, and no binding lookup happens past the null check inside
`ObjectFromInstanceBinding`.

### D6 — The resolution stays off the public API

No exported symbol, no cast entry point, nothing for a caller to know about. The
helper is unexported in `pkg/gdclassimpl`, the package that holds all 202 call
sites, so application code sees only the generated methods.

Alternatives considered:

- **Generic methods on the generated methods** (`GetChild[T Node](...) T`). Go
  1.27 allows type parameters on methods, but interfaces cannot declare them, so
  `*NodeImpl` would stop satisfying `Node` — the type every generated method
  returns and every caller holds. Methods also cannot be overloaded by type
  parameter, so a second name would be needed that interface holders cannot
  call. Rejected.
- **A caller-supplied type argument** (`GetChild[Label](...)`). Go does not infer
  `T` from the result type, so every call site would have to spell it. More
  importantly, a caller-supplied `T` is an assumption: when the engine returns a
  different class there is no fallback value of type `T`, and the signature has
  no error channel. Rejected.
- **Generic free functions** (`NodeGetChild[T Node](node Node, ...) T`). These
  work with interfaces but add 202 exported functions with class-prefixed names.
  Rejected.

## Risks / Trade-offs

- **Extra engine call per object return.** Even a cached binding costs one
  `object_get_instance_binding` token scan → acceptable: it is the same lookup
  the argument-decode path already performs, and it replaces a caller-side
  lookup that was strictly more work.
- **Bindings now anchor returned wrappers for the object's lifetime.** An object
  that had no binding gets a `cgo.Handle` when first returned → retention is
  bounded by live engine objects, and Ref returns (the ones with reference
  churn) are excluded.
- **Identity changes: two returns of one object become `==`.** Deliberate and
  consistent with `instance-binding-lookup`; callers that relied on fresh
  wrappers per call would see a difference, so the suite runs before commit.
- **The engine-class create callback panics for an unknown class name**
  (`GoCallback_GDExtensionBindingCreate`), and the fallback cannot catch a
  panic raised inside an engine callback → class names come from the engine and
  every generated class has a constructor; the same path already runs for
  arguments today.
- **Test output gate.** Binding creation logs at debug level only, and C stack
  tracing stays behind `enablePrintStacktrace`; the gate runs at `LOG_LEVEL=WARN`.
- **Custom-class instances whose binding cannot be created** still fall back, so
  the documented `GDClassBindingCreate` workaround keeps working.

## Migration Plan

1. Add `resolveGoStructObjectReturn` in `pkg/gdclassimpl/object_return.go` and
   its unit test; `go test ./pkg/gdclassimpl/`.
2. Edit the template return branch, run `GODOT=... make generate`, review the
   `classes.gen.go` diff for only the object-return lines, and confirm
   `object_return.go` is still in place.
3. Update `Example.TestObjectCasting` and the demo assertion, then run
   `GODOT=... make test` (runtime suite plus the editor-plugin gate).
4. Update `docs/overview.md` (narrowing section, capability map row).
5. Validate the change: `openspec validate resolve-object-returns-through-binding --strict`.

Rollback is a revert of the template and helper plus a regeneration; there is no
engine ABI change and no persisted state.

## Open Questions

- Whether a hot loop that returns objects needs a cheap per-object memo above
  the binding lookup. Deferrable: the lookup is the existing one, and a
  regression would be measured against a real workload before adding a cache.
