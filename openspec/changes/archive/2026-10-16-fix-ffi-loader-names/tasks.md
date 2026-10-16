# Tasks

## 1. Header metadata extraction

- [x] 1.1 Implement the `@name` map builder in `cmd/generate/ffi`: line scan of
      `godot_headers/godot/gdextension_interface.h` (relative to `projectPath`)
      pairing each interface-function typedef identifier with the `@name` value of
      its preceding doc block; missing header yields an empty map. Verify with new
      fixture-based unit tests covering: plain pairing, legacy-identifier typedef
      (`GDExtensionInterfaceGetVariantGetInternalPtrFunc` ->
      `variant_get_ptr_internal_getter`), multi-line typedef declaration, typedef
      with no preceding `@name` (absent from map), missing header file.
- [x] 1.2 Register an `interfaceFunctionName` template function (map hit wins,
      else fall back to `loadProcAddressName`) and switch the load line in
      `cmd/generate/ffi/ffi.go.tmpl` to use it. Verify with unit tests: metadata
      wins over identifier reconstruction; fallback emits the legacy transform.

## 2. Coverage guard and regeneration

- [x] 2.1 Add the coverage test: for the bundled header, every typedef returned by
      `CollectGDExtensionInterfaceFunctions` has a `@name` entry; on failure the
      test names the uncovered typedefs. Verify it passes on the current header and
      fails on a fixture with one `@name` removed.
- [x] 2.2 Run `make generate` and audit `pkg/ffi/ffi.gen.go`: exactly four
      `LoadProcAddress` strings change (`callable_custom_get_user_data` ->
      `callable_custom_get_userdata`, `classdb_register_extension_class_5` ->
      `classdb_register_extension_class5`, `classdb_register_extension_class_6` ->
      `classdb_register_extension_class6`, `get_variant_get_internal_ptr_func` ->
      `variant_get_ptr_internal_getter`); `pkg/ffi/ffi_wrapper.gen.*` unchanged. If
      any other line differs, stop and reconcile against the engine
      `REGISTER_INTERFACE_FUNC` inventory before continuing.
- [x] 2.3 Run `go test ./cmd/...` and `go vet ./...` - generator package green.

## 3. Integration and docs

- [x] 3.1 Update the AGENTS.md "Known Issues" block: remove the tolerated-error
      entries for `get_variant_get_internal_ptr_func`,
      `callable_custom_get_user_data`, and
      `classdb_register_extension_class_5/_6` (all resolved by this change).
      Verify the block no longer lists any name that now loads.
- [x] 3.2 Run `GODOT=/home/pcting/bin/godot make test` (background; full cgo build
      takes minutes). Verify: gate passes (extension loaded, positive PASSES, zero
      FAILURES, shutdown marker, no leaks) and the run log contains no "Attempt to
      get non-existent interface function" error lines at all.
- [x] 3.3 Commit on `fix-ffi-loader-names` with a `fix:` subject (generator load
      names from header `@name` metadata; regenerated `ffi.gen.go`; tests;
      AGENTS.md).

## 4. Wrap-up

- [x] 4.1 Write `openspec/changes/fix-ffi-loader-names/retrospective.md`
      (investigation-heavy change per config rules): the `LoadProcAddress` vs
      engine-registration audit method (want/have sets), the four before/after
      name pairs, why fixups drifted, and the C1 de-risking conclusion. Keep under
      1500 words.
