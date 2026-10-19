# Tasks

## 1. Diagnostics workflow

- [x] 1.1 Create `.github/workflows/ci_windows_crash_debug.yaml` (`on: workflow_dispatch` only): godot 4.7 download + window setup mirroring `ci_windows.yaml` (env vars `CI=1 LOG_LEVEL=WARN LANG`, Go setup, checkout), plus `scoop install gdb`.
- [x] 1.2 Add the gdb run (design D1): `--editor --quit` repro under `gdb --batch` with the exception-stop transcript (`bt 40`, `info registers rip rsp rbp`, `info sharedlibrary`, `thread apply all bt 30`), a second gdb pass of the full-suite commands only if the first transcript stops on something else, and generous step timeouts.
- [x] 1.3 Add the WER LocalDumps fallback (design D2): registry keys for `Godot_v4.7-stable_win64.exe` (DumpFolder `C:\crashdumps`, DumpType 1), grep the folder after the run.
- [x] 1.4 Add the success assertion + artifact upload (design D3): `GODOTGO_CRASH_EVIDENCE: captured` printed only on a real faulting-stop indicator in the transcript; final assertion step fails the job when evidence is absent; upload the gdb transcript(s), dump(s), and `test-output.log` as `windows-crash-evidence` with `if: always()`.

## 2. Evidence

- [ ] 2.1 Dispatch the workflow on this branch once implemented; verify the job turns red if the crash fails to reproduce (the assertion fires) and green-with-artifacts when it reproduces.
- [ ] 2.2 Resolve the faulting address against the DLL/PDB using the README's offline recipe (`llvm-symbolizer --obj=… --pdb=…` with base+RVA) and record the symbolized faulting stack in `tasks.md` evidence plus the gdb transcript group it came from.
- [ ] 2.3 Post the findings on issue #172 (observed faulting frame, hypothesis verdict between H1/H3 decorator-region vs H2 Go-thread-termination vs our-code H4/H5), then hand off to the follow-up fix change; leave this change implemented-but-not-archived until that verdict is recorded.
- [ ] 2.4 Only if the gdb transcript is ambiguous (design D4): run the one-off `os.Exit` discriminator on a temporary branch, record the result here, and revert the branch without merging.
