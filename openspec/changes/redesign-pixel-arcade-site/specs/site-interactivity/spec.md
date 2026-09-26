# Spec Delta

## Purpose

Specifies the interactive behaviors outside the mini-game — the PRESS START CTA, sound system, scroll reveals, hover feedback, easter egg, and typing terminal — that make the page respond to the visitor.

## ADDED Requirements

### Requirement: PRESS START call to action
The hero SHALL show a blinking "PRESS START" control; activating it (click, tap, or Enter) SHALL smooth-scroll the page to the first content section.

#### Scenario: Activate PRESS START
- **WHEN** the visitor clicks or taps PRESS START, or presses Enter while it is focused
- **THEN** the page smooth-scrolls to the first content section below the hero

### Requirement: Sound effects with user-controlled mute
The site SHALL generate short retro blip sounds for jump, collect, click, and game-over events. Sound SHALL be muted by default, audio SHALL initialize only after a user gesture, and a visible pixel sound toggle SHALL switch sound on/off with the choice persisting across sessions.

#### Scenario: First visit is silent
- **WHEN** a first-time visitor interacts with buttons or the game
- **THEN** no sound plays because the default state is muted

#### Scenario: Toggle persists
- **WHEN** the visitor enables sound via the toggle and later reloads the page
- **THEN** sound remains enabled without requiring another gesture beyond the browser's autoplay policy

### Requirement: Scroll-reveal sections
Content sections SHALL animate into view (pixel-dissolve/slide) the first time they enter the viewport, and SHALL remain visible thereafter.

#### Scenario: Section revealed once
- **WHEN** a section scrolls into view for the first time
- **THEN** it plays its reveal animation; scrolling away and back does not replay it

### Requirement: Hover and click feedback
Interactive elements SHALL give immediate pixel-style feedback: buttons visually depress on click, cards lift with a hard shadow shift on hover, and links flicker color on hover.

#### Scenario: Button press feedback
- **WHEN** the visitor presses a button
- **THEN** the button visibly depresses for the duration of the press

### Requirement: Konami code easter egg
Entering the sequence ↑ ↑ ↓ ↓ ← → ← → B A anywhere on the page SHALL activate a rainbow mode that cycles the accent palette and emits a burst of pixel particles.

#### Scenario: Secret sequence entered
- **WHEN** the visitor enters the Konami sequence
- **THEN** rainbow mode activates with cycling colors and a pixel particle burst

### Requirement: Self-typing terminal
Code terminal blocks SHALL type out their command text with a blinking block cursor when scrolled into view; under a reduced-motion preference the full text SHALL appear immediately without typing animation.

#### Scenario: Terminal types on reveal
- **WHEN** a terminal block first scrolls into view
- **THEN** its command text types out character by character ending with a blinking cursor
