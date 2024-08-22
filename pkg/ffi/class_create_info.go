package ffi

/*
#cgo CFLAGS: -I${SRCDIR}/../../../godot_headers -I${SRCDIR}/../../../pkg/log -I${SRCDIR}/../../../pkg/gdextension/ffi
#include <godot/gdextension_interface.h>
#include "ffi_wrapper.gen.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"
)

func NewGDExtensionClassCreationInfo3(
	isVirtual GDExtensionBool,
	isAbstract GDExtensionBool,
	isExposed GDExtensionBool,
	isRuntime GDExtensionBool,
	createInstanceFunc GDExtensionClassCreateInstance,
	freeInstanceFunc GDExtensionClassFreeInstance,
	getVirtualCallDataFunc GDExtensionClassGetVirtualCallData,
	callVirtualFunc GDExtensionClassCallVirtualWithData,
	toStringFunc GDExtensionClassToString,
	setFunc GDExtensionClassSet,
	getFunc GDExtensionClassGet,
	getPropertyListFunc GDExtensionClassGetPropertyList,
	freePropertyListFunc GDExtensionClassFreePropertyList2,
	propertyCanRevertFunc GDExtensionClassPropertyCanRevert,
	propertyGetRevertFunc GDExtensionClassPropertyGetRevert,
	validatePropertyFunc GDExtensionClassValidateProperty,
	notificationFunc GDExtensionClassNotification2,
	classUserdata unsafe.Pointer,
) GDExtensionClassCreationInfo3 {
	return (GDExtensionClassCreationInfo3)(C.GDExtensionClassCreationInfo3{
		is_virtual:                  (C.GDExtensionBool)(isVirtual),
		is_abstract:                 (C.GDExtensionBool)(isAbstract),
		is_exposed:                  (C.GDExtensionBool)(isExposed),
		is_runtime:                  (C.GDExtensionBool)(isRuntime),
		create_instance_func:        (C.GDExtensionClassCreateInstance)(createInstanceFunc),
		free_instance_func:          (C.GDExtensionClassFreeInstance)(freeInstanceFunc),
		get_virtual_call_data_func:  (C.GDExtensionClassGetVirtualCallData)(getVirtualCallDataFunc),
		call_virtual_with_data_func: (C.GDExtensionClassCallVirtualWithData)(callVirtualFunc),
		to_string_func:              (C.GDExtensionClassToString)(toStringFunc),
		set_func:                    (C.GDExtensionClassSet)(setFunc),
		get_func:                    (C.GDExtensionClassGet)(getFunc),
		get_property_list_func:      (C.GDExtensionClassGetPropertyList)(getPropertyListFunc),
		free_property_list_func:     (C.GDExtensionClassFreePropertyList2)(freePropertyListFunc),
		property_can_revert_func:    (C.GDExtensionClassPropertyCanRevert)(propertyCanRevertFunc),
		property_get_revert_func:    (C.GDExtensionClassPropertyGetRevert)(propertyGetRevertFunc),
		validate_property_func:      (C.GDExtensionClassValidateProperty)(validatePropertyFunc),
		notification_func:           (C.GDExtensionClassNotification2)(notificationFunc),
		class_userdata:              classUserdata,
	})
}

func (m *GDExtensionClassCreationInfo2) Destroy() {
	cm := (*C.GDExtensionClassCreationInfo2)(m)
	if cm.class_userdata != nil {
		C.free(cm.class_userdata)
	}
}
