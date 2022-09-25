#ifndef CGO_GODOT_GO_GDEXTENSION_BINDING_INIT_H
#define CGO_GODOT_GO_GDEXTENSION_BINDING_INIT_H

#include <godot/gdnative_interface.h>

// TODO: do we need this here?
void GDExtensionBindingInitializeLevel(void *userdata, GDNativeInitializationLevel p_level);
void GDExtensionBindingDeinitializeLevel(void *userdata, GDNativeInitializationLevel p_level);

void cgo_callfn_GDExtensionBindingInitializeLevel(void *userdata, GDNativeInitializationLevel p_level);
void cgo_callfn_GDExtensionBindingDeinitializeLevel(void *userdata, GDNativeInitializationLevel p_level);

#endif
