package builtin

import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
)

type WrappedImpl struct {
	// Owner must be public but do not directly modify.
	Owner *GodotObject
}

func (w *WrappedImpl) GetGodotObjectOwner() *GodotObject {
	return w.Owner
}

// AsGDExtensionObjectPtr returns the wrapped Godot object pointer. A nil receiver
// or a nil Owner yields a null pointer rather than panicking, so callers can pass a
// nil or typed-nil object straight into a call argument slot. Godot reads a null
// object pointer as null.
func (w *WrappedImpl) AsGDExtensionObjectPtr() GDExtensionObjectPtr {
	if w == nil || w.Owner == nil {
		return nil
	}
	return (GDExtensionObjectPtr)(unsafe.Pointer(w.Owner))
}

// AsGDExtensionConstObjectPtr is the const form of AsGDExtensionObjectPtr and
// shares its nil-receiver behaviour.
func (w *WrappedImpl) AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr {
	if w == nil || w.Owner == nil {
		return nil
	}
	return (GDExtensionConstObjectPtr)(unsafe.Pointer(w.Owner))
}

func (w *WrappedImpl) AsGDExtensionTypePtr() GDExtensionTypePtr {
	return (GDExtensionTypePtr)(unsafe.Pointer(&w.Owner))
}

func (w *WrappedImpl) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return (GDExtensionConstTypePtr)(unsafe.Pointer(&w.Owner))
}

func (w *WrappedImpl) SetGodotObjectOwner(owner *GodotObject) {
	w.Owner = owner
}

func (w *WrappedImpl) IsNil() bool {
	return w == nil || w.Owner == nil
}

func (w *WrappedImpl) Destroy() {
}

// NewWrappedFromGodotObject creates a new WrappedImpl from an existing GodotObject..
//
//	Wrapped::Wrapped(GodotObject *p_godot_object) in godot-cpp
func NewWrappedFromGodotObject(owner *GodotObject) *WrappedImpl {
	return &WrappedImpl{
		Owner: owner,
	}
}

type WrappedClassInstance struct {
	Instance Object
	// ClassName is the registered extension class name the instance was
	// created under, stored at construction so the class callback prologues
	// read a Go string they already own instead of paying an engine
	// crossing per callback (openspec trim-dispatch-overhead design D1).
	// An immutable Go string: the registration outlives the instance and
	// the field adds no program-lifetime retention.
	ClassName string
	// freeing marks that the engine has begun releasing this instance: the
	// instance-free callback sets it at the re-entrancy guard point, before
	// any Go-side teardown (Destroy) begins. Class callback prologues read
	// it through IsFreeing to decline deliveries that arrive while the
	// wrapper is dying. Both the free callback and the prologues are engine
	// callbacks on the main thread, so a plain bool suffices.
	freeing bool
}

// MarkFreeing records that the engine has begun releasing this instance.
// The instance-free callback is the sole caller, at the point its
// re-entrancy guard admits the first (and only) free; calling it again —
// e.g. from a re-entrant pass that the guard then trips — is a no-op.
func (w *WrappedClassInstance) MarkFreeing() {
	w.freeing = true
}

// IsFreeing reports whether the engine has begun releasing this instance.
// A nil wrapper reports false: it holds no instance to free.
func (w *WrappedClassInstance) IsFreeing() bool {
	return w != nil && w.freeing
}

func (w *WrappedClassInstance) GetClassName() string {
	return w.Instance.GetClassName()
}

func (w *WrappedClassInstance) GetParentClassName() string {
	return w.Instance.GetParentClassName()
}

func (w *WrappedClassInstance) GetGodotObjectOwner() *GodotObject {
	return w.Instance.GetGodotObjectOwner()
}

// godotObjectOwner returns the wrapped owner, or nil when the receiver or its
// Instance is nil. Shared by the object-pointer accessors so a nil or typed-nil
// instance never panics at the boundary.
func (w *WrappedClassInstance) godotObjectOwner() *GodotObject {
	if w == nil || w.Instance == nil {
		return nil
	}
	return w.Instance.GetGodotObjectOwner()
}

// AsGDExtensionObjectPtr returns the wrapped Godot object pointer. A nil receiver
// or nil Instance yields a null pointer rather than panicking, matching
// WrappedImpl.AsGDExtensionObjectPtr.
func (w *WrappedClassInstance) AsGDExtensionObjectPtr() GDExtensionObjectPtr {
	return (GDExtensionObjectPtr)(unsafe.Pointer(w.godotObjectOwner()))
}

// AsGDExtensionConstObjectPtr is the const form of AsGDExtensionObjectPtr and
// shares its nil-receiver behaviour.
func (w *WrappedClassInstance) AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr {
	return (GDExtensionConstObjectPtr)(unsafe.Pointer(w.godotObjectOwner()))
}

func (w *WrappedClassInstance) AsGDExtensionTypePtr() GDExtensionTypePtr {
	return (GDExtensionTypePtr)(unsafe.Pointer(w.Instance.GetGodotObjectOwner()))
}

func (w *WrappedClassInstance) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return (GDExtensionConstTypePtr)(unsafe.Pointer(w.Instance.GetGodotObjectOwner()))
}

func (w *WrappedClassInstance) SetGodotObjectOwner(owner *GodotObject) {
	w.Instance.SetGodotObjectOwner(owner)
}

func (w *WrappedClassInstance) IsNil() bool {
	return w == nil || w.Instance.GetGodotObjectOwner() == nil
}

func (w *WrappedClassInstance) Destroy() {
}
