#include "example.h"

void cgo_callback_example_ready(GDExtensionClassInstancePtr p_instance, const GDNativeTypePtr *p_args, GDNativeTypePtr r_ret) {
	Example_Ready((void*)(p_instance));
}
