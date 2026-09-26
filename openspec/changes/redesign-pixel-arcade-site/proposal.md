# Proposal

## Why

The current GitHub Pages site is the default `jekyll-theme-cayman` rendering of the README: generic, static, and forgettable. godot-go sits at the intersection of two communities that love playful, visual tooling (Godot) and pragmatic language advocacy (Go). A pixel-art arcade-themed site makes the project memorable, demonstrates that "Go can build games" viscerally, and re-frames the content as an advocacy page aimed at Go developers who want to ship real games.

## What Changes

- **BREAKING** Replace the Jekyll-theme-rendered `index.md` with a hand-crafted static site (`index.html` + `css/` + `js/`) served directly by GitHub Pages; remove the `theme:` and `readme_index` config from `_config.yml`.
- Add a full pixel-art visual theme: 16-bit arcade palette, "Press Start 2P" display font, hard pixel borders/shadows, `image-rendering: pixelated`, parallax star/cloud background.
- Add a playable hero mini-game ("Gopher Run"): canvas endless-runner, arrow-key/tap controls, collectibles, hazards, score + best score persisted in `localStorage`.
- Add site-wide interactivity: PRESS START scroll CTA, WebAudio blips with a sound toggle (muted by default), scroll-reveal section transitions, hover/click pixel effects, Konami-code rainbow-mode easter egg, self-typing fake terminal.
- Re-imagine all page content as a Go-advocacy narrative: hero title screen, "Why Go + Godot?" character-select cards, Level 1 Getting Started, Level 2 Build & Test, Boss-level FAQ, Continue-screen footer (GitHub, Discord, credits). The original page's reference link list is dropped.

## Capabilities

### New Capabilities

- `pixel-visual-theme`: The site-wide pixel-art aesthetic — palette, typography, borders/shadows, pixelated rendering, parallax background, and reduced-motion/degradation behavior.
- `hero-minigame`: The playable "Gopher Run" canvas game — controls, entities, scoring, game-over/restart, focus and visibility rules.
- `site-interactivity`: Interactive behaviors outside the game — PRESS START, sound toggle, scroll-reveal, hover effects, Konami code, typing terminal.
- `advocacy-content`: The page's information architecture and copy — hero, benefit cards, getting-started/build/test levels, FAQ, footer links.

### Modified Capabilities

_(none — no existing specs in this project)_

## Impact

- `index.md` removed; new `index.html`, `css/style.css`, `js/game.js`, `js/site.js` added at repo root.
- `_config.yml` stripped of theme/readme-index settings (Jekyll passes static files through untouched).
- External dependencies limited to Google Fonts (with system fallbacks); no build step, no JS frameworks, no image binaries (art is CSS/canvas/SVG).
- GitHub Pages deployment path unchanged (branch deploy continues to work).
