package builtin

import "runtime"

// Ref is the interface for reference-counted Godot objects.
//
// Go-side refcounting is managed by runtime.SetFinalizer — the user never
// calls Unref() directly. Each *RefCountedRef[T] allocation has exactly one
// finalizer that calls Unreference() when the Ref is garbage collected.
//
// Pointer copy of *RefCountedRef[T] is shared ownership (no count bump).
// Use Clone() to create an independent claim on the same underlying object.
type Ref interface {
	// Ptr returns the underlying RefCounted object.
	Ptr() RefCounted
	// Unref releases this reference. Managed automatically via finalizer.
	Unref()
	// Ref replaces the managed reference with one from another Ref.
	Ref(pFrom Ref)
	// IsValid returns whether this Ref holds a valid object.
	IsValid() bool
	// Clone creates an independent claim on the same underlying object.
	Clone() Ref
}

// TypedRefT is the constraint for types that can be stored in Ref.
type TypedRefT interface {
	comparable
	RefCounted
}

// RefCountedRef is the generic reference-counted smart pointer for Godot
// RefCounted objects. Go-side refcounting is fully automatic via
// runtime.SetFinalizer — the user never calls Unref() directly.
//
// Sharing model:
//   - Pointer copy (ref2 = ref1) — shared ownership, same finalizer, no count bump
//   - Clone() — new ownership, new finalizer, Reference() called
//
// This type satisfies the Ref interface.
type RefCountedRef[T TypedRefT] struct {
	// Reference is the underlying Godot RefCounted object, stored as the base
	// interface to allow nil-safe comparison.
	Reference RefCounted
}

// Ptr returns the underlying RefCounted object for method access.
func (r *RefCountedRef[T]) Ptr() RefCounted {
	return r.Reference
}

// TypedPtr returns the concrete type T.
func (r *RefCountedRef[T]) TypedPtr() T {
	return r.Reference.(T)
}

// Clone creates an independent claim on the same underlying object.
// Each clone has its own finalizer — both clones must go out of scope
// before the object is freed.
func (r *RefCountedRef[T]) Clone() Ref {
	return NewRef(r.Reference.(T))
}

// Unref releases this reference. If refcount drops to zero, the underlying
// Godot object is freed. Safe to call multiple times.
//
// NOTE: This is managed automatically via finalizer. Call only when
// deterministic cleanup is needed.
func (r *RefCountedRef[T]) Unref() {
	var zero T
	if r.Reference != zero && r.Reference.Unreference() {
		// object may have been freed
	}
	r.Reference = zero
}

// Ref replaces the managed reference with one from another Ref.
// The old reference is released, the new one is claimed.
func (r *RefCountedRef[T]) Ref(pFrom Ref) {
	r.Unref()
	if pFrom == nil {
		return
	}
	if src, ok := pFrom.(*RefCountedRef[T]); ok {
		r.Reference = src.Reference
		if r.Reference != nil {
			r.Reference.Reference()
		}
	}
}

// TypedRef replaces the managed reference with one from another RefCountedRef
// of the same type. The old reference is released, the new one is claimed.
// Kept for backward compatibility with generated code.
//
// Deprecated: use Ref() instead.
func (r *RefCountedRef[T]) TypedRef(from *RefCountedRef[T]) {
	r.Ref(from)
}

// IsValid returns whether this Ref holds a valid object.
func (r *RefCountedRef[T]) IsValid() bool {
	return r != nil && r.Reference != nil
}

// NewRef creates a RefCountedRef wrapping a Godot-owned object.
// The object is already ref-counted by Godot (returned from Godot API, load(), etc.).
// Does NOT call InitRef.
//
// A finalizer is registered that calls Unreference() when this Ref is garbage collected.
func NewRef[T TypedRefT](obj T) *RefCountedRef[T] {
	r := &RefCountedRef[T]{Reference: obj}
	obj.Reference() // claim our Go-side reference
	runtime.SetFinalizer(r, func(r *RefCountedRef[T]) {
		var zero T
		if r.Reference != zero {
			r.Reference.Unreference()
			r.Reference = zero
		}
	})
	return r
}

// NewRefInit creates a RefCountedRef wrapping a user-owned object.
// Calls InitRef() to mark the object as ref-counted, then Reference().
// Returns nil if InitRef fails (object doesn't support ref counting).
//
// A finalizer is registered that calls Unreference() when this Ref is garbage collected.
func NewRefInit[T TypedRefT](obj T) *RefCountedRef[T] {
	if !obj.InitRef() {
		return nil // failed — object doesn't support ref counting
	}
	return NewRef(obj) // calls Reference() + SetFinalizer
}

// NewTypedRef creates a RefCountedRef wrapping a user-owned object.
// Kept for backward compatibility with generated code during transition.
// Equivalent to NewRefInit.
func NewTypedRef[T TypedRefT](reference T) *RefCountedRef[T] {
	return NewRefInit(reference)
}

// TypedRef is an alias for RefCountedRef for backward compatibility with
// generated code. New code should use RefCountedRef.
//
// Deprecated: use RefCountedRef instead.
type TypedRef[T TypedRefT] = RefCountedRef[T]

// NewTypedRefGDExtensionIternalConstructor creates a RefCountedRef wrapping a Godot-owned
// object. Kept for backward compatibility with generated code during transition.
// Equivalent to NewRef.
func NewTypedRefGDExtensionIternalConstructor[T TypedRefT](reference T) *RefCountedRef[T] {
	return NewRef(reference)
}
