package gdnative

// #include <godot/gdnative_interface.h>
// #include "gdnative_binding_wrapper.h"
import "C"

type GodotObject [0]byte

// Base for all engine classes, to contain the pointer to the engine instance.
type Wrapped interface {
	// GetExtensionClass() TypeName
	GetGodotObjectOwner() *GodotObject
	SetGodotObjectOwner(owner *GodotObject)
}

type wrapped struct {
	// Must be public but you should not touch this.
	Owner *GodotObject
}

// func (w *wrapped) GetExtensionClass() TypeName {
// 	return (TypeName)("Wrapped")
// }

func (w *wrapped) GetGodotObjectOwner() *GodotObject {
	return w.Owner
}

func (w *wrapped) SetGodotObjectOwner(owner *GodotObject) {
	w.Owner = owner
}

// Comment out because lack of use
// func NewWrappedByGodotClassName(godotClassName string) Wrapped {
// 	owner := GDNativeInterface_classdb_construct_object(internal.gdnInterface, godotClassName)
// 	return &wrapped{
// 		Owner: (*GodotObject)(owner),
// 	}
// }

func NewWrappedFromGodotObject(pGodotObject *GodotObject) Wrapped {
	return &wrapped{
		Owner: pGodotObject,
	}
}
