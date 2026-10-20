# Proposal

## Why

The Windows teardown crash (issue #172) happens after the last engine callback into Go, during process teardown, and every in-process handler is silent there. The C-side stacktrace capture is a standing suspect: on Windows it links `dbghelp.dll` into the extension, so even a build that never enables capture at runtime still imports dbghelp, and dbghelp's process-detach path runs during exactly the phase that crashes.

Making the Windows capture opt-in removes dbghelp from a default process entirely (not just from the runtime path), which both prevents the suspected component from running in the teardown window and turns the build into a clean experiment: if the crash disappears without dbghelp loaded, the suspect is confirmed; if it persists, dbghelp is cleared. The capability stays available on request, so capture work is not lost while the crash is investigated.

## What Changes

- Windows native capture becomes compile-time opt-in behind `GODOT_GO_ENABLE_DBGHELP`: the `windows.h`/`dbghelp.h` includes, the capture body, and the availability report are gated together, so a default Windows build does not import dbghelp at all and `stacktraceCaptureAvailable()` reports false.
- The documented opt-in is `make build EXTRA_CGO_CFLAGS='-DGODOT_GO_ENABLE_DBGHELP'`, matching the existing `GODOT_GO_ENABLE_EXECINFO` pattern for musl.
- The Windows CI capture-evidence step accepts the probe's explicit `capture=0` state as healthy (native capture compiled out) and keeps failing a claimed-but-silent capture; the PDB file:line assertions apply whenever capture is compiled in.
- The execinfo paths (glibc/Apple automatic, musl external implementation) are unchanged; no Go API shape changes.

## Capabilities

### New Capabilities

_None._

### Modified Capabilities

- `native-stacktrace-capture`: Windows capture becomes opt-in at build time instead of always compiled in; the Windows CI line-info requirement applies only when capture is compiled in, and the default build mode reports capture unavailable.

## Impact

- `pkg/log/stacktrace.c` — conditional includes/body/availability for Windows.
- `pkg/core/stacktrace.go` — documentation of the new availability semantics.
- `README.md` — Windows capture section documents the toggle and its default.
- `.github/workflows/ci_windows.yaml` — capture-evidence step accepts the explicit compiled-out state.
- Windows users who rely on native capture must now opt in; nothing else about capture (banners, PDB line info, offline resolution) changes when it is enabled.
- Interim mitigation pending the diagnosis tracked in `openspec/changes/diagnose-windows-teardown-crash/`; re-enabling or replacing the toggle is a follow-up once #172 is resolved.
