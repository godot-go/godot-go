package gdutilfunc

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdutilfunc
#include <godot/gdextension_interface.h>
*/
import "C"
import (
	"runtime"
	"unsafe"
)

var (
	nullptr = unsafe.Pointer(nil)
	pnr     runtime.Pinner
)
