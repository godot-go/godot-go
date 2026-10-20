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

The crash website is therefore not captured by in-process handlers, and the two external capture routes divide up what survives:

- **WER LocalDumps was measured absent**: run 39222394609 reproduced the crash in both plain repros (exit codes 139 in both `.rc` files, ~21 s each) yet produced **no** minidump — the kernel-side unhandled-exception reporting does not dispatch during `ExitProcess`'s teardown phase, so WER cannot be the corpus.
- **What survives is the debug-event route, minus gdb's cost**: an attached debug client receives the unhandled exception as a debug event *directly*, regardless of whether in-process dispatch runs. The automated capture is therefore **procdump** (Sysinternals): `procdump -accepteula -mm -e -x C:\crashdumps .\Godot_v4.7-stable_win64.exe <args>` starts the repro with procdump as its debugger (`-x` grammar: dump folder, then image; `-o` is a bare overwrite flag, not the dump path — measured in run 39224719377). procdump does not resolve symbols or load debug info, so it does not reproduce gdb's synchronous first-chance storm. WER keys stay registered as harmless secondary evidence, and gdb remains a documented manual fallback recipe for interactive debugging.

**v3 measured the fault's relationship to instrumentation** (run 39225647458): under procdump as full monitor, the editor repro took ~5 min (vs 21 s plain), ran to completion, and exited `0x00000000` — **no second-chance event, no dump** — while the plain control run in the same job still crashed with exit 139. gdb (run 39220040720) similarly never stopped through 12+ minutes. **v4 completed the discriminator** (run 39227973049): in the same job, the plain control editor repro still crashed (exit 139), but both procdump-monitored runs — a full second-chance monitor on the suite repro and a display-mode watcher (`-e 1 -f ''`, first+second chance, include-all filter confirmed in the monitor banner) on the editor repro — ran to completion and logged **zero exception events of any kind**, exiting `0x00000000`. A godot+Go process produces benign first-chance exceptions in normal runs, so the empty watcher log rules reading (a) out: instrumented scheduling avoids the fault entirely; external debug-event capture cannot observe this crash. The remaining evidence routes are in-process breadcrumbs that survive on the plain (crashing, uninstrumented) runs — e.g. teardown-boundary markers with `CaptureStackBackTrace` module+RVA output symbolized offline with the README recipe — or the D4 `os.Exit` discriminator to localize the crash region without a stack. **v5 cleared the DbgHelp suspect** (run 39360425245): with the Windows capture compiled out entirely (`GODOT_GO_ENABLE_DBGHELP` absent, probe reports `capture=0`, so `dbghelp.dll` never loads), the plain editor repro still exits 139 — the crash is independent of DbgHelp. The suite second-chance procdump pass again produced no dump (same instrumentation aversion as v3/v4).

The gold evidence from a minidump is extractable offline without cdb/windbg: the minidump format is documented, and a small offline reader yields the exception record (faulting thread, exception code, `TargetRip`), the module list with base addresses, and raw stack memory. Return addresses are harvestable by pointer-scanning the stack against the module ranges, then symbolized with the README's `llvm-symbolizer --obj=… --pdb=…` recipe (base+RVA → VMA) against the demo DLL's linked PDB (windows-pdb-line-info) — which is why the upload step now also carries the built `libgodotgo` DLL and PDB next to the dumps.

Run detection without a debugger relies on process exit codes: each plain run records `${PIPESTATUS[0]}` into a `.rc` file — `139` (the msys bash decoration of STATUS_ACCESS_VIOLATION) or the make-style `2816` are the crash signatures — so "the crash did not reproduce" is still detectable, but a green job now requires a dump (the whole point is symbolizable evidence, not just reproduction).

### D2: The procdump run is self-asserting, and the artifact carries the offline-analysis inputs

LocalDumps keys are registered as secondary evidence; the enforced evidence is a procdump minidump. The assertion step accepts a run only when at least one `.dmp` exists on the dump path — an exit-code-only reproduction no longer paints the job green, because the change's deliverable is a symbolizable stack. Uploads carry, next to the dumps: the plain-run logs and `.rc` files, plus `test/demo/lib/*.dll` and `*.pdb` (the DLL rebuilt on the same runner) so `llvm-symbolizer --obj=… --pdb=…` can run offline on the dump-extracted faulting addresses. gdb remains available as a manual, runner-side fallback recipe for an engineer debugging interactively, but no automated gdb steps remain in the workflow.

### D3: The diagnostics job asserts its own success and remains dispatch-only

The gdb run ends with a marker line (`GODOTGO_CRASH_EVIDENCE: captured` printed only if gdb's transcript contains a faulting-stop indicator such as `received signal SIGSEGV` or `EXCEPTION_ACCESS_VIOLATION`). A final assertion step fails the job when the crash did not reproduce or no evidence was captured, so silently-empty runs are impossible. `on: workflow_dispatch` only — no `paths`/`push` triggers — because normal PR CI never pays this cost.

### D4: The `os.Exit` discriminator stays behind the gdb results, and is a revert-first experiment

Only if the gdb transcript is ambiguous about the two leading hypothesis families (H1/H3 "detach-region fault" vs H2 "Go-thread termination fault"): a temporary branch makes the Go shutdown callback call `os.Exit(<driver status>)` so the engine never returns from the shutdown callback; a single dispatch run tells us whether the crash disappears (supporting the game "engine-return-path/CRT" family) or persists (pointing at the engine-internal cleanup or load-time Go state). It must be reverted immediately after the evidence lands — it never merges, and it changes exit semantics that the runtime gate elsewhere depends on.

### D6: v6 — the silent fault is post-unload, and pinning the module both proves it and fixes it

The v5 conclusion ("the fault executes where exception dispatch is already dead") was directionally right but pinpointed the wrong region. Three measurements resolved it:

1. **A stack-safe in-process VEH works for faults before the unload.** Taking the handler lock before any call, caching the log path and opening the file handle at install, and reserving stack with `SetThreadStackGuarantee` captured the `--editor --quit` fault (run 39369282023): `EXCEPTION 0xC0000005`, `ACCESS type=0 fault=0x8`, faulting RIP `rva=0x53e90a2`, unwound with `RtlVirtualUnwind` to `HashMap::_insert` ← `DocTools::generate` (`this = 0x8`) ← `EditorHelp::_gen_extensions_docs` ← `MessageQueue::_call_function` ← `MessageQueue::flush`. That is an engine deferred-call use-after-null: the doc-cache worker queues `_gen_extensions_docs`, `EditorHelp`'s predestructor nulls `doc`, then cleanup's flush runs the deferred call. The engine-only control (empty project, no extension) exits 0; the extension only widens the startup window in which the call is still queued.
2. **The suite and `--editor --quit-after 12` faults are silent because they happen after the module is unmapped.** Raw exit codes from `Start-Process` were `0xC0000005` for both (run 39371788796). The hardened handler still logged nothing for them while logging the pre-unload fault in the same job. Godot releases the extension in `Main::cleanup` → `unregister_core_types()` → `memdelete(gdextension_manager)` → `~GDExtension` → `FreeLibrary`; once unmapped, the handler (and Go's own handlers) are gone, so the access violation cannot be reported in-process.
3. **Pinning the module is decisive.** `GetModuleHandleExA(GET_MODULE_HANDLE_EX_FLAG_PIN | GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS, ...)` at package init turned both faults into clean exits in the same job (run 39373187014):

| run | module pinned | raw exit |
| --- | --- | --- |
| suite `--quit-after 12` | no | `0xC0000005` |
| suite `--quit-after 12` | yes | `0x00000000` |
| editor `--quit-after 12` | no | `0xC0000005` |
| editor `--quit-after 12` | yes | `0x00000000` |

Verdict: H2 (Go runtime state outliving the module) is the only family consistent with all observations — the fault exists only while the module is unmapped, and keeping it mapped eliminates it. H1/H3 (detach-region decorators) are not involved. The fix is `fix-windows-teardown-unload` (merged as #180); the `--editor --quit` DocTools fault is a separate engine bug, left tolerated in `Gen Test Project Files`.

## Risks / Trade-offs

- gdb's symbol load for the ~185 MB extension DLL can be slow even with `auto-solib-add off`; the workflow keeps a generous step timeout and, failing that, the D2 dump path still yields module+offset facts from the dump headers.
- `--editor --quit` mode may crash through a slightly different final path than the headless suite run (editor has extra teardown phases); if the faulting stack from the editor repro looks nothing like the suite crash, the same job's gdb wrapper runs the suite-mode command as a second gdb pass and both transcripts are captured.
- gdb stopping on all first-chance exceptions (not just AVs) can produce interactive prompts under `--batch`; `set confirm off` plus explicit `set stop-on-solib-events 0` keeps batch mode non-interactive.
- Evidence exists only in artifacts/issue comments; if a fix change lands much later, the recorded transcript is the primary source of truth.
