package gdextension

// #include <godot/gdnative_interface.h>
// #include "classdb_wrapper.h"
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdnative"
	"github.com/godot-go/godot-go/pkg/log"
	. "github.com/godot-go/godot-go/pkg/util"
	"go.uber.org/zap"
)

func ClassDBAddPropertyGroup(t GDClass, p_name string, p_prefix string) {
	cn := t.GetClassName()

	if _, ok := gdRegisteredGDClasses.Get(cn); !ok {
		panic(fmt.Sprintf(`Trying to add property group "%s" to non-existing class "%s".`, p_name, cn))
	}

	GDNativeInterface_classdb_register_extension_class_property_group(
		internal.gdnInterface, internal.library, (string)(cn), (string)(p_name), p_prefix)
}

func ClassDBAddPropertySubgroup(t GDClass, p_name string, p_prefix string) {
	cn := t.GetClassName()

	if _, ok := gdRegisteredGDClasses.Get(cn); !ok {
		panic(fmt.Sprintf(`Trying to add property sub-group "%s" to non-existing class "%s".`, p_name, cn))
	}

	GDNativeInterface_classdb_register_extension_class_property_subgroup(
		internal.gdnInterface, internal.library, (string)(cn), (string)(p_name), p_prefix)
}

// ClassDBAddProperty default p_index = -1
func ClassDBAddProperty(
	p_instance GDClass,
	p_property_type GDNativeVariantType,
	p_property_name PropertyName,
	p_setter MethodName,
	p_getter MethodName,
) {
	t := reflect.TypeOf(p_instance)

	cn := p_instance.GetClassName()

	pn := p_property_name

	pi := -1

	var (
		ci *ClassInfo
		ok bool

		setterGDName MethodName
		getterGDName MethodName
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
			exp_args := (uint32)(Iff(pi >= 0, 2, 1))

			if exp_args != setter.ArgumentCount {
				panic(fmt.Sprintf(`Setter method "%s" must take a single argument.`, p_setter))
			}

			setterGDName = setter.GDName
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
		exp_args := (uint32)(Iff(pi >= 0, 1, 0))

		if exp_args != getter.ArgumentCount {
			panic(fmt.Sprintf(`Getter method "%s" must not take any argument.`, p_getter))
		}

		getterGDName = getter.GDName
	}

	// register property with plugin
	ci.PropertyNameMap[pn] = struct{}{}

	// register with Godot
	prop_info := NewGDNativePropertyInfo(
		string(cn),
		p_property_type,
		string(pn),
		uint32(PROPERTY_HINT_NONE),
		"",
		uint32(PROPERTY_USAGE_DEFAULT),
	)

	GDNativeInterface_classdb_register_extension_class_property(internal.gdnInterface, internal.library, (string)(ci.Name), &prop_info, (string)(setterGDName), (string)(getterGDName))

	// TODO: release CStrings?
}

func classDBGetMethod(p_class TypeName, p_method MethodName) *MethodBind {
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
		log.Panic("Class doesn't exist.", zap.String("class", (string)(typeName)))
		return nil
	}

	if _, ok = ci.MethodMap[goMethodName]; ok {
		log.Panic("Binding duplicate method.", zap.String("name", (string)(goMethodName)))
		return nil
	}

	if _, ok = ci.VirtualMethodMap[goMethodName]; ok {
		log.Panic("Method already bound as virtual.", zap.String("name", (string)(goMethodName)))
		return nil
	}

	// register our method bind within our plugin
	ci.MethodMap[goMethodName] = p_bind

	// and register with godot
	classDBBindMethodGodot(ci.Name, p_bind)

	return p_bind
}

func classDBBindMethodGodot(p_class_name TypeName, p_method *MethodBind) {
	var cDefArgs *GDNativeVariantPtr

	if len(p_method.DefaultArguments) > 0 {
		cDefArgs = (*GDNativeVariantPtr)(unsafe.Pointer(&p_method.DefaultArguments[0]))
	} else {
		cDefArgs = (*GDNativeVariantPtr)(nullptr)
	}

	method_info := NewGDNativeExtensionClassMethodInfo(
		(string)(p_method.GDName),
		unsafe.Pointer(p_method),
		(GDNativeExtensionClassMethodCall)(C.cgo_method_bind_method_call),
		(GDNativeExtensionClassMethodPtrCall)(C.cgo_method_bind_method_ptrcall),
		(uint32)(p_method.HintFlags),
		p_method.ArgumentCount,
		p_method.HasReturn,
		(GDNativeExtensionClassMethodGetArgumentType)(C.cgo_method_bind_bind_get_argument_type),
		(GDNativeExtensionClassMethodGetArgumentInfo)(C.cgo_method_bind_bind_get_argument_info),
		(GDNativeExtensionClassMethodGetArgumentMetadata)(C.cgo_method_bind_bind_get_argument_metadata),
		(uint32)(len(p_method.DefaultArguments)),
		cDefArgs,
	)

	GDNativeInterface_classdb_register_extension_class_method(internal.gdnInterface, internal.library, (string)(p_class_name), &method_info)
}

type SignalParam struct {
	Type GDNativeVariantType
	Name PropertyName
}

func ClassDBAddSignal(t GDClass, signalName string, params ...SignalParam) {
	log.Debug("ClassDBAddSignal called",
		zap.String("signalName", (string)(signalName)),
		zap.Any("params", params),
	)

	var (
		ci *ClassInfo
		ok bool
	)
	typeName := t.GetClassName()

	sigName := (SignalName)(signalName)

	if ci, ok = gdRegisteredGDClasses.Get(typeName); !ok {
		log.Panic("Class doesn't exist.", zap.String("class", (string)(typeName)))
		return
	}

	if _, ok = ci.SignalNameMap[sigName]; ok {
		log.Panic("Constant already registered.", zap.String("class", (string)(typeName)))
		return
	}

	ci.SignalNameMap[sigName] = struct{}{}

	paramArr := make([]GDNativePropertyInfo, len(params))

	for i, p := range params {
		paramArr[i] = NewGDNativePropertyInfo(
			string(typeName),
			p.Type,
			string(p.Name),
			(uint32)(PROPERTY_HINT_NONE),
			"",
			(uint32)(PROPERTY_USAGE_DEFAULT),
		)
	}

	var argsPtr *GDNativePropertyInfo

	if len(paramArr) > 0 {
		argsPtr = (*GDNativePropertyInfo)(unsafe.Pointer(&paramArr[0]))
	} else {
		argsPtr = (*GDNativePropertyInfo)(nullptr)
	}

	GDNativeInterface_classdb_register_extension_class_signal(internal.gdnInterface, internal.library, (string)(typeName), signalName, argsPtr, GDNativeInt(len(params)))
}

// ClassDBBindConstant binds a constant in godot.
func ClassDBBindConstant(t GDClass, p_constant_name string, p_constant_value int) {
	classDBBindIntegerConstant(t, "", p_constant_name, (GDNativeInt)(p_constant_value), false)
}

// ClassDBBindEnumConstant binds a enum value in godot.
func ClassDBBindEnumConstant(t GDClass, p_enum_name, p_constant_name string, p_constant_value int) {
	classDBBindIntegerConstant(t, p_enum_name, p_constant_name, (GDNativeInt)(p_constant_value), false)
}

// ClassDBBindBitfieldFlag binds a bitfield value in godot.
func ClassDBBindBitfieldFlag(t GDClass, p_enum_name, p_constant_name string, p_constant_value int) {
	classDBBindIntegerConstant(t, "", p_constant_name, (GDNativeInt)(p_constant_value), true)
}

func classDBBindIntegerConstant(t GDClass, p_enum_name, p_constant_name string, p_constant_value GDNativeInt, p_is_bitfield bool) {
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

	constName := (ConstantName)(p_constant_name)

	if ci, ok = gdRegisteredGDClasses.Get(typeName); !ok {
		log.Panic("Class doesn't exist.", zap.String("class", (string)(typeName)))
		return
	}

	if _, ok = ci.ConstantNameMap[constName]; ok {
		log.Panic("Constant already registered.", zap.String("class", (string)(typeName)))
		return
	}

	ci.ConstantNameMap[constName] = struct{}{}

	bitfield := (GDNativeBool)(BoolEncoder.EncodeArg(p_is_bitfield))

	GDNativeInterface_classdb_register_extension_class_integer_constant(internal.gdnInterface, internal.library, (string)(typeName), p_enum_name, p_constant_name, p_constant_value, bitfield)
}

func classDBBindVirtualMethod(t GDClass, p_method_name MethodName, p_arg_names ...string) {

	cn := t.GetClassName()

	ci, ok := gdRegisteredGDClasses.Get((TypeName)(cn))

	if !ok {
		log.Panic("Class doesn't exist.", zap.String("name", (string)(cn)))
		return
	}

	if _, ok = ci.MethodMap[p_method_name]; ok {
		log.Panic("Method already registered as non-virtual.", zap.String("name", (string)(p_method_name)))
		return
	}

	if _, ok = ci.VirtualMethodMap[p_method_name]; ok {
		log.Panic("Virtual method already registered as non-virtual.", zap.String("name", (string)(p_method_name)))
		return
	}

	// TODO: implement
	log.Panic("missing implementation")

	// ci.VirtualMethodMap[p_method_name] = p_call
}

func classDBInitialize(pLevel GDNativeInitializationLevel) {
	for _, ci := range gdRegisteredGDClasses.Values() {
		if ci.Level != pLevel {
			continue
		}

		// Nothing to do here for now...
	}
}

func classDBDeinitialize(pLevel GDNativeInitializationLevel) {
	for _, ci := range gdRegisteredGDClasses.Values() {
		if ci.Level != pLevel {
			continue
		}

		GDNativeInterface_classdb_unregister_extension_class(internal.gdnInterface, internal.library, (string)(ci.Name))

		// NOTE: godot-cpp iterates through the map to delete all method binds
		for n, mb := range ci.MethodMap {
			delete(ci.MethodMap, n)
			mb.Destroy()
		}
	}
}

func ClassDBRegisterClass(inst GDClass, bindMethodsFunc func(t GDClass)) {
	// Register this class within our plugin

	name := inst.GetClassName()
	parentName := inst.GetParentClassName()

	if name == parentName {
		log.Panic("class and parent cannot have the same name", zap.String("name", string(name)), zap.String("parent", string(parentName)))
	}

	log.Debug("ClassDBRegisterClass called", zap.String("name", (string)(name)))

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

	if string(name) != classType.Name() {
		log.Panic("GetClassName must match struct name", zap.String("name", string(name)), zap.String("struct_name", classType.Name()))
	}

	vf := reflect.VisibleFields(classType)

	if len(vf) == 0 {
		log.Panic("Missing GDNativeClass interface: no visible struct fields")
	}

	// need to ensure the GDNativeClass is always the first struct
	inheritType := vf[0].Type

	if inheritType == nil {
		log.Panic("Missing GDNativeClass interface: inherits type nil")
	}

	if fmt.Sprintf("%sImpl", string(parentName)) != inheritType.Name() {
		log.Panic("GetParentClassName must match struct name", zap.String("parent_name", string(parentName)), zap.String("struct_inherit_type", inheritType.Name()))
	}

	cl := NewClassInfo(name, parentName, level, classType, inheritType, parentPtr)

	if cl == nil {
		log.Panic("ClassInfo cannot be nil")
	}

	gdRegisteredGDClasses.Set(name, cl)

	if _, ok := gdNativeConstructors.Get((TypeName)(parentName)); !ok {
		log.Panic("Missing GDNativeClass interface: unhandled inherits type", zap.Any("class_type", classType), zap.Any("parent_type", parentName))
	}

	gdClassRegisterInstanceBindingCallbacks(name)

	log.Info("gdclass registered", zap.String("name", (string)(name)), zap.String("parent_type", (string)(parentName)))

	cName := C.CString((string)(name))

	// Register this class with Godot
	info := NewGDNativeExtensionClassCreationInfo(
		(GDNativeExtensionClassCreateInstance)(C.cgo_gdnative_extension_class_create_instance),
		(GDNativeExtensionClassFreeInstance)(C.cgo_gdnative_extension_class_free_instance),
		(GDNativeExtensionClassGetVirtual)(C.cgo_classdb_get_virtual_func),
		unsafe.Pointer(cName),
	)

	GDNativeInterface_classdb_register_extension_class(
		(*GDNativeInterface)(internal.gdnInterface),
		(GDNativeExtensionClassLibraryPtr)(internal.library),
		(string)(name),
		(string)(parentName),
		&info,
	)

	// call bind_methods etc. to register all members of the class
	gdClassInitializeClass(inst)

	// now register our class within ClassDB within Godot
	classDBInitializeClass(cl)

	bindMethodsFunc(inst)
}

func ClassDBBindMethod(p_instance GDClass, p_go_method_name MethodName, p_method_name MethodName, p_arg_names []string, p_default_values []*Variant) *MethodBind {
	return classDBBindMethod(p_instance, p_go_method_name, p_method_name, METHOD_FLAGS_DEFAULT, p_arg_names, p_default_values)
}

func ClassDBBindMethodStatic(p_instance GDClass, p_go_method_name MethodName, p_method_name MethodName, p_arg_names []string, p_default_values []*Variant) *MethodBind {
	return classDBBindMethod(p_instance, p_go_method_name, p_method_name, METHOD_FLAGS_DEFAULT|METHOD_FLAG_STATIC, p_arg_names, p_default_values)
}

func classDBBindMethod(p_instance GDClass, p_go_method_name MethodName, p_method_name MethodName, hintFlags MethodFlags, p_arg_names []string, p_default_values []*Variant) *MethodBind {
	log.Debug("classDBBindMethod called", zap.String("go_name", (string)(p_go_method_name)), zap.String("gd_name", (string)(p_method_name)), zap.Any("hint", hintFlags))

	t := reflect.TypeOf(p_instance)

	n := p_instance.GetClassName()

	m, ok := t.MethodByName((string)(p_go_method_name))

	if !ok {
		log.Panic("unable to find function", zap.String("gdclass", (string)(n)), zap.String("method_name", (string)(p_go_method_name)))
	}

	mb := NewMethodBind(m, p_method_name, p_arg_names, p_default_values, hintFlags)

	classDBBindMethodFi(0, mb)

	return mb
}

func classDBInitializeClass(p_cl *ClassInfo) {
}

//export GoCallback_ClassDBGetVirtualFunc
func GoCallback_ClassDBGetVirtualFunc(pUserdata unsafe.Pointer, pName *C.char) C.GDNativeExtensionClassCallVirtual {
	className := (TypeName)(C.GoString((*C.char)(pUserdata)))
	methodName := (MethodName)(C.GoString(pName))

	log.Debug("GoCallback_ClassDBGetVirtualFunc called", zap.String("type_name", (string)(className)), zap.String("method", (string)(methodName)))

	ci, ok := gdRegisteredGDClasses.Get(className)

	if !ok {
		log.Warn(fmt.Sprintf("class \"%s\" doesn't exist", className))
		return (C.GDNativeExtensionClassCallVirtual)(nullptr)
	}

	m, ok := ci.VirtualMethodMap[methodName]

	if !ok {
		return (C.GDNativeExtensionClassCallVirtual)(nullptr)
	}

	return (C.GDNativeExtensionClassCallVirtual)(m)
}

func BindVirtualMethod[T any](methodName string) {
	panic("TODO: implement BindVirtualMethod")
}
