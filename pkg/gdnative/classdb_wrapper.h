#ifndef CGO_GDNATIVE_CLASSDB_WRAPPER_H
#define CGO_GDNATIVE_CLASSDB_WRAPPER_H

#include <godot/gdnative_interface.h>

GDNativeExtensionClassCallVirtual cgo_classdb_get_virtual_func(void *p_userdata, const char *p_name);
GDNativeObjectPtr cgo_gdnative_extension_class_create_instance(void *data);
void cgo_gdnative_extension_class_free_instance(void *data, GDExtensionClassInstancePtr ptr);

#endif
