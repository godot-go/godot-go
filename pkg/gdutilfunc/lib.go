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
	// pnr is for program-lifetime retention only: values the engine keeps a
	// pointer to after the call that registers them returns. Per-call scratch
	// must use a body-scoped runtime.Pinner with defer Unpin() instead.
	pnr runtime.Pinner
)
