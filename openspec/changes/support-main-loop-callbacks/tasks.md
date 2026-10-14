# Tasks

## 1. Core binding (pkg/core)

- [ ] 1.1 Add the three nullable hook slots (startup/frame/shutdown) beside
      `GDExtensionBindingInitCallbacks` in `pkg/core/lib.go` and the three
      `InitObject.RegisterStartupCallback/RegisterFrameCallback/RegisterShutdownCallback`
      methods in `pkg/core/godot.go`; verify with `go build ./...` and
      `go vet ./pkg/core/`
- [ ] 1.2 Create `pkg/core/main_loop_callback.go` with exported
      `GoCallback_MainLoopStartup/Frame/Shutdown` trampolines (nil-check slot,
      invoke if set, no recover) and a registration helper that builds a
      `GDExtensionMainLoopCallbacks` struct — trampoline pointer only for
      filled slots — and calls
      `CallFunc_GDExtensionInterfaceRegisterMainLoopCallbacks(FFI.Library, ...)`;
      verify with `go build ./...`
- [ ] 1.3 Call the registration helper from `GDExtensionBindingInitializeLevel`
      when `pLevel == CORE`, gated on any slot non-nil AND
      `FFI.RegisterMainLoopCallbacks != nil`, with a debug log on the skip
      path; verify the demo extension still loads and logs normally via
      `GODOT=/home/pcting/bin/godot make test` (no new warnings, existing
      suite green)

## 2. Demo fixtures and driver assertions

- [ ] 2.1 In `test/pkg/example.go`, register startup/frame/shutdown hooks on
      the `InitObject` before `Init()` recording into Go atomics, and expose
      `main_loop_startup_ran()` / `main_loop_frame_count()` accessors on a
      registered fixture class; verify `make build` succeeds and the demo
      loads
- [ ] 2.2 In `test/demo/main.gd`, assert the startup evidence is already set
      in `_ready`, then await `process_frame` N times and assert the frame
      count grew by at least N between samples; verify with
      `GODOT=/home/pcting/bin/godot make test` — new assertions reported as
      passes
- [ ] 2.3 Make the shutdown hook log the unique marker literal
      `GODOTGO_MAIN_LOOP_SHUTDOWN` (level surviving `LOG_LEVEL=WARN`); verify
      the marker line appears in `test-output.log` after
      `GODOT=/home/pcting/bin/godot make test`

## 3. Harness gate (test/check_test_output.sh)

- [ ] 3.1 Add a shutdown-marker gate to `test/check_test_output.sh` requiring
      `GODOTGO_MAIN_LOOP_SHUTDOWN` with a distinct failure message naming the
      missing main-loop shutdown marker; verify both directions: checker
      passes on the fresh `test-output.log`, and fails with the distinct
      message on a copy of that log with the marker lines stripped

## 4. Docs and close-out

- [ ] 4.1 Add a "Main-loop callbacks" section to `docs/overview.md`: the
      three `InitObject` methods, register-before-`Init()` rule, CORE-level
      timing, pre-4.5 skip behavior, and the no-panic-recovery policy; verify
      by review against the delta spec scenarios
- [ ] 4.2 Mark roadmap C3 DONE in `docs/roadmap.md` (gap entry + ordering
      table + gantt); verify by review
- [ ] 4.3 Full gate: `GODOT=/home/pcting/bin/godot make test` green —
      existing assertions pass, zero failures, no leaked engine objects,
      shutdown-marker gate satisfied
