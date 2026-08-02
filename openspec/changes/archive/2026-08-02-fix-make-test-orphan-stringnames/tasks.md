## 1. Add copy constructor helpers

- [x] 1.1 Add `StringNameCopyConstructor` function in `pkg/builtin/char_string.go` that calls Godot's copy constructor (`constructor_1`) and pins the source pointer
- [x] 1.2 Add `NodePathCopyConstructor` function in `pkg/builtin/char_string.go` that calls Godot's copy constructor (`constructor_1`) and pins the source pointer

## 2. Fix return encoding (ptrcall path)

- [x] 2.1 In `GDExtensionTypePtrFromReflectValue` (Struct branch, `case StringName:`), replace `EncodeTypePtrArg` with `StringNameCopyConstructor` + `inst.Destroy()`
- [x] 2.2 In `GDExtensionTypePtrFromReflectValue` (Struct branch, `case NodePath:`), replace `EncodeTypePtrArg` with `NodePathCopyConstructor` + `inst.Destroy()`
- [x] 2.3 In `GDExtensionTypePtrFromReflectValue` (Array branch, `case StringName:`), replace `EncodeTypePtrArg` with `StringNameCopyConstructor` + `inst.Destroy()`
- [x] 2.4 In `GDExtensionTypePtrFromReflectValue` (Array branch, `case NodePath:`), replace `EncodeTypePtrArg` with `NodePathCopyConstructor` + `inst.Destroy()`

## 3. Fix return encoding (varcall/Variant path)

- [x] 3.1 In `GDExtensionVariantPtrFromReflectValue` (`case StringName:`), add `inst.Destroy()` after encoding
- [x] 3.2 In `GDExtensionVariantPtrFromReflectValue` (`case NodePath:`), add `inst.Destroy()` after encoding

## 4. Fix raw byte copy in NewStringNameWithGDExtensionConstStringNamePtr

- [x] 4.1 In `pkg/builtin/char_string.go`, change `NewStringNameWithGDExtensionConstStringNamePtr` to call the GDExtension copy constructor instead of raw byte copy
- [x] 4.2 In `pkg/core/method_bind_reflect.go`, fix argument decoding for StringName and NodePath from ptrcall to use copy constructors

## 5. Remove or fix NewSimpleGDExtensionPropertyInfo

- [x] 5.1 In `pkg/core/object.go`, remove `NewSimpleGDExtensionPropertyInfo` and replace its callers with `NewGDExtensionPropertyInfoFromNames`

## 6. Update AGENTS.md

- [x] 6.1 Add a section mandating cross-referencing godot-cpp before implementing any GDExtension boundary crossing
- [x] 6.2 Add a section mandating checking `openspec/specs/` before implementing

## 7. Verify

- [x] 7.1 Run `make build` to confirm compilation
- [x] 7.2 Run `make test` and verify zero orphan StringName warnings
