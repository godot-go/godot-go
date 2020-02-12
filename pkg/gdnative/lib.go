package gdnative

/*
#cgo CFLAGS: -DX86=1 -g -fPIC -std=c99 -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/gdnative
#include <godot/gdnative_interface.h>
#include "gdnative_wrapper.gen.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"os/signal"
	"syscall"
)

// GodotGoVersion holds the relese version
var GodotGoVersion = "0.1"

// This NOOP init() seems to get this file evaluated first
// so that the #cgo directive get evaluated first
func init() {
	// suggestion to help with C stacktraces: https://stackoverflow.com/a/35234443/320858
	// "This allows me to see the stack trace of the C code that crashed."
	signal.Ignore(syscall.SIGABRT)
}
