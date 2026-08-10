## 1. GDExtension API Surface

- [x] 1.1 Add `NewGDExtensionClassCreationInfo6` constructor to `pkg/ffi/class_create_info.go`, using `GDExtensionClassCreateInstance3` for the CreationInfo6 field type (available for future use)
- [x] 1.2 Rename `cgo_classcreationinfo_createinstance` → `cgo_classcreationinfo_createinstance2` in `classdb_callback.c` and `classdb_callback.h` (matches `GDExtensionClassCreateInstance2`)
- [x] 1.3 Rename `GoCallback_ClassCreationInfoCreateInstance` → `GoCallback_ClassCreationInfoCreateInstance2` in `classdb_callback.go`

## 2. Instance Lifecycle Refactor

- [x] 2.1 Add `SetConstructInfo` function in `pkg/builtin/wrapped_gdclass.go` matching godot-cpp pattern (with pnr.Pin on cbs)
- [x] 2.2 Add generic `CreateGDClassInstance[T GDClass]() T` using `SetConstructInfo`
- [x] 2.3 Add string-based `CreateGDClassInstance2(tn string) GDClass` using `WrappedPostInitialize`
- [x] 2.4 Restore `WrappedPostInitialize` with pnr.Pin calls (compat for CreateInstance2 callback)
- [x] 2.5 Restore `Internal.GDClassInstances` SyncMap as re-entrant FreeInstance guard
- [x] 2.6 Restore `pnr.Pin` calls in `GDClassRegisterInstanceBindingCallbacks` (Go 1.25+ cgo requirement)

## 3. FreeInstance Safety

- [x] 3.1 Use `defer Destroy()` in `FreeInstance` so guard catches re-entrant calls before Destroy fires
- [x] 3.2 Add SyncMap existence check in `FreeInstance` — panic if instance not found
- [x] 3.3 Intentionally omit `ptrHandle.Delete()` — prevents stale-pointer corruption

## 4. Build Configuration

- [x] 4.1 Add `GOAMD64=v3` to build targets (caps Go compiler at AVX2)
- [ ] 4.2 Add `-ldflags="-X runtime.godebugDefault=cpu.avx512f=off"` to disable AVX-512 in Go runtime (deferred — `GOAMD64=v3` is used instead via CGO_CFLAGS)
- [ ] 4.3 Add `build-asan` and `test-asan` Makefile targets (deferred)
- [x] 4.4 Add `vgcore.*` to `.gitignore`

## 5. Bug Fixes

- [x] 5.1 Fix `ClassInfo.HasError()` to check `c.Name` field
- [x] 5.2 Fix `StringName` argument decoding in `method_bind_reflect.go` (value copy)
- [x] 5.3 Remove unused `test/pkg_test/lib.go`
- [x] 5.4 Remove `util.CgoTestCall` debug calls from `variant.go`

## 6. Verification

- [x] 6.1 Run `make build` and fix any compilation errors
- [x] 6.2 Run `make test` (595/0 passes)
