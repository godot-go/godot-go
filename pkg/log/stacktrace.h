#ifndef CGO_GODOT_GO_STACKTRACE_H
#define CGO_GODOT_GO_STACKTRACE_H

#ifdef _M_IX86 || _M_X64
#  include <windows.h>
#  include <dbghelp.h>
#else
#  include <execinfo.h>
#endif

#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>

extern bool enablePrintStacktrace;

void printStacktrace();

#endif
