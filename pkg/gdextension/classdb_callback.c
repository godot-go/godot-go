#include <stdbool.h>
#include <stdio.h>
#include <godot/gdextension_interface.h>
#include "classdb_callback.h"
#include "stacktrace.h"

void GoCallback_ClassCreationInfoGetPropertyList(GDExtensionClassInstancePtr p_instance, uint32_t *r_count);
void GoCallback_ClassCreationInfoFreePropertyList(GDExtensionClassInstancePtr p_instance, uint32_t *r_count);
GDExtensionBool GoCallback_ClassCreationInfoPropertyCanRevert(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name);
GDExtensionBool GoCallback_ClassCreationInfoPropertyGetRevert(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionVariantPtr r_ret);
extern void* GoCallback_ClassCreationInfoGetVirtualCallWithData(void *p_userdata, GDExtensionConstStringNamePtr p_name);
extern void GoCallback_ClassCreationInfoCallVirtualWithData(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, void *p_userdata, const GDExtensionConstTypePtr *p_args, GDExtensionTypePtr r_ret);
extern GDExtensionBool GoCallback_ClassCreationInfoGet(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionVariantPtr r_ret);
extern GDExtensionBool GoCallback_ClassCreationInfoSet(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionConstVariantPtr p_value);
extern void GoCallback_ClassCreationInfoToString(GDExtensionClassInstancePtr p_instance, GDExtensionBool *r_is_valid, GDExtensionStringPtr p_out);
extern GDExtensionClassCallVirtual GoCallback_ClassCreationInfoGetVirtual(void *p_userdata, GDExtensionConstStringNamePtr p_name);
extern GDExtensionObjectPtr GoCallback_ClassCreationInfoCreateInstance(void *data);
extern void GoCallback_ClassCreationInfoFreeInstance(void *data, GDExtensionClassInstancePtr ptr);

void cgo_classcreationinfo_getpropertylist(GDExtensionClassInstancePtr p_instance, uint32_t *r_count) {
    printStacktrace();
    GoCallback_ClassCreationInfoGetPropertyList(p_instance, r_count);
}

void cgo_classcreationinfo_freepropertylist(GDExtensionClassInstancePtr p_instance, uint32_t *r_count) {
    printStacktrace();
    GoCallback_ClassCreationInfoFreePropertyList(p_instance, r_count);
}

GDExtensionBool cgo_classcreationinfo_propertycanrevert(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name) {
    printStacktrace();
    return GoCallback_ClassCreationInfoPropertyCanRevert(p_instance, p_name);
}

GDExtensionBool cgo_classcreationinfo_propertygetrevert(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionVariantPtr r_ret) {
    printStacktrace();
    return GoCallback_ClassCreationInfoPropertyGetRevert(p_instance, p_name, r_ret);
}

void* cgo_classcreationinfo_getvirtualcallwithdata(void *p_userdata, GDExtensionConstStringNamePtr p_name) {
    printStacktrace();
    return GoCallback_ClassCreationInfoGetVirtualCallWithData(p_userdata, p_name);
}

void cgo_classcreationinfo_callvirtualwithdata(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, void *p_userdata, const GDExtensionConstTypePtr *p_args, GDExtensionTypePtr r_ret) {
    printStacktrace();
    GoCallback_ClassCreationInfoCallVirtualWithData(p_instance, p_name, p_userdata, p_args, r_ret);
}

void cgo_classcreationinfo_tostring(GDExtensionClassInstancePtr p_instance, GDExtensionBool *r_is_valid, GDExtensionStringPtr p_out) {
    printStacktrace();
    GoCallback_ClassCreationInfoToString(p_instance, r_is_valid, p_out);
}

GDExtensionObjectPtr cgo_classcreationinfo_createinstance(void *data) {
    printStacktrace();
    return GoCallback_ClassCreationInfoCreateInstance(data);
}

void cgo_classcreationinfo_freeinstance(void *data, GDExtensionClassInstancePtr ptr) {
    printStacktrace();
    GoCallback_ClassCreationInfoFreeInstance(data, ptr);
}

GDExtensionBool cgo_classcreationinfo_get(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionVariantPtr r_ret) {
    printStacktrace();
    return GoCallback_ClassCreationInfoGet(p_instance, p_name, r_ret);
}

GDExtensionBool cgo_classcreationinfo_set(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionConstVariantPtr p_value) {
    printStacktrace();
    return GoCallback_ClassCreationInfoSet(p_instance, p_name, p_value);
}