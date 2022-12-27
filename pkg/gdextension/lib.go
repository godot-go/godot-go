package gdextension

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
#include <godot/gdextension_interface.h>
#include "classdb_wrapper.h"
#include "method_bind.h"
*/
import "C"
import (
	"unsafe"
)

const (
	MAX_ARG_COUNT = 64
	sizeOfPtr     = int(unsafe.Sizeof(uintptr(0)))
)
