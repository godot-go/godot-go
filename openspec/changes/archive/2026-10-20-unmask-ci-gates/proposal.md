# Proposal

## Why

Two gating workflows still hide failures with `continue-on-error`: both platforms' `Gen Test Project Files` step crashes in Godot's `--editor --quit` DocTools race (`make: Error 134` on Linux, `Error 2816` on Windows) and the step is masked, so CI cannot prove the generation step works. The crash-debug workflow also runs two masked procdump steps and now asserts a minidump that no longer exists — its premise (reproduce and capture the teardown crash) was resolved and archived. No CI workflow should hide a failing step.

## What Changes

- `make ci_gen_test_project_files` stops quitting during startup docs: `--editor --quit` becomes `--editor --frame-delay 16 --quit-after 600`. `--quit-after` counts iterations, so the frame delay guarantees roughly ten seconds of main-loop time after startup for the doc-cache worker to finish while `doc` is still alive; the deferred generation then runs in a normal frame instead of the engine's shutdown use-after-null.
- `.github/workflows/ci_linux.yaml`: drop `continue-on-error` from `Gen Test Project Files`.
- `.github/workflows/ci_windows.yaml`: drop `continue-on-error` from `Gen Test Project Files`.
- Delete `.github/workflows/ci_windows_crash_debug.yaml`: the teardown diagnosis is archived, the crash is fixed (#180), and the workflow's remaining masked steps plus its "a minidump must exist" assertion can no longer succeed.
- Result: a repository-wide invariant — no workflow uses `continue-on-error`.

## Capabilities

### New Capabilities

(none — CI/tooling only)

### Modified Capabilities

(none — no product behavior changes; `test-harness-integrity` already requires the test gate to fail on a nonzero run)

## Impact

- `Makefile` (`ci_gen_test_project_files`), `.github/workflows/ci_linux.yaml`, `.github/workflows/ci_windows.yaml`; removes `.github/workflows/ci_windows_crash_debug.yaml`.
- Adds roughly 10 seconds to each CI job's generation step (a cold cache measured ~17 s locally; a warm cache exits sooner if the worker finishes before the frame budget).
- The engine's DocTools bug is not fixed; CI simply stops quitting during startup docs. The workaround is documented in the Makefile comment.

## Non-goals

- Fixing Godot's deferred `_gen_extensions_docs` use-after-null.
- Changing any test logic, gate script, or product code.
- Keeping the crash-capture workflow as a regression check; the normal Windows job already runs both test invocations with the gate on.
