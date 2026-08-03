## Context

The GDExtension API has evolved through multiple struct versions for class registration. Godot 4.7 introduced `GDExtensionClassCreationInfo6` (using `CreateInstance3` and registered via `classdb_register_extension_class6`) and deprecated `CreationInfo4`/`CreateInstance2`/`classdb_register_extension_class4`. Since the project targets Godot 4.7.2, it should use the latest API where available.

The current Go-side class creation callback (`GoCallback_ClassCreationInfoCreateInstance`) passes a raw C string as class userdata and uses a `SyncMap` to track live instances. godot-cpp uses `cgo::handle`-like mechanisms via its own class pointer system. This design follows that pattern using Go's `cgo.Handle`.

See proposal.md for the full motivation.

## Goals / Non-Goals

**Goals:**
- Use `cgo.Handle` for per-instance callback lookups (replaces `SyncMap` InstanceID→Object mapping from origin/main's FreeInstance)
- Retain `Internal.GDClassInstances` SyncMap as a re-entrant `FreeInstance` guard
- Introduce `ClassUserdata` struct and `SetConstructInfo` pattern (godot-cpp naming)
- Add `NewGDExtensionClassCreationInfo6` constructor for future use

**Non-Goals:**
- Migrate to `CreationInfo6` + `RegisterExtensionClass6` (deferred — unavailable in binary)
- Delete `cgo.Handle` references in `FreeInstance` (intentionally kept to avoid corruption)
- Add compile-time or runtime fallbacks to older API versions
- Refactor virtual method dispatch, signal handling, or property registration

## Decisions

### Decision 1: Target CreationInfo6, not CreationInfo5

`CreationInfo5` is a typedef alias for `CreationInfo4` and was also deprecated in 4.7. `CreationInfo6` is the only non-deprecated struct. Two struct fields change: `create_instance_func` changes from `GDExtensionClassCreateInstance2` to `GDExtensionClassCreateInstance3`. Both have the same C signature (`void*, GDExtensionBool`), but `CreateInstance3` expects RefCounted subtypes to already have refcount=1 after `ConstructObject3`.

**Alternatives considered:** None. Targeting any deprecated struct is incorrect for a 4.7.2 baseline.

### Decision 2: ClassUserdata Struct for Per-Class Userdata

Previously, a C string (class name) was passed as `class_userdata`. The `ClassUserdata` struct bundles:

```go
type ClassUserdata struct {
    callbacks  GDExtensionInstanceBindingCallbacks
    className  string
    createFunc GDClassGoConstructor // func() GDClass
    typeData   reflect.Type
}
```

This enables structured access in the `CreateInstance2` callback without string parsing and carries the binding callbacks to avoid an extra SyncMap lookup in the hot creation path.

**Alternatives considered:** Keep raw C string + SyncMap lookups. Rejected because it adds unnecessary hash map overhead per instance creation and mixes data management concerns.

### Decision 3: cgo.Handle for Instance Lifetime Tracking

The old design maintained `Internal.GDClassInstances` (a `SyncMap[GDObjectInstanceID, GDClass]`) for looking up instances by ID, checked in `FreeInstance`. The new design stores a `cgo.Handle` to the `WrappedClassInstance` in `ObjectSetInstance` and retrieves it directly in callbacks — no hash map lookup needed for instance access.

**Handle deletion is intentionally omitted** from `FreeInstance`. Deleting the Handle invalidates the raw pointer that Godot holds internally. If Godot issues any subsequent callback with the stale pointer, `cgo.Handle` lookup returns garbage, causing heap corruption (glibc "corrupted double-linked list"). Origin/main never deleted Handles; we follow the same pattern.

**Alternatives considered:** Delete Handle on free. Rejected because of the corruption risk on re-entrant callbacks.

### Decision 4: ClassdbConstructObject2 (keep existing)

`ConstructObject3` (since 4.7) exists but is not yet used because it's tied to the `CreateInstance3`/`CreationInfo6` path which is deferred. The current code continues to use `ConstructObject2`.

**Alternatives considered:** Switch to ConstructObject3. Rejected — would introduce inconsistent refcount semantics (RefCounted=1) without the matching `CreateInstance3` callback.

### Decision 5: Defer CreationInfo6 + RegisterExtensionClass6 + ClassUserdata as class_userdata

During implementation, two blockers were discovered:

1. **`classdb_register_extension_class_6` is unavailable** in the Godot 4.7.2.rc binary. The FFI wrapper loads the proc address but Godot returns an error. This means `CreationInfo6` cannot be used for registration even though the struct and constructor exist.

2. **`ClassUserdata`/`cgo.Handle` as `class_userdata` causes `morestack on g0` crash.** When the `CreateInstance` callback unpacks a `cgo.Handle` from `class_userdata` and calls functions that trigger `runtime.Pinner.Pin()` (e.g., `NewStringNameWithLatin1Chars` → `AsGDExtensionConstStringNamePtr`), the Go runtime crashes with `fatal error: morestack on g0`. This happens because the callback runs on the system stack (g0) before a proper goroutine is set up, and `Pinner.Pin` internally calls `systemstack` which cannot grow g0's stack.

**Resolution:** The implementation uses:
- `CreationInfo4` + `RegisterExtensionClass4` for registration (available in the binary)
- C string as `class_userdata` (avoids the g0 crash)
- String-based `CreateGDClassInstance2(tn string)` called from the callback (avoids the g0 crash by using the origin/main `WrappedPostInitialize2` path which handles `Pinner.Pin` correctly)
- `cgo_classcreationinfo_createinstance2` C function cast as `GDExtensionClassCreateInstance2` type

The `NewGDExtensionClassCreationInfo6` constructor, `ClassUserdata` struct, generic `CreateGDClassInstance[T]`, and `SetConstructInfo` are all defined and available for future use when the blockers are resolved.

**Alternatives considered:** None viable. The g0 crash is a Go runtime limitation that requires either a Go runtime fix or a different approach to pinning in cgo callbacks.

### Decision 6: GDClassInstances SyncMap as Re-Entrant FreeInstance Guard

`Internal.GDClassInstances` is retained (not replaced) as a re-entrant guard in `FreeInstance`. The callback uses `defer Destroy()` so the SyncMap check runs before `Destroy` fires. If `Destroy` triggers a re-entrant `FreeInstance` call (by releasing the last Godot reference), the second invocation fails the SyncMap lookup and panics with a clear error instead of silently double-invoking the C++ destructor and corrupting the C heap.

**Alternatives considered:** Drop the SyncMap entirely. Rejected — the guard is necessary because Godot may invoke `FreeInstance` re-entrantly through `Destroy` in specific edge cases (object removal from scene tree with refcount reaching zero).

### Decision 7: pnr.Pin Callback Pointers in GDClassRegisterInstanceBindingCallbacks

Go 1.25+ enforces stricter cgo pointer rules. The binding callback pointers (`cgo_gdclass_binding_create_callback` et al.) are Go pointers to cgo-generated C thunks stored in a `GDExtensionInstanceBindingCallbacks` C struct. Without `pnr.Pin`, the Go runtime's cgo checker (`cgocheck=1`) may flag these as invalid, potentially corrupting the C-level `GDExtensionInstanceBindingCallbacks` struct in memory. The `pnr.Pin` calls (present in origin/main, briefly commented out) are restored.

### Decision 8: AVX-512 Disabled for Valgrind Compatibility

Go 1.25+ emits AVX-512 ZMM instructions in runtime functions (e.g., `memmove`). These are incompatible with Valgrind's VEX IR translator and with certain glibc builds. Two flags disable AVX-512:

- `GOAMD64=v3` — caps the Go compiler at AVX2 ISA (avoids compiler-generated AVX-512)
- `-ldflags="-X runtime.godebugDefault=cpu.avx512f=off"` — prevents runtime assembly from selecting AVX-512 paths at startup

These are set in the `build` and `build-full` Makefile targets. `CGO_CFLAGS` also includes `-mno-avx512f` to prevent LLVM from auto-vectorizing C code with AVX-512.

## Risks / Trade-offs

- **[Risk] `classdb_register_extension_class5` is known to be unsupported** in the current FFI wrapper (listed in AGENTS.md known issues). Using `class6` avoids this entirely.  
  → **Mitigation**: `RegisterExtensionClass6` is already generated and available; no known issues.

- **[Risk] Re-entrant FreeInstance** — if Godot calls `FreeInstance` re-entrantly through `Destroy()`, the second call would double-invoke the C++ destructor.  
  → **Mitigation**: `Internal.GDClassInstances` SyncMap guards against this; the second call panics instead of silently corrupting memory.

- **[Risk] ClassdbConstructObject3** returns objects with refcount=1 for RefCounted types. If the caller does not properly manage this, objects leak.  
  → **Mitigation**: Deferred until `CreationInfo6`/`RegisterExtensionClass6` migration. Currently using `ConstructObject2`.

- **[Risk] Deferred v6 migration** means the codebase still uses deprecated APIs.  
  → **Mitigation**: The `NewGDExtensionClassCreationInfo6` constructor and `ClassUserdata` struct are ready; migration is a single commit once `classdb_register_extension_class_6` becomes available.
