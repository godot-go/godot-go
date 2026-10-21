//go:build windows

// Package libbacktrace packages the PE/COFF libbacktrace copy that Godot
// vendors, for resolving Windows extension-module frames from DWARF.
package libbacktrace

/*
#cgo CFLAGS: -I${SRCDIR}
#include "backtrace.h"

// godotgoLinkAnchor keeps a relocation to the vendored state object. The
// Go linker only pulls archive members that resolve a live reference, and
// the capture code in package log calls the C API from a different cgo
// package. This anchor makes the dependency explicit on the Go side too.
static struct backtrace_state *(*volatile godotgoCreateState)(
	const char *, int, backtrace_error_callback, void *) = backtrace_create_state;

static int godotgoLinkAnchor(void) {
	return godotgoCreateState != 0;
}
*/
import "C"

func init() {
	C.godotgoLinkAnchor()
}
