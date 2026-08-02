package pkg

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
	. "github.com/godot-go/godot-go/pkg/gdutilfunc"
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

type ExampleBitfieldFlag int64

const (
	FlagOne ExampleBitfieldFlag = 1
	FlagTwo ExampleBitfieldFlag = 1 << 1
)

// Example implements GDClass evidence
var _ GDClass = (*Example)(nil)

type Example struct {
	ControlImpl
	customPosition   Vector2
	propertyFromList Vector3
	dprop            [3]Vector2
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

// func (e *Example) CustomRefFunc(pRef *ExampleRef) int32 {
// 	if pRef.IsValid() {
// 		return pRef.Ptr().(*ExampleRef).GetId()
// 	}
// 	return -1
// }

func (e *Example) ReturnSomething(base string, f32 float32, f64 float64,
	i int, i8 int8, i16 int16, i32 int32, i64 int64) string {
	println("  Return something called (8 values cancatenated as a string).")
	return fmt.Sprintf("1. %s, 2. %f, 3. %f, 4. %d, 5. %d, 6. %d, 7. %d, 8. %d", base+"42", f32, f64, i, i8, i16, i32, i64)
}

func (e *Example) ReturnSomethingConst() Viewport {
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

func (e *Example) ReturnEmptyRef() *ExampleRef {
	var ref ExampleRef
	return &ref
}

func (e *Example) GetV4() Vector4 {
	v4 := NewVector4WithFloat32Float32Float32Float32(1.2, 3.4, 5.6, 7.8)
	log.Debug("vector4 members",
		zap.Any("x", v4.MemberGetx()),
		zap.Any("y", v4.MemberGety()),
		zap.Any("z", v4.MemberGetz()),
		zap.Any("w", v4.MemberGetw()),
	)
	return v4
}

func (e *Example) TestNodeArgument(node *Example) *Example {
	log.Debug("example instances should be the same",
		zap.Any("reciever_owner", fmt.Sprintf("%p", e.Owner)),
		zap.Any("reciever_id", e.GetInstanceId()),
		zap.Any("arg_owner", fmt.Sprintf("%p", node.Owner)),
		// zap.Any("arg_id", node.GetInstanceId()),
	)
	return node
}

func (e *Example) DefArgs(p_a, p_b int32) int32 {
	ret := p_a + p_b
	log.Info("DefArgs called", zap.Int32("sum", ret))
	return ret
}

func (e *Example) TestArray() Array {
	arr := NewArray()
	arr.Insert(0, NewVariantInt64(1))
	arr.Insert(1, NewVariantInt64(2))
	v := NewVariantArray(arr)
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
	if v0.GetType() != GDEXTENSION_VARIANT_TYPE_INT {
		log.Panic("array value at index 0 is not a INT variant type",
			zap.Any("actual", v0.GetType()),
			zap.Any("expected", GDEXTENSION_VARIANT_TYPE_INT),
		)
	}
	if v0.ToInt64() != 1 {
		log.Panic("array value at index 0 is not equal",
			zap.Int64("actual", v0.ToInt64()),
			zap.Int64("expected", 1),
		)
	}
	if v1.GetType() != GDEXTENSION_VARIANT_TYPE_INT {
		log.Panic("array value at index 1 is not a INT variant type",
			zap.Any("actual", v1.GetType()),
			zap.Any("expected", GDEXTENSION_VARIANT_TYPE_INT),
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

func (e *Example) TestTArrayArg(arr PackedInt64Array) int64 {
	sum := int64(0)
	sz := arr.Size()
	for i := range sz {
		sum += arr.GetIndexed(i)
	}
	return sum
}

func (e *Example) TestTArray() Array {
	parr := NewPackedVector2Array()
	parr.Resize(2)
	parr.Set(0, NewVector2WithFloat32Float32(1.0, 2.0))
	parr.Set(1, NewVector2WithFloat32Float32(2.0, 3.0))
	arr := NewArrayWithPackedVector2Array(parr)
	return arr
}

func (e *Example) TestDictionary() Dictionary {
	dict := NewDictionary()
	world := NewVariantGoString("world")
	// defer world.Destroy()
	bar := NewVariantGoString("bar")
	// defer bar.Destroy()
	dict.SetKeyed("hello", world)
	dict.SetKeyed("foo", bar)
	return dict
}

func (e *Example) SetCustomPosition(pos Vector2) {
	e.customPosition = pos
}

func (e *Example) GetCustomPosition() Vector2 {
	return e.customPosition
}

func (e *Example) SetPropertyFromList(v Vector3) {
	e.propertyFromList = v
}

func (e *Example) GetPropertyFromList() Vector3 {
	return e.propertyFromList
}

func (e *Example) EmitCustomSignal(name string, value int64) {
	customSignal := NewStringNameWithLatin1Chars("custom_signal")
	defer customSignal.Destroy()
	snName := NewStringWithLatin1Chars(name)
	defer snName.Destroy()
	log.Info("EmitCustomSignal called",
		zap.String("name", name),
		zap.Int64("value", value),
	)
	arg0 := NewVariantString(snName)
	defer arg0.Destroy()
	arg1 := NewVariantInt64(value)
	defer arg1.Destroy()
	e.EmitSignal(
		customSignal,
		arg0,
		arg1,
	)
}

// TODO: dig into why casting is important
// func (e *Example) TestCastTo() {
// 	n := ObjectCastTo(e, "Node").(Node)
// 	if n == nil {
// 		log.Panic("failed to cast to cast Example to Node")
// 	}
// 	log.Debug("TestCastTo called", zap.Any("class", n.GetClassName()))
// }

func (e *Example) TestStatic(p_a, p_b int32) int32 {
	return p_a + p_b
}

func (e *Example) TestStatic2() {
	println("  void static")
}

func (e *Example) TestStringOps() string {
	s := NewStringWithUtf8Chars("A")
	defer s.Destroy()
	sB := NewStringWithUtf8Chars("B")
	defer sB.Destroy()
	sC := NewStringWithUtf8Chars("C")
	defer sC.Destroy()
	sD := NewStringWithUtf32Char(0x010E)
	defer sD.Destroy()
	sE := NewStringWithUtf8Chars("E")
	defer sE.Destroy()
	s = s.Add_String(sB)
	s = s.Add_String(sC)
	s = s.Add_String(sD)
	s = s.Add_String(sE)
	return s.ToUtf32()
}

func (e *Example) VarargsFunc(args ...Variant) Variant {
	ret := NewVariantInt64(int64(len(args)))
	return ret
}

func (e *Example) VarargsFuncVoid(args ...Variant) {
	e.EmitCustomSignal("varargs_func_void", int64(len(args)+1))
}

func (e *Example) VarargsFuncNv(args ...Variant) int {
	return 42 + len(args)
}

func (e *Example) V_PropertyCanRevert(p_name StringName) bool {
	gdSn := NewStringNameWithLatin1Chars("property_from_list")
	defer gdSn.Destroy()
	vec3 := NewVector3WithFloat32Float32Float32(42, 42, 42)
	return p_name.Equal_StringName(gdSn) && !e.propertyFromList.Equal_Vector3(vec3)
}

func (e *Example) V_Set(name string, value Variant) bool {
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
		return true
	}
	if name == "property_from_list" {
		e.propertyFromList = value.ToVector3()
		return true
	}
	return false
}

func (e *Example) V_Get(name string) (Variant, bool) {
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
		v := NewVariantVector2(vec2)
		return v, true
	case name == "property_from_list":
		v := NewVariantVector3(e.propertyFromList)
		return v, true
	}
	return Variant{}, false
}

func (e *Example) V_Ready() {
	log.Info("Example_Ready called",
		zap.String("inst", fmt.Sprintf("%p", e)),
	)
	// vector math tests
	v3 := NewVector3WithFloat32Float32Float32(1.1, 2.2, 3.3)
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
	v3 = v3.Add_Vector3(NewVector3WithFloat32Float32Float32(10, 20, 30))
	log.Info("Vector3: Add (1,2,3)",
		zap.Float32("x", v3.MemberGetx()),
		zap.Float32("y", v3.MemberGety()),
		zap.Float32("z", v3.MemberGetz()),
	)
	v3 = v3.Multiply_Vector3(NewVector3WithFloat32Float32Float32(5, 10, 15))
	log.Info("Vector3: Multiply (5,10,15)",
		zap.Float32("x", v3.MemberGetx()),
		zap.Float32("y", v3.MemberGety()),
		zap.Float32("z", v3.MemberGetz()),
	)
	v3 = v3.Subtract_Vector3(NewVector3WithFloat32Float32Float32(v3.MemberGetx(), v3.MemberGety(), 0))
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
	equal := v3.Equal_Vector3(NewVector3WithFloat32Float32Float32(0, 0, 1))
	log.Info("Vector3: Equality Check",
		zap.Float32("x", v3.MemberGetx()),
		zap.Float32("y", v3.MemberGety()),
		zap.Float32("z", v3.MemberGetz()),
		zap.Bool("equal", equal),
	)
	input := GetInputSingleton()
	uiRight := NewStringNameWithUtf8Chars("ui_right")
	defer uiRight.Destroy()
	input.IsActionPressed(uiRight, true)
	log.Info("NearestPo2(1025)",
		zap.Int64("result", NearestPo2(1025)),
	)
	Randomize()
	log.Info("RandiRange(0, 50)",
		zap.Int64("result", RandiRange(0, 50)),
	)
}

func (e *Example) V_Input(refEvent RefInputEvent) {
	event := refEvent.Ptr()
	if event == nil {
		log.Warn("Example.V_Input: null refEvent parameter")
		return
	}
	keyEvent, ok := ObjectCastTo(event, "InputEventKey").(InputEventKey)
	if !ok {
		log.Warn("event not InputEventKey")
		return
	}
	gdStringKeyLabel := keyEvent.AsTextKeyLabel()
	defer gdStringKeyLabel.Destroy()
	keyLabel := gdStringKeyLabel.ToUtf8()
	v := int64(keyEvent.GetUnicode())
	e.EmitCustomSignal(fmt.Sprintf("_input: %s", keyLabel), v)
}

func (e *Example) TestSetPositionAndSize(pos, size Vector2) {
	e.SetPosition(pos, true)
	e.SetSize(size, true)
}

func (e *Example) TestGetChildNode(nodePath string) Node {
	npStr := NewStringWithUtf8Chars(nodePath)
	defer npStr.Destroy()
	np := NewNodePathWithString(npStr)
	defer np.Destroy()
	names := np.GetConcatenatedNames()
	defer names.Destroy()
	log.Info("node path", zap.String("node_path", names.ToUtf8()))
	return e.GetNode(np)
}

func (e *Example) ImageRefFunc(refImage RefImage) string {
	if refImage != nil && refImage.IsValid() {
		return "valid"
	} else {
		return "invalid"
	}
}

// PassThroughRefImage receives a Ref<Image> (decode) and returns it (encode),
// exercising both Ref boundary directions.
func (e *Example) PassThroughRefImage(refImage RefImage) RefImage {
	if refImage == nil || !refImage.IsValid() {
		return nil
	}
	return refImage
}

// ReturnNilRefImage returns a nil Ref<Image> to Godot.
func (e *Example) ReturnNilRefImage() RefImage {
	return nil
}

// TestRefLifecycle exercises RefBase copy/unref semantics on a RefCounted
// object passed from Godot: copy (+1), unref (-1), double-unref safe.
// Returns 1 on success or a negative error code.
func (e *Example) TestRefLifecycle(ref RefRefCounted) int32 {
	if ref == nil || !ref.IsValid() {
		return -1
	}
	before := ref.Ptr().GetReferenceCount()
	dst := NewRef[RefCounted](nil)
	dst.Ref(ref)
	mid := ref.Ptr().GetReferenceCount()
	if mid != before+1 {
		return -100
	}
	dst.Unref()
	dst.Unref() // double-unref must be a safe no-op
	after := ref.Ptr().GetReferenceCount()
	if after != before {
		return -200
	}
	return 1
}

// TestFinalizerRelease drops a copied Ref and forces GC, verifying the
// finalizer releases the Go-held reference.
// Returns 1 on success or a negative error code.
func (e *Example) TestFinalizerRelease(ref RefRefCounted) int32 {
	if ref == nil || !ref.IsValid() {
		return -1
	}
	before := ref.Ptr().GetReferenceCount()
	func() {
		r := NewRef[RefCounted](nil)
		r.Ref(ref) // copy: +1 and installs finalizer
		mid := r.Ptr().GetReferenceCount()
		if mid != before+1 {
			return
		}
	}()
	for i := 0; i < 4; i++ {
		runtime.GC()
		runtime.Gosched()
	}
	after := ref.Ptr().GetReferenceCount()
	if after != before {
		return -400
	}
	return 1
}

// TestStringArgEcho decodes a Godot String argument and returns a fresh
// String built from its content (avoids aliasing the caller-owned arg).
func (e *Example) TestStringArgEcho(p_str String) String {
	return NewStringWithUtf8Chars(p_str.ToUtf8())
}

// TestStringNameArgEcho decodes a Godot StringName argument and returns its
// content as a fresh String.
func (e *Example) TestStringNameArgEcho(p_name StringName) String {
	return NewStringWithUtf8Chars(p_name.ToUtf8())
}

// TestNodePathArgEcho decodes a Godot NodePath argument and returns its
// content as a fresh String.
func (e *Example) TestNodePathArgEcho(p_path NodePath) String {
	return NewStringWithUtf8Chars(NewStringWithNodePath(p_path).ToUtf8())
}

// TestReturnStringName returns a fresh Godot StringName.
func (e *Example) TestReturnStringName() StringName {
	return NewStringNameWithUtf8Chars("returned_name")
}

// TestReturnNodePath returns a fresh Godot NodePath.
func (e *Example) TestReturnNodePath() NodePath {
	return NewNodePathWithString(NewStringWithUtf8Chars("root/child"))
}

// TestScalarEcho decodes bool/int64/float64/string scalars and returns them
// formatted into a Go string.
func (e *Example) TestScalarEcho(p_bool bool, p_i64 int64, p_f64 float64, p_str string) string {
	return fmt.Sprintf("%v:%d:%.1f:%s", p_bool, p_i64, p_f64, p_str)
}

// TestUint64Echo decodes a uint64 argument and returns it, preserving its
// bit pattern.
func (e *Example) TestUint64Echo(p_u64 uint64) uint64 {
	return p_u64
}

// TestReturnInt8 returns a narrow numeric (exercises varcall widening).
func (e *Example) TestReturnInt8() int8 {
	return 127
}

// TestReturnUint16 returns a narrow numeric (exercises varcall widening).
func (e *Example) TestReturnUint16() uint16 {
	return 65535
}

// TestReturnFloat32 returns a narrow float (exercises varcall widening).
func (e *Example) TestReturnFloat32() float32 {
	return 1.5
}

// TestGoStringRepeat returns a Go string; called repeatedly to exercise the
// varcall Go-string return lifecycle.
func (e *Example) TestGoStringRepeat() string {
	return "repeat_me_string_for_leak_check"
}

func (e *Example) TestEchoVector2i(p Vector2i) Vector2i { return p }
func (e *Example) TestEchoVector3(p Vector3) Vector3     { return p }
func (e *Example) TestEchoVector3i(p Vector3i) Vector3i  { return p }
func (e *Example) TestEchoRect2(p Rect2) Rect2           { return p }
func (e *Example) TestEchoRect2i(p Rect2i) Rect2i        { return p }
func (e *Example) TestEchoTransform2D(p Transform2D) Transform2D { return p }
func (e *Example) TestEchoVector4i(p Vector4i) Vector4i  { return p }
func (e *Example) TestEchoPlane(p Plane) Plane            { return p }
func (e *Example) TestEchoQuaternion(p Quaternion) Quaternion { return p }
func (e *Example) TestEchoAABB(p AABB) AABB              { return p }
func (e *Example) TestEchoBasis(p Basis) Basis            { return p }
func (e *Example) TestEchoTransform3D(p Transform3D) Transform3D { return p }
func (e *Example) TestEchoProjection(p Projection) Projection { return p }

func (e *Example) TestCharacterBody2D(body CharacterBody2D) {
	if body == nil {
		log.Warn("CharacterBody2D was nil")
		return
	}
	gdStrBody := body.ToString()
	// defer gdStrBody.Destroy()
	log.Info("TestCharacterBody2D called",
		zap.String("body", gdStrBody.ToUtf8()),
	)
	motion := NewVector2WithFloat32Float32(1.0, 2.0)
	refCollision := body.MoveAndCollide(motion, true, 0.5, true)
	collision := refCollision.Ptr()
	collisionV := NewVariantGodotObject(collision.GetGodotObjectOwner())
	log.Info("collision returned",
		zap.String("collision", collisionV.Stringify()),
	)
}

func (e *Example) TestParentIsNil() Control {
	log.Info("TestParentIsNil called")
	parent := e.GetParentControl()
	return parent
}

func (e *Example) TestStrUtility() string {
	gds := Str(
		NewVariantGoString("Hello, "),
		NewVariantGoString("World"),
		NewVariantGoString("! The answer is "),
		NewVariantInt64(42),
	)
	defer gds.Destroy()
	return gds.ToUtf8()
}

func (e *Example) TestVectorOps() int32 {
	arr := NewPackedInt32Array()
	defer arr.Destroy()
	arr.PushBack(20)
	arr.PushBack(10)
	arr.PushBack(30)
	arr.PushBack(45)
	ret := int32(0)
	for i := range arr.Size() {
		ret += int32(arr.GetIndexed(i))
	}
	return ret
}

func (e *Example) TestInstanceFromIdUtility() Object {
	id := e.GetInstanceId()
	obj := InstanceFromId(int64(id))
	log.Info("InstanceFromId(e.GetInstanceId())",
		zap.Uint64("expected", id),
		zap.String("actual", obj.(*ObjectImpl).ToGoString()),
	)
	return obj
}

// TestBitfield, TODO: change argument to accept type ExampleBitfieldFlag
func (e *Example) TestBitfield(flags int64) ExampleBitfieldFlag {
	return (ExampleBitfieldFlag)(flags)
}

func (e *Example) CallableBind() {
	methodName := NewStringNameWithLatin1Chars("emit_custom_signal")
	defer methodName.Destroy()
	c := NewCallableWithObjectStringName(e, methodName)
	defer c.Destroy()
	args := NewArray()
	defer args.Destroy()
	args.Append(NewVariantGoString("bound"))
	args.Append(NewVariantInt(11))
	c.Callv(args)
}

func (e *Example) TestVariantVector2iConversion(v Variant) Vector2i {
	return v.ToVector2i()
}

func (e *Example) V_ToString() string {
	return fmt.Sprintf("[ GDExtension::Example <--> Instance ID:%d ]", e.GetInstanceId())
}

func NewExampleFromOwnerObject(owner *GodotObject) GDClass {
	obj := &Example{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

func ValidateExampleProperty(property *GDExtensionPropertyInfo) {
	gdsnName := (*StringName)(property.Name())
	// Test hiding the "mouse_filter" property from the editor.
	if gdsnName.ToUtf8() == "mouse_filter" {
		property.SetUsage(PROPERTY_USAGE_NO_EDITOR)
	}
}

func GetExamplePropertyList() ([]GDExtensionPropertyInfo, func()) {
	props := make([]GDExtensionPropertyInfo, 4)
	cleanups := make([]func(), 0, 4)

	makeProp := func(className, propName string, vt GDExtensionVariantType) GDExtensionPropertyInfo {
		cn := NewStringNameWithLatin1Chars(className)
		pn := NewStringNameWithLatin1Chars(propName)
		hs := NewStringWithUtf8Chars("")
		cleanups = append(cleanups, func() {
			cn.Destroy()
			pn.Destroy()
			hs.Destroy()
		})
		return NewGDExtensionPropertyInfoFromNames(&cn, vt, &pn, &hs)
	}

	props[0] = makeProp("Example", "property_from_list", GDEXTENSION_VARIANT_TYPE_VECTOR3)
	for i := 0; i < 3; i++ {
		props[i+1] = makeProp("Example", fmt.Sprintf("dproperty_%d", i), GDEXTENSION_VARIANT_TYPE_VECTOR2)
	}

	cleanup := func() {
		for _, c := range cleanups {
			c()
		}
	}
	return props, cleanup
}

func RegisterClassExample() {
	props, cleanup := GetExamplePropertyList()
	defer cleanup()
	ClassDBRegisterClass(NewExampleFromOwnerObject, props, ValidateExampleProperty, func(t *Example) {
		// virtuals
		// virtuals
		ClassDBBindMethodVirtual(t, "V_ToString", "to_string", nil, nil)
		ClassDBBindMethodVirtual(t, "V_Ready", "_ready", nil, nil)
		ClassDBBindMethodVirtual(t, "V_Input", "_input", []string{"event"}, nil)
		ClassDBBindMethodVirtual(t, "V_Set", "_set", []string{"name", "value"}, nil)
		ClassDBBindMethodVirtual(t, "V_Get", "_get", []string{"name"}, nil)
		ClassDBBindMethodVirtual(t, "V_PropertyCanRevert", "_property_can_revert", []string{"name"}, nil)

		ClassDBBindMethod(t, "SimpleFunc", "simple_func", nil, nil)
		ClassDBBindMethod(t, "SimpleConstFunc", "simple_const_func", []string{"a"}, nil)
		// ClassDBBindMethod(t, "CustomRefFunc", "custom_ref_func", []string{"ref"}, nil)
		ClassDBBindMethod(t, "ImageRefFunc", "image_ref_func", []string{"image"}, nil)
		ClassDBBindMethod(t, "PassThroughRefImage", "pass_through_ref_image", []string{"image"}, nil)
		ClassDBBindMethod(t, "ReturnNilRefImage", "return_nil_ref_image", nil, nil)
		ClassDBBindMethod(t, "TestRefLifecycle", "test_ref_lifecycle", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestFinalizerRelease", "test_finalizer_release", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestStringArgEcho", "test_string_arg_echo", []string{"str"}, nil)
		ClassDBBindMethod(t, "TestStringNameArgEcho", "test_string_name_arg_echo", []string{"name"}, nil)
		ClassDBBindMethod(t, "TestNodePathArgEcho", "test_node_path_arg_echo", []string{"path"}, nil)
		ClassDBBindMethod(t, "TestReturnStringName", "test_return_string_name", nil, nil)
		ClassDBBindMethod(t, "TestReturnNodePath", "test_return_node_path", nil, nil)
		ClassDBBindMethod(t, "TestScalarEcho", "test_scalar_echo", []string{"p_bool", "p_i64", "p_f64", "p_str"}, nil)
		ClassDBBindMethod(t, "TestUint64Echo", "test_uint64_echo", []string{"u64"}, nil)
		ClassDBBindMethod(t, "TestReturnInt8", "test_return_int8", nil, nil)
		ClassDBBindMethod(t, "TestReturnUint16", "test_return_uint16", nil, nil)
		ClassDBBindMethod(t, "TestReturnFloat32", "test_return_float32", nil, nil)
		ClassDBBindMethod(t, "TestGoStringRepeat", "test_go_string_repeat", nil, nil)
		ClassDBBindMethod(t, "TestEchoVector2i", "test_echo_vector2i", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoVector3", "test_echo_vector3", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoVector3i", "test_echo_vector3i", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoRect2", "test_echo_rect2", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoRect2i", "test_echo_rect2i", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoTransform2D", "test_echo_transform2d", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoVector4i", "test_echo_vector4i", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoPlane", "test_echo_plane", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoQuaternion", "test_echo_quaternion", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoAABB", "test_echo_aabb", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoBasis", "test_echo_basis", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoTransform3D", "test_echo_transform3d", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestEchoProjection", "test_echo_projection", []string{"v"}, nil)
		ClassDBBindMethod(t, "ReturnSomething", "return_something", []string{"base", "f32", "f64", "i", "i8", "i16", "i32", "i64"}, nil)
		ClassDBBindMethod(t, "ReturnSomethingConst", "return_something_const", nil, nil)
		ClassDBBindMethod(t, "ReturnEmptyRef", "return_empty_ref", nil, nil)

		ClassDBBindMethod(t, "TestArray", "test_array", nil, nil)
		ClassDBBindMethod(t, "TestTArrayArg", "test_tarray_arg", []string{"array"}, nil)
		ClassDBBindMethod(t, "TestTArray", "test_tarray", nil, nil)
		ClassDBBindMethod(t, "TestDictionary", "test_dictionary", nil, nil)
		ClassDBBindMethod(t, "TestNodeArgument", "test_node_argument", []string{"example"}, nil)
		ClassDBBindMethod(t, "TestStringOps", "test_string_ops", nil, nil)
		ClassDBBindMethod(t, "TestStrUtility", "test_str_utility", nil, nil)
		ClassDBBindMethod(t, "TestVectorOps", "test_vector_ops", nil, nil)
		ClassDBBindMethod(t, "TestInstanceFromIdUtility", "test_instance_from_id_utility", nil, nil)

		// varargs
		ClassDBBindMethodVarargs(t, "VarargsFunc", "varargs_func", nil, nil)
		ClassDBBindMethodVarargs(t, "VarargsFuncNv", "varargs_func_nv", nil, nil)
		ClassDBBindMethodVarargs(t, "VarargsFuncVoid", "varargs_func_void", nil, nil)

		ClassDBBindMethod(t, "DefArgs", "def_args", []string{"a", "b"}, []Variant{NewVariantInt64(100), NewVariantInt64(200)})
		// ClassDBBindMethodStatic(t, "TestStatic", "test_static", []string{"a", "b"}, nil)
		// ClassDBBindMethodStatic(t, "TestStatic2", "test_static2", nil, nil)

		ClassDBBindMethod(t, "TestSetPositionAndSize", "test_set_position_and_size", nil, nil)
		ClassDBBindMethod(t, "TestGetChildNode", "test_get_child_node", nil, nil)
		ClassDBBindMethod(t, "TestCharacterBody2D", "test_character_body_2d", []string{"body"}, nil)
		ClassDBBindMethod(t, "TestParentIsNil", "test_parent_is_nil", nil, nil)

		// Properties
		ClassDBAddPropertyGroup(t, "Test group", "group_")
		ClassDBAddPropertySubgroup(t, "Test subgroup", "group_subgroup_")

		ClassDBBindMethod(t, "GetV4", "get_v4", nil, nil)
		ClassDBBindMethod(t, "GetCustomPosition", "get_custom_position", nil, nil)
		ClassDBBindMethod(t, "SetCustomPosition", "set_custom_position", []string{"position"}, nil)
		ClassDBAddProperty(t, GDEXTENSION_VARIANT_TYPE_VECTOR2, "group_subgroup_custom_position", "set_custom_position", "get_custom_position")

		// Signals.
		ClassDBAddSignal(t, "custom_signal",
			SignalParam{
				Type: GDEXTENSION_VARIANT_TYPE_STRING,
				Name: "name"},
			SignalParam{
				Type: GDEXTENSION_VARIANT_TYPE_INT,
				Name: "value",
			})
		ClassDBBindMethod(t, "EmitCustomSignal", "emit_custom_signal", []string{"name", "value"}, nil)

		// constants
		ClassDBBindEnumConstant(t, "Example.ExampleEnum", "FIRST", int(ExampleFirst))
		ClassDBBindEnumConstant(t, "Example.ExampleEnum", "ANSWER_TO_EVERYTHING", int(AnswerToEverything))
		ClassDBBindConstant(t, "CONSTANT_WITHOUT_ENUM", int(EXAMPLE_ENUM_CONSTANT_WITHOUT_ENUM))
		ClassDBBindBitfieldFlag(t, "Example.ExampleBitfieldFlag", "FLAG_ONE", int(FlagOne))
		ClassDBBindBitfieldFlag(t, "Example.ExampleBitfieldFlag", "FLAG_TWO", int(FlagTwo))

		ClassDBBindMethod(t, "TestBitfield", "test_bitfield", []string{"flags"}, nil)

		ClassDBBindMethod(t, "CallableBind", "callable_bind", nil, nil)
		ClassDBBindMethod(t, "TestVariantVector2iConversion", "test_variant_vector2i_conversion", []string{"variant"}, nil)

		// others
		// ClassDBBindMethod(t, "TestCastTo", "test_cast_to", nil, nil)
		log.Debug("Example registered")
	})
}

func UnregisterClassExample() {
	ClassDBUnregisterClass[*Example]()
	log.Debug("Example unregistered")
}
