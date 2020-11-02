package gdnativetest

import (
	"math/rand"
	"strings"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/gdnative"
	"github.com/godot-go/godot-go/pkg/log"
)

// PlayerCharacter is a custom NativeScript class that represents an entity in the game
// that the player controls
type PlayerCharacter struct {
	gdnative.KinematicBody2DImpl
	gdnative.UserDataIdentifiableImpl

	walkAnimation gdnative.AnimationPlayer
	velocity      gdnative.Vector2
	floorNormal   gdnative.Vector2
	name          gdnative.String
}

// ClassName is required for NativeScript class registration with Godot.
func (p *PlayerCharacter) ClassName() string {
	return "PlayerCharacter"
}

// BaseClass is required for NativeScript class registration with Godot.
func (p *PlayerCharacter) BaseClass() string {
	return "KinematicBody2D"
}

// RandomTeleport teleports the player to a random location within the range of distance.
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


// RandomName generates a random 10 character string to assign as the player's name.
func (p *PlayerCharacter) RandomName() {
	name := randomString(10)
	gstrName := gdnative.NewStringFromGoString(name)
	p.SetName(gstrName)
}

// Init should be used to initialize anything before the class gets added to a Scene.
func (p *PlayerCharacter) Init() {

}

// OnClassRegistered is designed have all methods, signals, and properties registered in the function.
func (p *PlayerCharacter) OnClassRegistered(e gdnative.ClassRegisteredEvent) {
	// methods
	e.RegisterMethod("_ready", "Ready")
	e.RegisterMethod("_physics_process", "PhysicsProcess")
	e.RegisterMethod("random_teleport", "RandomTeleport")
	e.RegisterMethod("random_name", "RandomName")
	e.RegisterMethod("run_ginkgo_testsuite", "RunGinkgoTestSuite")

	// signals
	e.RegisterSignal("moved", gdnative.RegisterSignalArg{"velocity", gdnative.GODOT_VARIANT_TYPE_VECTOR2})
	e.RegisterSignal("name_changed")

	// properties
	e.RegisterProperty("velocity", "SetVelocity", "GetVelocity", defaultVelocity)
	e.RegisterProperty("name", "SetName", "GetName", defaultName)
}

// RunGinkgoTestSuite is the entrypoint for the ginkgo test suite
func (p *PlayerCharacter) RunGinkgoTestSuite() {
	runTests()
}

// GetVelocity returns the velocity to Godot
func (p *PlayerCharacter) GetVelocity() gdnative.Variant {
	return gdnative.NewVariantVector2(p.velocity)
}

func (p *PlayerCharacter) setVelocity(v gdnative.Vector2) {
	m := 2 * 5 * 16 * 16
	n := v.Normalized()
	p.velocity = n.OperatorMultiplyScalar(float32(m))
}

// SetVelocity is called when the velocity property is modified by Godot
func (p *PlayerCharacter) SetVelocity(v gdnative.Variant) {
	vec2 := v.AsVector2()
	p.setVelocity(vec2)
}

// GetName returns the name to Godot
func (p *PlayerCharacter) GetName() gdnative.String {
	return p.name
}

// SetName is called when the name property is modified by Godot
func (p *PlayerCharacter) SetName(v gdnative.String) {
	if v != p.name {
		p.name = v
		p.EmitSignal("name_changed")
	}

}

// Ready is mapped to the _ready method, which is called after being added to the Scene.
func (p *PlayerCharacter) Ready() {
	log.Debug("start PlayerCharacter::Ready", gdnative.GodotObjectField("owner", p.GetOwnerObject()))
	nodePath := gdnative.NewNodePath("sprite/animation_player")

	log.Info("searching path...", gdnative.NodePathField("path", nodePath))

	n := p.GetNode(nodePath)
	pno := n.GetOwnerObject()
	tt := n.GetTypeTag()
	p.walkAnimation = gdnative.NewAnimationPlayerWithRef(pno, tt)

	if !p.walkAnimation.HasAnimation("walk-right") {
		log.Panic("unable to find walk-right animation")
	}

	if !p.walkAnimation.HasAnimation("walk-left") {
		log.Panic("unabel to find walk-left animation")
	}

	if !p.walkAnimation.HasAnimation("walk-down") {
		log.Panic("unable to find walk-down")
	}

	if !p.walkAnimation.HasAnimation("walk-up") {
		log.Panic("unable to find walk-up")
	}

	if !p.walkAnimation.HasAnimation("idle-right") {
		log.Panic("unable to find idle-right")
	}

	if !p.walkAnimation.HasAnimation("idle-left") {
		log.Panic("unable to find idle-left")
	}

	if !p.walkAnimation.HasAnimation("idle-down") {
		log.Panic("unable to find idle-down")
	}

	if !p.walkAnimation.HasAnimation("idle-up") {
		log.Panic("unable to find idle-up")
	}

	log.Debug("End PlayerCharacter::Ready", gdnative.GodotObjectField("owner", p.GetOwnerObject()))
}

// PhysicsProcess is mapped to the _physics_process method.
func (p *PlayerCharacter) PhysicsProcess(delta float64) {
	p.setVelocity(getKeyInputDirectionAsVector2())

	p.updateSprite(delta)

	v := p.velocity.OperatorMultiplyScalar(float32(delta))

	nv := p.MoveAndSlide(v, p.floorNormal, false, 4, 0.785398, true)

	variant := gdnative.NewVariantVector2(nv)
	defer variant.Destroy()
	p.EmitSignal("moved", &variant)
}

func (p *PlayerCharacter) updateSprite(delta float64) {
	x := p.velocity.GetX()
	y := p.velocity.GetY()

	a := p.walkAnimation
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

// Free should be called to clean up memory
func (p *PlayerCharacter) Free() {
	log.Debug("free PlayerCharacter")

	p.name.Destroy()

	p.walkAnimation = nil

	if p != nil {
		gdnative.Free(unsafe.Pointer(p))
		p = nil
	}
}

// NewPlayerCharacter is the constructor that creates an instance recognized by Godot
func NewPlayerCharacter() PlayerCharacter {
	log.Debug("NewPlayerCharacter")
	inst := *(gdnative.CreateCustomClassInstance("PlayerCharacter", "KinematicBody2D").(*PlayerCharacter))
	return inst
}

var (
	defaultVelocity gdnative.Variant
	defaultName gdnative.Variant
)

// PlayerCharacterNativescriptInit called after NativeScript initializes
func PlayerCharacterNativescriptInit() {
	defaultVelocity = gdnative.NewVariantVector2(gdnative.NewVector2(0.0, 0.0))
	defaultName = gdnative.NewVariantString("No_Name")

	gdnative.RegisterClass(&PlayerCharacter{})
}

// PlayerCharacterNativescriptTerminate called before NativeScript terminates
func PlayerCharacterNativescriptTerminate() {
	defaultVelocity.Destroy()
	defaultName.Destroy()
}

func init() {
	gdnative.RegisterInitCallback(PlayerCharacterNativescriptInit)
	gdnative.RegisterTerminateCallback(PlayerCharacterNativescriptTerminate)
}