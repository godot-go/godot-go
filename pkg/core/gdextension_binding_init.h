#ifndef CGO_GODOT_GO_GDEXTENSION_BINDING_INIT_H
#define CGO_GODOT_GO_GDEXTENSION_BINDING_INIT_H

#include <godot/gdextension_interface.h>

// TODO: do we need this here?
void GDExtensionBindingInitializeLevel(void *userdata, GDExtensionInitializationLevel p_level);
void GDExtensionBindingDeinitializeLevel(void *userdata, GDExtensionInitializationLevel p_level);

void cgo_callfn_GDExtensionBindingInitializeLevel(void *userdata, GDExtensionInitializationLevel p_level);
void cgo_callfn_GDExtensionBindingDeinitializeLevel(void *userdata, GDExtensionInitializationLevel p_level);

#endif
