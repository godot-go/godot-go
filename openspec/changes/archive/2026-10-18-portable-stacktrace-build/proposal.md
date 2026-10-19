# Proposal

## Why

`make build` fails on Alpine Linux (musl libc) because `pkg/log/stacktrace.h` unconditionally includes the glibc-only header `<execinfo.h>` (issue #127). Since that header is pulled into every package that crosses the GDExtension boundary (`pkg/log`, `pkg/ffi`, `pkg/core`, plus all callback wrappers), the build is broken on every musl-based system, not just in `pkg/log`.

## What Changes

- `pkg/log/stacktrace.h` no longer includes `<execinfo.h>` (or the Windows headers). It declares only `enablePrintStacktrace`, `printStacktrace()`, and the standard-library includes callers actually need.
- `pkg/log/stacktrace.c` feature-detects `<execinfo.h>` via `__has_include` (falling back to `__GLIBC__`/`__APPLE__` macros for compilers without `__has_include`). Where the header exists, the existing `backtrace()`/`backtrace_symbols()` implementation is compiled unchanged. Where it does not (musl, Android), `printStacktrace()` compiles to a no-op.
- README documents the optional musl upgrade path: `apk add libexecinfo-dev` plus `CGO_LDFLAGS='-lexecinfo'` restores real C backtraces on Alpine by enabling the execinfo code path with a preprocessor define.
- `.github/workflows/ci_linux.yaml` gains a build-only job in a `golang:*-alpine` container that runs `make generate` and `make build`, so musl compatibility is enforced in CI. No Godot runtime tests in that job (no Alpine Godot headless harness exists in CI).

## Capabilities

### New Capabilities

- `native-stacktrace-capture`: C-side callstack capture at GDExtension callback boundaries — its opt-in switch, platform-conditional availability, and graceful degradation on libc implementations without `execinfo`.

### Modified Capabilities

<!-- None: no existing spec's requirements change. Callback wrappers keep calling printStacktrace() unchanged. -->

## Impact

- Code: `pkg/log/stacktrace.h`, `pkg/log/stacktrace.c` only. No templates, no generated files, no call sites change — every `printStacktrace()` caller compiles against the trimmed header as before.
- Behavior: glibc/macOS/Windows builds identical. musl builds succeed; the C backtrace feature is a no-op there unless the libexecinfo opt-in is used (matching today's runtime state, where the feature is off by default and its only enable site is commented out).
- Docs: README build prerequisites note for Alpine.
- CI: one new build-only job in `ci_linux.yaml`.

## Non-goals

- Restoring or reimplementing C backtraces on musl in-tree (no `_Unwind_Backtrace` / `dladdr` implementation).
- Changing the default-on/off behavior of `enablePrintStacktrace` or reviving the commented-out enable site in `pkg/core/godot.go`.
- Removing the C stacktrace feature entirely.
- Running the Godot test suite (`make test`) on musl; CI coverage is compile-only.
- Supporting Android or other non-glibc, non-Apple libc with real backtraces (they get the no-op stub).
