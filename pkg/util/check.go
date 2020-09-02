package util

import (
	"github.com/godot-go/godot-go/pkg/gdnative"
	"github.com/godot-go/godot-go/pkg/log"
	"strings"
	"unsafe"
	"math/rand"
)

var (
	defaultVelocity gdnative.Variant
	defaultName gdnative.Variant
)

type Check struct {
	gdnative.KinematicBody2DImpl
	gdnative.UserDataIdentifiableImpl

	walkAnimation gdnative.AnimationPlayer
	velocity      gdnative.Vector2
	floorNormal   gdnative.Vector2
	name          gdnative.String
}

func (p *Check) ClassName() string {
	return "PlayerCharacter"
}

func (p *Check) BaseClass() string {
	return "KinematicBody2D"
}

func (p *Check) Init() {

}

func (p *Check) OnClassRegistered(e gdnative.ClassRegisteredEvent) {
}

func (h *Check) Ready() {
	
}


func (p *Check) Free() {
	log.WithFields(gdnative.WithObject(p.GetOwnerObject())).Trace("free PlayerCharacter")

	p.walkAnimation = nil

	if p != nil {
		gdnative.Free(unsafe.Pointer(p))
		p = nil
	}
}

func CheckCreateFunc(owner *gdnative.GodotObject, typeTag gdnative.TypeTag) gdnative.NativeScriptClass {
	m := &Check{}
	m.Owner = owner
	m.TypeTag = typeTag

	return m
}

func NewCheck() Check {
	log.Trace("NewCheck")
	inst := *(gdnative.CreateCustomClassInstance("Check", "KinematicBody2D").(*Check))
	return inst
}
