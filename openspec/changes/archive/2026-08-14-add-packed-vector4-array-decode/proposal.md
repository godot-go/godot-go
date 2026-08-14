## Why

`PackedVector4Array` is the only container built-in type that Godot supports but godot-go does not fully wire up. The Go→Godot *encoding* path already handles it (`PackedVector4ArrayEncoder` exists in `builtinclasses.gen.go`), but the Godot→Go *decode* path is incomplete in two ways: (1) the `ToPackedVector4Array()` accessor is missing from the generated `variant.gen.go` because `GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR4_ARRAY` is absent from the `encoderTypeMap` in `cmd/generate/builtin/templatefunctions.go`, and (2) neither the varcall nor ptrcall decode switch in `pkg/core/method_bind_reflect.go` has a `PackedVector4Array` case, so any Go method accepting it as an argument panics at runtime. This was flagged as out of scope by the code review of the preceding "implement-container-built-in-types" change but should be completed for full parity.

## What Changes

- Add `GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR4_ARRAY` to the `encoderTypeMap` in `cmd/generate/builtin/templatefunctions.go`, matching the existing `PackedVector3Array` entry, so that `make generate` produces the `ToPackedVector4Array()` and `NewVariantPackedVector4Array()` methods in `variant.gen.go`.
- Run `make generate` to regenerate `variant.gen.go` (and any other affected generated files).
- Add `case PackedVector4Array` to the varcall decode switch in `convertVariantToGoTypeReflectValue` (`pkg/core/method_bind_reflect.go`) using `arg.ToPackedVector4Array()`, matching the existing `arg.ToPackedVector3Array()` pattern.
- Add `case PackedVector4Array` to the ptrcall decode switch in `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs` (`pkg/core/method_bind_reflect.go`) using `NewPackedVector4ArrayWithPackedVector4Array(*pV)`, matching the existing `PackedVector3Array`/`PackedInt64Array` copy-constructor pattern.
- Add a `TestEchoPackedVector4Array` echo method + `ClassDBBindMethod` registration in `test/pkg/example.go`.
- Add ptrcall + varcall round-trip assertions in `test/demo/main.gd`.

## Capabilities

### New Capabilities

- `packed-vector4-array-decode`: Correct decoding and encoding of `PackedVector4Array` as a method argument across both the ptrcall and varcall call paths, matching the existing `PackedVector3Array` support.

### Modified Capabilities

- `container-built-in-types`: `PackedVector4Array` is added to the varcall and ptrcall argument decode paths, extending the capability to the full set of container types.

## Impact

- `cmd/generate/builtin/templatefunctions.go` — add `GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR4_ARRAY` to `encoderTypeMap`.
- `pkg/builtin/variant.gen.go` — regenerated; adds `ToPackedVector4Array()` and `NewVariantPackedVector4Array()`.
- `pkg/core/method_bind_reflect.go` — varcall decode: add `PackedVector4Array` case; ptrcall decode: add `PackedVector4Array` case.
- `test/pkg/example.go` — new echo test method + `ClassDBBindMethod` registration.
- `test/demo/main.gd` — GDScript assertions for ptrcall + varcall round-trip.
- `docs/overview.md` — add `PackedVector4Array` to the container built-in types table.

## Non-goals

- Go-native slice conversion (`[]Vector4`) as automatic argument/return marshalling; the boundary uses the Godot built-in `PackedVector4Array` type, consistent with all other container types.
- Adding other missing `Packed*Array` → `To*()` accessors (all others already exist).
- Changing the encoding path, which already handles `PackedVector4Array` correctly.
