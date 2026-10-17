package core

//revive:disable

/*
#include <godot/gdextension_interface.h>
#include "main_loop_callback.h"
*/
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
)

// Main-loop phase callbacks carry no user data (all three C signatures are
// void(void)), so the exported trampolines below dispatch to the package
// slots in lib.go. There is no panic recovery here: a panicking hook aborts
// the process, matching every other GoCallback_* trampoline and godot-cpp's
// unrecovered calls.

//export GoCallback_MainLoopStartup
func GoCallback_MainLoopStartup() {
	if fn := GDExtensionBindingMainLoopStartupCallback; fn != nil {
		fn()
	}
}

//export GoCallback_MainLoopFrame
func GoCallback_MainLoopFrame() {
	if fn := GDExtensionBindingMainLoopFrameCallback; fn != nil {
		fn()
	}
}

//export GoCallback_MainLoopShutdown
func GoCallback_MainLoopShutdown() {
	if fn := GDExtensionBindingMainLoopShutdownCallback; fn != nil {
		fn()
	}
}

// registerMainLoopCallbacks hands the phase trampolines to the engine,
// mirroring godot-cpp's godot.cpp initialize_level CORE-level registration:
// it must be called once, at the core initialization level, after the hook
// slots have their final values. The engine reads the struct synchronously
// and copies the three function pointers into its own storage, so the Go
// local is safe to drop when the call returns.
func registerMainLoopCallbacks() {
	if FFI.RegisterMainLoopCallbacks == nil {
		// Engine without register_main_loop_callbacks (pre-4.5): the loader
		// already warned about the missing interface function. Skip
		// registration so the extension keeps loading.
		log.Debug("main loop registration interface function unavailable; skipping main-loop callback registration")
		return
	}
	callbacks := C.cgo_main_loop_callbacks_build(
		C.bool(GDExtensionBindingMainLoopStartupCallback != nil),
		C.bool(GDExtensionBindingMainLoopFrameCallback != nil),
		C.bool(GDExtensionBindingMainLoopShutdownCallback != nil),
	)
	CallFunc_GDExtensionInterfaceRegisterMainLoopCallbacks(
		FFI.Library,
		(*GDExtensionMainLoopCallbacks)(unsafe.Pointer(&callbacks)),
	)
	log.Debug("registered main loop callbacks")
}
