# Retrospective: Windows teardown crash

## Goal

Find and fix the Godot process crash that followed a fully green Windows suite (`exit 139` / `STATUS_ACCESS_VIOLATION`) and made the Windows `Test` gate unusable.

## Investigation journey

**External observers first — and they all came up empty.** gdb never stopped on the fault (and slowed a 21 s repro past 12 minutes); procdump's second-chance and display modes logged zero exception events; WER LocalDumps stayed empty. Worse, every repro run under a debugger *exited 0*. The crash is instrumentation-averse: attaching a debugger changes the timing that lets the fault happen.

**Suspects were eliminated one at a time.** DbgHelp was cleared by compiling the Windows stack capture out entirely (`capture=0` probe, no `dbghelp.dll` import) and still crashing. Go GC/finalizers and `GODEBUG` flags were cleared by control runs.

**In-process capture was the only route.** A vectored exception handler installed at DLL load could see faults the external tools could not. The first version captured nothing, which looked like "the handler never runs"; the actual bug was the classifier: a null dereference lands in `MEM_FREE` space, which `VirtualQuery` reports as `PAGE_NOACCESS`, and the handler treated that as a guard page and skipped the walk. Only `PAGE_GUARD` is a stack-growth probe.

**The unwinder had to match what the engine ships.** `_Unwind_Backtrace` (libbacktrace's `simple.c`) added a `libgcc_s_seh-1.dll` dependency the extension does not ship. Replacing it with ntdll's `RtlLookupFunctionEntry` + `RtlVirtualUnwind` over the faulting `CONTEXT` required no new DLLs (the final import table is KERNEL32 + `api-ms-win-crt-*` + ntdll).

**The handler had to survive its own fault.** A 250-line single function triggered a GCC `internal compiler error` in CI's CodeView writer, so it was split into small functions. Then the handler was made stack-safe: take the `busy` lock before any call, cache the log path and open the file handle at install, and reserve emergency stack with `SetThreadStackGuarantee`.

That finally captured the **pre-unload** `--editor --quit` crash: `HashMap::_insert` ← `DocTools::generate` (`this = 0x8`) ← `EditorHelp::_gen_extensions_docs` ← `MessageQueue::_call_function` ← `MessageQueue::flush`.

**The suite crash was different, and silent even for a stack-safe handler.** Raw exit codes from `Start-Process` (not bash's flattened `139`) showed `0xC0000005` — a real access violation, not a stack overflow or fail-fast. The handler logged the pre-unload fault in the same job but nothing for the suite or `editor --quit-after 12`. That means the fault happens where the handler no longer exists: after the engine unmaps the extension. Godot's `Main::cleanup` does `unregister_core_types()` → `memdelete(gdextension_manager)` → `~GDExtension` → `FreeLibrary`.

**The decisive experiment was to remove the suspected operation, not to observe it.** Pinning the DLL at load (`GetModuleHandleExA` with `GET_MODULE_HANDLE_EX_FLAG_PIN`) made both runs exit `0` in the same job where the unpinned runs still exited `0xC0000005`:

| run | unpinned | pinned |
| --- | --- | --- |
| suite `--quit-after 12` | `0xC0000005` | `0x00000000` |
| editor `--quit-after 12` | `0xC0000005` | `0x00000000` |

## Root cause

Before (crashing):

```c
/* Main::cleanup, engine side */
unregister_core_types();           /* memdelete(gdextension_manager)          */
                                   /*   -> ~GDExtension -> FreeLibrary(dll)   */
                                   /* dll unmapped; Go sysmon/GC threads and  */
                                   /* in-module exception handlers now point  */
                                   /* into unmapped memory -> 0xC0000005      */
```

After (fixed):

```go
// pkg/core/module_windows.go, at package init
GetModuleHandleExA(
    GET_MODULE_HANDLE_EX_FLAG_PIN|GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS,
    &godotgo_pin_module, &mod)     // FreeLibrary can no longer unmap the module
```

A Go c-shared library is not unloadable: its runtime owns threads and handlers that `FreeLibrary` does not stop. Pinning keeps the code mapped until process exit; per-level deinitialization still runs normally. The consequence — no hot-reload/unload on Windows — is documented in `AGENTS.md` and the `extension-library-lifetime` spec.

## What remains unresolved

- **`--editor --quit` crashes for a different, engine-side reason.** The doc-cache worker queues `_gen_extensions_docs`; `EditorHelp`'s predestructor sets `doc = nullptr`; cleanup's `message_queue->flush()` then calls `DocTools::generate` with `this = 0x8`. That is Godot's bug, not the extension's (it only shifts the startup timing). `Gen Test Project Files` keeps `continue-on-error` until the engine guards the deferred call; the workaround is to let the doc cache finish before quitting (`--quit-after`).
- **Hot-reload on Windows is unsupported** by construction. The engine's unload is now a no-op there; restarting the process is the supported path.

## Learnings

1. **A VEH cannot report a fault that happens after its own module is unmapped.** "The handler never ran" and "the handler is gone" look identical from the log; the raw exit code and the engine's cleanup order separate them.
2. **`$?` = 139 is not a diagnosis.** MSYS bash flattens access violations, stack overflows, and more into 139; `Start-Process`'s raw `0xC0000005` was what ruled out a stack overflow.
3. **When observers perturb the bug, test the mechanism instead.** The pin experiment took one dispatch and answered the question that three external capture strategies could not.
4. **Name the unmapped-module hypothesis early.** The engine's `unregister_core_types` → `~GDExtension` → `FreeLibrary` path was visible in the source from the start; checking it before building more capture machinery would have saved several iterations.
5. **Small handler functions are a build requirement here**, not just style: GCC's CodeView writer ICEs on a large function in this toolchain.
