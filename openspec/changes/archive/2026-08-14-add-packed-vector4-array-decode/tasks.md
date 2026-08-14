## 1. Fix codegen entry point

- [x] 1.1 Add `GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR4_ARRAY` entry to `encoderTypeMap` in `cmd/generate/builtin/templatefunctions.go`, matching the existing `PackedVector3Array` entry
- [x] 1.2 Run `make generate` and verify `variant.gen.go` adds `ToPackedVector4Array()` and `NewVariantPackedVector4Array()`; verify diff is scoped to PackedVector4Array
- [x] 1.3 Verify `go build ./pkg/builtin/...` compiles

## 2. Varcall decode

- [x] 2.1 Add `case PackedVector4Array` to the `reflect.Array` switch in `convertVariantToGoTypeReflectValue` (`pkg/core/method_bind_reflect.go`) using `arg.ToPackedVector4Array()`, matching the existing `arg.ToPackedVector3Array()` pattern
- [x] 2.2 Verify `go build ./pkg/core/...` compiles

## 3. Ptrcall decode

- [x] 3.1 Add `case PackedVector4Array` to the `reflect.Array` switch in `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs` (`pkg/core/method_bind_reflect.go`) using `NewPackedVector4ArrayWithPackedVector4Array(*pV)`, matching the existing `PackedVector3Array` pattern
- [x] 3.2 Verify `go build ./pkg/core/...` compiles

## 4. Tests

- [x] 4.1 Add `TestEchoPackedVector4Array(arr PackedVector4Array) PackedVector4Array` echo method + `ClassDBBindMethod` registration in `test/pkg/example.go`
- [x] 4.2 Add ptrcall + varcall round-trip assertions in `test/demo/main.gd` for `test_echo_packed_vector4_array`

## 5. Update docs

- [x] 5.1 Add `PackedVector4Array` to the container built-in types table in `docs/overview.md`

## 6. Validate

- [x] 6.1 Run `go vet ./pkg/core/... ./pkg/builtin/...`
- [x] 6.2 Run `go build ./...`
- [x] 6.3 Run `GODOT=/path/to/godot make build` and `make test`; fix any failures until the suite is green

## 7. Return encoding dispatch (added during implementation)

The proposal assumed the Go→Godot encoding path was complete, but the reflect-value dispatch in `pkg/core/variant_refect_value.go` had no `PackedVector4Array` case, so returning a `PackedVector4Array` (the echo tests) panicked with "unhandled array value type to GDExtensionTypePtr".

- [x] 7.1 Add `case PackedVector4Array` to `GDExtensionTypePtrFromReflectValue` (ptrcall return, both `reflect.Struct` and `reflect.Array` branches) and to `GDExtensionVariantPtrFromReflectValue` (varcall return, `reflect.Array` branch), matching the `PackedVector3Array` pattern
- [x] 7.2 Rebuild and re-run `make test`; confirm 0 failures
