# Design

## Context

See proposal.md - Why. The engine side is settled: `callable_custom_create2`
consumes a `GDExtensionCallableCustomInfo2` whose header contract states only
`call_func` and `token` are strictly required, with
`(call_func, callable_userdata)` as automatic identity when
`hash/equal/less_than` are null. The repo already contains every supporting
mechanism:

- `GoCallback_MethodBindMethodCall` (`pkg/core/method_bind_callback.go`) —
  the varcall entry has the exact callback shape needed:
  `const GDExtensionConstVariantPtr *args` → borrowed views via
  `unsafe.Slice`, return bit-written into the uninitialized `r_return` slot,
  `GDExtensionCallError *` out-param, plus existing helpers
  `writeNilCallReturn` / `setCallErrorOK` and the `GDEXTENSION_CALL_ERROR_*`
  constants.
- `cgo.Handle` — the repo's userdata idiom (method binds, class instances).
- `classdb_callback.go/.c` — the //export + extern + `C.cgo_*` shim technique
  for installing Go callbacks into C structs
  (`pkg/core/classdb.go` struct literals).
- `GoCallback_ClassCreationInfoToString` — constructs the out-slot String with
  `GDExtensionStringPtrWithUtf8Chars` and writes `*r_is_valid` through the
  pointer (with a comment recording the historical bug of rebinding the local
  instead).
- `CallFunc_GDExtensionInterfaceCallableCustomCreate2` and the
  `VariantNewCopy`/`VariantDestroy` wrappers — already generated in `pkg/ffi`.
- A3 (`variant-copy-ownership`) — Callables already cross method boundaries
  as refcounted Variants with copy-ctor + deterministic release.

## Goals / Non-Goals

**Goals:**

- A minimal, leak-gate-clean custom-callable core that C2/C4 can consume.
- Failure semantics that can never abort the process from inside cgo.
- Behavior identical to godot-cpp's observable custom-callable semantics
  where the engine defines them (borrow args, pair identity).

**Non-Goals (design-level, beyond proposal Non-goals):**

- No changes to `pkg/ffi`, codegen, or the method-binding encoder.
- No thread-affinity enforcement: the engine may invoke closures on any
  thread; thread safety of user closure state is the user's problem.
- No `GDExtensionInterfaceCallableCustomGetUserData` use in v1 — needed only
  when Go must recognize an incoming Callable as one of ours (C2 territory).

## Decisions

### D1: Raw-Variant closure signature (v1 API)

`core.NewGoCallable(fn func(args []builtin.Variant) (builtin.Variant, error)) builtin.Callable`.
Alternative considered: reflection-typed `fn any` reusing the varcall binder's
type gates — deferred as a cheap additive wrapper; keeping C1 small proves
lifecycle/leak semantics in isolation first. User-approved.

### D2: Three callbacks only — `call`, `free`, `to_string`

`token = FFI.Library`, `object_id = 0`. Omit `is_valid` (handle alive ⇒
valid), `hash`/`equal`/`less_than` (engine falls back to
`(call_func, userdata)` pair identity — distinct handles are distinct
callables, copies share the custom pointer so they compare equal, dedup is
correct for free), and `get_argument_count` (arity unknown; the closure judges
its own). godot-cpp provides these because its userdata is a polymorphic C++
object; ours is a unique handle, so identity is already canonical.
Alternative (mirror godot-cpp fully) rejected: ~2x surface, no behavioral
gain.

### D3: Registry = `cgo.Handle`, freed deterministically in `free_func`

`NewGoCallable` registers the closure via `cgo.Handle.New` and stores the
handle value in `callable_userdata`; `GoCallback_GoCallableFree` calls
`h.Delete()` and clears the closure reference. Alternative: finalizer-based
cleanup — rejected per A3 D7 lesson (non-deterministic, leak-gate hostile).
`cgo.Handle` is internally synchronized, so cross-thread invocation is safe
without our own mutex.

### D4: Borrowed argument views, copy-if-store

Mirrors `GoCallback_MethodBindMethodCall`: arguments become borrowed
`builtin.Variant` views (`VariantViewFromConstPtr`-style) valid for the call;
no per-argument copy/destroy in the trampoline. Closures that retain an
argument copy it with the existing `NewVariantNativeCopy`. This matches
godot-cpp's `const Variant **p_args` borrow semantics and zero-copy
performance. Alternative (deep-copy all args, destroy after return) rejected:
costs a copy per call and diverges from reference semantics; misuse risk is
mitigated by documentation and an explicit copy helper, same policy the
varcall binder follows at its decode boundary.

### D5: Return value written by bit-copy into `r_return`

On success the closure's `builtin.Variant` (engine-owned construction, e.g.
method return encoders produce) is bit-copied into the uninitialized
`r_return` slot exactly as the varcall callback does; ownership transfers to
the engine. Defensive rule: a zero-value (unconstructed) Variant from the
closure is replaced by `writeNilCallReturn`.

### D6: Failure policy — recover, log, `GDEXTENSION_CALL_ERROR_INVALID_METHOD`

Closure error or recovered panic → zap log with details,
`writeNilCallReturn(r_return)`, `r_error.error =
GDEXTENSION_CALL_ERROR_INVALID_METHOD` with `argument/expected/passed = -1`.
No enum member means "user callback failed"; INVALID_METHOD is what makes
GDScript raise a clean "Invalid call" script error. Panics must never unwind
through cgo (process abort). Alternative (return nil with CALL_OK) rejected:
silent failure.

### D7: File layout follows the established pair convention

`pkg/core/go_callable.go` (API + registry + exported callbacks),
`pkg/core/go_callable.c/.h` (extern declarations + `cgo_callcustom_*` shim
functions installed into the info struct), unit + demo tests in
`pkg/core/*_test.go` and `test/pkg` + `test/demo`.

### D8: Test fixtures expose creation and registry count

Bound methods on the example class: one returning a counter closure (state in
a captured `int64` — main-thread only in fixtures), one returning an always-
erroring callable, one returning a panicking callable, one returning a
callable whose body re-invokes a bound method (reentrancy), plus
`GoCallableRegistryCount() int64` for the release assertions
(`TestRefBoundaryRelease`-last pattern before the shutdown marker).

## Risks / Trade-offs

- [Borrowed views stored past the call = dangling] → documented contract +
  `NewVariantNativeCopy`; copy-if-store scenario in the spec pins it.
- [Engine calls a Callable after our registry entry is gone] → impossible by
  construction: `free_func` runs exactly when the engine destroys the last
  reference, after which no invocation can arrive.
- [Closure panic mid-mutation leaves user state inconsistent] → inherent to
  recover policy; we log full stacktrace, engine stays up; per-closure state
  design is user responsibility.
- [`str(cb)` label colliding with engine expectations] → `to_string_func`
  follows the proven class `to_string` pattern including writing
  `*r_is_valid = true` through the pointer.
- [Deferred calls after registry free] → engine only defers live references;
  same argument as above (reference owns lifetime).
- [Debug `.so` link time] → pure cgo addition, no new interface functions;
  build cost unchanged beyond one `.c` file.

## Migration Plan

Additive only; no data or API migration. Rollback = revert the commit;
nothing else references the new symbols.

## Open Questions

- Exact label format of `str(cb)` (e.g. `GoCallable:<id>`) — cosmetic,
  decided during implementation.
