# Godot-cpp vs Godot-go StringName/String Memory Management Notes

## Date: 2026-04-11

## Status: RESOLVED

The orphaned StringName warnings were caused by Go-side lifecycle bugs (dangling pointers in `ClassInfo`, GC memory movement, premature destruction, and a `switch/case` bug in `PropertyInfo.Destroy()`). The fix stores `StringName` by value, pins at call sites, and allocates `GDExtensionClassMethodInfo` in C heap. See [stringname-mutation-analysis.md](stringname-mutation-analysis.md) for the authoritative analysis and the fix-orphan-stringname OpenSpec change for implementation details.

## Initial Observations

### Problem Statement
Orphaned StringName warnings appeared when running `make test`:
- 29 from `Example` class methods
- 2 from `Image` type
- 6 from `custom_signal`

These warnings indicated StringNames being destroyed after Godot had already taken its copy, or being destroyed twice.

### Initial Root Cause Hypothesis

The initial hypothesis was that godot-go passed Go-heap pointers to Godot and Godot stored them without copying. Investigation (see below) proved this wrong: Godot **copies** `StringName` immediately during every registration call. The real causes were Go-side lifecycle bugs:

1. **Dangling pointers**: `ClassInfo` stored a pointer to a local `StringName` that died when `NewClassInfo` returned.
2. **GC movement**: Go's GC could relocate heap-allocated `[8]uint8` values during registration calls.
3. **Premature destruction**: temporary `StringName`s (e.g. signal names) were `Destroy()`'d while Godot still referenced them.
4. **`switch/case` bug** in `PropertyInfo.Destroy()`: only the first non-nil field was destroyed.

### godot-go Pattern (fixed)

```go
// StringName is [8]uint8 - inline storage
type StringName [8]uint8

// Stored by value in ClassInfo - no dangling pointer
ci.NameStringName  // value field

// Local-copy-isolate-then-pin before every GDExtension call
cnName := ci.NameStringName
pnr.Pin(&cnName)
CallFunc_...(cnName.AsGDExtensionConstStringNamePtr(), ...)
```

### godot-cpp Pattern (reference)

```cpp
// StringName is uint8_t opaque[8] - inline storage
class StringName {
    uint8_t opaque[STRING_NAME_SIZE] = {};
    _FORCE_INLINE_ GDExtensionTypePtr _native_ptr() const {
        return const_cast<uint8_t(*)[STRING_NAME_SIZE]>(&opaque);
    }
};
```

## Key Differences

### 1. Memory Lifetime

| Aspect | godot-cpp | godot-go |
|--------|-----------|----------|
| Allocation | Stack (automatic) | Go heap / by value in structs |
| Pointer stability | Stack frame guarantees it | `runtime.Pinner` at call sites |
| Destruction | Safe after Godot call | Explicit `Destroy()` per value |

### 2. StringName Implementation

**godot-cpp**:
- `uint8_t opaque[8]` inline storage
- `_native_ptr()` returns pointer to opaque data
- RAII ensures cleanup happens in the correct scope

**godot-go**:
- `[8]uint8` value (identical layout)
- `AsGDExtensionConstStringNamePtr()` returns a pointer to the value
- `Destroy()` unrefs the interned entry; must be balanced with every `New*StringName`

### 3. PropertyInfo Handling

Godot copies every `StringName` field out of `GDExtensionPropertyInfo` during the registration call, so the Go struct only needs its 8-byte values pinned for the duration of the call — it does not need C-allocated backing.

## Resolved Investigation Tasks

- [x] Read godot-cpp `string_name.hpp` - Understand StringName structure
- [x] Read godot-cpp `class_db.cpp` - Understand ClassDB registration
- [x] Fix `NewGDExtensionPropertyInfo()` - remove the broken pointer-flow hypothesis
- [x] Fix `ClassDBAddProperty()` - read `ci.NameStringName` by value, pin before the call
- [x] Fix `ClassDBAddSignal()` - retain signal names in `ClassInfo.SignalNameStringNames`
- [x] Create comprehensive comparison document
- [x] Run tests to verify fix (43/43 pass)

## Final Outcome

- **Code fix**: `pkg/core/types.go`, `pkg/core/classdb.go`, `pkg/core/lib.go`, `pkg/core/method_bind.go`, `pkg/ffi/class_method_info.go`, `pkg/ffi/property_info.go`.
- **Verification**: `make test` — 43/43 tests pass, no cgo panics, no Go-side orphaned StringName warnings.
- **Remaining orphan**: Godot reports `Image` (static refs) at exit — engine-internal, not addressable from Go.

See the fix-orphan-stringname OpenSpec change for the full task breakdown and the retrospective.
