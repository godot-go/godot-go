# Design

## Context

The site currently ships a pixel-arcade theme (game, konami, pixel landscape/mecha). See `proposal.md` for the pivot motivation. Constraint: keep GitHub Pages branch deploy with no build step; brand SVGs must be vendored locally (no hotlinking).

## Goals / Non-Goals

**Goals:**
- Futuristic dark neon/glass identity with the official Godot + Go logos as the centerpiece
- Subtle, performant ambient animation (grid, glow, particles)
- Clean professional content sections with copyable code
- Reduced-motion friendly, no console errors

**Non-Goals:**
- No game, sound, or easter eggs
- No JS frameworks or bundlers
- No changes to the advocacy copy substance (features/FAQ content carried over)

## Decisions

**D1: Vendored brand SVGs.** `assets/godot-icon.svg` (white/blue robot head), `assets/go-logo-white.svg` (white wordmark), `assets/gopher.svg` (favicon). Referenced via `<img>`/`<link rel="icon">`. Alternatives (hotlinking, hand-drawn logos) rejected for reliability and brand fidelity.

**D2: CSS-driven glass + neon.** Glass = `rgba(255,255,255,0.04)` fill + `backdrop-filter: blur(12px)` + 1px `rgba(255,255,255,0.08)` border. Neon = cyan `#22d3ee` / violet `#8b5cf6` / Godot blue `#478cbf` used in gradients, glows (`box-shadow`/`text-shadow`), and the primary button gradient. Base `#050510`.

**D3: Ambient background = CSS grid + glow orbs + one canvas particle field.** Grid via repeating linear-gradients; glow orbs via blurred radial-gradient divs; particles via a single full-viewport `<canvas>` with ~70 slow-drifting dots and faint link lines (classic network look), paused under reduced motion. Chosen over a JS library (particles.js) to avoid a dependency.

**D4: Typography.** Inter (body/UI) + Space Grotesk (display headings) + JetBrains Mono (code), via Google Fonts with system fallbacks. Replaces the pixel font.

**D5: Copy buttons.** Each code block gets a button; JS writes the command text to `navigator.clipboard` (with `execCommand` fallback) and flips the label to "Copied" for ~1.5s.

**D6: Sticky nav.** `position: sticky; top: 0` with backdrop blur; anchor links use `scroll-behavior: smooth` (auto under reduced motion).

**D7: site.js scope.** Particles canvas + IntersectionObserver reveal + copy buttons + smooth-scroll. All game/konami/sound code removed; `js/game.js` deleted.

## Risks / Trade-offs

- [backdrop-filter unsupported in old browsers] → panels still readable (translucent fill without blur); progressive enhancement.
- [Google Fonts blocked] → system sans/mono fallback keeps layout stable.
- [Particle canvas CPU] → capped particle count, paused when tab hidden and under reduced motion.
- [Clipboard API permissions] → try/catch + execCommand fallback + visible confirmation.

## Migration Plan

1. Add `assets/` brand SVGs; rewrite `index.html`, `css/style.css`, `js/site.js`; delete `js/game.js`.
2. Push → GitHub Pages rebuilds. Rollback: `git revert`.
