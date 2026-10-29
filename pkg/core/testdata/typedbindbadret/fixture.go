// Package typedbindbadret is the return-family compile-rejection fixture:
// StringName returns must transfer ownership through the engine copy
// constructor, so StringName is in RetCopyEncodable but not retValueSlot —
// binding it through the value-return overload must fail to compile. The
// harness asserts the diagnostic names the constraint.
package typedbindbadret

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
func (f *Fixture) IsNil() bool                                  { return f == nil }
func (f *Fixture) AsGDExtensionObjectPtr() GDExtensionObjectPtr { return nil }
func (f *Fixture) AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr {
	return nil
}
func (f *Fixture) AsGDExtensionTypePtr() GDExtensionTypePtr { return nil }
func (f *Fixture) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return nil
}

// Name returns a copy-constructor return.
func (f *Fixture) Name() StringName { return StringName{} }

func bind() {
	f := &Fixture{}
	// Wrong family: the value overload's return parameter is constrained to
	// retValueSlot, which excludes every copy-constructor type.
	ClassDBBindValueMethod0(f, (*Fixture).Name, "name", nil, nil)
}
