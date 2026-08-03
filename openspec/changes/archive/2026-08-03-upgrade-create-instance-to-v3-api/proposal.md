## Why

The current codebase uses raw C strings for class userdata while `cgo.Handle` provides a safer, reference-counted alternative for the C/Go boundary. The `GDExtensionClassCreationInfo6` and `CreateInstance3` types are generated but not yet usable (Godot 4.7.2.rc binary lacks `classdb_register_extension_class_6`). This change prepares the codebase with the modern callback patterns — `cgo.Handle`-based instance retrieval, `SetConstructInfo`, `ClassUserdata` — plus defensive cgo/pinning fixes for Go 1.25+, while keeping the existing `CreateInstance2` naming and `CreationInfo4` registration path.

## What Changes

- Keep `GDExtensionClassCreateInstance2` for class instance construction callbacks (v6 registration path not yet available)
- Add `NewGDExtensionClassCreationInfo6` constructor to `pkg/ffi/class_create_info.go` (available for future use)
- Introduce `ClassUserdata` struct for structured per-class userdata (defined but not yet wired as `class_userdata` — see design.md Decision 5)
- Add `cgo.Handle`-based instance retrieval in `FreeInstance`, `GetPropertyList`, and other per-instance callbacks
- Add `SetConstructInfo` function matching godot-cpp's `Wrapped::_set_construct_info` pattern
- Add generic `CreateGDClassInstance[T GDClass]() T` for type-safe instance creation
- Add string-based `CreateGDClassInstance2(tn string) GDClass` for use from the `CreateInstance2` callback
- Keep `Internal.GDClassInstances` SyncMap as a re-entrant `FreeInstance` guard (prevents double-Destroy → C heap corruption)
- Restore `pnr.Pin` calls on cgo binding callback pointers in `GDClassRegisterInstanceBindingCallbacks` (required for Go 1.25+ cgo pointer rules)
- Add `GOAMD64=v3` + `-ldflags="-X runtime.godebugDefault=cpu.avx512f=off"` to build targets (Valgrind/glibc compatibility)
- Fix `ClassInfo.HasError()` to check `c.Name` field
- Fix `StringName` argument decoding in `method_bind_reflect.go` to use value copy
- Remove unused `test/pkg_test/lib.go`
- Remove `util.CgoTestCall` debug calls from `variant.go`

## Capabilities

### New Capabilities
- `class-instance-lifecycle`: GDExtension class instance creation with `cgo.Handle`-based instance retrieval, `SetConstructInfo` pattern, `ClassUserdata` type for future v6 migration, and `GDClassInstances` SyncMap as a re-entrant `FreeInstance` guard

### Modified Capabilities
<!-- None - existing specs unchanged -->

## Impact

- **`pkg/core/`**: `classdb.go`, `classdb_callback.go`, `classdb_callback.c`, `classdb_callback.h`, `lib.go`, `godot.go`, `types.go`, `method_bind_reflect.go`
- **`pkg/builtin/`**: `wrapped.go`, `wrapped_gdclass.go`, `char_string.go`, `variant.go`, `lib.go`
- **`pkg/ffi/`**: `ffi_wrapper.gen.go`, `lib.go`, `class_create_info.go`
- **Build**: `Makefile` (GOAMD64=v3, AVX-512 ldflags, test target restored, build-asan/test-asan targets)
- **Test**: Remove `test/pkg_test/lib.go`

## Non-goals

- No changes to existing spec-level behavior for StringName/NodePath lifecycle
- No changes to virtual method dispatch or signal handling
- Migration to `CreationInfo6` + `RegisterExtensionClass6` is deferred until `classdb_register_extension_class_6` is available
- `ClassUserdata` as `class_userdata` is deferred due to `morestack on g0` crash
