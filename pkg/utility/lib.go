package utility

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
#include <godot/gdextension_interface.h>
*/
import "C"
import (
	"unsafe"
)

var (
	nullptr = unsafe.Pointer(nil)
)
