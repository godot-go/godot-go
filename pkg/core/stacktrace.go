package core

/*
#include "stacktrace.h"
*/
import "C"

// SetPrintStacktrace toggles the C-side callstack capture performed at
// GDExtension callback boundaries. Output is produced only on builds where
// a native capture implementation is compiled in; see
// StacktraceCaptureAvailable.
func SetPrintStacktrace(enabled bool) {
	C.enablePrintStacktrace = C.bool(enabled)
}

// StacktraceCaptureAvailable reports whether this build compiles real native
// callstack capture (Windows DbgHelp or execinfo), or whether printStacktrace
// is compiled to a no-op (no platform backtrace interface, e.g. musl).
func StacktraceCaptureAvailable() bool {
	return bool(C.stacktraceCaptureAvailable())
}
