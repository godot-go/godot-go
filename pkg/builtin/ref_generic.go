package builtin

import "runtime"

// Ref is the refcount protocol implemented by all reference wrappers.
// Note: the concrete generic implementation is RefBase[T]; the interface
// cannot share the name because Go disallows an interface and a generic
// type with the same identifier.
type Ref interface {
	ToObject() RefCounted
	Ref(pFrom Ref)
	Unref()
	IsValid() bool
}

type RefCountedT interface {
	comparable
	RefCounted
}

// RefBase is a reference-counted smart pointer for Godot RefCounted
// objects. It stores the concrete type T directly and manages the
// underlying Godot reference count. Go-held references are released when
// the RefBase becomes unreachable via a finalizer, emulating godot-cpp's
// Ref destructor.
type RefBase[T RefCountedT] struct {
	m_ref T
}

func (r *RefBase[T]) Ptr() T {
	return r.m_ref
}

func (r *RefBase[T]) ToObject() RefCounted {
	return r.m_ref
}

func (r *RefBase[T]) IsValid() bool {
	var zero T
	return r != nil && r.m_ref != zero
}

// Ref copies a reference from another Ref, managing refcounts.
// After this, both refs share the same underlying object.
func (r *RefBase[T]) Ref(from Ref) {
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
		return // self-assignment, no-op
	}
	r.Unref()
	r.m_ref = o
	r.m_ref.Reference()
	runtime.SetFinalizer(r, (*RefBase[T]).Unref)
}

// Unref releases this reference. If the refcount drops to zero, the
// underlying Godot object is freed. Safe to call multiple times.
func (r *RefBase[T]) Unref() {
	runtime.SetFinalizer(r, nil)
	var zero T
	if r.m_ref != zero {
		r.m_ref.Unreference()
	}
	r.m_ref = zero
}

// NewRefInit wraps a user-owned object. InitRef() marks the object as
// ref-counted (+1) and the Ref owns that reference. Returns nil if the
// object does not support ref counting.
func NewRefInit[T RefCountedT](obj T) *RefBase[T] {
	if !obj.InitRef() {
		return nil
	}
	r := &RefBase[T]{m_ref: obj}
	runtime.SetFinalizer(r, (*RefBase[T]).Unref)
	return r
}

// NewRef wraps a Godot-owned object with transfer semantics: the reference
// count is not changed because the other side already holds a reference.
// Mirrors godot-cpp's _gde_internal_constructor. The caller must ensure a
// reference is held for the wrapper's lifetime; to take ownership, copy.
func NewRef[T RefCountedT](obj T) *RefBase[T] {
	return &RefBase[T]{m_ref: obj}
}
