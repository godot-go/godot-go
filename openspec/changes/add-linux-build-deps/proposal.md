# Proposal

## Why

The site's Get Started section lists only "clang-format and gcc" as prerequisites, which is incomplete and misleading: a developer following it will be missing the Go toolchain, GNU make, and a Godot 4 binary — all hard requirements in the project's own CI (`ci_linux.yaml`, `ci_windows.yaml`). Listing the real dependency sets per OS, pinned to the versions CI proves work, removes the first failed-build experience across platforms.

## What Changes

- Replace the one-line prerequisite note in the Get Started section with explicit per-OS dependency lists derived from the `godot-go` CI workflows and `Makefile`:
  - **Linux** (from `ci_linux.yaml`): Go 1.26+ (CI pins 1.26.5), gcc via `build-essential`, GNU make, Godot 4.7-stable binary on `PATH` or via `GODOT` env var
  - **Windows** (from `ci_windows.yaml`): Go 1.26+, a C compiler (mingw-w64 gcc), GNU make + bash (targets run `shell: bash`), Godot 4.7-stable win64 via `GODOT`; CI bootstraps `wget`/`unzip` via scoop
  - **macOS** (derived from `Makefile` — no upstream macOS CI exists): Go 1.26+, Xcode Command Line Tools (clang for cgo), GNU make, Godot 4.7-stable macOS build; the site states this derivation honestly
  - **Optional everywhere**: `clang-format` (formats generated C/H bindings; the build skips it silently if absent) and `goimports` (tidies generated Go; installed via `make installdeps`)
- Add per-OS copyable install commands for the package-manager-installable pieces: `sudo apt-get install -y build-essential clang-format` (Debian/Ubuntu), `scoop install wget unzip` (Windows), `xcode-select --install` (macOS), alongside the existing `go get` block.
- Keep the futuristic theme and page structure unchanged; this is a content correction to the Get Started section.

## Capabilities

### New Capabilities

- `build-dependencies`: The site's published build-dependency information for Linux, Windows, and macOS — the toolchain and engine versions the project's CI validates (or, for macOS, the Makefile implies), presented as prerequisites with copyable install commands.

### Modified Capabilities

_(none — `futuristic-theme` requirements are unchanged; only section content changes.)_

## Impact

- `index.html`: Get Started section rewritten (per-OS prerequisite lists + install code blocks).
- No CSS/JS changes expected; existing code-block and copy-button components are reused.
- Content is grounded in upstream `godot-go` `main`: `ci_linux.yaml` and `ci_windows.yaml` (Go 1.26.5, Godot 4.7-stable) and `Makefile` (cgo/gcc, make, darwin framework target, optional clang-format/goimports).
- macOS coverage is Makefile-derived and labeled as such; no CI parity claim is made for it.
