package gdextension

// #include <godot/gdextension_interface.h>
// #include "classdb_callback.h"
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

func ClassDBAddPropertyGroup(t GDClass, p_name string, p_prefix string) {
	cn := t.GetClassName()
	if _, ok := gdRegisteredGDClasses.Get(cn); !ok {
		panic(fmt.Sprintf(`Trying to add property group "%s" to non-existing class "%s".`, p_name, cn))
	}
	className := NewStringNameWithLatin1Chars(cn)
	defer className.Destroy()
	name := NewStringWithUtf8Chars(p_name)
	defer name.Destroy()
	prefix := NewStringWithUtf8Chars(p_prefix)
	defer prefix.Destroy()
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertyGroup(
		FFI.Library,
		className.AsGDExtensionConstStringNamePtr(),
		name.AsGDExtensionConstStringPtr(),
		prefix.AsGDExtensionConstStringPtr(),
	)
}

func ClassDBAddPropertySubgroup(t GDClass, p_name string, p_prefix string) {
	cn := t.GetClassName()
	if _, ok := gdRegisteredGDClasses.Get(cn); !ok {
		panic(fmt.Sprintf(`Trying to add property sub-group "%s" to non-existing class "%s".`, p_name, cn))
	}
	className := NewStringNameWithLatin1Chars(cn)
	defer className.Destroy()
	name := NewStringWithUtf8Chars(p_name)
	defer name.Destroy()
	prefix := NewStringWithUtf8Chars(p_prefix)
	defer prefix.Destroy()
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassPropertySubgroup(
		FFI.Library,
		className.AsGDExtensionConstStringNamePtr(),
		name.AsGDExtensionConstStringPtr(),
		prefix.AsGDExtensionConstStringPtr(),
	)
}

// ClassDBAddProperty default p_index = -1
func ClassDBAddProperty(
	inst GDClass,
	p_property_type GDExtensionVariantType,
	p_property_name string,
	p_setter string,
	p_getter string,
) {
	t := reflect.TypeOf(inst)
	cn := inst.GetClassName()
	pn := p_property_name
	ci, ok := gdRegisteredGDClasses.Get(cn)
	if !ok {
		log.Panic("Trying to add property to non-existing class.",
			zap.String("property", (string)(pn)),
			zap.String("p_class", (string)(cn)),
			zap.Any("type", t),
		)
	}
	if _, ok := ci.PropertyNameMap[pn]; ok {
		panic(fmt.Sprintf(`Property "%s" already exists in class "%s".`, pn, cn))
	}
	var (
		setter *MethodBindImpl
		getter *MethodBindImpl
	)
	if len(p_getter) == 0 {
		log.Panic(`Getter method must be specified.`)
	}
	mci, ok := ci.MethodMap[p_getter]
	if !ok {
		log.Panic("unable to find getter",
			zap.String("getter", p_getter),
		)
	}
	getter = mci.MethodBind
	if len(getter.MethodMetadata.GoArgumentTypes) != 0 {
		panic(fmt.Sprintf(`getter method "%s" must take a single argument.`, p_getter))
	}
	// specifying a setter is optional
	if len(p_setter) > 0 {
		mci, ok := ci.MethodMap[p_setter]
		if !ok {
			log.Panic("unable to find setter",
				zap.String("setter", p_setter),
			)
		}
		setter = mci.MethodBind
		if len(setter.MethodMetadata.GoArgumentTypes) != 1 {
			panic(fmt.Sprintf(`Setter method "%s" must take a single argument.`, p_setter))
		}
	}
	// register property with plugin
	ci.PropertyNameMap[pn] = struct{}{}
	className := NewStringNameWithLatin1Chars(cn)
	defer className.Destroy()
	propName := NewStringNameWithLatin1Chars(pn)
	defer propName.Destroy()
	hint := NewStringWithUtf8Chars("")
	defer hint.Destroy()
	// register with Godot
	prop_info := NewGDExtensionPropertyInfo(
		className.AsGDExtensionConstStringNamePtr(),
		p_property_type,
		propName.AsGDExtensionConstStringNamePtr(),
		uint32(PROPERTY_HINT_NONE),
		hint.AsGDExtensionConstStringPtr(),
		uint32(PROPERTY_USAGE_DEFAULT),
	)
	snSetterGDName := NewStringNameWithLatin1Chars(setter.MethodName)
	defer snSetterGDName.Destroy()
	snGetterGDName := NewStringNameWithLatin1Chars(getter.MethodName)
	defer snGetterGDName.Destroy()
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassProperty(
		FFI.Library,
		ci.NameAsStringNamePtr,
		&prop_info,
		snSetterGDName.AsGDExtensionConstStringNamePtr(),
		snGetterGDName.AsGDExtensionConstStringNamePtr(),
	)
}

type SignalParam struct {
	Type GDExtensionVariantType
	Name string
}

func ClassDBAddSignal(t GDClass, signalName string, params ...SignalParam) {
	log.Debug("ClassDBAddSignal called",
		zap.String("signalName", signalName),
		zap.Any("params", params),
	)
	typeName := t.GetClassName()
	ci, ok := gdRegisteredGDClasses.Get(typeName)
	if !ok {
		log.Panic("Class doesn't exist.", zap.String("class", typeName))
		return
	}
	if _, ok = ci.SignalNameMap[signalName]; ok {
		log.Panic("Constant already registered.", zap.String("class", typeName))
		return
	}
	ci.SignalNameMap[signalName] = struct{}{}
	paramArr := make([]GDExtensionPropertyInfo, len(params))
	for i, p := range params {
		snTypeName := NewStringNameWithLatin1Chars(typeName)
		// defer snTypeName.Destroy()
		snName := NewStringNameWithLatin1Chars(p.Name)
		// defer snName.Destroy()
		hint := NewStringWithUtf8Chars("")
		// defer hint.Destroy()
		paramArr[i] = NewGDExtensionPropertyInfo(
			snTypeName.AsGDExtensionConstStringNamePtr(),
			p.Type,
			snName.AsGDExtensionConstStringNamePtr(),
			(uint32)(PROPERTY_HINT_NONE),
			hint.AsGDExtensionConstStringPtr(),
			(uint32)(PROPERTY_USAGE_DEFAULT),
		)
		defer paramArr[i].Destroy()
	}
	var pi *GDExtensionPropertyInfo
	if len(paramArr) > 0 {
		pi = (*GDExtensionPropertyInfo)(unsafe.Pointer(&paramArr[0]))
	} else {
		pi = (*GDExtensionPropertyInfo)(nullptr)
	}
	snTypeName := NewStringNameWithLatin1Chars(typeName)
	defer snTypeName.Destroy()
	snSignalName := NewStringNameWithLatin1Chars(signalName)
	defer snSignalName.Destroy()
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassSignal(
		FFI.Library,
		snTypeName.AsGDExtensionConstStringNamePtr(),
		snSignalName.AsGDExtensionConstStringNamePtr(),
		pi,
		GDExtensionInt(len(params)))
}

func ClassDBBindMethod(inst GDClass, goMethodName string, gdMethodName string, argNames []string, defaultValues []Variant) *MethodBindImpl {
	return classDBBindMethod(inst, goMethodName, gdMethodName, METHOD_FLAGS_DEFAULT, argNames, defaultValues)
}

func ClassDBBindMethodStatic(inst GDClass, goMethodName string, gdMethodName string, argNames []string, defaultValues []Variant) *MethodBindImpl {
	return classDBBindMethod(inst, goMethodName, gdMethodName, METHOD_FLAG_STATIC, argNames, defaultValues)
}

func ClassDBBindMethodVirtual(inst GDClass, goMethodName string, gdMethodName string, argNames []string, defaultValues []Variant) *MethodBindImpl {
	return classDBBindMethod(inst, goMethodName, gdMethodName, METHOD_FLAG_VIRTUAL, argNames, defaultValues)
}

func ClassDBBindMethodVarargs(
	inst GDClass,
	goMethodName string,
	gdMethodName string,
	argNames []string,
	defaultValues []Variant,
) *MethodBindImpl {
	return classDBBindMethod(inst, goMethodName, gdMethodName, METHOD_FLAG_VARARG, argNames, defaultValues)
}

func classDBBindMethod(
	inst GDClass,
	goMethodName string,
	methodName string,
	methodFlags MethodFlags,
	argNames []string,
	defaultValues []Variant,
) *MethodBindImpl {
	t := reflect.TypeOf(inst)
	className := inst.GetClassName()
	log.Debug("classDBBindMethod called",
		zap.Reflect("inst", inst),
		zap.String("go_name", goMethodName),
		zap.String("gd_name", methodName),
		zap.Any("flags", methodFlags),
		zap.Any("t", t),
		zap.Any("class", className),
	)
	m, ok := t.MethodByName(goMethodName)
	if !ok {
		log.Panic("unable to find function",
			zap.String("gdclass", className),
			zap.String("method_name", goMethodName),
		)
	}
	ptrcallFunc := m.Func
	methodMetadata := NewMethodMetadata(m, className, methodName, argNames, defaultValues, methodFlags)
	if methodMetadata.IsVirtual {
		if !strings.HasPrefix(goMethodName, "V_") {
			log.Panic(`virtual method name must have a prefix of "V_".`)
		}
	} else {
		if strings.HasPrefix(goMethodName, "V_") {
			log.Panic(`method name cannot have a prefix of "V_".`)
		}
	}
	mb := NewMethodBind(
		className,
		methodName,
		goMethodName,
		*methodMetadata,
		ptrcallFunc,
	)
	cmi := NewGDExtensionClassMethodInfoFromMethodBind(mb)
	bi := &MethodBindAndClassMethodInfo{
		MethodBind:      mb,
		ClassMethodInfo: cmi,
	}

	ci, ok := gdRegisteredGDClasses.Get(className)
	if !ok {
		log.Panic("Class doesn't exist.", zap.String("class", className))
		return nil
	}
	if _, ok = ci.MethodMap[methodName]; ok {
		log.Panic("Binding duplicate method.",
			zap.String("go_name", goMethodName),
			zap.String("gd_name", methodName),
		)
		return nil
	}
	if _, ok = ci.VirtualMethodMap[methodName]; ok {
		log.Panic("Method already bound as virtual.",
			zap.String("go_name", goMethodName),
			zap.String("gd_name", methodName),
		)
		return nil
	}
	// register our method bind within our plugin
	if (methodFlags & METHOD_FLAG_VIRTUAL) == METHOD_FLAG_VIRTUAL {
		ci.VirtualMethodMap[methodName] = bi
	} else {
		ci.MethodMap[methodName] = bi
	}
	// and register with godot
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassMethod(
		FFI.Library,
		ci.NameAsStringNamePtr,
		cmi,
	)
	return mb
}

// ClassDBBindConstant binds a constant in godot.
func ClassDBBindConstant(t GDClass, p_constant_name string, p_constant_value int) {
	classDBBindIntegerConstant(t, "", p_constant_name, (GDExtensionInt)(p_constant_value), false)
}

// ClassDBBindEnumConstant binds a enum value in godot.
func ClassDBBindEnumConstant(t GDClass, p_enum_name, p_constant_name string, p_constant_value int) {
	classDBBindIntegerConstant(t, p_enum_name, p_constant_name, (GDExtensionInt)(p_constant_value), false)
}

// ClassDBBindBitfieldFlag binds a bitfield value in godot.
func ClassDBBindBitfieldFlag(t GDClass, p_enum_name, p_constant_name string, p_constant_value int) {
	classDBBindIntegerConstant(t, "", p_constant_name, (GDExtensionInt)(p_constant_value), true)
}

func classDBBindIntegerConstant(t GDClass, p_enum_name, p_constant_name string, p_constant_value GDExtensionInt, p_is_bitfield bool) {
	log.Debug("classDBBindIntegerConstant called",
		zap.String("enum", (string)(p_enum_name)),
		zap.String("constant", (string)(p_constant_name)),
		zap.Any("value", p_constant_value),
		zap.Any("is_bitfield", p_is_bitfield),
	)

	var (
		ci *ClassInfo
		ok bool
	)
	typeName := t.GetClassName()

	if ci, ok = gdRegisteredGDClasses.Get(typeName); !ok {
		log.Panic("Class doesn't exist.", zap.String("class", typeName))
		return
	}

	if _, ok = ci.ConstantNameMap[p_constant_name]; ok {
		log.Panic("Constant already registered.", zap.String("class", typeName))
		return
	}

	ci.ConstantNameMap[p_constant_name] = struct{}{}

	bitfield := (GDExtensionBool)(BoolEncoder.EncodeArg(p_is_bitfield))

	snTypeName := NewStringNameWithLatin1Chars(typeName)
	defer snTypeName.Destroy()

	snEnumName := NewStringNameWithLatin1Chars(p_enum_name)
	defer snEnumName.Destroy()

	snConstantName := NewStringNameWithLatin1Chars(p_constant_name)
	defer snConstantName.Destroy()

	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassIntegerConstant(
		FFI.Library,
		snTypeName.AsGDExtensionConstStringNamePtr(),
		snEnumName.AsGDExtensionConstStringNamePtr(),
		snConstantName.AsGDExtensionConstStringNamePtr(),
		p_constant_value,
		bitfield,
	)
}

func ClassDBRegisterClass(inst GDClass, bindMethodsFunc func(t GDClass)) {
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
	for _, ci := range gdRegisteredGDClasses.Values() {
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
	cl := NewClassInfo(className, parentName, level, classType, inheritType, parentPtr)
	if cl == nil {
		log.Panic("ClassInfo cannot be nil")
	}
	gdRegisteredGDClasses.Set(className, cl)
	if _, ok := gdNativeConstructors.Get(parentName); !ok {
		log.Panic("Missing GDExtensionClass interface: unhandled inherits type", zap.Any("class_type", classType), zap.Any("parent_type", parentName))
	}
	gdClassRegisterInstanceBindingCallbacks(className)
	log.Info("gdclass registered",
		zap.String("class", className),
		zap.String("parent_type", parentName),
	)
	cName := C.CString(className)
	// Register this class with Godot
	info := NewGDExtensionClassCreationInfo2(
		(GDExtensionClassCreateInstance)(C.cgo_classcreationinfo_createinstance),
		(GDExtensionClassFreeInstance)(C.cgo_classcreationinfo_freeinstance),
		(GDExtensionClassGetVirtualCallData)(C.cgo_classcreationinfo_getvirtualcallwithdata),
		(GDExtensionClassCallVirtualWithData)(C.cgo_classcreationinfo_callvirtualwithdata),
		(GDExtensionClassToString)(C.cgo_classcreationinfo_tostring),
		(GDExtensionClassSet)(C.cgo_classcreationinfo_set),
		(GDExtensionClassGet)(C.cgo_classcreationinfo_get),
		(GDExtensionClassGetPropertyList)(C.cgo_classcreationinfo_getpropertylist),
		(GDExtensionClassFreePropertyList)(C.cgo_classcreationinfo_freepropertylist),
		unsafe.Pointer(cName),
	)
	snName := NewStringNameWithLatin1Chars(className)
	defer snName.Destroy()
	snParentName := NewStringNameWithLatin1Chars(parentName)
	defer snParentName.Destroy()
	// register with Godot
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClass2(
		(GDExtensionClassLibraryPtr)(FFI.Library),
		snName.AsGDExtensionConstStringNamePtr(),
		snParentName.AsGDExtensionConstStringNamePtr(),
		&info,
	)
	// call bindMethodsFunc as a callback for users to register their methods on the class
	bindMethodsFunc(inst)
}

func classDBInitialize(pLevel GDExtensionInitializationLevel) {
	for _, ci := range gdRegisteredGDClasses.Values() {
		if ci.Level != pLevel {
			continue
		}
		// Nothing to do here for now...
	}
}

func classDBDeinitialize(pLevel GDExtensionInitializationLevel) {
	for _, ci := range gdRegisteredGDClasses.Values() {
		if ci.Level != pLevel {
			continue
		}
		name := NewStringNameWithLatin1Chars(ci.Name)
		defer name.Destroy()
		CallFunc_GDExtensionInterfaceClassdbUnregisterExtensionClass(
			FFI.Library,
			name.AsGDExtensionConstStringNamePtr(),
		)
		for n, mb := range ci.VirtualMethodMap {
			delete(ci.VirtualMethodMap, n)
			mb.ClassMethodInfo.Destroy()
		}
		for n, mb := range ci.MethodMap {
			delete(ci.MethodMap, n)
			mb.ClassMethodInfo.Destroy()
		}
	}
}
