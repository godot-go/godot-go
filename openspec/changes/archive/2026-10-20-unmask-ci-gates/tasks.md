# Tasks

## 1. Stop quitting during startup docs

- [x] 1.1 Change `ci_gen_test_project_files` in the `Makefile` from `--editor --quit` to `--editor --frame-delay 16 --quit-after 600`, with a comment naming the DocTools race and the iterations-vs-seconds detail. Verify: with the doc cache removed, the target exits `0` locally (`GODOT=<binary> make ci_gen_test_project_files`), prints no crash/abort, and writes the cache.

## 2. Remove every mask

- [x] 2.1 Remove `continue-on-error: true` from `Gen Test Project Files` in `.github/workflows/ci_linux.yaml`. Verify: the workflow YAML parses and the step has no `continue-on-error`.
- [x] 2.2 Remove `continue-on-error: true` from `Gen Test Project Files` in `.github/workflows/ci_windows.yaml`. Verify: same.
- [x] 2.3 Delete `.github/workflows/ci_windows_crash_debug.yaml` (diagnosis archived, crash fixed, its dump assertion can no longer pass). Verify: `ls .github/workflows/` no longer lists it.
- [x] 2.4 Verify the repository-wide invariant: `rg continue-on-error .github/workflows/` returns no matches.

## 3. Validate and land

- [x] 3.1 Run both workflows on the branch and confirm `Gen Test Project Files` succeeds (exit 0) and the test jobs are green on Linux and Windows. Verify: the dispatch/PR checks are green and the generation step's log ends without an `Error`.

## Post-merge check (outside this change)

- After merge, confirm the `main` push runs are green with no masked steps (both `run tests` jobs).
- Archive this change before the PR is merged (`chore: archive unmask-ci-gates change`); it has `skip_specs: true`, so no spec sync is needed.
