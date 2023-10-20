package pkg

/*
#cgo CFLAGS: -I${SRCDIR} -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
*/
import "C"

import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextension"
	"github.com/godot-go/godot-go/pkg/gdextensionffi"
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
	log.Debug("TestDemoInit called")
	initObj := NewInitObject(
		(gdextensionffi.GDExtensionInterfaceGetProcAddress)(p_get_proc_address),
		(gdextensionffi.GDExtensionClassLibraryPtr)(p_library),
		(*gdextensionffi.GDExtensionInitialization)(r_initialization),
	)

	initObj.RegisterSceneInitializer(RegisterExampleTypes)
	initObj.RegisterSceneTerminator(UnregisterExampleTypes)

	return initObj.Init()
}
