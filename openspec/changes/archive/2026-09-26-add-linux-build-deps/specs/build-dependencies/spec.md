# Spec Delta

## Purpose

Publishes the build-dependency information for godot-go on the site — the toolchain and engine versions the project's GitHub Actions CI validates on Linux and Windows, and the Makefile implies on macOS — so developers can prepare a working build environment in one pass on each platform.

## ADDED Requirements

### Requirement: Complete Linux prerequisite list
The Get Started section SHALL list every hard build dependency for Linux as validated by the upstream `ci_linux.yaml` workflow: Go 1.26 or newer, a C compiler (gcc via `build-essential`), GNU make, and a Godot 4.7-stable binary available on `PATH` or via the `GODOT` environment variable.

#### Scenario: All hard dependencies visible
- **WHEN** a visitor opens the Get Started section on Linux
- **THEN** Go (≥1.26), gcc/build-essential, GNU make, and Godot 4.7-stable are each listed as prerequisites

#### Scenario: Godot binary discovery documented
- **WHEN** a visitor reads the Godot prerequisite
- **THEN** it states the binary must be on `PATH` or pointed to by the `GODOT` environment variable

### Requirement: Complete Windows prerequisite list
The Get Started section SHALL list every hard build dependency for Windows as validated by the upstream `ci_windows.yaml` workflow: Go 1.26 or newer, a C compiler (mingw-w64 gcc), GNU make with a bash shell (CI targets run `shell: bash`), and a Godot 4.7-stable win64 binary pointed to by the `GODOT` environment variable.

#### Scenario: Windows dependencies visible
- **WHEN** a visitor opens the Get Started section for Windows
- **THEN** Go (≥1.26), mingw-w64 gcc, GNU make + bash, and Godot 4.7-stable win64 are each listed as prerequisites

### Requirement: macOS prerequisites with honest verification status
The Get Started section SHALL list the macOS build dependencies derived from the upstream `Makefile` (darwin framework build target): Go 1.26 or newer, Xcode Command Line Tools (clang for cgo), GNU make, and a Godot 4.7-stable macOS build. The section SHALL state that macOS dependencies are Makefile-derived and not validated by an upstream CI workflow.

#### Scenario: macOS dependencies visible
- **WHEN** a visitor opens the Get Started section for macOS
- **THEN** Go (≥1.26), Xcode Command Line Tools, GNU make, and Godot 4.7-stable macOS are each listed as prerequisites

#### Scenario: Verification status disclosed
- **WHEN** a visitor reads the macOS prerequisites
- **THEN** the section clearly notes there is no upstream macOS CI and the list comes from the Makefile

### Requirement: Optional tools marked optional
The Get Started section SHALL present `clang-format` and `goimports` as optional on all platforms, each with its role: `clang-format` formats the generated C/H bindings (the build skips it silently when absent) and `goimports` tidies generated Go sources (installed via `make installdeps`).

#### Scenario: Optional tools labeled
- **WHEN** a visitor reads the prerequisite lists
- **THEN** clang-format and goimports are explicitly marked optional and the build is not described as failing without them

### Requirement: Copyable install commands per platform
The Get Started section SHALL provide copyable command blocks covering the package-manager-installable dependencies for each platform — `sudo apt-get install -y build-essential clang-format` (Debian/Ubuntu), `scoop install wget unzip` (Windows), and for macOS the Homebrew commands `brew install go clang-format` plus `brew install --cask godot` — alongside the one-time `xcode-select --install` for the Xcode Command Line Tools, which Homebrew cannot install. Each block SHALL use the existing copy-button component.

#### Scenario: Copy the apt command
- **WHEN** a visitor clicks the copy button on the apt install block
- **THEN** the exact `sudo apt-get install -y build-essential clang-format` command is copied to the clipboard with visible confirmation

#### Scenario: Copy the macOS Homebrew commands
- **WHEN** a visitor clicks the copy button on the macOS Homebrew block
- **THEN** the exact `brew install go clang-format` and `brew install --cask godot` commands are copied to the clipboard with visible confirmation

### Requirement: Tabbed per-OS presentation
The Get Started section SHALL present all prerequisites in a single section with one tab per operating system (Linux, Windows, macOS). Selecting a tab SHALL show only that OS's prerequisites and install command; the visitor's own OS SHALL be auto-selected on load when detectable; tabs SHALL be keyboard-navigable (arrow keys, Home, End) per the ARIA tabs pattern.

#### Scenario: Switch tabs
- **WHEN** a visitor selects the Windows tab
- **THEN** only the Windows prerequisites and install command are visible, and the Linux and macOS panels are hidden

#### Scenario: Visitor OS auto-selected
- **WHEN** a visitor whose browser reports macOS loads the page
- **THEN** the macOS tab is selected by default

### Requirement: Versions match upstream CI
The Linux and Windows dependency versions shown on the site SHALL match the versions pinned in the upstream `godot-go` CI workflows (`ci_linux.yaml` and `ci_windows.yaml`, both pinning Go 1.26.5 and Godot 4.7-stable) at the time of publishing. The macOS versions SHALL match the engine version the Makefile targets.

#### Scenario: Version parity check
- **WHEN** the published prerequisite versions are compared against `ci_linux.yaml` and `ci_windows.yaml`
- **THEN** the Go and Godot versions listed on the site equal the CI-pinned versions
