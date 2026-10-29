// Package typedbindobjectarg is the object-argument fallback-boundary fixture:
// Object and Ref parameters sit outside nativeSlot on purpose, because decoding
// them is per-call instance-binding work (class-name resolution plus
// wrapper/ref construction, see resolveObjectArgument) that a uniform raw load
// cannot express. No typed overload therefore admits such a signature, and the
// bind stays reflective through the name-string API. This package exercises the
// Object half; typedbindrefarg exercises the Ref half. Both must fail to
// compile; the harness asserts the diagnostic names the constraint.
package typedbindobjectarg

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

// Adopt takes an engine-class object argument.
func (f *Fixture) Adopt(node Node2D) int64 { return 0 }

func bind() {
	f := &Fixture{}
	ClassDBBindValueMethod1(f, (*Fixture).Adopt, "adopt", []string{"node"}, nil)
}
