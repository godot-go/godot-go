# Tasks

## 1. Core binding (pkg/core)

- [x] 1.1 Add the three nullable hook slots (startup/frame/shutdown) beside
      `GDExtensionBindingInitCallbacks` in `pkg/core/lib.go` and the three
      `InitObject.RegisterStartupCallback/RegisterFrameCallback/RegisterShutdownCallback`
      methods in `pkg/core/godot.go`; verify with `go build ./...` and
      `go vet ./pkg/core/`
- [x] 1.2 Create `pkg/core/main_loop_callback.go` with exported
      `GoCallback_MainLoopStartup/Frame/Shutdown` trampolines (nil-check slot,
      invoke if set, no recover) and a registration helper that builds a
      `GDExtensionMainLoopCallbacks` struct — trampoline pointer only for
      filled slots — and calls
      `CallFunc_GDExtensionInterfaceRegisterMainLoopCallbacks(FFI.Library, ...)`;
      verify with `go build ./...`
- [x] 1.3 Call the registration helper from `GDExtensionBindingInitializeLevel`
      when `pLevel == CORE`, gated on any slot non-nil AND
      `FFI.RegisterMainLoopCallbacks != nil`, with a debug log on the skip
      path; verify the demo extension still loads and logs normally via
      `GODOT=/home/pcting/bin/godot make test` (no new warnings, existing
      suite green)

## 2. Demo fixtures and driver assertions

- [x] 2.1 In `test/pkg/example.go`, add package-level atomic counters and
      `main_loop_startup_ran()` / `main_loop_frame_count()` accessors on the
      Example class (bound via `ClassDBBindMethod`), and in
      `test/pkg/lib.go` register startup/frame/shutdown hooks on the
      `InitObject` before `Init()` recording into them; verify `make build`
      succeeds and the demo loads
- [x] 2.2 Switch the Makefile headless invocation from `--quit` to
      `--quit-after 5` so GDScript coroutines can observe several full
      `Main::iteration`s (`--quit` runs exactly one: a second
      `await process_frame` would never resume and the banner would be lost);
      verify `GODOT=/home/pcting/bin/godot make test` still completes green
      with the same exit status
- [x] 2.3 In `test/demo/main.gd`, assert the startup evidence is already set
      in `_ready`; then one priming `await get_tree().process_frame`
      (resumption happens inside the signal dispatch, before that
      iteration's hook), then two measurement awaits each asserting the
      frame count grew by at least one since the previous sample; verify
      with `GODOT=/home/pcting/bin/godot make test` — new assertions
      reported as passes
- [x] 2.4 Make the shutdown hook log the unique marker literal
      `GODOTGO_MAIN_LOOP_SHUTDOWN` (level surviving `LOG_LEVEL=WARN`); verify
      the marker line appears in `test-output.log` after
      `GODOT=/home/pcting/bin/godot make test`

## 3. Harness gate (test/check_test_output.sh)

- [x] 3.1 Add a shutdown-marker gate to `test/check_test_output.sh` requiring
      `GODOTGO_MAIN_LOOP_SHUTDOWN` with a distinct failure message naming the
      missing main-loop shutdown marker; verify both directions: checker
      passes on the fresh `test-output.log`, and fails with the distinct
      message on a copy of that log with the marker lines stripped

## 4. Docs and close-out

- [x] 4.1 Add a "Main-loop callbacks" section to `docs/overview.md`: the
      three `InitObject` methods, register-before-`Init()` rule, CORE-level
      timing, pre-4.5 skip behavior, and the no-panic-recovery policy; verify
      by review against the delta spec scenarios
- [x] 4.2 Mark roadmap C3 DONE in `docs/roadmap.md` (gap entry + ordering
      table + gantt); verify by review
- [x] 4.3 Full gate: `GODOT=/home/pcting/bin/godot make test` green —
      existing assertions pass, zero failures, no leaked engine objects,
      shutdown-marker gate satisfied
