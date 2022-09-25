#ifndef CGO_GODOT_GO_WRAPPED_H
#define CGO_GODOT_GO_WRAPPED_H

#include <godot/gdnative_interface.h>

void *cgo_gdclass_binding_create_callback(void *p_token, void *p_instance);
void cgo_gdclass_binding_free_callback(void *p_token, void *p_instance, void *p_binding);
GDNativeBool cgo_gdclass_binding_reference_callback(void *p_token, void *p_instance, GDNativeBool p_reference);

#endif
