## Why

The vector built-in types (Rect2, Rect2i, Transform2D, Plane, Quaternion, AABB, Basis, Transform3D, Projection, and several integer variants) are missing from the argument decoding paths in both varcall and ptrcall dispatch, causing panics when GDScript passes them as method arguments to Go. Only Vector2, Vector4, and a subset of the i-prefixed types are handled.

## What Changes

- Add all missing vector types (Rect2, Rect2i, Transform2D, Plane, Quaternion, AABB, Basis, Transform3D, Projection, Vector2i, Vector3i, Vector4i, Vector3) to the varcall argument decode switch in `convertVariantToGoTypeReflectValue`
- Add all missing vector types to the ptrcall argument decode switch in `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`
- Add Go test methods and GDScript assertions exercising each vector type as both a method argument (ptrcall and varcall) and a return value
- Fix any existing vector-related bugs discovered during testing (e.g. `Vector3` missing from ptrcall path)

## Capabilities

### New Capabilities

- `vector-built-in-types`: Correct argument decoding and return encoding for all Godot vector/matrix built-in types (Vector2, Vector2i, Vector3, Vector3i, Rect2, Rect2i, Vector4, Vector4i, Transform2D, Plane, Quaternion, AABB, Basis, Transform3D, Projection) across both ptrcall and varcall paths.

### Modified Capabilities

*(none — this is a new capability)*

## Non-goals

- Changing the Go type definitions or memory layout of vector types (they remain `[N]uint8` fixed-size arrays)
- Adding conversion helpers between Go slice types and Packed*Array types
- Implementing vector operators (+, -, *, /) in Go code — these are already supported through the engine's built-in method bindings
- Adding support for container types like Dictionary or Array with vector elements

## Impact

- `pkg/core/method_bind_reflect.go`: Extended switch cases in both `convertVariantToGoTypeReflectValue` (varcall) and `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs` (ptrcall)
- `test/pkg/example.go`: New test methods for each vector type
- `test/demo/main.gd`: New GDScript assertions
- No ABI or generated code changes
