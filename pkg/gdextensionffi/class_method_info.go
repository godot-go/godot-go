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
	stringNameDestructor := (GDExtensionPtrDestructor)(CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING_NAME))
	CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(m.name))
	cm := (*C.GDExtensionClassMethodInfo)(m)
	if cm != nil {
		CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(unsafe.Pointer(cm.name)))
	}
	if cm.return_value_info != nil {
		(*GDExtensionPropertyInfo)(cm.return_value_info).Destroy()
	}
	if cm.argument_count > 0 && cm.arguments_info != nil {
		(*GDExtensionPropertyInfo)(cm.arguments_info).Destroy()
	}
}