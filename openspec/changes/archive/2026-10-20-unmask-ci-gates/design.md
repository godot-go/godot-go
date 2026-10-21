# Design

## Context

`ci_gen_test_project_files` runs `godot --headless --editor --quit` to materialize `.godot/` and `extension_list.cfg` before the test run. With the extension loaded, the editor's doc-cache worker is still loading when the immediate quit begins; `EditorHelp`'s predestructor sets `doc = nullptr`, then `Main::cleanup`'s `message_queue->flush()` runs the deferred `_gen_extensions_docs`, which calls `DocTools::generate` with a null `doc` (observed `this = 0x8`). On Linux it aborts (`Error 134`), on Windows it access-violates (`Error 2816`); both CI workflows masked the step, so the failure was invisible.

The crash-debug workflow (`ci_windows_crash_debug.yaml`) was built to reproduce and capture the now-fixed teardown crash. Its two procdump steps are masked and its final step requires a minidump; after the pin fix there is no crash and no dump.

## Goals / Non-Goals

**Goals:**
- Make every CI step honest: no `continue-on-error` anywhere.
- Keep the generation step doing its real job on both platforms.

**Non-Goals:**
- Fixing the engine bug (upstream), or changing test gates/scripts.
- Preserving the obsolete crash-capture workflow.

## Decisions

### D1: Give the doc-cache worker main-loop time instead of quitting during startup (`--frame-delay 16 --quit-after 600`)

`--quit-after` counts iterations, not seconds. The first frame spends several seconds in startup, during which the doc-cache worker runs; a bare `--quit-after 12` can still reach the last frame before the worker finishes, leaving its deferred `_gen_extensions_docs`/`generate_doc` call queued for `Main::cleanup`'s flush after `EditorHelp` nulled `doc`. `--frame-delay 16` forces every iteration to take at least ~16 ms, so 600 iterations guarantee ~10 s of main-loop time *after* startup. The worker then finishes and its deferred call is flushed while `doc` is alive.

Measured locally on a cold cache (doc cache removed): `--quit-after 12` aborts (exit 134), `--quit-after 300` passes, and the shipped `--frame-delay 16 --quit-after 600` exits 0 in ~17 s and writes the cache. `make test`'s editor run keeps `--quit-after 12` because the generation step has already warmed the cache in CI; using a timeout or a custom auto-quit hook would add machinery for a workaround.

### D2: Delete the crash-debug workflow rather than unmask it

Its capture contract is "reproduce the teardown crash and assert a dump"; the crash is fixed and the diagnosis is archived. Unmasking would leave a dispatch-only job guaranteed to fail. The capture recipes remain in the archived diagnosis design and git history, and the normal Windows job now covers the regression with `make test` gated.

### D3: Treat the invariant as repository-wide

`ci_build_godot_linux.yaml`, `ci_build_godot_windows.yaml`, and `version_bump.yaml` contain no masks today; the change removes the remaining four and verifies with one search.

## Risks / Trade-offs

- **The Gen step could still fail on a slow runner if the worker outlives the ~10 s frame budget** → the worker runs throughout startup as well; the budget is only the tail margin. If a runner proves slower, raise `--quit-after` (the step timeout is 3 minutes).
- **Deleting the debug workflow loses a ready-made capture harness** → recipes are in the archived design; the harness was single-purpose and now asserts a fault that cannot occur.
- **The engine bug can still be hit by anyone running `--editor --quit`** → documented in the archived diagnosis; CI no longer uses that pattern.

## Migration Plan

None. CI-only. Rollback is restoring the workflow files/Makefile flag from git.
