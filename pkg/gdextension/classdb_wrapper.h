#ifndef CGO_GODOT_GO_CLASSDB_WRAPPER_H
#define CGO_GODOT_GO_CLASSDB_WRAPPER_H

#include <godot/gdextension_interface.h>

GDExtensionClassCallVirtual cgo_classdb_get_virtual_func(void *p_userdata, const char *p_name);
GDExtensionObjectPtr cgo_gdextension_extension_class_create_instance(void *data);
void cgo_gdextension_extension_class_free_instance(void *data, GDExtensionClassInstancePtr ptr);

#endif
