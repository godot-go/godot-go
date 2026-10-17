#ifndef CGO_GODOT_GO_METHOD_BIND_H
#define CGO_GODOT_GO_METHOD_BIND_H

#include <godot/gdextension_interface.h>
#include <stdint.h>

// cgo_set_handle_slot is the shared handle-slot writer in pkg/log/cgo_ptr.h.
#include "cgo_ptr.h"

void cgo_method_bind_method_call(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionVariantPtr *p_args, const GDExtensionInt p_argument_count, GDExtensionVariantPtr r_return, GDExtensionCallError *r_error);
void cgo_method_bind_method_ptrcall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionTypePtr *p_args, GDExtensionTypePtr r_ret);

#endif
