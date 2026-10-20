# Tasks

## 1. C-side opt-in gating

- [ ] 1.1 Gate the `<windows.h>`/`<dbghelp.h>` includes, the DbgHelp capture body, and `stacktraceCaptureAvailable()` in `pkg/log/stacktrace.c` behind `GODOT_GO_ENABLE_DBGHELP`, keeping the execinfo branches untouched. Verify: a default Windows build links and its module import table contains no `dbghelp.dll`; a build with `EXTRA_CGO_CFLAGS='-DGODOT_GO_ENABLE_DBGHELP'` compiles the capture path; `go build ./pkg/log ./pkg/core` succeeds on Linux.
- [ ] 1.2 Update the `pkg/core/stacktrace.go` doc comments to state the Windows opt-in and the new meaning of `StacktraceCaptureAvailable()`. Verify: `go vet ./pkg/core` passes and the comments name the macro.

## 2. Documentation

- [ ] 2.1 Rewrite the README Windows capture paragraph: opt-in default, the `EXTRA_CGO_CFLAGS='-DGODOT_GO_ENABLE_DBGHELP'` invocation, the reason (issue #172 mitigation), and that `capture=0` on the probe line means compiled out. Verify: README contains the macro name and the compiled-out probe meaning, and still documents the existing PDB/offline-resolution recipe.

## 3. CI evidence alignment

- [ ] 3.1 Update the Windows "Verify Stacktrace Capture Evidence" step in `.github/workflows/ci_windows.yaml`: explicit `capture=0` passes with a mitigation note; `capture=1` without banner or without a file:line frame still fails. Verify: the workflow YAML parses and a Windows run whose probe reports `capture=0` passes the step.

## 4. Integration verification

- [ ] 4.1 Run the Windows CI suite on the branch and confirm the probe reports `capture=0`, the evidence step passes, and no DbgHelp-related output appears. Verify: run logs show `stacktrace-probe capture=0` and a green evidence step.
- [x] 4.2 Record whether the teardown crash persists without DbgHelp loaded (plain uninstrumented repro exit code: 139 = dbghelp cleared, 0 = suspect confirmed) and feed the result to `openspec/changes/diagnose-windows-teardown-crash/`. Measured on run 39360425245 (v4 workflow + this branch): the plain editor repro still exits 139 with capture compiled out and the probe reporting `capture=0` — the teardown crash is independent of DbgHelp, so dbghelp is cleared as a cause and the opt-in is an isolation/mitigation, not a fix. The suite second-chance procdump pass again produced no dump (same instrumentation-aversion as before).

## Workflow follow-up

- Once issue #172 is resolved, decide whether to re-enable the Windows capture by default or keep it opt-in permanently; add an opt-in capture build to Windows CI if the toggle outlives the investigation.
