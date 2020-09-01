#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
#include <gdnative_wrappergen.h>
#include <cgo_gateway_instance_binding.h>

void *cgo_gateway_alloc_instance_binding_data(void *data, const void *type_tag, godot_object *instance) {
	return go_alloc_instance_binding_data(data, (void *)type_tag, instance);
}

void cgo_gateway_free_instance_binding_data(void *data, void *wrapper) {
	go_free_instance_binding_data(data, wrapper);
}
