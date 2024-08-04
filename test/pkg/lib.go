package pkg

/*
#cgo CFLAGS: -I${SRCDIR} -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
*/
import "C"

import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/core"
	"github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
)

func RegisterExampleTypes() {
	log.Debug("RegisterExampleTypes called")

	RegisterClassExampleRef()
	RegisterClassExample()
}

func UnregisterExampleTypes() {
	log.Debug("UnregisterExampleTypes called")

	UnregisterClassExample()
}

//export TestDemoInit
func TestDemoInit(p_get_proc_address unsafe.Pointer, p_library unsafe.Pointer, r_initialization unsafe.Pointer) bool {
	log.Info("TestDemoInit called")
	initObj := NewInitObject(
		(ffi.GDExtensionInterfaceGetProcAddress)(p_get_proc_address),
		(ffi.GDExtensionClassLibraryPtr)(p_library),
		(*ffi.GDExtensionInitialization)(r_initialization),
	)

	initObj.RegisterSceneInitializer(RegisterExampleTypes)
	initObj.RegisterSceneTerminator(UnregisterExampleTypes)

	return initObj.Init()
}
