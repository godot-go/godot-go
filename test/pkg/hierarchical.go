package pkg

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
	"github.com/godot-go/godot-go/pkg/log"
)

// TestHierarchicalBase is the shallow level of the two-level hierarchy demo.
// It embeds the generated engine Impl leaf directly and declares a
// base-level qualified virtual implementation.
type TestHierarchicalBase struct {
	ControlImpl
}

func (t *TestHierarchicalBase) GetClassName() string {
	return "TestHierarchicalBase"
}

func (t *TestHierarchicalBase) GetParentClassName() string {
	return "Control"
}

// V_TestHierarchicalBase_GetMaximumSize is the base-level implementation of
// the _get_maximum_size virtual under the qualified naming convention.
func (t *TestHierarchicalBase) V_TestHierarchicalBase_GetMaximumSize() Vector2 {
	return NewVector2WithFloat32Float32(11, 22)
}

func NewTestHierarchicalBaseFromOwnerObject(owner *GodotObject) GDClass {
	obj := &TestHierarchicalBase{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

// TestHierarchicalDerived embeds the wrapper level alongside the engine
// Impl leaf, so both levels' implementations coexist in its method set
// instead of shadowing each other. Registration machinery requires the
// engine Impl as the first field with a matching declared parent, so the
// engine sees both classes as Control subclasses; the Go-side embedding
// chain is what drives qualified-virtual resolution.
type TestHierarchicalDerived struct {
	ControlImpl
	TestHierarchicalBase
}

func (t *TestHierarchicalDerived) GetClassName() string {
	return "TestHierarchicalDerived"
}

func (t *TestHierarchicalDerived) GetParentClassName() string {
	return "Control"
}

// V_TestHierarchicalDerived_GetMaximumSize is the most-derived
// implementation; it extends rather than replaces the base behavior by
// delegating explicitly through the promoted selector.
func (t *TestHierarchicalDerived) V_TestHierarchicalDerived_GetMaximumSize() Vector2 {
	base := t.TestHierarchicalBase.V_TestHierarchicalBase_GetMaximumSize()
	return NewVector2WithFloat32Float32(base.MemberGetx()+100, base.MemberGety())
}

func NewTestHierarchicalDerivedFromOwnerObject(owner *GodotObject) GDClass {
	obj := &TestHierarchicalDerived{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

func RegisterClassTestHierarchicalBase() {
	ClassDBRegisterClass(NewTestHierarchicalBaseFromOwnerObject, nil, nil, func(t *TestHierarchicalBase) {
		ClassDBBindMethodVirtual(t, "V_TestHierarchicalBase_GetMaximumSize", "_get_maximum_size", nil, nil)
	})
	log.Debug("TestHierarchicalBase registered")
}

func RegisterClassTestHierarchicalDerived() {
	ClassDBRegisterClass(NewTestHierarchicalDerivedFromOwnerObject, nil, nil, func(t *TestHierarchicalDerived) {
		ClassDBBindMethodVirtual(t, "V_TestHierarchicalDerived_GetMaximumSize", "_get_maximum_size", nil, nil)
	})
	log.Debug("TestHierarchicalDerived registered")
}

func UnregisterClassTestHierarchicalClasses() {
	ClassDBUnregisterClass[*TestHierarchicalDerived]()
	ClassDBUnregisterClass[*TestHierarchicalBase]()
	log.Debug("hierarchical test classes unregistered")
}
