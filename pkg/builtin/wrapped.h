#ifndef CGO_GODOT_GO_WRAPPED_H
#define CGO_GODOT_GO_WRAPPED_H

#include <godot/gdextension_interface.h>

void *cgo_gdclass_binding_create_callback(void *p_token, void *p_instance);
void cgo_gdclass_binding_free_callback(void *p_token, void *p_instance, void *p_binding);
GDExtensionBool cgo_gdclass_binding_reference_callback(void *p_token, void *p_instance, GDExtensionBool p_reference);

#endif
