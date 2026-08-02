## 1. Varcall Argument Decode — Missing Vector Types

- [x] 1.1 Add `Rect2`, `Rect2i`, `Transform2D`, `Plane`, `Quaternion`, `AABB`, `Basis`, `Transform3D`, and `Projection` cases to `convertVariantToGoTypeReflectValue` in `pkg/core/method_bind_reflect.go`. For types with `To*()` accessors (`Transform2D`, `Plane`, `Quaternion`, `AABB`, `Basis`, `Transform3D`, `Projection`), use `arg.ToTransform2D()` etc. For `Rect2` and `Rect2i`, use `Rect2Encoder.DecodeVariantPtr(arg.NativeConstPtr())`.

## 2. Ptrcall Argument Decode — Missing Vector Types

- [x] 2.1 Add `Vector2i`, `Vector3`, `Vector3i`, `Vector4i`, `Rect2`, `Rect2i`, `Transform2D`, `Plane`, `Quaternion`, `AABB`, `Basis`, `Transform3D`, and `Projection` cases to both the `reflect.Array` and `reflect.Struct` branches of `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs` in `pkg/core/method_bind_reflect.go`. Pattern: `pV := (*T)(unsafe.Pointer(arg))` followed by `NewTWithT(*pV)`.

## 3. Tests

- [x] 3.1 Add Go test methods in `test/pkg/example.go` for each missing vector type: echo methods that accept and return the same vector type (e.g. `TestEchoRect2(p Rect2) Rect2`). Register each via `ClassDBBindMethod`.
- [x] 3.2 Add GDScript assertions in `test/demo/main.gd` exercising every new Go method via both ptrcall (`example.test_echo_rect2(...)`) and varcall (`example.call("test_echo_rect2", ...)`) paths.
- [x] 3.3 Run `GODOT=/path/to/godot make build` and `make test`. Fix any failures until the suite is green.

## 4. Validation

- [x] 4.1 Run `go vet ./pkg/... ./test/...` and `go build ./...`.
