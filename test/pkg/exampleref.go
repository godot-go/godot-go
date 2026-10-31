package pkg

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
	"github.com/godot-go/godot-go/pkg/log"
)

// ExampleRef implements GDClass evidence
var _ RefCounted = (*ExampleRef)(nil)

type ExampleRef struct {
	RefCountedImpl
	Id int32
}

func (c *ExampleRef) Self() *ExampleRef { return c }

func (c *ExampleRef) GetClassName() string {
	return "ExampleRef"
}

func (c *ExampleRef) GetParentClassName() string {
	return "RefCounted"
}

// IsNil is declared on the outer type per the Wrapped contract: the call
// promoted through the embedded RefCountedImpl would fault on a typed nil.
func (c *ExampleRef) IsNil() bool {
	return c == nil
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
	ClassDBRegisterClass(NewExampleRefFromOwnerObject, nil, nil, func(t *ExampleRef) {
		ClassDBBindValueMethod0(t, (*ExampleRef).GetId, "get_id", nil, nil)
		// SetId stays name-string-bound: the typed overload surface has no
		// fixed-arity void-return family, so this class also exercises the
		// reflective fallback for the property set path.
		ClassDBBindMethod(t, "SetId", "set_id", []string{"id"}, nil)
		ClassDBAddProperty(t, GDEXTENSION_VARIANT_TYPE_INT, "group_subgroup_id", "set_id", "get_id")
		log.Debug("ExampleRef registered")
	})
}

func UnregisterClassExampleRef() {
	ClassDBUnregisterClass[*ExampleRef]()
	log.Debug("ExampleRef unregistered")
}
