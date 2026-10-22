package core

// #include <godot/gdextension_interface.h>
// #include "classdb_callback.h"
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassinit"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

func ClassDBAddPropertyGroup(t GDClass, p_name string, p_prefix string) {
	cn := t.GetClassName()
	ci, ok := Internal.GDRegisteredGDClasses.Get(cn)
	if !ok {
		panic(fmt.Sprintf(`Trying to add property group "%s" to non-existing class "%s".`, p_name, cn))
	}
	name := NewStringWithUtf8Chars(p_name)
	defer name.Destroy()
	prefix := NewStringWithUtf8Chars(p_prefix)
	defer prefix.Destroy()
	log.Info("register property group",
		zap.String("class", cn),
		zap.String("name", p_name),
		zap.String("prefix", p_prefix),
	)
	cnName := ci.NameStringName
	pnr.Pin(&cnName)
	pnr.Pin(&name)
	pnr.Pin(&prefix)
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertyGroup(
		FFI.Library,
		cnName.AsGDExtensionConstStringNamePtr(),
		name.AsGDExtensionConstStringPtr(),
		prefix.AsGDExtensionConstStringPtr(),
	)
}

func ClassDBAddPropertySubgroup(t GDClass, p_name string, p_prefix string) {
	cn := t.GetClassName()
	ci, ok := Internal.GDRegisteredGDClasses.Get(cn)
	if !ok {
		panic(fmt.Sprintf(`Trying to add property sub-group "%s" to non-existing class "%s".`, p_name, cn))
	}
	name := NewStringWithUtf8Chars(p_name)
	defer name.Destroy()
	prefix := NewStringWithUtf8Chars(p_prefix)
	defer prefix.Destroy()
	log.Info("register property sub-group",
		zap.String("class", cn),
		zap.String("name", p_name),
		zap.String("prefix", p_prefix),
	)
	cnName := ci.NameStringName
	pnr.Pin(&cnName)
	pnr.Pin(&name)
	pnr.Pin(&prefix)
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertySubgroup(
		FFI.Library,
		cnName.AsGDExtensionConstStringNamePtr(),
		name.AsGDExtensionConstStringPtr(),
		prefix.AsGDExtensionConstStringPtr(),
	)
}

// ClassDBAddProperty registers a property with default metadata: hint
// PROPERTY_HINT_NONE, an empty hint string, and PROPERTY_USAGE_DEFAULT.
// Use ClassDBAddPropertyInfo for inspector hints, usage flags, and class
// qualifiers.
func ClassDBAddProperty(
	inst GDClass,
	p_property_type GDExtensionVariantType,
	p_property_name string,
	p_setter string,
	p_getter string,
) {
	ClassDBAddPropertyInfo(inst, PropertySpec{
		VariantType: p_property_type,
		Name:        p_property_name,
		Setter:      p_setter,
		Getter:      p_getter,
	})
}

type SignalParam struct {
	Type GDExtensionVariantType
	Name string
}

// classDBSignalProbe returns a probe answering ClassDB.class_has_signal for
// className — the engine's inheritance-inclusive member set, the same set the
// engine's own add_signal guard consults, so the answer cannot disagree with
// the registration it protects. The second result reports whether the engine
// could be consulted at all: a missing ClassDB singleton or method bind
// yields (false, false) so the caller can degrade to the Go-side record. The
// method hash 471820014 is ClassDB.class_has_signal per extension_api.json,
// matching the generated ClassDBImpl.ClassHasSignal call body.
func classDBSignalProbe(className string) func(signalName string) (carries bool, available bool) {
	return func(signalName string) (bool, bool) {
		singleton := GetSingleton("ClassDB")
		if singleton == nil {
			return false, false
		}
		var pinner runtime.Pinner
		defer pinner.Unpin()

		tableName := NewStringNameWithLatin1Chars("ClassDB")
		defer tableName.Destroy()
		methodName := NewStringNameWithLatin1Chars("class_has_signal")
		defer methodName.Destroy()
		fn := CallFunc_GDExtensionInterfaceClassdbGetMethodBind(
			tableName.AsGDExtensionConstStringNamePtrPinned(&pinner),
			methodName.AsGDExtensionConstStringNamePtrPinned(&pinner),
			471820014,
		)
		if fn == nil {
			return false, false
		}

		classSn := NewStringNameWithLatin1Chars(className)
		defer classSn.Destroy()
		signalSn := NewStringNameWithLatin1Chars(signalName)
		defer signalSn.Destroy()

		var ret bool
		retPtr := (GDExtensionTypePtr)(unsafe.Pointer(&ret))
		pinner.Pin(retPtr)
		argPtrSlice := make([]GDExtensionConstTypePtr, 2)
		cArgs := (*GDExtensionConstTypePtr)(unsafe.SliceData(argPtrSlice))
		pinner.Pin(cArgs)
		argPtrSlice[0] = (GDExtensionConstTypePtr)(unsafe.Pointer(&classSn))
		pinner.Pin(argPtrSlice[0])
		argPtrSlice[1] = (GDExtensionConstTypePtr)(unsafe.Pointer(&signalSn))
		pinner.Pin(argPtrSlice[1])
		CallFunc_GDExtensionInterfaceObjectMethodBindPtrcall(fn, singleton, cArgs, retPtr)
		return ret, true
	}
}

// signalNameCollision reports whether the class behind ci already carries
// signalName: either a name the class itself recorded, or — when the engine
// could be consulted — any name the engine says the class carries, which
// includes names inherited from ancestors. The second result reports whether
// the answer is final: a recorded collision is decisive without the engine,
// and an engine-consulted answer is final. False means the decision degraded
// to the recorded names — inheritance was not checked — and the caller warns.
func signalNameCollision(ci *ClassInfo, signalName string, probe func(signalName string) (carries bool, available bool)) (collision bool, decisive bool) {
	if _, ok := ci.SignalNameSet[signalName]; ok {
		return true, true
	}
	if probe != nil {
		if carries, available := probe(signalName); available {
			return carries, true
		}
	}
	return false, false
}

func ClassDBAddSignal(t GDClass, signalName string, params ...SignalParam) {
	log.Debug("ClassDBAddSignal called",
		zap.String("signalName", signalName),
		zap.Any("params", params),
	)
	typeName := t.GetClassName()
	ci, ok := Internal.GDRegisteredGDClasses.Get(typeName)
	if !ok {
		log.Panic("Class doesn't exist.", zap.String("class", typeName))
		return
	}
	collision, decisive := signalNameCollision(ci, signalName, classDBSignalProbe(typeName))
	if collision {
		log.Panic(fmt.Sprintf("Signal already registered: class %q already carries signal %q", typeName, signalName),
			zap.String("class", typeName),
			zap.String("signal", signalName),
		)
		return
	}
	if !decisive {
		log.Warn("signal duplicate check consulted recorded names only; inheritance was not checked",
			zap.String("class", typeName),
			zap.String("signal", signalName),
		)
	}
	ci.SignalNameSet[signalName] = struct{}{}

	// Build temporary StringName/String objects for signal parameter PropertyInfo.
	// Godot copies the pointer data from the PropertyInfo array, so these can be
	// destroyed immediately after the registration call.
	paramSnNames := make([]StringName, len(params))
	paramHintStrings := make([]String, len(params))
	paramArr := make([]GDExtensionPropertyInfo, len(params))
	for i, p := range params {
		paramSnNames[i] = NewStringNameWithLatin1Chars(p.Name)
		paramHintStrings[i] = NewStringWithUtf8Chars("")
		paramArr[i] = NewGDExtensionPropertyInfo(
			ci.NameStringName.AsGDExtensionConstStringNamePtr(),
			p.Type,
			paramSnNames[i].AsGDExtensionConstStringNamePtr(),
			(uint32)(PROPERTY_HINT_NONE),
			paramHintStrings[i].AsGDExtensionConstStringPtr(),
			(uint32)(PROPERTY_USAGE_DEFAULT),
		)
	}

	var pi *GDExtensionPropertyInfo
	if len(paramArr) > 0 {
		pi = (*GDExtensionPropertyInfo)(unsafe.Pointer(&paramArr[0]))
		pnr.Pin(&paramArr)
	} else {
		pi = (*GDExtensionPropertyInfo)(nullptr)
	}

	// Store the signal name StringName in ClassInfo for lifecycle management.
	// Godot retains a reference to this StringName in its ClassDB, so we must
	// not destroy it until the class is unregistered (handled in ClassInfo.Destroy()).
	snSignalName := NewStringNameWithLatin1Chars(signalName)
	ci.SignalNameStringNames[signalName] = snSignalName
	log.Info("register signal",
		zap.String("class", typeName),
		zap.String("name", signalName),
	)
	cnName := ci.NameStringName
	pnr.Pin(&cnName)
	pnr.Pin(&snSignalName)
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassSignal(
		FFI.Library,
		cnName.AsGDExtensionConstStringNamePtr(),
		snSignalName.AsGDExtensionConstStringNamePtr(),
		pi,
		GDExtensionInt(len(params)))

	// Destroy temporary signal parameter StringNames and hint strings.
	// Godot already copied the pointer data from the PropertyInfo array above.
	for i := range paramSnNames {
		paramSnNames[i].Destroy()
		paramHintStrings[i].Destroy()
	}
}

func ClassDBBindMethod[T GDClass](
	inst T,
	goMethodName string,
	gdMethodName string,
	argNames []string,
	defaultValues []Variant,
) {
	classDBBindMethod(inst, goMethodName, gdMethodName, METHOD_FLAGS_DEFAULT, argNames, defaultValues)
}

// ClassDBBindStaticMethod binds a Go package function as a static method of
// T's registered class. inst is used only to resolve the class name; the
// function is registered with METHOD_FLAG_STATIC, so GDScript calls it
// through the class path (MyClass.some_static(...)) with no instance, and
// the engine dispatches it with a null instance pointer. fn must be a Go
// function; its parameters (no receiver) become the method arguments.
func ClassDBBindStaticMethod[T GDClass](
	inst T,
	gdMethodName string,
	fn any,
	argNames []string,
	defaultValues []Variant,
) {
	classDBBindStaticMethod(inst, gdMethodName, METHOD_FLAG_STATIC, fn, argNames, defaultValues)
}

func ClassDBBindMethodVirtual[T GDClass](
	inst T,
	goMethodName string,
	gdMethodName string,
	argNames []string,
	defaultValues []Variant,
) {
	classDBBindMethod(inst, goMethodName, gdMethodName, METHOD_FLAG_VIRTUAL, argNames, defaultValues)
}

// pascalFromGdVirtual converts a GDScript virtual name to its Go convention
// segment: "_ready" -> "Ready", "to_string" -> "ToString",
// "_property_can_revert" -> "PropertyCanRevert".
func pascalFromGdVirtual(gdMethodName string) string {
	trimmed := strings.TrimPrefix(gdMethodName, "_")
	parts := strings.Split(trimmed, "_")
	var sb strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		sb.WriteString(strings.ToUpper(part[:1]))
		sb.WriteString(part[1:])
	}
	return sb.String()
}

// virtualEmbeddedChainTypes returns the receiver type's embedding chain,
// most-derived first: the type itself (deref'd), then each anonymous field's
// chain in declaration order.
func virtualEmbeddedChainTypes(t reflect.Type) []reflect.Type {
	if t == nil {
		return nil
	}
	typ := t
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return []reflect.Type{typ}
	}
	types := []reflect.Type{typ}
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		if !field.Anonymous {
			continue
		}
		types = append(types, virtualEmbeddedChainTypes(field.Type)...)
	}
	return types
}

// resolveQualifiedVirtualMethod finds the most-derived implementation of a
// Godot virtual along T's embedding chain. Each level contributes one
// candidate named V_<LevelTypeName>_<PascalGdName>; because qualified names
// embed their declaring type's name they cannot collide across levels, so
// querying the outermost promoted method set per candidate (levels ordered
// most-derived first) yields the nearest implementation.
func resolveQualifiedVirtualMethod(t reflect.Type, gdMethodName string) (m reflect.Method, ok bool, searched []string) {
	pascal := pascalFromGdVirtual(gdMethodName)
	chainTypes := virtualEmbeddedChainTypes(t)
	searched = make([]string, 0, len(chainTypes))
	for _, level := range chainTypes {
		candidate := "V_" + level.Name() + "_" + pascal
		searched = append(searched, candidate)
		if found, foundOk := t.MethodByName(candidate); foundOk {
			return found, true, searched
		}
	}
	return reflect.Method{}, false, searched
}

func ClassDBBindMethodVarargs[T GDClass](
	inst T,
	goMethodName string,
	gdMethodName string,
	argNames []string,
	defaultValues []Variant,
) {
	classDBBindMethod(inst, goMethodName, gdMethodName, METHOD_FLAG_VARARG, argNames, defaultValues)
}

// ensureMethodNameAvailable is the pre-allocation duplicate guard: it
// resolves the class registration and refuses a gdMethodName that is already
// taken by a method or virtual bind. Both public bind paths call it before
// constructing any engine-dependent metadata, so a rejected bind allocates
// nothing to leak; classDBRegisterMethod keeps its own guards as the
// authoritative last check before touching the engine.
func ensureMethodNameAvailable(className string, gdMethodName string, goMethodName string) {
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
	if !ok {
		log.Panic("Class doesn't exist.", zap.String("class", className))
		return
	}
	if err := ci.HasError(); err != nil {
		log.Panic("ClassInfo not valid",
			zap.String("class", className),
			zap.Error(err),
			zap.String("go_name", goMethodName),
			zap.String("gd_name", gdMethodName),
		)
		return
	}
	if _, ok := ci.MethodMap[gdMethodName]; ok {
		log.Panic("Binding duplicate method.",
			zap.String("go_name", goMethodName),
			zap.String("gd_name", gdMethodName),
		)
		return
	}
	if _, ok := ci.VirtualMethodMap[gdMethodName]; ok {
		log.Panic("Method already bound as virtual.",
			zap.String("go_name", goMethodName),
			zap.String("gd_name", gdMethodName),
		)
		return
	}
}

func classDBBindMethod[T GDClass](
	inst T,
	goMethodName string,
	gdMethodName string,
	methodFlags MethodFlags,
	argNames []string,
	defaultValues []Variant,
) {
	t := reflect.TypeFor[T]()
	className := inst.GetClassName()
	isVirtualBind := (methodFlags & METHOD_FLAG_VIRTUAL) == METHOD_FLAG_VIRTUAL
	log.Debug("classDBBindMethod called",
		zap.Any("inst", inst),
		zap.String("go_name", goMethodName),
		zap.String("gd_name", gdMethodName),
		zap.Any("flags", methodFlags),
		zap.Any("t", t),
		zap.Any("class", className),
	)
	var m reflect.Method
	if isVirtualBind {
		// Clean break: the deprecated flat V_<Method> form must not register.
		if strings.HasPrefix(goMethodName, "V_") &&
			!strings.Contains(strings.TrimPrefix(goMethodName, "V_"), "_") {
			log.Panic("flat virtual method name is no longer supported; use the qualified form",
				zap.String("gd_method_name", gdMethodName),
				zap.String("flat_go_name", goMethodName),
				zap.String("expected_shape", "V_<ClassName>_"+pascalFromGdVirtual(gdMethodName)),
			)
		}
		resolved, ok, searched := resolveQualifiedVirtualMethod(t, gdMethodName)
		if !ok {
			log.Panic("no qualified virtual implementation found along the embedding chain",
				zap.String("gd_method_name", gdMethodName),
				zap.String("searched_names", strings.Join(searched, ", ")),
				zap.String("expected_shape", "V_<ClassName>_"+pascalFromGdVirtual(gdMethodName)),
			)
		}
		m = resolved
		goMethodName = m.Name
	} else {
		var ok bool
		m, ok = t.MethodByName(goMethodName)
		if !ok {
			log.Panic("unable to find function",
				zap.String("gdclass", className),
				zap.String("method_name", goMethodName),
				zap.String("gd_method_name", gdMethodName),
			)
		}
	}
	log.Debug("method found",
		zap.Reflect("method", m),
	)
	// Reject duplicate names before any engine-dependent metadata is
	// allocated: a refused bind must leak nothing (StringNames included).
	ensureMethodNameAvailable(className, gdMethodName, goMethodName)
	md := NewGoMethodMetadata(m, className, gdMethodName, goMethodName, argNames, defaultValues, methodFlags)
	if md.IsVirtual {
		if !strings.HasPrefix(goMethodName, "V_") {
			log.Panic(`virtual method name must have a prefix of "V_".`)
		}
	} else {
		if strings.HasPrefix(goMethodName, "V_") {
			log.Panic(`method name cannot have a prefix of "V_".`)
		}
	}
	classDBRegisterMethod(md, goMethodName)
}

// classDBBindStaticMethod is the internal path behind ClassDBBindStaticMethod.
// Unlike classDBBindMethod there is no method-name resolution step: the
// callable is the package function itself, reflected without a receiver.
func classDBBindStaticMethod[T GDClass](
	inst T,
	gdMethodName string,
	methodFlags MethodFlags,
	fn any,
	argNames []string,
	defaultValues []Variant,
) {
	className := inst.GetClassName()
	log.Debug("classDBBindStaticMethod called",
		zap.Any("inst", inst),
		zap.String("gd_name", gdMethodName),
		zap.Any("flags", methodFlags),
		zap.Any("class", className),
	)
	// A variadic package function implies the engine's vararg flag; deriving
	// it from the function's own signature keeps the flag/function alignment
	// gate in validateBindSignature unfalsifiable by callers.
	// Reject duplicate names before any engine-dependent metadata is
	// allocated: a refused bind must leak nothing (StringNames included).
	ensureMethodNameAvailable(className, gdMethodName, "")
	if fv := reflect.ValueOf(fn); fv.IsValid() && fv.Kind() == reflect.Func && fv.Type().IsVariadic() {
		methodFlags |= METHOD_FLAG_VARARG
	}
	md := newGoMethodMetadataStatic(fn, className, gdMethodName, argNames, defaultValues, methodFlags)
	classDBRegisterMethod(md, md.GoMethodName)
}

// classDBRegisterMethod is the shared registration tail for every class
// method bind (instance, virtual, static): build the class method info,
// reject duplicate names in the per-class registries, record the bind, and
// hand it to the engine under the class name StringName.
func classDBRegisterMethod(md *GoMethodMetadata, goMethodName string) {
	className := md.ClassName
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
	if !ok {
		log.Panic("Class doesn't exist.", zap.String("class", className))
		return
	}
	if err := ci.HasError(); err != nil {
		log.Panic("ClassInfo not valid",
			zap.String("class", className),
			zap.Error(err),
			zap.String("go_name", goMethodName),
			zap.String("gd_name", md.GdMethodName),
		)
		return
	}
	if _, ok = ci.MethodMap[md.GdMethodName]; ok {
		log.Panic("Binding duplicate method.",
			zap.String("go_name", goMethodName),
			zap.String("gd_name", md.GdMethodName),
		)
		return
	}
	if _, ok = ci.VirtualMethodMap[md.GdMethodName]; ok {
		log.Panic("Method already bound as virtual.",
			zap.String("go_name", goMethodName),
			zap.String("gd_name", md.GdMethodName),
		)
		return
	}
	cmi := NewGDExtensionClassMethodInfoFromMethodBind(md)
	bi := &MethodBindAndClassMethodInfo{
		GoMethodMetadata: md,
		ClassMethodInfo:  cmi,
	}
	hasVarargs := (md.MethodFlags & METHOD_FLAG_VARARG) == METHOD_FLAG_VARARG
	// keep track of the method
	if (md.MethodFlags & METHOD_FLAG_VIRTUAL) == METHOD_FLAG_VIRTUAL {
		ci.VirtualMethodMap[md.GdMethodName] = bi
		log.Info("register class virtual method",
			zap.String("class", className),
			zap.String("go_name", goMethodName),
			zap.String("gd_name", md.GdMethodName),
			zap.String("bind", bi.GoMethodMetadata.String()),
			zap.Bool("has_varargs", hasVarargs),
		)
	} else {
		ci.MethodMap[md.GdMethodName] = bi
		log.Info("register class method",
			zap.String("class", className),
			zap.String("go_name", goMethodName),
			zap.String("gd_name", md.GdMethodName),
			zap.String("bind", bi.GoMethodMetadata.String()),
			zap.Bool("has_varargs", hasVarargs),
		)
	}
	// and register with godot
	cnName := ci.NameStringName
	pnr.Pin(&cnName)
	pnr.Pin(cmi)
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassMethod(
		FFI.Library,
		cnName.AsGDExtensionConstStringNamePtr(),
		cmi,
	)
}

// ClassDBBindConstant binds a constant in godot.
func ClassDBBindConstant(t GDClass, p_constant_name string, p_constant_value int) {
	classDBBindIntegerConstant(t, "", p_constant_name, (GDExtensionInt)(p_constant_value), false)
}

// ClassDBBindEnumConstant binds a enum value in godot.
func ClassDBBindEnumConstant(t GDClass, p_enum_name, p_constant_name string, p_constant_value int) {
	classDBBindIntegerConstant(t, p_enum_name, p_constant_name, (GDExtensionInt)(p_constant_value), false)
}

// ClassDBBindBitfieldFlag binds a bitfield value in godot. The flag is bound
// under p_enum_name so the enum member exists and records bitfield-ness. A
// name qualified as "Class.EnumName" is accepted: the engine registers the
// enum under the portion following its first separator.
func ClassDBBindBitfieldFlag(t GDClass, p_enum_name, p_constant_name string, p_constant_value int) {
	classDBBindIntegerConstant(t, p_enum_name, p_constant_name, (GDExtensionInt)(p_constant_value), true)
}

func classDBBindIntegerConstant(t GDClass, p_enum_name, p_constant_name string, p_constant_value GDExtensionInt, p_is_bitfield bool) {
	log.Debug("classDBBindIntegerConstant called",
		zap.String("enum", p_enum_name),
		zap.String("constant", p_constant_name),
		zap.Any("value", p_constant_value),
		zap.Any("is_bitfield", p_is_bitfield),
	)
	var (
		ci *ClassInfo
		ok bool
	)
	typeName := t.GetClassName()
	if ci, ok = Internal.GDRegisteredGDClasses.Get(typeName); !ok {
		log.Panic("Class doesn't exist.", zap.String("class", typeName))
		return
	}
	if _, ok = ci.ConstantNameSet[p_constant_name]; ok {
		log.Panic(fmt.Sprintf("Constant already registered: class %q already binds constant %q", typeName, p_constant_name),
			zap.String("class", typeName),
			zap.String("constant", p_constant_name),
		)
		return
	}
	ci.ConstantNameSet[p_constant_name] = struct{}{}
	var bitfield GDExtensionBool
	if p_is_bitfield {
		bitfield = (GDExtensionBool)(1)
	} else {
		bitfield = (GDExtensionBool)(0)
	}
	snTypeName := NewStringNameWithLatin1Chars(typeName)
	defer snTypeName.Destroy()
	snEnumName := NewStringNameWithLatin1Chars(p_enum_name)
	defer snEnumName.Destroy()
	snConstantName := NewStringNameWithLatin1Chars(p_constant_name)
	defer snConstantName.Destroy()
	log.Info("register int constant",
		zap.String("type", snTypeName.ToUtf8()),
		zap.String("enum", snEnumName.ToUtf8()),
		zap.String("const", snConstantName.ToUtf8()),
		zap.Int("value", (int)(p_constant_value)),
	)
	pnr.Pin(&snTypeName)
	pnr.Pin(&snEnumName)
	pnr.Pin(&snConstantName)
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassIntegerConstant(
		FFI.Library,
		snTypeName.AsGDExtensionConstStringNamePtr(),
		snEnumName.AsGDExtensionConstStringNamePtr(),
		snConstantName.AsGDExtensionConstStringNamePtr(),
		p_constant_value,
		bitfield,
	)
}

func ClassDBRegisterClass[T Object](
	constructor GDClassGoConstructorFromOwner,
	propertyList []PropertyDescriptor,
	validateProperty func(*GDExtensionPropertyInfo),
	bindMethodsFunc func(t T),
) {
	t := reflect.TypeFor[T]()
	objectInst := reflect.Zero(t).Interface().(Object)
	inst := objectInst.(T)

	// Register this class within our plugin
	className := inst.GetClassName()
	parentName := inst.GetParentClassName()
	if className == parentName {
		log.Panic("class and parent cannot have the same name",
			zap.String("class", className),
			zap.String("parent", parentName),
		)
	}
	log.Debug("ClassDBRegisterClass called",
		zap.String("class", className),
	)
	level := classdbCurrentLevel
	var parentPtr *ClassInfo
	for _, ci := range Internal.GDRegisteredGDClasses.Values() {
		if ci.Name == parentName {
			parentPtr = ci
			break
		}
	}
	if parentPtr == nil {
		log.Debug("parent not found in classdb", zap.String("parentName", (string)(parentName)))
	}
	classType := reflect.TypeOf(inst)
	if classType == nil {
		log.Panic("Type cannot be nil")
	}
	if classType.Kind() == reflect.Ptr {
		classType = classType.Elem()
	}
	if className != classType.Name() {
		log.Panic("GetClassName must match struct name",
			zap.String("class", className),
			zap.String("struct_name", classType.Name()),
		)
	}
	vf := reflect.VisibleFields(classType)
	if len(vf) == 0 {
		log.Panic("Missing GDExtensionClass interface: no visible struct fields")
	}
	// need to ensure the GDExtensionClass is always the first struct
	inheritType := vf[0].Type
	if inheritType == nil {
		log.Panic("Missing GDExtensionClass interface: inherits type nil")
	}
	if fmt.Sprintf("%sImpl", parentName) != inheritType.Name() {
		log.Panic("GetParentClassName must match struct name", zap.String("parent_name", parentName), zap.String("struct_inherit_type", inheritType.Name()))
	}
	cl := NewClassInfo(className, parentName, level, classType, inheritType, parentPtr, propertyList, validateProperty)
	if cl == nil {
		log.Panic("ClassInfo cannot be nil")
	}
	Internal.GDRegisteredGDClasses.Set(className, cl)
	if _, ok := GDNativeConstructors.Get(parentName); !ok {
		log.Panic("Missing GDExtensionClass interface: unhandled inherits type", zap.Any("class_type", classType), zap.Any("parent_type", parentName))
	}
	Internal.GDClassConstructors.Set(className, constructor)
	GDRegisteredGDClassEncoders.Set(className, CreateObjectEncoder[T]())
	GDClassRegisterInstanceBindingCallbacks(className)
	cName := C.CString(className)
	// Register this class with Godot
	info := NewGDExtensionClassCreationInfo4(
		GDExtensionBool(0),
		GDExtensionBool(0),
		GDExtensionBool(1),
		GDExtensionBool(0),
		(GDExtensionConstStringPtr)(nil),
		(GDExtensionClassSet)(C.cgo_classcreationinfo_set),
		(GDExtensionClassGet)(C.cgo_classcreationinfo_get),
		(GDExtensionClassGetPropertyList)(C.cgo_classcreationinfo_getpropertylist),
		(GDExtensionClassFreePropertyList2)(C.cgo_classcreationinfo_freepropertylist2),
		(GDExtensionClassPropertyCanRevert)(C.cgo_classcreationinfo_propertycanrevert),
		(GDExtensionClassPropertyGetRevert)(C.cgo_classcreationinfo_propertygetrevert),
		(GDExtensionClassValidateProperty)(C.cgo_classcreationinfo_validateproperty),
		(GDExtensionClassNotification2)(C.cgo_classcreationinfo_notification),
		(GDExtensionClassToString)(C.cgo_classcreationinfo_tostring),
		(GDExtensionClassReference)(nil),
		(GDExtensionClassUnreference)(nil),
		(GDExtensionClassCreateInstance2)(C.cgo_classcreationinfo_createinstance2),
		(GDExtensionClassFreeInstance)(C.cgo_classcreationinfo_freeinstance),
		(GDExtensionClassRecreateInstance)(nil),
		(GDExtensionClassGetVirtualCallData2)(C.cgo_classcreationinfo_getvirtualcallwithdata2),
		(GDExtensionClassCallVirtualWithData)(C.cgo_classcreationinfo_callvirtualwithdata),
		unsafe.Pointer(cName),
	)
	snName := NewStringNameWithLatin1Chars(className)
	defer snName.Destroy()
	snParentName := NewStringNameWithLatin1Chars(parentName)
	defer snParentName.Destroy()
	log.Info("gdclass registered",
		zap.String("class", className),
		zap.String("parent_type", parentName),
	)
	// register with Godot
	pnr.Pin(&snName)
	pnr.Pin(&snParentName)
	pnr.Pin(&info)
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClass4(
		(GDExtensionClassLibraryPtr)(FFI.Library),
		snName.AsGDExtensionConstStringNamePtr(),
		snParentName.AsGDExtensionConstStringNamePtr(),
		&info,
	)
	// call bindMethodsFunc as a callback for users to register their methods on the class
	bindMethodsFunc(inst)
}

func ClassDBUnregisterClass[T Object]() {
	t := reflect.TypeFor[T]()
	objectInst := reflect.Zero(t).Interface().(Object)
	inst := objectInst.(T)
	className := inst.GetClassName()
	log.Info("ClassDBUnregisterClass called",
		zap.String("class", className),
	)
	cl, ok := Internal.GDRegisteredGDClasses.Get(className)
	if !ok {
		log.Panic("Class doesn't exist.", zap.String("class", className))
		return
	}
	Internal.GDRegisteredGDClasses.Delete(className)
	name := NewStringNameWithLatin1Chars(className)
	defer name.Destroy()
	pnr.Pin(&name)
	CallFunc_GDExtensionInterfaceClassdbUnregisterExtensionClass(
		FFI.Library,
		name.AsGDExtensionConstStringNamePtr(),
	)
	for _, mb := range cl.VirtualMethodMap {
		// Pin metadata before Destroy to satisfy cgo "Go pointer to unpinned Go pointer" check.
		// GoMethodMetadata contains reflect.Value (Go pointer) which must be pinned when
		// passing StringName fields to cgo functions.
		pnr.Pin(&mb.GoMethodMetadata)
		mb.GoMethodMetadata.Destroy()
	}
	cl.VirtualMethodMap = nil
	for _, mb := range cl.MethodMap {
		pnr.Pin(&mb.GoMethodMetadata)
		mb.GoMethodMetadata.Destroy()
	}
	cl.MethodMap = nil
	cl.Destroy()
}
