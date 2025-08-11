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
	Reference RefCounted
}

func (cx *TypedRef[T]) Ptr() RefCounted {
	return (RefCounted)(cx.Reference)
}

func (cx *TypedRef[T]) TypedPtr() T {
	return cx.Reference.(T)
}

func (cx *TypedRef[T]) Ref(pFrom Ref) {
	cx.TypedRef(pFrom.(*TypedRef[T]))
}

// Ref increments a reference counter
func (cx *TypedRef[T]) TypedRef(from *TypedRef[T]) {
	var zero T
	if from.Reference == cx.Reference {
		return
	}
	cx.Unref()
	cx.Reference = from.Reference
	if cx.Reference != zero {
		(RefCounted)(cx.Reference).Reference()
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
	cx.Reference = r
}

func (cx *TypedRef[T]) Unref() {
	var zero T
	if cx.Reference != zero && cx.Reference.Unreference() {
		// cx.Reference.Destroy()
		// release memory
		// runtime.Unpin(cx.reference)
	}
	cx.Reference = zero
}

func (cx *TypedRef[T]) IsValid() bool {
	return cx != nil && cx.Reference != nil
}

func NewTypedRef[T TypedRefT](reference T) *TypedRef[T] {
	ref := TypedRef[T]{}
	ref.RefPointer(reference)
	ptr := &ref
	pnr.Pin(ptr)
	return ptr
}

func NewTypedRefGDExtensionIternalConstructor[T TypedRefT](reference T) *TypedRef[T] {
	ref := TypedRef[T]{}
	ref.Reference = reference
	ptr := &ref
	pnr.Pin(ptr)
	return ptr
}
