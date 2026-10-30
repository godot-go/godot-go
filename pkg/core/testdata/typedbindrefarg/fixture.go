// Package typedbindrefarg is the reference-argument half of the
// object-argument fallback boundary (see typedbindobjectarg): a
// RefCounted-typed parameter needs the decoder's ref-acquire and wrapper
// construction, so it sits outside nativeSlot and no typed overload admits it.
// This package must fail to compile; the harness asserts the diagnostic names
// the constraint.
package typedbindrefarg

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

type Fixture struct {
	owner *GodotObject
}

func (f *Fixture) Self() *Fixture { return f }

func (f *Fixture) Destroy()                                     {}
func (f *Fixture) GetGodotObjectOwner() *GodotObject            { return f.owner }
func (f *Fixture) SetGodotObjectOwner(owner *GodotObject)       { f.owner = owner }
func (f *Fixture) GetClassName() string                         { return "Fixture" }
func (f *Fixture) GetParentClassName() string                   { return "Object" }
func (f *Fixture) AsGDExtensionObjectPtr() GDExtensionObjectPtr { return nil }
func (f *Fixture) AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr {
	return nil
}
func (f *Fixture) AsGDExtensionTypePtr() GDExtensionTypePtr { return nil }
func (f *Fixture) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return nil
}

// AdoptRef takes a RefCounted reference argument.
func (f *Fixture) AdoptRef(reference *RefBase[Node2D]) int64 { return 0 }

func bind() {
	f := &Fixture{}
	ClassDBBindValueMethod1(f, (*Fixture).AdoptRef, "adopt_ref", []string{"reference"}, nil)
}
