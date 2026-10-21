# Proposal

## Why

macOS is already a first-class build target in-tree — the Makefile has a `darwin` branch, `test/demo/example.gdextension` maps both macOS architectures, and `pkg/log/stacktrace.c` uses the native execinfo path on Apple — but no workflow ever runs it. Linux and Windows are covered; macOS regressions (build, extension load, teardown, capture probe) stay invisible until a user hits them. A macOS run today would also fail spuriously: both gate scripts strip ANSI with GNU-only `sed` escapes, so BSD `sed` leaves the colour codes in place and the summary banner never parses.

## What Changes

- Add `.github/workflows/ci_macos.yaml`, a two-leg matrix on native runners:
  - `macos-latest` (macOS 26, arm64) with `GOARCH=arm64`
  - `macos-26-intel` (macOS 26, x64) with `GOARCH=amd64`
  - `fail-fast: false`, so one architecture's failure does not hide the other's result.
- Each leg mirrors the Windows job: Godot 4.7 stable install, `update_godot_headers_from_binary`, `make generate`, `make build`, `make ci_gen_test_project_files`, `make test`, capture evidence, artifact upload. Artifacts carry the architecture in the name (`macos-<arch>-test-lib`) to avoid collisions.
- Capture evidence is scoped to what macOS can prove: `stacktrace-probe capture=1` and the `___ C start backtrace ___` banner. The Windows `file:line` assertion is **not** ported — execinfo prints `backtrace_symbols` output, not DWARF source lines.
- Make the ANSI strip in `test/check_test_output.sh` and `test/check_editor_output.sh` POSIX-portable (emit a literal ESC byte instead of GNU `\x1b`), so the same gate semantics hold under BSD and GNU `sed`.
- Add a macOS badge to `README.md` alongside the existing Linux and Windows badges.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `test-harness-integrity`: the gate's log parsing must strip colour codes identically on every supported CI platform, so a clean run on BSD `sed` is judged clean rather than failing as "no driver summary".

## Impact

- New: `.github/workflows/ci_macos.yaml`.
- Changed: `test/check_test_output.sh`, `test/check_editor_output.sh`, `README.md`.
- CI: one new workflow, two jobs per PR that touches Go/C/header/GDExtension files (macOS runners are free on this public repository).
- No Go, C, generated, or API changes.

## Non-goals

- Fixing macOS shutdown/unload behaviour if the first run exposes a fault (the Windows module-pin problem). That would be its own change.
- Producing a real `.framework` bundle layout; the platform keeps dlopen-ing the plain Mach-O file Godot already resolves.
- macOS packaging, signing, notarization, or release artifacts.
- Source-line (`file:line`) stacktrace evidence on macOS.
- Changing the Linux or Windows workflows beyond the shared gate-script fix.
- Adding the new check to branch protection or required status checks.
