# Tasks

## 1. Assets & cleanup

- [x] 1.1 Vendor brand SVGs into `assets/` (godot-icon.svg, go-logo-white.svg, gopher.svg) and verify each file exists and is valid SVG
- [x] 1.2 Delete `js/game.js` and confirm no references to it remain in `index.html`

## 2. Futuristic theme (css/style.css)

- [x] 2.1 Define the dark neon/glass design tokens (base #050510, cyan/violet/godot-blue accents, glass fill/blur/border, Inter + Space Grotesk + JetBrains Mono) and base element styles; verify dark theme renders with glass panels
- [x] 2.2 Style the ambient background layers (CSS grid, glow orbs, particle canvas container) behind content; verify grid + glow visible and behind all content
- [x] 2.3 Style the sticky nav, hero logo lockup (Godot icon + Go wordmark with glow), glass feature cards with line icons, code blocks with copy buttons, FAQ accordion, and footer; verify each section renders in the new theme
- [x] 2.4 Add `prefers-reduced-motion` overrides disabling particles/glow/blur animations with content fully visible; verify under emulated reduced-motion

## 3. Page structure (index.html)

- [x] 3.1 Build sticky nav (Godot icon + godot-go brand, anchors: Features / Get Started / FAQ / GitHub) and favicon link to `assets/gopher.svg`; verify nav sticky and favicon set
- [x] 3.2 Build hero with the Godot icon + Go wordmark lockup, tagline, description, and CTA buttons; verify both logos visible with glow
- [x] 3.3 Build Features section as 6 glass cards with inline line-style SVG icons; verify all six render
- [x] 3.4 Build Get Started + Build & Test sections as code blocks (install, `make generate && make build`, `make test`) each with a copy button; verify commands copyable
- [x] 3.5 Build FAQ accordion (≥4 Q&A) and footer (GitHub, Discord, credits); verify expand/collapse and links

## 4. Interactivity (js/site.js)

- [x] 4.1 Implement the particle canvas (slow-drifting dots + faint link lines, capped count, paused when tab hidden and under reduced motion); verify it animates and pauses correctly
- [x] 4.2 Implement IntersectionObserver scroll-reveal for sections; verify one-time reveal
- [x] 4.3 Implement copy-to-clipboard for code blocks with "Copied" confirmation and execCommand fallback; verify copy works
- [x] 4.4 Ensure NO game/konami/sound code remains; verify by grepping site.js for game/konami/audio symbols

## 5. Integration & verification

- [x] 5.1 Full browser pass: desktop + mobile viewport, reduced-motion, no console errors; logos render, nav scrolls, copy works, FAQ toggles; record results
- [x] 5.2 Verify Liquid-sequence safety (no `{{`/`{%`) and that the page serves correctly from a local static server; preview for user sign-off
