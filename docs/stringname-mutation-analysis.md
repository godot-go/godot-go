# StringName Mutation Analysis

## Overview

Analysis of whether Godot writes to caller-provided `StringName` memory during GDExtension registration calls, and the safety implications for storing `StringName` by value in Go.

## Godot StringName Internals

`StringName` is **not a string**. It is an 8-byte pointer to a refcounted interned entry in Godot's global hash table:

```
StringName = [8 bytes] = pointer to _Data

_Data (interned heap):
  - refcount: uint32
  - static_count: uint32
  - name: String
  - hash: uint32
  - prev/next: linked list in bucket
```

Two `StringName` values with the same text share the same `_Data` pointer. Equality is pointer comparison (O(1)).

### The Refcount Dance

```
Caller creates StringName("position"):
  - Godot finds/creates _Data entry in interned table
  - refcount = 1
  - 8 bytes written into caller's buffer

Caller passes StringName to Godot registration:
  - Godot copy constructor reads 8 bytes from caller
  - refcount.ref() → refcount = 2
  - Godot owns its own local copy

Caller calls Destroy():
  - refcount.unref() → refcount = 1
  - Godot's copy remains valid ✓
```

### The Orphan Condition

At Godot shutdown, `StringName::cleanup()` checks:

```cpp
if (d->static_count.get() != d->refcount.get()) {
    print_line("Orphan StringName: " + d->name);
}
```

An orphan means the refcount didn't balance — someone created a refcount but never destroyed it.

## Mutation Analysis

### Godot's Copy Constructor — Read-Only from Caller's Perspective

```cpp
StringName::StringName(const StringName &p_name) {
    _data = nullptr;  // Godot's NEW local starts empty
    if (p_name._data && p_name._data->refcount.ref()) {
        _data = p_name._data;  // Write to GODOT'S local only
    }
}
```

**Key insight**: The copy constructor reads `_data` from the source and writes `_data` into `this` (Godot's local). It never modifies the source's 8 bytes.

### Registration Functions — All Read-Only

Every GDExtension registration function follows the same pattern:

```cpp
void GDExtension::_register_extension_class_property_indexed(
    GDExtensionConstStringNamePtr p_class_name,
    const GDExtensionPropertyInfo *p_info,
    GDExtensionConstStringNamePtr p_setter,
    GDExtensionConstStringNamePtr p_getter,
    GDExtensionInt p_index) {
    
    // All copies — read from caller, write to Godot locals
    StringName class_name = *reinterpret_cast<const StringName *>(p_class_name);
    StringName setter     = *reinterpret_cast<const StringName *>(p_setter);
    StringName getter     = *reinterpret_cast<const StringName *>(p_getter);
    String property_name  = *reinterpret_cast<const StringName *>(p_info->name);
}
```

Verified across all registration functions: `_register_extension_class5`, `_register_extension_class_method`, `_register_extension_class_signal`, `_register_extension_class_property_indexed`, `_register_extension_class_property_group`, `_register_extension_class_property_subgroup`, `_register_extension_class_integer_constant`, `_unregister_extension_class`.

### The Non-Const Cast (gdextension.cpp:196)

```cpp
void update(const GDExtensionClassMethodInfo *p_method_info) {
#ifdef TOOLS_ENABLED
    name = *reinterpret_cast<StringName *>(p_method_info->name);
    //   ^^^ drops const — but operator= writes to THIS, not source
#endif
}
```

`operator=` writes to `this->_data`, not to the source:

```cpp
StringName &StringName::operator=(const StringName &p_name) {
    unref();                    // unref THIS's current _data
    if (p_name._data && p_name._data->refcount.ref()) {
        _data = p_name._data;   // write to THIS, not p_name
    }
    return *this;
}
```

### PropertyInfo Copy from GDExtensionPropertyInfo

```cpp
PropertyInfo(const GDExtensionPropertyInfo &pinfo) :
    type((Variant::Type)pinfo.type),
    name(*reinterpret_cast<StringName *>(pinfo.name)),      // copy ctor — read only
    class_name(*reinterpret_cast<StringName *>(pinfo.class_name)),  // copy ctor — read only
    hint((PropertyHint)pinfo.hint),
    hint_string(*reinterpret_cast<String *>(pinfo.hint_string)),
    usage(pinfo.usage) {}
```

## Complete Mutation Matrix

| Context | Godot Writes to Caller's 8 Bytes? | Example |
|---------|---|---|
| All `classdb_register_extension_class_*` registration calls | **No** | `p_class_name`, `p_setter`, `p_getter`, `p_signal_name`, `p_info->name`, `p_info->class_name` |
| `PropertyInfo(*p_info)` internal copy | **No** | Copy constructor reads, never writes back |
| `object_get_class_name()` output parameter | **Yes** | `r_class_name` — by design, caller provides uninitialized buffer |
| Callback: `GDExtensionClassSet`/`Get` | **No** | `p_name` is Godot-owned, extension reads it |
| Builtin method output (ptrcall) | **Yes** | `r_ret` is caller buffer, Godot writes result |

## Registration Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────────┐
│              Godot Registration Call — Read-Only Copy                     │
│                                                                          │
│  Go Memory (caller)          Godot Stack (callee)                        │
│  ┌────────────────┐          ┌────────────────┐                          │
│  │ [8 bytes]      │─────────▶│ [8 bytes]      │  copy ctor              │
│  │ ci.NameStringName │          │ local class_name │                          │
│  │  (caller's)    │ READ ONLY │  (Godot's)     │                          │
│  └────────────────┘          └────────────────┘                          │
│                                                                          │
│  Both now point to same _Data interned entry:                            │
│                                                                          │
│  ┌──────────────────────────────────┐                                    │
│  │ _Data (interned heap)            │                                    │
│  │ refcount: 2  (was 1, now +1)    │                                    │
│  │ name: "MyClass"                  │                                    │
│  └──────────────────────────────────┘                                    │
│       ▲                  ▲                                               │
│       │                  │                                               │
│  caller's ref       Godot's ref                                         │
│                                                                          │
│  After caller calls Destroy():                                           │
│  refcount: 2 → 1   (Godot's ref still valid)                             │
└──────────────────────────────────────────────────────────────────────────┘
```

## Implications for godot-go

### Why Storing by Value Is Safe

Since Godot never writes to the caller's 8 bytes during registration:

1. **Storing `[8]uint8` by value** in Go structs is safe — Godot treats it as read-only during calls
2. **Pin at call sites** — `runtime.Pinner` prevents GC from moving the memory during the registration call
3. **Refcount is balanced** — copy constructor increments, `Destroy()` decrements, no orphan

### Pre-Existing Risk: Godot-Owned StringNames in Callbacks

```go
// classdb_callback.go — Godot provides pName, we cast to *StringName
func GoCallback_ClassCreationInfoGet(pInstance, pName, rRet) {
    gdName := (*StringName)(pName)  // points to GODOT's memory
    // Safe to READ (e.g., gdName.AsString())
    // DANGEROUS to call gdName.Destroy() — would decrement Godot's refcount!
}
```

This is a separate concern from the registration issue. The callback receives Godot-owned StringNames. Calling `Destroy()` on them would incorrectly decrement Godot's refcount.

### Comparison with Current Approach

| | Current (Pointer) | Option B (Value) |
|---|---|---|
| `ClassInfo` stores | `GDExtensionConstStringNamePtr` (dangling pointer to local) | `StringName` value `[8]uint8` |
| GC movement during call | ✅ Risk (pointer becomes stale) | ❌ Safe with `Pin` |
| GC collection of source | ✅ Risk (local goes out of scope) | ❌ Safe (stored in struct) |
| Dangling pointer | ✅ Bug (`NewClassInfo` returns, local dies) | ❌ No pointer, no dangling |
| Pin requirement | Not needed (pointer is stale anyway) | Required at call sites |
| Memory layout | Indirection through pointer | Inline 8 bytes |

## Sources

- Godot `core/string/string_name.h` — StringName struct, copy constructor, destructor, operator=
- Godot `core/string/string_name.cpp` — refcount management, interned table, cleanup
- Godot `core/extension/gdextension.cpp` — all `_register_extension_class_*` functions
- Godot `core/object/object.h` — `PropertyInfo` copy from `GDExtensionPropertyInfo`
- godot-cpp `src/core/class_db.cpp` — reference implementation of registration pattern
