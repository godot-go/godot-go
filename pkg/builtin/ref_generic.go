package builtin

type Ref interface {
	Ptr() RefCounted
	Unref()
	Ref(pFrom Ref)
	IsValid() bool
}

type TypedRefT interface {
	comparable
	RefCounted
}

// Ref is a helper struct for RefCounted Godot Objects.
type TypedRef[T TypedRefT] struct {
	// HasReference
	reference RefCounted
}

func (cx *TypedRef[T]) Ptr() RefCounted {
	return (RefCounted)(cx.reference)
}

func (cx *TypedRef[T]) TypedPtr() T {
	return cx.reference.(T)
}

func (cx *TypedRef[T]) Ref(pFrom Ref) {
	cx.TypedRef(pFrom.(*TypedRef[T]))
}

// Ref increments a reference counter
func (cx *TypedRef[T]) TypedRef(from *TypedRef[T]) {
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

func (cx *TypedRef[T]) RefPointer(r T) {
	var zero T
	if r == zero {
		panic("reference cannot be nil")
	}
	if !r.InitRef() {
		panic("init ref failure")
	}
	cx.reference = r
}

func (cx *TypedRef[T]) Unref() {
	var zero T
	if cx.reference != zero && cx.reference.Unreference() {
		cx.reference.Destroy()
		// release memory
		// runtime.Unpin(cx.reference)
	}
	cx.reference = zero
}

func (cx *TypedRef[T]) IsValid() bool {
	return cx != nil && cx.reference != nil
}

func NewTypedRef[T TypedRefT](reference T) *TypedRef[T] {
	ref := TypedRef[T]{}
	ref.RefPointer(reference)
	return &ref
}

func newTypedRefGDExtensionIternalConstructor[T TypedRefT](reference T) *TypedRef[T] {
	ref := TypedRef[T]{}
	ref.reference = reference
	return &ref
}
