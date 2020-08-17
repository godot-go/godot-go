#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
#include <cgo_gateway_register_class.h>
#include <gdnative.wrappergen.h>


// This is a gateway function for the create method.
void *cgo_gateway_create_func(godot_object *owner, void *method_data) {
	return go_create_func(owner, method_data);
}

// This is a gateway function for the destroy method.
void cgo_gateway_destroy_func(godot_object *owner, void *method_data, void *user_data) {
	go_destroy_func(owner, method_data, user_data);
}

// This is a gateway function for the destroy method.
godot_variant cgo_gateway_method_func(godot_object *owner, void *method_data, void *user_data, int nargs, godot_variant **args) {
	return go_method_func(owner, method_data, user_data, nargs, args);
}

void cgo_gateway_property_set_func(godot_object *owner, void *method_data, void *user_data, godot_variant *value) {
	return go_property_set_func(owner, method_data, user_data, value);
}

godot_variant cgo_gateway_property_get_func(godot_object *owner, void *method_data, void *user_data) {
	return go_property_get_func(owner, method_data, user_data);
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

// This is a gateway function for the free method.
void cgo_gateway_property_set_free_func(void *method_data) {
	go_property_set_free_func(method_data);
}

// This is a gateway function for the free method.
void cgo_gateway_property_get_free_func(void *method_data) {
	go_property_get_free_func(method_data);
}
