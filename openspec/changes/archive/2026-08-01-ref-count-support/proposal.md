## Why

godot-go's `TypedRef[T]` exists but is incomplete: Go methods cannot return a `Ref` to Godot (the return encoder panics), the `Ref` API is awkward and requires unsafe type assertions, and there is no way for Go-held references to be released when a `Ref` becomes unreachable. This blocks building Go types that extend `RefCounted` (Resources, custom ref-counted types) with correct memory management.

**Concrete problems today (all verified against source):**

- **Returning a `Ref` from a Go method to Godot panics** — `GDExtensionTypePtrFromReflectValue` (`pkg/core/variant_refect_value.go:52`) and `GDExtensionVariantPtrFromReflectValue` (`:267`) have no handler for `Ref` interface values, so both the ptrcall and varcall return paths hit `default: log.Panic`.
- **No deterministic release for Go-held references** — `TypedRef` has no finalizer and relies on explicit `Unref()`. A `Ref` copied into Go and then dropped leaks one reference.
- **`TypedRef[T]` requires a runtime type assertion** — `TypedPtr()` does `cx.Reference.(T)` (`pkg/builtin/ref_generic.go`), which panics if the stored object isn't actually `T`.
- **Generated `RefXImpl` types delegate ~15 methods each** across ~200 types (~3000 lines of boilerplate).

The inverse direction (Godot→Go) is already implemented correctly: both the ptrcall and Variant decode paths (`pkg/core/method_bind_reflect.go:164-184`, `:509-554`) wrap refs with **transfer semantics** (no refcount increment) via `GDClassRefConstructors`.

**Non-problems the original proposal incorrectly identified:** Godot's class-creation `referenceFunc`/`unreferenceFunc` callbacks and the instance-binding reference callbacks are **optional notifications**, not the refcount mechanism. Godot's `RefCounted::reference()/unreference()` manage the count internally (`../godot/core/object/ref_counted.cpp:74-108`); the extension callbacks are only invoked as post-hoc notifications. godot-cpp ships all of them as `nullptr`/no-ops (`../godot-cpp/include/godot_cpp/core/class_db.hpp:280-281`) and works. Wiring them to call `RefCounted.Reference()/Unreference()` (as the original proposal did) would re-enter the same refcount function and double-count every copy/destroy, and the binding callbacks cannot even reach the Go wrapper (their second parameter is the binding data, which is `nullptr` for godot-go). These must NOT be wired.

## What Changes

### Core `Ref[T]` Type
- **Replace `TypedRef[T]` with `Ref[T]`**: Stores `T` directly (not the `RefCounted` base), eliminating the runtime type assertion. `Ptr()` returns `T`.
- **`Ref` interface split**: Remove `Ptr() RefCounted` from the `Ref` interface (it cannot coexist with a concrete `Ptr() T`); add `ToObject() RefCounted` for boundary encoding. The interface becomes the refcount protocol: `ToObject() RefCounted; Ref(from Ref); Unref(); IsValid()`.
- **Finalizer-backed release**: `NewRefInit` and `Ref(from)` install `runtime.SetFinalizer` so an unreachable Go-held reference is released, emulating godot-cpp's destructor. `Unref()` clears the finalizer and zeroes the stored ref (idempotent).
- **Ownership constructors (godot-cpp semantics, keeping today's behavior):**
  - `NewRefInit[T](obj)` — user-created object: `obj.InitRef()` (marks ref-counted, +1), then stores; owns the reference.
  - `NewRef[T](obj)` — Godot-owned/transfer object: stores **without** changing the refcount (godot-cpp `_gde_internal_constructor`); used across the Go↔Godot boundary where the other side already holds the reference.
  - `Ref(from)` (copy) — `Reference()` (+1); `Unref()` — `Unreference()` (-1).

### GDExtension Ref Integration
- **Fix Go→Godot return encoding** (ptrcall and varcall): detect `Ref` interface values in `GDExtensionTypePtrFromReflectValue` / `GDExtensionVariantPtrFromReflectValue`, extract `ToObject()`, and encode as an `Object`. The ptrcall return slot is the raw `Object*` inside a pre-initialized Variant (`../godot/core/extension/gdextension.cpp:132-142`), so encoding the object pointer directly is correct; Godot wraps its own `Ref`.
- **Map `Ref[T]` to `VARIANT_TYPE_OBJECT`** in `ReflectTypeToGDExtensionVariantType`.
- **Leave the class-creation and binding reference callbacks untouched** (`nil` / no-op) — verified correct, godot-cpp parity.

### Codegen Updates
- **Generated `RefImageImpl TypedRef[Image]`** → **`type RefImage struct { *Ref[Image] }`** — thin wrappers via embedding. The ~15 delegation methods per type disappear (~3000 lines → ~600 lines across ~200 types).
- **Template changes**: `cmd/generate/gdclassimpl/` templates for Ref type generation.

### Test Coverage
- Ref counting: create, copy, unref, verify refcount via `GetReferenceCount()`.
- Godot round-trip: Go returns `Ref[T]`, Godot receives it correctly (ptrcall and varcall).
- User-defined `RefCounted`: custom class extending `Resource` with `Ref[CustomResource]`.
- Go-held reference release: drop a `Ref`, force GC, verify refcount drops (finalizer).

## Capabilities

### New Capabilities
- `ref-smart-pointer`: Implements `Ref[T]` as a type-safe reference-counted smart pointer for Godot `RefCounted` objects, with copy/unref semantics, finalizer-backed release, and GDExtension encoding/decoding.

### Modified Capabilities
- None

## Impact

### Files Changed
| File | Change |
|------|--------|
| `pkg/builtin/ref_generic.go` | Rewrite `TypedRef[T]` → `Ref[T]`, split `Ref` interface, add `NewRef`/`NewRefInit` and finalizers |
| `pkg/builtin/classes.ref.interfaces.gen.go` | Regenerate per-class `RefX` interfaces (drop `Ptr()`, keep concrete `TypedPtr()`) |
| `pkg/core/variant_refect_value.go` | Add `Ref` handling to ptrcall + varcall return encoders |
| `pkg/core/variant_reflect_type.go` | Ensure `Ref[T]` → `VARIANT_TYPE_OBJECT` mapping |
| `pkg/core/method_bind_reflect.go` | Verify decode paths still use transfer constructors; no double-reference |
| `cmd/generate/gdclassimpl/templatefunctions.go` | Update Ref helpers |
| `cmd/generate/gdclassimpl/templates/*.tmpl` | Regenerate `RefImage` etc. as embedding structs |
| `test/demo/` | Add Ref lifecycle, round-trip, finalizer, and user-defined RefCounted tests |

### Systems Affected
- Method binding (encode/decode Ref types)
- Variant encoding (Ref → Object)
- Ref-counted object lifecycle (Go-held reference release)
- Code generation (~200 Ref types regenerated)

### Breaking Changes
- `TypedRef[T]` renamed to `Ref[T]`; `Ptr()` becomes concrete, `Ref` interface loses `Ptr()`, gains `ToObject()`.
- Generated `RefImageImpl` types become embedding structs; per-class `RefX` interfaces change.
- `NewTypedRef[T]` / `NewTypedRefGDExtensionIternalConstructor[T]` replaced by `NewRef[T]` / `NewRefInit[T]`. No production usage today.

### Verification
- 43/43 existing tests still pass
- New tests: Ref create/copy/unref lifecycle, Godot round-trip (ptrcall + varcall), user-defined RefCounted class, finalizer-backed release
- `GetReferenceCount()` verifies correct refcount at each step
- No memory leaks on extension unload (valgrind)
