#ifndef CGO_GODOT_GO_METHOD_BIND_H
#define CGO_GODOT_GO_METHOD_BIND_H

#include <godot/gdextension_interface.h>

// GDExtensionVariantType cgo_method_bind_bind_get_argument_type(void *p_method_userdata, int32_t p_argument);
// void cgo_method_bind_bind_get_argument_info(void *p_method_userdata, int32_t p_argument, GDExtensionPropertyInfo *r_info);
// GDExtensionClassMethodArgumentMetadata cgo_method_bind_bind_get_argument_metadata(void *p_method_userdata, int32_t p_argument);

void cgo_method_bind_method_call(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionVariantPtr *p_args, const GDExtensionInt p_argument_count, GDExtensionVariantPtr r_return, GDExtensionCallError *r_error);
void cgo_method_bind_method_ptrcall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionTypePtr *p_args, GDExtensionTypePtr r_ret);

#endif
