package gdnativetest

import (
	"math/rand"
	"strings"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/gdnative"
	"github.com/godot-go/godot-go/pkg/log"
)

type PlayerCharacter struct {
	gdnative.KinematicBody2DImpl
	gdnative.UserDataIdentifiableImpl

	walkAnimation gdnative.AnimationPlayer
	velocity      gdnative.Vector2
	floorNormal   gdnative.Vector2
	name          gdnative.String
}

func (p *PlayerCharacter) ClassName() string {
	return "PlayerCharacter"
}

func (p *PlayerCharacter) BaseClass() string {
	return "KinematicBody2D"
}

func (p *PlayerCharacter) RandomTeleport(distance float64) {
	pos := p.GetPosition()
	v := gdnative.NewVector2(rand.Float32() - 0.5, rand.Float32() - 0.5)
	normalized := v.Normalized()
	p.SetPosition(pos.OperatorAdd(normalized.OperatorMultiplyScalar(float32(distance))))
}

func randomString(len int) string {
	randomInt := func (min, max int) int {
		return min + rand.Intn(max-min)
	}

    bytes := make([]byte, len)
    for i := 0; i < len; i++ {
        bytes[i] = byte(randomInt(65, 90))
    }
    return string(bytes)
}

func (p *PlayerCharacter) RandomName() {
	name := randomString(10)
	gstrName := gdnative.NewStringFromGoString(name)
	p.SetName(gstrName)
}

func (p *PlayerCharacter) Init() {

}

func (p *PlayerCharacter) OnClassRegistered(e gdnative.ClassRegisteredEvent) {
	// methods
	e.RegisterMethod("_ready", "Ready")
	e.RegisterMethod("_physics_process", "PhysicsProcess")
	e.RegisterMethod("random_teleport", "RandomTeleport")
	e.RegisterMethod("random_name", "RandomName")

	// signals
	e.RegisterSignal("moved", gdnative.RegisterSignalArg{"velocity", gdnative.GODOT_VARIANT_TYPE_VECTOR2})
	e.RegisterSignal("name_changed")

	// properties
	e.RegisterProperty("velocity", "SetVelocity", "GetVelocity", defaultVelocity)
	e.RegisterProperty("name", "SetName", "GetName", defaultName)
}

func (h *PlayerCharacter) GetVelocity() gdnative.Variant {
	return gdnative.NewVariantVector2(h.velocity)
}

func (h *PlayerCharacter) setVelocity(v gdnative.Vector2) {
	m := 2 * 5 * 16 * 16
	n := v.Normalized()
	h.velocity = n.OperatorMultiplyScalar(float32(m))
}

func (h *PlayerCharacter) SetVelocity(v gdnative.Variant) {
	vec2 := v.AsVector2()
	h.setVelocity(vec2)
}

func (h *PlayerCharacter) GetName() gdnative.String {
	return h.name
}

func (h *PlayerCharacter) SetName(v gdnative.String) {
	if v != h.name {
		h.name = v
		h.EmitSignal("name_changed")
	}

}

func (h *PlayerCharacter) Ready() {
	log.Debug("start PlayerCharacter::Ready", gdnative.GodotObjectField("owner", h.GetOwnerObject()))
	path := "sprite/animation_player"
	strP := gdnative.NewStringFromGoString(path)
	p := gdnative.NewNodePath(strP)

	log.Info("searching path...", gdnative.StringField("path", path))

	n := h.GetNode(p)
	pno := n.GetOwnerObject()
	tt := n.GetTypeTag()
	h.walkAnimation = gdnative.NewAnimationPlayerWithRef(pno, tt)

	if !h.walkAnimation.HasAnimation("walk-right") {
		log.Panic("unable to find walk-right animation")
	}

	if !h.walkAnimation.HasAnimation("walk-left") {
		log.Panic("unabel to find walk-left animation")
	}

	if !h.walkAnimation.HasAnimation("walk-down") {
		log.Panic("unable to find walk-down")
	}

	if !h.walkAnimation.HasAnimation("walk-up") {
		log.Panic("unable to find walk-up")
	}

	if !h.walkAnimation.HasAnimation("idle-right") {
		log.Panic("unable to find idle-right")
	}

	if !h.walkAnimation.HasAnimation("idle-left") {
		log.Panic("unable to find idle-left")
	}

	if !h.walkAnimation.HasAnimation("idle-down") {
		log.Panic("unable to find idle-down")
	}

	if !h.walkAnimation.HasAnimation("idle-up") {
		log.Panic("unable to find idle-up")
	}

	log.Debug("End PlayerCharacter::Ready", gdnative.GodotObjectField("owner", h.GetOwnerObject()))
}

func (h *PlayerCharacter) PhysicsProcess(delta float64) {
	h.setVelocity(getKeyInputDirectionAsVector2())

	h.updateSprite(delta)

	v := h.velocity.OperatorMultiplyScalar(float32(delta))

	nv := h.MoveAndSlide(v, h.floorNormal, false, 4, 0.785398, true)

	variant := gdnative.NewVariantVector2(nv)
	defer variant.Destroy()
	h.EmitSignal("moved", &variant)
}

func (h *PlayerCharacter) updateSprite(delta float64) {
	x := h.velocity.GetX()
	y := h.velocity.GetY()

	a := h.walkAnimation
	ca := a.GetCurrentAnimation()

	if x > 0 {
		if ca != "walk-right" {
			a.Play("walk-right", -1, 1.0, false)
		}
	} else if x < 0 {
		if ca != "walk-left" {
			a.Play("walk-left", -1, 1.0, true)
		}
	} else if y > 0 {
		if ca != "walk-down" {
			a.Play("walk-down", -1, 1.0, false)
		}
	} else if y < 0 {
		if ca != "walk-up" {
			a.Play("walk-up", -1, 1.0, false)
		}
	} else if ca != "" {
		tokens := strings.Split(ca, "-")

		if len(tokens) != 2  {
			log.Panic("unable to parse animation name", gdnative.StringField("name", ca))
		}

		var animationName string
		switch tokens[1] {
		case "up":
			animationName = "idle-up"
		case "down":
			animationName = "idle-down"
		case "left":
			animationName = "idle-left"
		case "right":
			animationName = "idle-right"
		default:
			log.Warn("unhandled animation name", gdnative.StringField("name", ca))
		}

		if ca != animationName {
			a.Play(animationName, -1, 1.0, false)
		}
	}
}

func isActionPressedToInt8(a string) int8 {
	if gdnative.GetSingletonInput().IsActionPressed(a) {
		return 1
	} else {
		return 0
	}
}

func getKeyInputDirectionAsVector2() gdnative.Vector2 {
	return gdnative.NewVector2(
		float32(isActionPressedToInt8("ui_right")-isActionPressedToInt8("ui_left")),
		float32(isActionPressedToInt8("ui_down")-isActionPressedToInt8("ui_up")),
	)
}

func (p *PlayerCharacter) Free() {
	log.Debug("free PlayerCharacter")

	p.name.Destroy()

	p.walkAnimation = nil

	if p != nil {
		gdnative.Free(unsafe.Pointer(p))
		p = nil
	}
}

func NewPlayerCharacter() PlayerCharacter {
	log.Debug("NewPlayerCharacter")
	inst := *(gdnative.CreateCustomClassInstance("PlayerCharacter", "KinematicBody2D").(*PlayerCharacter))
	return inst
}

var (
	defaultVelocity gdnative.Variant
	defaultName gdnative.Variant
)

func PlayerCharacterNativescriptInit() {
	defaultVelocity = gdnative.NewVariantVector2(gdnative.NewVector2(0.0, 0.0))
	defaultName = gdnative.NewVariantString(gdnative.NewStringFromGoString("No_Name"))
}

func PlayerCharacterNativescriptTerminate() {
	defaultVelocity.Destroy()
	defaultName.Destroy()
}
