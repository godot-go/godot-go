#ifndef CGO_GODOT_GO_CALLABLE_CUSTOM_WRAPPER_H
#define CGO_GODOT_GO_CALLABLE_CUSTOM_WRAPPER_H

#include <godot/gdextension_interface.h>

// cgo_set_handle_slot is the shared handle-slot writer in pkg/log/cgo_ptr.h.
#include "cgo_ptr.h"

// cgo_callcustom_call signature must match GDExtensionCallableCustomCall
// callback when the engine invokes the custom callable from script.
void cgo_callcustom_call(void *callable_userdata, const GDExtensionConstVariantPtr *p_args, GDExtensionInt p_argument_count, GDExtensionVariantPtr r_return, GDExtensionCallError *r_error);

// cgo_callcustom_free signature must match GDExtensionCallableCustomFree
// callback when the engine destroys the last reference to the callable.
void cgo_callcustom_free(void *callable_userdata);

// cgo_callcustom_to_string signature must match GDExtensionCallableCustomToString
// callback when the engine renders the callable (str(cb)).
void cgo_callcustom_to_string(void *callable_userdata, GDExtensionBool *r_is_valid, GDExtensionStringPtr r_out);

#endif
