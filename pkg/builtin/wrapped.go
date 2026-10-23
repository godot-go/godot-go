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
