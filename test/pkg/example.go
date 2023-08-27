package pkg

import (
	"fmt"

	"github.com/godot-go/godot-go/pkg/gdextension"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type ExampleEnum int64

const (
	ExampleFirst       ExampleEnum = 0
	AnswerToEverything ExampleEnum = 42
)

const (
	EXAMPLE_ENUM_CONSTANT_WITHOUT_ENUM = 314
)

// Example implements GDClass evidence
var _ gdextension.GDClass = new(Example)

type Example struct {
	gdextension.ControlImpl
	customPosition gdextension.Vector2
	propertyFromList gdextension.Vector3
}

func (c *Example) GetClassName() string {
	return "Example"
}

func (c *Example) GetParentClassName() string {
	return "Control"
}

func (e *Example) SimpleFunc() {
	println("  Simple func called.")
	e.EmitCustomSignal("simple_func", 3)
}

func (e *Example) SimpleConstFunc(a int64) {
	fmt.Printf("  Simple const func called %d.\n", a)
	e.EmitCustomSignal("simple_const_func", 4)
}

func (e *Example) ReturnSomething(base string, f32 float32, f64 float64,
	i int, i8 int8, i16 int16, i32 int32, i64 int64) string {
	println("  Return something called (8 values cancatenated as a string).")
	return fmt.Sprintf("1. %s, 2. %f, 3. %f, 4. %d, 5. %d, 6. %d, 7. %d, 8. %d", base + "42", f32, f64, i, i8, i16, i32, i64)
}

func (e *Example) ReturnSomethingConst() (gdextension.Viewport, error) {
	println("  Return something const called.")
	if !e.IsInsideTree() {
		return nil, fmt.Errorf("unable to get viewport")
	}
	result := e.GetViewport()
	if result == nil {
		println("null viewport encountered")
		return nil, fmt.Errorf("null viewport")
	}
	fmt.Printf("viewport instance id: %v\n", result.GetInstanceId())
	return result, nil
}

func (e *Example) ReturnEmptyRef() ExampleRef {
	var ref ExampleRef
	return ref
}

func (e *Example) GetV4() gdextension.Vector4 {
	v4 := gdextension.NewVector4WithFloat32Float32Float32Float32(1.2, 3.4, 5.6, 7.8)
	log.Debug("vector4 members",
		zap.Any("x", v4.MemberGetx()),
		zap.Any("y", v4.MemberGety()),
		zap.Any("z", v4.MemberGetz()),
		zap.Any("w", v4.MemberGetw()),
	)
	return v4
}

func (e *Example) DefArgs(p_a, p_b int32) int32 {
	return p_a + p_b
}

func (e *Example) TestArray() gdextension.Array {
	arr := gdextension.NewArray()
	arr.Resize(2)
	arr.SetIndexed(0, gdextension.NewVariantInt64(1))
	arr.SetIndexed(1, gdextension.NewVariantInt64(2))
	return arr
}

func (e *Example) TestDictionary() gdextension.Dictionary {
	dict := gdextension.NewDictionary()
	v := gdextension.NewVariantDictionary(dict)
	hello := gdextension.NewStringNameWithUtf8Chars("hello")
	defer hello.Destroy()
	world := gdextension.NewStringWithUtf8Chars("world")
	defer world.Destroy()
	foo := gdextension.NewStringNameWithUtf8Chars("foo")
	defer foo.Destroy()
	bar := gdextension.NewStringWithUtf8Chars("bar")
	defer bar.Destroy()
	v.SetNamed(hello, gdextension.NewVariantString(world))
	v.SetNamed(foo, gdextension.NewVariantString(bar))

	return dict
}

func (e *Example) SetCustomPosition(pos gdextension.Vector2) {
	e.customPosition = pos
}

func (e *Example) GetCustomPosition() gdextension.Vector2 {
	return e.customPosition
}

func (e *Example) SetPropertyFromList(v gdextension.Vector3) {
	e.propertyFromList = v
}

func (e *Example) GetPropertyFromList() gdextension.Vector3 {
	return e.propertyFromList
}

func (e *Example) EmitCustomSignal(name string, value int64) {
	customSignal := gdextension.NewStringNameWithUtf8Chars("custom_signal")
	defer customSignal.Destroy()
	snName := gdextension.NewStringWithLatin1Chars(name)
	defer snName.Destroy()
	log.Info("EmitCustomSignal called",
		zap.String("name", name),
		zap.Int64("value", value),
	)
	e.EmitSignal(
		customSignal,
		gdextension.NewVariantString(snName),
		gdextension.NewVariantInt64(value),
	)
}

func (e *Example) TestCastTo() {
	n, ok := e.CastTo((gdextension.Node)(nil)).(gdextension.Node)
	if !ok {
		log.Panic("failed to cast to cast Example to Node")
	}
	log.Debug("TestCastTo called", zap.Any("class", n.GetClassName()))
}

// func ExampleTestStatic(p_a, p_b int32) int32 {
// 	return p_a + p_b
// }

// func ExampleTestStatic2() {
// 	println("  void static")
// }

func (e *Example) TestStringOps() string {
	s := gdextension.NewStringWithUtf8Chars("A");
	defer s.Destroy()
	sB := gdextension.NewStringWithUtf8Chars("B");
	defer sB.Destroy()
	sC := gdextension.NewStringWithUtf8Chars("C");
	defer sC.Destroy()
	sD := gdextension.NewStringWithUtf32Char(0x010E)
	defer sD.Destroy()
	sE := gdextension.NewStringWithUtf8Chars("E");
	defer sE.Destroy()
	s = s.Add_String(sB)
	s = s.Add_String(sC)
	s = s.Add_String(sD)
	s = s.Add_String(sE)
	return s.ToUtf32()
}

func (e *Example) V_Ready() {
	log.Info("Example_Ready called",
		zap.String("inst", fmt.Sprintf("%p", e)),
	)
	// vector math tests
	v3 := gdextension.NewVector3WithFloat32Float32Float32(1.1, 2.2, 3.3)
	log.Info("Vector3: Created (1.1, 2.2, 3.3)",
		zap.Float32("x", v3.MemberGetx()),
		zap.Float32("y", v3.MemberGety()),
		zap.Float32("z", v3.MemberGetz()),
	)
	v3 = v3.Multiply_float(2.0)
	log.Info("Vector3: Multiply Vector3 by 2",
		zap.Float32("x", v3.MemberGetx()),
		zap.Float32("y", v3.MemberGety()),
		zap.Float32("z", v3.MemberGetz()),
	)
	v3 = v3.Add_Vector3(gdextension.NewVector3WithFloat32Float32Float32(10, 20, 30))
	log.Info("Vector3: Add (1,2,3)",
		zap.Float32("x", v3.MemberGetx()),
		zap.Float32("y", v3.MemberGety()),
		zap.Float32("z", v3.MemberGetz()),
	)
	v3 = v3.Multiply_Vector3(gdextension.NewVector3WithFloat32Float32Float32(5, 10, 15))
	log.Info("Vector3: Multiply (5,10,15)",
		zap.Float32("x", v3.MemberGetx()),
		zap.Float32("y", v3.MemberGety()),
		zap.Float32("z", v3.MemberGetz()),
	)
	v3 = v3.Subtract_Vector3(gdextension.NewVector3WithFloat32Float32Float32(v3.MemberGetx(), v3.MemberGety(), 0))
	log.Info("Vector3: Substract (x,y,0)",
		zap.Float32("x", v3.MemberGetx()),
		zap.Float32("y", v3.MemberGety()),
		zap.Float32("z", v3.MemberGetz()),
	)
	v3 = v3.Normalized()
	log.Info("Vector3: Normalized",
		zap.Float32("x", v3.MemberGetx()),
		zap.Float32("y", v3.MemberGety()),
		zap.Float32("z", v3.MemberGetz()),
	)
	equal := v3.Equal_Vector3(gdextension.NewVector3WithFloat32Float32Float32(0, 0, 1))
	log.Info("Vector3: Equality Check",
		zap.Float32("x", v3.MemberGetx()),
		zap.Float32("y", v3.MemberGety()),
		zap.Float32("z", v3.MemberGetz()),
		zap.Bool("equal", equal),
	)
	input := gdextension.GetInputSingleton()
	uiRight := gdextension.NewStringNameWithUtf8Chars("ui_right")
	defer uiRight.Destroy()
	input.IsActionPressed(uiRight, true)
}

func (e *Example) V_Input(refEvent gdextension.Ref) {
	event := refEvent.Ptr()
	if event == nil {
		log.Warn("Example.V_Input: null refEvent parameter")
		return
	}
	keyEvent, ok := (*event).CastTo((gdextension.InputEventKey)(nil)).(gdextension.InputEventKey)
	if !ok {
		log.Error("Example.V_Input: unable to cast event to InputEventKey")
		return
	}
	keyLabel := keyEvent.GetKeyLabel()
	v := int64(keyEvent.GetUnicode())
	e.EmitCustomSignal(fmt.Sprintf("_input: %d", keyLabel), v);
}