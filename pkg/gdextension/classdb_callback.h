#ifndef CGO_GODOT_GO_CLASSDB_WRAPPER_H
#define CGO_GODOT_GO_CLASSDB_WRAPPER_H

#include <godot/gdextension_interface.h>

// cgo_classcreationinfo_tostring should match GDExtensionClassToString
void cgo_classcreationinfo_tostring(GDExtensionClassInstancePtr p_instance, GDExtensionBool *r_is_valid, GDExtensionStringPtr p_out);

// cgo_classcreationinfo_getvirtualcallwithdata should match GDExtensionClassGetVirtuaCallData
// callback when godot wants to get the virtual function call
void* cgo_classcreationinfo_getvirtualcallwithdata(void *p_userdata, GDExtensionConstStringNamePtr p_name);

// cgo_classcreationinfo_callvirtualwithdata should match GDExtensionClassCallVirtualWithData
// callback when godot wants to call a method in go marked as a virtual
void cgo_classcreationinfo_callvirtualwithdata(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, void *p_userdata, const GDExtensionConstTypePtr *p_args, GDExtensionTypePtr r_ret);

// cgo_classcreationinfo_createinstance signature should match GDExtensionClassCreateInstance
GDExtensionObjectPtr cgo_classcreationinfo_createinstance(void *data);

// cgo_classcreationinfo_freeinstance signature shuold match GDExtensionClassFreeInstance
void cgo_classcreationinfo_freeinstance(void *data, GDExtensionClassInstancePtr ptr);

// TODO: implement code to utilize _get _set below

// cgo_classdb_get_func should match GDExtensionClassGet
// callback when godot wants to get a property of a class
GDExtensionBool cgo_classdb_get_func(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionVariantPtr r_ret);

// cgo_classdb_set_func should match GDExtensionClassSet
// callback when godot wants to set a property of a class
GDExtensionBool cgo_classdb_set_func(GDExtensionClassInstancePtr p_instance, GDExtensionConstStringNamePtr p_name, GDExtensionConstVariantPtr p_value);

#endif
