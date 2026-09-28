package ffi

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/ffi
#include <godot/gdextension_interface.h>
#include "ffi_wrapper.gen.h"
#include "stacktrace.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
	// "github.com/CannibalVox/cgoalloc"
)

var (
	// pnr is for program-lifetime retention only: values the engine keeps a
	// pointer to after the call that registers them returns. Per-call scratch
	// must use a body-scoped runtime.Pinner with defer Unpin() instead.
	pnr = runtime.Pinner{}
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

func (e *GDExtensionInitialization) SetUserData(
	userdata unsafe.Pointer,
) {
	e.userdata = userdata
}

func (e *GDExtensionInitialization) SetInitializationLevel(level GDExtensionInitializationLevel) {
	e.minimum_initialization_level = (C.GDExtensionInitializationLevel)(level)
}

func init() {
	FFI.GodotVersion = new(GDExtensionGodotVersion)
	pnr.Pin(FFI.GodotVersion)
	// mem := cgoalloc.DefaultAllocator{}

	// FFI.GodotVersion = (*GDExtensionGodotVersion)(mem.Malloc(int(unsafe.Sizeof(GDExtensionGodotVersion{}))))
}
