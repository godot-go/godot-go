# Spec Delta

## Purpose

Specifies the playable "Gopher Run" endless-runner embedded in the hero section, so visitors can experience a real Go-and-Godot-flavored game without leaving the page.

## ADDED Requirements

### Requirement: Desktop keyboard controls
While the game canvas has focus, the game SHALL respond to arrow keys for movement and Up/Space for jumping, and SHALL NOT capture arrow keys when the canvas does not have focus.

#### Scenario: Jump with keyboard
- **WHEN** the canvas is focused and the player presses Up or Space
- **THEN** the gopher jumps

#### Scenario: Page scroll not hijacked
- **WHEN** the canvas is not focused and the visitor presses the arrow keys
- **THEN** the page scrolls normally and the game ignores the input

### Requirement: Touch controls
On touch devices, tapping the game canvas SHALL make the gopher jump.

#### Scenario: Tap to jump
- **WHEN** a touch visitor taps the canvas
- **THEN** the gopher jumps

### Requirement: Collectibles and scoring
The game SHALL spawn collectible items — "Go" gems worth 10 points and Godot icons worth 50 points — and display the current score in a HUD during play.

#### Scenario: Collect a gem
- **WHEN** the gopher overlaps a Go gem
- **THEN** the gem disappears, the score increases by 10, and a collect effect plays

### Requirement: Hazards and game over
The game SHALL include hazards (spike crates and pits); colliding with one ends the run and displays a "GAME OVER" screen offering restart.

#### Scenario: Hit a hazard
- **WHEN** the gopher collides with a spike crate or falls into a pit
- **THEN** the run ends and a GAME OVER screen with a restart prompt is shown

#### Scenario: Restart
- **WHEN** the GAME OVER screen is showing and the player presses Enter or taps the canvas
- **THEN** a new run starts with score reset to zero

### Requirement: Best score persistence
The game SHALL record the highest score achieved and display it as BEST, persisting across browser sessions.

#### Scenario: Best survives reload
- **WHEN** a player scored 120 in a previous session and reloads the page
- **THEN** the HUD shows BEST: 120 (or higher if beaten since)

### Requirement: Visibility and focus awareness
The game loop SHALL run only while the canvas is visible in the viewport and the browser tab is focused, and SHALL pause otherwise.

#### Scenario: Scroll away pauses the game
- **WHEN** the canvas is scrolled out of the viewport
- **THEN** the game loop stops consuming CPU and resumes when the canvas becomes visible again

### Requirement: No-JS fallback
Without JavaScript, the hero area SHALL display a static pixel frame (e.g., "INSERT COIN") instead of a broken or blank canvas.

#### Scenario: JavaScript disabled
- **WHEN** a visitor loads the page with JavaScript disabled
- **THEN** the hero shows a static pixel-art frame and the rest of the page remains readable
