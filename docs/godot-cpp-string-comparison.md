# Godot-CPP vs Godot-Go StringName Implementation Analysis

## Executive Summary

Godot-Go previously reported orphaned StringName warnings during `make test` due to Go-side lifecycle bugs: dangling pointers stored in `ClassInfo` and Go-GC memory movement during registration calls. The fix stores `StringName` by value, pins values at call sites, and allocates `GDExtensionClassMethodInfo` in C heap. This aligns godot-go's lifecycle model with godot-cpp's: **Godot only reads the 8-byte `StringName` value during registration, so the caller's copy stays valid as long as it is pinned for the duration of the call.**

---

## Memory Layout Comparison

### godot-cpp Pattern

**File**: `/home/pcting/workspace/godot-cpp/gen/include/godot_cpp/variant/string_name.hpp`

```cpp
class StringName {
    uint8_t opaque[8];  // Inline storage - 8 bytes on stack/struct
    _native_ptr() returns pointer to opaque data
};
```

**Key Characteristics**:
- **Inline Storage**: `StringName` stores data directly as `uint8_t opaque[8]`
- **No External Allocation**: No separate heap allocation for StringName data
- **Pointer Semantics**: `_native_ptr()` returns pointer to inline opaque data

**Destruction Flow**:
```cpp
StringName::~StringName() {
    if (unique) {
        internal_destructor();  // Unrefs the interned _Data entry
    }
}
```

### godot-go Pattern

**File**: `pkg/builtin/builtinclasses.gen.go`

```go
type StringName [8]uint8  // Inline storage, 8 bytes
```

**Key Characteristics**:
- **Inline Storage by Value**: `[8]uint8` value, storable in Go structs
- **Explicit Destruction**: `Destroy()` unrefs the interned `_Data` entry
- **Pinning Required**: Go's GC can move heap-allocated values; `runtime.Pinner` keeps the 8 bytes stable during a GDExtension call
- **C-Heap Allocation for Structs**: structs containing C pointers to Go heap (e.g. `GDExtensionClassMethodInfo`) are allocated via `C.malloc()`
---

## Why the Value-Storage Pattern Is Safe

Godot's `StringName` copy constructor **only reads** the source's 8 bytes and increments the refcount on the shared interned `_Data`. It never writes back to caller memory:

```cpp
StringName::StringName(const StringName &p_name) {
    _data = nullptr;
    if (p_name._data && p_name._data->refcount.ref()) {
        _data = p_name._data;  // Write to GODOT'S local only
    }
}
```

The same is true of every `_register_extension_class_*` function — they all copy the `StringName` out of the caller's memory and store their own refcounted copy. See [stringname-mutation-analysis.md](stringname-mutation-analysis.md) for the complete mutation matrix.

---

## ClassDB Registration — Current Pattern

### ClassDBAddProperty (godot-go, fixed)

**File**: `pkg/core/classdb.go`

```go
func ClassDBAddProperty(
	inst GDClass,
	p_property_type GDExtensionVariantType,
	p_property_name string,
	p_setter string,
	p_getter string,
) {
	ci, ok := Internal.GDRegisteredGDClasses.Get(cn)
	// ...
	propName := NewStringNameWithLatin1Chars(pn)
	defer propName.Destroy()
	hint := NewStringWithUtf8Chars("")
	defer hint.Destroy()
	pnr.Pin(&hint)

	prop_info := NewGDExtensionPropertyInfo(
		ci.NameStringName.AsGDExtensionConstStringNamePtr(),
		p_property_type,
		propName.AsGDExtensionConstStringNamePtr(),
		uint32(PROPERTY_HINT_NONE),
		hint.AsGDExtensionConstStringPtr(),
		uint32(PROPERTY_USAGE_DEFAULT),
	)
	pnr.Pin(&prop_info)

	// Class name is stored by value in ClassInfo — copy to an isolated
	// local, pin it, then pass it. Godot reads the 8 bytes and copies.
	cnName := ci.NameStringName
	pnr.Pin(&cnName)
	pnr.Pin(&propName)
	pnr.Pin(&snSetterGDName)
	pnr.Pin(&snGetterGDName)
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertyIndexed(
		FFI.Library,
		cnName.AsGDExtensionConstStringNamePtr(),
		&prop_info,
		snSetterGDName.AsGDExtensionConstStringNamePtr(),
		snGetterGDName.AsGDExtensionConstStringNamePtr(),
		-1,
	)
	// defer Destroy()s run — safe because Godot already took its own copy
}
```

**Why It Works**:
1. `ClassInfo` stores `StringName` by **value** (`NameStringName`), so no dangling pointer to a dead local.
2. Values are copied to isolated locals and **pinned** (`pnr.Pin`) so the GC cannot move them during the call.
3. Godot **copies** the StringName data immediately during the interface call.
4. `Destroy()` after the call is safe — Godot holds its own refcounted copy.
5. The local-copy-isolate-then-pin pattern satisfies the Go 1.26 cgo checker, which walks the entire parent struct and would panic on unpinned Go pointers in `ClassInfo`'s maps/slices.

### GDExtensionClassMethodInfo — C Heap Allocation

`GDExtensionClassMethodInfo` contains C pointers to Go heap (`name`, `arguments_info`, `default_arguments`). The Go cgo checker cannot track pinning across function call boundaries, so the struct is allocated in C heap:

```go
func NewGDExtensionClassMethodInfo(...) *GDExtensionClassMethodInfo {
    cptr := C.malloc(C.sizeof_GDExtensionClassMethodInfo)
    ret := (*GDExtensionClassMethodInfo)(cptr)
    // ... fill fields ...
    return ret  // C heap pointer — cgo checker treats as foreign memory
}
```

`Destroy()` on the returned struct unrefs the contained StringNames. See `pkg/ffi/class_method_info.go`.

---

## Key Differences Summary

| Aspect | godot-cpp | godot-go |
|--------|-----------|----------|
| **StringName Storage** | `uint8_t opaque[8]` inline | `[8]uint8` value |
| **Allocation** | Stack or struct member | Go heap / by value in structs |
| **Pointer Stability** | Stack frame guarantees it | `runtime.Pinner` at call sites |
| **Structs with C pointers** | Stack, Godot copies during call | `C.malloc()` (C heap) |
| **Godot Copy Timing** | Immediate during API call | Immediate during API call |
| **Destruction** | Automatic on stack unwind | Explicit `Destroy()` per value |
| **Dangling Pointer Risk** | None | None (by-value storage) |

---

## Recommended Pattern for godot-go

1. **Store `StringName` by value** in Go structs that outlive a single call (e.g. `ClassInfo.NameStringName`).
2. **Pin at call sites**: copy the value to an isolated local and `pnr.Pin(&local)` for the duration of the GDExtension call.
3. **C-heap allocation** for structs that hold C pointers to Go heap (`GDExtensionClassMethodInfo`).
4. **Balance every `New*StringName` with `Destroy()`** — the copy ctor refcounts, `Destroy()` unrefs.

This is what `pkg/core/types.go`, `pkg/core/classdb.go`, `pkg/core/method_bind.go`, and `pkg/ffi/class_method_info.go` implement.

---

## Testing Verification

```bash
GODOT=/path/to/godot make test
```

Expected: 43/43 tests pass, no cgo panics, no Go-side orphaned StringName warnings. `make test` exits with zero orphaned StringName warnings — the previous `Image` orphan was traced to a Go-side lifecycle bug in `getObjectInstanceBinding()` and eliminated (see `openspec/changes/fix-orphan-stringname/`).

---

## References

- Godot-CPP StringName: `/home/pcting/workspace/godot-cpp/gen/include/godot_cpp/variant/string_name.hpp`
- Godot-CPP StringName Implementation: `/home/pcting/workspace/godot-cpp/gen/src/variant/string_name.cpp`
- Godot-CPP ClassDB: `/home/pcting/workspace/godot-cpp/src/core/class_db.cpp`
- Godot-Go Property Info: `pkg/ffi/property_info.go`
- Godot-Go ClassDB: `pkg/core/classdb.go`
- Godot-Go StringName analysis: [docs/stringname-mutation-analysis.md](stringname-mutation-analysis.md)
- Godot GDExtension Docs: https://docs.godotengine.org/en/stable/tutorials/scripting/gdextension/gdextension_c_example.html
