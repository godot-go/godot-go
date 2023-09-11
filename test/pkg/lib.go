package pkg

/*
#cgo CFLAGS: -I${SRCDIR} -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
#include "example.h"
*/
import "C"

import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextension"
	"github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
)

func RegisterExampleTypes() {
	log.Debug("RegisterExampleTypes called")
	ClassDBRegisterClass(&ExampleRef{}, func(t GDClass) {
		ClassDBBindMethod(t, "GetId", "get_id", nil, nil)
		ClassDBBindMethod(t, "SetId", "set_id", []string{"id"}, nil)
		ClassDBAddProperty(t, gdextensionffi.GDEXTENSION_VARIANT_TYPE_INT, "group_subgroup_id", "set_id", "get_id")
		log.Debug("ExampleRef registered")
	})

	ClassDBRegisterClass(&Example{}, func(t GDClass) {
		ClassDBBindMethodVirtual(t, "V_Ready", "_ready", nil, nil)
		ClassDBBindMethodVirtual(t, "V_Input", "_input", []string{"event"}, nil)
		ClassDBBindMethodVirtual(t, "V_Set", "_set", []string{"name", "value"}, nil)
		ClassDBBindMethodVirtual(t, "V_Get", "_get", []string{"name"}, nil)
		// ClassDBBindMethodVirtual(t, "V_GetPropertyList", "_get_property_list", []string{"name", "value"}, nil)
		// fn := func(obj Object, args ...Variant) Variant {
		// 	return (obj.(*Example)).VarargsFunc(args...)
		// }
		// fn := (*Example).VarargsFunc
		// ClassDBBindMethodVarargs(t, fn, "varargs_func", nil, nil)
		// ClassDBBindMethodStatic(t, "TestStatic", "test_static", []string{"a", "b"}, nil)
		// ClassDBBindMethodStatic(t, "TestStatic2", "test_static2", nil, nil)
		ClassDBBindMethod(t, "SimpleFunc", "simple_func", nil, nil)
		ClassDBBindMethod(t, "SimpleConstFunc", "simple_const_func", []string{"a"}, nil)
		ClassDBBindMethod(t, "ReturnSomething", "return_something", []string{"base", "f32", "f64", "i", "i8", "i16", "i32", "i64"}, nil)
		ClassDBBindMethod(t, "ReturnSomethingConst", "return_something_const", nil, nil)
		ClassDBBindMethod(t, "GetV4", "get_v4", nil, nil)
		ClassDBBindMethod(t, "DefArgs", "def_args", []string{"a", "b"}, []Variant{NewVariantInt64(100), NewVariantInt64(200)})
		ClassDBBindMethod(t, "TestArray", "test_array", nil, nil)
		ClassDBBindMethod(t, "TestDictionary", "test_dictionary", nil, nil)
		ClassDBBindMethod(t, "TestStringOps", "test_string_ops", nil, nil)

		// Properties
		ClassDBAddPropertyGroup(t, "Test group", "group_")
		ClassDBAddPropertySubgroup(t, "Test subgroup", "group_subgroup_")
		ClassDBBindMethod(t, "GetCustomPosition", "get_custom_position", nil, nil)
		ClassDBBindMethod(t, "SetCustomPosition", "set_custom_position", []string{"position"}, nil)
		ClassDBAddProperty(t, gdextensionffi.GDEXTENSION_VARIANT_TYPE_VECTOR2, "group_subgroup_custom_position", "set_custom_position", "get_custom_position")
		ClassDBBindMethod(t, "GetPropertyFromList", "get_property_from_list", nil, nil)
		ClassDBBindMethod(t, "SetPropertyFromList", "set_property_from_list", []string{"v"}, nil)
		ClassDBAddProperty(t, gdextensionffi.GDEXTENSION_VARIANT_TYPE_VECTOR3, "group_subgroup_property_from_list", "set_property_from_list", "get_property_from_list")

		// Signals.
		ClassDBAddSignal(t, "custom_signal",
			SignalParam{
				Type: gdextensionffi.GDEXTENSION_VARIANT_TYPE_STRING,
				Name: "name"},
			SignalParam{
				Type: gdextensionffi.GDEXTENSION_VARIANT_TYPE_INT,
				Name: "value",
			})
		ClassDBBindMethod(t, "EmitCustomSignal", "emit_custom_signal", []string{"name", "value"}, nil)

		// constants
		ClassDBBindEnumConstant(t, "ExampleEnum", "FIRST", int(ExampleFirst))
		ClassDBBindEnumConstant(t, "ExampleEnum", "ANSWER_TO_EVERYTHING", int(AnswerToEverything))
		ClassDBBindConstant(t, "CONSTANT_WITHOUT_ENUM", int(EXAMPLE_ENUM_CONSTANT_WITHOUT_ENUM))

		// others
		ClassDBBindMethod(t, "TestCastTo", "test_cast_to", nil, nil)
		log.Debug("Example registered")
	})
}

func UnregisterExampleTypes() {
	log.Debug("UnregisterExampleTypes called")
}

//export TestDemoInit
func TestDemoInit(p_get_proc_address unsafe.Pointer, p_library unsafe.Pointer, r_initialization unsafe.Pointer) bool {
	log.Debug("ExampleLibraryInit called")
	initObj := NewInitObject(
		(gdextensionffi.GDExtensionInterfaceGetProcAddress)(p_get_proc_address),
		(gdextensionffi.GDExtensionClassLibraryPtr)(p_library),
		(*gdextensionffi.GDExtensionInitialization)(r_initialization),
	)

	initObj.RegisterSceneInitializer(RegisterExampleTypes)
	initObj.RegisterSceneTerminator(UnregisterExampleTypes)

	return initObj.Init()
}
