package pkg

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/godot-go/godot-go/pkg/gdextension"
	"github.com/godot-go/godot-go/pkg/gdextensionffi"
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
	customPosition   gdextension.Vector2
	propertyFromList gdextension.Vector3
	dprop [3]gdextension.Vector2
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
	return fmt.Sprintf("1. %s, 2. %f, 3. %f, 4. %d, 5. %d, 6. %d, 7. %d, 8. %d", base+"42", f32, f64, i, i8, i16, i32, i64)
}

func (e *Example) ReturnSomethingConst() gdextension.Viewport {
	println("  Return something const called.")
	if !e.IsInsideTree() {
		return nil
	}
	result := e.GetViewport()
	if result == nil {
		println("null viewport encountered")
		return nil
	}
	fmt.Printf("viewport instance id: %v\n", result.GetInstanceId())
	return result
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
	ret := p_a + p_b
	log.Info("DefArgs called", zap.Int32("sum", ret))
	return ret
}

func (e *Example) TestArray() gdextension.Array {
	arr := gdextension.NewArray()
	arr.Insert(0, gdextension.NewVariantInt64(1))
	arr.Insert(1, gdextension.NewVariantInt64(2))
	v := gdextension.NewVariantArray(arr)
	v0, err := v.GetIndexed(0)
	if err != nil {
		log.Panic("error getting index 0: %w",
			zap.Error(err),
		)
	}
	gdV0 := v0.ToString()
	v1, err := v.GetIndexed(1)
	if err != nil {
		log.Panic("error getting index 0: %w",
			zap.Error(err),
		)
	}
	gdV1 := v1.ToString()
	log.Info("arr size",
		zap.Any("size", arr.Size()),
		zap.String("v[0]", gdV0.ToUtf8()),
		zap.String("v[1]", gdV1.ToUtf8()),
	)
	if v0.GetType() != gdextensionffi.GDEXTENSION_VARIANT_TYPE_INT {
		log.Panic("array value at index 0 is not a INT variant type",
			zap.Any("actual", v0.GetType()),
			zap.Any("expected", gdextensionffi.GDEXTENSION_VARIANT_TYPE_INT),
		)
	}
	if v0.ToInt64() != 1 {
		log.Panic("array value at index 0 is not equal",
			zap.Int64("actual", v0.ToInt64()),
			zap.Int64("expected", 1),
		)
	}
	if v1.GetType() != gdextensionffi.GDEXTENSION_VARIANT_TYPE_INT {
		log.Panic("array value at index 1 is not a INT variant type",
			zap.Any("actual", v1.GetType()),
			zap.Any("expected", gdextensionffi.GDEXTENSION_VARIANT_TYPE_INT),
		)
	}
	if v1.ToInt64() != 2 {
		log.Panic("array value at index 1 is not equal",
			zap.Int64("actual", v0.ToInt64()),
			zap.Int64("expected", 2),
		)
	}
	ret := v.ToArray()
	pr := ret.PickRandom()
	log.Info("pick random", zap.Int64("val", pr.ToInt64()))
	return ret
}

func (e *Example) TestDictionary() gdextension.Dictionary {
	dict := gdextension.NewDictionary()
	world := gdextension.NewVariantGoString("world")
	defer world.Destroy()
	bar := gdextension.NewVariantGoString("bar")
	defer bar.Destroy()
	dict.SetKeyed("hello", world)
	dict.SetKeyed("foo", bar)
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
	customSignal := gdextension.NewStringNameWithLatin1Chars("custom_signal")
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
	n := gdextension.ObjectCastTo(e, "Node").(gdextension.Node)
	if n != nil {
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
	s := gdextension.NewStringWithUtf8Chars("A")
	defer s.Destroy()
	sB := gdextension.NewStringWithUtf8Chars("B")
	defer sB.Destroy()
	sC := gdextension.NewStringWithUtf8Chars("C")
	defer sC.Destroy()
	sD := gdextension.NewStringWithUtf32Char(0x010E)
	defer sD.Destroy()
	sE := gdextension.NewStringWithUtf8Chars("E")
	defer sE.Destroy()
	s = s.Add_String(sB)
	s = s.Add_String(sC)
	s = s.Add_String(sD)
	s = s.Add_String(sE)
	return s.ToUtf32()
}

func (e *Example) VarargsFunc(args ...gdextension.Variant) gdextension.Variant {
	ret := gdextension.NewVariantInt64(int64(len(args)))
	return ret
}

// func (e *Example) V_GetPropertyList(props []gdextension.PropertyInfo) {
// 	props = append(props, gdextension.NewPropertyInfo(gdextentionffi.GDEXTENSION_VARIANT_TYPE_VECTOR3, "property_from_list"))
// 	for i := 0; i < 3; i++ {
// 		props = append(props, gdextension.NewPropertyInfo(gdextentionffi.GDEXTENSION_VARIANT_TYPE_VECTOR2, fmt.Sprintf("dproperty_%d", i)))
// 	}
// }

func (e *Example) V_PropertyCanRevert(p_name *gdextension.StringName) bool {
	gdSn := gdextension.NewStringNameWithLatin1Chars("property_from_list")
	vec3 := gdextension.NewVector3WithFloat32Float32Float32(42, 42, 42)
	return p_name.Equal_StringName(gdSn) && !e.propertyFromList.Equal_Vector3(vec3)
}

func (e *Example) V_Set(name string, value gdextension.Variant) bool {
	if strings.HasPrefix(name, "dproperty") {
		tokens := strings.SplitN(name, "_", 2)
		if len(tokens) != 2 {
			log.Error("invalid property name",
				zap.String("name", name),
			)
		}
		index, err := strconv.Atoi(tokens[1])
		if err != nil {
			log.Error("invalid index parsed from property name",
				zap.String("name", name),
			)
		}
		e.dprop[index] = value.ToVector2()
		return true;
	}
	if name == "property_from_list" {
		e.propertyFromList = value.ToVector3()
		return true;
	}
	return false;
}

func (e *Example) V_Get(name string) (gdextension.Variant, bool) {
	switch {
	case strings.HasPrefix(name, "dproperty"):
		tokens := strings.SplitN(name, "_", 2)
		if len(tokens) != 2 {
			log.Error("invalid property name",
				zap.String("name", name),
			)
		}
		index, err := strconv.Atoi(tokens[1])
		if err != nil {
			log.Error("invalid index parsed from property name",
				zap.String("name", name),
			)
		}
		vec2 := e.dprop[index]
		log.Info("--> return vec2",
			zap.Float32("x", vec2.MemberGetx()),
			zap.Float32("y", vec2.MemberGety()),
		)
		v := gdextension.NewVariantVector2(vec2)
		return v, true
	case name == "property_from_list":
		v := gdextension.NewVariantVector3(e.propertyFromList)
		return v, true
	}
	return gdextension.Variant{}, false
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
	uiRight := gdextension.NewStringNameWithLatin1Chars("ui_right")
	defer uiRight.Destroy()
	input.IsActionPressed(uiRight, true)
}

func (e *Example) V_Input(refEvent gdextension.Ref) {
	defer refEvent.Unref()
	event := refEvent.Ptr()
	if event == nil {
		log.Warn("Example.V_Input: null refEvent parameter")
		return
	}
	keyEvent := event.(gdextension.InputEventKey)
	gdStringKeyLabel := keyEvent.AsTextKeyLabel()
	keyLabel := gdStringKeyLabel.ToUtf8()
	v := int64(keyEvent.GetUnicode())
	e.EmitCustomSignal(fmt.Sprintf("_input: %s", keyLabel), v)
}
