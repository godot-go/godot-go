package gdnativetest

import (
	"github.com/pcting/godot-go/pkg/gdnative"
	"github.com/pcting/godot-go/pkg/log"
	"strings"
	"unsafe"
	"math/rand"
)

var (
	defaultVelocity gdnative.Variant
	defaultName gdnative.Variant
)

type PlayerCharacter struct {
	gdnative.KinematicBody2DImpl
	gdnative.UserDataIdentifiable

	walkAnimation gdnative.AnimationPlayer
	velocity    gdnative.Vector2
	floorNormal gdnative.Vector2
	name         gdnative.String
}

func (p *PlayerCharacter) ClassName() string {
	return "PlayerCharacter"
}

func (p *PlayerCharacter) BaseClass() string {
	return "KinematicBody2D"
}

func (p *PlayerCharacter) Blink() {
	pos := p.GetPosition()
	v := gdnative.NewVector2(0.5 - rand.Float32(), 0.5 - rand.Float32())
	normalized := v.Normalized()
	p.SetPosition(pos.OperatorAdd(normalized.OperatorMultiplyScalar(50.0)))
}

func (p *PlayerCharacter) OnClassRegistered(e gdnative.ClassRegisteredEvent) {
	e.RegisterMethod("_ready", "Ready")
	e.RegisterMethod("_physics_process", "PhysicsProcess")
	e.RegisterMethod("blink", "Blink")
	e.RegisterSignal("moved", gdnative.RegisterSignalArg{"velocity", gdnative.GODOT_VARIANT_TYPE_VECTOR2})

	
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

func (h *PlayerCharacter) GetName() gdnative.Variant {
	v := gdnative.NewVariantString(h.name)
	return v
}

func (h *PlayerCharacter) SetName(v gdnative.Variant) {
	h.name = v.AsString()
}

func (h *PlayerCharacter) Ready() {
	log.WithFields(gdnative.WithObject(h.GetOwnerObject())).Trace("Start PlayerCharacter::Ready")
	strP := gdnative.NewStringFromGoString("sprite/animation_player")
	p := gdnative.NewNodePath(strP)

	log.WithFields(gdnative.WithGodotString("path", p.AsString())).Info("searching path...")

	n := h.GetNode(p)
	pno := n.GetOwnerObject()
	tt := n.GetTypeTag()
	h.walkAnimation = gdnative.NewAnimationPlayerWithRef(pno, tt)

	if !h.walkAnimation.HasAnimation(walkRight) {
		log.Panic("unable to find walk-right animation")
	}
	
	if !h.walkAnimation.HasAnimation(walkLeft) {
		log.Panic("unabel to find walk-left animation")
	}
	
	if !h.walkAnimation.HasAnimation(walkDown) {
		log.Panic("unable to find walk-down")
	}

	if !h.walkAnimation.HasAnimation(walkUp) {
		log.Panic("unable to find walk-up")
	}

	if !h.walkAnimation.HasAnimation(idleRight) {
		log.Panic("unable to find idle-right")
	}
	
	if !h.walkAnimation.HasAnimation(idleLeft) {
		log.Panic("unable to find idle-left")
	}
	
	if !h.walkAnimation.HasAnimation(idleDown) {
		log.Panic("unable to find idle-down")
	}
	
	if !h.walkAnimation.HasAnimation(idleUp) {
		log.Panic("unable to find idle-up")
	}

	log.WithFields(gdnative.WithObject(h.GetOwnerObject())).Trace("End PlayerCharacter::Ready")
}

func (h *PlayerCharacter) PhysicsProcess(delta float64) {
	h.setVelocity(getKeyInputDirectionAsVector2())

	h.updateSprite(delta)

	v := h.velocity.OperatorMultiplyScalar(float32(delta))

	nv := h.MoveAndSlide(v, h.floorNormal, false, 4, 0.785398, true)

	variant := gdnative.NewVariantVector2(nv)
	defer variant.Destroy()
	h.EmitSignal(moved, &variant)
}

func (h *PlayerCharacter) updateSprite(delta float64) {
	x := h.velocity.GetX()
	y := h.velocity.GetY()

	a := h.walkAnimation
	ca := a.GetCurrentAnimation()
	pca := &ca

	if x > 0 {
		if !pca.OperatorEqual(walkRight) {
			a.Play(walkRight, -1, 1.0, false)
		}
	} else if x < 0 {
		if !pca.OperatorEqual(walkLeft) {
			a.Play(walkLeft, -1, 1.0, true)
		}
	} else if y > 0 {
		if !pca.OperatorEqual(walkDown) {
			a.Play(walkDown, -1, 1.0, false)
		}
	} else if y < 0 {
		if !pca.OperatorEqual(walkUp) {
			a.Play(walkUp, -1, 1.0, false)
		}
	} else {
		// switch to idle animation if the character isn't moving
		name := pca.AsGoString()

		if name != "" {
			tokens := strings.Split(name, "-")
			
			if len(tokens) != 2  {
				log.WithField("name", name).Panic("unable to parse animation name")
			}

			var animationName gdnative.String
			switch tokens[1] {
			case "up":
				animationName = idleUp
			case "down":
				animationName = idleDown
			case "left":
				animationName = idleLeft
			case "right":
				animationName = idleRight
			default:
				log.WithField("name", name).Warn("unhandled animation name")
			}

			if !pca.OperatorEqual(animationName) {
				a.Play(animationName, -1, 1.0, false)
			}
		}
	}
}

func isActionPressedToInt8(a gdnative.String) int8 {
	if gdnative.GetSingletonInput().IsActionPressed(a) {
		return 1
	} else {
		return 0
	}
}

func getKeyInputDirectionAsVector2() gdnative.Vector2 {
	return gdnative.NewVector2(
		float32(isActionPressedToInt8(uiRight)-isActionPressedToInt8(uiLeft)),
		float32(isActionPressedToInt8(uiDown)-isActionPressedToInt8(uiUp)),
	)
}

func (p *PlayerCharacter) Free() {
	log.WithFields(gdnative.WithObject(p.GetOwnerObject())).Trace("free PlayerCharacter")

	p.walkAnimation = nil

	if p != nil {
		gdnative.Free(unsafe.Pointer(p))
		p = nil
	}
}

func PlayerCharacterCreateFunc(owner *gdnative.GodotObject, typeTag gdnative.TypeTag) gdnative.NativeScriptClass {
	log.WithFields(gdnative.WithObject(owner)).Trace("create_func new PlayerCharacter")

	m := &PlayerCharacter{}
	m.Owner = owner
	m.TypeTag = typeTag

	return m
}

func NewPlayerCharacter() PlayerCharacter {
	log.Trace("NewPlayerCharacter")
	inst := *(gdnative.CreateCustomClassInstance("PlayerCharacter", "KinematicBody2D").(*PlayerCharacter))
	return inst
}

var (
	moved gdnative.String
	velocity gdnative.String
	velocityVariant gdnative.Variant

	uiRight gdnative.String
	uiLeft gdnative.String
	uiUp gdnative.String
	uiDown gdnative.String

	walkRight gdnative.String
	walkLeft gdnative.String
	walkUp gdnative.String
	walkDown gdnative.String

	idleRight gdnative.String
	idleLeft gdnative.String
	idleUp gdnative.String
	idleDown gdnative.String
)

func PlayerCharacterNativescriptInit() {
	moved = gdnative.NewStringFromGoString("moved")
	velocity = gdnative.NewStringFromGoString("velocity")
	velocityVariant = gdnative.NewVariantString(velocity)

	uiRight = gdnative.NewStringFromGoString("ui_right")
	uiLeft = gdnative.NewStringFromGoString("ui_left")
	uiUp = gdnative.NewStringFromGoString("ui_up")
	uiDown = gdnative.NewStringFromGoString("ui_down")

	walkRight = gdnative.NewStringFromGoString("walk-right")
	walkLeft = gdnative.NewStringFromGoString("walk-left")
	walkUp = gdnative.NewStringFromGoString("walk-up")
	walkDown = gdnative.NewStringFromGoString("walk-down")

	idleRight = gdnative.NewStringFromGoString("idle-right")
	idleLeft = gdnative.NewStringFromGoString("idle-left")
	idleUp = gdnative.NewStringFromGoString("idle-up")
	idleDown = gdnative.NewStringFromGoString("idle-down")

	defaultVelocity = gdnative.NewVariantVector2(gdnative.NewVector2(0.0, 0.0))
	defaultName = gdnative.NewVariantString(gdnative.NewStringFromGoString("No_Name"))
}

func PlayerCharacterNativescriptTerminate() {
	moved.Destroy()
	velocity.Destroy()
	velocityVariant.Destroy()

	uiRight.Destroy()
	uiLeft.Destroy()
	uiUp.Destroy()
	uiDown.Destroy()

	walkRight.Destroy()
	walkLeft.Destroy()
	walkUp.Destroy()
	walkDown.Destroy()

	idleRight.Destroy()
	idleLeft.Destroy()
	idleUp.Destroy()
	idleDown.Destroy()

	defaultVelocity.Destroy()
	defaultName.Destroy()
}
