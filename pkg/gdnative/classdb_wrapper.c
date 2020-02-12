#include <execinfo.h>
#include <stdio.h>
#include <godot/gdnative_interface.h>
#include "gdnative_binding_wrapper.h"

extern GDNativeExtensionClassCallVirtual GoCallback_ClassDBGetVirtualFunc(void *p_userdata, char *p_name);
extern GDNativeObjectPtr GoCallback_GDNativeExtensionClassCreateInstance(void *data);
extern void GoCallback_GDNativeExtensionClassFreeInstance(void *data, GDExtensionClassInstancePtr ptr);

GDNativeExtensionClassCallVirtual cgo_classdb_get_virtual_func(void *p_userdata, const char *p_name) {
    // output stacktrace for debugging issues
    printf("===============\ncgo_classdb_get_virtual_func(p_userdata=%s, p_name=%s) stacktrace\n", p_userdata, p_name);
    void* callstack[128];
    int i, frames = backtrace(callstack, 128);
    char** strs = backtrace_symbols(callstack, frames);
    for (i = 0; i < frames; ++i) {
        printf("%s\n", strs[i]);
    }
    free(strs);
    printf("===============\n\n");
    return GoCallback_ClassDBGetVirtualFunc(p_userdata, p_name);
}

GDNativeObjectPtr cgo_gdnative_extension_class_create_instance(void *data) {
    return GoCallback_GDNativeExtensionClassCreateInstance(data);
}


void cgo_gdnative_extension_class_free_instance(void *data, GDExtensionClassInstancePtr ptr) {
    return GoCallback_GDNativeExtensionClassFreeInstance(data, ptr);
}
