# Tasks

## 1. Fix The Validity Flag Write-Through
- [ ] 1.1 In `GoCallback_ClassCreationInfoToString` (`pkg/core/classdb_callback.go`), replace the local rebind `r_is_valid = (*C.GDExtensionBool)(&isValid)` with a write through the pointer: `*r_is_valid = 1` on every path that supplies a string.

## 2. Dispatch The User's to_string Virtual
- [ ] 2.1 In `GoCallback_ClassCreationInfoToString`, look up the instance's class in `Internal.GDRegisteredGDClasses` and resolve `ci.VirtualMethodMap["to_string"]`; if present, invoke it via `GoMethodMetadata.Func.Call([]reflect.Value{reflect.ValueOf(inst)})` (same pattern as `_notification`) and use the returned Go `string`.
- [ ] 2.2 If the class is not registered or has no `to_string` virtual, fall back to the existing default format `[ GDExtension::<class> <--> Instance ID:<id> ]`, still reported valid.
- [ ] 2.3 Guard the reflection return: if the resolved virtual's return value is not a Go `string`, log a warning and use the fallback format instead of panicking.
- [ ] 2.4 Replace `GDExtensionStringPtrWithLatin1Chars` with `GDExtensionStringPtrWithUtf8Chars` so the out Godot `String` is constructed from the Go string's UTF-8 bytes.

## 3. Tests
- [ ] 3.1 Extract the resolution logic (virtual lookup + reflection + fallback) into a testable Go helper (e.g. `resolveToString(inst) string`) that the cgo callback calls, so it can be unit-tested without a live engine handle.
- [ ] 3.2 Add a Go unit test (e.g. `pkg/core/classdb_to_string_test.go`) proving dispatch: register a class whose `to_string` virtual returns a distinctive marker string, call the helper, and assert the marker — not the default format — is produced; assert a class with no virtual produces the non-empty default format.
- [ ] 3.3 Confirm the existing acceptance assertion at `test/demo/main.gd:28` (`example.to_string()` equals `[ GDExtension::Example <--> Instance ID:<id> ]`) now passes, proving the dispatched `V_Example_ToString` value reaches Godot through the validity-fixed callback.

## 4. Verification
- [ ] 4.1 Run `go build ./...` — clean build.
- [ ] 4.2 Run `GODOT=/home/pcting/bin/godot make test` — full suite green with the `main.gd:28` assertion passing; ignore the known startup warnings (`get_variant_get_internal_ptr_func`, `callable_custom_get_user_data`, `classdb_register_extension_class_5/6`).
