package pkg

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
	"github.com/godot-go/godot-go/pkg/log"
)

// ExampleRef implements GDClass evidence
var _ RefCounted = new(ExampleRef)

type ExampleRef struct {
	RefCountedImpl
	Id int32
}

func (c *ExampleRef) GetClassName() string {
	return "ExampleRef"
}

func (c *ExampleRef) GetParentClassName() string {
	return "RefCounted"
}

func (e *ExampleRef) SetId(id int32) {
	e.Id = id
}

func (e *ExampleRef) GetId() int32 {
	return e.Id
}

func NewExampleRefFromOwnerObject(owner *GodotObject) GDClass {
	obj := &ExampleRef{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

func RegisterClassExampleRef() {
	ClassDBRegisterClass(&ExampleRef{}, NewExampleRefFromOwnerObject, []GDExtensionPropertyInfo{}, nil, func(t GDClass) {
		ClassDBBindMethod(t, "GetId", "get_id", nil, nil)
		ClassDBBindMethod(t, "SetId", "set_id", []string{"id"}, nil)
		ClassDBAddProperty(t, GDEXTENSION_VARIANT_TYPE_INT, "group_subgroup_id", "set_id", "get_id")
		log.Debug("ExampleRef registered")
	})
}
