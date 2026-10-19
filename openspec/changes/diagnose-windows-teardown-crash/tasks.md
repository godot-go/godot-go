# Tasks

## 1. Diagnostics workflow

- [x] 1.1 Create `.github/workflows/ci_windows_crash_debug.yaml` (`on: workflow_dispatch` only): godot 4.7 download + window setup mirroring `ci_windows.yaml` (env vars `CI=1 LOG_LEVEL=WARN LANG`, Go setup, checkout), plus `scoop install gdb`.
- [x] 1.2 Implement the plain-run capture (design D1, v2): the `--editor --quit` repro and the suite phase-1 command run without instrumentation, each recording its exit code (`139` / `2816` are the crash signatures via `${PIPESTATUS[0]}`) into a `.rc` file. Measured in run 39222394609: both repros reproduced the crash (exit 139) but no WER dump was produced — the kernel-side unhandled filter does not dispatch during `ExitProcess` teardown. gdb's v1 being infeasible (run 39220040720, 6 s repro → >12 min with no stop) is retained only as a manual fallback recipe.
- [x] 1.3 Implement the procdump capture (design D1, v3): `scoop install procdump`, crashdumps folder created, and both repros launched under `procdump -accepteula -e -o C:\crashdumps -x` (second-chance/unhandled fault → dump); WER LocalDumps keys remain registered as secondary evidence.
- [x] 1.4 Add the success assertion + artifact upload (design D2/D3): `GODOTGO_CRASH_EVIDENCE: captured` printed only when at least one `.dmp` exists on the dump path (an exit-code-only reproduction no longer passes); final assertion step fails the job otherwise; upload run logs, `.rc` files, dumps **and** the built `test/demo/lib` DLL + PDB (offline `llvm-symbolizer` inputs) as `windows-crash-evidence` with `if: always()`.

## 2. Evidence

- [x] 2.1 Dispatch the workflow across its iterations (main ref for v1, branch ref for v2/v3): verified the assertion turns the job red when evidence is missing (v1 gdb run 39220040720 cancelled mid-run) and green when the crash reproduces with evidence (v2 run 39222394609: both repros exit 139, `GODOTGO_CRASH_EVIDENCE: captured`; the absence of WER dumps drove the v3 revision).
- [ ] 2.2 Extract the faulting thread, exception record, module bases, and faulting addresses from the procdump minidump offline (documented minidump reader; pointer-scan the stack against module ranges), then symbolize against the DLL/PDB with the README's recipe (`llvm-symbolizer --obj=… --pdb=…`, base+RVA → VMA) and record the symbolized faulting stack in `tasks.md` evidence.
- [ ] 2.3 Post the findings on issue #172 (observed faulting frame, hypothesis verdict between H1/H3 decorator-region vs H2 Go-thread-termination vs our-code H4/H5), then hand off to the follow-up fix change; leave this change implemented-but-not-archived until that verdict is recorded.
- [ ] 2.4 Only if the gdb transcript is ambiguous (design D4): run the one-off `os.Exit` discriminator on a temporary branch, record the result here, and revert the branch without merging.
