## Why

godot-go has a `TypedRef[T]` type that exists but is incomplete and unused. It lacks GDExtension Ref integration (can't pass `Ref` to Godot API functions), has no variant encoding/decoding, and doesn't properly bridge Godot's reference-counted object lifecycle. This blocks building any Go types that extend `RefCounted` (Resources, etc.) with correct memory management.

**Concrete problems today:**
- Returning `RefImage` from a Go method to Godot **panics** — `GDExtensionTypePtrFromReflectValue` has no handler for Ref types
- Godot's `referenceFunc`/`unreferenceFunc` class creation callbacks are `nil` — when Godot copies/destroys a `Ref` to a user-defined class, the refcount is never updated, causing use-after-free or memory leaks
- `TypedRef[T]` API is awkward: `ref.TypedPtr().GetWidth()` instead of `ref.Ptr().GetWidth()`
- Binding reference callbacks (`GoCallback_GDClassBindingReference`) are no-ops

## What Changes

### Core `Ref[T]` Type
- **Replace `TypedRef[T]` with `Ref[T]`**: Stores `T` directly (not `RefCounted` base), eliminating type assertions. `Ptr()` returns `T` for transparent access (godot-cpp `operator->` equivalent).
- **Two constructors**: `NewRef[T](obj)` for Godot-owned objects (calls `Reference()`), `NewRefInit[T](obj)` for user-owned objects (calls `InitRef()` then `Reference()`).
- **Pure finalizer lifecycle**: `runtime.SetFinalizer` guarantees exactly-one `Unreference()` per `*Ref[T]` allocation. No manual `Unref()` — Go GC manages the entire lifecycle.
- **Clone for sharing**: `Clone()` creates an independent claim on the same underlying object. Each clone has its own finalizer and calls `Reference()`.

### GDExtension Ref Integration
- **Wire `referenceFunc`/`unreferenceFunc`** in `NewGDExtensionClassCreationInfo4` — currently `nil`, which causes refcount drift for user-defined `RefCounted` classes. Callbacks extract `WrappedClassInstance` and call `RefCounted.Reference()`/`Unreference()`.
- **Fix binding reference callbacks** — `GoCallback_GDClassBindingReference` and `GoCallback_GDExtensionBindingReference` are no-ops. Wire to update Go-side binding references.

### Ref Encoding (Go→Godot)
- **Extract `Ptr()` as `Object*`**: In `GDExtensionTypePtrFromReflectValue`, detect `Ref` types and encode `ref.Ptr()` as raw object pointer. Mirrors ptrcall semantics — Godot handles its own `Ref` wrapping internally.
- **Variant encoding**: `Ref[T]` maps to `VARIANT_TYPE_OBJECT` in `ReflectTypeToGDExtensionVariantType` (already partially handled, needs Ref detection before Object check).

### Codegen Updates
- **Generated `RefImageImpl TypedRef[Image]`** → **`RefImage struct { *Ref[Image] }`** — thin wrappers via embedding. Eliminates ~15 delegation methods per type (~3000 lines → ~600 lines across ~200 types).
- **Template changes**: `cmd/generate/gdclassimpl/` templates for Ref type generation. No `Unref()`/`Ref()` delegation needed — finalizer handles everything.

### Test Coverage
- Ref counting: create, clone, GC collect, verify refcount via `GetReferenceCount()`.
- Godot round-trip: Go returns `Ref[T]`, Godot receives it correctly.
- User-defined `RefCounted`: custom class extending `Resource` with `Ref[CustomResource]`.
- Binding lifecycle: Godot copies/destroys `Ref` to user class, refcount stays correct.

## Capabilities

### New Capabilities
- `ref-smart-pointer`: Implements `Ref[T]` as a type-safe reference-counted smart pointer for Godot `RefCounted` objects, with finalizer-based lifecycle management, variant encoding, and GDExtension integration.

### Modified Capabilities
- None

## Impact

### Files Changed
| File | Change |
|------|--------|
| `pkg/builtin/ref_generic.go` | Rewrite `TypedRef[T]` → `Ref[T]` with `runtime.SetFinalizer`, no `Unref()` |
| `pkg/builtin/lib.go` | Update `RefCountedConstructor` type signature |
| `pkg/builtin/wrapped_gdclass.go` | Fix `GoCallback_GDClassBindingReference` to call `Reference()`/`Unreference()` |
| `pkg/gdclassinit/wrapped_gdextension_class.go` | Fix `GoCallback_GDExtensionBindingReference` |
| `pkg/core/classdb.go` | Wire `referenceFunc`/`unreferenceFunc` in `ClassDBRegisterClass` |
| `pkg/core/classdb_callback.go` | Add `GoCallback_ClassCreationInfoReference`/`Unreference` C exports |
| `pkg/core/method_bind_reflect.go` | Add Ref type detection in return encoding (`GDExtensionTypePtrFromReflectValue`) |
| `pkg/core/variant_reflect_type.go` | Ensure `Ref[T]` → `VARIANT_TYPE_OBJECT` mapping |
| `cmd/generate/gdclassimpl/templatefunctions.go` | Update `goEncodeIsReference` and related helpers |
| `cmd/generate/gdclassimpl/templates/*.tmpl` | Regenerate `RefImage` etc. as embedding structs |
| `test/demo/` | Add Ref lifecycle test cases |

### Systems Affected
- Method binding (encode/decode Ref types)
- Variant encoding (Ref → Object)
- Object lifecycle management (refcount callbacks)
- Code generation (~200 Ref types regenerated)

### Breaking Changes
- `TypedRef[T]` renamed to `Ref[T]` with API changes (no production usage)
- Generated `RefImageImpl` types become thin wrapper structs with embedding
- `NewTypedRef[T]` / `NewTypedRefGDExtensionIternalConstructor[T]` replaced with `NewRef[T]` / `NewRefInit[T]`

### Verification
- 43/43 existing tests still pass
- New tests: Ref create/clone/GC lifecycle, Godot round-trip, user-defined RefCounted class
- `GetReferenceCount()` verifies correct refcount at each step
- No memory leaks on extension unload (valgrind)
