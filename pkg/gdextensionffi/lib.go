package gdextensionffi

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextensionffi
#include <godot/gdextension_interface.h>
#include "ffi_wrapper.gen.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"

	"github.com/godot-go/godot-go/pkg/util"
)

func NewGDExtensionInstanceBindingCallbacks(
	createCallback GDExtensionInstanceBindingCreateCallback,
	freeCallback GDExtensionInstanceBindingFreeCallback,
	referenceCallback GDExtensionInstanceBindingReferenceCallback,
) GDExtensionInstanceBindingCallbacks {
	return GDExtensionInstanceBindingCallbacks{
		create_callback:    (C.GDExtensionInstanceBindingCreateCallback)(createCallback),
		free_callback:      (C.GDExtensionInstanceBindingFreeCallback)(freeCallback),
		reference_callback: (C.GDExtensionInstanceBindingReferenceCallback)(referenceCallback),
	}
}

func (e *GDExtensionInitialization) SetCallbacks(
	initCallback *[0]byte,
	deinitCallback *[0]byte,
) {
	e.initialize = initCallback
	e.deinitialize = deinitCallback
}

func (e *GDExtensionInitialization) SetInitializationLevel(level GDExtensionInitializationLevel) {
	e.minimum_initialization_level = (C.GDExtensionInitializationLevel)(level)
}

func NewGDExtensionPropertyInfo(
	className GDExtensionConstStringNamePtr,
	propertyType GDExtensionVariantType,
	propertyName GDExtensionConstStringNamePtr,
	hint uint32,
	hintString GDExtensionConstStringPtr,
	usage uint32,
) GDExtensionPropertyInfo {
	return (GDExtensionPropertyInfo)(C.GDExtensionPropertyInfo{
		_type:       (C.GDExtensionVariantType)(propertyType),
		name:        (C.GDExtensionStringNamePtr)(unsafe.Pointer(propertyName)),
		class_name:  (C.GDExtensionStringNamePtr)(unsafe.Pointer(className)),
		hint:        (C.uint32_t)(hint),
		hint_string: (C.GDExtensionStringPtr)(unsafe.Pointer(hintString)),
		usage:       (C.uint32_t)(usage),
	})
}

func (p *GDExtensionPropertyInfo) Destroy() {
	cp := (*C.GDExtensionPropertyInfo)(p)

	C.free(unsafe.Pointer(cp.name))
	C.free(unsafe.Pointer(cp.class_name))
	C.free(unsafe.Pointer(cp.hint_string))
}

func NewGDExtensionClassMethodInfo(
	name GDExtensionConstStringNamePtr,
	methodUserdata unsafe.Pointer,
	callFunc GDExtensionClassMethodCall,
	ptrcallFunc GDExtensionClassMethodPtrCall,
	methodFlags uint32,
	hasReturnValue bool,
	returnValueInfo *GDExtensionPropertyInfo,
	returnValueMetadata GDExtensionClassMethodArgumentMetadata,
	argumentCount uint32,
	argumentsInfo *GDExtensionPropertyInfo,
	argumentsMetadata *GDExtensionClassMethodArgumentMetadata,
	defaultArgumentCount uint32,
	defaultArguments *GDExtensionVariantPtr,
) GDExtensionClassMethodInfo {
	return (GDExtensionClassMethodInfo)(C.GDExtensionClassMethodInfo{
		name:            (C.GDExtensionStringNamePtr)(unsafe.Pointer(name)),
		method_userdata: methodUserdata,
		call_func:       (C.GDExtensionClassMethodCall)(callFunc),
		ptrcall_func:    (C.GDExtensionClassMethodPtrCall)(ptrcallFunc),

		// Bitfield of `GDExtensionClassMethodFlags`.
		method_flags: (C.uint32_t)(methodFlags),

		/* If `has_return_value` is false, `return_value_info` and `return_value_metadata` are ignored. */
		has_return_value:      (C.GDExtensionBool)(util.BoolToUint8(hasReturnValue)),
		return_value_info:     (*C.GDExtensionPropertyInfo)(returnValueInfo),
		return_value_metadata: (C.GDExtensionClassMethodArgumentMetadata)(returnValueMetadata),

		/* Arguments: `arguments_info` and `arguments_metadata` are array of size `argument_count`.
		* Name and hint information for the argument can be omitted in release builds. Class name should always be present if it applies.
		 */
		argument_count:     (C.uint32_t)(argumentCount),
		arguments_info:     (*C.GDExtensionPropertyInfo)(argumentsInfo),
		arguments_metadata: (*C.GDExtensionClassMethodArgumentMetadata)(argumentsMetadata),

		/* Default arguments: `default_arguments` is an array of size `default_argument_count`. */
		default_argument_count: (C.uint32_t)(defaultArgumentCount),
		default_arguments:      (*C.GDExtensionVariantPtr)(defaultArguments),
	})
}

func (m *GDExtensionClassMethodInfo) Destroy() {
	cm := (*C.GDExtensionClassMethodInfo)(m)

	C.free(unsafe.Pointer(cm.name))
}

func NewGDExtensionClassCreationInfo(
	createInstanceFunc GDExtensionClassCreateInstance,
	freeInstanceFunc GDExtensionClassFreeInstance,
	getVirtualFunc GDExtensionClassGetVirtual,
	classUserdata unsafe.Pointer,
) GDExtensionClassCreationInfo {
	return (GDExtensionClassCreationInfo)(C.GDExtensionClassCreationInfo{
		create_instance_func: (C.GDExtensionClassCreateInstance)(createInstanceFunc),
		free_instance_func:   (C.GDExtensionClassFreeInstance)(freeInstanceFunc),
		get_virtual_func:     (C.GDExtensionClassGetVirtual)(getVirtualFunc),
		class_userdata:       classUserdata,
	})
}
