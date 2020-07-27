#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
#include <cgo_gateway_class.h>
#include <gdnative.wrappergen.h>

godot_object *go_godot_get_class_constructor_new(godot_gdnative_core_api_struct * p_api, const char * p_classname) {
	godot_class_constructor constructor = p_api->godot_get_class_constructor(p_classname);
	return constructor();
}
