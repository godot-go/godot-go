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

### D1: gdb in a dispatch-only workflow captures the faulting stack

An `m:\\` `--editor --quit` repro run under `gdb --batch` with explicit exception handling: gdb sees the first-chance STATUS_ACCESS_VIOLATION on the dying process even when the in-process dispatch chain is gone — the debugger intercepts events the process's own handlers never reach. Practical shape:

```
gdb --batch
  -ex 'set pagination off'
  -ex 'set confirm off'
  -ex 'set auto-solib-add off'        # load our DLL's symbols lazily/symbolically
  -ex run                             # stops at SEGV/AV
  -ex 'bt 40'                         # faulting thread
  -ex 'info registers rip rsp rbp'
  -ex 'info sharedlibrary'            # which module owns the fault site
  -ex 'thread apply all bt 30'        # all threads; distinguishes thread-kill
                                      # variants (H2) from main-thread detach (H1/H3)
  --args ./Godot_v4.7-stable_win64.exe --headless --verbose --path test/demo/ --editor --quit
```

gdb is installed in the debug workflow via `scoop install gdb` (mingw-w64 gdb), matching the repo's existing scoop usage. A first-chance AV stop is consistent with the exception code already observed (`2816 = 0x0B00`, exception 11).

Considered alternatives and why they are fallbacks, not primary: WER LocalDumps (D2) yields a binary dump whose offline analysis needs cdb/windbg we do not have on the host; godot's or Go's in-process handlers are already proven silent; a custom in-module VEH reuses our stacktrace.c machinery but registers through the same dispatch chain the crash demonstrably avoids.

`set auto-solib-add off` keeps gdb from grinding on all loaded DLLs; the trace still lists module base addresses so the faulting module can be resolved by offset from the gdb transcript alone.

### D2: WER LocalDumps as a machine-readable fallback

The same job pre-registers the LocalDumps keys for `Godot_v4.7-stable_win64.exe` (`DumpFolder=C:\\crashdumps`, `DumpType=1` minidump), uploads any produced `.dmp`. If gdb's mingw staging misbehaves (H2-class cracks can tickle debuggers with weird loader-state events), the dump at least records exception code, faulting module, and raw stacks for later analysis by whatever tooling the fix work introduces.

### D3: The diagnostics job asserts its own success and remains dispatch-only

The gdb run ends with a marker line (`GODOTGO_CRASH_EVIDENCE: captured` printed only if gdb's transcript contains a faulting-stop indicator such as `received signal SIGSEGV` or `EXCEPTION_ACCESS_VIOLATION`). A final assertion step fails the job when the crash did not reproduce or no evidence was captured, so silently-empty runs are impossible. `on: workflow_dispatch` only — no `paths`/`push` triggers — because normal PR CI never pays this cost.

### D4: The `os.Exit` discriminator stays behind the gdb results, and is a revert-first experiment

Only if the gdb transcript is ambiguous about the two leading hypothesis families (H1/H3 "detach-region fault" vs H2 "Go-thread termination fault"): a temporary branch makes the Go shutdown callback call `os.Exit(<driver status>)` so the engine never returns from the shutdown callback; a single dispatch run tells us whether the crash disappears (supporting the game "engine-return-path/CRT" family) or persists (pointing at the engine-internal cleanup or load-time Go state). It must be reverted immediately after the evidence lands — it never merges, and it changes exit semantics that the runtime gate elsewhere depends on.

## Risks / Trade-offs

- gdb's symbol load for the ~185 MB extension DLL can be slow even with `auto-solib-add off`; the workflow keeps a generous step timeout and, failing that, the D2 dump path still yields module+offset facts from the dump headers.
- `--editor --quit` mode may crash through a slightly different final path than the headless suite run (editor has extra teardown phases); if the faulting stack from the editor repro looks nothing like the suite crash, the same job's gdb wrapper runs the suite-mode command as a second gdb pass and both transcripts are captured.
- gdb stopping on all first-chance exceptions (not just AVs) can produce interactive prompts under `--batch`; `set confirm off` plus explicit `set stop-on-solib-events 0` keeps batch mode non-interactive.
- Evidence exists only in artifacts/issue comments; if a fix change lands much later, the recorded transcript is the primary source of truth.
