# Tasks

## 1. Get Started content (index.html)

- [x] 1.1 Replace the one-line prerequisite note with OS-tagged prerequisite list rows (glass cards) covering: Go ≥1.26, gcc via build-essential (Linux), mingw-w64 gcc (Windows), Xcode Command Line Tools (macOS), GNU make (all; + bash on Windows), Godot 4.7-stable (on `PATH` or `GODOT` env var); verify all rows render with correct OS tags
- [x] 1.2 Add optional entries for clang-format (formats generated C/H bindings; build skips silently if absent) and goimports (tidies generated Go; `make installdeps`), each visibly tagged "optional" and all-platform; verify tags render
- [x] 1.3 Add per-OS install code blocks above the existing `go get` block, each with the existing copy-button component: `sudo apt-get install -y build-essential clang-format` (Debian/Ubuntu), `scoop install wget unzip` (Windows, with inline note that mingw-w64 gcc/make come with the dev environment), `xcode-select --install` (macOS); verify each exact command copies with confirmation
- [x] 1.4 Add the macOS disclosure note "Makefile-derived · no upstream CI" on the macOS prerequisite row and install block; verify it renders
- [x] 1.5 Update the section sub-line to note Linux/Windows versions match the project's CI (Go 1.26.x, Godot 4.7-stable) and macOS follows the Makefile; verify text renders

## 3. Tabbed restructure

- [x] 3.1 Replace the six-card grid with a single tabbed section: ARIA tablist (Linux/Windows/macOS), one tabpanel per OS with that OS's prerequisite list and install code block; shared `go get` block below the tabs
- [x] 3.2 Add tab-switching JS with keyboard support (Arrow/Home/End) and visitor-OS auto-select via `navigator.userAgentData`/`platform`
- [x] 3.3 Add tab styling consistent with the theme (pill tabs, active gradient state)
- [x] 3.4 Browser pass: tab switching shows correct panels, keyboard nav works, auto-select works, copy buttons work in each panel, desktop + mobile, no console errors
- [x] 3.5 macOS panel: lead with Homebrew block (`brew install go clang-format` + `brew install --cask godot`, versions verified against formulae.brew.sh: go 1.27.1, clang-format 23.1.2, godot cask 4.7.2 with PATH symlink); keep `xcode-select --install` as one-time block with note that brew cannot install CLT; verify copy button copies both brew commands

## 4. Verification

- [x] 2.1 Browser pass at desktop and mobile widths: prerequisite list with OS tags, optional tags, all four code blocks with working copy buttons, macOS disclosure note; no console errors
- [x] 2.2 Version parity check against upstream `ci_linux.yaml` and `ci_windows.yaml` (Go 1.26.5 → "1.26+", Godot 4.7-stable) and `Makefile` (darwin target, clang-format/goimports optional); record the comparison for all three platforms
- [x] 2.3 Liquid-sequence safety (no `{{`/`{%`) and local static-server render check
