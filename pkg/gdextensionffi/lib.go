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
)

var (
	nullptr = unsafe.Pointer(nil)
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
