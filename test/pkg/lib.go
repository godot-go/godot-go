package pkg

/*
#cgo CFLAGS: -I${SRCDIR} -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
#include "example.h"
*/
import "C"

import (
	"unsafe"

	"github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/gdextension"
	"github.com/godot-go/godot-go/pkg/log"
)

func RegisterExampleTypes() {
	log.Debug("RegisterExampleTypes called")
	gdextension.ClassDBRegisterClass(&ExampleRef{}, func(t gdextension.GDClass) {
		gdextension.ClassDBBindMethod(t, "GetId", "get_id", nil, nil)
		gdextension.ClassDBBindMethod(t, "SetId", "set_id", []string{"id"}, nil)
		gdextension.ClassDBAddProperty(t, gdextensionffi.GDEXTENSION_VARIANT_TYPE_INT, "group_subgroup_id", "set_id", "get_id")
	log.Debug("ExampleRef registered")
})

	gdextension.ClassDBRegisterClass(&Example{}, func(t gdextension.GDClass) {
		gdextension.ClassDBBindMethodVirtual(t, "V_Ready", "_ready", nil, nil)
		gdextension.ClassDBBindMethodVirtual(t, "V_Input", "_input", []string{"event"}, nil)
		// gdextension.ClassDBBindMethodStatic(t, "TestStatic", "test_static", []string{"a", "b"}, nil)
		// gdextension.ClassDBBindMethodStatic(t, "TestStatic2", "test_static2", nil, nil)
		gdextension.ClassDBBindMethod(t, "SimpleFunc", "simple_func", nil, nil)
		gdextension.ClassDBBindMethod(t, "SimpleConstFunc", "simple_const_func", []string{"a"}, nil)
		gdextension.ClassDBBindMethod(t, "ReturnSomething", "return_something", []string{"base", "f32", "f64", "i", "i8", "i16", "i32", "i64"}, nil)
		gdextension.ClassDBBindMethod(t, "ReturnSomethingConst", "return_something_const", nil, nil)
		gdextension.ClassDBBindMethod(t, "GetV4", "get_v4", nil, nil)
		defArgA := gdextension.NewVariantInt64(100)
		defArgB := gdextension.NewVariantInt64(200)
		gdextension.ClassDBBindMethod(t, "DefArgs", "def_args", []string{"a", "b"}, []*gdextension.Variant{&defArgA, &defArgB})
		gdextension.ClassDBBindMethod(t, "TestArray", "test_array", nil, nil)
		gdextension.ClassDBBindMethod(t, "TestDictionary", "test_dictionary", nil, nil)
		gdextension.ClassDBBindMethod(t, "TestStringOps", "test_string_ops", nil, nil)

		// Properties
		gdextension.ClassDBAddPropertyGroup(t, "Test group", "group_")
		gdextension.ClassDBAddPropertySubgroup(t, "Test subgroup", "group_subgroup_")
		gdextension.ClassDBBindMethod(t, "GetCustomPosition", "get_custom_position", nil, nil)
		gdextension.ClassDBBindMethod(t, "SetCustomPosition", "set_custom_position", []string{"position"}, nil)
		gdextension.ClassDBAddProperty(t, gdextensionffi.GDEXTENSION_VARIANT_TYPE_VECTOR2, "group_subgroup_custom_position", "set_custom_position", "get_custom_position")
		gdextension.ClassDBBindMethod(t, "GetPropertyFromList", "get_property_from_list", nil, nil)
		gdextension.ClassDBBindMethod(t, "SetPropertyFromList", "set_property_from_list", []string{"v"}, nil)
		gdextension.ClassDBAddProperty(t, gdextensionffi.GDEXTENSION_VARIANT_TYPE_VECTOR3, "group_subgroup_property_from_list", "set_property_from_list", "get_property_from_list")

		// Signals.
		gdextension.ClassDBAddSignal(t, "custom_signal",
			gdextension.SignalParam{
				Type: gdextensionffi.GDEXTENSION_VARIANT_TYPE_STRING,
				Name: "name"},
			gdextension.SignalParam{
				Type: gdextensionffi.GDEXTENSION_VARIANT_TYPE_INT,
				Name: "value",
			})
		gdextension.ClassDBBindMethod(t, "EmitCustomSignal", "emit_custom_signal", []string{"name", "value"}, nil)

		// constants
		gdextension.ClassDBBindEnumConstant(t, "ExampleEnum", "FIRST", int(ExampleFirst))
		gdextension.ClassDBBindEnumConstant(t, "ExampleEnum", "ANSWER_TO_EVERYTHING", int(AnswerToEverything))
		gdextension.ClassDBBindConstant(t, "CONSTANT_WITHOUT_ENUM", int(EXAMPLE_ENUM_CONSTANT_WITHOUT_ENUM))

		// others
		gdextension.ClassDBBindMethod(t, "TestCastTo", "test_cast_to", nil, nil)
		log.Debug("Example registered")
	})
}

func UnregisterExampleTypes() {
	log.Debug("UnregisterExampleTypes called")
}

//export TestDemoInit
func TestDemoInit(p_get_proc_address unsafe.Pointer, p_library unsafe.Pointer, r_initialization unsafe.Pointer) bool {
	log.Debug("ExampleLibraryInit called")
	initObj := gdextension.NewInitObject(
		(gdextensionffi.GDExtensionInterfaceGetProcAddress)(p_get_proc_address),
		(gdextensionffi.GDExtensionClassLibraryPtr)(p_library),
		(*gdextensionffi.GDExtensionInitialization)(r_initialization),
	)

	initObj.RegisterSceneInitializer(RegisterExampleTypes)
	initObj.RegisterSceneTerminator(UnregisterExampleTypes)

	return initObj.Init()
}