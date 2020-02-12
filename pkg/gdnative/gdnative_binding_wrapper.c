#include <godot/gdnative_interface.h>
#include "gdnative_binding_wrapper.h"

// GDClass
extern void *GoCallback_GDClassBindingCreate(void *p_token, void *p_instance);
extern void GoCallback_GDClassBindingFree(void *p_token, void *p_instance, void *p_binding);
extern GDNativeBool GoCallback_GDClassBindingReference(void *p_token, void *p_instance, GDNativeBool p_reference);

void cgo_wrapper_binding_initialize(void *userdata, GDNativeInitializationLevel p_level) {
    GDExtensionBindingInitializeLevel(userdata, p_level);
}

void cgo_wrapper_binding_deinitialize(void *userdata, GDNativeInitializationLevel p_level) {
    GDExtensionBindingDeinitializeLevel(userdata, p_level);
}

void *cgo_gdclass_binding_create_callback(void *p_token, void *p_instance) {
	return GoCallback_GDClassBindingCreate(p_token, p_instance);
}

void cgo_gdclass_binding_free_callback(void *p_token, void *p_instance, void *p_binding) {
	GoCallback_GDClassBindingFree(p_token, p_instance, p_binding);
}

GDNativeBool cgo_gdclass_binding_reference_callback(void *p_token, void *p_instance, GDNativeBool p_reference) {
	return GoCallback_GDClassBindingReference(p_token, p_instance, p_reference);
}
