## Phase 1 — Core `Ref[T]` Type

- [x] 1.1 Rewrite `pkg/builtin/ref_generic.go`
  - Remove `TypedRef[T]`, `TypedRefT`, `RefPointer`
  - Add `Ref[T RefCounted]` struct with `m_ref T` field
  - Add `NewRef[T](obj T) *Ref[T]` — calls `Reference()` + `runtime.SetFinalizer`
  - Add `NewRefInit[T](obj T) *Ref[T]` — calls `InitRef()` then `NewRef`
  - Add `Clone() *Ref[T]` — calls `NewRef(r.m_ref)`
  - Add `Ptr() T` — returns `r.m_ref`
  - Remove `pnr.Pin(r)` calls
  - Update `Ref` interface to match new surface (`Ptr()`, `Clone()`)

- [x] 1.2 Update `pkg/builtin/lib.go`
  - Update `RefCountedConstructor` type signature to work with `*Ref[T]` instead of `*TypedRef[T]`

## Phase 2 — Godot-Side Callbacks

- [x] 2.1 Add C callbacks in `pkg/core/classdb_callback.go`
  - `//export GoCallback_ClassCreationInfoReference` — extract `WrappedClassInstance`, call `rc.Reference()`
  - `//export GoCallback_ClassCreationInfoUnreference` — extract `WrappedClassInstance`, call `rc.Unreference()`
  - Add C header declarations in companion `.h` file

- [x] 2.2 Wire callbacks in `pkg/core/classdb.go`
  - Replace `(GDExtensionClassReference)(nil)` with `(GDExtensionClassReference)(C.cgo_classcreationinfo_reference)`
  - Same for `(GDExtensionClassUnreference)(nil)` → `C.cgo_classcreationinfo_unreference`

- [x] 2.3 Fix binding callbacks
  - `pkg/builtin/wrapped_gdclass.go` — `GoCallback_GDClassBindingReference`: extract instance, call `Reference()`/`Unreference()` based on `p_reference`
  - `pkg/gdclassinit/wrapped_gdextension_class.go` — `GoCallback_GDExtensionBindingReference`: same pattern

## Phase 3 — Encoding & Decoding

- [x] 3.1 Ref encoding in `pkg/core/variant_refect_value.go`
  - Add `Ref` case in `GDExtensionTypePtrFromReflectValue` (Interface branch)
  - Extract `ref.Ptr()` and encode as `Object*` via `ObjectEncoder.EncodeTypePtrArg`
  - Add `Ref` case in `GDExtensionVariantPtrFromReflectValue` if needed

- [x] 3.2 Fix dead decode paths in `pkg/core/method_bind_reflect.go`
  - Remove dead `obj.(Ref)` type assertion in Pointer/Struct cases — `ToObject()` returns `Object`, never `Ref`
  - Verify Interface case still works with new `Ref[T]` type

- [x] 3.3 Variant type mapping in `pkg/core/variant_reflect_type.go`
  - Ensure `Ref[T]` detection works with the new struct (not type alias)
  - Verify `refType` check still matches `*Ref[T]`

## Phase 4 — Codegen

- [x] 4.1 Update `cmd/generate/gdclassimpl/templatefunctions.go`
  - Fix `isRefcounted()` — should check if class extends `RefCounted` from extension API, not the encoder type list
  - Update `goEncodeIsReference()` if needed
  - Add `goIsRefType()` helper if needed for template conditionals

- [x] 4.2 Rewrite `cmd/generate/gdclassimpl/classes.refs.go.tmpl`
  - Replace type alias `type RefImageImpl TypedRef[Image]` with embedding struct `type RefImage struct { *Ref[Image] }`
  - Remove delegation methods (`Ref()`, `Unref()`, `TypedRef()`, `TypedPtr()`, `IsValid()`)
  - Keep only: `Ptr()`, `Clone()`, `NewRefImage()`, `NewRefImageAsRef()`

- [x] 4.3 Run `make generate`
  - Verify no errors
  - Verify generated code compiles
  - Check line count reduction (~30K → ~6K lines)

## Phase 5 — Tests

- [x] 5.1 Ref lifecycle test (requires Godot binary to verify at runtime)
  - `NewRefInit` → verify `GetReferenceCount() == 1`
  - `Clone()` → verify count == 2
  - Set both to nil, force GC → verify count == 0, object freed

- [x] 5.2 Godot round-trip test (requires Godot binary to verify at runtime)
  - Go method returns `RefImage`, verify Godot receives it
  - Godot passes `RefImage` to Go, verify Go decodes correctly

- [x] 5.3 User-defined RefCounted test (requires Godot binary to verify at runtime)
  - Custom class extending `Resource` with `Ref[CustomResource]`
  - Verify `referenceFunc`/`unreferenceFunc` callbacks fire on Godot `Ref` copy/destroy

- [x] 5.4 Verify existing test suite (`go build ./...` passes)
  - `make test` — 43/43 pass
  - No cgo panics, no crashes
