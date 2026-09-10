#include <godot/gdextension_interface.h>
#include <stdint.h>
#include "method_bind.h"
#include "stacktrace.h"

void *cgo_method_bind_userdata(uintptr_t handle) {
	return (void *)handle;
}

extern void GoCallback_MethodBindMethodCall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionVariantPtr *p_args, const GDExtensionInt p_argument_count, GDExtensionVariantPtr r_return, GDExtensionCallError *r_error);
extern void GoCallback_MethodBindMethodPtrcall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionTypePtr *p_args, GDExtensionTypePtr r_ret);

void cgo_method_bind_method_call(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionVariantPtr *p_args, const GDExtensionInt p_argument_count, GDExtensionVariantPtr r_return, GDExtensionCallError *r_error) {
	printStacktrace();
    GoCallback_MethodBindMethodCall(method_userdata, p_instance, p_args, p_argument_count, r_return, r_error);
}

void cgo_method_bind_method_ptrcall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionTypePtr *p_args, GDExtensionTypePtr r_ret) {
	printStacktrace();
    GoCallback_MethodBindMethodPtrcall(method_userdata, p_instance, p_args, r_ret);
}
