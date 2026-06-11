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
- Go GC manages ref lifecycle — no manual `Unref()` required.
- Support user-defined GDExtension classes that extend `RefCounted` (Resources, etc.).

**Non-Goals:**
- Changing the general-purpose `Object` lifecycle or binding system.
- Implementing custom allocators or ref-counted memory pools.
- Modifying generated code in `*.gen.*` files directly — only template changes.
- Providing deterministic cleanup (finalizer fires "eventually").

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

**Rationale**: Eliminates type assertions, stores the exact type. `Ptr()` returns the concrete type directly.

### 2. Two Constructors — Godot-Owned vs User-Owned

```go
// Godot-owned object (returned from Godot API, load(), etc.)
// Does NOT call InitRef — the object is already ref-counted by Godot.
func NewRef[T RefCounted](obj T) *Ref[T] {
    r := &Ref[T]{m_ref: obj}
    obj.Reference()  // claim our Go-side reference
    runtime.SetFinalizer(r, func(r *Ref[T]) {
        r.m_ref.Unreference()
    })
    return r
}

// User-owned object (newly created, e.g., NewImage())
// Calls InitRef() to mark as ref-counted, then delegates to NewRef.
func NewRefInit[T RefCounted](obj T) *Ref[T] {
    if !obj.InitRef() {
        return nil  // failed — object doesn't support ref counting
    }
    return NewRef[T](obj)  // calls Reference() + SetFinalizer
}
```

**Rationale**: `InitRef()` marks an object as ref-counted (can only be called once). Godot-owned objects are already initialized. User-owned objects need initialization. `NewRefInit` delegates to `NewRef` after init — no duplication of finalizer setup. No `Pin(r)` — `Ref[T]` stores Go interface (GC-stable).

### 3. Pure Finalizer — No Manual Unref

**Decision**: Go's GC manages the entire refcounting lifecycle via `runtime.SetFinalizer`. The user never calls `Unref()`.

```go
// Clone creates an independent claim on the same underlying object.
// Each clone has its own finalizer — both must go out of scope
// before the object is freed.
func (r *Ref[T]) Clone() *Ref[T] {
    return NewRef(r.m_ref)  // Reference() + new finalizer
}

// Ptr returns the underlying RefCounted object for method access.
func (r *Ref[T]) Ptr() T {
    return r.m_ref
}
```

**Rationale**: Go is a garbage-collected language. Manual `Unref()` is error-prone (forget to call it = leak, call it too early = use-after-free). `runtime.SetFinalizer` guarantees exactly-one `Unreference()` per `*Ref[T]` allocation.

**Sharing model**:
```go
ref1 := NewRef[Image](img)     // Reference() → Godot refcount: 1, finalizer A
ref2 := ref1.Clone()           // Reference() → Godot refcount: 2, finalizer B

ref1 = nil  // finalizer A fires eventually → refcount: 1
ref2 = nil  // finalizer B fires eventually → refcount: 0, Godot frees
```

**Pointer copy is shared ownership** (no count bump):
```go
ref2 := ref1  // same *Ref[T], same finalizer — NOT a new claim
```

**`Clone()` is new ownership** (count bump):
```go
ref2 := ref1.Clone()  // new *Ref[T], new finalizer, Reference() called
```

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

### 7. Generated Ref Types — Minimal Wrappers via Embedding

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
func (r *RefImage) Clone() *Ref[Image] { return r.Ref.Clone() }

func NewRefImage(img Image) *RefImage {
    return &RefImage{Ref: NewRefInit[Image](img)}
}

func NewRefImageAsRef(rc RefCounted) *Ref[Image] {
    return NewRef[Image](rc.(Image))
}
```

**Rationale**: With pure finalizer semantics, there's no `Unref()`, `Ref()`, `TypedRef()` to delegate. Each generated type is just `Ptr()`, `Clone()`, and constructors. `Ref[T]` methods (`Clone()`, `Ptr()`) are inherited automatically via embedding.

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│  Ref[T] Data Flow (Finalizer-Based)                                  │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Go creates Ref<Image>:                                         │ │
│  │    img := NewImage()                                            │ │
│  │    ref := NewRefInit[Image](img)  → InitRef() + Reference()     │ │
│  │                         → runtime.SetFinalizer(r, Unreference)  │ │
│  │    ref.Ptr().GetWidth()            → direct method access        │ │
│  │    (GC collects ref eventually → Unreference() automatically)   │ │
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
│  │      → NewRef[Image](obj) → Reference() + SetFinalizer          │ │
│  └─────────────────────────────────────────────────────────────────┘ │
│                                                                      │
│  ┌─────────────────────────────────────────────────────────────────┐ │
│  │  Go shares Ref<Image> with another variable:                    │ │
│  │    ref2 := ref1.Clone()  → Reference() + new finalizer          │ │
│  │    ref1 = nil   → finalizer A fires eventually → Unreference()  │ │
│  │    ref2 = nil   → finalizer B fires eventually → Unreference()  │ │
│  └─────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────┘
```

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| Pointer copy without `Clone()` — user copies `*Ref[T]`, both variables share one finalizer | **By design** — pointer copy is shared ownership. `Clone()` creates new claim. Document clearly. |
| Finalizer timing — object freed "eventually" after last ref dies | Godot refcount is shared between Go and Godot sides. As long as Godot-side `Ref<T>` still holds a ref, the object stays alive regardless of Go finalizer timing. |
| Global map holds `*Ref[T]` forever — finalizer never fires | **Expected** — cache entry keeps object alive. User deletes from map to release. Same as any manual ref system. |
| No `Pin(r)` — is `Ref[T]` GC-stable? | `Ref[T]` stores Go interface `T` (GC-stable pointer + type). Verify with cgocheck=1. No C pointers in struct. |
| Godot frees object while Go `Ref[T]` still holds interface | `NewRef()` calls `Reference()` — Go holds a refcount. Shared-ownership semantics. Godot won't free while Go holds ref. |
| Non-RefCounted class gets reference callbacks | Type assertion `wci.Instance.(RefCounted)` — fails for Node/etc., safe no-op |
| Template changes affect ~200 types | Mechanical regeneration. Verify with `make generate` + test suite. |

## File Changes

| File | Change |
|------|--------|
| `pkg/builtin/ref_generic.go` | Rewrite: `TypedRef[T]` → `Ref[T]` with `runtime.SetFinalizer`, no `Unref()` |
| `pkg/builtin/wrapped_gdclass.go` | Fix `GoCallback_GDClassBindingReference` to call `Reference()`/`Unreference()` |
| `pkg/gdclassinit/wrapped_gdextension_class.go` | Fix `GoCallback_GDExtensionBindingReference` |
| `pkg/core/classdb.go` | Wire `referenceFunc`/`unreferenceFunc` in `NewGDExtensionClassCreationInfo4` |
| `pkg/core/classdb_callback.go` | Add `GoCallback_ClassCreationInfoReference`/`Unreference` C exports |
| `pkg/core/method_bind_reflect.go` | Add Ref detection in return encoding (`GDExtensionTypePtrFromReflectValue`) |
| `pkg/core/variant_reflect_type.go` | Ensure `Ref[T]` maps to `VARIANT_TYPE_OBJECT` |
| `cmd/generate/gdclassimpl/templatefunctions.go` | Update `goEncodeIsReference` and related helpers |
| `cmd/generate/gdclassimpl/templates/` | Update Ref type template for embedding struct (no Unref/Ref delegation) |
| `test/demo/` | Add Ref lifecycle test cases |
