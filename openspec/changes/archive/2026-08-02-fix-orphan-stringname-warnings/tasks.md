## 1. Fix return encoding path

- [x] 1.1 Restore `inst.Destroy()` after `StringNameEncoder.EncodeVariantPtrArg` in `GDExtensionVariantPtrFromReflectValue` for StringName and NodePath cases
- [x] 1.2 Replace `EncodeTypePtrArg` with `CopyConstructor(rOut, ...) + inst.Destroy()` for StringName/NodePath in `GDExtensionTypePtrFromReflectValue` Struct branch
- [x] 1.3 Replace `EncodeTypePtrArg` with `CopyConstructor(rOut, ...) + inst.Destroy()` for StringName/NodePath in `GDExtensionTypePtrFromReflectValue` Array branch

## 2. Fix argument decoding path

- [x] 2.1 Revert varcall StringName decode to byte-copy + Destroy borrow semantics in `convertVariantToGoTypeReflectValue`
- [x] 2.2 Revert varcall NodePath decode to byte-copy + Destroy borrow semantics in `convertVariantToGoTypeReflectValue`
- [x] 2.3 Revert ptrcall StringName decode to byte-copy borrow semantics in `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`
- [x] 2.4 Revert ptrcall NodePath decode to byte-copy borrow semantics in `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`

## 3. Verify

- [x] 3.1 Run `make build` to verify compilation
- [x] 3.2 Run `make test` to verify 0 orphan StringName warnings and all 595 tests pass
- [x] 3.3 Verify no "corrupted double-linked list" or other memory corruption errors at exit
