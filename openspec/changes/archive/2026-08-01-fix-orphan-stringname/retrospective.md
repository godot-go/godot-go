# Retrospective: Orphan StringName Investigation

## Before vs After

```
BEFORE: Orphan StringName: Example (3), varargs (3), Image (2) — 3 types, 7 refs
AFTER:  Orphan StringName: Image (2) — 1 type, 1 non-static ref
```

43/43 tests pass, 0 failures, no cgo panics or double-frees.

## Root Cause Analysis

### 1. `PropertyInfo.Destroy()` — Switch-Case Bug

`switch/case` only executes the first matching branch — leaking `class_name` and `hint_string`. Changed to independent `if` statements.

```go
// BEFORE — only first non-nil field destroyed
switch {
case cp.name != nil:
    CallFunc_...(cp.name)
case cp.class_name != nil:
    CallFunc_...(cp.name)
// AFTER — all non-nil fields destroyed
if cp.name != nil {
    CallFunc_...(cp.name)
}
if cp.class_name != nil {
    CallFunc_...(cp.name)
}
```

**File:** `pkg/ffi/property_info.go`

---

### 2. Signal Names Destroyed Prematurely

`ClassDBAddSignal` created `StringName` → passed to Godot → immediately destroyed via `defer`. Godot retained a reference, so refcount dropped to 0 prematurely.

**Fix:** Store signal `StringName` in `ClassInfo.SignalNameStringNames map[string]StringName`. Destroy on class unregister.

**Files:** `pkg/core/classdb.go`, `pkg/core/types.go`

---

### 3. Property List Double-Destroy

`ClassInfo.Destroy()` destroyed property list StringNames, but Godot's `GoCallback_ClassCreationInfoFreePropertyList2` already destroys them — causing double-free and intern table corruption.

**Fix:** Skip property list destruction in `ClassInfo.Destroy()`.

**File:** `pkg/core/types.go`

---

### 4. `GoMethodMetadata` StringNames Never Destroyed

`GoMethodMetadata` stored StringName objects for every registered method (method name, return/arg PropertyInfos). `ClassDBUnregisterClass` called `ClassMethodInfo.Destroy()` but never cleaned up `GoMethodMetadata`'s fields.

**Fix:** Added `GoMethodMetadata.Destroy()` that:
1. Copies each StringName to a local variable (isolates from `reflect.Value` in struct)
2. Pins the local (satisfies cgo "Go pointer to unpinned Go pointer" check)
3. Calls `.Destroy()` on each StringName/String

**Files:** `pkg/core/method_bind.go`, `pkg/core/classdb.go`

---

### 5. Variadic Methods — Temporary StringNames

`NewGDExtensionClassMethodInfoFromMethodBind` created temporary StringName objects for variadic methods on every call. Never destroyed.

**Fix:** Create variadic StringNames in `NewGoMethodMetadata`, store in struct fields, reuse instead of creating temporaries. Cleanup in `GoMethodMetadata.Destroy()`.

**Files:** `pkg/core/method_bind.go`, `cmd/generate/gdclassimpl/classes.go.tmpl`

---

### 6. cgo Safety — "Go pointer to unpinned Go pointer"

`GoMethodMetadata` contains `reflect.Value` (Go pointer). Calling `.Destroy()` on its StringName fields triggered cgo panic — the checker walks the entire parent struct, sees the Go pointer, panics.

**Fix:** Copy field to isolated local → pin local → call `.Destroy()`.

```go
methodName := md.gdeMethodNameStringName  // copy [8]uint8 to local
pnr.Pin(&methodName)                       // pin isolated local
methodName.Destroy()                        // safe
```

**File:** `pkg/core/method_bind.go`

---

### 7. Double-Free in `ClassMethodInfo.Destroy()`

`GDExtensionClassMethodInfo.Destroy()` destroyed the same StringNames that `GoMethodMetadata.Destroy()` manages — causing double-unref.

**Fix:** Only call `GoMethodMetadata.Destroy()` during unregistration. Skip `ClassMethodInfo.Destroy()`.

**File:** `pkg/core/classdb.go`

## Architecture: StringName Lifecycle

```
Go Creation (refcount +1)  →  Registration (Godot copy, refcount +1)
  → Unregistration (Go Destroy, refcount -1)  →  Godot memdelete (refcount -1, entry removed)
```

Go-side: Every `NewStringName` has matching `Destroy()` ✅
Godot-side: PropertyInfo copies released on `memdelete` ✅

## Remaining Orphan Was Go-side: `Image` (static: 0, total: 1) — FIXED

The last remaining orphan, `Image`, was initially misdiagnosed as a Godot-side
limitation (`static: 1, total: 2` on Godot 4.6.3). After the 4.7.2 upgrade it
appeared as `Image (static: 0, total: 1)` and was re-investigated.

**Isolation:** `make test` was run three ways against `test/demo/main.gd`:
1. `Image.new()` commented out entirely — no orphan.
2. `Image.new()` present, image never passed to Go — no orphan.
3. `Image.new()` + `example.image_ref_func(image)` — orphan appears.

This proved the orphan is created when a valid RefCounted object crosses into
Go, not by `Image.new()` itself.

**Root cause:** `getObjectInstanceBinding()` (`pkg/builtin/variant.go`) calls
Godot's `object_get_class_name`, which does `memnew_placement` (a placement-new)
on the caller-provided `StringName` storage. Go never called `.Destroy()` on it
(`// defer snClassName.Destroy()` was commented out), leaking one refcount for
the object's class name (`"Image"`).

**Fix:** Restore `defer snClassName.Destroy()` in `getObjectInstanceBinding()`.

**Related latent fix:** The two `object_get_class_name` call sites in
`pkg/core/method_bind_reflect.go` (ptrcall arg decode) passed `NewStringName()`
(pre-initialized) storage that Godot's placement-new then overwrote, leaking the
initial constructor refcount. Changed them to zero-value `StringName{}` storage
(consistent with the variant.go pattern).

**Result:** `make test` reports zero orphaned StringName warnings.

## Files Modified

| File | Change |
|------|--------|
| `cmd/generate/gdclassimpl/classes.go.tmpl` | `defer v.Destroy()` for vararg Variant wrappers |
| `pkg/core/classdb.go` | `GoMethodMetadata.Destroy()` on unregister; removed double-free |
| `pkg/core/classdb_callback.go` | Signal name tracking in `ClassInfo` |
| `pkg/core/method_bind.go` | `GoMethodMetadata.Destroy()`, variadic fields, cgo safety |
| `pkg/core/types.go` | Skip property list destruction; signal name cleanup |
| `pkg/ffi/property_info.go` | `switch/case` → independent `if` in `Destroy()` |
| `pkg/gdclassimpl/classes.gen.go` | Regenerated with vararg fix |
| `pkg/builtin/variant.go` | Restored `defer snClassName.Destroy()` in `getObjectInstanceBinding()` |
| `pkg/core/method_bind_reflect.go` | `StringName{}` (not `NewStringName()`) storage for `object_get_class_name` placement-new |

## Key Learnings

1. **`switch/case` vs `if`**: `switch` only executes first matching branch — use `if` for processing all non-nil fields.
2. **cgo pinning with nested Go pointers**: Copy field to isolated local before pinning — cgo walks the entire parent struct.
3. **StringName interning**: 8-byte handle into global table. Copy = refcount +1, Destroy = refcount -1. Orphan = refcount > static_count at exit.
4. **PropertyInfo lifecycle**: Godot copies PropertyInfo during registration, destroys on `memdelete`. Go manages through `GoMethodMetadata.Destroy()`.
5. **Godot-owned StringName output params**: `object_get_class_name` (and similar `GDExtensionUninitializedStringNamePtr` outputs) placement-construct the StringName in the caller's storage. Pass zero-value `StringName{}` storage and always `Destroy()` it after use — never pre-construct with `NewStringName()`.
6. **Verify assumptions with isolation tests**: The `Image` orphan appeared Godot-side (`static: 1`) on 4.6.3 but Go-side (`static: 0`) after the 4.7.2 upgrade. Commenting out code paths in `main.gd` pinpointed the real trigger instead of trusting the earlier "Godot-side limitation" conclusion.
