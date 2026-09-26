# Tasks

## 1. Scaffold & config

- [x] 1.1 Create `index.html` skeleton with semantic landmarks (header/hero, main sections, footer), Google Fonts link for "Press Start 2P" with monospace fallback, and links to `css/style.css`, `js/game.js`, `js/site.js`; verify page loads locally via `python3 -m http.server` with no console errors
- [x] 1.2 Update `_config.yml` to remove `theme:` and `readme_index` settings and delete `index.md`; verify the local server serves the new `index.html` at `/` and no file contains `{{` or `{%` Liquid sequences (grep check)

## 2. Pixel visual theme (css/style.css)

- [x] 2.1 Define CSS custom-property palette (navy base, Go cyan, Godot blue, magenta, lime, gold), base typography (pixel font for headings/UI, monospace body), and 4px hard box-shadow border/button primitives; verify buttons/cards render with crisp hard edges at 400% browser zoom
- [x] 2.2 Build a parallax pixel landscape background (sky gradient with stars + generated mountain/hill/foreground SVG bands) whose layers move at different rates on scroll; verify visible depth difference while scrolling
- [x] 2.3 Add `prefers-reduced-motion` overrides disabling parallax, blinking, and decorative animations with all content visible; verify by enabling OS reduced-motion and reloading

## 3. Advocacy content (index.html)

- [x] 3.1 Hero title screen: project name, tagline "Write your game in Go. Ship it with Godot.", one-line GDExtension/cgo description, game canvas placeholder, PRESS START CTA — all visible without scrolling; verify at 1280×800 viewport
- [x] 3.2 "Why Go + Godot?" character-select section with 6 benefit cards (static typing, goroutines, compile speed, single-binary cross-platform, game-friendly GC, stdlib/tooling), each with inline-SVG pixel icon + headline + blurb; verify all six render with icons
- [x] 3.3 Level 1 Getting Started: requirements as "party member" chips (clang-format, gcc) and pixel terminal block containing `go get github.com/godot-go/godot-go`; verify command text is selectable
- [x] 3.4 Level 2 Build & Test: terminal blocks for `make generate && make build` and `make test` with one-line explanations; verify copyable
- [x] 3.5 Boss-level FAQ accordion with ≥4 Q&As (performance, GC pauses, Godot version support, API stability); verify expand/collapse works with click and keyboard
- [x] 3.6 Continue-screen footer: GitHub repo link, Godot Discord invite, credits (ShadowApex godot-go, godot-cpp), original reference links; verify all hrefs match the original index.md values

## 4. Mini-game (js/game.js)

- [x] 4.1 Canvas setup: 640×220 logical resolution, `imageSmoothingEnabled=false`, CSS upscale with `image-rendering: pixelated`, static "INSERT COIN" fallback frame drawn without JS (noscript + initial paint); verify crisp pixels and fallback with JS disabled
- [x] 4.2 Fixed-timestep loop (accumulator at 60Hz) with gopher physics: run left/right, gravity, jump; controls via canvas `keydown` (←/→/↑/Space, canvas has `tabindex="0"`) and touch tap; verify jump/move on desktop and tap-to-jump via touch emulation
- [x] 4.3 Scrolling ground/pits, spike hazards, Go gems (+10) and Godot icons (+50) with AABB collision, collect effects, and HUD (SCORE / BEST); verify scoring and hazard death by playing
- [x] 4.4 Game-over state with "GAME OVER — TRY AGAIN? [ENTER]" and restart handling; BEST persisted to `localStorage` (try/catch wrapped) and shown in HUD; verify best survives reload
- [x] 4.5 Visibility discipline: IntersectionObserver + `visibilitychange` pause/resume the loop; keyboard scoped to canvas so unfocused arrows scroll the page; verify CPU idle when scrolled away and normal page scroll when canvas unfocused

## 5. Site interactivity (js/site.js)

- [x] 5.1 PRESS START: click/tap/Enter smooth-scrolls to first content section; verify scroll lands on the benefits section
- [x] 5.2 WebAudio blip synth (jump/collect/click/game-over), lazy `AudioContext` on first gesture, muted by default, pixel sound toggle persisted in `localStorage` (`godotgo.sound`); verify silence on first visit, sounds after enabling, persistence after reload
- [x] 5.3 Scroll-reveal: IntersectionObserver adds `.revealed` once per section with pixel-dissolve CSS; terminal blocks type out with blinking cursor on reveal (full text immediately under reduced motion); verify one-time reveal and reduced-motion behavior
- [x] 5.4 Hover/click feedback: button depress, card lift with hard-shadow shift, link color flicker; verify on hover/press
- [x] 5.5 Konami code (↑↑↓↓←→←→BA) → `body.rainbow` hue-cycling + pixel particle burst; verify sequence triggers effect anywhere on page

## 6. Integration & verification

- [x] 6.1 Full manual test matrix: desktop Chrome/Firefox, mobile viewport (tap-to-jump, layout), reduced-motion mode, JS-disabled mode, 400% zoom — no console errors, all specs' scenarios pass; record results
- [x] 6.2 Verify Jekyll passthrough safety (no Liquid sequences, `_config.yml` minimal) and that the page renders correctly from a local static server exactly as it will on GitHub Pages; preview final page for user sign-off
