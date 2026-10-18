# Proposal

## Why

A Go closure or method cannot be handed to GDScript as a `Callable`: the
`callable_custom_*` FFI wrappers are generated but unused, so connecting Go
behavior to signals requires exposing a bound class method and connecting
from script (roadmap C1). This is the keystone dependency for C2 (await /
signal bridge) and the `docs/overview.md`-promised goroutine patterns, and
its last prerequisite — loading `callable_custom_get_userdata` and the other
name-mismatched interface functions — landed with `fix-ffi-loader-names`.

## What Changes

- New public API in `pkg/core`: `NewGoCallable(fn func(args []builtin.Variant) (builtin.Variant, error)) builtin.Callable`
  creates an engine custom `Callable` backed by a Go closure. Raw-Variant
  signature only; typed reflection sugar is deferred (see Non-goals).
- Custom-callable plumbing in `pkg/core` (Go + C shim, following the
  `classdb_callback.go/.c` precedent): `callable_custom_create2` with three
  callbacks — `call`, `free`, `to_string` — plus `token = FFI.Library`,
  `object_id = 0`. `is_valid`/`hash`/`equal`/`less_than`/`get_argument_count`
  are omitted; the engine uses the `(call_func, callable_userdata)` pair as
  identity, giving correct dedup and copy-equality for free.
- Closure-registry keyed by `cgo.Handle` (the repo's existing userdata idiom);
  `free_func` deletes the handle deterministically when the engine drops the
  last reference, keeping the ObjectDB/leak gate clean.
- Call semantics: arguments arrive as borrowed `Variant` views valid for the
  call duration (godot-cpp parity, mirroring `GoCallback_MethodBindMethodCall`);
  storing an argument requires an explicit copy. Closure errors and panics are
  recovered, logged, reported through `GDExtensionCallError`, and never unwind
  through cgo.
- Test fixtures: `test/pkg` exposes a bound method returning a counter
  Callable plus a registry-count probe; `test/demo` GDScript exercises
  `call`, `call_deferred`, `str(cb)`, double-`connect` dedup, and
  release-on-drop before the shutdown marker.
- `docs/roadmap.md`: replace the stale "Risk to spike first" paragraph — the
  spike outcome was that the missing-symbol risk was an interface-loader name
  bug, already fixed.

## Capabilities

### New Capabilities

- `custom-callable-binding`: creating engine Callables backed by Go closures —
  creation, invocation and argument semantics, identity/dedup, failure
  reporting, and deterministic registry cleanup.

### Modified Capabilities

None. `variant-copy-ownership`, `interface-function-loading`, and
`ref-method-boundary` requirements are consumed as-is, not changed.

## Impact

- New files: `pkg/core/go_callable.go`, `pkg/core/go_callable.c/.h` (shim),
  tests; `test/pkg` fixture methods; `test/demo/main.gd` scenarios.
- No codegen, no `pkg/ffi` changes (wrappers already generated), no
  public API breaks — purely additive.
- Relies on `callable_custom_create2` (always loaded) and A3's Callable
  copy-constructor marshalling across method boundaries.

## Non-goals

- Typed sugar (`NewGoCallable(fn any)` with reflection-based arg/return
  typing) — additive wrapper in a later change.
- `hash`/`equal`/`is_valid`/`get_argument_count`/`less_than` callbacks —
  unnecessary with pair-identity; revisit only if a future need appears.
- Go-side `await`/signal connection (C2), signal-emission helpers (C4).
- Binding Go *methods* (as opposed to closures) to existing class instances
  with `object_id` identity.
