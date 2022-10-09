#include <godot/gdextension_interface.h>
#include "method_bind.h"
#include "stacktrace.h"

// extern GDExtensionVariantType GoCallback_MethodBindBindGetArgumentType(void *p_method_userdata, int32_t p_argument);
// extern void GoCallback_MethodBindBindGetArgumentInfo(void *p_method_userdata, int32_t p_argument, GDExtensionPropertyInfo *r_info);
// extern GDExtensionClassMethodArgumentMetadata GoCallback_MethodBindBindGetArgumentMetadata(void *p_method_userdata, int32_t p_argument);
extern void GoCallback_MethodBindBindCall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionVariantPtr *p_args, const GDExtensionInt p_argument_count, GDExtensionVariantPtr r_return, GDExtensionCallError *r_error);
extern void GoCallback_MethodBindBindPtrcall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionTypePtr *p_args, GDExtensionTypePtr r_ret);

// GDExtensionVariantType cgo_method_bind_bind_get_argument_type(void *p_method_userdata, int32_t p_argument) {
// 	printStacktrace();
// 	return GoCallback_MethodBindBindGetArgumentType(p_method_userdata, p_argument);
// }

// void cgo_method_bind_bind_get_argument_info(void *p_method_userdata, int32_t p_argument, GDExtensionPropertyInfo *r_info) {
// 	printStacktrace();
// 	GoCallback_MethodBindBindGetArgumentInfo(p_method_userdata, p_argument, r_info);
// }

// GDExtensionClassMethodArgumentMetadata cgo_method_bind_bind_get_argument_metadata(void *p_method_userdata, int32_t p_argument) {
// 	printStacktrace();
// 	return GoCallback_MethodBindBindGetArgumentMetadata(p_method_userdata, p_argument);
// }

void cgo_method_bind_method_call(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionVariantPtr *p_args, const GDExtensionInt p_argument_count, GDExtensionVariantPtr r_return, GDExtensionCallError *r_error) {
	printStacktrace();
    GoCallback_MethodBindBindCall(method_userdata, p_instance, p_args, p_argument_count, r_return, r_error);
}

void cgo_method_bind_method_ptrcall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionTypePtr *p_args, GDExtensionTypePtr r_ret) {
	printStacktrace();
    GoCallback_MethodBindBindPtrcall(method_userdata, p_instance, p_args, r_ret);
}
