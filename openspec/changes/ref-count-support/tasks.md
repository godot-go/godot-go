# Tasks

## 1. Core `Ref[T]` Type

- [ ] 1.1 Rewrite `pkg/builtin/ref_generic.go`: replace `TypedRef[T]` with `Ref[T] struct { m_ref T }` storing the concrete type
- [ ] 1.2 Split the `Ref` interface: remove `Ptr()`, add `ToObject() RefCounted`; keep `Ref(from Ref)`, `Unref()`, `IsValid()`
- [ ] 1.3 Implement `NewRefInit[T](obj)` — `InitRef()` + store + `runtime.SetFinalizer` (user-owned)
- [ ] 1.4 Implement `NewRef[T](obj)` — transfer semantics, no refcount change, no finalizer (Godot-owned)
- [ ] 1.5 Implement `Ref(from Ref)` copy with self-assignment guard, `Reference()`, finalizer install
- [ ] 1.6 Implement `Unref()` — `Unreference()`, clear finalizer, zero `m_ref` (idempotent)
- [ ] 1.7 Remove `pnr.Pin` from the ref constructors

## 2. GDExtension Ref Integration

- [ ] 2.1 Add `Ref` detection to `GDExtensionTypePtrFromReflectValue` (`pkg/core/variant_refect_value.go`) — encode `ToObject()` as Object via `ObjectEncoder`, handle nil
- [ ] 2.2 Add `Ref` detection to `GDExtensionVariantPtrFromReflectValue` — encode via `ObjectEncoder.EncodeVariantPtrArg`, handle nil
- [ ] 2.3 Ensure `Ref[T]` maps to `VARIANT_TYPE_OBJECT` in `ReflectTypeToGDExtensionVariantType` (`pkg/core/variant_reflect_type.go`)
- [ ] 2.4 Verify decode paths (`pkg/core/method_bind_reflect.go`) still compile and use transfer constructors after the `Ref` interface split; update `GDClassRefConstructors` wiring if needed
- [ ] 2.5 Leave `referenceFunc`/`unreferenceFunc` nil and binding reference callbacks as no-ops (no change)

## 3. Codegen Updates

- [ ] 3.1 Update `cmd/generate/gdclassimpl/templates/*.tmpl` to generate `type RefImage struct { *Ref[Image] }` embedding wrappers
- [ ] 3.2 Update `cmd/generate/gdclassimpl/templatefunctions.go` ref helpers (constructors `NewRefX`, `NewRefXAsRef`, `NewRefXGDExtensionIternalConstructor`)
- [ ] 3.3 Regenerate `pkg/gdclassimpl/classes.refs.gen.go` and `pkg/builtin/classes.ref.interfaces.gen.go` with `make generate`

## 4. Tests

- [ ] 4.1 Add Ref create/copy/unref lifecycle test verifying `GetReferenceCount()` at each step
- [ ] 4.2 Add double-unref safety test (explicit + finalizer)
- [ ] 4.3 Add Go→Godot round-trip test: Go method returns `RefImage` (ptrcall and varcall), Godot receives it
- [ ] 4.4 Add nil Ref return test
- [ ] 4.5 Add user-defined `RefCounted` test: custom `Resource` class with `Ref[CustomResource]`, verify Godot-side Ref copy/destroy refcount
- [ ] 4.6 Add finalizer release test: drop `Ref`, force GC, verify refcount drops

## 5. Validation

- [ ] 5.1 Run `make generate` and confirm no diff drift in `.gen.*` files
- [ ] 5.2 Run `make test` — all 43 existing tests still pass plus new tests
- [ ] 5.3 Verify `cgocheck=1` runs clean (no cgo pointer rule violations)
- [ ] 5.4 Update `docs/` if the Ref API is documented anywhere
