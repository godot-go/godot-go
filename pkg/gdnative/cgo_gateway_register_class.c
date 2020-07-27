#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
#include <cgo_gateway_register_class.h>
#include <gdnative.wrappergen.h>


// This is a gateway function for the create method.
void *cgo_gateway_create_func(godot_object *obj, void *method_data) {
	long tt = (long)(method_data);
	char* cname = get_class_name_from_type_tag(tt);
	// printf("cgo_gateway_create_func: type_tag: (%d) %s\n", tt, cname);
	free(cname);
	return go_create_func(obj, method_data);
}

// This is a gateway function for the destroy method.
void cgo_gateway_destroy_func(godot_object *obj, void *method_data, void *user_data) {
	go_destroy_func(obj, method_data, user_data);
}

// This is a gateway function for the destroy method.
godot_variant cgo_gateway_method_func(godot_object *obj, void *method_data, void *user_data, int nargs, godot_variant **args) {
	long mt = (long)(method_data);
	char *cname = get_method_name_from_method_tag(mt);
	// printf("cgo_gateway_method_func: method_data: %s, nargs: %d, args: %p\n", cname, nargs, args);
	free(cname);
	return go_method_func(obj, method_data, user_data, nargs, args);
}

// This is a gateway function for the free method.
void cgo_gateway_create_free_func(void *method_data) {
	go_create_free_func(method_data);
}

// This is a gateway function for the free method.
void cgo_gateway_destroy_free_func(void *method_data) {
	go_destroy_free_func(method_data);
}

// This is a gateway function for the free method.
void cgo_gateway_method_free_func(void *method_data) {
	go_method_free_func(method_data);
}
