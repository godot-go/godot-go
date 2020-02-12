#ifndef CGO_METHOD_BIND_H
#define CGO_METHOD_BIND_H

#include <godot/gdnative_interface.h>

GDNativeVariantType cgo_method_bind_bind_get_argument_type(void *p_method_userdata, int32_t p_argument);
void cgo_method_bind_bind_get_argument_info(void *p_method_userdata, int32_t p_argument, GDNativePropertyInfo *r_info);
GDNativeExtensionClassMethodArgumentMetadata cgo_method_bind_bind_get_argument_metadata(void *p_method_userdata, int32_t p_argument);

void cgo_method_bind_method_call(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDNativeVariantPtr *p_args, const GDNativeInt p_argument_count, GDNativeVariantPtr r_return, GDNativeCallError *r_error);
void cgo_method_bind_method_ptrcall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDNativeTypePtr *p_args, GDNativeTypePtr r_ret);

#endif
