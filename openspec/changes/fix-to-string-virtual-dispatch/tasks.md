# Tasks

## 1. Fix The Validity Flag Write-Through
- [x] 1.1 In `GoCallback_ClassCreationInfoToString` (`pkg/core/classdb_callback.go`), replace the local rebind `r_is_valid = (*C.GDExtensionBool)(&isValid)` with a write through the pointer: `*r_is_valid = 1` on every path that supplies a string.

## 2. Dispatch The User's to_string Virtual
- [x] 2.1 In `GoCallback_ClassCreationInfoToString`, look up the instance's class in `Internal.GDRegisteredGDClasses` and resolve `ci.VirtualMethodMap["to_string"]`; if present, invoke it via `GoMethodMetadata.Func.Call([]reflect.Value{reflect.ValueOf(inst)})` (same pattern as `_notification`) and use the returned Go `string`.
- [x] 2.2 If the class is not registered or has no `to_string` virtual, fall back to the existing default format `[ GDExtension::<class> <--> Instance ID:<id> ]`, still reported valid.
- [x] 2.3 Guard the reflection return: if the resolved virtual's return value is not a Go `string`, log a warning and use the fallback format instead of panicking.
- [x] 2.4 Replace `GDExtensionStringPtrWithLatin1Chars` with `GDExtensionStringPtrWithUtf8Chars` so the out Godot `String` is constructed from the Go string's UTF-8 bytes.

## 3. Tests
- [x] 3.1 Extract the resolution logic (virtual lookup + reflection + fallback) into a testable Go helper (e.g. `resolveToString(inst) string`) that the cgo callback calls, so it can be unit-tested without a live engine handle.
  - Landed as `pkg/core/to_string.go`: `resolveToString(inst Object)` plus `resolveToStringFor(recv any, className string, instanceId uint64)`. The split keeps the dispatch rules reachable without a live engine object behind the receiver.
- [x] 3.2 Add a Go unit test (e.g. `pkg/core/classdb_to_string_test.go`) proving dispatch: register a class whose `to_string` virtual returns a distinctive marker string, call the helper, and assert the marker — not the default format — is produced; assert a class with no virtual produces the non-empty default format.
  - Six cases: bound virtual returns marker; no virtual bound; class unregistered; virtual returns non-string; virtual returns two values; default-format shape pinned.
- [x] 3.3 Confirm the existing acceptance assertion at `test/demo/main.gd:28` now passes.
  - **Strengthened rather than merely confirmed.** The original expectation was byte-identical to the binding's fallback format, so it passed whether or not dispatch ran — it could not fail for the bug this change fixes. `V_Example_ToString` now returns `[ GDExtension::Example <--> Instance ID:<id> | é中 ]`, a string only the virtual produces, and the assertion was updated to match. This also covers 2.4: `é` (2 bytes) and `中` (3 bytes) round-trip through the UTF-8 constructor, which the previous ASCII-only assertion left untested.
  - Falsified: with the virtual lookup sabotaged to always miss, the suite reports 1078 passes / 1 failure and exits 2. Restored, it reports 1079 / 0 and exits 0.

## 4. Verification
- [x] 4.1 Run `go build ./...` — clean build.
- [x] 4.2 Run `GODOT=/home/pcting/bin/godot make test` — full suite green with the `main.gd:28` assertion passing; ignore the known startup warnings (`get_variant_get_internal_ptr_func`, `callable_custom_get_user_data`, `classdb_register_extension_class_5/6`).
  - **1079 passes, 0 failures, 0 leaked instances, 0 panics.** First fully green run of this suite.
