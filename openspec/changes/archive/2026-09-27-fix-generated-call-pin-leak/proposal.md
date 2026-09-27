# Proposal

## Why

Generated call paths pin every scratch object to a package-global `runtime.Pinner` that is never unpinned. Six packages declare one — `pkg/builtin/lib.go:33`, `pkg/core/lib.go:38`, `pkg/ffi/lib.go:19`, `pkg/gdclassimpl/lib.go:12`, `pkg/gdclassinit/lib.go:21`, `pkg/gdutilfunc/lib.go:15` — and together they take **65,995** `pnr.Pin(...)` calls (63,086 of them in `pkg/gdclassimpl/classes.gen.go`) against **zero** `Unpin` calls anywhere under `pkg/`. Per the Go runtime, a Pinner releases its referents only after the Pinner itself becomes unreachable — a package-level variable never does — so every generated call permanently retains its return slot, owner pointer cell, argument pointer slice, and every argument slot.

The leak is observable today: `CollisionShape2D.GetShape()` allocates a `Shape2DImpl` wrapper per call and pins it forever, keeping the refcounted Godot resource reachable at shutdown and tripping the engine's exit check (`Leaked instance: CircleShape2D ... RID allocations of type 'GodotShape2D' were leaked at exit`). 51 `SetShape` calls with no `GetShape` leaked nothing; adding `GetShape` leaked (sibling change `fix-object-arg-ptrcall-encoding`, task 4.3). In a frame loop — thousands of calls per frame — retention is unbounded.

## What Changes

- Replace the global never-unpinned pinner in **generated per-call paths** with a scoped per-call `runtime.Pinner` released by `defer pinner.Unpin()` when the call completes — the pattern already written (but commented out) in `cmd/generate/ffi/ffi_wrapper.go.tmpl`.
- Edit templates only — `classes.go.tmpl`, `builtinclasses.go.tmpl`, `variant.go.tmpl`, `utilityfunctions.go.tmpl`, and the FFI return-value pin in `cmd/generate/ffi/templatefunctions.go` — then propagate with `make generate`. No `*.gen.*` file is hand-edited.
- Apply the same scoped pattern to the two hand-written per-call helpers that share the leak: `pkg/ffi/builtin_ptrcall.go` and `pkg/core/method_bind_callback.go`.
- Keep registration-time pins on the package pinner — `GoMethodMetadata` PropertyInfo `StringName`/`String` retention, instance-binding callbacks, `FFI.GodotVersion`, `ClassInfo` — because C reads them for the metadata's lifetime and they are bounded by registration count, not call count.
- Add regression tests: weak-pointer collectability of dropped return wrappers, and a `GetShape` loop that exits with a clean engine leak check.

## Capabilities

### New Capabilities
- `native-call-pinning`: How Go↔Godot call paths bound the lifetime of pinned Go memory — call-scoped scratch is unpinned when the call completes, while values C retains beyond the call stay pinned for a matching lifetime.

### Modified Capabilities
(none)

## Impact

- Templates under `cmd/generate/` and their regenerated outputs (`pkg/gdclassimpl/classes.gen.go`, `pkg/builtin/*.gen.go`, `pkg/gdutilfunc/utilityfunctions.gen.go`, `pkg/ffi/ffi_wrapper.gen.go`).
- Hand-written helpers `pkg/ffi/builtin_ptrcall.go` and `pkg/core/method_bind_callback.go`.
- `test/pkg/` and `test/demo/main.gd` — leak regression tests.
- No public API changes; call signatures unchanged.

## Non-goals

- Not eliminating or redesigning pinning (no arenas or pools), and not auditing every existing pin away — the pins stay; their lifetime shrinks to the call.
- Not changing registration-time or one-time pinning; those remain on the package pinner by design (see design.md Non-Goals).
- Not editing `*.gen.*` files directly; all changes flow through `*.tmpl` plus `make generate`.
- Performance tuning beyond removing the leak (per-call Pinner cost is amortized by the runtime's per-P pinner cache).
