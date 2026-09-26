# Spec Delta

## Purpose

Defines the site-wide pixel-art aesthetic — palette, typography, hard-edged rendering, and animated background — so the site presents a cohesive retro arcade identity on every screen size.

## ADDED Requirements

### Requirement: Hard-edged pixel rendering
All decorative visual elements (borders, shadows, sprites, canvas output) SHALL render with square, non-anti-aliased pixel edges at any display scale.

#### Scenario: Pixel edges stay crisp when zoomed
- **WHEN** a visitor zooms the browser to 400%
- **THEN** borders, sprites, and canvas art still show square pixel edges with no smoothing or blurring

### Requirement: Arcade palette and display typography
The site SHALL use a 16-bit arcade palette (deep navy base with high-contrast neon accents) and a pixel display font for headings and UI labels, with a monospace fallback.

#### Scenario: Fonts fail to load
- **WHEN** the external pixel webfont is unavailable (offline or blocked)
- **THEN** the page still renders fully readable text using the monospace fallback and the layout does not break

### Requirement: Parallax landscape background
The page background SHALL form a pixel-art landscape of at least three depth layers (e.g., sky with stars, distant mountains, hills, foreground strip) that move at different apparent rates on scroll, creating depth. A giant steampunk mecha robot SHALL stand as a landmark within the landscape, positioned behind the foreground hills, with an animated glowing furnace and rising steam from its smokestacks.

#### Scenario: Scrolling reveals depth
- **WHEN** the visitor scrolls the page
- **THEN** background layers move at visibly different rates relative to the content

#### Scenario: Mecha landmark present
- **WHEN** the page renders with JavaScript enabled
- **THEN** a large steampunk mecha is visible in the background with a pulsing furnace glow and animated steam puffs

#### Scenario: Mecha respects reduced motion
- **WHEN** the visitor has `prefers-reduced-motion: reduce` set
- **THEN** the mecha's steam and furnace animations are disabled while the mecha remains visible

### Requirement: Reduced-motion respect
When the operating system reports a reduced-motion preference, the site SHALL disable parallax, blinking, and decorative animation while keeping all content fully visible and usable.

#### Scenario: Reduced motion enabled
- **WHEN** a visitor has `prefers-reduced-motion: reduce` set
- **THEN** no parallax, blinking, or decorative looping animation plays, and all sections are immediately visible
