package gdextension

// Ref is a helper struct for RefCounted Godot Objects.
type Ref struct {
	reference *RefCounted
}

func (cx *Ref) Ptr() *RefCounted {
	return cx.reference
}

// Ref increments a reference counter
func (cx *Ref) Ref(pFrom *Ref) {
	if pFrom.reference == cx.reference {
		return
	}

	cx.Unref()

	cx.reference = pFrom.reference
	if cx.reference != nil {
		(*cx.reference).Reference()
	}
}

func (cx *Ref) RefPointer(pRef *RefCounted) {
	if pRef == nil {
		panic("pRef cannot be nil")
	}

	if (*pRef).InitRef() {
		cx.reference = pRef
	}
}

func (cx *Ref) Unref() {
	if cx.reference != nil && (*cx.reference).Unreference() {
		if destroyable, ok := (*cx.reference).(HasDestructor); ok {
			destroyable.Destroy()
		}
		// release memory
		// runtime.Unpin(cx.reference)
	}
	cx.reference = nil
}

func NewRef(reference *RefCounted) *Ref {
	ref := &Ref{}

	ref.RefPointer(reference)

	return ref
}
