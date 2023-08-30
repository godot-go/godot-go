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
	. "github.com/godot-go/godot-go/pkg/util"
	"go.uber.org/zap"
)

func ClassDBAddPropertyGroup(t GDClass, p_name string, p_prefix string) {
	cn := t.GetClassName()

	if _, ok := gdRegisteredGDClasses.Get(cn); !ok {
		panic(fmt.Sprintf(`Trying to add property group "%s" to non-existing class "%s".`, p_name, cn))
	}

	className := NewStringNameWithUtf8Chars(cn)
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

	className := NewStringNameWithUtf8Chars(cn)
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
	p_instance GDClass,
	p_property_type GDExtensionVariantType,
	p_property_name string,
	p_setter string,
	p_getter string,
) {
	t := reflect.TypeOf(p_instance)

	cn := p_instance.GetClassName()

	pn := p_property_name

	pi := -1

	var (
		ci *ClassInfo
		ok bool

		setterGDName string
		getterGDName string
	)

	if ci, ok = gdRegisteredGDClasses.Get(cn); !ok {
		log.Panic("Trying to add property to non-existing class.",
			zap.String("property", (string)(pn)),
			zap.String("p_class", (string)(cn)),
			zap.Any("type", t),
		)
	}

	if _, ok := ci.PropertyNameMap[pn]; ok {
		panic(fmt.Sprintf(`Property "%s" already exists in class "%s".`, pn, cn))
	}

	var setter *MethodBind

	if len(p_setter) > 0 {
		setter = classDBGetMethod(cn, p_setter)

		if setter == nil {
			log.Debug("Setter method not found for property",
				zap.Any("setter", p_setter),
				zap.Any("class", cn),
				zap.Any("property", pn),
			)
		} else {
			expArgs := Iff[int](pi >= 0, 2, 1)

			if expArgs != len(setter.GoArgumentTypes) {
				panic(fmt.Sprintf(`Setter method "%s" must take a single argument.`, p_setter))
			}

			setterGDName = setter.Name
		}
	}

	if len(p_getter) == 0 {
		log.Panic(`Getter method must be specified.`)
	}

	getter := classDBGetMethod(cn, p_getter)

	if len(p_getter) == 0 {
		panic(`Getter method not found for property.`)
	}

	if getter == nil {
		log.Panic("Getter method not found for property",
			zap.Any("setter", p_setter),
			zap.Any("class", cn),
			zap.Any("property", pn),
		)
	} else {
		expArgs := Iff[int](pi >= 0, 1, 0)

		if expArgs != len(getter.GoArgumentTypes) {
			panic(fmt.Sprintf(`Getter method "%s" must not take any argument.`, p_getter))
		}

		getterGDName = getter.Name
	}

	// register property with plugin
	ci.PropertyNameMap[pn] = struct{}{}

	className := NewStringNameWithUtf8Chars(cn)
	defer className.Destroy()

	propName := NewStringNameWithUtf8Chars(pn)
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

	snSetterGDName := NewStringNameWithUtf8Chars(setterGDName)
	defer snSetterGDName.Destroy()

	snGetterGDName := NewStringNameWithUtf8Chars(getterGDName)
	defer snGetterGDName.Destroy()

	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassProperty(
		FFI.Library,
		ci.NameAsStringNamePtr,
		&prop_info,
		snSetterGDName.AsGDExtensionConstStringNamePtr(),
		snGetterGDName.AsGDExtensionConstStringNamePtr(),
	)
}

func classDBGetMethod(p_class string, p_method string) *MethodBind {
	var (
		ci *ClassInfo
		ok bool
	)

	if ci, ok = gdRegisteredGDClasses.Get(p_class); !ok {
		panic(fmt.Sprintf(`Class "%s" not found.`, p_class))
	}

	for ci != nil {
		if method, ok := ci.MethodMap[p_method]; ok {
			return method
		}

		ci = ci.ParentPtr
	}

	return nil
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

	var (
		ci *ClassInfo
		ok bool
	)
	typeName := t.GetClassName()

	if ci, ok = gdRegisteredGDClasses.Get(typeName); !ok {
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
		snTypeName := NewStringNameWithUtf8Chars(typeName)
		// defer snTypeName.Destroy()

		snName := NewStringNameWithUtf8Chars(p.Name)
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

	var argsPtr *GDExtensionPropertyInfo

	if len(paramArr) > 0 {
		argsPtr = (*GDExtensionPropertyInfo)(unsafe.Pointer(&paramArr[0]))
	} else {
		argsPtr = (*GDExtensionPropertyInfo)(nullptr)
	}

	snTypeName := NewStringNameWithUtf8Chars(typeName)
	defer snTypeName.Destroy()

	snSignalName := NewStringNameWithUtf8Chars(signalName)
	defer snSignalName.Destroy()

	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassSignal(
		FFI.Library,
		snTypeName.AsGDExtensionConstStringNamePtr(),
		snSignalName.AsGDExtensionConstStringNamePtr(),
		argsPtr,
		GDExtensionInt(len(params)))
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

		name := NewStringNameWithUtf8Chars(ci.Name)
		defer name.Destroy()

		CallFunc_GDExtensionInterfaceClassdbUnregisterExtensionClass(
			FFI.Library,
			name.AsGDExtensionConstStringNamePtr(),
		)

		// NOTE: godot-cpp iterates through the map to delete all method binds
		for n, mb := range ci.MethodMap {
			delete(ci.MethodMap, n)
			mb.Destroy()
		}
	}
}

func ClassDBBindMethod(p_instance GDClass, p_go_method_name string, p_method_name string, p_arg_names []string, p_default_values []*Variant) *MethodBind {
	return classDBBindMethod(p_instance, p_go_method_name, p_method_name, METHOD_FLAGS_DEFAULT, p_arg_names, p_default_values)
}

func ClassDBBindMethodVarargs(p_instance GDClass, p_go_method_name string, p_method_name string, p_arg_names []string, p_default_values []*Variant) *MethodBind {
	return classDBBindMethod(p_instance, p_go_method_name, p_method_name, METHOD_FLAGS_DEFAULT, p_arg_names, p_default_values)
}

func ClassDBBindMethodStatic(p_instance GDClass, p_go_method_name string, p_method_name string, p_arg_names []string, p_default_values []*Variant) *MethodBind {
	return classDBBindMethod(p_instance, p_go_method_name, p_method_name, METHOD_FLAGS_DEFAULT|METHOD_FLAG_STATIC, p_arg_names, p_default_values)
}

func ClassDBBindMethodVirtual(p_instance GDClass, p_go_method_name string, p_method_name string, p_arg_names []string, p_default_values []*Variant) *MethodBind {
	return classDBBindMethod(p_instance, p_go_method_name, p_method_name, METHOD_FLAGS_DEFAULT|GDEXTENSION_METHOD_FLAG_VIRTUAL, p_arg_names, p_default_values)
}

func classDBBindMethod(p_instance GDClass, p_go_method_name string, p_method_name string, hintFlags MethodFlags, p_arg_names []string, p_default_values []*Variant) *MethodBind {
	t := reflect.TypeOf(p_instance)
	n := p_instance.GetClassName()
	log.Debug("classDBBindMethod called",
		zap.Reflect("inst", p_instance),
		zap.String("go_name", p_go_method_name),
		zap.String("gd_name", p_method_name),
		zap.Any("hint", hintFlags),
		zap.Any("t", t),
		zap.Any("n", n),
	)
	if (hintFlags & GDEXTENSION_METHOD_FLAG_VIRTUAL) == GDEXTENSION_METHOD_FLAG_VIRTUAL {
		if !strings.HasPrefix(p_go_method_name, "V_") {
			log.Panic(`virtual method name must have a prefix of "V_".`)
		}
	} else {
		if strings.HasPrefix(p_go_method_name, "V_") {
			log.Panic(`method name cannot have a prefix of "V_".`)
		}
	}
	m, ok := t.MethodByName(p_go_method_name)
	if !ok {
		log.Panic("unable to find function", zap.String("gdclass", n), zap.String("method_name", p_go_method_name))
	}
	mb := NewMethodBind(m, p_method_name, p_arg_names, p_default_values, hintFlags)
	typeName := mb.InstanceClass
	goMethodName := mb.GoName
	godotMethodName := mb.Name
	ci, ok := gdRegisteredGDClasses.Get(typeName)
	if !ok {
		log.Panic("Class doesn't exist.", zap.String("class", typeName))
		return nil
	}
	if _, ok = ci.MethodMap[godotMethodName]; ok {
		log.Panic("Binding duplicate method.",
			zap.String("go_name", goMethodName),
			zap.String("godot_name", godotMethodName),
		)
		return nil
	}
	if _, ok = ci.VirtualMethodMap[godotMethodName]; ok {
		log.Panic("Method already bound as virtual.",
			zap.String("go_name", goMethodName),
			zap.String("godot_name", godotMethodName),
		)
		return nil
	}
	// register our method bind within our plugin
	if (hintFlags & GDEXTENSION_METHOD_FLAG_VIRTUAL) == GDEXTENSION_METHOD_FLAG_VIRTUAL {
		ci.VirtualMethodMap[godotMethodName] = mb
	} else {
		ci.MethodMap[godotMethodName] = mb
	}
	log.Debug("called C.cgo_callfn_GDExtensionInterface_classdb_register_extension_class_method",
		zap.String("method", mb.Name),
		zap.String("instance_class", mb.InstanceClass),
		zap.String("go_name", mb.GoName),
	)
	// and register with godot
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClassMethod(
		FFI.Library,
		ci.NameAsStringNamePtr,
		&mb.ClassMethodInfo,
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

	snTypeName := NewStringNameWithUtf8Chars(typeName)
	defer snTypeName.Destroy()

	snEnumName := NewStringNameWithUtf8Chars(p_enum_name)
	defer snEnumName.Destroy()

	snConstantName := NewStringNameWithUtf8Chars(p_constant_name)
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

	name := inst.GetClassName()
	parentName := inst.GetParentClassName()

	if name == parentName {
		log.Panic("class and parent cannot have the same name", zap.String("name", name), zap.String("parent", parentName))
	}

	log.Debug("ClassDBRegisterClass called", zap.String("name", name))

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

	if name != classType.Name() {
		log.Panic("GetClassName must match struct name", zap.String("name", name), zap.String("struct_name", classType.Name()))
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

	cl := NewClassInfo(name, parentName, level, classType, inheritType, parentPtr)

	if cl == nil {
		log.Panic("ClassInfo cannot be nil")
	}

	gdRegisteredGDClasses.Set(name, cl)

	if _, ok := gdNativeConstructors.Get(parentName); !ok {
		log.Panic("Missing GDExtensionClass interface: unhandled inherits type", zap.Any("class_type", classType), zap.Any("parent_type", parentName))
	}

	gdClassRegisterInstanceBindingCallbacks(name)

	log.Info("gdclass registered", zap.String("name", name), zap.String("parent_type", parentName))

	cName := C.CString(name)

	// Register this class with Godot
	info := NewGDExtensionClassCreationInfo(
		(GDExtensionClassCreateInstance)(C.cgo_classcreationinfo_createinstance),
		(GDExtensionClassFreeInstance)(C.cgo_classcreationinfo_freeinstance),
		(GDExtensionClassGetVirtuaCallData)(C.cgo_classcreationinfo_getvirtualcallwithdata),
		(GDExtensionClassCallVirtualWithData)(C.cgo_classcreationinfo_callvirtualwithdata),
		(GDExtensionClassToString)(C.cgo_classcreationinfo_tostring),
		unsafe.Pointer(cName),
	)

	snName := NewStringNameWithUtf8Chars(name)
	defer snName.Destroy()

	snParentName := NewStringNameWithUtf8Chars(parentName)
	defer snParentName.Destroy()

	// register with Godot
	CallFunc_GDExtensionInterfaceClassdbRegisterExtensionClass(
		(GDExtensionClassLibraryPtr)(FFI.Library),
		snName.AsGDExtensionConstStringNamePtr(),
		snParentName.AsGDExtensionConstStringNamePtr(),
		&info,
	)

	// call bindMethodsFunc as a callback for users to register their methods on the class
	bindMethodsFunc(inst)
}
