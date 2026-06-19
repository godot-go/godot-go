# Godot-cpp vs Godot-go StringName/String Memory Management Notes

## Date: 2026-04-11

## Initial Observations

### Problem Statement
Orphaned StringName warnings appear when running `make test`:
- 29 from `Example` class methods
- 2 from `Image` type  
- 6 from `custom_signal`

These warnings indicate StringNames are being destroyed after Godot has already tried to access them, or double-freed.

### Root Cause Hypothesis

**godot-go Pattern (Problematic)**:
```go
// StringName is [8]uint8 - inline storage on Go heap
type StringName [8]uint8

// When we pass pointer to C:
className := NewStringNameWithLatin1Chars(cn)  // Go heap allocation
defer className.Destroy()                       // Frees Go memory
prop_info := NewGDExtensionPropertyInfo(
    className.AsGDExtensionConstStringNamePtr(), // Pointer to Go memory
    ...
)
// Godot receives pointer but doesn't copy immediately
// When defer runs: Go memory freed
// Later: C struct Destroy() tries to free already-freed memory!
```

**godot-cpp Pattern (Correct)**:
```cpp
// StringName is uint8_t opaque[8] - inline storage
class StringName {
    uint8_t opaque[STRING_NAME_SIZE] = {};
    _FORCE_INLINE_ GDExtensionTypePtr _native_ptr() const { 
        return const_cast<uint8_t(*)[STRING_NAME_SIZE]>(&opaque); 
    }
};

// When passed to ClassDB:
StringName class_name("MyClass");  // Stack allocation
StringName prop_name("my_property");
GDExtensionPropertyInfo prop_info = {
    .name = prop_name._native_ptr(),  // Pointer to stack memory
    // ...
};
ClassDB::add_property(class_name, prop_info, setter, getter);
// Inside add_property():
internal::gdextension_interface_classdb_register_extension_class_property_indexed(
    library, info.name._native_ptr(), &prop_info, ...);
// Godot copies prop_info.name IMMEDIATELY via GDExtension API
// Stack StringName destructor runs safely after Godot has copied
```

## Key Differences

### 1. Memory Lifetime

| Aspect | godot-cpp | godot-go |
|--------|-----------|----------|
| Allocation | Stack (automatic) | Go heap (managed) |
| Pointer ownership | Godot copies immediately | Godot stores pointer |
| Destruction | Safe after Godot call | Problematic timing |

### 2. StringName Implementation

**godot-cpp**:
- `uint8_t opaque[8]` inline storage
- `_native_ptr()` returns pointer to opaque data
- Destructor calls GDExtension destructor on the same memory
- RAII ensures cleanup happens in correct scope

**godot-go**:
- `[8]uint8` inline storage (similar)
- `AsGDExtensionConstStringNamePtr()` returns pointer
- `Destroy()` calls GDExtension destructor
- **BUT**: When pointer stored in C struct, Go destructor and C destructor both try to free

### 3. PropertyInfo Handling

**godot-cpp** (`class_db.cpp:72-112`):
```cpp
void ClassDB::add_property(const StringName &p_class, const PropertyInfo &p_pinfo, ...) {
    GDExtensionPropertyInfo prop_info = {
        static_cast<GDExtensionVariantType>(p_pinfo.type),
        p_pinfo.name._native_ptr(),     // Pointer to StringName opaque
        p_pinfo.class_name._native_ptr(), // Pointer to StringName opaque
        p_pinfo.hint,
        p_pinfo.hint_string._native_ptr(), // Pointer to String opaque
        p_pinfo.usage,
    };
    // Godot API call - copies prop_info internally
    internal::gdextension_interface_classdb_register_extension_class_property_indexed(
        internal::library, info.name._native_ptr(), &prop_info, ...);
    // prop_info is stack-local, StringNames are stack-local
    // Godot has copied the data, safe to return
}
```

**godot-go** (`property_info.go:17-73`):
```go
func NewGDExtensionPropertyInfo(
    className GDExtensionConstStringNamePtr,
    propertyType GDExtensionVariantType,
    propertyName GDExtensionConstStringNamePtr,
    ...
) GDExtensionPropertyInfo {
    // PROBLEM: Trying to convert Go pointers back to StringNames
    // then allocate new C memory - inefficient and wrong!
    var classNamePtr C.GDExtensionStringNamePtr
    if className != nil {
        classNamePtr = (C.GDExtensionStringNamePtr)(C.malloc(C.size_t(8)))
        CallFunc_GDExtensionInterfaceStringNameNewWithLatin1Chars(
            (GDExtensionUninitializedStringNamePtr)(classNamePtr),
            NewStringNameFromGDExtensionConstStringNamePtr(className).ToUtf8(), // WRONG!
            0,
        )
    }
    // ...
}
```

**Issue**: The helper functions `NewStringNameFromGDExtensionConstStringNamePtr()` and `NewStringFromGDExtensionConstStringPtr()` don't exist! The code is broken.

### 4. Correct godot-go Pattern

Should match godot-cpp: Allocate StringName in C memory immediately, pass pointer to Godot, Godot copies, C memory freed when C struct destroyed.

```go
// Helper function (already exists in method_bind.go)
func allocCStringNameWithLatin1Chars(content string) GDExtensionStringNamePtr {
    ptr := (GDExtensionStringNamePtr)(unsafe.Pointer(C.malloc(C.size_t(StringNameSize))))
    CallFunc_GDExtensionInterfaceStringNameNewWithLatin1Chars(
        (GDExtensionUninitializedStringNamePtr)(ptr),
        content,
        0, // p_is_static = false
    )
    return ptr
}

// Use in NewGDExtensionPropertyInfo
func NewGDExtensionPropertyInfo(...) GDExtensionPropertyInfo {
    var classNamePtr C.GDExtensionStringNamePtr
    if className != nil {
        // Allocate directly in C memory with content
        classNamePtr = (C.GDExtensionStringNamePtr)(C.malloc(C.size_t(StringNameSize)))
        CallFunc_GDExtensionInterfaceStringNameNewWithLatin1Chars(
            (GDExtensionUninitializedStringNamePtr)(classNamePtr),
            goStringFromPtr(className), // Convert pointer to string first
            0,
        )
    }
    // ...
}
```

Wait, this is still circular. We need to rethink the API.

### Better API Design

Instead of:
```go
NewGDExtensionPropertyInfo(classNamePtr, type, propertyNamePtr, ...)
```

Should be:
```go
NewGDExtensionPropertyInfo(className string, type, propertyName string, ...)
```

Then internally:
```go
func NewGDExtensionPropertyInfo(className string, ...) GDExtensionPropertyInfo {
    classNamePtr := allocCStringNameWithLatin1Chars(className)
    propertyNamePtr := allocCStringNameWithLatin1Chars(propertyName)
    hintStringPtr := allocCStringWithUtf8Chars("")
    
    ret := GDExtensionPropertyInfo{...}
    pnr.Pin(&ret) // Pin the struct so C pointers remain valid
    
    return ret
}
```

## Investigation Tasks

- [x] Read godot-cpp `string_name.hpp` - Understand StringName structure
- [x] Read godot-cpp `class_db.cpp` - Understand ClassDB registration
- [ ] Fix `NewGDExtensionPropertyInfo()` - Remove non-existent helpers
- [ ] Fix `ClassDBAddProperty()` - Use C-allocated StringNames
- [ ] Fix `ClassDBAddSignal()` - Use C-allocated StringNames
- [ ] Create comprehensive comparison document
- [ ] Run tests to verify fix

## Current Status

**Files to fix**:
1. `pkg/ffi/property_info.go:17-73` - `NewGDExtensionPropertyInfo()` uses non-existent helpers
2. `pkg/core/classdb.go:71-157` - `ClassDBAddProperty()` - creates Go StringNames then passes pointers
3. `pkg/core/classdb.go:164-224` - `ClassDBAddSignal()` - creates Go StringNames then passes pointers

**Pattern to apply**:
- Use `allocCStringNameWithLatin1Chars()` for all StringName allocations that go into C structs
- Ensure Godot copies the data before Go memory is freed
- Properly destroy C-allocated StringNames in C struct Destroy() methods
