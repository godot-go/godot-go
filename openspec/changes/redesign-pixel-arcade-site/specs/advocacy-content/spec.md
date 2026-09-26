# Spec Delta

## Purpose

Defines the page's information architecture and copy: a Go-advocacy narrative that persuades Go developers to try godot-go, organized as game screens from title screen through boss FAQ to the continue screen.

## ADDED Requirements

### Requirement: Hero title screen
The hero SHALL present the project name, the tagline "Write your game in Go. Ship it with Godot.", a one-line description of godot-go (Go bindings for Godot 4 via GDExtension/cgo), the mini-game, and the PRESS START CTA.

#### Scenario: First impression
- **WHEN** a visitor lands on the page
- **THEN** they see the project name, tagline, one-line description, playable game, and PRESS START without scrolling

### Requirement: Benefits showcase
The page SHALL present at least six Go-advocacy benefit cards styled as character-select options, covering: static typing, goroutines/concurrency, fast compile times, single-binary cross-platform shipping, game-friendly garbage collection, and standard library/tooling. Each card SHALL have a pixel icon and a short punchy blurb.

#### Scenario: Benefits visible as selectable cards
- **WHEN** a visitor reaches the benefits section
- **THEN** six or more benefit cards are displayed, each with a pixel icon and headline plus blurb

### Requirement: Getting-started level
The page SHALL present requirements (clang-format, gcc) as "party members" and the install command `go get github.com/godot-go/godot-go` inside a pixel terminal block.

#### Scenario: Install path is clear
- **WHEN** a convinced developer reaches the getting-started section
- **THEN** they can see the prerequisites and copy the exact install command from the terminal block

### Requirement: Build and test level
The page SHALL present the build command `make generate && make build` and the test command `make test` in terminal style, each with a one-line explanation.

#### Scenario: Build commands copyable
- **WHEN** a developer reaches the build-and-test section
- **THEN** both commands are selectable/copyable text with brief explanations

### Requirement: Boss-level FAQ
The page SHALL present a FAQ as an accordion of at least four common developer concerns (e.g., runtime performance, GC pauses, Godot version support, API stability), each expandable to reveal its answer.

#### Scenario: Expand an answer
- **WHEN** a visitor clicks a FAQ question
- **THEN** the answer expands in place; clicking again collapses it

### Requirement: Continue-screen footer
The footer SHALL link to the GitHub repository, the Godot Engine Discord invite, and credits (ShadowApex's godot-go, godot-cpp), styled as an arcade "continue screen".

#### Scenario: Find the project links
- **WHEN** a visitor scrolls to the footer
- **THEN** GitHub, Discord, and credit links are present and functional

### Requirement: Content accuracy
All commands, links, and project claims on the page SHALL match the actual godot-go project as reflected in the original `index.md` content being replaced.

#### Scenario: No stale claims
- **WHEN** the new page is compared against the original index.md
- **THEN** every command, URL, and factual claim carried over is unchanged in substance
