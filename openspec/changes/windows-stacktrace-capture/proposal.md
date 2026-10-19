# Proposal: windows-stacktrace-capture

## Why

`pkg/log/stacktrace.c`'s `_WIN32` capture branch is fully commented-out DbgHelp scaffolding: on Windows, `printStacktrace()` compiles to a no-op even when `enablePrintStacktrace` is set, and no test or CI job exercises capture on any platform (the only enable site in `pkg/core/godot.go` is commented out and tests run at `LOG_LEVEL=WARN`). The portable-stacktrace-build change made the feature's degradation portable but left its actual Windows operation dead and unverified.

## What Changes

- Implement the `_WIN32` branch of `printStacktrace()` with DbgHelp: `RtlCaptureContext` + a capped `StackWalk64` loop, symbol names via `SymFromAddr` when resolvable, and a `module.dll+0xoffset (ip 0x…)` fallback line otherwise (the address form addr2line/gdb can resolve — mingw builds carry DWARF, the system `dbghelp.dll` cannot read it, so names will often be unavailable and the offset form is the useful one).
- Add `stacktraceCaptureAvailable()` to the C interface (true where capture is compiled in: `_WIN32` or execinfo; false on the musl stub), exposed to Go through `pkg/core` (`SetPrintStacktrace`, `StacktraceCaptureAvailable`).
- Link `-ldbghelp` for Windows targets via `#cgo windows LDFLAGS` in the package that includes the header.
- Add a narrow capture probe to the test suite: `Example.test_stacktrace_probe("start"|"hit"|"stop")` prints one `stacktrace-probe capture=<0|1>` line, opens a three-call capture window (≈2 banners per run — no log spam), and the GDScript driver asserts the phase returns.
- Extend the runtime gate `test/check_test_output.sh`: when the log claims `capture=1`, at least one `___ C start backtrace ___` banner MUST appear; a run without the probe line at all fails. New fixtures pin both failure modes and the stub (`capture=0`) pass.
- Add a dedicated Windows CI step asserting the capture evidence in `test-output.log`. Discovered during implementation: the Windows `Test` step runs `make test` with `continue-on-error: true` (a pre-existing teardown segfault, tracked separately), so the in-suite gate alone cannot fail that job.

The Linux job gates `make test` normally, so the probe is enforced there through the existing gate.

## Impact

- Affected specs: `native-stacktrace-capture` (ADDED: Windows native capture; CI-visible capture probe)
- Affected code: `pkg/log/stacktrace.c/.h`, `pkg/log/lib.go` (cgo LDFLAGS), `pkg/core/stacktrace.go` (Go helpers), `test/pkg/example.go`, `test/demo/main.gd`, `test/check_test_output.sh`, `test/fixtures/*`, `.github/workflows/ci_windows.yaml`
- Depends on: the reworked `stacktrace.c` cascade from `portable-stacktrace-build` (stacked branch)
- No behavior change by default: the enable flag stays false outside the probe window; musl/glibc execinfo paths untouched except the additive availability predicate.

Closes #127's follow-up expectation (Windows capture actually works); no issue number of its own.
