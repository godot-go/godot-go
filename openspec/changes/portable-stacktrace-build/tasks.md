# Tasks

## 1. Portable stacktrace capture (pkg/log)

- [x] 1.1 Trim `pkg/log/stacktrace.h` to the platform-neutral surface (`<stdbool.h>`, `<stdio.h>`, `<stdlib.h>`, `enablePrintStacktrace`, `printStacktrace()` prototype) and verify with `cc -fsyntax-only -I. pkg/log/stacktrace.c` plus `go build ./...` on the glibc host
- [x] 1.2 Rework `pkg/log/stacktrace.c` per design D2: move the platform includes in, add the `GODOT_GO_ENABLE_EXECINFO` / `__has_include` / legacy-macro cascade, keep the existing backtrace body under `GODOT_GO_HAVE_EXECINFO`, add the empty `#else` body, and preserve the commented `_WIN32` scaffolding; verify with `cc -fsyntax-only -I. pkg/log/stacktrace.c` on the glibc host (execinfo path selected) and by defining `GODOT_GO_HAVE_EXECINFO`-blocking conditions are NOT forced (default host build unchanged)
- [x] 1.3 Verify no includer regressions: run `make generate` and `make build` on the glibc host and confirm the build completes with `printStacktrace()` call sites untouched in `pkg/ffi/ffi_wrapper.gen.c`

## 2. Opt-in plumbing and docs

- [x] 2.1 Add `EXTRA_CGO_CFLAGS` / `EXTRA_CGO_LDFLAGS` pass-through variables (empty defaults) appended to the inline `CGO_CFLAGS` / `CGO_LDFLAGS` in the Makefile `build` and `build-full` recipes, and verify `make build` succeeds unset on the glibc host and that `make build EXTRA_CGO_CFLAGS='-DGODOT_GO_ENABLE_EXECINFO'` shows the define in the `-x` compiler output (link failure without libexecinfo on glibc-host is not exercised; 3.1 covers the linked case)
- [x] 2.2 Add the Alpine paragraph to `README.md` under Getting Started -> Building Godot-Go (builds as-is with C backtrace disabled; `apk add libexecinfo-dev` + `make build EXTRA_CGO_CFLAGS='-DGODOT_GO_ENABLE_EXECINFO' EXTRA_CGO_LDFLAGS='-lexecinfo'` for real backtraces) and verify the rendered section matches the flags actually implemented in 1.2/2.1

## 3. Alpine CI coverage

- [ ] 3.1 Add the `build_alpine` job to `.github/workflows/ci_linux.yaml` per design D4 (container `golang:<version>-alpine` matching the existing job's Go version, `apk add --no-cache build-base`, `make generate`, `make build`; compile-only), and verify by running the same step sequence inside `docker run --rm -v $PWD:/src -w /src golang:<version>-alpine` and confirming both generate and build complete
- [ ] 3.2 Extend the `build_alpine` verification with the opt-in path: in the same container run `apk add --no-cache libexecinfo-dev` then `make build EXTRA_CGO_CFLAGS='-DGODOT_GO_ENABLE_EXECINFO' EXTRA_CGO_LDFLAGS='-lexecinfo'` and confirm the linked binary is produced (add the same two steps to the CI job as a second build step)

## 4. Integration checks

- [x] 4.1 Run the full `GODOT=/path/to/godot make test` suite on the glibc host and confirm green (ignoring the Known Issues list in AGENTS.md)
- [x] 4.2 Run `openspec validate portable-stacktrace-build` and confirm the change validates against `specs/native-stacktrace-capture/spec.md`

## Workflow follow-up

- Open a PR referencing issue #127; confirm the new `build_alpine` CI job runs green on the PR.
- Archive the change after review requirements are satisfied.
