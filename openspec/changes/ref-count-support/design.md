## Context

godot-go's `TypedRef[T]` is incomplete — it cannot encode `Ref` types in method return values, Godot's class-level reference callbacks are `nil`, and the binding reference callbacks are no-ops. This makes user-defined `RefCounted` classes (Resources, custom ref-counted types) unusable with correct memory management.

Two callback paths exist for ref counting in GDExtension, with different responsibilities:

```
┌─────────────────────────────────────────────────────────────────────┐
│  Callback Path 1: Class Creation Info (referenceFunc/unreferenceFunc)│
│  └── Called by Godot's Ref<T> class when copying/destroying a Ref  │
│  └── PRIMARY refcounting mechanism for user-defined classes        │
│  └── Should call RefCounted.Reference()/Unreference()              │
│  └── Currently: nil (NOT WIRED) 💥                                 │
│                                                                     │
│  Callback Path 2: Instance Binding (GoCallback_GDClassBindingReference)│
│  └── Called by Godot's binding system for Go-side object tracking  │
│  └── Secondary — manages Go binding lifecycle                      │
│  └── Should track Go-side binding reference                        │
│  └── Currently: returns 1 (NO-OP) 💥                               │
└─────────────────────────────────────────────────────────────────────┘
```

## Goals / Non-Goals

**Goals:**
- `Ref[T]` works as a drop-in smart pointer for `RefCounted` objects.
- Godot-side `Ref<T>` copy/destroy correctly calls `Reference()`/`Unreference()` on user-defined classes.
- Go methods returning `Ref[T]` encode correctly to Godot (ptrcall and varcall).
- Go methods accepting `Ref[T]` arguments decode correctly from Godot.
- API is ergonomic: `ref.Ptr().Method()` not `ref.TypedPtr().Method()`.
- Support user-defined GDExtension classes that extend `RefCounted` (Resources, etc.).

**Non-Goals:**
- Changing the general-purpose `Object` lifecycle or binding system.
- Implementing custom allocators or ref-counted memory pools.
- Modifying generated code in `*.gen.*` files directly — only template changes.

## Decisions

### 1. Replace `TypedRef[T]` with `Ref[T]`

Current `TypedRef[T]` stores `RefCounted` (the base interface), requiring type assertions:

```go
// Current — awkward, requires type assertion
type TypedRef[T TypedRefT] struct {
    Reference RefCounted  // stores as base interface
}
func (cx *TypedRef[T]) TypedPtr() T { return cx.Reference.(T) }  // type assertion
```

New `Ref[T]` stores `T` directly:

```go
// New — type-safe, no assertions
type Ref[T RefCounted] struct {
    m_ref T  // stored as concrete type
}
func (r *Ref[T]) Ptr() T { return r.m_ref }
```

**Rationale**: Eliminates type assertions, stores the exact type, matches godot-cpp's `Ref<T>` pattern. `Ptr()` is the godot-cpp equivalent of `operator->`.

### 2. Two Constructors — Godot-Owned vs User-Owned

```go
// Godot-owned object (returned from Godot API, load(), etc.)
// Does NOT call InitRef — the object is already ref-counted by Godot.
func NewRef[T RefCounted](obj T) *Ref[T] {
    r := &Ref[T]{m_ref: obj}
    obj.Reference()  // claim our Go-side reference
    pnr.Pin(r)
    return r
}

// User-owned object (newly created, e.g., NewImage())
// Calls InitRef() to mark as ref-counted, then Reference().
func NewRefInit[T RefCounted](obj T) *Ref[T] {
    if !obj.InitRef() {
        return nil  // failed — object doesn't support ref counting
    }
    r := &Ref[T]{m_ref: obj}
    obj.Reference()  // claim our Go-side reference
    pnr.Pin(r)
    return r
}
```

**Rationale**: `InitRef()` marks an object as ref-counted (can only be called once). Godot-owned objects are already initialized. User-owned objects need initialization. This mirrors godot-cpp's `_gde_internal_constructor` for Godot-owned objects.

### 3. Ref Copy and Unref Semantics

```go
// Copy reference from another Ref, managing refcounts.
// After this, both refs share the same underlying object.
func (r *Ref[T]) Ref(from *Ref[T]) {
    if from == nil {
        r.Unref()
        return
    }
    var zero T
    if from.m_ref == nil {
        r.Unref()
        return
    }
    if r.m_ref == from.m_ref {
        return  // self-assignment, no-op
    }
    r.Unref()
    r.m_ref = from.m_ref
    r.m_ref.Reference()
}

// Release this reference. If refcount drops to zero, the
// underlying Godot object is freed. Safe to call multiple times.
func (r *Ref[T]) Unref() {
    var zero T
    if r.m_ref != nil && r.m_ref != zero {
        r.m_ref.Unreference()
    }
    r.m_ref = zero  // prevents double-unref
}
```

**Rationale**: Self-assignment guard prevents accidental double-ref. `Unref()` zeroes `m_ref` to make double-unref a safe no-op.

### 4. Ref Encoding — Extract `Ptr()` as `Object*`

For ptrcall return values, Godot expects raw `Object*`. `Ref[T]` encoding extracts the underlying object:

```
GDExtensionTypePtrFromReflectValue(RefImage, rReturn)
  → detect Ref type → extract ref.Ptr() → encode as Object*
```

**Why not `RefSetObject`?** `RefSetObject` is for Variant-level `Ref` manipulation. For method binding, Godot's ptrcall expects `Object*` and handles its own `Ref` wrapping internally. Extracting `Ptr()` is simpler and correct.

**Encoding path** (in `variant_refect_value.go`):
```go
// In GDExtensionTypePtrFromReflectValue
case reflect.Interface:
    switch inst := value.Interface().(type) {
    case Object:
        ObjectEncoder.EncodeTypePtrArg(inst, rOut)
    case Ref:  // NEW
        ref := inst  // extract the Ref interface
        // Get the underlying Object and encode
        obj := ref.Ptr()  // returns RefCounted → cast to Object
        ObjectEncoder.EncodeTypePtrArg(obj.(Object), rOut)
```

### 5. Wire `referenceFunc`/`unreferenceFunc` in Class Creation

**This is the critical missing link.** Currently both are `nil`:

```go
// Current — BROKEN for user-defined RefCounted classes
info := NewGDExtensionClassCreationInfo4(
    ...
    (GDExtensionClassReference)(nil),      // never called
    (GDExtensionClassUnreference)(nil),    // never called
    ...
)
```

**After fix:**

```go
info := NewGDExtensionClassCreationInfo4(
    ...
    (GDExtensionClassReference)(C.cgo_classcreationinfo_reference),
    (GDExtensionClassUnreference)(C.cgo_classcreationinfo_unreference),
    ...
)
```

**Callback implementation** (in `classdb_callback.go`):
```go
//export GoCallback_ClassCreationInfoReference
func GoCallback_ClassCreationInfoReference(p_instance C.GDExtensionClassInstancePtr) {
    wci := cgo.Handle(p_instance).Value().(*WrappedClassInstance)
    if rc, ok := wci.Instance.(RefCounted); ok {
        rc.Reference()
    }
    // Non-RefCounted classes (Node, etc.) hit the !ok branch — safe no-op
}

//export GoCallback_ClassCreationInfoUnreference
func GoCallback_ClassCreationInfoUnreference(p_instance C.GDExtensionClassInstancePtr) {
    wci := cgo.Handle(p_instance).Value().(*WrappedClassInstance)
    if rc, ok := wci.Instance.(RefCounted); ok {
        rc.Unreference()
    }
}
```

**Rationale**: Godot calls these when its own `Ref<T>` copies/destroys a reference. Without them, user-defined `RefCounted` classes cannot be used in Godot-side `Ref` — refcount drift causes use-after-free or leaks. Non-RefCounted classes (Node, etc.) hit the type assertion failure and no-op safely.

### 6. Fix Binding Reference Callbacks

Both current callbacks are no-ops. Wire them to update Go-side binding references:

```go
// For GDExtension classes (class-level binding)
func GoCallback_GDExtensionBindingReference(p_type_name, p_token, p_instance, p_reference bool) bool {
    wci := cgo.Handle(p_instance).Value().(*WrappedClassInstance)
    if rc, ok := wci.Instance.(RefCounted); ok {
        if p_reference {
            rc.Reference()
        } else {
            rc.Unreference()
        }
    }
    return true
}

// For GDClass (instance-level binding)
func GoCallback_GDClassBindingReference(p_token, p_instance, p_reference) C.GDExtensionBool {
    wci := cgo.Handle(p_instance).Value().(*WrappedClassInstance)
    if rc, ok := wci.Instance.(RefCounted); ok {
        if p_reference {
            rc.Reference()
        } else {
            rc.Unreference()
        }
    }
    return 1
}
```

**Rationale**: These complement the class-level `referenceFunc`/`unreferenceFunc`. The binding callbacks manage the Go-side binding wrapper lifecycle. Both layers need correct refcounting.

### 7. Generated Ref Types — Thin Wrappers via Embedding

Current generated code uses type alias and delegation (~15 methods per type, ~200 types):

```go
// Current — ~3000 lines of delegation
type RefImageImpl TypedRef[Image]
func (r *RefImageImpl) Ptr() RefCounted { rg := (*TypedRef[Image])(r); return rg.Ptr().(RefCounted) }
func (r *RefImageImpl) TypedPtr() Image { rg := (*TypedRef[Image])(r); return rg.TypedPtr() }
func (r *RefImageImpl) Ref(from Ref) { rg := (*TypedRef[Image])(r); rg.Ref(from) }
func (r *RefImageImpl) TypedRef(from *RefImageImpl) { /* ... */ }
func (r *RefImageImpl) Unref() { rg := (*TypedRef[Image])(r); rg.Unref() }
func (r *RefImageImpl) IsValid() bool { return r != nil && r.Reference != nil }
```

**After fix — embedding, no delegation:**

```go
// New — ~600 lines total
type RefImage struct {
    *Ref[Image]
}

func (r *RefImage) Ptr() Image { return r.Ref.Ptr() }

func NewRefImage(ref Image) RefImage {
    return RefImage{Ref: NewRefInit[Image](ref)}
}

func NewRefImageAsRef(ref RefCounted) Ref {
    return NewRef[Image](ref.(Image))
}
```

**Rationale**: Embedding eliminates ~15 delegation methods per type. `Ref[T]` methods (`Ref()`, `Unref()`, `IsValid()`) are inherited automatically via embedding.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│  Ref[T] Data Flow                                                    │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Go creates Ref<Image>:                                         │ │
│  │    img := NewImage()                                            │ │
│  │    ref := NewRefInit[Image](img)  → InitRef() + Reference()     │ │
│  │    ref.Ptr().GetWidth()            → direct method access        │ │
│  │    ref.Unref()                     → Unreference()               │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Godot creates Ref<MyResource> (user-defined):                  │ │
│  │    Godot: myRef = Ref<MyResource>(obj)                          │ │
│  │      → referenceFunc callback → wci.Instance.Reference()        │ │
│  │    Godot: myRef2 = myRef                                        │ │
│  │      → referenceFunc callback → wci.Instance.Reference()        │ │
│  │    Godot: myRef = null                                          │ │
│  │      → unreferenceFunc callback → wci.Instance.Unreference()    │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Go returns Ref<Image> to Godot:                                │ │
│  │    func GetImage() RefImage { ... }                             │ │
│  │      → GDExtensionTypePtrFromReflectValue(ref, rOut)            │ │
│  │      → detect Ref → ref.Ptr() → encode as Object*               │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Godot returns Ref<Image> to Go:                                │ │
│  │    Godot: variant = sprite.GetTexture()                         │ │
│  │      → Variant.ToObject() → GodotObject                         │ │
│  │      → GDClassRefConstructors → NewRefImageAsRef()              │ │
│  │      → NewRef[Image](obj) → Reference()                         │ │
│  └─────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Double `Unref()` — user calls `Unref()`, then GC finalizer also calls it | `Unref()` zeroes `m_ref` — second call is safe no-op |
| `Pin(r)` holds GC hostage for `Ref[T]` lifetime | `Ref[T]` stores Go interface (GC-stable). Verify with cgocheck=1. Remove pin if unnecessary. |
| Godot frees object while Go `Ref[T]` still holds interface | `NewRef()` calls `Reference()` — Go holds a refcount. Shared-ownership semantics. |
| Non-RefCounted class gets reference callbacks | Type assertion `wci.Instance.(RefCounted)` — fails for Node/etc., safe no-op |
| Template changes affect ~200 types | Mechanical regeneration. Verify with `make generate` + test suite. |

## Trade-off: Pin vs No-Pin for `Ref[T]`

`Ref[T]` stores a Go interface (`T`), which is already GC-stable (the interface header is on the stack, the data pointer is heap-allocated and stable). The `pnr.Pin(r)` may be unnecessary — the struct itself doesn't contain C pointers that would be moved by GC.

**Investigation needed**: Remove `pnr.Pin(r)` and verify with `cgocheck=1` that no cgo pointer issues arise. If safe, remove pin to avoid holding GC hostage indefinitely.

## File Changes

| File | Change |
|------|--------|
| `pkg/builtin/ref_generic.go` | Rewrite: `TypedRef[T]` → `Ref[T]`, add `NewRef`/`NewRefInit` constructors |
| `pkg/builtin/wrapped_gdclass.go` | Fix `GoCallback_GDClassBindingReference` to call `Reference()`/`Unreference()` |
| `pkg/gdclassinit/wrapped_gdextension_class.go` | Fix `GoCallback_GDExtensionBindingReference` |
| `pkg/core/classdb.go` | Wire `referenceFunc`/`unreferenceFunc` in `NewGDExtensionClassCreationInfo4` |
| `pkg/core/classdb_callback.go` | Add `GoCallback_ClassCreationInfoReference`/`Unreference` C exports |
| `pkg/core/method_bind_reflect.go` | Add Ref detection in return encoding (`GDExtensionTypePtrFromReflectValue`) |
| `pkg/core/variant_reflect_type.go` | Ensure `Ref[T]` maps to `VARIANT_TYPE_OBJECT` |
| `cmd/generate/gdclassimpl/templatefunctions.go` | Update `goEncodeIsReference` and related helpers |
| `cmd/generate/gdclassimpl/templates/` | Update Ref type template for embedding struct |
| `test/demo/` | Add Ref lifecycle test cases |
