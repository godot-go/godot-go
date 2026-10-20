# Tasks

## 1. Pin the Windows module

- [x] 1.1 Add `pkg/core/module_windows.go` (build tag `windows`): a cgo helper that calls `GetModuleHandleExA(GET_MODULE_HANDLE_EX_FLAG_PIN | GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS, ...)` on the module's own address at package initialization, returning the `BOOL` so a failure is logged with `log.Warn`. Verify: `GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=<mingw wrapper> go build ./pkg/core/` succeeds and the package initializes the pin (code inspection of the init).
- [x] 1.2 Keep the pin unconditional: the `GODOT_GO_PIN_DLL` experiment switch must not reach product code. Verify: `rg GODOT_GO_PIN_DLL pkg/` returns no matches.
- [x] 1.3 Document the constraint in the package comment: the module is never unmapped on Windows, so hot-reload/unload is unsupported and a restart is the supported path. Verify: the doc comment states both the mechanism and the limitation.

## 2. Restore the Windows gate

- [x] 2.1 Remove `continue-on-error: true` from the `Test` step in `.github/workflows/ci_windows.yaml`; leave it on `Gen Test Project Files` (the separate `--editor --quit` engine DocTools race). Verify: `python3 -c "import yaml; ..."` parses the workflow and the `Test` step has no `continue-on-error`.
- [x] 2.2 Validate the fix on Windows: run the branch's `ci_windows` job (dispatch or PR) and confirm both `make test` invocations exit `0` (suite `--quit-after 12` and editor `--quit-after 12`) with the `Test` step green. Verify: run 39375069018 is green and the job log shows `make test` completing without `godot-exited-non-zero`.
- [x] 2.3 Confirm Linux is unaffected: the same PR's Linux job stays green. Verify: run 39375071034 is green.

## 3. Documentation

- [x] 3.1 Add a Windows module-lifetime note to `AGENTS.md` (architecture/testing section): the extension DLL is pinned and must not be expected to hot-reload. Verify: `grep -n "pinned" AGENTS.md` shows the note.

## Workflow follow-up

- Archive this change after merge (`chore: archive fix-windows-teardown-unload change`) and sync `extension-library-lifetime` into `openspec/specs/`.
- Update and archive `diagnose-windows-teardown-crash` with the captured root cause and the pin experiment evidence, then post the summary on issue #172.
