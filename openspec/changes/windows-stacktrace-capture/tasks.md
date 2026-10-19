# Tasks

## 1. Gate-first (fixtures red before product)

- [x] 1.1 Add the stacktrace rules to `test/check_test_output.sh`: (a) probe line `stacktrace-probe capture=<0|1>` must be present in a clean run; (b) `capture=1` claims require >=1 `___ C start backtrace ___` banner. Place before the catch-all non-zero exit check. Verify `make test_gate` fails on `clean_pass.log`/`godot_nonzero.log` (they predate the probe).
- [x] 1.2 Update `clean_pass.log` and `godot_nonzero.log` to carry probe line + banner; add fixtures `stacktrace_silent.log` (capture=1, no banner -> fail), `stacktrace_stub.log` (capture=0 -> pass), `stacktrace_probe_missing.log` (no probe -> fail); extend the `run_fixture_tests.sh` table. Verify `make test_gate` green.
- [x] 1.3 Confirm product-level red: `make test` on the host fails with `FAIL[stacktrace-probe-missing]`.

## 2. Windows capture and availability predicate

- [x] 2.1 Implement the `_WIN32` branch in `pkg/log/stacktrace.c` per design D1/D2 (SRWLOCK-guarded DbgHelp walk, capped frames, `module+0xoffset (ip 0x…)` fallback, same banners); add `stacktraceCaptureAvailable()` to `pkg/log/stacktrace.c`/`.h` returning true on `_WIN32` or `GODOT_GO_HAVE_EXECINFO`, false otherwise. Verify the glibc path unchanged (`make build` + host suite).
- [x] 2.2 Add `#cgo windows LDFLAGS: -ldbghelp` to the cgo preamble of the package that compiles `stacktrace.c` (`pkg/log/lib.go`), plus exported Go helpers `SetPrintStacktrace(bool)` / `StacktraceCaptureAvailable() bool` in `pkg/core` (new `stacktrace.go`; matches the header's documented seam choice)
- [x] 2.3 No mingw toolchain on the dev host (apt install unavailable); the `_WIN32` branch was instead syntax-checked against stub windows/dbghelp headers (`cc -fsyntax-only -Wall -std=gnu90 -D_WIN32`), and the Windows CI job is the real compile verification

## 3. Test-suite probe

- [x] 3.1 Add `Example.TestStacktraceProbe` (`test/pkg/example.go`) + registration; start prints `stacktrace-probe capture=<0|1>` and opens the window, hit/stop return fixed strings and close it.
- [x] 3.2 Add the `test_stacktrace_probe(example)` section to `test/demo/main.gd` asserting the three phase returns.
- [x] 3.3 Host `make test` green: banner present in `test-output.log` at least once, absent after the probe window closes, assertion count +3.

## 4. Integration checks

- [x] 4.1 `GODOT=/path/to/godot make test` fully green on the glibc host (Known Issues ignored) and `make test_gate` green.
- [x] 4.2 `openspec validate windows-stacktrace-capture` passes.

## 5. Implementation-discovered follow-ups

- [x] 5.1 Windows `Test` step found to run `make test` under `continue-on-error: true` (pre-existing godot teardown segfault, exit 139, reproduces on `main` without this change) — the in-suite gate cannot fail that job. Added a final `Verify Stacktrace Capture Evidence` step to `ci_windows.yaml` grepping `test-output.log` for the probe line and banner (design D7).
- [x] 5.2 Confirm the new verify step passes on this PR's Windows CI run ("windows stacktrace capture verified") — run 39182125987, with the masked `FAIL[godot-exited-non-zero]` (#172) still visible underneath, proving the evidence step gates independently of the red `make test` step.

## 6. Companion Go stack + offline resolution docs (follow-on, approved in chat)

- [x] 6.1 Gate-first: added the `stacktrace-go-stack-missing` rule (probe line present requires the marker followed by a `file.go:line` frame). RED: `clean_pass.log`, `godot_nonzero.log`, `stacktrace_stub.log` fail with the new code (`stacktrace_silent.log` keeps failing its own earlier silent-capture rule by design).
- [x] 6.2 Fixtures: go-stack block added to the four probe-bearing logs, new `stacktrace_no_go_stack.log` (clean run minus the companion) pinned to `FAIL[stacktrace-go-stack-missing]`; 22/22 fixture cases green. Product RED confirmed (`make test` -> `FAIL[stacktrace-go-stack-missing]`).
- [x] 6.3 `Example.TestStacktraceProbe("hit")` prints `stacktrace-probe go-stack` + `runtime.Stack` output; host `make test` green at 1430 assertions with real frames (`test/pkg/example.go:…` plus the reflect dispatch chain).
- [x] 6.4 README: offline `go tool addr2line` recipe with the GNU-addr2line-hangs warning, and the companion-stack note.

## 7. Offline resolution artifacts (follow-on, approved in chat)

- [x] 7.1 `Upload Test Artifacts` step in `ci_windows.yaml` (`upload-artifact@v4`, `if: always()`, 14-day retention): `test-output.log`, `test/demo/lib/*.dll`, plus a `*.pdb` glob reserved for the stacked line-info change.
- [x] 7.2 Downloaded the artifact (run 39193759198) and verified against the real mingw-linked DLL: preferred base `0x1e8d10000` (not the conventional `0x140000000` — discovered via `llvm-objdump -p`, must not be assumed); the run's own frames `+0x98823`/`+0x8f6b2` resolve to `runtime.asmcgocall` `runtime/asm_amd64.s:973` and `runtime.cgocall` `runtime/cgocall.go:185`.
- [x] 7.3 README: Windows offline recipe (`llvm-objdump -p` base discovery + `go tool addr2line`) exactly as executed against the artifact.

## Workflow follow-up

- PR base: `fix-portable-stacktrace-build` (stacked; merge bottom-up per the pull-request skill).
- Confirm `build on alpine` + both test jobs green; the windows job is the first-ever proof of Windows capture end to end.
