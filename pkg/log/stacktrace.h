#ifndef CGO_GODOT_GO_STACKTRACE_H
#define CGO_GODOT_GO_STACKTRACE_H

// Platform-specific backtrace headers (execinfo.h, windows.h, and the
// vendored libbacktrace) belong in stacktrace.c only. Including them here
// would impose a glibc dependency on every package that includes this
// header, breaking builds on libc implementations without execinfo (e.g.
// musl). See issue #127.
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>

extern bool enablePrintStacktrace;

void printStacktrace();

// stacktraceCaptureAvailable reports whether printStacktrace() captures a
// real native callstack in this build (Windows libbacktrace or execinfo), or
// is compiled to a no-op (no platform backtrace interface, e.g. musl).
bool stacktraceCaptureAvailable(void);

#endif
