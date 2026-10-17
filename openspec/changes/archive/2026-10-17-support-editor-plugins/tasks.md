# Tasks

## 1. Headless-editor spike (design D5)

- [x] 1.1 Run the current test demo in the editor (`godot --headless -e --quit-after 12 --path test/demo`) and record: does the extension load and shut down clean, is `--quit-after` honored, what leak/shutdown markers appear. Verification: recorded notes; if `--quit-after` is not honored, confirm `--quit` works and note it for 4.2.
- [x] 1.2 Throwaway spike: register a minimal `EditorPlugin`-derived fixture class and call the raw `CallFunc_GDExtensionInterfaceEditorAddPlugin` wrapper from a demo editor-level init callback. Verification: spike run shows the Go constructor firing (printed marker), the editor quits, and stderr reports zero ObjectDB leaks. Keep findings in the change folder; revert spike code or promote it into task 4.1.

## 2. Registry and public API (pkg/core)

- [x] 2.1 Add the editor plugin registry (`pkg/core/editor_plugins.go`): mutex-guarded ordered class-name list with pure decision seams (`shouldAdd`/`shouldRemove` style: duplicate-add rejected, unknown-remove rejected, accepted calls mutate state only on success). Verification: engine-free unit tests cover accept, duplicate, unknown, and drain-clears-order; `go test ./pkg/core` green.
- [x] 2.2 Implement `AddEditorPlugin[T GDClass](inst T)` / `RemoveEditorPlugin[T GDClass](inst T)`: `GetEditorInterfaceSingleton() == nil` → recoverable `log.Panic`; then registry decision; then per-call StringName build → FFI call → `Destroy()`. Verification: unit tests pin guard ordering on the pure seams (nil singleton never reaches registry); whole repo builds (`make build`).

## 3. Automatic unload

- [x] 3.1 Drain the registry in `GDExtensionBindingDeinitializeLevel` for the EDITOR level, after the user terminator callback, in insertion order (design D4). Verification: engine-free test drives the drain seam with a recorded fake removal list and asserts terminator-before-drain ordering and registry emptiness; `go test ./pkg/core` green.

## 4. Harness fixtures and editor gate

- [x] 4.1 Promote the spike into a permanent fixture: `ExampleEditorPlugin` (struct embedding `EditorPluginImpl`, class name equal to struct name), registered at editor init level and added via `AddEditorPlugin`; constructor prints the driver PASS signal exactly once. Verification: editor run from 1.1 shows the fixture PASS marker and clean shutdown.
- [x] 4.2 Add the Makefile editor-run stage and extend `test/check_test_output.sh` with an editor-mode gate: fixture marker present, clean shutdown marker, zero ObjectDB leaks, exit code 0. Verification: `GODOT=/home/pcting/bin/godot make test` executes both runtime and editor stages; deliberately breaking 4.1 fails the editor gate.
- [x] 4.3 Add a runtime-build misuse probe to the runtime harness: call `AddEditorPlugin` from runtime code inside `recover()`, assert the recoverable panic fires and no engine registration happened (PASS verdict printed by the existing driver). Verification: runtime gate counts the new PASS and zero FAILURES.
- [x] 4.4 Optional C1 composition probe in the editor run: fixture calls `AddToolMenuItem` with a Go `Callable` (via the C1 bridge) and prints a PASS marker; if infeasible headless, record the reason in the tasks file verification notes instead. Verification: marker present or documented skip.

## 5. Verification and closeout

- [x] 5.1 Full gate green: `GODOT=/home/pcting/bin/godot make test` passes runtime (baseline 1381 PASSES plus new probes, 0 FAILURES, zero leaks) and the new editor stage. Verification: gate script exit 0.
- [x] 5.2 Mutation-proof the change: temporarily remove the task 3.1 auto-drain call and confirm the editor gate fails; restore and reconfirm green. Verification: fail/green sequence recorded in verification notes.
- [x] 5.3 Lint/format: `gofmt`/`goimports` clean; `/home/pcting/go/bin/revive -config .revive.toml` reports no new finding categories vs baseline. Verification: command output diffed against baseline.
- [x] 5.4 Docs closeout: `docs/overview.md` gains the editor plugin capability; `docs/roadmap.md` C4 marked DONE (section note, Phase table row, both gantt `done` tags). Verification: manual review against shipped API.

## Verification notes

### 1.1 Editor baseline (2026-10-16)

- `godot --headless --verbose --path test/demo/ --editor --quit-after 12`: exit 0.
  Extension loads in the editor, `GODOTGO_MAIN_LOOP_SHUTDOWN` fires during
  teardown, zero ObjectDB leaks, `--quit-after` is honored with `--editor`.
- No `TESTS FINISHED` driver summary exists in editor mode (the main scene does
  not run), so the editor gate (task 4.2) must key on the fixture marker
  instead of the runtime driver banner.
- Orphan-StringName and leak reports are only emitted with `--verbose`;
  the runtime baseline log (`test-output.log`, verbose) contains zero orphans,
  which stays the invariant to preserve.
- Pre-existing, C4-unrelated: `ERROR: Object 'Example' already has member ''`
  appears in current runtime and editor logs alike.

### 1.2 Spike findings (2026-10-16)

- Register an `EditorPlugin`-derived Go class at the EDITOR level + raw
  `editor_add_plugin` → engine queues → `EditorNode` ctor drains →
  instantiation routes through `GoCallback_ClassCreationInfoCreateInstance2`;
  editor exits clean, zero ObjectDB leaks.
- **Engine-side instantiation uses `reflect.New(ci.ClassType)`, NOT the
  constructor function passed to `ClassDBRegisterClass`** (that map only feeds
  the decode path in `method_bind_reflect.go`). The observable fixture signal
  must be a bound virtual: `ClassDBBindMethodVirtual(t, "V_..._EnterTree",
  "_enter_tree", …)` fired when the editor added the plugin to the tree.
- Manual `classdb_construct_object3` probe returned non-nil — registration
  routing healthy.
- Verbose run after spike: orphan `ExampleEditorPlugin` (static: 0, total: 2)
  remained because the spike never unregistered the class; 4.1 must register
  an editor terminator calling `ClassDBUnregisterClass[*ExampleEditorPlugin]()`
  (mirrors `UnregisterExampleTypes`). Orphans `_enter_tree`/`EditorPlugin`
  (static: 1) are engine `SNAME` static interns by design, triggered by the
  add/virtual paths.
- Spike artifact: `test/pkg/editor_plugin_spike.go` — to be promoted into the
  4.1 fixture, not kept as-is.

### 4.x Harness results (2026-10-16)

- Permanent fixture in `test/pkg/editor_plugin_tests.go`: editor run prints
  `GODOTGO_EDITOR_PLUGIN_ENTERED` and `GODOTGO_EDITOR_C1_CALLABLE_PASS`
  (WARN level), shuts down with `GODOTGO_MAIN_LOOP_SHUTDOWN`, exits 0, zero
  leaks and zero orphan StringNames.
- 4.4 headless feasibility: `AddToolMenuItem` with a `NewGoCallable` Callable
  works headless; invoking the same Callable through `Callv` dispatched into
  the Go closure exactly once (menu itself never clicked — headless has no
  GUI, the direct call is the proof signal).
- **TOOLS-build landmine**: `main.cpp` gates `initialize_extensions(EDITOR)`
  on `TOOLS_ENABLED`, not on the editor session — EDITOR-level init runs even
  when a TOOLS build runs the project as a game, with no `EditorInterface`
  singleton present. The first full-gate run proved this: the D3 guard
  correctly panicked during the runtime stage when the fixture added
  unconditionally. Fix: the fixture self-gates on
  `GetEngineSingleton().IsEditorHint()` with a symmetric terminator flag;
  design D3 amended.
- 4.2: `make test` runs runtime stage, then editor stage (`--editor
  --quit-after 12`, `editor-output.log`, gate `check_editor_output.sh`;
  runtime gate untouched). The gate's own failure modes are pinned by eight
  fixtures in `test/fixtures/` (incl. `editor_static_intern_ok.log` pinning
  the static-1 tolerance); fixture runner extended to drive both gates —
  18/18 fixture tests green. Fixed pre-existing drift found en route: runtime
  fixtures `clean_pass.log`/`godot_nonzero.log` predated the C3
  shutdown-marker check (198a926) and failed `make test_gate` before this
  change; markers appended to both.

### 5.1/5.2 Gate and mutation proof (2026-10-16)

- 5.1: `GODOT=/home/pcting/bin/godot make test` exit 0 — runtime `PASS:
  extension loaded, 1382 assertions ran, 0 failures, main-loop shutdown
  marker present, no leaked engine objects` (baseline 1381 + the 4.3 probe)
  and editor `PASS: editor plugin fixture instantiated, main-loop shutdown
  marker present, no leaked engine objects, no orphaned StringNames`.
- 5.2 M1 (auto-drain call removed from `deinitializeBindingLevel`): unit test
  `TestDeinitializeBindingLevelDrainsAfterTerminator` fails, and the in-engine
  editor gate fails with `FAIL[orphaned-string-names]` — the never-removed
  plugin leaves `Orphan StringName: ExampleEditorPlugin (static: 0,
  total: 1)` because the engine-side registration outlives the extension.
- 5.2 M2 / 4.2 deliberate fixture break (`AddEditorPlugin` disabled): editor
  gate fails with `FAIL[no-fixture-marker]` (godot exit 0 — the failure is
  silent without the gate).
- Both mutations restored; full `make test` reconfirmed green, same PASS
  lines as 5.1.
