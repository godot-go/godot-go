package builtin

import (
	. "github.com/godot-go/godot-go/pkg/ffi"
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
	// IsNil reports whether the wrapper pointer itself is nil. A typed nil
	// held by this interface can only answer through a method declared on
	// the outer type — a call promoted through an embedded parent faults
	// while Go addresses the embedded field. Generated classes emit it from
	// classes.go.tmpl; a hand-written wrapper embedding an Impl must declare
	// `func (x *MyType) IsNil() bool { return x == nil }` itself.
	IsNil() bool
}
