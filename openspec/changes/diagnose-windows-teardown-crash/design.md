# Design

## Context

From issue #172 triage (2026-10-19, comment on the issue) plus new source-reads:

```
WINDOWS TEARDOWN TIMELINE (observed in CI, runs 39212143087 / 39213753803)
-----------------------------------------------------------------------------
godot running suite                    1430 asserts, 0 failures
      v
GDExtensionManager::shutdown()         -> Go shutdown callback fires
                                          GODOTGO_MAIN_LOOP_SHUTDOWN printed
      v
Main::cleanup()                        deinit levels run; our
                                       GDExtensionBindingDeinitializeLevel
                                       executes per level without complaint
      v
return to main / WinMain               nothing in logs starts here
      v
CRT exit() + ExitProcess               <<-- CRASH ZONE, total silence
      |   godot never calls close_library for extensions at engine exit
      |   (verified: close_library fires only on reload / load-error paths)
      v
bash prints "Segmentation fault" 139   |  gen step: make Error 2816 (AV 11)
-----------------------------------------------------------------------------
```

- Two run modes crash identically: the full suite (`exit 139`) and `--editor --quit` (~10 s; `Error 2816`). The repro is minimal and does not involve test content.
- Godot's mingw build registers a CRT `signal(SIGSEGV)` handler (crash_handler_windows_signal.cpp) that would print a dbghelp-symbolized `Program crashed` backtrace; observed output contains no such banner.
- Go's runtime on Windows installs its own exception filter and would print `SIGSEGV: segmentation violation` + goroutine dump; observed output contains no such dump either.
- Conclusion: the fault executes where exception dispatch is already dead — the `ExitProcess` / DLL-detach / thread-kill region. In-process handlers cannot observe it there; an external process (debugger, WER) can.

Toolchain facts that constrain the evidence mechanism: `windows-latest` runners run with admin rights (registry writes for LocalDumps are okay), scoop is already set up in the windows workflow, and the demo DLL itself has a linked PDB (windows-pdb-line-info) so attributed `module+offset` values can be resolved offline with the llvm-symbolizer recipe already in the README.

## Goals / Non-Goals

**Goals:**
- Produce the faulting instruction pointer and a full stack of the teardown crash, even though the crash fires where in-process exception handling is dead.
- Make the diagnostics workflow assert its own success: if gdb does not stop on the fault (crash did not reproduce) or evidence capture broke, the job fails loudly rather than producing empty artifacts.
- Keep normal CI (`ci_windows.yaml`) untouched, including the merged windows gates.

**Non-Goals:**
- Fixing the crash (a follow-up change designs that from the observed stack).
- Un-masking or re-gating the windows `Test`/`Gen Test Project Files` steps.
- Diagnosing the unrelated `test_pin_scratch_stays_flat` heap flake.
- Reproducing locally (the crash is Windows-specific; host has no windows binary/wine path).

## Decisions

### D1: The dump-first strategy: WER LocalDumps captures the faulting context; gdb is a documented manual fallback, not the automated path

The first implementation wrapped the repro in `gdb --batch` with first-chance SIGSEGV/ILL/FPE passed through, expecting the unhandled teardown AV to stop the debugger cheaply at the very end. Run 39220040720 measured the opposite: the uninstrumented `--editor --quit` repro completes (and crashes) in ~6 s, but under gdb's synchronous debug-event handling it was still inside godot's editor help-cache generation after 12+ minutes with no stop occurred — every first-chance exception (constant under a Go c-shared DLL: guard-page stack growth, thread events, module loads) triggers a synchronous gdb round-trip, and all that overhead buys nothing because the crash only stops gdb at the end. The cancelled run's `gdb-editor-quit.log` transcript is retained as evidence of the amplification.

The crash website is therefore not captured by a debugger. Instead, the plain (uninstrumented) repro runs fire the crash in normal time, and Windows Error Reporting's `LocalDumps` captures it: keys are registered for `Godot_v4.7-stable_win64.exe` (`DumpFolder=C:\crashdumps`, `DumpType=1` minidump), and the crash's dying process leaves a `.dmp` regardless of which in-process handler is dead. LocalDumps is invoked by the kernel-side WER path, not by the process's exception chain, so it covers the dispatch-dead phase.

The gold evidence from a minidump is extractable offline without cdb/windbg: the minidump format is documented, and a small offline reader yields the exception record (faulting thread, exception code, `TargetRip`), the module list with base addresses, and raw stack memory. Return addresses are harvestable by pointer-scanning the stack against the module ranges, then symbolized with the README's `llvm-symbolizer --obj=… --pdb=…` recipe (base+RVA → VMA) against the demo DLL's linked PDB (windows-pdb-line-info).

Run detection without gdb relies on process exit codes: each plain run records `${PIPESTATUS[0]}` into a `.rc` file — `139` (the msys bash decoration of STATUS_ACCESS_VIOLATION) or the make-style `2816` are the crash signatures — so "the crash did not reproduce" is still detectable when WER silently refuses (see Risks).

### D2: WER LocalDumps is also self-asserting

LocalDumps keys are registered for `Godot_v4.7-stable_win64.exe` (`DumpFolder=C:\crashdumps`, `DumpType=1`, `DumpCount=10`) in the debug workflow, and the dump folder is uploaded. The offline reader (D1) consumes the dumps directly; no windbg toolchain is introduced. gdb remains available as a manual, runner-side fallback recipe for an engineer debugging interactively, but no automated gdb steps remain in the workflow.

### D3: The diagnostics job asserts its own success and remains dispatch-only

The gdb run ends with a marker line (`GODOTGO_CRASH_EVIDENCE: captured` printed only if gdb's transcript contains a faulting-stop indicator such as `received signal SIGSEGV` or `EXCEPTION_ACCESS_VIOLATION`). A final assertion step fails the job when the crash did not reproduce or no evidence was captured, so silently-empty runs are impossible. `on: workflow_dispatch` only — no `paths`/`push` triggers — because normal PR CI never pays this cost.

### D4: The `os.Exit` discriminator stays behind the gdb results, and is a revert-first experiment

Only if the gdb transcript is ambiguous about the two leading hypothesis families (H1/H3 "detach-region fault" vs H2 "Go-thread termination fault"): a temporary branch makes the Go shutdown callback call `os.Exit(<driver status>)` so the engine never returns from the shutdown callback; a single dispatch run tells us whether the crash disappears (supporting the game "engine-return-path/CRT" family) or persists (pointing at the engine-internal cleanup or load-time Go state). It must be reverted immediately after the evidence lands — it never merges, and it changes exit semantics that the runtime gate elsewhere depends on.

## Risks / Trade-offs

- gdb's symbol load for the ~185 MB extension DLL can be slow even with `auto-solib-add off`; the workflow keeps a generous step timeout and, failing that, the D2 dump path still yields module+offset facts from the dump headers.
- `--editor --quit` mode may crash through a slightly different final path than the headless suite run (editor has extra teardown phases); if the faulting stack from the editor repro looks nothing like the suite crash, the same job's gdb wrapper runs the suite-mode command as a second gdb pass and both transcripts are captured.
- gdb stopping on all first-chance exceptions (not just AVs) can produce interactive prompts under `--batch`; `set confirm off` plus explicit `set stop-on-solib-events 0` keeps batch mode non-interactive.
- Evidence exists only in artifacts/issue comments; if a fix change lands much later, the recorded transcript is the primary source of truth.
