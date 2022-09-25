#ifndef CGO_GODOT_GO_STACKTRACE_H
#define CGO_GODOT_GO_STACKTRACE_H

#include <execinfo.h>
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>

extern bool printDebugStacktrace;

void printStacktrace();

#endif
