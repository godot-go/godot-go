# Tasks

## 1. Diagnostics workflow

- [x] 1.1 Create `.github/workflows/ci_windows_crash_debug.yaml` (`on: workflow_dispatch` only): godot 4.7 download + window setup mirroring `ci_windows.yaml` (env vars `CI=1 LOG_LEVEL=WARN LANG`, Go setup, checkout), plus `scoop install gdb`.
- [x] 1.2 Implement the plain-run capture (design D1, v2): the `--editor --quit` repro and the suite phase-1 command run without instrumentation, each recording its exit code (`139` / `2816` are the crash signatures via `${PIPESTATUS[0]}`) into a `.rc` file. The gdb-based v1 was measured infeasible in run 39220040720 (6 s repro → >12 min with no stop; gdb's synchronous first-chance handling amplifies godot's exception traffic) and is retained only as a manual fallback recipe.
- [x] 1.3 Add the WER LocalDumps capture (design D2): registry keys for `Godot_v4.7-stable_win64.exe` (DumpFolder `C:\crashdumps`, DumpType 1), and after the runs list the folder contents.
- [x] 1.4 Add the success assertion + artifact upload (design D3, v2): `GODOTGO_CRASH_EVIDENCE: captured` printed only when a `.dmp` exists in `C:\crashdumps` **or** a run's `.rc` records a crash signature (`139` / `2816`); final assertion step fails the job when neither reproduced; upload the run logs, `.rc` files, and dumps as `windows-crash-evidence` with `if: always()`.

## 2. Evidence

- [ ] 2.1 Dispatch the workflow on this branch once implemented; verify the job turns red if the crash fails to reproduce (the assertion fires) and green-with-artifacts when it reproduces.
- [ ] 2.2 Resolve the faulting address against the DLL/PDB using the README's offline recipe (`llvm-symbolizer --obj=… --pdb=…` with base+RVA) and record the symbolized faulting stack in `tasks.md` evidence plus the gdb transcript group it came from.
- [ ] 2.3 Post the findings on issue #172 (observed faulting frame, hypothesis verdict between H1/H3 decorator-region vs H2 Go-thread-termination vs our-code H4/H5), then hand off to the follow-up fix change; leave this change implemented-but-not-archived until that verdict is recorded.
- [ ] 2.4 Only if the gdb transcript is ambiguous (design D4): run the one-off `os.Exit` discriminator on a temporary branch, record the result here, and revert the branch without merging.
