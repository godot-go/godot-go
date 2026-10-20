package core

/*
#include "stacktrace.h"
*/
import "C"

// SetPrintStacktrace toggles the C-side callstack capture performed at
// GDExtension callback boundaries. Output is produced only on builds where
// a native capture implementation is compiled in; see
// StacktraceCaptureAvailable. On Windows the native implementation is opt-in
// (CGO_CFLAGS -DGODOT_GO_ENABLE_DBGHELP) while issue #172 is under
// investigation.
func SetPrintStacktrace(enabled bool) {
	C.enablePrintStacktrace = C.bool(enabled)
}

// StacktraceCaptureAvailable reports whether this build compiles real native
// callstack capture (Windows DbgHelp behind GODOT_GO_ENABLE_DBGHELP, or
// execinfo), or whether printStacktrace is compiled to a no-op (no platform
// backtrace interface, such as musl, or a Windows build without the DbgHelp
// opt-in).
func StacktraceCaptureAvailable() bool {
	return bool(C.stacktraceCaptureAvailable())
}
