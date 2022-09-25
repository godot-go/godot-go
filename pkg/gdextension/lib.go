package gdextension

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
#include <godot/gdnative_interface.h>
#include "classdb_wrapper.h"
#include "method_bind.h"
*/
import "C"

const (
	MAX_ARG_COUNT = 64
)
