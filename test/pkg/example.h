#ifndef GODOT_GO_TEST_EXAMPLE_H
#define GODOT_GO_TEST_EXAMPLE_H
#include <godot/gdnative_interface.h>

extern void Example_Ready(void *inst);

void cgo_callback_example_ready(GDExtensionClassInstancePtr p_instance, const GDNativeTypePtr *p_args, GDNativeTypePtr r_ret);

#endif
