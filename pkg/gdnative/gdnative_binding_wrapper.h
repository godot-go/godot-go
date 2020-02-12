#ifndef CGO_GDNATIVE_BINDING_WRAPPER_H
#define CGO_GDNATIVE_BINDING_WRAPPER_H

#include <godot/gdnative_interface.h>

void GDExtensionBindingInitializeLevel(void *userdata, GDNativeInitializationLevel p_level);
void GDExtensionBindingDeinitializeLevel(void *userdata, GDNativeInitializationLevel p_level);

void cgo_wrapper_binding_initialize(void *userdata, GDNativeInitializationLevel p_level);
void cgo_wrapper_binding_deinitialize(void *userdata, GDNativeInitializationLevel p_level);

void *cgo_gdclass_binding_create_callback(void *p_token, void *p_instance);
void cgo_gdclass_binding_free_callback(void *p_token, void *p_instance, void *p_binding);
GDNativeBool cgo_gdclass_binding_reference_callback(void *p_token, void *p_instance, GDNativeBool p_reference);

#endif
