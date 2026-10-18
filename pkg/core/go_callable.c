#include <godot/gdextension_interface.h>
#include <stdint.h>
#include "go_callable.h"
#include "stacktrace.h"

extern void GoCallback_GoCallableCall(void *callable_userdata, const GDExtensionConstVariantPtr *p_args, GDExtensionInt p_argument_count, GDExtensionVariantPtr r_return, GDExtensionCallError *r_error);
extern void GoCallback_GoCallableFree(void *callable_userdata);
extern void GoCallback_GoCallableToString(void *callable_userdata, GDExtensionBool *r_is_valid, GDExtensionStringPtr r_out);

void cgo_callcustom_call(void *callable_userdata, const GDExtensionConstVariantPtr *p_args, GDExtensionInt p_argument_count, GDExtensionVariantPtr r_return, GDExtensionCallError *r_error) {
	printStacktrace();
	GoCallback_GoCallableCall(callable_userdata, p_args, p_argument_count, r_return, r_error);
}

void cgo_callcustom_free(void *callable_userdata) {
	printStacktrace();
	GoCallback_GoCallableFree(callable_userdata);
}

void cgo_callcustom_to_string(void *callable_userdata, GDExtensionBool *r_is_valid, GDExtensionStringPtr r_out) {
	printStacktrace();
	GoCallback_GoCallableToString(callable_userdata, r_is_valid, r_out);
}
