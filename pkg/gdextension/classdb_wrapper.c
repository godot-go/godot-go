#include <stdbool.h>
#include <stdio.h>
#include <godot/gdnative_interface.h>
#include "classdb_wrapper.h"
#include "stacktrace.h"

extern GDNativeExtensionClassCallVirtual GoCallback_ClassDBGetVirtualFunc(void *p_userdata, char *p_name);
extern GDNativeObjectPtr GoCallback_GDNativeExtensionClassCreateInstance(void *data);
extern void GoCallback_GDNativeExtensionClassFreeInstance(void *data, GDExtensionClassInstancePtr ptr);

GDNativeExtensionClassCallVirtual cgo_classdb_get_virtual_func(void *p_userdata, const char *p_name) {
    printStacktrace();
    return GoCallback_ClassDBGetVirtualFunc(p_userdata, (char *)p_name);
}

GDNativeObjectPtr cgo_gdnative_extension_class_create_instance(void *data) {
    printStacktrace();
    return GoCallback_GDNativeExtensionClassCreateInstance(data);
}

void cgo_gdnative_extension_class_free_instance(void *data, GDExtensionClassInstancePtr ptr) {
    printStacktrace();
    return GoCallback_GDNativeExtensionClassFreeInstance(data, ptr);
}
