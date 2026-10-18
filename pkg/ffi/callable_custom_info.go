package ffi

/*
#cgo CFLAGS: -I${SRCDIR}/../../../godot_headers -I${SRCDIR}/../../../pkg/log -I${SRCDIR}/../../../pkg/gdextension/ffi
#include <godot/gdextension_interface.h>
#include "ffi_wrapper.gen.h"
#include "cgo_ptr.h"
*/
import "C"
import (
	"runtime/cgo"
	"unsafe"
)

// NewGDExtensionCallableCustomInfo2 builds the info struct consumed by
// callable_custom_create2. Only call_func and token are strictly required by
// the engine contract; pass zero/nil for the optional callbacks to let the
// engine use the (call_func, callable_userdata) pair as the identity of the
// callable for hashing and comparison.
func NewGDExtensionCallableCustomInfo2(
	callableUserdata cgo.Handle,
	token unsafe.Pointer,
	objectId GDObjectInstanceID,
	callFunc GDExtensionCallableCustomCall,
	isValidFunc GDExtensionCallableCustomIsValid,
	freeFunc GDExtensionCallableCustomFree,
	hashFunc GDExtensionCallableCustomHash,
	equalFunc GDExtensionCallableCustomEqual,
	lessThanFunc GDExtensionCallableCustomLessThan,
	toStringFunc GDExtensionCallableCustomToString,
) GDExtensionCallableCustomInfo2 {
	info := (C.GDExtensionCallableCustomInfo2)(C.GDExtensionCallableCustomInfo2{
		token:                   token,
		object_id:               C.GDObjectInstanceID(objectId),
		call_func:               (C.GDExtensionCallableCustomCall)(callFunc),
		is_valid_func:           (C.GDExtensionCallableCustomIsValid)(isValidFunc),
		free_func:               (C.GDExtensionCallableCustomFree)(freeFunc),
		hash_func:               (C.GDExtensionCallableCustomHash)(hashFunc),
		equal_func:              (C.GDExtensionCallableCustomEqual)(equalFunc),
		less_than_func:          (C.GDExtensionCallableCustomLessThan)(lessThanFunc),
		to_string_func:          (C.GDExtensionCallableCustomToString)(toStringFunc),
		get_argument_count_func: (C.GDExtensionCallableCustomGetArgumentCount)(nil),
	})
	// The callable userdata is a packed cgo.Handle; write it through the C
	// shim so the integer never occupies a Go pointer slot (openspec
	// fix-instance-handle-lifecycle).
	C.cgo_set_handle_slot(
		(*C.uintptr_t)(unsafe.Pointer(&info.callable_userdata)),
		C.uintptr_t(callableUserdata),
	)
	return (GDExtensionCallableCustomInfo2)(info)
}
