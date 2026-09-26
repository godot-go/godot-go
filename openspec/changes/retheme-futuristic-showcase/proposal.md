# Proposal

## Why

The pixel-arcade theme, while playful, doesn't read as a credible home for a developer-facing binding library. The project should present godot-go as a serious, modern tool. A futuristic, professional dark-neon/glass aesthetic — built around the official Godot and Go logos — showcases the two brands the project bridges and signals quality to evaluating developers.

## What Changes

- **BREAKING** Remove all arcade elements: the Gopher Run mini-game (`js/game.js`), Konami code, PRESS START, sound blips/toggle, pixel font, and the pixel landscape/mecha background.
- Replace the visual system with a futuristic dark theme: deep near-black base, cyan + violet neon accents, glassmorphism panels, an animated grid + glow background, and a subtle particle field.
- Showcase the **official Godot and Go logos** prominently: a hero logo lockup (Godot icon + Go wordmark) with glow treatments, plus the Godot icon in the nav and the gopher as the favicon. Brand SVGs vendored under `assets/`.
- Keep the advocacy content (features, getting started, build & test, FAQ, footer) but restyle it as glass cards and modern code blocks with copy buttons.
- Add a sticky top navigation bar with smooth-scroll anchors.

## Capabilities

### New Capabilities

- `futuristic-theme`: The futuristic professional visual system — dark neon/glass palette, typography, animated grid/glow/particle background, official logo showcase, glass cards, code blocks with copy, sticky nav, and reduced-motion behavior.

### Modified Capabilities

_(none — the prior pixel-arcade change was never archived; its `pixel-visual-theme`, `hero-minigame`, and `site-interactivity` capabilities are superseded and dropped by this change. The `advocacy-content` capability is retained unchanged.)_

## Impact

- `js/game.js` deleted; `js/site.js` rewritten (particles, reveal, copy buttons, smooth scroll — no game/konami/sound).
- `index.html` rewritten (nav, logo hero, glass sections); `css/style.css` rewritten (futuristic theme).
- New vendored brand assets: `assets/godot-icon.svg`, `assets/go-logo-white.svg`, `assets/gopher.svg`.
- Fonts switch from "Press Start 2P" to a modern sans (Inter/Space Grotesk) + JetBrains Mono for code.
- GitHub Pages deployment path unchanged.
