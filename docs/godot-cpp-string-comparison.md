# Godot-CPP vs Godot-Go StringName Implementation Analysis

## Executive Summary

Godot-Go experiences orphaned StringName warnings during `make test` due to a fundamental difference in memory management patterns between godot-cpp and godot-go. The godot-cpp approach uses **stack-allocated StringName objects with inline storage**, while godot-go uses **heap-allocated StringName objects with explicit destruction**.

**Root Cause**: Orphaned StringName warnings (29 from Example class, 2 from Image type, 6 from custom_signal) occur because godot-go stores Go-heap-allocated StringName pointers in C structs, while godot-cpp uses stack allocation with Godot copying the data immediately.

---

## Memory Layout Comparison

### godot-cpp Pattern (Correct)

**File**: `/home/pcting/workspace/godot-cpp/gen/include/godot_cpp/variant/string_name.hpp`

```cpp
class StringName {
    uint8_t opaque[8];  // Inline storage - 8 bytes on stack/struct
    // ...
    _native_ptr() returns pointer to opaque data
};
```

**Key Characteristics**:
- **Inline Storage**: `StringName` stores data directly as `uint8_t opaque[8]`
- **Small Object**: Fits entirely within containing struct or on stack
- **No External Allocation**: No separate heap allocation for StringName data
- **Pointer Semantics**: `_native_ptr()` returns pointer to inline opaque data

**Constructor Flow**:
```cpp
StringName::StringName(const String &p_name) {
    internal_constructor(p_name);  // Copies String data into opaque[8]
}

StringName::StringName(const char *p_name) {
    internal_constructor_from_cstr(p_name);  // Copies C string into opaque[8]
}
```

**Destruction Flow**:
```cpp
StringName::~StringName() {
    if (unique) {
        internal_destructor();  // Cleans up internal data only
    }
}
```

### godot-go Pattern (Problematic)

**File**: `pkg/ffi/string_name.gen.go`

```go
type StringName struct {
    data [8]uint8  // Inline storage in Go heap object
}

func (sn *StringName) Destroy() {
    // Frees Go heap memory after defer
    internal_destructor()
}
```

**Key Characteristics**:
- **Inline Storage in Go Heap**: `[8]uint8` stored in Go-managed heap object
- **Pointer Passed to C**: `AsGDExtensionConstStringNamePtr()` returns pointer to Go heap
- **Explicit Destruction**: `Destroy()` called explicitly after use
- **Lifetime Mismatch**: Go heap pointer may outlive C usage or cause double-free

**Constructor Flow**:
```go
func NewStringNameWithLatin1Chars(s string) StringName {
    var sn StringName
    // Populates sn.data [8]uint8 on Go heap
    internal_constructor(s)
    return sn  // Returns by value, but may be stored as pointer in C structs
}
```

**Destruction Flow**:
```go
func (sn *StringName) Destroy() {
    // Calls Godot destructor on Go-heap-allocated memory
    internal_destructor()
}
```

---

## Critical Code Flow Comparison

### godot-cpp: ClassDB Registration

**File**: `/home/pcting/workspace/godot-cpp/src/core/class_db.cpp`

```cpp
// Register property
StringName class_name("MyClass");
StringName property_name("my_property");
StringName setter_name("_set_my_property");
StringName getter_name("_get_my_property");

GDExtensionPropertyInfo prop_info;
prop_info.class_name = class_name._native_ptr();  // Pointer to stack opaque
prop_info.name = property_name._native_ptr();     // Pointer to stack opaque
prop_info.setter = setter_name._native_ptr();     // Pointer to stack opaque
prop_info.getter = getter_name._native_ptr();     // Pointer to stack opaque

// Pass to Godot - Godot COPIES data immediately
CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertyIndexed(
    library,
    class_name._native_ptr(),
    &prop_info,
    setter_name._native_ptr(),
    getter_name._native_ptr(),
    -1,
);

// StringName destructors run safely - Godot has already copied the data
```

**Why It Works**:
1. StringName objects live on **stack** of registration function
2. `_native_ptr()` returns pointer to **inline opaque[8]** on stack
3. Godot **copies** the StringName data immediately during interface call
4. Stack StringName destructors run after Godot has copied
5. **No dangling pointers, no double-free**

### godot-go: ClassDB Registration

**File**: `pkg/core/classdb.go:126-149`

```go
// Register property
prop_info := NewGDExtensionPropertyInfo(
    NewStringNameWithLatin1Chars(cn),        // Go heap StringName
    p_property_type,
    NewStringNameWithLatin1Chars(pn),        // Go heap StringName
    uint32(PROPERTY_HINT_NONE),
    NewStringWithUtf8Chars(""),
    uint32(PROPERTY_USAGE_DEFAULT),
)
snSetterGDName := NewStringNameWithLatin1Chars(setter.GdMethodName)
defer snSetterGDName.Destroy()
snGetterGDName := NewStringNameWithLatin1Chars(getter.GdMethodName)
defer snGetterGDName.Destroy()

CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertyIndexed(
    FFI.Library,
    ci.NameAsStringNamePtr,
    &prop_info,  // Contains pointers to Go-heap StringNames
    snSetterGDName.AsGDExtensionConstStringNamePtr(),
    snGetterGDName.AsGDExtensionConstStringNamePtr(),
    -1,
)
// defer runs - Destroy() called on Go-heap StringNames
```

**The Problem**:
1. StringName objects stored in **Go heap**
2. `NewGDExtensionPropertyInfo()` creates C struct containing **pointers to Go heap**
3. Godot may **store pointers without copying** immediately
4. `defer Destroy()` frees Go heap memory
5. **Double-free or use-after-free** when Godot tries to access freed memory

---

## The Broken Function: NewGDExtensionPropertyInfo

**File**: `pkg/ffi/property_info.go`

```go
func NewGDExtensionPropertyInfo(
    p_class_name *StringName,
    p_type GDExtensionVariantType,
    p_name *StringName,
    p_hint GDExtensionPropertyHint,
    p_hint_string *String,
    p_usage GDExtensionPropertyUsageFlags,
) GDExtensionPropertyInfo {
    // Creates C-allocated PropertyInfo struct
    cPropInfo := (*GDExtensionPropertyInfo)(C.malloc(C.sizeof_GDExtensionPropertyInfo))
    defer C.free(unsafe.Pointer(cPropInfo))
    
    // But the pointers inside point to Go-heap StringNames!
    cPropInfo.class_name = (**GDExtensionStringNameData)(unsafe.Pointer(p_class_name))
    cPropInfo.name = (**GDExtensionStringNameData)(unsafe.Pointer(p_name))
    cPropInfo.hint_string = (**GDExtensionStringData)(unsafe.Pointer(p_hint_string))
    
    return *cPropInfo  // Returns copy, but pointers still reference Go heap
}
```

**The Issue**:
- Allocates C memory for the struct
- But copies **Go-heap pointers** into that struct
- When struct is returned/copied, pointers still reference Go heap
- `Destroy()` on Go heap StringNames causes memory issues

---

## Godot-CPP's Correct Pattern in Detail

### StringName Constructor Variants

From `string_name.cpp`:

```cpp
// From String - copies String's internal data into opaque[8]
StringName::StringName(const String &p_name) {
    if (p_name.is_empty()) {
        _data = _empty_data;
    } else {
        _native_ptr() points to opaque[8] filled with String data
        internal_constructor(p_name.utf8().ptr());
    }
}

// From C string - copies C string into opaque[8]
StringName::StringName(const char *p_name) {
    if (p_name == nullptr || p_name[0] == 0) {
        _data = _empty_data;
    } else {
        internal_constructor(p_name);  // Copies into opaque[8]
    }
}

// From literal - compile-time known, stored inline
StringName::StringName(const char16_t *p_name) {
    internal_constructor(p_name);  // Copies into opaque[8]
}
```

### StringName Destruction

```cpp
StringName::~StringName() {
    if (unique) {
        // Only cleans up internal reference counting
        // Does NOT free external memory because there is none
        // opaque[8] is destroyed as part of stack/struct cleanup
    }
}
```

### Property Registration Flow

```cpp
void ClassDB::_register_extension_class_property(
        const StringName &p_class,
        const PropertyInfo &p_info,
        int p_index) {
    
    StringName class_name = p_class;  // Stack StringName
    StringName setter_name = p_info.setter;  // Stack StringName  
    StringName getter_name = p_info.getter;  // Stack StringName
    StringName property_name = p_info.name;  // Stack StringName
    
    GDExtensionPropertyInfo prop_info;
    prop_info.class_name = class_name._native_ptr();  // Stack pointer
    prop_info.name = property_name._native_ptr();     // Stack pointer
    prop_info.setter = setter_name._native_ptr();     // Stack pointer
    prop_info.getter = getter_name._native_ptr();     // Stack pointer
    
    // Godot interface call - copies data immediately
    internal_interface_ptr->classdb_register_extension_class_property(
        _get_library(),
        prop_info.class_name,
        &prop_info,
        p_index);
    
    // Stack StringNames destroyed - safe because Godot already copied
}
```

---

## Godot-Go's Problematic Flow

### Property Registration Flow

```go
func ClassDBAddProperty(...) {
    // Go heap StringNames created
    classSn := NewStringNameWithLatin1Chars(cn)        // Go heap
    nameSn := NewStringNameWithLatin1Chars(pn)         // Go heap
    setterSn := NewStringNameWithLatin1Chars(setter)   // Go heap
    getterSn := NewStringNameWithLatin1Chars(getter)   // Go heap
    
    // C struct allocated with Go heap pointers inside
    prop_info := NewGDExtensionPropertyInfo(
        &classSn,  // Pointer to Go heap
        p_property_type,
        &nameSn,   // Pointer to Go heap
        ...,
    )
    
    // Call Godot with pointers to Go heap
    CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertyIndexed(
        FFI.Library,
        ci.NameAsStringNamePtr,
        &prop_info,  // Contains Go heap pointers
        &setterSn,   // Pointer to Go heap
        &getterSn,   // Pointer to Go heap
        -1,
    )
    
    // defer Destroy() runs - frees Go heap memory
    // Godot may still have pointers to that memory!
}
```

### The Orphaned StringName Warning

When Godot tries to access a StringName that has been freed:

```
ERROR: Attempt to get non-existent interface function: 'get_variant_get_internal_ptr_func'.
   at: get_interface_function (core/extension/gdextension.cpp:728)

WARN    ffi/ffi.gen.go:390  GDExtensionInterfaceGetProcAddress Error  
    {"name": "get_variant_get_internal_ptr_func"}
```

This occurs because:
1. Go heap StringName memory was freed by `Destroy()`
2. Godot tries to access the StringName later
3. Memory may be reused or corrupted
4. Reference counting fails or double-free occurs

---

## Key Differences Summary

| Aspect | godot-cpp (Correct) | godot-go (Problematic) |
|--------|---------------------|------------------------|
| **StringName Storage** | `uint8_t opaque[8]` inline | `[8]uint8` inline (but in Go heap) |
| **Allocation** | Stack or struct member | Go heap |
| **Pointer Lifetime** | Tied to stack frame | Managed by Go GC + explicit Destroy |
| **Godot Copy Timing** | Immediate during API call | Unclear - may store pointers |
| **Destruction** | Automatic on stack unwind | Explicit `Destroy()` call |
| **Memory Ownership** | Clear - stack owns data | Ambiguous - Go heap + C struct |
| **Double-Free Risk** | None | Yes - Go heap freed, Godot may access |

---

## Recommended Fix for godot-go

### Option 1: Stack-Allocated StringNames (godot-cpp style)

Create StringNames in C memory and pass directly:

```go
func ClassDBAddProperty(...) {
    // Allocate StringNames in C memory
    var cClassName C.GDExtensionStringNameData
    var cPropertyName C.GDExtensionStringNameData
    var cSetterName C.GDExtensionStringNameData
    var cGetterName C.GDExtensionStringNameData
    
    // Initialize from Go strings - data copied into C memory
    C.string_name_init_from_cstr(&cClassName, C.CString(cn))
    C.string_name_init_from_cstr(&cPropertyName, C.CString(pn))
    C.string_name_init_from_cstr(&cSetterName, C.CString(setter))
    C.string_name_init_from_cstr(&cGetterName, C.CString(getter))
    
    // Create PropertyInfo with C-allocated StringNames
    var cPropInfo C.GDExtensionPropertyInfo
    cPropInfo.class_name = &cClassName
    cPropInfo.name = &cPropertyName
    cPropInfo.setter = &cSetterName
    cPropInfo.getter = &cGetterName
    
    // Call Godot interface
    CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertyIndexed(
        FFI.Library,
        ci.NameAsStringNamePtr,
        &cPropInfo,
        &cSetterName,
        &cGetterName,
        -1,
    )
    
    // Clean up C-allocated StringNames
    C.string_name_destroy(&cClassName)
    C.string_name_destroy(&cPropertyName)
    C.string_name_destroy(&cSetterName)
    C.string_name_destroy(&cGetterName)
}
```

### Option 2: Immediate StringName Destruction (hybrid)

Pass StringName pointers to Godot, then immediately destroy Go heap versions after Godot copies:

```go
func ClassDBAddProperty(...) {
    classSn := NewStringNameWithLatin1Chars(cn)
    nameSn := NewStringNameWithLatin1Chars(pn)
    setterSn := NewStringNameWithLatin1Chars(setter)
    getterSn := NewStringNameWithLatin1Chars(getter)
    
    prop_info := NewGDExtensionPropertyInfo(
        &classSn,
        p_property_type,
        &nameSn,
        ...,
    )
    
    CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertyIndexed(
        FFI.Library,
        ci.NameAsStringNamePtr,
        &prop_info,
        &setterSn,
        &getterSn,
        -1,
    )
    
    // IMMEDIATE destroy - after Godot call returns (assuming Godot copied)
    classSn.Destroy()
    nameSn.Destroy()
    setterSn.Destroy()
    getterSn.Destroy()
}
```

**Note**: This assumes Godot copies StringName data immediately during the interface call. Testing required to verify.

---

## Files to Modify

1. **`pkg/ffi/property_info.go`** - Fix `NewGDExtensionPropertyInfo()` to handle C-allocated StringNames
2. **`pkg/core/classdb.go`** - Fix `ClassDBAddProperty()` and `ClassDBAddSignal()` to use C-allocated StringNames
3. **`pkg/ffi/string_name.gen.go`** - Potentially add C-allocator functions for StringNames

---

## Testing Verification

After fixes, verify orphaned StringName warnings are eliminated:

```bash
GODOT=/path/to/godot make test
```

Expected: No orphaned StringName errors in test output.

---

## References

- Godot-CPP StringName: `/home/pcting/workspace/godot-cpp/gen/include/godot_cpp/variant/string_name.hpp`
- Godot-CPP StringName Implementation: `/home/pcting/workspace/godot-cpp/gen/src/variant/string_name.cpp`
- Godot-CPP ClassDB: `/home/pcting/workspace/godot-cpp/src/core/class_db.cpp`
- Godot-Go Property Info: `pkg/ffi/property_info.go`
- Godot-Go ClassDB: `pkg/core/classdb.go`
- Godot GDExtension Docs: https://docs.godotengine.org/en/stable/tutorials/scripting/gdextension/gdextension_c_example.html
