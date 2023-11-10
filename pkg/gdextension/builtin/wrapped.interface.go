package builtin

import (
	. "github.com/godot-go/godot-go/pkg/gdextension/ffi"
)

type GDExtensionClass interface {
	Wrapped
}

type HasDestructor interface {
	Destroy()
}

// Base for all engine classes, to contain the pointer to the engine instance.
type Wrapped interface {
	HasDestructor
	GetGodotObjectOwner() *GodotObject
	SetGodotObjectOwner(owner *GodotObject)
	GetClassName() string
	GetParentClassName() string
	AsGDExtensionObjectPtr() GDExtensionObjectPtr
	AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr
	AsGDExtensionTypePtr() GDExtensionTypePtr
	AsGDExtensionConstTypePtr() GDExtensionConstTypePtr
}
