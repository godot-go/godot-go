## Context

See `proposal.md - Why` and `proposal.md - What Changes` for motivation and scope.

The vector built-in types are all `[N]uint8` fixed-size arrays in Go. Return encoding (both TypePtr and VariantPtr) already handles all 15 vector types through generated encoder objects (`variant_refect_value.go`). Argument decoding is the gap: only 6 of 15 types have varcall cases (`convertVariantToGoTypeReflectValue`) and only 2 of 15 have ptrcall cases (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`).

## Goals / Non-Goals

**Goals:**
- Add all missing vector types to the varcall arg decode switch (`convertVariantToGoTypeReflectValue`)
- Add all missing vector types to the ptrcall arg decode switch (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`)
- Write test methods and GDScript assertions covering every vector type in both call paths

**Non-Goals:**
- No changes to Go type definitions or memory layout (they remain `[N]uint8`)
- No changes to `variant_refect_value.go` (return encoding is already complete)
- No changes to generated encoder files (`builtinclasses.gen.go`, `variant.gen.go`)
- No Packed*Array conversion helpers

## Decisions

### Decision 1: Use To*() accessors for varcall decode where available; use encoder for Rect2/Rect2i

**Approach:** All vector types except Rect2 and Rect2i have generated `To*()` methods on `Variant` (e.g. `arg.ToTransform2D()`). The varcall switch will use these for the 11 types that have them. For `Rect2` and `Rect2i`, which lack `ToRect2()` / `ToRect2i()`, the `Rect2Encoder.DecodeVariantPtr(arg.NativeConstPtr())` call will be used — the generic encoder supports all built-in types.

**Rationale:** Matches the existing code pattern (8 existing cases already use `arg.To*()`). The encoder approach for Rect2/Rect2i avoids adding generated methods and is consistent with how the encoder is used elsewhere.

### Decision 2: Use copy constructors for ptrcall decode

**Approach:** For all vector types in the ptrcall `reflect.Array` switch: cast `(*T)(unsafe.Pointer(arg))`, then call `NewTWithT(*pV)` to create an owned copy.

**Rationale:** Matches the existing pattern for Vector2 and Vector4. The copy constructor ensures Go owns the resulting value. The `[N]uint8` types are fixed-size, so the pointer dereference and copy constructor are equivalent to a memcpy of the fixed bytes.

### Decision 3: Add both `reflect.Array` and `reflect.Struct` branches

**Approach:** The Go runtime may represent `[N]uint8` as either `reflect.Array` or `reflect.Struct` depending on how the type is used and the Go version. Both branches in the ptrcall switch will be updated identically.

**Rationale:** The existing code already has both branches with identical patterns for Vector2 and Vector4. Both need coverage to avoid surprise panics on different Go versions.

## Risks / Trade-offs

- [Low] The ptrcall decode for `[N]uint8` types uses `unsafe.Pointer` casts. This is safe because the Godot engine provides a pointer to memory of the correct layout and size — the type system validation (`reflect.Array` kind check + type switch) prevents misuse. The same pattern is already used for Vector2, Vector4, String, and NodePath.
- [Low] Rect2/Rect2i varcall decode uses the encoder instead of a dedicated To*() method. If a future code generation pass adds `ToRect2()`, the switch should be updated. This is a trivial change.
