#include <godot/gdnative_interface.h>
#include "gdextension_binding_init.h"
#include "stacktrace.h"

void cgo_callfn_GDExtensionBindingInitializeLevel(void *userdata, GDNativeInitializationLevel p_level) {
	printStacktrace();
    GDExtensionBindingInitializeLevel(userdata, p_level);
}

void cgo_callfn_GDExtensionBindingDeinitializeLevel(void *userdata, GDNativeInitializationLevel p_level) {
	printStacktrace();
    GDExtensionBindingDeinitializeLevel(userdata, p_level);
}
