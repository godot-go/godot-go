package pkg

import (
	"github.com/godot-go/godot-go/pkg/gdextension"
)

var (
	ExampleRef_InstanceCount int32
	ExampleRef_LastId int32
)

// Example implements GDClass evidence
var _ gdextension.GDClass = new(ExampleRef)

type ExampleRef struct {
	gdextension.RefCountedImpl
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
