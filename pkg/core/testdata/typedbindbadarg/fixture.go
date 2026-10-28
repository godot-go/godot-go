// Package typedbindbadarg is the argument-union compile-rejection fixture:
// float32 is deliberately excluded from nativeSlot (the reflective decoder
// double-loads then truncates, so a raw load would break parity), and no
// typed overload admits it. This package must fail to compile; the harness
// asserts the diagnostic names the constraint.
package typedbindbadarg

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

// Scaled takes a float32: outside every argument constraint union.
func (f *Fixture) Scaled(a float32) {}

func bind() {
	f := &Fixture{}
	ClassDBBindValueMethod1(f, (*Fixture).Scaled, "scaled", []string{"a"}, nil)
}
