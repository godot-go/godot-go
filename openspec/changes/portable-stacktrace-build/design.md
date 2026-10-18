# Design

## Context

`pkg/log/stacktrace.h` is included by four Go packages (via cgo preambles) and by every C callback wrapper, generated or hand-written. It currently `#include <execinfo.h>` on all non-Windows platforms, so the glibc-only dependency propagates project-wide; on musl the compile fails at every one of those sites (issue #127). The actual backtrace implementation lives only in `pkg/log/stacktrace.c`, guarded by the global `enablePrintStacktrace` (default `false`; its sole enable site in `pkg/core/godot.go` is commented out). All other translation units only call `printStacktrace()`.

Constraints:

- The Go toolchain provides no automatic `musl` build tag (verified against go1.25 `go help buildconstraint` and GOROOT), so libc detection cannot ride on `//go:build`.
- cgo respects `//go:build` on `.c` files, but gating there would remove real capture from default glibc builds too — a workflow change for maintainers.
- The `Makefile` hard-assigns `CGO_CFLAGS`/`CGO_LDFLAGS` inline on the `go build` line in `build` and `build-full`, so a user's exported values are overwritten; any documented opt-in needs a pass-through hook.
- CI (`ci_linux.yaml`) runs generate → build → test on ubuntu (glibc) with a downloaded Godot binary; no musl coverage today.

## Goals / Non-Goals

**Goals:**

- Build succeeds on musl with only the base toolchain; glibc/Apple/Windows behavior byte-identical.
- The public header stays platform-neutral; platform specifics live in the one `.c` file.
- A user who wants real C backtraces on musl can get them via build flags only, linking a user-supplied execinfo-ABI library, no source edits.
- A CI job keeps musl compilability from regressing.

**Non-Goals:**

- See proposal.md Non-goals. Design-level addition: no auto-detection of libexecinfo at build time (cgo has no configure step); the opt-in is explicit.

## Decisions

### D1: Trim the public header, keep call sites untouched

`stacktrace.h` keeps only `<stdbool.h>`/`<stdio.h>`/`<stdlib.h>`, the `enablePrintStacktrace` extern, and the `printStacktrace()` prototype. `<execinfo.h>` and `<windows.h>`/`<dbghelp.h>` move into `stacktrace.c`.

- Alternative: leave the header and add platform guards around the include — still forces every includer through a platform decision and keeps the header as the failure point.
- Alternative: delete the feature and all `printStacktrace()` call sites (touches two codegen templates, six hand-written C files, forces regeneration). Rejected: the seam is a maintainer debug tool and the diff surface is ~10x for zero extra build benefit.

### D2: Feature-detect with `__has_include`, legacy macros as fallback, explicit define as override

In `stacktrace.c`:

```c
#if defined(GODOT_GO_ENABLE_EXECINFO)
#  include <execinfo.h>            /* user forced: external execinfo-ABI library */
#  define GODOT_GO_HAVE_EXECINFO 1
#elif defined(__has_include)
#  if __has_include(<execinfo.h>)  /* glibc and Apple ship it; musl does not */
#    include <execinfo.h>
#    define GODOT_GO_HAVE_EXECINFO 1
#  endif
#elif defined(__GLIBC__) || defined(__APPLE__)
#  include <execinfo.h>
#  define GODOT_GO_HAVE_EXECINFO 1
#endif
```

The existing backtrace body compiles under `#if defined(GODOT_GO_HAVE_EXECINFO)`; the `#else` branch is an empty function body. The `_WIN32` branch (whose body is fully commented out today) is preserved as-is inside the `.c`.

- Why `__has_include` first: it asks the real question ("is the header here?") instead of enumerating libc families; supported by GCC >= 5 / clang >= 3.3, far below any toolchain this project targets in 2026.
- Why legacy macros kept: pre-C23 toolchains without `__has_include` still land on glibc/Apple correctly; anything older or exotic can pass `-DGODOT_GO_ENABLE_EXECINFO` manually.
- Alternative: platform macros only (`__GLIBC__`/`__APPLE__`/`__BIONIC__`) — brittle for BSDs and future libcs.
- Alternative: `_Unwind_Backtrace`/`dladdr` portable unwinder — rejected: Go frames lack `.eh_frame` data, so unwinding across the Go/C mix in this process is unreliable; high risk for a default-off debug aid.

### D3: Makefile pass-through for the opt-in

Add `EXTRA_CGO_CFLAGS` and `EXTRA_CGO_LDFLAGS` (empty defaults) appended to the inline `CGO_CFLAGS`/`CGO_LDFLAGS` values in the `build` and `build-full` recipes. The opt-in command shape is:

```
make build EXTRA_CGO_CFLAGS='-DGODOT_GO_ENABLE_EXECINFO -I/path/to/execinfo/include' EXTRA_CGO_LDFLAGS='-L/path/to/execinfo/lib -lexecinfo'
```

- Reality check (2026-10, second CI run): Alpine's repositories ship no `libexecinfo` package in main/community/testing, and neither `libunwind` nor `gcompat` exports `backtrace`/`backtrace_symbols` (verified against the v3.23/edge APKINDEXes and `nm -D` on the shipped `.so`s). The original design assumed `apk add libexecinfo-dev` existed; it does not. The flag contract itself is unchanged and libc-agnostic — the external implementation is user-supplied on Alpine, and the CI job verifies the path against a minimal execinfo-ABI shim compiled in the job.

- Alternative: document `go build` directly — bypasses the Makefile's flags and duplicates them.
- Alternative: `GOFLAGS`-based injection — awkward with per-env assignments in recipes.

### D4: Build-only Alpine CI job in `ci_linux.yaml`

New job `build_alpine`: `runs-on: ubuntu-latest`, `container: golang:1.27-alpine`, steps: `apk add --no-cache build-base`, `make generate`, `make build` (default, backtrace compiled out), then the opt-in build against a minimal execinfo-ABI shim compiled in the job (see D3 reality check). No Godot download and no `make test`: `make generate` consumes the committed `godot_headers/`, so the pipeline is fully headless, and there is no Alpine Godot runtime harness. `paths` filters gain `.github/workflows/ci_linux.yaml` already covers itself; no new filter needed.

- Image pin: `go.mod` requires `go >= 1.27.0`, and official `golang:*` images set `GOTOOLCHAIN=local`, so the container must run the 1.27 line directly. (The ubuntu job's `setup-go` "1.26.5" does not constrain what actually runs: setup-go resolves `go.mod` toolchain directives itself; first CI run proved `golang:1.26.5-alpine` fails at `go generate`.)

- Alternative: full test job with an Alpine Godot build — out of scope (see proposal Non-goals; would mirror `ci_build_godot_linux.yaml` under musl).

### D5: README note in "Building Godot-Go"

Under `README.md` -> Getting Started -> Building Godot-Go: one short paragraph — on Alpine the extension builds as-is with the C backtrace disabled; Alpine ships no installable execinfo package (D3 reality check), so real capture requires a user-supplied execinfo-ABI library plus the D3 command pointing `-I`/`-L` at it.

## Risks / Trade-offs

- [musl users silently lose C backtraces] → Feature is default-off and unreachable at runtime today (enable site commented out); documented opt-in (D3/D5) restores it with a user-supplied library; Go-side zap stacktraces remain at Warn+.
- [`__has_include` missing on an exotic compiler misselects the branch] → Legacy macro fallback covers glibc/Apple; explicit `-DGODOT_GO_ENABLE_EXECINFO` override; Alpine CI pins the musl case.
- [Future edit re-adds an execinfo include to a shared header] → Alpine CI job fails on exactly that regression.
- [Alpine container misses a build dependency] → `build-base` provides gcc/ar; `golang:alpine` ships git; job is compile-only so no runtime deps.
- [`-DGODOT_GO_ENABLE_EXECINFO` without a linked execinfo implementation] → Link error with a clear `undefined reference to backtrace`, which points at the missing library; the README lists both halves of the opt-in (define + `-I`/`-L`/`-lexecinfo`) together.

## Migration Plan

No user-facing API changes; additive CI + docs. Rollback is reverting the two-file `pkg/log` change and the workflow/README/Makefile edits. Downstream users on musl need no action beyond `make build`.
