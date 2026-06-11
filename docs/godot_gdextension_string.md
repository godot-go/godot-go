# Godot GDExtension String Handling

This document describes how strings are handled in the Godot GDExtension API and their implementation in godot-go.

## Overview

Godot's String type is a flexible string container that internally stores text as UTF-16. The GDExtension API provides multiple encoding options for creating and converting strings, supporting Latin-1, UTF-8, UTF-16, UTF-32, and wide character encodings.

## String Type Representation

### Godot Internal Representation
- **String**: Stored as `[8]uint8` (64-bit builds) or `[4]uint8` (32-bit builds) - uses Small String Optimization (SSO)
- **StringName**: Stored as `[4]uint8` - optimized for identifiers and interned strings
- **Pointer Types**:
  - `GDExtensionStringPtr` - mutable string pointer
  - `GDExtensionConstStringPtr` - immutable string pointer
  - `GDExtensionUninitializedStringPtr` - uninitialized string memory

### Memory Layout
```go
// Godot-go representation
type String [8]uint8
type StringName [4]uint8
```

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

// StringNames are interned and more efficient for repeated use
// They don't need Destroy() calls in most cases
```

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
