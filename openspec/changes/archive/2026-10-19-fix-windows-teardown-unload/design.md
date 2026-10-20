# Design

## Context

See `proposal.md` — Why. The relevant engine shutdown order (`main/main.cpp`):

1. `GDExtensionManager::shutdown()` (main-loop shutdown callback).
2. `message_queue->flush()`, `delete_main_loop()`, script/renderer teardown.
3. `unregister_core_extensions()` (level `CORE` deinit).
4. `memdelete(engine)`, then `unregister_core_types()`.
5. `unregister_core_types()` deletes `ClassDB` **before** `gdextension_manager` (`core/register_core_types.cpp`).
6. `memdelete(gdextension_manager)` drops the extension `Ref`s; `~GDExtension` calls `close_library()`, which is `FreeLibrary` on Windows (`core/extension/gdextension.cpp`).

A Go c-shared library is not unloadable in practice: the Go runtime owns OS threads (sysmon, GC workers) that outlive `FreeLibrary`, and its in-process exception handlers live in the same module. Once the module is unmapped, both a runtime thread resuming and a stale exception-handler entry are access violations outside the module.

Measured evidence (dispatch run 39373187014, Godot 4.7 stable, Windows runner):

| run | module pinned | raw exit |
| --- | --- | --- |
| suite `--quit-after 12` | no | `0xC0000005` |
| suite `--quit-after 12` | yes (`GODOT_GO_PIN_DLL=1`) | `0x00000000` |
| editor `--quit-after 12` | no | `0xC0000005` |
| editor `--quit-after 12` | yes | `0x00000000` |

The in-process crash logger wrote nothing for the unpinned suite/editor faults even after it was made stack-safe (lock taken first, log path and file handle captured at install, `SetThreadStackGuarantee`), while it did capture the *pre-unload* `--editor --quit` DocTools fault. That places the silent fault after `FreeLibrary`, not inside engine cleanup before it. External tools (gdb, procdump, WER LocalDumps) were also silent for the same reason: the exception is raised in a process state whose handlers are gone or whose module is unmapped.

## Goals / Non-Goals

**Goals:**

- Make every Windows run that loads the extension reach normal process exit, so `make test` and the CI gate can be trusted again.
- Keep the level deinitialization contract untouched.
- Keep the fix Windows-only and dependency-free.

**Non-Goals:**

- Making extension unload/reload work on Windows; the Go runtime cannot stop its threads on demand.
- Fixing the engine's `--editor --quit` deferred-doc race (it is a genuine engine use-after-free before unload; the `Gen Test Project Files` step stays tolerated).
- Capturing post-unmap stacks; the fix removes the fault instead of observing it.

## Decisions

### D1: Pin the module at load (`GetModuleHandleEx` with `GET_MODULE_HANDLE_EX_FLAG_PIN`)

`pkg/core/module_windows.go` (build tag `windows`) calls a small cgo helper at package initialization:

```c
GetModuleHandleExA(
    GET_MODULE_HANDLE_EX_FLAG_PIN | GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS,
    (LPCSTR)(uintptr_t)&godotgo_pin_module, &mod);
```

`PIN` adds a reference that only process exit releases, so the engine's `FreeLibrary` cannot unmap the module. The helper returns `BOOL`; a failure is logged as a warning and leaves the pre-existing behavior.

Alternatives considered:

- **One extra `LoadLibrary` at init.** Same effect in the common case, but a normal reference can still be dropped by any additional `FreeLibrary`, and it does not express "never unload". Rejected.
- **Patch the engine / custom Godot build.** Not shippable for users and affects every GDExtension. Rejected.
- **Stop Go runtime threads before unload.** No supported API; `runtime.GOMAXPROCS`/`LockOSThread` do not cover sysmon. Rejected.
- **Keep the crash logger in a second, never-unloaded helper DLL.** Would turn the silent fault into a captured stack, but still crashes the process. Useful only as diagnostics; not a fix. Rejected for this change.

### D2: Pin unconditionally on Windows

The module cannot be safely unmapped regardless of run mode, so this is not opt-in. The diagnostics-only `GODOT_GO_PIN_DLL` switch used to prove the mechanism is removed when the code lands.

### D3: Leave deinitialization alone

Level deinit (`EDITOR` → `SCENE` → `SERVERS` → `CORE`) still runs; only the unmap is prevented. Keeping the module mapped does not keep extension *state* alive: instances, classes, and callbacks are already torn down by those callbacks, and `ClassDB` is destroyed before the manager.

### D4: Re-enable the Windows `Test` gate in the same change

The fix is exactly what the gate was waiting for. `Test` loses `continue-on-error`; `Gen Test Project Files` keeps it because its `--editor --quit` crash is the separate engine DocTools race.

## Risks / Trade-offs

- **Extension hot-reload cannot replace code on Windows** → the engine's unload no longer frees the module, so a "reload" keeps the old code. It already crashed before; document the restart requirement as the supported path.
- **Pinned module keeps the engine's temporary `~libgodotgo*.dll` copy mapped** → the file stays locked until process exit. Godot creates this copy per process; the OS releases it at exit, so no cleanup change is needed.
- **`GetModuleHandleEx` fails** (unexpected) → the warning is logged and the old crash returns; the CI gate surfaces it. No silent downgrade.
- **Diagnostics now lose post-unmap evidence** → intentional: there is no post-unmap fault left to observe.

## Migration Plan

No migration. Windows builds pin automatically; Linux/macOS are untouched. Rollback is deleting `pkg/core/module_windows.go` (and restoring `continue-on-error` if the crash must be tolerated again).
