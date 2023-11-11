#include <godot/gdextension_interface.h>
#include "gdextension_binding_init.h"
#include "stacktrace.h"

void cgo_callfn_GDExtensionBindingInitializeLevel(void *userdata, GDExtensionInitializationLevel p_level) {
	printStacktrace();
    GDExtensionBindingInitializeLevel(userdata, p_level);
}

void cgo_callfn_GDExtensionBindingDeinitializeLevel(void *userdata, GDExtensionInitializationLevel p_level) {
	printStacktrace();
    GDExtensionBindingDeinitializeLevel(userdata, p_level);
}
