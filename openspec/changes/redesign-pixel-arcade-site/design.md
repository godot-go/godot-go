# Design

## Context

The repo is a GitHub Pages branch-deployed site: `_config.yml` (jekyll-theme-cayman + readme-index plugin) and `index.md` (README content). See `proposal.md` for motivation. Constraints: GitHub Pages rebuilds on push; we want zero build tooling, zero binary image assets, and the existing Pages deployment path untouched.

## Goals / Non-Goals

**Goals:**
- Single-page static site with full pixel-art control (no theme fighting us)
- A fun but small hero game that never overshadows the content
- All art generated from code (CSS box-shadow sprites, inline SVG, canvas drawing)
- Works without JS (readable), degrades under reduced-motion, no console errors

**Non-Goals:**
- Multi-page docs site, blog, or API reference
- Audio files, image binaries, JS frameworks, bundlers, or CI changes
- Cross-browser pixel-perfection in legacy browsers (modern evergreen only)

## Decisions

**D1: Hand-crafted static files over Jekyll theme or JS framework.**
`index.html` + `css/style.css` + `js/game.js` + `js/site.js` at repo root. Jekyll (which Pages runs by default) copies files without front matter verbatim, so a minimal `_config.yml` keeps Pages happy while giving us raw HTML control. Alternatives rejected: custom Jekyll theme (friction with canvas/pixel CSS for zero benefit), React/Vue (build toolchain overkill for one page).
- Guard: avoid Liquid trigger sequences (`{{`, `{%`) in HTML/JS/CSS so Jekyll never interprets them. JS template literals use `${...}` and are safe.

**D2: Canvas mini-game with fixed-timestep loop.**
Logical resolution 640×220 upscaled with `image-rendering: pixelated`. `requestAnimationFrame` drives an accumulator stepping physics at fixed 60Hz ticks; gravity + jump velocity; AABB collision for collectibles/hazards/ground. Deterministic step makes collision behavior inspectable and testable. All sprites drawn programmatically (rect primitives) — no sprite sheets. ~300–400 lines vanilla JS.

**D3: Focus/visibility discipline for the game.**
`IntersectionObserver` starts/stops the loop with canvas visibility; `visibilitychange` pauses on tab blur. Keyboard listeners attach to the canvas element (with `tabindex="0"`), not `window`, so arrow keys never hijack page scroll unless the game is focused.

**D4: WebAudio-synthesized blips, no audio files.**
Tiny oscillator envelopes (square wave, ~50–120ms) per event type. `AudioContext` created lazily on first user gesture (autoplay policy). Muted by default; toggle persists in `localStorage` key `godotgo.sound`. All `localStorage` access wrapped in try/catch (private-mode safety).

**D5: Pixel art via CSS, not images.**
Borders/shadows: layered `box-shadow` at 4px steps (hard edges, no blur). Sprites in hero/cards: inline SVG with `shape-rendering="crispEdges"` or CSS box-shadow pixel sprites. Canvas: `imageSmoothingEnabled = false`.

**D6: Reveal/typing via IntersectionObserver + CSS classes.**
One observer adds `.revealed` once; CSS handles the pixel-dissolve (stepped `clip-path`/opacity keyframes). Terminal typing is a small JS char-by-char timer triggered by the same observer. Under `prefers-reduced-motion`, CSS disables animations and the terminal renders full text immediately (JS checks the media query before animating).

**D7: Konami listener in `site.js`.**
Window-level keydown buffer matching ↑↑↓↓←→←→BA; on match adds `body.rainbow` (CSS hue-cycling animation on accent custom properties) + a short canvas-free particle burst of absolutely-positioned pixel divs.

**D8: Fonts.**
"Press Start 2P" via Google Fonts `<link>` with `font-display: swap`; fallback stack `"Courier New", monospace`. Body copy in monospace for readability; pixel font reserved for headings/UI to limit line-height pain.

## Risks / Trade-offs

- [Jekyll mangles a file via Liquid] → Avoid `{{`/`{%` sequences; verify rendered output locally with `jekyll build` if available, else inspect served output after deploy.
- [Google Fonts blocked/slow] → `font-display: swap` + monospace fallback keeps layout stable.
- [Canvas perf on low-end mobile] → Small logical resolution, capped entity count, loop paused when off-screen/unfocused.
- [localStorage unavailable] → try/catch wrappers; game still plays, just without persisted BEST/sound.
- [Game overshadows advocacy] → Game confined to hero height budget (~220px logical); content sections are the bulk of the page.
- [Pixel font hurts readability] → Pixel font only for headings/labels ≤ 2 sizes; paragraphs in monospace.

## Migration Plan

1. Add new files (`index.html`, `css/`, `js/`), update `_config.yml`, delete `index.md` in one commit.
2. Push → GitHub Pages rebuilds automatically (no pipeline changes).
3. Rollback: `git revert` the commit restores the Cayman/README site exactly.
