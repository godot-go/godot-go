package pkg

/*
#cgo CFLAGS: -I${SRCDIR} -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
#include "example.h"
*/
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/gdextensionffi"
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
}

func (c *Example) GetClassName() string {
	return "Example"
}

func (c *Example) GetParentClassName() string {
	return "Control"
}

func (e *Example) TestStatic(p_a, p_b int32) int32 {
	return p_a + p_b
}

func (e *Example) TestStatic2() {
	println("  void static")
}

func (e *Example) SimpleFunc() {
	println("  Simple func called.")
}

func (e *Example) SimpleConstFunc(a int64) {
	fmt.Printf("  Simple const func called %d.\n", a)
}

func (e *Example) ReturnSomething(base string, f32 float32, f64 float64,
	i int, i8 int8, i16 int16, i32 int32, i64 int64) string {
	println("  Return something called (8 values cancatenated as a string).")
	return fmt.Sprintf("(1. %s, 2. %f, 3. %f, 4. %d, 5. %d, 6. %d, 7. %d, 8. %d)", base, f32, f64, i, i8, i16, i32, i64)
}

func (e *Example) ReturnSomethingConst() (gdextension.Viewport, error) {
	println("  Return something const called.")
	if e.IsInsideTree() {
		result := e.GetViewport()

		if result == nil {
			println("null viewport encountered")
			return nil, fmt.Errorf("null viewport")
		}

		fmt.Printf("viewport instance id: %v\n", result.GetInstanceId())



		return result, nil
	}
	return nil, fmt.Errorf("unable to get viewport")
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

	v.SetNamed(gdextension.NewStringNameWithLatin1Chars("hello"), gdextension.NewVariantString(gdextension.NewStringWithLatin1Chars("world")))
	v.SetNamed(gdextension.NewStringNameWithLatin1Chars("foo"), gdextension.NewVariantString(gdextension.NewStringWithLatin1Chars("bar")))

	return dict
}

func (e *Example) SetCustomPosition(pos gdextension.Vector2) {
	e.customPosition = pos
}

func (e *Example) GetCustomPosition() gdextension.Vector2 {
	return e.customPosition
}

func (e *Example) EmitCustomSignal(name string, value int64) {
	e.EmitSignal(
		gdextension.NewStringNameWithLatin1Chars("custom_signal"),
		gdextension.NewVariantString(gdextension.NewStringWithLatin1Chars(name)),
		gdextension.NewVariantInt64(value),
	)
	log.Debug("EmitCustomSignal called",
		zap.String("name", name),
		zap.Int64("value", value),
	)
}

func (e *Example) TestCastTo() {
	n, ok := e.CastTo("Node").(gdextension.Node)
	if !ok {
		log.Panic("failed to cast to cast Example to Node")
	}
	log.Debug("TestCastTo called", zap.Any("class", n.GetClassName()))
}

//export Example_Ready
func Example_Ready(inst unsafe.Pointer) {
	log.Info("Example_Ready called")

	input := gdextension.GetInputSingleton()

	uiRight := gdextension.NewStringNameWithLatin1Chars("ui_right")

	input.IsActionPressed(uiRight, true)
}

func RegisterExampleTypes() {
	log.Debug("RegisterExampleTypes called")
	gdextension.ClassDBRegisterClass(&Example{}, func(t gdextension.GDClass) {
		gdextension.ClassDBBindMethodVirtual(t, "_ready", (gdextensionffi.GDExtensionClassCallVirtual)(C.cgo_callback_example_ready))
		gdextension.ClassDBBindMethodStatic(t, "TestStatic", "test_static", []string{"a", "b"}, nil)
		gdextension.ClassDBBindMethodStatic(t, "TestStatic2", "test_static2", nil, nil)
		gdextension.ClassDBBindMethod(t, "SimpleFunc", "simple_func", nil, nil)
		gdextension.ClassDBBindMethod(t, "SimpleConstFunc", "simple_const_func", []string{"a"}, nil)
		gdextension.ClassDBBindMethod(t, "ReturnSomething", "return_something", []string{"base", "f32", "f64", "i", "i8", "i16", "i32", "i64"}, nil)
		gdextension.ClassDBBindMethod(t, "ReturnSomethingConst", "return_something_const", nil, nil)
		gdextension.ClassDBBindMethod(t, "GetV4", "get_v4", nil, nil)
		defArgA := gdextension.NewVariantInt64(100)
		defArgB := gdextension.NewVariantInt64(200)
		gdextension.ClassDBBindMethod(t, "DefArgs", "def_args", []string{"a", "b"}, []*gdextension.Variant{&defArgA, &defArgB})
		gdextension.ClassDBBindMethod(t, "TestArray", "test_array", nil, nil)
		gdextension.ClassDBBindMethod(t, "TestDictionary", "test_dictionary", nil, nil)

		gdextension.ClassDBAddPropertyGroup(t, "Test group", "group_")
		gdextension.ClassDBAddPropertySubgroup(t, "Test subgroup", "group_subgroup_")

		gdextension.ClassDBBindMethod(t, "GetCustomPosition", "get_custom_position", nil, nil)
		gdextension.ClassDBBindMethod(t, "SetCustomPosition", "set_custom_position", []string{"position"}, nil)

		gdextension.ClassDBAddProperty(t, gdextensionffi.GDEXTENSION_VARIANT_TYPE_VECTOR2, "group_subgroup_custom_position", "SetCustomPosition", "GetCustomPosition")

		// Signals.
		gdextension.ClassDBAddSignal(t, "custom_signal",
			gdextension.SignalParam{
				Type: gdextensionffi.GDEXTENSION_VARIANT_TYPE_STRING,
				Name: "name"},
			gdextension.SignalParam{
				Type: gdextensionffi.GDEXTENSION_VARIANT_TYPE_INT,
				Name: "value",
			})
		gdextension.ClassDBBindMethod(t, "EmitCustomSignal", "emit_custom_signal", []string{"name", "value"}, nil)

		// constants
		gdextension.ClassDBBindEnumConstant(t, "ExampleEnum", "FIRST", int(ExampleFirst))
		gdextension.ClassDBBindEnumConstant(t, "ExampleEnum", "ANSWER_TO_EVERYTHING", int(AnswerToEverything))
		gdextension.ClassDBBindConstant(t, "CONSTANT_WITHOUT_ENUM", int(EXAMPLE_ENUM_CONSTANT_WITHOUT_ENUM))

		// others
		gdextension.ClassDBBindMethod(t, "TestCastTo", "test_cast_to", nil, nil)
	})
}

func UnregisterExampleTypes() {
	log.Debug("UnregisterExampleTypes called")
}

//export TestDemoInit
func TestDemoInit(p_interface unsafe.Pointer, p_library unsafe.Pointer, r_initialization unsafe.Pointer) bool {
	log.Debug("ExampleLibraryInit called")
	initObj := gdextension.NewInitObject(
		(*gdextensionffi.GDExtensionInterface)(unsafe.Pointer(p_interface)),
		(gdextensionffi.GDExtensionClassLibraryPtr)(p_library),
		(*gdextensionffi.GDExtensionInitialization)(unsafe.Pointer(r_initialization)),
	)

	initObj.RegisterSceneInitializer(RegisterExampleTypes)
	initObj.RegisterSceneTerminator(UnregisterExampleTypes)

	return initObj.Init()
}
