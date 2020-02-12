package main

// #cgo CFLAGS: -DX86=1 -g -fPIC -std=c99 -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/gdnative
// #include <godot/gdnative_interface.h>
import "C"

import (
	"fmt"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/gdnative"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// ExampleRef implements GDClass evidence
var _ gdnative.GDClass = new(ExampleRef)

type ExampleRef struct {
	gdnative.RefCounted
}

func (c *ExampleRef) GetExtensionClass() gdnative.TypeName {
	return (gdnative.TypeName)("Example")
}

func (c *ExampleRef) GetExtensionParentClass() gdnative.TypeName {
	return (gdnative.TypeName)("RefCounted")
}

type ExampleEnum int64

const (
	ExampleFirst       ExampleEnum = 0
	AnswerToEverything ExampleEnum = 42
)

const (
	EXAMPLE_ENUM_CONSTANT_WITHOUT_ENUM = 314
)

// Example implements GDClass evidence
var _ gdnative.GDClass = new(Example)

type Example struct {
	gdnative.Control
	customPosition gdnative.Vector2
}

func (c *Example) GetExtensionClass() gdnative.TypeName {
	return (gdnative.TypeName)("Example")
}

func (c *Example) GetExtensionParentClass() gdnative.TypeName {
	return (gdnative.TypeName)("Control")
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

func (e *Example) ReturnSomethingConst() (gdnative.Viewport, error) {
	println("  Return something const called.")
	if e.IsInsideTree() {
		result := e.GetViewport()
		return result, nil
	}
	return gdnative.Viewport{}, fmt.Errorf("unable to get viewport")
}

// func (e *Example) ReturnExtendedRef() *ExampleRef {
// 	return NewExampleRef()
// }

func (e *Example) GetV4() gdnative.Vector4 {
	v4 := gdnative.NewVector4WithFloat32Float32Float32Float32(1.2, 3.4, 5.6, 7.8)
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

func (e *Example) TestArray() gdnative.Array {
	arr := gdnative.NewArray()

	arr.Resize(2)
	arr.Insert(0, gdnative.NewVariantInt64(1))
	arr.Insert(1, gdnative.NewVariantInt64(2))

	return arr
}

func (e *Example) TestDictionary() gdnative.Dictionary {
	dict := gdnative.NewDictionary()
	v := gdnative.NewVariantDictionary(dict)

	v.SetNamed(gdnative.NewStringNameWithLatin1Chars("hello"), gdnative.NewVariantString(gdnative.NewStringWithLatin1Chars("world")))
	v.SetNamed(gdnative.NewStringNameWithLatin1Chars("foo"), gdnative.NewVariantString(gdnative.NewStringWithLatin1Chars("bar")))

	return dict
}

func (e *Example) SetCustomPosition(pos gdnative.Vector2) {
	e.customPosition = pos
}

func (e *Example) GetCustomPosition() gdnative.Vector2 {
	return e.customPosition
}

func (e *Example) EmitCustomSignal(name string, value int64) {
	e.EmitSignal(gdnative.NewStringNameWithLatin1Chars("custom_signal"))
	log.Debug("EmitCustomSignal called")
}

func RegisterExampleTypes() {
	log.Debug("RegisterExampleTypes called")
	gdnative.ClassDBRegisterClass(&Example{}, func(t gdnative.GDClass) {
		gdnative.ClassDBBindMethodStatic(t, "TestStatic", "test_static", []string{"a", "b"}, nil)
		gdnative.ClassDBBindMethodStatic(t, "TestStatic2", "test_static2", nil, nil)
		gdnative.ClassDBBindMethod(t, "SimpleFunc", "simple_func", nil, nil)
		gdnative.ClassDBBindMethod(t, "SimpleConstFunc", "simple_const_func", []string{"a"}, nil)
		gdnative.ClassDBBindMethod(t, "ReturnSomething", "return_something", []string{"base", "f32", "f64", "i", "i8", "i16", "i32", "i64"}, nil)
		gdnative.ClassDBBindMethod(t, "ReturnSomethingConst", "return_something_const", nil, nil)
		// gdnative.ClassDBBindMethod(t, "ReturnExtendedRef", "return_extended_ref", nil, nil)
		gdnative.ClassDBBindMethod(t, "GetV4", "get_v4", nil, nil)
		defArgA := gdnative.NewVariantInt64(100)
		defArgB := gdnative.NewVariantInt64(200)
		gdnative.ClassDBBindMethod(t, "DefArgs", "def_args", []string{"a", "b"}, []*gdnative.Variant{&defArgA, &defArgB})
		gdnative.ClassDBBindMethod(t, "TestArray", "test_array", nil, nil)
		gdnative.ClassDBBindMethod(t, "TestDictionary", "test_dictionary", nil, nil)

		gdnative.ClassDBAddPropertyGroup(t, "Test group", "group_")
		gdnative.ClassDBAddPropertySubgroup(t, "Test subgroup", "group_subgroup_")

		gdnative.ClassDBBindMethod(t, "GetCustomPosition", "get_custom_position", nil, nil)
		gdnative.ClassDBBindMethod(t, "SetCustomPosition", "set_custom_position", []string{"position"}, nil)

		gdnative.ClassDBAddProperty(t, gdnative.GDNATIVE_VARIANT_TYPE_VECTOR2, (gdnative.PropertyName)("group_subgroup_custom_position"), "SetCustomPosition", "GetCustomPosition")

		// Signals.
		gdnative.ClassDBAddSignal(t, "custom_signal",
			gdnative.SignalParam{
				Type: gdnative.GDNATIVE_VARIANT_TYPE_STRING,
				Name: "name"},
			gdnative.SignalParam{
				Type: gdnative.GDNATIVE_VARIANT_TYPE_INT,
				Name: "value",
			})
		gdnative.ClassDBBindMethod(t, "EmitCustomSignal", "emit_custom_signal", []string{"name", "value"}, nil)

		// constants
		gdnative.ClassDBBindEnumConstant(t, "ExampleEnum", "FIRST", int(ExampleFirst))
		gdnative.ClassDBBindEnumConstant(t, "ExampleEnum", "ANSWER_TO_EVERYTHING", int(AnswerToEverything))

		gdnative.ClassDBBindConstant(t, "CONSTANT_WITHOUT_ENUM", int(EXAMPLE_ENUM_CONSTANT_WITHOUT_ENUM))
	})
}

func UnregisterExampleTypes() {
	log.Debug("UnregisterExampleTypes called")
}

//export ExampleLibraryInit
func ExampleLibraryInit(p_interface *C.GDNativeInterface, p_library C.GDNativeExtensionClassLibraryPtr, r_initialization *C.GDNativeInitialization) bool {
	log.Debug("ExampleLibraryInit called")
	initObj := gdnative.NewInitObject(
		(*gdnative.GDNativeInterface)(unsafe.Pointer(p_interface)),
		(gdnative.GDNativeExtensionClassLibraryPtr)(p_library),
		(*gdnative.GDNativeInitialization)(unsafe.Pointer(r_initialization)),
	)

	initObj.RegisterSceneInitializer(RegisterExampleTypes)
	initObj.RegisterSceneTerminator(UnregisterExampleTypes)

	return initObj.Init()
}

func main() {

}
