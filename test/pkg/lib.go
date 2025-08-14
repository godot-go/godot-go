package pkg

/*
#cgo CFLAGS: -I${SRCDIR} -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
*/
import "C"

import (
	"runtime"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/core"
	"github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
)

var (
	pnr runtime.Pinner
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

var threadName = "godot-go"

//export TestDemoInit
func TestDemoInit(p_get_proc_address unsafe.Pointer, p_library unsafe.Pointer, r_initialization unsafe.Pointer) bool {
	initObj := NewInitObject(
		(ffi.GDExtensionInterfaceGetProcAddress)(p_get_proc_address),
		(ffi.GDExtensionClassLibraryPtr)(p_library),
		(*ffi.GDExtensionInitialization)(r_initialization),
	)

	initObj.RegisterSceneInitializer(RegisterExampleTypes)
	initObj.RegisterSceneTerminator(UnregisterExampleTypes)

	return initObj.Init()
}
