## Why

`docs/overview.md` promises that the basic built-in types (`bool`, `int`, `float`, `String`, `StringName`, `NodePath`) are supported across the Go↔Godot boundary, but the Godot→Go argument decode paths are incomplete: Go methods taking a Godot `String`, `StringName`, or `NodePath` argument `log.Panic` on both the ptrcall and varcall paths. Two return-encoding bugs also affect these types: narrow numeric returns through the varcall path read past the end of the value, and Go `string` returns leak a temporary Godot `String`.

## What Changes

- Add Godot `String`, `StringName`, and `NodePath` cases to the argument decode switches in `pkg/core/method_bind_reflect.go` (varcall + ptrcall), using the existing `Variant.ToString*` / `String.ToUtf8()` helpers. This fixes the panics for the documented basic types.
- Fix the varcall return encoder for narrow integer and float types in `pkg/builtin/variant_number_encoder.go` so Godot reads the widened `int64`/`float64` value, not a pointer to a narrower Go value (out-of-bounds read).
- Restore the `Destroy()` call for the intermediate Godot `String` in `pkg/builtin/variant_string_encoder.go` so Go `string` returns through the varcall path do not leak.
- Fix unsigned argument decoding in the varcall path so `uint64` values above `MaxInt64` are not sign-wrapped (currently decoded via `ToInt64`).
- Fix a copy-paste logging bug: the varcall decode function logs `"ptrcall arg parsed"`.
- Add round-trip tests covering the basic types (bool, int/uint widths, float widths, Go `string`, Godot `String`, `StringName`, `NodePath`) as method arguments and return values in both call styles.

## Capabilities

### New Capabilities
- `basic-built-in-types`: Correct decoding and encoding of the basic built-in types (`bool`, `int`/`uint` widths, `float` widths, Go `string`, Godot `String`, `StringName`, `NodePath`) as method arguments and return values across both the ptr call and Variant call paths.

### Modified Capabilities
<!-- None -->

## Impact

- `pkg/core/method_bind_reflect.go` — argument decode switches (varcall + ptrcall).
- `pkg/builtin/variant_number_encoder.go` — varcall reflect return encoding for narrow numerics.
- `pkg/builtin/variant_string_encoder.go` — intermediate `String` lifecycle.
- `test/pkg/example.go`, `test/demo/main.gd` — new round-trip tests and GDScript assertions.

## Non-goals

- Full support for the container built-in types (`Array`, `Packed*Array`, `Dictionary`, `Signal`, `Callable`) and `Packed*Array` → Go slice conversion, which remains documented as not yet implemented.
- Parity for the vector and engine built-in types (`Vector2/3/4`, `Color`, `Rect2`, `Transform2D/3D`, etc.) in the ptrcall argument decode path, which is a separate gap; the fix here adds cases only for the documented basic types.
- Unrelated method-binding features (default arguments, varargs, static methods).
