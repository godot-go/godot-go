## 1. Varcall decode: add `Array` case

- [x] 1.1 Add `case Array` to the `reflect.Array` switch in `convertVariantToGoTypeReflectValue` (`pkg/core/method_bind_reflect.go`) using `arg.ToArray()`, matching the existing `arg.ToPackedByteArray()` pattern
- [x] 1.2 Verify `go build ./pkg/core/...` compiles

## 2. Ptrcall decode: add `Array` and eight `Packed*Array` cases

- [x] 2.1 Add `case Array` to `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs` `reflect.Array` switch, using `NewArrayWithArray(*pV)` copy constructor, matching the existing `PackedInt64Array` pattern
- [x] 2.2 Add `case PackedByteArray` using `NewPackedByteArrayWithPackedByteArray(*pV)`
- [x] 2.3 Add `case PackedInt32Array` using `NewPackedInt32ArrayWithPackedInt32Array(*pV)`
- [x] 2.4 Add `case PackedFloat32Array` using `NewPackedFloat32ArrayWithPackedFloat32Array(*pV)`
- [x] 2.5 Add `case PackedFloat64Array` using `NewPackedFloat64ArrayWithPackedFloat64Array(*pV)`
- [x] 2.6 Add `case PackedStringArray` using `NewPackedStringArrayWithPackedStringArray(*pV)`
- [x] 2.7 Add `case PackedVector2Array` using `NewPackedVector2ArrayWithPackedVector2Array(*pV)`
- [x] 2.8 Add `case PackedVector3Array` using `NewPackedVector3ArrayWithPackedVector3Array(*pV)`
- [x] 2.9 Add `case PackedColorArray` using `NewPackedColorArrayWithPackedColorArray(*pV)`
- [x] 2.10 Verify `go build ./pkg/core/...` compiles

## 3. Test methods (Go side)

- [x] 3.1 Add `TestEchoArray(arr Array) Array` echo method + `ClassDBBindMethod` registration in `test/pkg/example.go`
- [x] 3.2 Add `TestEchoPackedByteArray(arr PackedByteArray) PackedByteArray` echo method + registration
- [x] 3.3 Add `TestEchoPackedInt32Array(arr PackedInt32Array) PackedInt32Array` echo method + registration
- [x] 3.4 Add `TestEchoPackedInt64Array(arr PackedInt64Array) PackedInt64Array` echo method + registration
- [x] 3.5 Add `TestEchoPackedFloat32Array(arr PackedFloat32Array) PackedFloat32Array` echo method + registration
- [x] 3.6 Add `TestEchoPackedFloat64Array(arr PackedFloat64Array) PackedFloat64Array` echo method + registration
- [x] 3.7 Add `TestEchoPackedStringArray(arr PackedStringArray) PackedStringArray` echo method + registration
- [x] 3.8 Add `TestEchoPackedVector2Array(arr PackedVector2Array) PackedVector2Array` echo method + registration
- [x] 3.9 Add `TestEchoPackedVector3Array(arr PackedVector3Array) PackedVector3Array` echo method + registration
- [x] 3.10 Add `TestEchoPackedColorArray(arr PackedColorArray) PackedColorArray` echo method + registration

## 4. Test assertions (GDScript side)

- [x] 4.1 Add ptrcall + varcall round-trip assertions in `test/demo/main.gd` for `test_echo_array`
- [x] 4.2 Add ptrcall + varcall round-trip assertions for `test_echo_packed_byte_array`
- [x] 4.3 Add ptrcall + varcall round-trip assertions for `test_echo_packed_int32_array`
- [x] 4.4 Add ptrcall + varcall round-trip assertions for `test_echo_packed_int64_array`
- [x] 4.5 Add ptrcall + varcall round-trip assertions for `test_echo_packed_float32_array`
- [x] 4.6 Add ptrcall + varcall round-trip assertions for `test_echo_packed_float64_array`
- [x] 4.7 Add ptrcall + varcall round-trip assertions for `test_echo_packed_string_array`
- [x] 4.8 Add ptrcall + varcall round-trip assertions for `test_echo_packed_vector2_array`
- [x] 4.9 Add ptrcall + varcall round-trip assertions for `test_echo_packed_vector3_array`
- [x] 4.10 Add ptrcall + varcall round-trip assertions for `test_echo_packed_color_array`

## 5. Update docs

- [x] 5.1 Update `docs/overview.md` to remove "NOT YET IMPLEMENTED" annotations for `Array`, `PackedByteArray`, `PackedInt32Array`, `PackedInt64Array`, `PackedFloat32Array`, `PackedFloat64Array`, `PackedStringArray`, `PackedVector2Array`, `PackedVector3Array`, `PackedColorArray`

## 6. Validate

- [x] 6.1 Run `go vet ./pkg/core/... ./pkg/builtin/...`
- [x] 6.2 Run `go build ./...`
- [x] 6.3 Run `GODOT=/path/to/godot make build` and `make test`; fix any failures until the suite is green
