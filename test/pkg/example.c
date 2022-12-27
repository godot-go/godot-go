#include "example.h"

void cgo_callback_example_ready(GDExtensionClassInstancePtr p_instance, const GDExtensionTypePtr *p_args, GDExtensionTypePtr r_ret) {
	Example_Ready((void*)(p_instance));
}
