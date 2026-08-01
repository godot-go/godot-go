# Godot GDExtension String Handling

This document describes how strings are handled in the Godot GDExtension API and their implementation in godot-go.

## Overview

Godot's String type is a flexible string container that internally stores text as UTF-16. The GDExtension API provides multiple encoding options for creating and converting strings, supporting Latin-1, UTF-8, UTF-16, UTF-32, and wide character encodings.

## String Type Representation

### Godot Internal Representation
- **String**: Stored as `[8]uint8` (64-bit builds) or `[4]uint8` (32-bit builds) - uses Small String Optimization (SSO)
- **StringName**: Stored as `[8]uint8` (64-bit builds) or `[4]uint8` (32-bit builds) - a pointer into Godot's global interned hash table, optimized for identifiers
- **Pointer Types**:
  - `GDExtensionStringPtr` - mutable string pointer
  - `GDExtensionConstStringPtr` - immutable string pointer
  - `GDExtensionUninitializedStringPtr` - uninitialized string memory

### Memory Layout
```go
// Godot-go representation
type String [8]uint8
type StringName [8]uint8
```

`StringName` is **not a string**. Its 8 bytes are a pointer to a refcounted interned `_Data` entry in Godot's global hash table. Two `StringName` values with the same text share the same `_Data` pointer, making equality an O(1) pointer comparison. See [stringname-mutation-analysis.md](stringname-mutation-analysis.md) for the full internals analysis.

## String Creation

### From Go Strings (UTF-8)

The most common way to create Godot strings from Go code:

```go
// UTF-8 (recommended for most cases)
str := builtin.NewStringWithUtf8Chars("Hello, World!")

// Latin-1 (for ASCII-compatible content)
str := builtin.NewStringWithLatin1Chars("Hello, World!")
```

### From Character Arrays

For advanced use cases with specific encodings:

```go
// UTF-16
var utf16Str [5]char16_t
// ... populate utf16Str ...
str := NewStringWithUtf16Chars(&utf16Str[0])

// UTF-32
var utf32Str Char32T = 'A'
str := builtin.NewStringWithUtf32Char(utf32Str)

// Wide characters (platform-dependent)
var wideStr [10]WcharT
// ... populate wideStr ...
str := NewStringWithWideChars(&wideStr[0])
```

### Constructors from Other Godot Types

```go
// From StringName
str := builtin.NewStringWithStringName(stringName)

// From NodePath
str := builtin.NewStringWithNodePath(nodePath)

// Default constructor (empty string)
str := builtin.NewString()
```

## String Conversion

### To Go Strings

```go
// Convert to UTF-8 Go string (recommended)
goStr := godotStr.ToUtf8()

// Convert to Latin-1
asciiStr := godotStr.ToAscii()

// Convert to UTF-32
utf32Str := godotStr.ToUtf32()
```

### To Godot StringName

```go
stringName := builtin.NewStringNameWithString(godotStr)
// Or directly from content
stringName := builtin.NewStringNameWithUtf8Chars("content")
stringName := builtin.NewStringNameWithLatin1Chars("content")
```

### StringName to String

```go
str := stringName.AsString()
```

## Encoding Details

### Latin-1 (ISO-8859-1)
- Single-byte encoding (0-255)
- Compatible with ASCII for values 0-127
- Best for: Western European text, simple ASCII content
- Functions: `StringNewWithLatin1Chars`, `StringToLatin1Chars`

### UTF-8
- Variable-length encoding (1-4 bytes per character)
- ASCII-compatible, widely supported
- Best for: Most use cases, web content, cross-platform compatibility
- Functions: `StringNewWithUtf8Chars`, `StringToUtf8Chars`

### UTF-16
- Variable-length encoding (2-4 bytes per character)
- Godot's internal storage format
- Best for: Windows compatibility, JavaScript interoperability
- Functions: `StringNewWithUtf16Chars`, `StringToUtf16Chars`

### UTF-32
- Fixed-length encoding (4 bytes per character)
- Simple indexing, one code point per unit
- Best for: Text processing, character counting
- Functions: `StringNewWithUtf32CharsAndLen`, `StringToUtf32Chars`

### Wide Characters
- Platform-dependent (4 bytes on Linux/macOS, 2 bytes on Windows)
- Best for: Platform-specific native code interop
- Functions: `StringNewWithWideChars`, `StringToWideChars`

## Usage Patterns

### Basic String Operations

```go
// Create and use strings
str1 := builtin.NewStringWithUtf8Chars("Hello")
str2 := builtin.NewStringWithUtf8Chars(" World")

// Concatenation (via operator methods)
combined := str1.OperatorPlusEqString(str2)

// Length
length := str1.Len()

// Substring
substr := str1.Substring(0, 5)

// Cleanup - important for proper memory management
defer str1.Destroy()
defer str2.Destroy()
defer combined.Destroy()
```

### String in Variants

```go
// Go string to Variant
var enc = builtin.Utf8GoStringEncoder
variant := enc.EncodeVariantPtr("hello")

// Variant to Go string
goStr := enc.DecodeVariantPtr(variant)
```

### StringName for Identifiers

```go
// Use StringName for property names, method names, signals
propName := builtin.NewStringNameWithUtf8Chars("position")
methodName := builtin.NewStringNameWithUtf8Chars("queue_free")
defer propName.Destroy()
defer methodName.Destroy()

// StringNames are interned; each New*StringName must be balanced by Destroy()
// StringName is an 8-byte handle into Godot's interned table; the copy
// constructor increments the refcount, Destroy() decrements it.
```

### StringName Lifetime and Pinning

`StringName` is an 8-byte value. Storing it **by value** in Go structs is safe and is the recommended pattern (`ClassInfo` stores `NameStringName` / `ParentNameStringName` by value). When passing a `StringName` to a GDExtension interface call:

1. Godot's copy constructor only **reads** the 8 bytes and increments the refcount on the shared `_Data` — it never writes back to caller memory.
2. Go's GC can move heap-allocated values, so pin the value for the duration of the call with `runtime.Pinner`.
3. When the value lives inside a struct that also holds Go pointers (maps, slices), copy it to an isolated local variable before pinning — the Go cgo checker walks the whole parent struct and panics on unpinned Go pointers otherwise.

```go
// Register property — local-copy-isolate-then-pin
ci := Internal.GDRegisteredGDClasses.Get(cn)
cnName := ci.NameStringName      // copy [8]uint8 to isolated local
pnr.Pin(&cnName)                 // pin before the call
CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertyIndexed(
	FFI.Library,
	cnName.AsGDExtensionConstStringNamePtr(),
	// ...
)
```

For structs with C pointers to Go heap (e.g. `GDExtensionClassMethodInfo`), allocate in C heap via `C.malloc()` instead — the cgo checker cannot track pinning across call boundaries and treats C heap memory as foreign.

## Performance Considerations

1. **Encoding Choice**: Use UTF-8 for most cases. Latin-1 is faster for ASCII-only content.
2. **StringName vs String**: Use StringName for identifiers that are reused frequently.
3. **Memory Management**: Always call `Destroy()` on strings created with constructors that don't use SSO.
4. **Avoid Repeated Conversions**: Cache string conversions when working in loops.

## Common Pitfalls

### Encoding Mismatch
```go
// ❌ Wrong: Assuming Go strings are Latin-1
str := builtin.NewStringWithLatin1Chars("日本語") // May lose data!

// ✅ Correct: Use UTF-8 for non-ASCII content
str := builtin.NewStringWithUtf8Chars("日本語")
```

### Missing Cleanup
```go
// ❌ Wrong: Forgetting to destroy strings
str := builtin.NewString()
// ... use str ...
// Memory leak!

// ✅ Correct: Always cleanup
str := builtin.NewString()
defer str.Destroy()
```

### Unnecessary Conversions
```go
// ❌ Inefficient: Multiple conversions
goStr := godotStr.ToUtf8()
godotStr2 := builtin.NewStringWithUtf8Chars(goStr)

// ✅ Efficient: Work with Godot types directly
godotStr2 := godotStr.ToUpper() // Keep as Godot String
```

## API Reference

### Creation Functions

| Function | Encoding | Use Case |
|----------|----------|----------|
| `NewString()` | N/A | Empty string |
| `NewStringWithUtf8Chars()` | UTF-8 | Most Go strings |
| `NewStringWithLatin1Chars()` | Latin-1 | ASCII/Western European |
| `NewStringWithUtf16Chars()` | UTF-16 | Windows/JS interop |
| `NewStringWithUtf32Chars()` | UTF-32 | Character processing |
| `NewStringWithWideChars()` | Wide | Platform-native |

### Conversion Functions

| Function | Output | Use Case |
|----------|--------|----------|
| `ToUtf8()` | UTF-8 string | Most Go strings |
| `ToAscii()` | Latin-1 string | ASCII content |
| `ToUtf32()` | UTF-32 string | Character processing |

## See Also

- [Godot GDExtension C Example](https://docs.godotengine.org/en/stable/tutorials/scripting/gdextension/gdextension_c_example.html)
- [extension_api.json](../godot_headers/extension_api.json) - Full API metadata
- [gdextension_interface.h](../godot_headers/godot/gdextension_interface.h) - C interface definitions
