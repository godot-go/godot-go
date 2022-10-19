package gdnative

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdnative
#include <godot/gdnative_interface.h>
#include "gdnative_wrapper.gen.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"

	"github.com/godot-go/godot-go/pkg/util"
)

func NewGDNativeInstanceBindingCallbacks(
	createCallback GDNativeInstanceBindingCreateCallback,
	freeCallback GDNativeInstanceBindingFreeCallback,
	referenceCallback GDNativeInstanceBindingReferenceCallback,
) GDNativeInstanceBindingCallbacks {
	return GDNativeInstanceBindingCallbacks{
		create_callback:    (C.GDNativeInstanceBindingCreateCallback)(createCallback),
		free_callback:      (C.GDNativeInstanceBindingFreeCallback)(freeCallback),
		reference_callback: (C.GDNativeInstanceBindingReferenceCallback)(referenceCallback),
	}
}

func (e *GDNativeInitialization) SetCallbacks(
	initCallback *[0]byte,
	deinitCallback *[0]byte,
) {
	e.initialize = initCallback
	e.deinitialize = deinitCallback
}

func (e *GDNativeInitialization) SetInitializationLevel(level GDNativeInitializationLevel) {
	e.minimum_initialization_level = (C.GDNativeInitializationLevel)(level)
}

func NewGDNativePropertyInfo(
	className string,
	propertyType GDNativeVariantType,
	propertyName string,
	hint uint32,
	hintString string,
	usage uint32,
) GDNativePropertyInfo {
	cClassName := C.CString(className)
	cPropName := C.CString(propertyName)
	cHintString := C.CString(hintString)

	pi := C.GDNativePropertyInfo{
		_type:       (C.GDNativeVariantType)(propertyType),
		name:        cPropName,
		class_name:  cClassName,
		hint:        (C.uint32_t)(hint),
		hint_string: cHintString,
		usage:       (C.uint32_t)(usage),
	}

	return (GDNativePropertyInfo)(pi)
}

func (p *GDNativePropertyInfo) Destroy() {
	cp := (*C.GDNativePropertyInfo)(p)

	C.free(unsafe.Pointer(cp.name))
	C.free(unsafe.Pointer(cp.class_name))
	C.free(unsafe.Pointer(cp.hint_string))
}

func NewGDNativeExtensionClassMethodInfo(
	name string,
	methodUserdata unsafe.Pointer,
	callFunc GDNativeExtensionClassMethodCall,
	ptrcallFunc GDNativeExtensionClassMethodPtrCall,
	methodFlags uint32,
	argumentCount uint32,
	hasReturnValue bool,
	getArgumentTypeFunc GDNativeExtensionClassMethodGetArgumentType,
	getArgumentInfoFunc GDNativeExtensionClassMethodGetArgumentInfo,
	getArgumentMetadataFunc GDNativeExtensionClassMethodGetArgumentMetadata,
	defaultArgumentCount uint32,
	defaultArguments *GDNativeVariantPtr,
) GDNativeExtensionClassMethodInfo {
	cName := C.CString(name)

	return (GDNativeExtensionClassMethodInfo)(C.GDNativeExtensionClassMethodInfo{
		name:                       cName,
		method_userdata:            methodUserdata,
		call_func:                  (C.GDNativeExtensionClassMethodCall)(callFunc),
		ptrcall_func:               (C.GDNativeExtensionClassMethodPtrCall)(ptrcallFunc),
		method_flags:               (C.uint32_t)(methodFlags),
		argument_count:             (C.uint32_t)(argumentCount),
		has_return_value:           (C.GDNativeBool)(util.BoolToUint8(hasReturnValue)),
		get_argument_type_func:     (C.GDNativeExtensionClassMethodGetArgumentType)(getArgumentTypeFunc),
		get_argument_info_func:     (C.GDNativeExtensionClassMethodGetArgumentInfo)(getArgumentInfoFunc),
		get_argument_metadata_func: (C.GDNativeExtensionClassMethodGetArgumentMetadata)(getArgumentMetadataFunc),
		default_argument_count:     (C.uint32_t)(defaultArgumentCount),
		default_arguments:          (*C.GDNativeVariantPtr)(defaultArguments),
	})
}

func (m *GDNativeExtensionClassMethodInfo) Destroy() {
	cm := (*C.GDNativeExtensionClassMethodInfo)(m)

	C.free(unsafe.Pointer(cm.name))
}

func NewGDNativeExtensionClassCreationInfo(
	createInstanceFunc GDNativeExtensionClassCreateInstance,
	freeInstanceFunc GDNativeExtensionClassFreeInstance,
	getVirtualFunc GDNativeExtensionClassGetVirtual,
	classUserdata unsafe.Pointer,
) GDNativeExtensionClassCreationInfo {
	return (GDNativeExtensionClassCreationInfo)(C.GDNativeExtensionClassCreationInfo{
		create_instance_func: (C.GDNativeExtensionClassCreateInstance)(createInstanceFunc),
		free_instance_func:   (C.GDNativeExtensionClassFreeInstance)(freeInstanceFunc),
		get_virtual_func:     (C.GDNativeExtensionClassGetVirtual)(getVirtualFunc),
		class_userdata:       classUserdata,
	})
}
