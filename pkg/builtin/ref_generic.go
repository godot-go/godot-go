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

// NewRefTransfer takes ownership of a reference the engine has already
// counted and explicitly handed over. The refcount is not changed -- the +1
// already exists -- but a finalizer is installed so that transferred
// reference is released exactly once when the wrapper is dropped, mirroring
// godot-cpp's Ref destructor.
//
// Use this only where the engine genuinely transferred a reference into the
// value being wrapped, such as a ptrcall return slot filled through
// PtrToArg<Ref<T>>::convert, which constructs a Ref and so calls
// reference(). Wrapping an object that someone else holds with this
// constructor releases a reference that was never acquired.
//
// A wrapper over no engine object gets no finalizer: Unref would issue an
// unreference ptrcall against a null owner.
func NewRefTransfer[T RefCountedT](obj T) *RefBase[T] {
	r := &RefBase[T]{m_ref: obj}
	if ObjectArgPtr(obj) != nil {
		runtime.SetFinalizer(r, (*RefBase[T]).Unref)
	}
	return r
}

// NewRef wraps an object without changing its reference count and without
// taking any ownership: the caller or some other party holds the reference
// that keeps the object alive, and dropping this wrapper releases nothing.
//
// This is the borrowing constructor. A wrapper built here that outlives the
// real owner's reference observes a freed object, and it never releases the
// engine reference even when the engine handed one over. Where the engine
// transferred a +1 into the value being wrapped -- a ptrcall return slot,
// for instance -- use NewRefTransfer instead so that reference is released.
func NewRef[T RefCountedT](obj T) *RefBase[T] {
	return &RefBase[T]{m_ref: obj}
}
