## Context

godot-go's `TypedRef[T]` cannot return `Ref` types to Godot (the return encoders panic), has no deterministic release for Go-held references, and requires runtime type assertions. The Godot→Go direction is already implemented with correct **transfer semantics**: both the ptrcall arg decoder and the Variant decoders wrap refs without incrementing the refcount via `GDClassRefConstructors` (`pkg/core/method_bind_reflect.go:164-184`, `:509-554`), and the generated `NewRefXAsRef` / `NewRefXGDExtensionIternalConstructor` constructors store the object without touching the refcount (`pkg/gdclassimpl/classes.refs.gen.go`).

The two callback paths the original design wanted to "wire up" are **optional notifications, not the refcount mechanism**:

```
┌────────────────────────────────────────────────────────────────────────────┐
│  Callback Path 1: Class Creation Info (referenceFunc/unreferenceFunc)      │
│  └── Invoked FROM WITHIN Godot's RefCounted::reference()/unreference()     │
│      (ref_counted.cpp:82-84, 102-104), AFTER the internal refcount changed │
│  └── Notifications only — godot-cpp ships them as nullptr                  │
│  └── Wiring them to call Reference()/Unreference() re-enters the same      │
│      refcount function → double-counting → leaks on every copy/destroy    │
│  └── Verdict: leave nil (correct today)                                    │
│                                                                             │
│  Callback Path 2: Instance Binding (GoCallback_GDClassBindingReference)     │
│  └── Signature (gdextension_interface.h:197): (token, p_binding, p_ref)    │
│  └── 2nd arg is the BINDING DATA, not the instance — nullptr for godot-go  │
│  └── Return value = "can die"; returning 1/true is already correct         │
│  └── The Go wrapper lives as the extension class instance, freed by        │
│      free_instance_func — these callbacks never see it                     │
│  └── Verdict: leave no-op (correct today)                                  │
└────────────────────────────────────────────────────────────────────────────┘
```

The real work is: (a) fix the Go→Godot return encoders, (b) give `Ref[T]` deterministic release semantics, and (c) remove the type-assertion and delegation boilerplate.

## Goals / Non-Goals

**Goals:**
- `Ref[T]` works as a drop-in, type-safe smart pointer for `RefCounted` objects.
- Go methods returning `Ref[T]` encode correctly to Godot (ptrcall and varcall).
- Go methods accepting `Ref[T]` arguments decode correctly, without adding or leaking references.
- Go-held references are released when a `Ref` becomes unreachable (finalizer).
- API is ergonomic: `ref.Ptr().GetWidth()` with no runtime type assertion.
- Support user-defined GDExtension classes that extend `RefCounted` (Resources, etc.).

**Non-Goals:**
- Wiring `referenceFunc`/`unreferenceFunc` or the binding reference callbacks — verified unnecessary and harmful (see Context).
- Changing the general-purpose `Object` lifecycle or binding system.
- Implementing custom allocators or ref-counted memory pools.
- Modifying generated code in `*.gen.*` files directly — only template changes.

## Decisions

### 1. Replace `TypedRef[T]` with `Ref[T]`, split the `Ref` interface

Current `TypedRef[T]` stores the `RefCounted` base interface and type-asserts in `TypedPtr()`:

```go
// Current — runtime type assertion, can panic
type TypedRef[T TypedRefT] struct {
    Reference RefCounted  // stores as base interface
}
func (cx *TypedRef[T]) TypedPtr() T { return cx.Reference.(T) }  // assert!
```

New `Ref[T]` stores `T` directly:

```go
// New — type-safe, no assertions
type Ref[T RefCounted] struct {
    m_ref T  // stored as concrete type
}
func (r *Ref[T]) Ptr() T          { return r.m_ref }
func (r *Ref[T]) ToObject() RefCounted { return r.m_ref }
```

**Interface split (why `Ptr()` cannot stay on `Ref`):** The reflection decode paths dispatch on `t.Implements(refType)` (`method_bind_reflect.go:29,166`). If the `Ref` interface still required `Ptr() RefCounted`, no concrete type could also expose `Ptr() T` (Go forbids two methods with the same name and different return types). We therefore remove `Ptr()` from the `Ref` interface and add `ToObject()`:

```go
type Ref interface {
    ToObject() RefCounted
    Ref(pFrom Ref)
    Unref()
    IsValid() bool
}
```

The decoder still works (`RefImage` embeds `Ref`, so it implements the `Ref` interface), and the encoder can extract the wrapped object via `ToObject()`. The per-class interface becomes:

```go
type RefImage interface {
    Ref
    TypedPtr() Image  // or Ptr() Image — see Decision 7
}
```

**Rationale**: Eliminates the runtime assertion (`Ptr()`/`TypedPtr()` returns `m_ref` directly), matches godot-cpp's `Ref<T>` ownership model, and keeps the reflection-based decode dispatch intact.

### 2. Ownership constructors — godot-cpp semantics, transfer vs init

Two constructors cover the two ownership situations. These **preserve today's semantics** (current `NewRefX` = init, `NewRefXAsRef`/`NewRefXGDExtensionIternalConstructor` = transfer) and rename them consistently:

```go
// User-owned object (newly created, e.g., NewImage(), or a user-defined Resource).
// InitRef() marks it ref-counted (+1); the Ref owns that reference.
func NewRefInit[T RefCounted](obj T) *Ref[T] {
    if !obj.InitRef() {
        return nil  // failed — object doesn't support ref counting
    }
    r := &Ref[T]{m_ref: obj}
    runtime.SetFinalizer(r, (*Ref[T]).Unref)
    return r
}

// Godot-owned object (returned from Godot API, passed across the boundary).
// Transfer: NO refcount change — the other side already holds the reference.
// Mirrors godot-cpp's _gde_internal_constructor (ref.hpp:208-212).
// SAFETY: the caller/Godot must hold a reference for this wrapper's lifetime;
//         to take ownership, copy via Ref(from) which calls Reference().
func NewRef[T RefCounted](obj T) *Ref[T] {
    r := &Ref[T]{m_ref: obj}
    return r
}
```

**Finalizer**: `NewRefInit` and `Ref(from)` install `runtime.SetFinalizer` so a Go-held reference is released when the `Ref` becomes unreachable — the Go equivalent of godot-cpp's `~Ref()`. Transfer wrappers (`NewRef`) get **no** finalizer: they own no reference, so they must not decrement.

**Rationale**: `InitRef()` can only be called once and marks the object as ref-counted. Godot-owned objects are already ref-counted. godot-cpp distinguishes exactly this with `ref_pointer<true>` (init) vs `_gde_internal_constructor` (transfer).

### 3. Ref copy and unref semantics

```go
// Copy reference from another Ref, managing refcounts.
// After this, both refs share the same underlying object.
func (r *Ref[T]) Ref(from Ref) {
    var zero T
    if from == nil || !from.IsValid() {
        r.Unref()
        return
    }
    o, ok := from.ToObject().(T)
    if !ok || o == zero {
        r.Unref()
        return
    }
    if r.m_ref == o {
        return  // self-assignment, no-op
    }
    r.Unref()
    r.m_ref = o
    r.m_ref.Reference()
    runtime.SetFinalizer(r, (*Ref[T]).Unref)
}

// Release this reference. If refcount drops to zero, the
// underlying Godot object is freed. Safe to call multiple times.
func (r *Ref[T]) Unref() {
    runtime.SetFinalizer(r, nil)
    var zero T
    if r.m_ref != nil && r.m_ref != zero {
        r.m_ref.Unreference()
    }
    r.m_ref = zero  // prevents double-unref
}
```

**Rationale**: Self-assignment guard prevents accidental double-ref. `Unref()` zeroes `m_ref` and clears the finalizer so double-unref (explicit + finalizer) is a safe no-op.

### 4. Ref encoding (Go→Godot) — the actual bug fix

For ptrcall returns, Godot gives the extension a raw `Object*` slot inside a pre-initialized Variant of the declared return type (`../godot/core/extension/gdextension.cpp:132-142`). For OBJECT returns this is exactly the wrapped object pointer, and Godot then re-wraps it. So encoding the wrapped object as an `Object` is correct — no `ref_set_object` needed.

**ptrcall path** (`pkg/core/variant_refect_value.go`, `GDExtensionTypePtrFromReflectValue`):

```go
case reflect.Interface:
    switch inst := value.Interface().(type) {
    case Object:
        ObjectEncoder.EncodeTypePtrArg(inst, rOut)
    case Ref:  // NEW
        obj := inst.ToObject()
        if obj == nil {
            ObjectEncoder.EncodeTypePtrArg(nil, rOut)  // or zero-fill rOut
            return
        }
        ObjectEncoder.EncodeTypePtrArg(obj.(Object), rOut)
    default:
        log.Panic(...)
    }
```

**varcall path** (`GDExtensionVariantPtrFromReflectValue`, line ~263): same detection, but encode via `ObjectEncoder.EncodeVariantPtrArg` so the Variant is of type `OBJECT` holding the object.

**Reference balance on return**: the returned Go `Ref` becomes unreachable after encoding and is released by its finalizer (or explicitly by the caller). Godot's Variant/`Ref` wrapping takes its own reference, so the object survives correctly.

**Rationale**: verified against Godot's extension method-bind return handling; reuses the existing `ObjectEncoder`, which already knows how to extract the `GodotObject` pointer from an `Object`.

### 5. Ref decoding (Godot→Go) — keep, verify transfer semantics

The Godot→Go direction is already correct and must be preserved:
- ptrcall arg decode: `ref_get_object` → `GDClassRefConstructors` transfer constructor (`method_bind_reflect.go:509-554`).
- Variant arg decode: `arg.ToObject()` → transfer constructor (`method_bind_reflect.go:164-184`).
- Variant return decode: same via `:293-305` / `:317-329`.

No `Reference()` is called during decode, so no leak. Go code that retains a decoded ref must copy (`Ref(from)`), which adds a reference that the finalizer will release.

**No changes required** other than confirming the renamed constructors still wire through `GDClassRefConstructors` and the `Ref` interface split (Decision 1) still satisfies `t.Implements(refType)`.

### 6. Rejected: wiring reference/unreference and binding reference callbacks

The original design proposed wiring these to call `Reference()`/`Unreference()`. This is **rejected**:

1. **Double-counting / recursion**: the class-creation callbacks are invoked from *inside* `RefCounted::reference()/unreference()` (`ref_counted.cpp:82-84, 102-104`). Calling `rc.Reference()` there re-enters the same function after the internal count already changed, permanently inflating the refcount by one per copy/destroy.
2. **Wrong parameter**: `GDExtensionInstanceBindingReferenceCallback` receives `(p_token, p_binding, p_reference)` (`gdextension_interface.h:197`). `p_binding` is the binding data, which is `nullptr` for godot-go (create returns `nullptr`); `cgo.Handle(p_binding)` would panic on a zero handle. The callbacks never receive the instance.
3. **Already correct**: the binding callbacks' return value is "can the object die"; returning `1`/`true` is the right answer. The Go wrapper is freed by `free_instance_func` when Godot's internal refcount reaches zero — the callbacks have nothing to do.
4. **godot-cpp precedent**: `reference_func`/`unreference_func` are `nullptr` (`class_db.hpp:280-281`) and the binding reference callback is unimplemented; user-defined `RefCounted` classes work correctly.

### 7. Generated Ref types — thin wrappers via embedding

Current generated code delegates ~15 methods per type:

```go
type RefImageImpl TypedRef[Image]
func (r *RefImageImpl) Ptr() RefCounted { ... }   // + 14 more delegations
```

After fix:

```go
type RefImage struct {
    *Ref[Image]
}

func (r *RefImage) TypedPtr() Image { return r.Ref.Ptr() }

func NewRefImage(ref Image) RefImage {
    return RefImage{Ref: NewRefInit[Image](ref)}
}

func NewRefImageAsRef(ref RefCounted) Ref {
    return NewRef[Image](ref.(Image))
}

func NewRefImageGDExtensionIternalConstructor(ref Image) RefImage {
    return RefImage{Ref: NewRef[Image](ref)}
}
```

`Ref`-interface methods (`ToObject()`, `Ref()`, `Unref()`, `IsValid()`) are inherited via embedding, so `var _ Ref = &RefImage{}` and `var _ RefImage = &RefImage{}` still hold. `Ptr()` is promoted from `*Ref[Image]` and returns the concrete `Image`, giving `ref.Ptr().GetWidth()`.

**Rationale**: embedding eliminates ~15 delegation methods per type (~3000 lines → ~600 lines across ~200 types); constructors keep godot-cpp semantics (init vs transfer).

## Architecture

```
┌───────────────────────────────────────────────────────────────────────────┐
│  Ref[T] Data Flow                                                          │
│                                                                            │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │  Go creates Ref<Image>:                                              │ │
│  │    img := NewImage()                                                 │ │
│  │    ref := NewRefInit[Image](img)  → InitRef() +1, finalizer set      │ │
│  │    ref.Ptr().GetWidth()           → direct concrete access           │ │
│  │    ref.Unref()  (or GC finalizer) → Unreference() -1                 │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │  Godot → Go (decode, transfer — no refcount change):                 │ │
│  │    Godot: myRef passed as arg / returned in Variant                  │ │
│  │      → ref_get_object / Variant.ToObject() → owner Object*          │ │
│  │      → GDClassRefConstructors → NewRefXAsRef() [transfer]            │ │
│  │      → Go Retains? call Ref(from) [Reference()] — finalizer releases │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
│                                                                            │
│  ┌───────────────────────────────────────────────────────────────────────┐ │
│  │  Go → Godot (encode, the fix):                                       │ │
│  │    func GetImage() RefImage { ... }                                  │ │
│  │      → GDExtensionTypePtrFromReflectValue / ...VariantPtr...         │ │
│  │      → detect Ref → ToObject() → ObjectEncoder (Object* / Variant)  │ │
│  │      → Go Ref released by finalizer after the call returns           │ │
│  └───────────────────────────────────────────────────────────────────────┘ │
└───────────────────────────────────────────────────────────────────────────┘
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Double `Unref()` — explicit then finalizer | `Unref()` zeroes `m_ref` and clears the finalizer — second call is a safe no-op |
| Finalizer timing is nondeterministic / runs on GC | Document explicit `Unref()` for hot paths; finalizer is a correctness safety net |
| Transfer wrapper (`NewRef`) outlives Godot's reference → dangling | Documented invariant; callers needing ownership must `Ref(from)` (copy) |
| Removing `Ptr()` from `Ref` interface | Only internal usages exist (verified); decode dispatch still works via embedding; no production API consumers |
| Calling into cgo from a finalizer | Allowed (finalizers run on a dedicated goroutine); keep finalizer bodies minimal |
| Template changes affect ~200 types | Mechanical regeneration; verify with `make generate` + test suite |

## Trade-off: Pin vs Finalizer for `Ref[T]`

Today `TypedRef` calls `pnr.Pin(ptr)` to keep the struct GC-stable. `Ref[T]` stores only a Go interface (`T`) — no C pointers that the GC would move — so pinning is unnecessary. We **remove `pnr.Pin`** and rely on `runtime.SetFinalizer` for release. Verify with `cgocheck=1` that no cgo pointer rules are violated.

## File Changes

| File | Change |
|------|--------|
| `pkg/builtin/ref_generic.go` | Rewrite: `TypedRef[T]` → `Ref[T]`, split `Ref` interface (drop `Ptr()`, add `ToObject()`), add `NewRef`/`NewRefInit`, finalizers, remove `pnr.Pin` |
| `pkg/builtin/classes.ref.interfaces.gen.go` | Regenerate per-class `RefX` interfaces (embed `Ref`, keep concrete `TypedPtr()`/`Ptr()`) |
| `pkg/core/variant_refect_value.go` | Add `Ref` detection to `GDExtensionTypePtrFromReflectValue` and `GDExtensionVariantPtrFromReflectValue` |
| `pkg/core/variant_reflect_type.go` | Ensure `Ref[T]` maps to `VARIANT_TYPE_OBJECT` |
| `pkg/core/method_bind_reflect.go` | Confirm decode paths use transfer constructors; adjust for renamed constructors if needed |
| `cmd/generate/gdclassimpl/templatefunctions.go` | Update Ref helpers / constructors |
| `cmd/generate/gdclassimpl/templates/*.tmpl` | Ref type template → embedding struct |
| `test/demo/` | Add Ref lifecycle, round-trip, finalizer, user-defined RefCounted tests |
