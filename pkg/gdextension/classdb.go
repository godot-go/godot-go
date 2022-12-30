package gdextension

// #include <godot/gdextension_interface.h>
// #include "classdb_wrapper.h"
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"reflect"
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

	className := NewStringNameWithLatin1Chars(cn)
	defer className.Destroy()

	name := NewStringWithLatin1Chars(p_name)
	defer name.Destroy()

	prefix := NewStringWithLatin1Chars(p_prefix)
	defer prefix.Destroy()

	GDExtensionInterface_classdb_register_extension_class_property_group(
		internal.gdnInterface, internal.library,
		className.AsGDExtensionStringNamePtr(),
		name.AsGDExtensionStringPtr(),
		prefix.AsGDExtensionStringPtr(),
	)
}

func ClassDBAddPropertySubgroup(t GDClass, p_name string, p_prefix string) {
	cn := t.GetClassName()

	if _, ok := gdRegisteredGDClasses.Get(cn); !ok {
		panic(fmt.Sprintf(`Trying to add property sub-group "%s" to non-existing class "%s".`, p_name, cn))
	}

	className := NewStringNameWithLatin1Chars(cn)
	defer className.Destroy()

	name := NewStringWithLatin1Chars(p_name)
	defer name.Destroy()

	prefix := NewStringWithLatin1Chars(p_prefix)
	defer prefix.Destroy()

	GDExtensionInterface_classdb_register_extension_class_property_subgroup(
		internal.gdnInterface, internal.library,
		className.AsGDExtensionStringNamePtr(),
		name.AsGDExtensionStringPtr(),
		prefix.AsGDExtensionStringPtr(),
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

	className := NewStringNameWithLatin1Chars(cn)
	defer className.Destroy()

	propName := NewStringNameWithLatin1Chars(pn)
	defer propName.Destroy()

	hint := NewStringWithLatin1Chars("")
	defer hint.Destroy()

	// register with Godot
	prop_info := NewGDExtensionPropertyInfo(
		className.AsGDExtensionStringNamePtr(),
		p_property_type,
		propName.AsGDExtensionStringNamePtr(),
		uint32(PROPERTY_HINT_NONE),
		hint.AsGDExtensionStringPtr(),
		uint32(PROPERTY_USAGE_DEFAULT),
	)

	snSetterGDName := NewStringNameWithLatin1Chars(setterGDName)
	defer snSetterGDName.Destroy()

	snGetterGDName := NewStringNameWithLatin1Chars(getterGDName)
	defer snGetterGDName.Destroy()

	GDExtensionInterface_classdb_register_extension_class_property(
		internal.gdnInterface, internal.library,
		ci.NameAsStringNamePtr,
		&prop_info,
		snSetterGDName.AsGDExtensionStringNamePtr(),
		snGetterGDName.AsGDExtensionStringNamePtr(),
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

func classDBBindMethodFi(p_flags uint32, p_bind *MethodBind) *MethodBind {
	var (
		ci *ClassInfo
		ok bool
	)
	typeName := p_bind.InstanceClass

	goMethodName := p_bind.GoName

	if ci, ok = gdRegisteredGDClasses.Get(typeName); !ok {
		log.Panic("Class doesn't exist.", zap.String("class", typeName))
		return nil
	}

	if _, ok = ci.MethodMap[goMethodName]; ok {
		log.Panic("Binding duplicate method.", zap.String("name", goMethodName))
		return nil
	}

	if _, ok = ci.VirtualMethodMap[goMethodName]; ok {
		log.Panic("Method already bound as virtual.", zap.String("name", goMethodName))
		return nil
	}

	// register our method bind within our plugin
	ci.MethodMap[goMethodName] = p_bind

	// and register with godot
	classDBBindMethodGodot(ci.NameAsStringNamePtr, p_bind)

	return p_bind
}

func classDBBindMethodGodot(pClassName GDExtensionConstStringNamePtr, pMethod *MethodBind) {
	log.Debug("called C.cgo_callfn_GDExtensionInterface_classdb_register_extension_class_method",
		zap.String("method", pMethod.Name),
		zap.String("instance_class", pMethod.InstanceClass),
		zap.String("go_name", pMethod.GoName),
	)

	GDExtensionInterface_classdb_register_extension_class_method(
		internal.gdnInterface,
		internal.library,
		pClassName,
		&pMethod.ClassMethodInfo,
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
		snTypeName := NewStringNameWithLatin1Chars(typeName)
		defer snTypeName.Destroy()

		snName := NewStringNameWithLatin1Chars(p.Name)
		defer snName.Destroy()

		hint := NewStringWithLatin1Chars("")
		defer hint.Destroy()

		paramArr[i] = NewGDExtensionPropertyInfo(
			snTypeName.AsGDExtensionStringNamePtr(),
			p.Type,
			snName.AsGDExtensionStringNamePtr(),
			(uint32)(PROPERTY_HINT_NONE),
			hint.AsGDExtensionStringPtr(),
			(uint32)(PROPERTY_USAGE_DEFAULT),
		)
		defer paramArr[i].Destroy(internal.gdnInterface)
	}

	var argsPtr *GDExtensionPropertyInfo

	if len(paramArr) > 0 {
		argsPtr = (*GDExtensionPropertyInfo)(unsafe.Pointer(&paramArr[0]))
	} else {
		argsPtr = (*GDExtensionPropertyInfo)(nullptr)
	}

	snTypeName := NewStringNameWithLatin1Chars(typeName)
	defer snTypeName.Destroy()

	snSignalName := NewStringNameWithLatin1Chars(signalName)
	defer snSignalName.Destroy()

	GDExtensionInterface_classdb_register_extension_class_signal(
		internal.gdnInterface,
		internal.library,
		snTypeName.AsGDExtensionStringNamePtr(),
		snSignalName.AsGDExtensionStringNamePtr(),
		argsPtr,
		GDExtensionInt(len(params)))
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

	GDExtensionInterface_classdb_register_extension_class_integer_constant(
		internal.gdnInterface,
		internal.library,
		snTypeName.AsGDExtensionStringNamePtr(),
		snEnumName.AsGDExtensionStringNamePtr(),
		snConstantName.AsGDExtensionStringNamePtr(),
		p_constant_value,
		bitfield,
	)
}

func ClassDBBindMethodVirtual(t GDClass, p_method_name string, p_call GDExtensionClassCallVirtual) {

	cn := t.GetClassName()

	ci, ok := gdRegisteredGDClasses.Get(cn)

	if !ok {
		log.Panic("Class doesn't exist.", zap.String("name", cn))
		return
	}

	if _, ok = ci.MethodMap[p_method_name]; ok {
		log.Panic("Method already registered as non-virtual.", zap.String("name", p_method_name))
		return
	}

	if _, ok = ci.VirtualMethodMap[p_method_name]; ok {
		log.Panic("Virtual method already registered as non-virtual.", zap.String("name", (string)(p_method_name)))
		return
	}

	ci.VirtualMethodMap[p_method_name] = p_call
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

		GDExtensionInterface_classdb_unregister_extension_class(
			internal.gdnInterface,
			internal.library,
			name.AsGDExtensionStringNamePtr(),
		)

		// NOTE: godot-cpp iterates through the map to delete all method binds
		for n, mb := range ci.MethodMap {
			delete(ci.MethodMap, n)
			mb.Destroy(internal.gdnInterface)
		}
	}
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
		(GDExtensionClassCreateInstance)(C.cgo_gdextension_extension_class_create_instance),
		(GDExtensionClassFreeInstance)(C.cgo_gdextension_extension_class_free_instance),
		(GDExtensionClassGetVirtual)(C.cgo_classdb_get_virtual_func),
		unsafe.Pointer(cName),
	)

	snName := NewStringNameWithLatin1Chars(name)
	defer snName.Destroy()

	snParentName := NewStringNameWithLatin1Chars(parentName)
	defer snParentName.Destroy()

	// register with Godot
	GDExtensionInterface_classdb_register_extension_class(
		(*GDExtensionInterface)(internal.gdnInterface),
		(GDExtensionClassLibraryPtr)(internal.library),
		snName.AsGDExtensionStringNamePtr(),
		snParentName.AsGDExtensionStringNamePtr(),
		&info,
	)

	// call bind_methods etc. to register all members of the class
	gdClassInitializeClass(inst)

	bindMethodsFunc(inst)
}

func ClassDBBindMethod(p_instance GDClass, p_go_method_name string, p_method_name string, p_arg_names []string, p_default_values []*Variant) *MethodBind {
	return classDBBindMethod(p_instance, p_go_method_name, p_method_name, METHOD_FLAGS_DEFAULT, p_arg_names, p_default_values)
}

func ClassDBBindMethodStatic(p_instance GDClass, p_go_method_name string, p_method_name string, p_arg_names []string, p_default_values []*Variant) *MethodBind {
	return classDBBindMethod(p_instance, p_go_method_name, p_method_name, METHOD_FLAGS_DEFAULT|METHOD_FLAG_STATIC, p_arg_names, p_default_values)
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

	m, ok := t.MethodByName(p_go_method_name)

	if !ok {
		log.Panic("unable to find function", zap.String("gdclass", n), zap.String("method_name", p_go_method_name))
	}

	mb := NewMethodBind(m, p_method_name, p_arg_names, p_default_values, hintFlags)

	classDBBindMethodFi(0, mb)

	return mb
}

//export GoCallback_ClassDBGetVirtualFunc
func GoCallback_ClassDBGetVirtualFunc(pUserdata unsafe.Pointer, pName *C.char) C.GDExtensionClassCallVirtual {
	className := C.GoString((*C.char)(pUserdata))
	methodName := C.GoString(pName)

	log.Debug("GoCallback_ClassDBGetVirtualFunc called", zap.String("type_name", className), zap.String("method", methodName))

	ci, ok := gdRegisteredGDClasses.Get(className)

	if !ok {
		log.Warn(fmt.Sprintf("class \"%s\" doesn't exist", className))
		return (C.GDExtensionClassCallVirtual)(nullptr)
	}

	m, ok := ci.VirtualMethodMap[methodName]

	if !ok {
		return (C.GDExtensionClassCallVirtual)(nullptr)
	}

	return (C.GDExtensionClassCallVirtual)(m)
}
