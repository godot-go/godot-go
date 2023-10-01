package gdextension

type Ref interface {
	Ptr() RefCounted
	Unref()
	Ref(pFrom Ref)
	IsValid() bool
}

type typedRefT interface {
	comparable
	RefCounted
}

// Ref is a helper struct for RefCounted Godot Objects.
type typedRef[T typedRefT] struct {
	// HasReference
	reference RefCounted
}

func (cx *typedRef[T]) Ptr() RefCounted {
	return (RefCounted)(cx.reference)
}

func (cx *typedRef[T]) TypedPtr() T {
	return cx.reference.(T)
}

func (cx *typedRef[T]) Ref(pFrom Ref) {
	cx.TypedRef(pFrom.(*typedRef[T]))
}

// Ref increments a reference counter
func (cx *typedRef[T]) TypedRef(from *typedRef[T]) {
	var zero T
	if from.reference == cx.reference {
		return
	}
	cx.Unref()
	cx.reference = from.reference
	if cx.reference != zero {
		(RefCounted)(cx.reference).Reference()
	}
}

func (cx *typedRef[T]) RefPointer(r T) {
	var zero T
	if r == zero {
		panic("reference cannot be nil")
	}
	if !r.InitRef() {
		panic("init ref failure")
	}
	cx.reference = r
}

func (cx *typedRef[T]) Unref() {
	var zero T
	if cx.reference != zero && cx.reference.Unreference() {
		cx.reference.Destroy()
		// release memory
		// runtime.Unpin(cx.reference)
	}
	cx.reference = zero
}

func (cx *typedRef[T]) IsValid() bool {
	return cx != nil && cx.reference != nil
}

func NewTypedRef[T typedRefT](reference T) *typedRef[T] {
	ref := typedRef[T]{}
	ref.RefPointer(reference)
	return &ref
}

func newTypedRefGDExtensionIternalConstructor[T typedRefT](reference T) *typedRef[T] {
	ref := typedRef[T]{}
	ref.reference = reference
	return &ref
}
