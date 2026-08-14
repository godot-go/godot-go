## Why

`docs/overview.md` documents the container built-in types (`Array`, `PackedByteArray`, `PackedInt32Array`, `PackedInt64Array`, `PackedFloat32Array`, `PackedFloat64Array`, `PackedStringArray`, `PackedVector2Array`, `PackedVector3Array`, `PackedColorArray`) as "NOT YET IMPLEMENTED." The Go→Godot *encoding* paths already handle all of these types in both the ptrcall and varcall return code; the gap is entirely in the Godot→Go *decode* argument paths. Specifically: (1) the varcall decode (`convertVariantToGoTypeReflectValue`) is missing the `Array` case, and (2) the ptrcall decode (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`) is missing `Array` and eight of the nine `Packed*Array` types — only `PackedInt64Array` is handled via a byte-copy constructor. Methods that accept these types as arguments therefore panic at runtime on the missing path, and there are no round-trip echo tests covering them.

## What Changes

- Add the `Array` case to the varcall decode switch in `pkg/core/method_bind_reflect.go` (`convertVariantToGoTypeReflectValue`), using `arg.ToArray()`, matching the existing `arg.ToPackedByteArray()` pattern for the Packed types that are already present.
- Add `Array` and the eight missing `Packed*Array` types (`PackedByteArray`, `PackedInt32Array`, `PackedFloat32Array`, `PackedFloat64Array`, `PackedStringArray`, `PackedVector2Array`, `PackedVector3Array`, `PackedColorArray`) to the ptrcall decode switch in `pkg/core/method_bind_reflect.go` (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`), using the `*FromPtr` copy-constructor pattern already established by `PackedInt64Array` (e.g. `NewPackedByteArrayWithPackedByteArray(*pV)`).
- Add round-trip echo test methods in `test/pkg/example.go` and `ClassDBBindMethod` registrations for every container type, covering both the ptrcall and varcall call styles.
- Add GDScript assertions in `test/demo/main.gd` exercising all container types as arguments and return values via both `example.method(...)` (ptrcall) and `example.call("method", ...)` (varcall).
- Update `docs/overview.md` to remove the "NOT YET IMPLEMENTED" annotations for the container built-in types whose decode/encode paths are now complete.

## Capabilities

### New Capabilities

- `container-built-in-types`: Correct decoding and encoding of all container built-in types (`Array`, `PackedByteArray`, `PackedInt32Array`, `PackedInt64Array`, `PackedFloat32Array`, `PackedFloat64Array`, `PackedStringArray`, `PackedVector2Array`, `PackedVector3Array`, `PackedColorArray`) as method arguments and return values across both the ptrcall and varcall call paths.

### Modified Capabilities

- `basic-built-in-types`: Add `Array` decoding to the varcall path to complete the container coverage described in the existing spec.

## Impact

- `pkg/core/method_bind_reflect.go` — varcall decode (`convertVariantToGoTypeReflectValue`): add `Array` case; ptrcall decode (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`): add `Array` + 8 `Packed*Array` cases.
- `test/pkg/example.go` — new echo test methods + `ClassDBBindMethod` registrations for all container types.
- `test/demo/main.gd` — new GDScript assertions for container round-trips in both call styles.
- `docs/overview.md` — remove "NOT YET IMPLEMENTED" annotations.
- No API/signature changes; purely additive decode cases and tests.

## Non-goals

- Go-native slice conversion (`[]Variant`, `[]byte`, `[]int32`, etc.) as automatic argument/return marshalling. The change uses the Godot built-in container types (`Array`, `Packed*Array`) for boundary crossing, consistent with how `Vector2` and other built-ins are handled.
- Adding new container types beyond those listed in `docs/overview.md`.
- Changing the encoding paths (Go→Godot), which already handle all container types correctly.
