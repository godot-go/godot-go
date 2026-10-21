# Proposal

## Why

macOS is a build target in-tree — a `darwin` branch in the Makefile, macOS entries in `test/demo/example.gdextension`, execinfo in `pkg/log/stacktrace.c` — yet no workflow runs it, so build, load, teardown and capture regressions stay invisible. A macOS run would also fail spuriously: both gate scripts strip ANSI with GNU-only `sed` escapes that BSD `sed` ignores. The first CI run showed how untested that path was: darwin does not compile, because `util.SetThreadName` exists only for linux and windows.

## What Changes

- Add `.github/workflows/ci_macos.yaml`, a two-leg matrix on native runners — `macos-latest` (arm64) and `macos-26-intel` (x64) — with `fail-fast: false`, so one architecture's failure cannot hide the other's.
- Each leg mirrors the Windows job: Godot 4.7 stable, header update, `make generate`, `make build`, `make ci_gen_test_project_files`, `make test`, capture evidence, per-architecture artifact upload.
- Capture evidence stops at `stacktrace-probe capture=1` plus the `___ C start backtrace ___` banner; the Windows `file:line` assertion does not port to execinfo.
- Make the ANSI strip in `test/check_test_output.sh` and `test/check_editor_output.sh` POSIX-portable (a literal ESC byte instead of GNU `\x1b`).
- Add the missing darwin `util.SetThreadName` as a no-op, matching Windows, so the extension compiles for macOS at all.
- Re-measure the pin-flatness probe before failing: arm64 grew 416 KiB in one sample but retained nothing through 40k iterations in the next, so a one-off heap step must not read as retained pins.
- Add a macOS badge to `README.md`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `test-harness-integrity`: the gate's log parsing must strip colour codes identically on every supported CI platform, so a clean run on BSD `sed` is judged clean rather than failing as "no driver summary".

## Impact

- New: `.github/workflows/ci_macos.yaml`, `pkg/util/threadname_darwin.go`.
- Changed: `test/check_test_output.sh`, `test/check_editor_output.sh`, `test/pkg/pin_flatness_tests.go`, `README.md`.
- CI: two jobs per PR that touches Go/C/header/GDExtension files (free on this public repository).
- No C, generated, or API changes.

## Non-goals

- Fixing macOS shutdown/unload behaviour if the first run exposes a fault (the Windows module-pin problem). That would be its own change.
- Producing a real `.framework` bundle layout; the platform keeps dlopen-ing the plain Mach-O file Godot already resolves.
- macOS packaging, signing, notarization, or release artifacts.
- Source-line (`file:line`) stacktrace evidence on macOS.
- Changing the Linux or Windows workflows beyond the shared gate-script fix.
- Adding the new check to branch protection or required status checks.
