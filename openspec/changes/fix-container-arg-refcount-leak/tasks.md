## 1. Container copy-constructor helpers

- [x] 1.1 Add copy-constructor helpers beside `StringNameCopyConstructor` in `pkg/builtin` for `Array`, `PackedByteArray`, `PackedInt32Array`, `PackedInt64Array`, `PackedFloat32Array`, `PackedFloat64Array`, `PackedStringArray`, `PackedVector2Array`, `PackedVector3Array`, `PackedColorArray`, each calling the type's `constructor_1` via `CallBuiltinConstructor` (pattern: `char_string.go` `StringNameCopyConstructor`)
- [x] 1.2 Verify `go build ./pkg/builtin/...` compiles

## 2. Ptrcall path (return encoding + arg decode)

- [x] 2.1 In `GDExtensionTypePtrFromReflectValue` (`pkg/core/variant_refect_value.go`), replace `XEncoder.EncodeTypePtrArg(inst, rOut)` with `XCopyConstructor(rOut, inst.NativeConstPtr())` + `inst.Destroy()` for the ten container types, matching the existing `StringName`/`NodePath` cases
- [x] 2.2 In `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs` (`pkg/core/method_bind_reflect.go`), change the ten container `reflect.Array` cases from `v := NewXWithX(*pV)` to byte-copy borrow `v := *(*X)(arg)`, matching the existing `String`/`StringName`/`NodePath` cases
- [x] 2.3 Verify `go build ./pkg/core/... ./pkg/builtin/...` compiles

## 3. Varcall arg decode

- [x] 3.1 In `convertVariantToGoTypeReflectValue` (`pkg/core/method_bind_reflect.go`), decode the `Array`, `Dictionary`, and nine `Packed*Array` cases as owned copies (`v := arg.ToX(); return reflect.ValueOf(v)`). An initial attempt used the `StringName`/`NodePath` byte-copy borrow (construct → byte-copy → `Destroy()`), but that is unsafe for containers: when the supplied Variant is a *different* container type than the Go parameter (e.g. `Array[int]` → `PackedInt64Array`), the type-from-variant constructor allocates a buffer the Variant does not own, so `Destroy()` left the caller reading freed memory (regressed `test_tarray_arg`). Owned decode is correct in both same-type and cross-type cases.
- [x] 3.2 Verify `go build ./pkg/core/...` compiles
- [x] 3.3 Release the owned varcall container args in `GoMethodMetadata.Call` (`pkg/core/method_bind.go`) via `destroyOwnedContainerArgs`, invoked *after* the return value is encoded so an echoed container return already holds its own reference

## 4. Tests

- [x] 4.1 Ensure echo methods exist in `test/pkg/example.go` for all ten container types (add any missing) and add arg-consume (non-echo) methods that read a container argument and return a scalar derived from it
- [x] 4.2 Register the arg-consume methods via `ClassDBBindMethod` in `test/pkg/example.go`
- [x] 4.3 Add ptrcall (`example.method(...)`) and varcall (`example.call("method", ...)`) round-trip assertions in `test/demo/main.gd` for echo and arg-consume of each container type, asserting returned/observed contents match

## 5. Validate

- [x] 5.1 Run `go vet ./pkg/core/... ./pkg/builtin/...`
- [x] 5.2 Run `go build ./...`
- [x] 5.3 Run `GODOT=/path/to/godot make build` and `make test`; fix any failures until the suite is green, confirming no orphaned-object or double-free symptoms for container round-trips

## 6. Tests for review-finding spec scenarios

- [x] 6.1 Add `TestMutateArray`/`TestMutatePackedInt64Array` (in-place `Append` on the received argument) and `TestCloneEchoArray`/`TestRebuildArrayPlus` to `test/pkg/example.go`, register via `ClassDBBindMethod`; assert in `test/demo/main.gd`: ptrcall mutation propagates to the caller's container, Array mutations also propagate through varcall (`Array` is shared-reference, no CoW) while `PackedInt64Array` varcall mutation is isolated by the owned decode's CoW; a byte-identical defensive clone round-trips as a borrow echo and a rebuilt array (bytes differing from every argument) round-trips without mutating or leaking
- [x] 6.2 Sensitivity-verify the new tests: temporarily reverting the ptrcall `PackedInt64Array` decode to an owned copy fails exactly the ptrcall propagation assertion; temporarily disabling `isPtrcallBorrowEcho` produces `_ref` refcount-underflow errors and a crash on echo round-trips; both temporary mutations reverted afterwards
- [x] 6.3 Run `GODOT=/path/to/godot make build` and `make test` — suite green (646 passes, 0 failures)
