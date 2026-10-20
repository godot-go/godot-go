# Proposal

## Why

On Windows every Godot run that loads the Go extension faults after a fully clean suite (`1430 passed, 0 failed`) with `STATUS_ACCESS_VIOLATION` (`0xC0000005`). The fault is silent: the in-process vectored handler never sees it because the engine unmaps the extension module during shutdown (`Main::cleanup` → `unregister_core_types` → `~GDExtensionManager` → `FreeLibrary`) and the fault happens afterwards, in code or handlers that lived in the unmapped module. A Go c-shared library cannot be unloaded: its runtime keeps its own threads (sysmon, GC workers) running, and unmapping the module turns their next instruction into an access violation. Pinning the module with `GetModuleHandleEx(..., GET_MODULE_HANDLE_EX_FLAG_PIN, ...)` makes both Windows repros exit cleanly (measured: raw `0xC0000005` unpinned, `0x00000000` pinned, suite and editor run).

## What Changes

- On Windows, the extension pins its own module in the process at load time, so the engine's shutdown `FreeLibrary` never unmaps it. Per-level deinitialization still runs normally; only the unmap is prevented.
- Windows CI stops tolerating a crashing test run: the `Test` step loses `continue-on-error` and gates the job again.
- Document the platform constraint: on Windows the module stays mapped until process exit, so extension hot-reload/unload is unsupported there.

## Capabilities

### New Capabilities

- `extension-library-lifetime`: platform behavior of the extension's native module — when it may be unmapped, and what an unload request means for the Go runtime.

### Modified Capabilities

(none — initialization-level deinitialization semantics are unchanged; only the Windows unmap is prevented)

## Impact

- `pkg/core` gains a Windows-only module pin at package initialization.
- `.github/workflows/ci_windows.yaml` tests are gated again; no test logic changes.
- No public Go API change, no Linux/macOS behavior change.

## Non-goals

- Fixing the engine's `--editor --quit` DocTools deferred-call race; that separate engine bug stays tolerated in `Gen Test Project Files`.
- Supporting extension hot-reload/unload on Windows; the Go runtime cannot be safely unloaded.
- Changing the diagnostics-only crash workflow.
