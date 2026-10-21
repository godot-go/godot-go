# Tasks

## 1. Portable gate parsing

- [x] 1.1 Replace the GNU-only `\x1b` ANSI strip in `test/check_test_output.sh` with an ESC byte emitted by `printf`, and verify: `grep -n '\\x' test/check_test_output.sh` returns nothing, and the gate parses a colour-wrapped log (`printf '\033[92m   PASSES: 1077\033[39m\n'` plus the other required markers) to a passing verdict
- [x] 1.2 Apply the same change to `test/check_editor_output.sh` and verify the same way, including that its editor-marker checks still fire on a log missing `GODOTGO_EDITOR_PLUGIN_ENTERED`
- [x] 1.3 Run `make test_gate` and confirm the fixture suite still passes unchanged, proving the strip change does not alter verdicts on the existing fixtures

## 2. darwin build gap

- [x] 2.1 Add `pkg/util/threadname_darwin.go` as a no-op with the Windows rationale, and verify `GOOS=darwin GOARCH=arm64 go build ./pkg/util` and `GOOS=darwin GOARCH=amd64 go build ./pkg/util` both succeed, with `GOOS=darwin go list -f '{{.GoFiles}}' ./pkg/util` listing the new file
- [ ] 2.2 Confirm the previously failing symbol resolves: `GOOS=darwin GOARCH=arm64 go vet ./pkg/util` is clean, and the CI build leg that reported `undefined: util.SetThreadName` now reaches the test steps

## 3. macOS workflow

- [x] 3.1 Create `.github/workflows/ci_macos.yaml` with the sibling triggers (dispatch, PR/push path filters including the workflow file itself) and a `fail-fast: false` matrix binding `macos-latest`+`arm64` and `macos-26-intel`+`amd64`; verify the file parses (`actionlint` if installed, otherwise a YAML parse) and that each leg resolves a distinct runner
- [x] 3.2 Add the Godot 4.7 stable install (`curl` the universal zip, run `Godot.app/Contents/MacOS/Godot`, defensive `xattr -dr`), header update, `make generate`, `git diff`, and `make build` steps with `GOOS=darwin GOARCH=${{ matrix.arch }}`; verify with `make -n build GOOS=darwin GOARCH=arm64 | grep -o 'libgodotgo-test-macos-arm64.framework'` and the amd64 equivalent, confirming each leg names its own `.gdextension` entry
- [x] 3.3 Add `make ci_gen_test_project_files` (3 min) and `make test` (7 min) steps; verify the workflow passes `CI=1`, `LOG_LEVEL`, and `LANG` exactly as the Windows job does, and that `GODOT` points at the app binary
- [x] 3.4 Add the capture-evidence step asserting `stacktrace-probe capture=1` and `___ C start backtrace ___` only (no `file:line`), plus artifact upload named `macos-${{ matrix.arch }}-test-lib` containing `test-output.log`, `editor-output.log`, and the built `.framework`; verify the YAML parses and the names interpolate the architecture
- [x] 3.5 Add the macOS badge to `README.md` next to the Linux and Windows badges and verify its target URL matches `.github/workflows/ci_macos.yaml`

## 4. Proof on the platform

- [ ] 4.1 Confirm both legs of the PR run complete; verify with `gh run view --log` that both compile the extension (no `undefined: util.SetThreadName`) and that no leg reports `FAIL[driver-summary]` (the ANSI-strip regression), and record each leg's step timings against the 3/7 min step budgets
- [ ] 4.2 Download both artifacts and confirm each contains the logs and a Mach-O `.framework` of the matching architecture (`file test/demo/lib/*.framework` reports `arm64` and `x86_64` respectively)
- [ ] 4.3 Confirm the capture banner is present in both legs' `test-output.log` and that teardown exited zero; if a leg faults at shutdown instead, stop and report it rather than weakening the job

## Workflow follow-up

- Open the pull request from `feat-add-macos-ci` against `main` and let the new workflow run there.
- Archive the change once merged, then decide separately whether to mark the macOS legs as required checks.
