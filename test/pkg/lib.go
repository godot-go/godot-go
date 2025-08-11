package pkg

/*
#cgo CFLAGS: -I${SRCDIR} -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
*/
import "C"

import (
	"fmt"
	"runtime"
	"syscall"
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
	if runtime.GOOS == "linux" {
		// PR_SET_NAME is 15 on Linux
		_, _, errno := syscall.Syscall6(syscall.SYS_PRCTL, syscall.PR_SET_NAME, uintptr(unsafe.Pointer(syscall.StringBytePtr(threadName))), 0, 0, 0, 0)
		if errno != 0 {
			fmt.Printf("Error setting thread name: %v\n", errno)
		}
	}

	initObj := NewInitObject(
		(ffi.GDExtensionInterfaceGetProcAddress)(p_get_proc_address),
		(ffi.GDExtensionClassLibraryPtr)(p_library),
		(*ffi.GDExtensionInitialization)(r_initialization),
	)

	initObj.RegisterSceneInitializer(RegisterExampleTypes)
	initObj.RegisterSceneTerminator(UnregisterExampleTypes)

	return initObj.Init()
}
