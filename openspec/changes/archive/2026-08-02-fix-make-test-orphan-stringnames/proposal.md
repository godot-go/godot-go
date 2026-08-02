## Why

`make test` shows Godot orphan StringName warnings at exit for "child", "root", and "returned_name", indicating StringName leaks from the Go binding. These warnings clutter test output and signal improper lifecycle management. This was previously fixed down to 1 orphan but regressed to 3 during the Godot 4.7.2 update and new test additions.

The deeper issue: godot-go's encoding/decoding layer uses raw byte copies (memcpy) for StringName and NodePath when they cross the C/Go boundary, bypassing Godot's internal refcount management. godot-cpp (the reference GDExtension implementation) always uses the GDExtension copy constructor, ensuring refcount correctness. Every StringName that crosses the C/Go boundary must use the GDExtension copy constructor, not raw byte copy.

## What Changes

- Fix return encoding for StringName/NodePath in `GDExtensionTypePtrFromReflectValue` (both Struct and Array branches) to use copy constructor + Destroy
- Fix return variant encoding for StringName/NodePath in `GDExtensionVariantPtrFromReflectValue` to Destroy after encoding
- Fix `NewStringNameWithGDExtensionConstStringNamePtr` to use the GDExtension copy constructor instead of raw byte copy
- Fix argument decode for StringName/NodePath in `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs` to use copy constructor
- Remove or fix `NewSimpleGDExtensionPropertyInfo` (DEPRECATED, leaks StringNames)
- Update AGENTS.md to mandate cross-referencing godot-cpp before implementing
- Verify zero orphan StringName warnings at `make test`

## Capabilities

No spec-level behavior changes — this is a pure memory leak fix. The GDScript-visible API is unchanged. `skip_specs: true` set in `.openspec.yaml`.

## Impact

- `pkg/builtin/char_string.go` — add `StringNameCopyConstructor`/`NodePathCopyConstructor` helpers; fix `NewStringNameWithGDExtensionConstStringNamePtr`
- `pkg/builtin/variant_builtinclass_encoder.go` — fix encode/decode for refcounted types
- `pkg/core/variant_refect_value.go` — fix return encoding for StringName/NodePath
- `pkg/core/method_bind_reflect.go` — fix argument decode for StringName/NodePath
- `pkg/core/object.go` — remove or fix `NewSimpleGDExtensionPropertyInfo`
- `AGENTS.md` — add godot-cpp cross-reference mandate
