# Proposal

## Why

Godot's MinGW Windows builds capture native stacks with the libbacktrace copy vendored at `thirdparty/libbacktrace` (`platform/windows/crash_handler_windows_signal.cpp`), while godot-go's Windows capture uses DbgHelp. That path is opt-in because of the old teardown suspicion, requires CodeView/PDB and an lld link, and cannot read mingw DWARF live. The module-pin fix (#180) removed the unload hazard, so godot-go can adopt the same library as Godot, restore default Windows capture, and drop the PDB machinery.

## What Changes

- Replace the Windows DbgHelp capture in `pkg/log/stacktrace.c` with a vendored copy of Godot 4.7-stable's PE/COFF libbacktrace (BSD-3-Clause, Windows-only build constraints), and use libbacktrace to symbolize frames in the extension DLL.
- Windows capture becomes compiled in by default: remove `GODOT_GO_ENABLE_DBGHELP`, the `dbghelp.dll` import, and the "capture unavailable" default state.
- Remove the CodeView/PDB build path: the `-gcodeview` directive in `pkg/log` and the lld `/DEBUG --pdb` link flags in the `Makefile`. C translation units keep DWARF.
- Frames in the extension DLL print a function plus source `file:line` when libbacktrace resolves them; frames in other modules keep the `module+0xoffset (ip ...)` form.
- Linux/macOS keep `execinfo`; musl keeps the existing no-op and `GODOT_GO_ENABLE_EXECINFO` opt-in.
- Windows CI expects `capture=1`, a backtrace banner, and at least one DWARF-resolved `file:line` frame; it stops treating `capture=0` as healthy and stops uploading a PDB.

## Capabilities

### New Capabilities

(none)

### Modified Capabilities

- `native-stacktrace-capture`: replace the Windows DbgHelp, opt-in, and PDB requirements with libbacktrace-based, default-compiled-in Windows capture, and update the CI evidence scenarios.

## Impact

- Code: `pkg/log/stacktrace.c`, `pkg/log/stacktrace.h`, `pkg/log/lib.go`, `pkg/core/stacktrace.go`.
- New vendored sources under `pkg/log/libbacktrace/` (exact cgo packaging is settled by an early implementation spike).
- Build and CI: `Makefile`, `.github/workflows/ci_windows.yaml`, `README.md`.
- No public Go API change: `SetPrintStacktrace` and `StacktraceCaptureAvailable` keep their signatures. Windows builds lose the opt-in flag and the PDB artifact; capture becomes available by default.
- Spec delta: `openspec/specs/native-stacktrace-capture/spec.md`.

## Non-goals

- Using libbacktrace on Linux/macOS/musl instead of `execinfo`; only Windows changes.
- Symbolizing engine or GDScript frames, or adding a crash-time signal handler.
- Resolving Go frames through libbacktrace; the existing `runtime.Stack` companion stays the Go-frame source.
- Changing the callback-boundary enable flag or producing output while capture is off.
- Fixing upstream libbacktrace or tracking newer upstream releases.
