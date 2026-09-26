# Design

## Context

See `proposal.md` — Why. The current Get Started section is a single `<p>` prerequisite note plus one code block. The site is a hand-built static page (no build step) that must stay Liquid-safe for GitHub Pages. Upstream CI covers Linux and Windows only; macOS support exists in the `Makefile` (darwin `.framework` build target) but has no CI workflow.

## Goals / Non-Goals

**Goals:**
- Present the CI-validated (Linux, Windows) and Makefile-derived (macOS) dependency sets in the existing visual language without new CSS.
- Make the mapping from site copy to upstream CI verifiable, and make macOS's weaker verification status explicit rather than implied.

**Non-Goals:**
- Proposing or documenting a macOS CI workflow for the upstream repo (out of scope for the site).
- Pinning exact patch versions in prose beyond what CI pins (Go 1.26.x, Godot 4.7-stable).

## Decisions

**D1: Single Get Started section with per-OS tabs.** Prerequisites live in one section behind a tab bar (Linux / Windows / macOS). Each tab panel lists that OS's prerequisites and its install command block; the `go get` block is shared below the tabs. The earlier six-card grid is replaced. Alternative (three stacked OS sections) rejected — forces scrolling through irrelevant platforms.

**D2: Standard ARIA tabs + visitor-OS auto-select.** Tab bar uses `role="tablist"` with `role="tab"` buttons (`aria-selected`, `aria-controls`) and `role="tabpanel"` panels toggled via the `hidden` attribute. Keyboard: ArrowLeft/Right, Home/End move selection. On load, the visitor's OS is auto-selected via `navigator.userAgentData`/`navigator.platform` heuristics (fallback Linux). Install commands stay in the existing `.codeblock` + copy-button component. Alternative (JS-free radio tabs) rejected — no auto-select, clunkier markup.

**D6: macOS installs via Homebrew.** The macOS panel leads with `brew install go clang-format` and `brew install --cask godot` (verified against formulae.brew.sh: `go` 1.27.1 ≥ CI's 1.26 floor; `clang-format` 23.1.2; cask `godot` 4.7.2, which symlinks the `godot` binary onto `PATH` — satisfying the `PATH`/`GODOT` requirement). `xcode-select --install` stays as a separate one-time block because Homebrew cannot install the Xcode Command Line Tools; a note explains brew itself already requires them. Alternative (xcode-select only) rejected — left every other dependency manual on the one platform where a package manager is ubiquitous.

**D3: Optional tools inline with the lists.** clang-format and goimports appear in the prerequisite lists with an "optional" tag and one-line role each, applying to all platforms — keeps one scannable list.

**D4: macOS derivation disclosed.** The macOS prerequisite row and its install block carry a short "Makefile-derived · no upstream CI" note. The alternative — presenting macOS identically to Linux/Windows — would falsely imply CI validation.

**D5: Source annotation.** The section sub-line cites that Linux/Windows versions match the project's CI and macOS follows the Makefile, so future editors know where parity comes from. No machine-readable cross-repo check (out of scope for a static site).

## Risks / Trade-offs

- [Upstream bumps Go/Godot pins and site drifts] — the parity requirement gives a concrete re-check target; a human re-runs the comparison when CI changes.
- [macOS list may be incomplete without CI validation] — disclosed per D4; if upstream adds macOS CI, the note is removed and parity extended.
- [apt/scoop commands assume Debian/Ubuntu and scoop-equipped Windows] — stated explicitly in each block title; other setups map the package names.
