# Proposal

## Why

Godot 4.5+ lets a GDExtension register `startup` / `frame` / `shutdown`
callbacks on the main loop (`register_main_loop_callbacks`), which language
bindings use as the per-frame ledge for their own runtime work. godot-go has
init levels and terminators but no frame hook, so Go extensions cannot observe
main-loop phases at all and roadmap item C2 (await/goroutine bridge) has
nothing to sit on. The FFI layer already loads and wraps the interface
function — only the `pkg/core` binding layer is missing.

## What Changes

- Three new `InitObject` registration methods mirroring godot-cpp master
  exactly: `RegisterStartupCallback`, `RegisterFrameCallback`,
  `RegisterShutdownCallback`, each taking a `GDExtensionBindingCallback`
  (`func()`). One slot per phase, assignment overwrites (last-write-wins),
  set before `Init()`.
- Engine registration fires at the CORE init level when at least one slot is
  filled, dispatching through exported Go trampolines into the Go slots.
- Older engines (pre-4.5, missing interface function) keep loading:
  registration is skipped with a log line, no crash.
- A panicking hook aborts the process — same policy as every existing
  `GoCallback_*` trampoline and godot-cpp's unrecovered C++ calls.
- Test demo gains fixtures proving each phase fires; `make test`'s checker
  (`test/check_test_output.sh`) gains a mandatory shutdown-marker gate so the
  `shutdown` phase stays regression-tested.
- No codegen changes: `GDExtensionMainLoopCallbacks`, its `CallFunc_` wrapper,
  and the `register_main_loop_callbacks` loader entry already exist in
  `pkg/ffi`.

## Capabilities

### New Capabilities

- `main-loop-callback-registration`: how Go-defined extensions register and
  receive main-loop `startup`/`frame`/`shutdown` phase callbacks, including
  registration timing, engine-version tolerance, and dispatch ownership.

### Modified Capabilities

- `test-harness-integrity`: the "genuinely clean run passes" green condition
  gains a mandatory main-loop shutdown marker in the run output; its absence
  fails the target with a distinct message.

## Impact

- `pkg/core`: new main-loop callback state, exported trampolines,
  registration call inside `GDExtensionBindingInitializeLevel`, three
  `InitObject` methods (`godot.go`, new `main_loop_callback.go`).
- `test/pkg/example.go`: hook fixtures (startup flag, frame counter, shutdown
  marker log) exposed to GDScript via a registered class.
- `test/demo/main.gd`: assertions for startup-before-ready and frame counting
  across awaited frames.
- `test/check_test_output.sh`: new shutdown-marker gate.
- `Makefile`: headless invocation changes from `--quit` to `--quit-after 5`
  so the driver can observe multiple main-loop iterations (single-iteration
  runs cannot show frame-hook growth and lose the summary banner on any
  second `await process_frame`).
- `docs/overview.md`, `docs/roadmap.md`: document the feature, mark C3 done.
- No public API breakage; existing callers untouched.

## Non-goals

- Multiple subscribers per phase (fan-out) or deregistration — godot-cpp
  parity means one slot per phase.
- Panic recovery or per-hook fault isolation inside trampolines.
- Any consumer built on the frame hook (C2 await bridge, deferred Callable
  pump) — C3 only keeps the ledge clear.
- Per-level main-loop wiring or editor-runtime-only phases (C4).
