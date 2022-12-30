#include <stdbool.h>
#include <stdio.h>
#include <godot/gdextension_interface.h>
#include "classdb_wrapper.h"
#include "stacktrace.h"

extern GDExtensionClassCallVirtual GoCallback_ClassDBGetVirtualFunc(void *p_userdata, GDExtensionConstStringNamePtr p_name);
extern GDExtensionObjectPtr GoCallback_GDExtensionClassCreateInstance(void *data);
extern void GoCallback_GDExtensionClassFreeInstance(void *data, GDExtensionClassInstancePtr ptr);

GDExtensionClassCallVirtual cgo_classdb_get_virtual_func(void *p_userdata, GDExtensionConstStringNamePtr p_name) {
    printStacktrace();
    return GoCallback_ClassDBGetVirtualFunc(p_userdata, p_name);
}

GDExtensionObjectPtr cgo_gdextension_extension_class_create_instance(void *data) {
    printStacktrace();
    return GoCallback_GDExtensionClassCreateInstance(data);
}

void cgo_gdextension_extension_class_free_instance(void *data, GDExtensionClassInstancePtr ptr) {
    printStacktrace();
    return GoCallback_GDExtensionClassFreeInstance(data, ptr);
}
