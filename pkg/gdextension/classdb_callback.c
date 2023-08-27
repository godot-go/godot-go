#include <stdbool.h>
#include <stdio.h>
#include <godot/gdextension_interface.h>
#include "classdb_callback.h"
#include "stacktrace.h"

extern void* GoCallback_ClassCreationInfoGetVirtualCallWithData(void *p_userdata, GDExtensionConstStringNamePtr p_name);
extern void GoCallback_ClassCreationInfoCallVirtualWithData(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, void *p_userdata, const GDExtensionConstTypePtr *p_args, GDExtensionTypePtr r_ret);
extern GDExtensionBool GoCallback_ClassDBGetFunc(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionVariantPtr r_ret);
extern GDExtensionBool GoCallback_ClassDBSetFunc(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionConstVariantPtr p_value);
extern void GoCallback_ClassCreationInfoToString(GDExtensionClassInstancePtr p_instance, GDExtensionBool *r_is_valid, GDExtensionStringPtr p_out);
extern GDExtensionClassCallVirtual GoCallback_ClassCreationInfoGetVirtual(void *p_userdata, GDExtensionConstStringNamePtr p_name);
extern GDExtensionObjectPtr GoCallback_ClassCreationInfoCreateInstance(void *data);
extern void GoCallback_ClassCreationInfoFreeInstance(void *data, GDExtensionClassInstancePtr ptr);

void* cgo_classcreationinfo_getvirtualcallwithdata(void *p_userdata, GDExtensionConstStringNamePtr p_name) {
    printStacktrace();
    return GoCallback_ClassCreationInfoGetVirtualCallWithData(p_userdata, p_name);
}

void cgo_classcreationinfo_callvirtualwithdata(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, void *p_userdata, const GDExtensionConstTypePtr *p_args, GDExtensionTypePtr r_ret) {
    printStacktrace();
    return GoCallback_ClassCreationInfoCallVirtualWithData(p_instance, p_name, p_userdata, p_args, r_ret);
}

void cgo_classcreationinfo_tostring(GDExtensionClassInstancePtr p_instance, GDExtensionBool *r_is_valid, GDExtensionStringPtr p_out) {
    printStacktrace();
    return GoCallback_ClassCreationInfoToString(p_instance, r_is_valid, p_out);
}

// GDExtensionClassCallVirtual cgo_classdb_get_virtual_func(void *p_userdata, GDExtensionConstStringNamePtr p_name) {
//     printStacktrace();
//     return GoCallback_ClassCreationInfoGetVirtual(p_userdata, p_name);
// }

GDExtensionObjectPtr cgo_classcreationinfo_createinstance(void *data) {
    printStacktrace();
    return GoCallback_ClassCreationInfoCreateInstance(data);
}

void cgo_classcreationinfo_freeinstance(void *data, GDExtensionClassInstancePtr ptr) {
    printStacktrace();
    return GoCallback_ClassCreationInfoFreeInstance(data, ptr);
}

GDExtensionBool cgo_classdb_get_func(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionVariantPtr r_ret) {
    printStacktrace();
    return GoCallback_ClassDBGetFunc(p_instance, p_name, r_ret);
}

GDExtensionBool cgo_classdb_set_func(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionConstVariantPtr p_value) {
    printStacktrace();
    return GoCallback_ClassDBSetFunc(p_instance, p_name, p_value);
}