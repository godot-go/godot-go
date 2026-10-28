// Package typedbindok is the positive compile fixture for the typed bind
// overloads: every call here must type-check, proving the fixture harness
// itself builds real binds through the constraint unions. Never executed.
package typedbindok

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// Fixture is a minimal bound-class shape: the GDClass contract plus the one
// self() line the selfTyped constraint requires.
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

// Add takes nativeSlot members and returns a retValueSlot member.
func (f *Fixture) Add(a int64, b float64) float64 { return float64(a) + b }

// Name returns a copy-constructor return, so it belongs to the CopyRet family.
func (f *Fixture) Name() StringName { return StringName{} }

// Echo returns its borrowed StringName argument unchanged.
func (f *Fixture) Echo(name StringName) StringName { return name }

func bind() {
	f := &Fixture{}
	ClassDBBindValueMethod2(f, (*Fixture).Add, "add", []string{"a", "b"}, nil)
	ClassDBBindCopyRetMethod0(f, (*Fixture).Name, "name", nil, nil)
	ClassDBBindCopyRetMethod1(f, (*Fixture).Echo, "echo", []string{"name"}, nil)
	ClassDBBindStaticValueMethod0(f, func() int64 { return 1 }, "one", nil, nil)
}
