## 1. Argument Decoding (Godot → Go)

- [x] 1.1 Add `String`, `StringName`, and `NodePath` cases to `convertVariantToGoTypeReflectValue` (`pkg/core/method_bind_reflect.go` varcall path) using `arg.ToString()`, `arg.ToStringName()`, `arg.ToNodePath()`
- [x] 1.2 Add `String`, `StringName`, and `NodePath` cases to `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs` (ptrcall path) using byte-copy helpers (`NewStringNameWithGDExtensionConstStringNamePtr` for StringName)
- [x] 1.3 Replace signed `arg.ToInt64()` derivation for `reflect.Uint*` kinds with the typed `ToUint*` accessors in the varcall path, preserving `uint64` > `MaxInt64`
- [x] 1.4 Fix the copy-paste log label (`"ptrcall arg parsed"` → correct label) in the varcall decode function

## 2. Return Encoding (Go → Godot)

- [x] 2.1 Fix `createNumberEncoder.encodeReflectVariantPtrArg` (`pkg/builtin/variant_number_encoder.go`) to widen narrow numeric values through an `E` temp via `encodeTypePtrArg` before calling the variant constructor
- [x] 2.2 Restore `defer enc.Destroy()` at `variant_string_encoder.go:68` and `:81`, and add the matching `Destroy()` in `encodeReflectVariantPtrArg`

## 3. Tests

- [x] 3.1 Add Go methods in `test/pkg/example.go` accepting and returning each basic type (`bool`, `int64`, `float64`, Go `string`, Godot `String`, `StringName`, `NodePath`) and register them via `ClassDBBindMethod`
- [x] 3.2 Add GDScript assertions in `test/demo/main.gd` covering ptrcall and varcall round-trips for all basic types, a large `uint64` argument, narrow numeric returns, and repeated Go `string` returns
- [x] 3.3 Run `GODOT=/path/to/godot make build` and `make test`; fix any failures until the suite is green

## 4. Validation

- [x] 4.1 Run `go vet ./pkg/core/... ./pkg/builtin/...` and `go build ./...`
- [x] 4.2 Confirm the new test assertions exercise both the ptrcall and varcall paths and there are no `log.Panic("unsupported ...")` hits for the basic types
