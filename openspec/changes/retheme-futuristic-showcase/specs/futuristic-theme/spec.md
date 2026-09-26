# Spec Delta

## Purpose

Defines the futuristic professional visual system for the godot-go site — a dark neon/glass aesthetic with an animated background and a prominent showcase of the official Godot and Go logos — replacing the prior pixel-arcade presentation.

## ADDED Requirements

### Requirement: Dark neon/glass visual system
The site SHALL use a dark near-black base with cyan and violet neon accents, glassmorphism panels (translucent fill, backdrop blur, hairline borders), and a modern sans-serif typeface with a monospace face for code.

#### Scenario: Dark theme renders
- **WHEN** the page loads
- **THEN** the background is dark near-black, panels are translucent glass with subtle borders, and accent colors are cyan/violet

### Requirement: Official logo showcase
The hero SHALL prominently display the official Godot icon and the official Go wordmark logo as a lockup with glow treatment. The Godot icon SHALL also appear in the navigation bar, and the Go gopher SHALL be the site favicon. All brand SVGs are served from the local `assets/` directory.

#### Scenario: Logos visible in hero
- **WHEN** a visitor lands on the page
- **THEN** the Godot icon and the Go wordmark are both clearly visible in the hero, each with a glow effect

#### Scenario: Brand assets load locally
- **WHEN** the page requests brand imagery
- **THEN** the SVGs are served from `assets/` (no hotlinking to external hosts)

### Requirement: Animated ambient background
The page SHALL have a subtle animated background composed of a grid, soft glow orbs, and a slow-drifting particle field that does not distract from content.

#### Scenario: Background animates subtly
- **WHEN** the visitor views the page
- **THEN** the grid, glow, and particles animate slowly and remain behind all content

### Requirement: Glass feature cards and copyable code
Benefits SHALL be presented as glass cards with line-style icons. Install/build/test commands SHALL appear in styled code blocks, each with a copy-to-clipboard button that copies the exact command.

#### Scenario: Copy a command
- **WHEN** a visitor clicks a code block's copy button
- **THEN** the exact command text is copied to the clipboard and the button shows brief confirmation

### Requirement: Sticky navigation
The page SHALL have a sticky top navigation bar with anchor links (Features, Get Started, FAQ, GitHub) that smooth-scroll to their sections.

#### Scenario: Nav smooth-scrolls
- **WHEN** a visitor clicks a nav link
- **THEN** the page smooth-scrolls to the corresponding section

### Requirement: Arcade elements removed
The site SHALL NOT include the mini-game, Konami code, PRESS START control, sound effects, sound toggle, pixel font, or pixel landscape/mecha background.

#### Scenario: No arcade artifacts
- **WHEN** the page loads with JavaScript enabled
- **THEN** no game canvas, sound toggle, or Konami handler is present

### Requirement: Reduced-motion respect
When the visitor prefers reduced motion, the particle field and glow animations SHALL be disabled while all content remains fully visible and usable.

#### Scenario: Reduced motion enabled
- **WHEN** a visitor has `prefers-reduced-motion: reduce` set
- **THEN** particles and glow animations are off and all content is immediately visible
