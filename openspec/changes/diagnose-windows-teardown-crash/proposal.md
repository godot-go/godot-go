# Proposal

## Why

GitHub issue #172: on Windows CI every `godot` process that loads the `libgodotgo` GDExtension and then exits crashes during teardown, after a fully clean suite report. The crash fires in two independent run modes — the test suite run (`exit 139`, `FAIL[godot-exited-non-zero]` after `1430 passed, 0 failed`) and the editor project-files regression step (`make: Error 2816`, exception code 0xB = STATUS_ACCESS_VIOLATION) on a nearly instant `--editor --quit` — so the entire repro is "godot loaded the Go c-shared DLL, then exited".

Neither of the two handlers that could report the fault produces any output: godot's mingw CRT `signal(SIGSEGV)` crash handler never prints its `Program crashed` banner, and Go's runtime never prints a `SIGSEGV: segmentation violation` dump. The crash therefore happens in a phase where exception dispatch is already dead (almost certainly the loader-lock region of `ExitProcess`/DLL detach or thread kill), and every fix hypothesis so far (Go runtime thread termination, mingw CRT detach inside the c-shared DLL, our dbghelp/deinit state, engine cleanup) is unfalsifiable without the faulting stack.

We cannot fix what we have never seen. This change adds the diagnostics to *see* the crash; the fix itself is a separate, follow-up change.

## What Changes

Windows CI gains a diagnostics-only workflow (`workflow_dispatch`, plus the same godot/toolchain setup as `ci_windows.yaml`) that runs the fast deterministic repro under outsiders' control:

1. The `--editor --quit` repro (and optionally the full suite path) executed under `gdb --batch`, which intercepts the faulting exception before in-process dispatch dies — even during `ExitProcess` teardown it pays off, unlike the process's own handlers.
2. A Windows Error Reporting `LocalDumps` minidump fallback writing an on-disk crash dump of the dying godot process, uploaded as an artifact.
3. An assertion that the run produced a faulting stack (crash indicators in the log), so the diagnostics job itself reports loudly when the crash *did not* reproduce or evidence capture silently broke.
4. Any crash evidence (gdb transcript, minidump, module+offset of the faulting instruction) uploaded as `windows-crash-evidence` artifacts regardless of step failure.

No product code, no Makefile targets, and no normal-CI-path behavior change: `ci_windows.yaml` stays untouched, and ordinary PR runs behave exactly as today.

## Capabilities

### New Capabilities

(none — tooling/diagnostics only, no spec-level behavior change)

### Modified Capabilities

(none)

## Impact

- New `.github/workflows/ci_windows_crash_debug.yaml` (dispatch-only).
- Windows runner minutes only when dispatched; no dependency on, or change to, the existing windows CI gates (`Verify Stacktrace Capture Evidence` remains intact).
- One-off diagnostic evidence: when all three diagnostics are green, findings are transcribed into the change record and issue #172, and the follow-up fix change is designed from the observed faulting stack.
