package pkg_test

/*
#cgo CFLAGS: -I${SRCDIR} -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
*/
import "C"

import (
	"runtime"
	"unsafe"

	_ "github.com/godot-go/godot-go/pkg/core"
	_ "github.com/godot-go/godot-go/pkg/ffi"
	_ "github.com/godot-go/godot-go/pkg/log"
)

var (
	pnr runtime.Pinner
)

//export TestDemoInit
func TestDemoInit(p_get_proc_address unsafe.Pointer, p_library unsafe.Pointer, r_initialization unsafe.Pointer) bool {
	return false
}
