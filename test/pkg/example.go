package pkg

import (
	"fmt"
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
	"github.com/godot-go/godot-go/pkg/gdclassinit"
	. "github.com/godot-go/godot-go/pkg/gdutilfunc"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"unsafe"
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
	// SecondFlag belongs to the second, distinct bitfield enum (witness for
	// enum independence).
	SecondFlag ExampleBitfieldFlag = 1 << 2
)

// Example implements GDClass evidence
var _ GDClass = (*Example)(nil)

type Example struct {
	ControlImpl
	customPosition    Vector2
	propertyFromList  Vector3
	dprop             [3]Vector2
	notificationCodes []int32
	// Property hint fixtures (support-property-hints)
	hintedRange   float64
	hintedEnum    int64
	hintedTexture RefTexture2D
	hintedNode    Node
	storageOnly   int64
	readonlyValue int64
	// String-boundary retention fixture (ratify-string-marshalling): a
	// borrowed StringName kept past its call via the owned Copy().
	stringCopyHeld StringName
	// Slice retention fixture (support-native-slice-marshalling): a decoded
	// Go-owned []byte kept past its call for cross-frame re-reads.
	sliceHeld []byte
}

func (c *Example) Self() *Example { return c }

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

// PartialDefArgs has a single trailing default (b=200), exercising the
// trailing-default fill for a 0 < defaults < declared binding.
var partialDefArgsRan bool

func (e *Example) PartialDefArgs(p_a, p_b int32) int32 {
	partialDefArgsRan = true
	ret := p_a + p_b
	log.Info("PartialDefArgs called", zap.Int32("sum", ret))
	return ret
}

// TestPartialDefArgsRejectsShort drives a zero-argument call to partial_def_args
// (which declares two params with one trailing default) through the engine call
// API. The call is unsatisfiable (missing 2 > 1 default) and must be rejected
// without running the body. Returns 0 when the body correctly did not run.
func (e *Example) TestPartialDefArgsRejectsShort() int64 {
	partialDefArgsRan = false
	func() {
		defer func() { _ = recover() }()
		sn := NewStringNameWithLatin1Chars("partial_def_args")
		defer sn.Destroy()
		e.Call(sn)
	}()
	if partialDefArgsRan {
		return 1
	}
	return 0
}

// callErrorProbeRan records whether the CallErrorProbe method body executed, so
// the harness can confirm a rejected varcall never invokes the bound Go method.
var callErrorProbeRan bool

// CallErrorProbe is a sentinel bound to a single required int64 parameter. Its
// body flipping the package flag is the observable signal that a call was
// accepted and dispatched through the varcall callback.
func (e *Example) CallErrorProbe(p1 int64) int64 {
	callErrorProbeRan = true
	return p1
}

// TestCallErrorReporting drives the varcall argument-validation paths through the
// engine call API and returns the number of failed expectations (0 == all pass).
// Rejected calls must leave callErrorProbeRan false (body never runs); an
// accepted call must set it. The engine reports a call error for rejects, which
// the generated Object.Call surfaces as a Go panic; we recover because the
// assertion is on the body-ran flag, not the panic itself. main.gd only needs
// this to return 0, and its continuation proves the process survived every
// rejected call.
func (e *Example) TestCallErrorReporting() int64 {
	var failures int64

	probe := func(method string, args ...Variant) (ran bool) {
		callErrorProbeRan = false
		func() {
			defer func() { _ = recover() }()
			sn := NewStringNameWithLatin1Chars(method)
			defer sn.Destroy()
			e.Call(sn, args...)
		}()
		for _, a := range args {
			a.Destroy()
		}
		return callErrorProbeRan
	}

	// Rejected: too few arguments for the single required int64.
	if probe("call_error_probe") {
		failures++
	}
	// Rejected: too many arguments.
	if probe("call_error_probe", NewVariantInt64(1), NewVariantInt64(2)) {
		failures++
	}
	// Rejected: non-convertible argument type (Array into an int64 parameter).
	// Exercises a refcounted container argument on the reject path; the marshalled
	// view is a non-owning copy and the reject frees nothing, so the leak gate
	// stays green. We still release the variants we own.
	arr := NewArray()
	arrRef := NewVariantArray(arr)
	if probe("call_error_probe", arrRef) {
		failures++
	}
	arr.Destroy()
	// Rejected: non-convertible argument type (Vector2 into an int64 parameter).
	if probe("call_error_probe", NewVariantVector2(NewVector2())) {
		failures++
	}
	// Accepted: correct arity and type dispatches and runs the body.
	if !probe("call_error_probe", NewVariantInt64(5)) {
		failures++
	}

	return failures
}

// TestSingletonHelpers exercises the generated typed singleton accessors
// (openspec: support-singleton-accessors): wrapper identity, runtime
// availability, nil for engine-absent singletons with absence never frozen,
// and identical wrappers for concurrent cold-path callers. Returns a failure
// count.
func (e *Example) TestSingletonHelpers() int64 {
	var failures int64

	// Identity: every call returns the same wrapper value.
	if GetEngineSingleton() != GetEngineSingleton() {
		failures++
	}

	// Runtime singletons resolve non-nil; the wrapper answers to its class.
	if GetOSSingleton() == nil || GetProjectSettingsSingleton() == nil ||
		GetInputSingleton() == nil {
		failures++
	}
	if os := GetOSSingleton(); os == nil || os.GetClassName() != "OS" {
		failures++
	}

	// Editor-only singleton in a runtime build: nil on repeated calls. Two
	// lookups prove absence is returned safely and re-attempted rather than
	// frozen as a broken wrapper (the never-store-on-NULL cold path).
	if GetEditorInterfaceSingleton() != nil || GetEditorInterfaceSingleton() != nil {
		failures++
	}

	// Concurrent cold path: TranslationServer is untouched until here, so all
	// racers enter the builder; the atomic holder must publish one wrapper to
	// every caller with no torn interface value observable.
	const racers = 8
	gate := make(chan struct{})
	results := make(chan TranslationServer, racers)
	for i := 0; i < racers; i++ {
		go func() {
			<-gate
			results <- GetTranslationServerSingleton()
		}()
	}
	close(gate)
	first := <-results
	if first == nil {
		failures++
	}
	for i := 1; i < racers; i++ {
		if got := <-results; got != first {
			failures++
		}
	}

	return failures
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

// --- Property hint fixture accessors ---

func (e *Example) SetHintedRange(value float64) { e.hintedRange = value }
func (e *Example) GetHintedRange() float64      { return e.hintedRange }

func (e *Example) SetHintedEnum(value int64) { e.hintedEnum = value }
func (e *Example) GetHintedEnum() int64      { return e.hintedEnum }

func (e *Example) SetHintedTexture(value RefTexture2D) { e.hintedTexture = value }
func (e *Example) GetHintedTexture() RefTexture2D      { return e.hintedTexture }

func (e *Example) SetHintedNode(value Node) { e.hintedNode = value }
func (e *Example) GetHintedNode() Node      { return e.hintedNode }

func (e *Example) SetStorageOnly(value int64) { e.storageOnly = value }
func (e *Example) GetStorageOnly() int64      { return e.storageOnly }

func (e *Example) GetReadonlyValue() int64 { return e.readonlyValue }

// Main-loop hook recorders (support-main-loop-callbacks). Hooks run on the
// engine main thread; recorders stay allocation-free and never create engine
// objects, so the shutdown recorder cannot race the exit leak check.
var (
	mainLoopStartupRan atomic.Bool
	mainLoopFrameCount atomic.Int64
)

func RecordMainLoopStartup() {
	mainLoopStartupRan.Store(true)
}

func RecordMainLoopFrame() {
	mainLoopFrameCount.Add(1)
}

// LogMainLoopShutdownMarker logs the marker literal that
// test/check_test_output.sh requires: evidence the shutdown phase really
// fired in the headless run. Writing it through the log package while the
// engine tears down is also the live witness for the
// diagnostic-output-channel requirement "Diagnostics remain available while
// the engine tears down": the entry reaches captured output without any
// engine-provided output facility.
func LogMainLoopShutdownMarker() {
	log.Warn("GODOTGO_MAIN_LOOP_SHUTDOWN")
}

func (e *Example) MainLoopStartupRan() bool  { return mainLoopStartupRan.Load() }
func (e *Example) MainLoopFrameCount() int64 { return mainLoopFrameCount.Load() }

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

// TestObjectCast exercises the binding-identity contract end to end: a wrapper
// resolves through the instance binding to the wrapper standing for the same
// engine object, widening it to an ancestor keeps that owner, asserting it to an
// unrelated class fails instead of yielding a wrong wrapper, and the inheritance
// predicate accepts ancestors while rejecting unrelated classes. It returns a
// non-zero failure code so the GDScript assertion names which clause broke.
func (e *Example) TestObjectCast() int {
	owner := e.GetGodotObjectOwner()
	self, err := ObjectFromInstanceBinding(owner)
	if err != nil || self == nil || self.GetGodotObjectOwner() != owner {
		return 1
	}
	asControl, ok := self.(Control)
	if !ok || asControl.GetGodotObjectOwner() != owner {
		return 2
	}
	if _, ok := self.(Image); ok {
		return 3
	}
	if !ObjectInheritsClass(e.AsGDExtensionObjectPtr(), "Control") {
		return 4
	}
	if !ObjectInheritsClass(e.AsGDExtensionObjectPtr(), "Node") {
		return 5
	}
	if ObjectInheritsClass(e.AsGDExtensionObjectPtr(), "Image") {
		return 6
	}
	return 0
}

// TestObjectCasting witnesses both directions of a class change against a live
// engine object. A generated object return already stands for the object's
// runtime class, so narrowing it is the assertion alone and repeated returns are
// the identical wrapper; re-resolving the binding remains as coverage for
// wrappers that did not come from a generated return. It returns a non-zero
// failure code so the GDScript assertion names which clause broke.
func (e *Example) TestObjectCasting() int {
	// One NodePath for both calls: NodePath stores its components as
	// StringNames, so a throwaway path per call would leave an interned
	// reference outstanding at engine exit.
	name := NewStringWithUtf8Chars("Label")
	defer name.Destroy()
	path := NewNodePathWithString(name)
	defer path.Destroy()

	// The method declares Node, but the wrapper stands for the runtime class:
	// this Label is a Label to Go, with no helper call in between.
	node := e.GetNode(path)
	label, ok := node.(Label)
	if !ok {
		return 1
	}

	// Two returns of the same engine object are the identical wrapper.
	again := e.GetNode(path)
	if again != node {
		return 2
	}

	// Re-resolving the binding is still the path for wrappers that did not come
	// from a generated return, and it yields that same wrapper.
	wrapper, err := ObjectFromInstanceBinding(node.GetGodotObjectOwner())
	if err != nil || wrapper == nil {
		return 3
	}
	if wrapper != node {
		return 4
	}

	// Widening needs no lookup: the Label wrapper embeds every ancestor.
	widened, ok := label.(Node)
	if !ok || widened.GetGodotObjectOwner() != node.GetGodotObjectOwner() {
		return 5
	}
	return 0
}

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

// VarargsFuncNvBool declares a second return value on a variadic bind. Only the
// first value is delivered to the engine; the bool is dropped with a warning.
func (e *Example) VarargsFuncNvBool(args ...Variant) (int, bool) {
	return 70 + len(args), true
}

func (e *Example) V_Example_PropertyCanRevert(p_name StringName) bool {
	gdSn := NewStringNameWithLatin1Chars("property_from_list")
	defer gdSn.Destroy()
	vec3 := NewVector3WithFloat32Float32Float32(42, 42, 42)
	// Copy the field out first: cgo pointer checks forbid passing pointers
	// into composite objects that hold unpinned Go pointers (e.g. slices).
	pf := e.propertyFromList
	return p_name.Equal_StringName(gdSn) && !pf.Equal_Vector3(vec3)
}

func (e *Example) V_Example_PropertyGetRevert(p_name StringName) (Variant, bool) {
	gdSn := NewStringNameWithLatin1Chars("property_from_list")
	defer gdSn.Destroy()
	if p_name.Equal_StringName(gdSn) {
		vec3 := NewVector3WithFloat32Float32Float32(42, 42, 42)
		return NewVariantVector3(vec3), true
	}
	return Variant{}, false
}

func (e *Example) V_Example_Notification(what int32, reversed bool) {
	e.notificationCodes = append(e.notificationCodes, what)
}

func (e *Example) GetNotificationCodes() PackedInt32Array {
	arr := NewPackedInt32Array()
	for _, code := range e.notificationCodes {
		arr.Append(int64(code))
	}
	return arr
}

func (e *Example) V_Example_Set(name string, value Variant) bool {
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

func (e *Example) V_Example_Get(name string) (Variant, bool) {
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

func (e *Example) V_Example_Ready() {
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

func (e *Example) V_Example_Input(refEvent RefInputEvent) {
	event := refEvent.Ptr()
	if event == nil {
		log.Warn("Example.V_Input: null refEvent parameter")
		return
	}
	keyEvent, ok := event.(InputEventKey)
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

// TestEchoStringNameArg returns the received StringName argument unchanged.
// The argument is a borrow in both call styles, so the return path must
// classify it as a borrow echo and leave it undestroyed.
func (e *Example) TestEchoStringNameArg(p_name StringName) StringName {
	return p_name
}

// TestFreshStringName ignores the received argument's value and returns a
// newly constructed StringName with different contents, guaranteeing a non-
// echo classification (same-content clones can be interned byte-identical to
// the borrow) and exercising the destroy-after-encode accounting.
func (e *Example) TestFreshStringName(p_name StringName) StringName {
	_ = p_name
	return NewStringNameWithUtf8Chars("returned_fresh_name")
}

// TestEchoNodePathArg is TestEchoStringNameArg for NodePath.
func (e *Example) TestEchoNodePathArg(p_path NodePath) NodePath {
	return p_path
}

// TestFreshNodePath is TestFreshStringName for NodePath.
func (e *Example) TestFreshNodePath(p_path NodePath) NodePath {
	_ = p_path
	return NewNodePathWithString(NewStringWithUtf8Chars("fresh/path"))
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

// TestStringParamTextConversion receives a mixed-type value at a native
// string parameter alongside the expected text and counts mismatches. The
// value crosses the same decode path as the expected string, so a non-zero
// return means the conversion produced different text than the engine's.
func (e *Example) TestStringParamTextConversion(p_expected string, p_value string) int64 {
	if p_value != p_expected {
		return 1
	}
	return 0
}

// TestStringCopyStore receives a call-scoped borrow of a StringName, retains
// it via the owned Copy(), and stores it for cross-frame verification.
func (e *Example) TestStringCopyStore(p_name StringName) int64 {
	e.stringCopyHeld = p_name.Copy()
	return 0
}

// TestStringCopyVerifyRelease re-reads the StringName stored by an earlier
// frame, proving the Copy() outlived the borrow's call, then releases the
// owned reference.
func (e *Example) TestStringCopyVerifyRelease(p_expected string) int64 {
	failures := int64(0)
	if e.stringCopyHeld.ToUtf8() != p_expected {
		failures++
	}
	// Move the owned reference into a local cell before releasing it.
	// Destroy on the field itself hands cgo an address inside the
	// heap-allocated Example, and cgocheck=1 then scans the whole object
	// for pointers to unpinned Go memory - sibling slice and interface
	// fields trip that scan and panic. A local cell holds only the 8
	// content bytes, so the scan finds nothing. Value-receiver reads like
	// ToUtf8 above copy the cell to the stack first and are unaffected.
	held := e.stringCopyHeld
	e.stringCopyHeld = StringName{}
	held.Destroy()
	return failures
}

// TestCopyContentPreservation copies each of the three engine string types
// and checks the copy's text against the source's before releasing both.
func (e *Example) TestCopyContentPreservation() int64 {
	failures := int64(0)

	strOrig := NewStringWithUtf8Chars("copy_str_content")
	strCopy := strOrig.Copy()
	if strCopy.ToUtf8() != "copy_str_content" || strOrig.ToUtf8() != "copy_str_content" {
		failures++
	}
	strOrig.Destroy()
	strCopy.Destroy()

	nameOrig := NewStringNameWithUtf8Chars("copy_name_content")
	nameCopy := nameOrig.Copy()
	if nameCopy.ToUtf8() != "copy_name_content" || nameOrig.ToUtf8() != "copy_name_content" {
		failures++
	}
	nameOrig.Destroy()
	nameCopy.Destroy()

	pathStr := NewStringWithUtf8Chars("copy/path/content")
	pathOrig := NewNodePathWithString(pathStr)
	pathStr.Destroy()
	pathCopy := pathOrig.Copy()
	pathOrigText := NewStringWithNodePath(pathOrig)
	pathCopyText := NewStringWithNodePath(pathCopy)
	if pathCopyText.ToUtf8() != pathOrigText.ToUtf8() || pathOrigText.ToUtf8() != "copy/path/content" {
		failures++
	}
	pathOrigText.Destroy()
	pathCopyText.Destroy()
	pathOrig.Destroy()
	pathCopy.Destroy()

	return failures
}

// --- support-native-slice-marshalling fixtures ---------------------------
// Probes exercise slice argument decoding without returning a slice (the
// echo-return side ships with the return encoder). A decoded slice is a
// Go-owned deep copy (design D1); length and checksum probes prove the
// elements landed without corruption.

func (e *Example) TestSliceLenBytes(p_in []byte) int64 { return int64(len(p_in)) }

func (e *Example) TestSliceLenInt32(p_in []int32) int64 { return int64(len(p_in)) }

func (e *Example) TestSliceLenInt64(p_in []int64) int64 { return int64(len(p_in)) }

func (e *Example) TestSliceLenFloat32(p_in []float32) int64 { return int64(len(p_in)) }

func (e *Example) TestSliceLenFloat64(p_in []float64) int64 { return int64(len(p_in)) }

func (e *Example) TestSliceLenStrings(p_in []string) int64 { return int64(len(p_in)) }

func (e *Example) TestSliceLenVector2(p_in []Vector2) int64 { return int64(len(p_in)) }

func (e *Example) TestSliceLenVector3(p_in []Vector3) int64 { return int64(len(p_in)) }

func (e *Example) TestSliceLenVector4(p_in []Vector4) int64 { return int64(len(p_in)) }

func (e *Example) TestSliceLenColor(p_in []Color) int64 { return int64(len(p_in)) }

// TestSliceChecksumBytes sums every decoded byte; paired with the length it
// catches element corruption, ordering, and stride faults in the bulk copy.
func (e *Example) TestSliceChecksumBytes(p_in []byte) int64 {
	sum := int64(0)
	for _, b := range p_in {
		sum += int64(b)
	}
	return sum
}

func (e *Example) TestSliceChecksumInt64(p_in []int64) int64 {
	sum := int64(0)
	for _, v := range p_in {
		sum += v
	}
	return sum
}

// TestSliceChecksumStrings sums the UTF-8 byte lengths, so multi-byte text
// decodes as content and not as raw handles.
func (e *Example) TestSliceChecksumStrings(p_in []string) int64 {
	total := int64(0)
	for _, s := range p_in {
		total += int64(len(s))
	}
	return total
}

func (e *Example) TestSliceChecksumFloat64(p_in []float64) float64 {
	sum := float64(0)
	for _, v := range p_in {
		sum += v
	}
	return sum
}

// TestSliceEcho*K return the decoded slice unchanged: the return encoder
// builds a fresh Packed*K cell and GDScript asserts element equality against
// the original argument on both the call() and direct paths.

func (e *Example) TestSliceEchoBytes(p_in []byte) []byte         { return p_in }
func (e *Example) TestSliceEchoInt32(p_in []int32) []int32       { return p_in }
func (e *Example) TestSliceEchoInt64(p_in []int64) []int64       { return p_in }
func (e *Example) TestSliceEchoFloat32(p_in []float32) []float32 { return p_in }
func (e *Example) TestSliceEchoFloat64(p_in []float64) []float64 { return p_in }
func (e *Example) TestSliceEchoStrings(p_in []string) []string   { return p_in }
func (e *Example) TestSliceEchoVector2(p_in []Vector2) []Vector2 { return p_in }
func (e *Example) TestSliceEchoVector3(p_in []Vector3) []Vector3 { return p_in }
func (e *Example) TestSliceEchoVector4(p_in []Vector4) []Vector4 { return p_in }
func (e *Example) TestSliceEchoColor(p_in []Color) []Color       { return p_in }

// TestSliceEchoNilBytes returns a nil slice with no input: the encoding must
// produce an empty PackedByteArray, never a null Variant (design D6).
func (e *Example) TestSliceEchoNilBytes() []byte { return nil }

// TestSliceRetainStore keeps a decoded slice on the Example, past the
// call's borrow lifetime: the slice is Go-owned memory with nothing left to
// release engine-side once the call returns (design D1).
func (e *Example) TestSliceRetainStore(p_in []byte) int64 {
	e.sliceHeld = p_in
	return 0
}

// TestSliceRetainVerify re-reads the stored slice on a later frame, proving
// the decoded copy has no dependence on the engine array that fed it.
func (e *Example) TestSliceRetainVerify(p_expected_len int64, p_expected_sum int64) int64 {
	failures := int64(0)
	if int64(len(e.sliceHeld)) != p_expected_len {
		failures++
	}
	sum := int64(0)
	for _, b := range e.sliceHeld {
		sum += int64(b)
	}
	if sum != p_expected_sum {
		failures++
	}
	return failures
}

// TestSliceMutateProbe returns the sum of the decoded elements and then
// scribbles 0xFF over every slot. If decoding had aliased the engine buffer,
// the scribble would rewrite the GDScript-side original, which the driver
// asserts against after the call.
func (e *Example) TestSliceMutateProbe(p_in []byte) int64 {
	sum := int64(0)
	for _, b := range p_in {
		sum += int64(b)
	}
	for i := range p_in {
		p_in[i] = 0xFF
	}
	return sum
}

func (e *Example) TestEchoVector2i(p Vector2i) Vector2i                                 { return p }
func (e *Example) TestEchoVector3(p Vector3) Vector3                                    { return p }
func (e *Example) TestEchoVector3i(p Vector3i) Vector3i                                 { return p }
func (e *Example) TestEchoRect2(p Rect2) Rect2                                          { return p }
func (e *Example) TestEchoRect2i(p Rect2i) Rect2i                                       { return p }
func (e *Example) TestEchoTransform2D(p Transform2D) Transform2D                        { return p }
func (e *Example) TestEchoVector4i(p Vector4i) Vector4i                                 { return p }
func (e *Example) TestEchoPlane(p Plane) Plane                                          { return p }
func (e *Example) TestEchoQuaternion(p Quaternion) Quaternion                           { return p }
func (e *Example) TestEchoAABB(p AABB) AABB                                             { return p }
func (e *Example) TestEchoBasis(p Basis) Basis                                          { return p }
func (e *Example) TestEchoTransform3D(p Transform3D) Transform3D                        { return p }
func (e *Example) TestEchoProjection(p Projection) Projection                           { return p }
func (e *Example) TestEchoArray(arr Array) Array                                        { return arr }
func (e *Example) TestEchoPackedByteArray(arr PackedByteArray) PackedByteArray          { return arr }
func (e *Example) TestEchoPackedInt32Array(arr PackedInt32Array) PackedInt32Array       { return arr }
func (e *Example) TestEchoPackedInt64Array(arr PackedInt64Array) PackedInt64Array       { return arr }
func (e *Example) TestEchoPackedFloat32Array(arr PackedFloat32Array) PackedFloat32Array { return arr }
func (e *Example) TestEchoPackedFloat64Array(arr PackedFloat64Array) PackedFloat64Array { return arr }
func (e *Example) TestEchoPackedStringArray(arr PackedStringArray) PackedStringArray    { return arr }
func (e *Example) TestEchoPackedVector2Array(arr PackedVector2Array) PackedVector2Array { return arr }
func (e *Example) TestEchoPackedVector3Array(arr PackedVector3Array) PackedVector3Array { return arr }
func (e *Example) TestEchoPackedVector4Array(arr PackedVector4Array) PackedVector4Array { return arr }
func (e *Example) TestEchoPackedColorArray(arr PackedColorArray) PackedColorArray       { return arr }

func (e *Example) TestConsumeArray(arr Array) int64 {
	return arr.Size()
}

func (e *Example) TestConsumePackedByteArray(arr PackedByteArray) int64 {
	return arr.Size()
}

func (e *Example) TestConsumePackedInt32Array(arr PackedInt32Array) int64 {
	return arr.Size()
}

func (e *Example) TestConsumePackedInt64Array(arr PackedInt64Array) int64 {
	return arr.Size()
}

func (e *Example) TestConsumePackedFloat32Array(arr PackedFloat32Array) int64 {
	return arr.Size()
}

func (e *Example) TestConsumePackedFloat64Array(arr PackedFloat64Array) int64 {
	return arr.Size()
}

func (e *Example) TestConsumePackedStringArray(arr PackedStringArray) int64 {
	return arr.Size()
}

func (e *Example) TestConsumePackedVector2Array(arr PackedVector2Array) int64 {
	return arr.Size()
}

func (e *Example) TestConsumePackedVector3Array(arr PackedVector3Array) int64 {
	return arr.Size()
}

func (e *Example) TestConsumePackedColorArray(arr PackedColorArray) int64 {
	return arr.Size()
}

// TestMutateArray appends an element to the received array argument in place.
// Via ptrcall the argument is a borrow sharing caller storage, so the mutation
// is visible to the GDScript caller. Godot Array/Dictionary are shared-
// reference types without copy-on-write, so even varcall's owned decode shares
// the same underlying container and the mutation is visible there as well.
func (e *Example) TestMutateArray(arr Array) {
	arr.Append(NewVariantInt64(42))
}

// TestCloneEchoArray returns an unmutated defensive clone of the received
// array argument. The clone is byte-identical to the borrow, so
// isPtrcallBorrowEcho classifies it as an echo and skips destroying it,
// retaining one reference per call (documented bounded leak in
// container-lifecycle-management).
func (e *Example) TestCloneEchoArray(arr Array) Array {
	return NewArrayWithArray(arr)
}

// TestRebuildArrayPlus returns a freshly built Array holding the received
// argument's elements plus a trailing 7. Unlike a copy-constructor clone, the
// result shares no storage with the argument (Godot Array is a shared-reference
// type with no copy-on-write), so its bytes differ from every borrowed
// argument: isPtrcallBorrowEcho must not classify it as an echo, and its owned
// reference is destroyed after encoding.
func (e *Example) TestRebuildArrayPlus(arr Array) Array {
	out := NewArray()
	for i := int64(0); i < arr.Size(); i++ {
		out.Append(arr.Get(i))
	}
	out.Append(NewVariantInt64(7))
	return out
}

func (e *Example) TestEchoDictionary(d Dictionary) Dictionary { return d }

func (e *Example) TestConsumeDictionary(d Dictionary) int64 {
	return d.Size()
}

// TestMutateDictionary adds one known key to the received dictionary argument.
// Godot Dictionary is a shared-reference type without copy-on-write, so even
// varcall's owned decode shares the same underlying container and the new key
// is visible to the GDScript caller.
func (e *Example) TestMutateDictionary(d Dictionary) {
	d.SetKeyed("go_added", NewVariantInt64(42))
}

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

// TestStacktraceProbe drives a short C-side capture window so the runtime
// gate can verify native backtraces actually print. "start" reports whether
// capture is compiled in (one line, `stacktrace-probe capture=<0|1>`,
// straight from stacktraceCaptureAvailable) and enables capture; "hit" is
// called while capture is on, so its callback prints the backtrace before
// the body runs, and the body then prints this goroutine's Go stack — file
// and line for the Go side on every platform, independent of whether the
// native capture is compiled in; "stop" turns capture off so no further
// output leaks into the rest of the run.
func (e *Example) TestStacktraceProbe(p_phase String) String {
	switch p_phase.ToUtf8() {
	case "start":
		capture := 0
		if StacktraceCaptureAvailable() {
			capture = 1
		}
		fmt.Printf("stacktrace-probe capture=%d\n", capture)
		SetPrintStacktrace(true)
		return NewStringWithUtf8Chars("on")
	case "hit":
		stack := make([]byte, 8192)
		n := runtime.Stack(stack, false)
		fmt.Print("stacktrace-probe go-stack\n", string(stack[:n]))
		return NewStringWithUtf8Chars("hit")
	case "stop":
		SetPrintStacktrace(false)
		return NewStringWithUtf8Chars("off")
	}
	return NewStringWithUtf8Chars("unknown-phase")
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

// TestRegistrationBookkeeping asserts what registration left behind: the class
// is in the registration registry at the scene level, its constructor is
// callable, and the encoder installed for it accepts an instance. Non-zero codes
// identify the failing clause.
func (e *Example) TestRegistrationBookkeeping() int {
	ci, ok := Internal.GDRegisteredGDClasses.Get("Example")
	if !ok {
		return 1
	}
	if ci.Level != GDEXTENSION_INITIALIZATION_SCENE {
		return 2
	}
	ctor, ok := Internal.GDClassConstructors.Get("Example")
	if !ok {
		return 3
	}
	enc, ok := gdclassinit.GDRegisteredGDClassEncoders.Get("Example")
	if !ok || enc == nil {
		return 4
	}
	marker := 0
	owner := (*GodotObject)(unsafe.Pointer(&marker))
	inst := ctor(owner)
	if inst == nil {
		return 5
	}
	var out *GodotObject
	enc.EncodeReflectTypePtrArg(
		reflect.ValueOf(inst),
		(GDExtensionUninitializedTypePtr)(unsafe.Pointer(&out)),
	)
	if out != owner {
		return 6
	}
	return 0
}

func (e *Example) V_Example_ToString() string {
	// Deliberately unlike the binding's default format, and carrying non-ASCII.
	// The default format is ASCII-only and is also what a *failed* dispatch
	// produces, so an assertion written against it could not tell the two apart.
	// This string proves the virtual ran and that multi-byte UTF-8 survives the
	// crossing into Godot's String.
	return fmt.Sprintf("[ GDExtension::Example <--> Instance ID:%d | é中 ]", e.GetInstanceId())
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

// GetExamplePropertyList describes the Example properties exposed through the
// class get_property_list callback. Backing engine strings and stable cells
// are built per-class by registration (openspec
// fix-class-property-list-lifecycle).
func GetExamplePropertyList() []PropertyDescriptor {
	props := make([]PropertyDescriptor, 5)

	makeProp := func(className, propName string, vt GDExtensionVariantType) PropertyDescriptor {
		return PropertyDescriptor{
			ClassName:  className,
			Type:       vt,
			Name:       propName,
			Hint:       uint32(PROPERTY_HINT_NONE),
			HintString: "",
			Usage:      uint32(PROPERTY_USAGE_DEFAULT),
		}
	}

	props[0] = makeProp("Example", "property_from_list", GDEXTENSION_VARIANT_TYPE_VECTOR3)
	for i := 0; i < 3; i++ {
		props[i+1] = makeProp("Example", fmt.Sprintf("dproperty_%d", i), GDEXTENSION_VARIANT_TYPE_VECTOR2)
	}
	// Carries a non-empty hint string through the class callback so the
	// persistence gate can assert hint-string content end to end
	// (openspec fix-class-property-list-lifecycle review).
	props[4] = PropertyDescriptor{
		ClassName:  "Example",
		Type:       GDEXTENSION_VARIANT_TYPE_STRING,
		Name:       "property_hint_string",
		Hint:       uint32(PROPERTY_HINT_NONE),
		HintString: "review-carrier",
		Usage:      uint32(PROPERTY_USAGE_DEFAULT),
	}

	return props
}

func RegisterClassExample() {
	props := GetExamplePropertyList()
	ClassDBRegisterClass(NewExampleFromOwnerObject, props, ValidateExampleProperty, func(t *Example) {
		// virtuals
		// virtuals
		ClassDBBindMethodVirtual(t, "V_Example_ToString", "to_string", nil, nil)
		ClassDBBindMethodVirtual(t, "V_Example_Ready", "_ready", nil, nil)
		ClassDBBindMethodVirtual(t, "V_Example_Input", "_input", []string{"event"}, nil)
		// Callback virtuals through the dedicated typed binds (typed-method-
		// dispatch D6): the engine slots run compiled fixed-shape closures.
		// to_string stays name-string-bound, keeping the reflective fallback
		// under the to-string suite while every other callback suite runs
		// its compiled path.
		ClassDBBindVirtualSet(t, (*Example).V_Example_Set)
		ClassDBBindVirtualGet(t, (*Example).V_Example_Get)
		ClassDBBindVirtualPropertyCanRevert(t, (*Example).V_Example_PropertyCanRevert)
		ClassDBBindVirtualPropertyGetRevert(t, (*Example).V_Example_PropertyGetRevert)
		ClassDBBindVirtualNotification(t, (*Example).V_Example_Notification)

		ClassDBBindMethod(t, "SimpleFunc", "simple_func", nil, nil)
		ClassDBBindMethod(t, "SimpleConstFunc", "simple_const_func", []string{"a"}, nil)
		// ClassDBBindMethod(t, "CustomRefFunc", "custom_ref_func", []string{"ref"}, nil)
		ClassDBBindMethod(t, "ImageRefFunc", "image_ref_func", []string{"image"}, nil)
		ClassDBBindMethod(t, "PassThroughRefImage", "pass_through_ref_image", []string{"image"}, nil)
		ClassDBBindMethod(t, "ReturnNilRefImage", "return_nil_ref_image", nil, nil)
		ClassDBBindMethod(t, "TestRefLifecycle", "test_ref_lifecycle", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestFinalizerRelease", "test_finalizer_release", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestStringArgEcho", "test_string_arg_echo", []string{"str"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestStringNameArgEcho, "test_string_name_arg_echo", []string{"name"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestNodePathArgEcho, "test_node_path_arg_echo", []string{"path"}, nil)
		ClassDBBindCopyRetMethod0(t, (*Example).TestReturnStringName, "test_return_string_name", nil, nil)
		ClassDBBindCopyRetMethod0(t, (*Example).TestReturnNodePath, "test_return_node_path", nil, nil)
		// Borrow-echo fixtures through the typed copy-return family (design
		// D5): echo calls classify as borrows and skip the source release;
		// fresh returns are classified owned and destroyed after encoding.
		// The shutdown leak gates compare the totals against the reflective
		// baseline the same fixtures produced before conversion.
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoStringNameArg, "test_echo_string_name_arg", []string{"name"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestFreshStringName, "test_fresh_string_name", []string{"name"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoNodePathArg, "test_echo_node_path_arg", []string{"path"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestFreshNodePath, "test_fresh_node_path", []string{"path"}, nil)
		// Object-argument encoding (openspec: fix-object-arg-ptrcall-encoding).
		ClassDBBindMethod(t, "TestObjectArgAddChild", "test_object_arg_add_child", []string{"child"}, nil)
		ClassDBBindMethod(t, "TestObjectArgIdentity", "test_object_arg_identity", []string{"parent", "child"}, nil)
		ClassDBBindMethod(t, "TestObjectArgSetShape", "test_object_arg_set_shape", []string{"owner", "shape"}, nil)
		ClassDBBindMethod(t, "TestObjectArgSetShapeTypedNil", "test_object_arg_set_shape_typed_nil", []string{"owner"}, nil)
		ClassDBBindMethod(t, "TestObjectArgSetShapeInvalidRef", "test_object_arg_set_shape_invalid_ref", []string{"owner"}, nil)
		ClassDBBindMethod(t, "TestObjectArgNilPlainObject", "test_object_arg_nil_plain_object", []string{"node"}, nil)
		ClassDBBindMethod(t, "TestObjectArgRefcountStability", "test_object_arg_refcount_stability", []string{"owner", "shape"}, nil)
		ClassDBBindMethod(t, "TestObjectArgBaseClassParam", "test_object_arg_base_class_param", []string{"shape"}, nil)
		ClassDBBindMethod(t, "TestObjectArgReturnNode", "test_object_arg_return_node", []string{"node"}, nil)
		ClassDBBindMethod(t, "TestPinScratchStaysFlat", "test_pin_scratch_stays_flat", []string{"small", "large"}, nil)
		ClassDBBindMethod(t, "TestInstanceChurnBegin", "test_instance_churn_begin", nil, nil)
		ClassDBBindMethod(t, "TestInstanceChurnCapture", "test_instance_churn_capture", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestInstanceChurnEnd", "test_instance_churn_end", []string{"phase", "batch"}, nil)
		ClassDBBindMethod(t, "TestInstanceChurnVerify", "test_instance_churn_verify", nil, nil)
		ClassDBBindMethod(t, "TestBindingChurnBegin", "test_binding_churn_begin", nil, nil)
		ClassDBBindMethod(t, "TestBindingChurnCapture", "test_binding_churn_capture", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestBindingChurnVerify", "test_binding_churn_verify", nil, nil)
		ClassDBBindMethod(t, "TestPropListTeardownCheck", "test_prop_list_teardown_check", nil, nil)
		ClassDBBindMethod(t, "TestPropListChurnBegin", "test_prop_list_churn_begin", nil, nil)
		ClassDBBindMethod(t, "TestPropListChurnEnd", "test_prop_list_churn_end", []string{"batch"}, nil)
		ClassDBBindMethod(t, "TestObjectArgRelease", "test_object_arg_release", nil, nil)
		ClassDBBindMethod(t, "TestReturnRefcountStability", "test_return_refcount_stability", []string{"owner", "iterations"}, nil)
		ClassDBBindMethod(t, "TestReturnDroppedReleases", "test_return_dropped_releases", []string{"owner"}, nil)
		ClassDBBindMethod(t, "TestReturnUnrefIdempotence", "test_return_unref_idempotence", []string{"owner"}, nil)
		ClassDBBindMethod(t, "TestBorrowedRefNeverReleases", "test_borrowed_ref_never_releases", []string{"owner"}, nil)
		ClassDBBindMethod(t, "TestNullReturnSchedulesNoRelease", "test_null_return_schedules_no_release", []string{"owner"}, nil)
		ClassDBBindMethod(t, "TestReturnOwnershipRelease", "test_return_ownership_release", nil, nil)
		ClassDBBindMethod(t, "TestReturnRefPointeeNarrowing", "test_return_ref_pointee_narrowing", []string{"owner"}, nil)
		ClassDBBindMethod(t, "TestReturnRefPointeeIdentity", "test_return_ref_pointee_identity", []string{"owner"}, nil)
		ClassDBBindMethod(t, "TestPtrcallDecodePlainObject", "test_ptrcall_decode_plain_object", []string{"node"}, nil)
		ClassDBBindMethod(t, "TestPtrcallDecodeSubclass", "test_ptrcall_decode_subclass", []string{"shape"}, nil)
		ClassDBBindMethod(t, "TestPtrcallDecodeNullObject", "test_ptrcall_decode_null_object", nil, nil)
		ClassDBBindMethod(t, "TestPtrcallDecodeRefRegression", "test_ptrcall_decode_ref_regression", []string{"shape"}, nil)
		ClassDBBindMethod(t, "TestPtrcallDecodeUndecodableInterface", "test_ptrcall_decode_undecodable_interface", []string{"node"}, nil)
		ClassDBBindMethod(t, "TestPtrcallDecodeRelease", "test_ptrcall_decode_release", nil, nil)
		// support-variant-reflect-completeness: Ref argument/return crossing.
		ClassDBBindMethod(t, "TestRefBoundaryValid", "test_ref_boundary_valid", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestRefBoundaryEcho", "test_ref_boundary_echo", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestRefBoundaryNullReturn", "test_ref_boundary_null_return", nil, nil)
		ClassDBBindMethod(t, "TestRefBoundaryBareRefValid", "test_bare_ref_valid", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestRefBoundarySubclassSliced", "test_ref_boundary_subclass_sliced", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestRefBoundaryRefcountAtLeastTwo", "test_ref_boundary_refcount", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestRefBoundaryStash", "test_ref_boundary_stash", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestRefBoundaryStashAlive", "test_ref_boundary_stash_alive", nil, nil)
		ClassDBBindMethod(t, "TestRefBoundaryRelease", "test_ref_boundary_release", nil, nil)
		ClassDBBindMethod(t, "TestRefBoundaryRejectProbe", "test_ref_boundary_reject_probe", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestRefBoundaryRejectProbeCount", "test_ref_boundary_reject_probe_count", nil, nil)
		ClassDBBindMethod(t, "TestRefBoundaryUserRefId", "test_ref_boundary_user_ref_id", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestRefBoundaryBareRefEcho", "test_bare_ref_echo", []string{"ref"}, nil)
		ClassDBBindMethod(t, "TestRefBoundaryPtrcallCell", "test_ref_boundary_ptrcall_cell", []string{"shape"}, nil)
		ClassDBBindMethod(t, "TestRefBoundaryPtrcallBareRefCell", "test_ref_boundary_ptrcall_bare_ref_cell", []string{"shape"}, nil)
		ClassDBBindMethod(t, "TestRefBoundaryPtrcallNilCell", "test_ref_boundary_ptrcall_nil_cell", nil, nil)
		ClassDBBindMethod(t, "TestRefBoundaryPtrcallMismatchCell", "test_ref_boundary_ptrcall_mismatch_cell", []string{"node"}, nil)
		// support-go-callable-bridge: Go closure -> Callable bridge.
		ClassDBBindMethod(t, "CreateCounterCallable", "create_counter_callable", nil, nil)
		ClassDBBindMethod(t, "CreateProbeCallable", "create_probe_callable", nil, nil)
		ClassDBBindMethod(t, "CreateErrorCallable", "create_error_callable", nil, nil)
		ClassDBBindMethod(t, "CreatePanicCallable", "create_panic_callable", nil, nil)
		ClassDBBindMethod(t, "CreateCopyStoredCallable", "create_copy_stored_callable", nil, nil)
		ClassDBBindMethod(t, "ReleaseGoCallableStoredCopy", "release_go_callable_stored_copy", nil, nil)
		ClassDBBindMethod(t, "CreateReentrantCallable", "create_reentrant_callable", nil, nil)
		ClassDBBindMethod(t, "CallableBindTarget", "callable_bind_target", []string{"value"}, nil)
		ClassDBBindMethod(t, "GoCallableInvocationCount", "go_callable_invocation_count", nil, nil)
		ClassDBBindMethod(t, "ResetGoCallableInvocationCount", "reset_go_callable_invocation_count", nil, nil)
		ClassDBBindMethod(t, "GoCallableRegistryCount", "go_callable_registry_count", nil, nil)
		// support-signal-await: signal -> goroutine await bridge.
		ClassDBBindMethod(t, "AwaitHitStart", "await_hit_start", []string{"target"}, nil)
		ClassDBBindMethod(t, "AwaitHitCode", "await_hit_code", nil, nil)
		ClassDBBindMethod(t, "AwaitMarshalStart", "await_marshal_start", []string{"target", "want"}, nil)
		ClassDBBindMethod(t, "AwaitMarshalCode", "await_marshal_code", nil, nil)
		ClassDBBindMethod(t, "AwaitOneShotStart", "await_oneshot_start", []string{"target"}, nil)
		ClassDBBindMethod(t, "AwaitOneShotCode", "await_oneshot_code", nil, nil)
		ClassDBBindMethod(t, "AwaitOneShotSilenceCode", "await_oneshot_silence_code", nil, nil)
		ClassDBBindMethod(t, "AwaitOneShotRearm", "await_oneshot_rearm", []string{"target"}, nil)
		ClassDBBindMethod(t, "AwaitOneShotRearmCode", "await_oneshot_rearm_code", nil, nil)
		ClassDBBindMethod(t, "AwaitMisuseRecovered", "await_misuse_recovered", []string{"target"}, nil)
		ClassDBBindMethod(t, "AwaitFreedStart", "await_freed_start", []string{"target"}, nil)
		ClassDBBindMethod(t, "AwaitFreedCode", "await_freed_code", nil, nil)
		ClassDBBindMethod(t, "AwaitShutdownArm", "await_shutdown_arm", []string{"target"}, nil)
		ClassDBBindMethod(t, "AwaitTimerStart", "await_timer_start", []string{"timer"}, nil)
		ClassDBBindMethod(t, "AwaitTimerCode", "await_timer_code", nil, nil)
		// support-static-methods: Go package functions bound as class statics.
		// static_add rides the typed static bind (compiled value family);
		// the remaining statics keep the name-string path.
		ClassDBBindStaticValueMethod2(t, ExampleStaticAdd, "static_add", []string{"a", "b"}, nil)
		ClassDBBindStaticMethod(t, "static_with_default", ExampleStaticWithDefault,
			[]string{"base", "bonus"}, []Variant{NewVariantInt64(10)})
		ClassDBBindStaticMethod(t, "static_sum_varargs", ExampleStaticSumVarargs, nil, nil)
		ClassDBBindStaticMethod(t, "static_object_class", ExampleStaticObjectClass, []string{"node"}, nil)
		ClassDBBindStaticMethod(t, "static_echo_node", ExampleStaticEchoNode, []string{"node"}, nil)
		ClassDBBindStaticMethod(t, "static_arity_probe", ExampleStaticArityProbe, []string{"a"}, nil)
		ClassDBBindMethod(t, "StaticBadArityRejected", "static_bad_arity_rejected", nil, nil)
		ClassDBBindMethod(t, "EditorPluginRejectedOutsideEditor", "editor_plugin_rejected_outside_editor", nil, nil)
		ClassDBBindMethod(t, "StaticInstanceCallObservation", "static_instance_call_observation", nil, nil)
		ClassDBBindMethod(t, "StaticDuplicateBindRecovered", "static_duplicate_bind_recovered", nil, nil)
		ClassDBBindMethod(t, "StaticNonCallableRecovered", "static_non_callable_recovered", nil, nil)
		// support-variant-reflect-completeness: 36-type matrix audit echoes.
		ClassDBBindMethod(t, "TestAuditBool", "test_audit_bool", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditInt", "test_audit_int", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditFloat", "test_audit_float", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditString", "test_audit_string", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditVector2", "test_audit_vector2", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditVector2i", "test_audit_vector2i", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditRect2", "test_audit_rect2", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditRect2i", "test_audit_rect2i", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditVector3", "test_audit_vector3", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditVector3i", "test_audit_vector3i", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditTransform2D", "test_audit_transform2d", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditVector4", "test_audit_vector4", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditVector4i", "test_audit_vector4i", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditPlane", "test_audit_plane", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditQuaternion", "test_audit_quaternion", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditAABB", "test_audit_aabb", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditBasis", "test_audit_basis", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditTransform3D", "test_audit_transform3d", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditProjection", "test_audit_projection", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditColor", "test_audit_color", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditStringName", "test_audit_string_name", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditNodePath", "test_audit_node_path", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditRID", "test_audit_rid", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditObject", "test_audit_object", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditCallable", "test_audit_callable", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditSignal", "test_audit_signal", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditDictionary", "test_audit_dictionary", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditArray", "test_audit_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestAuditVariant", "test_audit_variant", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestUserDefinedNodeArg", "test_user_defined_node_arg", []string{"node", "expected_id"}, nil)
		ClassDBBindMethod(t, "TestUserDefinedArgStable", "test_user_defined_arg_stable", []string{"node"}, nil)
		ClassDBBindMethod(t, "TestEngineClassArgStillResolves", "test_engine_class_arg_still_resolves", []string{"node", "expected_id"}, nil)
		ClassDBBindMethod(t, "TestUnresolvableBindingIsTypedError", "test_unresolvable_binding_is_typed_error", nil, nil)
		ClassDBBindMethod(t, "TestVarcallDecodeWrongClass", "test_varcall_decode_wrong_class", []string{"label"}, nil)
		ClassDBBindMethod(t, "TestVarcallDecodeSubclassStillAccepted", "test_varcall_decode_subclass_still_accepted", []string{"label"}, nil)
		ClassDBBindMethod(t, "TestVarcallDecodeOwnedPrefixReleased", "test_varcall_decode_owned_prefix_released", []string{"arr", "label"}, nil)
		ClassDBBindMethod(t, "TestVarcallDecodeSuccessReleasesOwnedContainer", "test_varcall_decode_success_releases_owned_container", []string{"arr"}, nil)
		ClassDBBindMethod(t, "TestVarcallRejectProbe", "test_varcall_reject_probe", []string{"target"}, nil)
		ClassDBBindMethod(t, "TestVarcallRejectProbeCount", "test_varcall_reject_probe_count", nil, nil)

		ClassDBBindMethod(t, "TestScalarEcho", "test_scalar_echo", []string{"p_bool", "p_i64", "p_f64", "p_str"}, nil)
		ClassDBBindMethod(t, "TestUint64Echo", "test_uint64_echo", []string{"u64"}, nil)
		ClassDBBindMethod(t, "TestReturnInt8", "test_return_int8", nil, nil)
		ClassDBBindMethod(t, "TestReturnUint16", "test_return_uint16", nil, nil)
		ClassDBBindMethod(t, "TestReturnFloat32", "test_return_float32", nil, nil)
		ClassDBBindMethod(t, "TestGoStringRepeat", "test_go_string_repeat", nil, nil)
		ClassDBBindMethod(t, "TestStringParamTextConversion", "test_string_param_text_conversion", []string{"p_expected", "p_value"}, nil)
		ClassDBBindMethod(t, "TestStringCopyStore", "test_string_copy_store", []string{"p_name"}, nil)
		ClassDBBindMethod(t, "TestStringCopyVerifyRelease", "test_string_copy_verify_release", []string{"p_expected"}, nil)
		ClassDBBindMethod(t, "TestCopyContentPreservation", "test_copy_content_preservation", nil, nil)
		ClassDBBindMethod(t, "TestSliceLenBytes", "test_slice_len_bytes", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceLenInt32", "test_slice_len_int32", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceLenInt64", "test_slice_len_int64", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceLenFloat32", "test_slice_len_float32", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceLenFloat64", "test_slice_len_float64", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceLenStrings", "test_slice_len_strings", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceLenVector2", "test_slice_len_vector2", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceLenVector3", "test_slice_len_vector3", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceLenVector4", "test_slice_len_vector4", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceLenColor", "test_slice_len_color", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceChecksumBytes", "test_slice_checksum_bytes", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceChecksumInt64", "test_slice_checksum_int64", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceChecksumStrings", "test_slice_checksum_strings", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceChecksumFloat64", "test_slice_checksum_float64", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceEchoBytes", "test_slice_echo_bytes", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceEchoInt32", "test_slice_echo_int32", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceEchoInt64", "test_slice_echo_int64", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceEchoFloat32", "test_slice_echo_float32", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceEchoFloat64", "test_slice_echo_float64", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceEchoStrings", "test_slice_echo_strings", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceEchoVector2", "test_slice_echo_vector2", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceEchoVector3", "test_slice_echo_vector3", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceEchoVector4", "test_slice_echo_vector4", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceEchoColor", "test_slice_echo_color", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceEchoNilBytes", "test_slice_echo_nil_bytes", nil, nil)
		ClassDBBindMethod(t, "TestSliceRetainStore", "test_slice_retain_store", []string{"p_in"}, nil)
		ClassDBBindMethod(t, "TestSliceRetainVerify", "test_slice_retain_verify", []string{"p_expected_len", "p_expected_sum"}, nil)
		ClassDBBindMethod(t, "TestSliceMutateProbe", "test_slice_mutate_probe", []string{"p_in"}, nil)
		// Value-family echoes through typed binds (compiled ptrcall/varcall).
		ClassDBBindValueMethod1(t, (*Example).TestEchoVector2i, "test_echo_vector2i", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoVector3, "test_echo_vector3", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoVector3i, "test_echo_vector3i", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoRect2, "test_echo_rect2", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoRect2i, "test_echo_rect2i", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoTransform2D, "test_echo_transform2d", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoVector4i, "test_echo_vector4i", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoPlane, "test_echo_plane", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoQuaternion, "test_echo_quaternion", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoAABB, "test_echo_aabb", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoBasis, "test_echo_basis", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoTransform3D, "test_echo_transform3d", []string{"v"}, nil)
		ClassDBBindValueMethod1(t, (*Example).TestEchoProjection, "test_echo_projection", []string{"v"}, nil)
		// Container echoes through the typed copy-return family: the echo
		// classification must keep every shared-storage borrow unreleased
		// (borrow-echo-detection), which the shutdown leak gates assert.
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoArray, "test_echo_array", []string{"v"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoPackedByteArray, "test_echo_packed_byte_array", []string{"v"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoPackedInt32Array, "test_echo_packed_int32_array", []string{"v"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoPackedInt64Array, "test_echo_packed_int64_array", []string{"v"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoPackedFloat32Array, "test_echo_packed_float32_array", []string{"v"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoPackedFloat64Array, "test_echo_packed_float64_array", []string{"v"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoPackedStringArray, "test_echo_packed_string_array", []string{"v"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoPackedVector2Array, "test_echo_packed_vector2_array", []string{"v"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoPackedVector3Array, "test_echo_packed_vector3_array", []string{"v"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoPackedVector4Array, "test_echo_packed_vector4_array", []string{"v"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoPackedColorArray, "test_echo_packed_color_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestConsumeArray", "test_consume_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestConsumePackedByteArray", "test_consume_packed_byte_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestConsumePackedInt32Array", "test_consume_packed_int32_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestConsumePackedInt64Array", "test_consume_packed_int64_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestConsumePackedFloat32Array", "test_consume_packed_float32_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestConsumePackedFloat64Array", "test_consume_packed_float64_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestConsumePackedStringArray", "test_consume_packed_string_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestConsumePackedVector2Array", "test_consume_packed_vector2_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestConsumePackedVector3Array", "test_consume_packed_vector3_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestConsumePackedColorArray", "test_consume_packed_color_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestMutateArray", "test_mutate_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestCloneEchoArray", "test_clone_echo_array", []string{"v"}, nil)
		ClassDBBindMethod(t, "TestRebuildArrayPlus", "test_rebuild_array_plus", []string{"v"}, nil)
		ClassDBBindCopyRetMethod1(t, (*Example).TestEchoDictionary, "test_echo_dictionary", []string{"d"}, nil)
		ClassDBBindMethod(t, "TestConsumeDictionary", "test_consume_dictionary", []string{"d"}, nil)
		ClassDBBindMethod(t, "TestMutateDictionary", "test_mutate_dictionary", []string{"d"}, nil)
		ClassDBBindMethod(t, "ReturnSomething", "return_something", []string{"base", "f32", "f64", "i", "i8", "i16", "i32", "i64"}, nil)
		ClassDBBindMethod(t, "ReturnSomethingConst", "return_something_const", nil, nil)
		ClassDBBindMethod(t, "ReturnEmptyRef", "return_empty_ref", nil, nil)

		ClassDBBindCopyRetMethod0(t, (*Example).TestArray, "test_array", nil, nil)
		ClassDBBindValueMethod1(t, (*Example).TestTArrayArg, "test_tarray_arg", []string{"array"}, nil)
		ClassDBBindCopyRetMethod0(t, (*Example).TestTArray, "test_tarray", nil, nil)
		ClassDBBindCopyRetMethod0(t, (*Example).TestDictionary, "test_dictionary", nil, nil)
		ClassDBBindMethod(t, "TestNodeArgument", "test_node_argument", []string{"example"}, nil)
		ClassDBBindMethod(t, "TestStringOps", "test_string_ops", nil, nil)
		ClassDBBindMethod(t, "TestStrUtility", "test_str_utility", nil, nil)
		ClassDBBindMethod(t, "TestVectorOps", "test_vector_ops", nil, nil)
		ClassDBBindMethod(t, "TestInstanceFromIdUtility", "test_instance_from_id_utility", nil, nil)

		// varargs
		ClassDBBindMethodVarargs(t, "VarargsFunc", "varargs_func", nil, nil)
		ClassDBBindMethodVarargs(t, "VarargsFuncNv", "varargs_func_nv", nil, nil)
		ClassDBBindMethodVarargs(t, "VarargsFuncNvBool", "varargs_func_nv_bool", nil, nil)
		ClassDBBindMethodVarargs(t, "VarargsFuncVoid", "varargs_func_void", nil, nil)

		ClassDBBindMethod(t, "DefArgs", "def_args", []string{"a", "b"}, []Variant{NewVariantInt64(100), NewVariantInt64(200)})
		ClassDBBindMethod(t, "PartialDefArgs", "partial_def_args", []string{"a", "b"}, []Variant{NewVariantInt64(200)})
		ClassDBBindMethod(t, "TestPartialDefArgsRejectsShort", "test_partial_def_args_rejects_short", nil, nil)
		ClassDBBindMethod(t, "CallErrorProbe", "call_error_probe", []string{"p1"}, nil)
		ClassDBBindMethod(t, "TestCallErrorReporting", "test_call_error_reporting", nil, nil)
		ClassDBBindMethod(t, "TestSingletonHelpers", "test_singleton_helpers", nil, nil)
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
		ClassDBBindCopyRetMethod0(t, (*Example).GetNotificationCodes, "get_notification_codes", nil, nil)
		ClassDBBindMethod(t, "GetCustomPosition", "get_custom_position", nil, nil)
		ClassDBBindMethod(t, "SetCustomPosition", "set_custom_position", []string{"position"}, nil)
		ClassDBAddProperty(t, GDEXTENSION_VARIANT_TYPE_VECTOR2, "group_subgroup_custom_position", "set_custom_position", "get_custom_position")

		// Properties with inspector metadata (ClassDBAddPropertyInfo).
		ClassDBBindMethod(t, "GetHintedRange", "get_hinted_range", nil, nil)
		ClassDBBindMethod(t, "SetHintedRange", "set_hinted_range", []string{"value"}, nil)
		ClassDBAddPropertyInfo(t, PropertySpec{
			VariantType: GDEXTENSION_VARIANT_TYPE_FLOAT,
			Name:        "hinted_range",
			Setter:      "set_hinted_range",
			Getter:      "get_hinted_range",
			Hint:        PROPERTY_HINT_RANGE,
			HintString:  "0:100:1",
		})

		ClassDBBindMethod(t, "GetHintedEnum", "get_hinted_enum", nil, nil)
		ClassDBBindMethod(t, "SetHintedEnum", "set_hinted_enum", []string{"value"}, nil)
		ClassDBAddPropertyInfo(t, PropertySpec{
			VariantType: GDEXTENSION_VARIANT_TYPE_INT,
			Name:        "hinted_enum",
			Setter:      "set_hinted_enum",
			Getter:      "get_hinted_enum",
			Hint:        PROPERTY_HINT_ENUM,
			HintString:  "idle,walk,run",
		})

		ClassDBBindMethod(t, "GetHintedTexture", "get_hinted_texture", nil, nil)
		ClassDBBindMethod(t, "SetHintedTexture", "set_hinted_texture", []string{"texture"}, nil)
		ClassDBAddPropertyInfo(t, PropertySpec{
			VariantType: GDEXTENSION_VARIANT_TYPE_OBJECT,
			Name:        "hinted_texture",
			Setter:      "set_hinted_texture",
			Getter:      "get_hinted_texture",
			Hint:        PROPERTY_HINT_RESOURCE_TYPE,
			HintString:  "Texture2D",
		})

		ClassDBBindMethod(t, "GetHintedNode", "get_hinted_node", nil, nil)
		ClassDBBindMethod(t, "SetHintedNode", "set_hinted_node", []string{"node"}, nil)
		ClassDBAddPropertyInfo(t, PropertySpec{
			VariantType: GDEXTENSION_VARIANT_TYPE_OBJECT,
			Name:        "hinted_node",
			Setter:      "set_hinted_node",
			Getter:      "get_hinted_node",
			ClassName:   "Node",
		})

		ClassDBBindMethod(t, "GetStorageOnly", "get_storage_only", nil, nil)
		ClassDBBindMethod(t, "SetStorageOnly", "set_storage_only", []string{"value"}, nil)
		ClassDBAddPropertyInfo(t, PropertySpec{
			VariantType: GDEXTENSION_VARIANT_TYPE_INT,
			Name:        "storage_only",
			Setter:      "set_storage_only",
			Getter:      "get_storage_only",
			Usage:       PROPERTY_USAGE_STORAGE,
		})

		// Read-only property: registration with no setter name.
		ClassDBBindMethod(t, "GetReadonlyValue", "get_readonly_value", nil, nil)
		ClassDBAddPropertyInfo(t, PropertySpec{
			VariantType: GDEXTENSION_VARIANT_TYPE_INT,
			Name:        "readonly_value",
			Getter:      "get_readonly_value",
		})

		// Main-loop hook evidence accessors (support-main-loop-callbacks).
		ClassDBBindMethod(t, "MainLoopStartupRan", "main_loop_startup_ran", nil, nil)
		ClassDBBindMethod(t, "MainLoopFrameCount", "main_loop_frame_count", nil, nil)

		// Signals.
		ClassDBAddSignal(t, "custom_signal",
			SignalParam{
				Type: GDEXTENSION_VARIANT_TYPE_STRING,
				Name: "name"},
			SignalParam{
				Type: GDEXTENSION_VARIANT_TYPE_INT,
				Name: "value",
			})
		ClassDBAddSignal(t, "empty_signal")
		ClassDBBindMethod(t, "EmitCustomSignal", "emit_custom_signal", []string{"name", "value"}, nil)
		ClassDBBindMethod(t, "TestDuplicateSignalRejection", "test_duplicate_signal_rejection", nil, nil)
		ClassDBBindMethod(t, "TestConstantCollisionRejection", "test_constant_collision_rejection", nil, nil)
		ClassDBBindMethod(t, "TestPropertyGroupUnknownClass", "test_property_group_unknown_class", nil, nil)

		// Issue #151 regression: vararg call return buffers (see vararg_return_tests.go).
		ClassDBBindMethod(t, "TestVarargReturnHeapIntegrity", "test_vararg_return_heap_integrity", []string{"iterations"}, nil)
		ClassDBBindMethod(t, "TestVarargErrorReturns", "test_vararg_error_returns", []string{"node"}, nil)

		// constants
		ClassDBBindEnumConstant(t, "Example.ExampleEnum", "FIRST", int(ExampleFirst))
		ClassDBBindEnumConstant(t, "Example.ExampleEnum", "ANSWER_TO_EVERYTHING", int(AnswerToEverything))
		ClassDBBindConstant(t, "CONSTANT_WITHOUT_ENUM", int(EXAMPLE_ENUM_CONSTANT_WITHOUT_ENUM))
		ClassDBBindBitfieldFlag(t, "Example.ExampleBitfieldFlag", "FLAG_ONE", int(FlagOne))
		ClassDBBindBitfieldFlag(t, "Example.ExampleBitfieldFlag", "FLAG_TWO", int(FlagTwo))
		ClassDBBindBitfieldFlag(t, "Example.SecondBitfield", "SECOND_FLAG", int(SecondFlag))

		ClassDBBindMethod(t, "TestBitfield", "test_bitfield", []string{"flags"}, nil)

		ClassDBBindMethod(t, "TestStacktraceProbe", "test_stacktrace_probe", []string{"phase"}, nil)

		ClassDBBindMethod(t, "CallableBind", "callable_bind", nil, nil)
		ClassDBBindMethod(t, "TestVariantVector2iConversion", "test_variant_vector2i_conversion", []string{"variant"}, nil)
		// document-binding-substrate witnesses: wrapper identity, inheritance,
		// and the bookkeeping registration leaves behind.
		ClassDBBindMethod(t, "TestObjectCast", "test_object_cast", nil, nil)
		ClassDBBindMethod(t, "TestObjectCasting", "test_object_casting", nil, nil)
		ClassDBBindMethod(t, "TestRegistrationBookkeeping", "test_registration_bookkeeping", nil, nil)

		// others
		// ClassDBBindMethod(t, "TestCastTo", "test_cast_to", nil, nil)
		log.Debug("Example registered")
	})
}

func UnregisterClassExample() {
	ClassDBUnregisterClass[*Example]()
	log.Debug("Example unregistered")
}
