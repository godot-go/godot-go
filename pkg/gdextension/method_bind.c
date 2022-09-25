#include <godot/gdnative_interface.h>
#include "method_bind.h"
#include "stacktrace.h"

extern GDNativeVariantType GoCallback_MethodBindBindGetArgumentType(void *p_method_userdata, int32_t p_argument);
extern void GoCallback_MethodBindBindGetArgumentInfo(void *p_method_userdata, int32_t p_argument, GDNativePropertyInfo *r_info);
extern GDNativeExtensionClassMethodArgumentMetadata GoCallback_MethodBindBindGetArgumentMetadata(void *p_method_userdata, int32_t p_argument);
extern void GoCallback_MethodBindBindCall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDNativeVariantPtr *p_args, const GDNativeInt p_argument_count, GDNativeVariantPtr r_return, GDNativeCallError *r_error);
extern void GoCallback_MethodBindBindPtrcall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDNativeTypePtr *p_args, GDNativeTypePtr r_ret);

GDNativeVariantType cgo_method_bind_bind_get_argument_type(void *p_method_userdata, int32_t p_argument) {
	printStacktrace();
	return GoCallback_MethodBindBindGetArgumentType(p_method_userdata, p_argument);
}

void cgo_method_bind_bind_get_argument_info(void *p_method_userdata, int32_t p_argument, GDNativePropertyInfo *r_info) {
	printStacktrace();
	GoCallback_MethodBindBindGetArgumentInfo(p_method_userdata, p_argument, r_info);
}

GDNativeExtensionClassMethodArgumentMetadata cgo_method_bind_bind_get_argument_metadata(void *p_method_userdata, int32_t p_argument) {
	printStacktrace();
	return GoCallback_MethodBindBindGetArgumentMetadata(p_method_userdata, p_argument);
}

void cgo_method_bind_method_call(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDNativeVariantPtr *p_args, const GDNativeInt p_argument_count, GDNativeVariantPtr r_return, GDNativeCallError *r_error) {
	printStacktrace();
    GoCallback_MethodBindBindCall(method_userdata, p_instance, p_args, p_argument_count, r_return, r_error);
}

void cgo_method_bind_method_ptrcall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDNativeTypePtr *p_args, GDNativeTypePtr r_ret) {
	printStacktrace();
    GoCallback_MethodBindBindPtrcall(method_userdata, p_instance, p_args, r_ret);
}
