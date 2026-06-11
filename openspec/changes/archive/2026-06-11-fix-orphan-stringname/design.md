## Context

Godot-Go implements `StringName` as a Go `[8]uint8` struct. Godot's `StringName` is an 8-byte pointer into a global interned hash table. When Godot receives a `GDExtensionConstStringNamePtr`, its copy constructor **only reads** the 8 bytes and increments the refcount on the shared `_Data`. Godot never writes back to caller memory during registration (see `docs/stringname-mutation-analysis.md` for full analysis).

Two root causes produce orphaned StringName warnings and instability:

1. **Dangling pointers in `ClassInfo`**: `NewClassInfo` creates a local `StringName`, stores `&local` as `NameAsStringNamePtr` in `ClassInfo`, and returns. The local goes out of scope — the pointer dangles.
2. **GC memory movement**: Go's GC can relocate heap-allocated `[8]uint8` values during registration calls, producing stale pointers.

## Goals / Non-Goals

**Goals:**
- Eliminate dangling pointers in `ClassInfo` that reference freed local variables.
- Ensure `StringName` values passed to GDExtension interface calls have stable, GC-pinned addresses for the duration of the call.
- Pass Go 1.26 cgo safety checks (`cgocheck=1`) without panics on "Go pointer to unpinned Go pointer".
- Achieve 43/43 passing tests with no crashes.

**Non-Goals:**
- Eliminating Godot's internal "Orphan StringName" warnings at exit — Godot holds its own refcounts on interned StringName entries, which it reports as orphans on shutdown. This is Godot-side behavior, not addressable from Go.
- Changing the general-purpose Go `StringName` API for end-users.
- C-side string allocation or caching.
- Modifying generated code in `*.gen.*` files.

## Decisions

### 1. Store `StringName` by Value in `ClassInfo`

Replace pointer fields with inline value fields:

```go
type ClassInfo struct {
    NameStringName       StringName  // [8]uint8 — lives as long as ClassInfo lives
    ParentNameStringName StringName
    // Remove: NameAsStringNamePtr, ParentNameAsStringNamePtr
}
```

- **Rationale**: Eliminates dangling pointers entirely. The `[8]uint8` value lives inside the struct that owns it, not a separate local variable that goes out of scope. This mirrors godot-cpp's pattern where `ClassInfo` stores `StringName` by value.
- **Alternative considered**: C-managed memory allocation (malloc/free). Rejected because it adds manual lifecycle complexity and a C wrapper layer when Go value storage + pinning achieves the same result safely.

### 2. Pin at Registration Call Sites

Use `runtime.Pinner` at each call site where a `StringName` is passed to Godot:

```go
func ClassDBAddProperty(inst GDClass, ...) {
    ci := Internal.GDRegisteredGDClasses.Get(cn)
    pnr.Pin(&ci.NameStringName)  // prevent GC movement during call

    CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertyIndexed(
        FFI.Library,
        ci.NameStringName.AsGDExtensionConstStringNamePtr(),  // Godot reads 8 bytes, copies
        &prop_info,
        // ...
    )
    // Pin scope ends here — GC may relocate after this point, but Godot has its own copy
}
```

- **Rationale**: Pin scope is the registration call only. Godot's copy ctor increments refcount during the call, so after the call returns, Godot owns an independent ref. Pinning prevents GC movement only during the read window.
- **Alternative considered**: Persistent pinning (pin once in `NewClassInfo`, never unpin). Rejected because it holds the GC hostage for the lifetime of the extension with no benefit — the value is only accessed during registration calls.

### 3. Eliminate Redundant `StringName` Creation in `ClassDBAddProperty`

Currently `ClassDBAddProperty` creates a new `StringName` for the class name on every call, even though `ClassInfo` already stores it. After storing by value, read directly:

```go
// Before (creates temp, stores dangling ptr, passes temp)
className := NewStringNameWithLatin1Chars(cn)
defer className.Destroy()
CallFunc_...(ci.NameAsStringNamePtr, ...)  // dangling!

// After (reads stored value, no temp needed)
pnr.Pin(&ci.NameStringName)
CallFunc_...(ci.NameStringName.AsGDExtensionConstStringNamePtr(), ...)
```

- **Rationale**: Reduces allocations and eliminates the dangling pointer in one change.

### 4. Local-Copy-Isolate-Then-Pin Pattern

You cannot `pnr.Pin(&ci.NameStringName)` directly — `ClassInfo` contains Go pointers (`map[string]*MethodBindAndClassMethodInfo`, `[]GDExtensionPropertyInfo`, etc.), and the Go 1.26 cgo checker walks the entire parent struct when you pass a field address to cgo. It sees those Go pointers and panics with "Go pointer to unpinned Go pointer", even though you pinned the specific field.

**Solution**: Copy the `[8]uint8` value to a local variable, pin the local, then pass it:

```go
// Wrong — cgo checker sees ci's maps/slices
pnr.Pin(&ci.NameStringName)  // PANIC!
CallFunc_...(ci.NameStringName.AsGDExtensionConstStringNamePtr(), ...)

// Correct — isolated local has no Go pointers
cnName := ci.NameStringName  // copy [8]uint8 to local stack
pnr.Pin(&cnName)             // pin the isolated local
CallFunc_...(cnName.AsGDExtensionConstStringNamePtr(), ...)
```

This pattern is used at every call site that accesses `ClassInfo` fields: `ClassDBAddPropertyGroup`, `ClassDBAddPropertySubgroup`, `ClassDBAddProperty`, `ClassDBAddSignal`, `classDBBindMethod`, `CreateGDClassInstance`, and `ClassInfo.Destroy()`.

### 5. C Heap Allocation for `GDExtensionClassMethodInfo`

`NewGDExtensionClassMethodInfo` constructs a Go struct containing C pointers to Go heap memory (`name`, `arguments_info`, `default_arguments`, etc.). Even when pinned with `pnr.Pin(cmi)`, the Go 1.26 cgo checker panics when the struct is passed through the generated wrapper function to C — it cannot track pinning across function call boundaries.

**Solution**: Allocate in C heap via `C.malloc()`:

```go
func NewGDExtensionClassMethodInfo(...) *GDExtensionClassMethodInfo {
    cptr := C.malloc(C.sizeof_GDExtensionClassMethodInfo)
    ret := (*GDExtensionClassMethodInfo)(cptr)
    (*C.GDExtensionClassMethodInfo)(ret).name = (C.GDExtensionStringNamePtr)(name)
    // ... fill remaining fields ...
    return ret  // C heap pointer, cgo checker treats as foreign
}
```

The cgo checker sees a C heap pointer and allows it through. Godot reads the struct fields during the registration call and copies the data it needs. After the call returns, the C heap memory can be safely freed (or leaked for the extension lifetime).

**Trade-off**: `Destroy()` currently does not call `C.free()` on the struct itself. This is safe for single-load extensions (typical GDExtension lifecycle) but would leak on unload+reload cycles. Tracked as follow-up task 5.3.

### 6. `GDExtensionPropertyInfo` Remains a Go Struct

`NewGDExtensionPropertyInfo` continues to construct a Go struct. The struct itself lives on the Go stack (or heap) and is passed by pointer to Godot. Godot reads the struct fields during the call — it doesn't store the pointer.

- **Rationale**: Godot reads `GDExtensionPropertyInfo` by value during the registration call (copy ctor on `p_info->name`, `p_info->class_name`). It never stores the pointer to the struct itself. Only the 8-byte values inside need to be stable during the call, which pinning provides.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│  Before (broken)                                                     │
│                                                                      │
│  NewClassInfo():                                                     │
│    nameStringName := NewStringName(...)  ← local on stack/heap      │
│    return ClassInfo{                                                │
│      NameAsStringNamePtr: &nameStringName  ← DANGLING POINTER       │
│    }                                                                │
│    ← nameStringName goes out of scope, pointer is invalid           │
│                                                                      │
│  ClassDBAddProperty():                                               │
│    ci.NameAsStringNamePtr → reads stale memory 💥                   │
├─────────────────────────────────────────────────────────────────────┤
│  After (fixed)                                                       │
│                                                                      │
│  NewClassInfo():                                                     │
│    nameStringName := NewStringName(...)                              │
│    return ClassInfo{                                                │
│      NameStringName: nameStringName  ← VALUE COPIED INTO STRUCT     │
│    }                                                                │
│    ← struct owns the value, lives as long as ClassInfo               │
│                                                                      │
│  ClassDBAddProperty():                                               │
│    cnName := ci.NameStringName  ← copy to isolated local            │
│    pnr.Pin(&cnName)          ← GC can't move during call            │
│    CallFunc_...(cnName.AsGDExtensionConstStringNamePtr(), ...)      │
│    // Godot copy ctor: reads 8 bytes, increments refcount            │
│    // Pin scope ends, Godot owns its own ref ✓                      │
│                                                                      │
│  GDExtensionClassMethodInfo:                                         │
│    cptr := C.malloc(...)       ← C heap allocation                  │
│    CallFunc_...(cmi)           ← cgo checker allows C heap ptr      │
│    // Godot reads struct fields, copies what it needs                │
└─────────────────────────────────────────────────────────────────────┘

```

### cgo Safety Model

```
┌─────────────────────────────────────────────────────────────────────┐
│  Go 1.26 cgo checker (cgocheck=1)                                   │
│                                                                      │
│  pnr.Pin(&ci.NameStringName) → ❌ FAILS                             │
│  (ci has map[string]*... and []GDExtensionPropertyInfo fields)       │
│                                                                      │
│  cnName := ci.NameStringName                                        │
│  pnr.Pin(&cnName)            → ✅ PASSES                            │
│  (cnName is isolated [8]uint8, no Go pointers)                       │
│                                                                      │
│  cmi := NewGDExtensionClassMethodInfo(...) → ❌ FAILS               │
│  (Go struct contains C pointers to Go heap)                          │
│                                                                      │
│  cmi := C.malloc(...)         → ✅ PASSES                            │
│  (C heap pointer, cgo checker treats as foreign memory)              │
└─────────────────────────────────────────────────────────────────────┘
```

## Risks / Trade-offs

- **Risk**: Pin discipline must be followed at every call site. Forgetting to pin risks GC movement.
  - **Mitigation**: Centralize pin logic in helper functions where possible. Code review checks for `Pin` at every `CallFunc_GDExtension...` that passes a `StringName` pointer.

- **Risk**: Pre-existing issue — Godot-owned `StringName` pointers in callbacks (e.g., `classdb_callback.go`) could be accidentally `Destroy()`'d.
  - **Mitigation**: Out of scope for this change. Document the constraint. Consider a follow-up change to add runtime guards or read-only wrapper types.

- **Trade-off**: Slight verbosity increase — every registration call site needs an explicit `Pin` call. This is more explicit than the broken implicit behavior.

## File Changes

| File | Change |
|------|--------|
| `pkg/core/types.go` | `ClassInfo`: replace `NameAsStringNamePtr`/`ParentNameAsStringNamePtr` with `NameStringName`/`ParentNameStringName` value fields. `Destroy()` copies to locals before calling `.Destroy()`. Remove `unsafe` import. |
| `pkg/core/classdb.go` | All 8 registration functions updated: read from `ci.NameStringName`, copy to locals, pin all temporaries (`String`, `StringName`, `PropertyInfo`, `MethodInfo`). Removed redundant `NewStringNameWithLatin1Chars(cn)` allocations. |
| `pkg/core/lib.go` | `CreateGDClassInstance`: use `ci.ParentNameStringName` instead of creating new `StringName`, copy to local, pin before cgo call. |
| `pkg/core/method_bind.go` | `NewGDExtensionClassMethodInfoFromMethodBind`: add `pnr.Pin(&gdMethodNameStringName)` for the local `StringName` used in `NewGDExtensionClassMethodInfo`. |
| `pkg/ffi/class_method_info.go` | `NewGDExtensionClassMethodInfo`: changed from Go stack allocation to C heap (`C.malloc()`). Removed `pnr.Pin(ret)` — no longer needed since memory is C-owned. |
| `pkg/ffi/property_info.go` | No change — `GDExtensionPropertyInfo` is passed by pointer, Godot reads fields during call. |
| `openspec/changes/fix-orphan-stringname/tasks.md` | All 16 tasks marked complete; added Section 5 (follow-up known issues). |
| `openspec/changes/fix-orphan-stringname/proposal.md` | Updated with actual implementation details, known issues, and accurate file list. |
| `openspec/changes/fix-orphan-stringname/design.md` | Updated with local-copy pattern, C heap allocation decision, cgo safety model, and corrected architecture diagram. |
