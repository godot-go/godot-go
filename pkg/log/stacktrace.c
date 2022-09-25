#include "stacktrace.h"

bool printDebugStacktrace = false;

void printStacktrace() {
    if (printDebugStacktrace) {
        printf("===============\n\n");
        void* callstack[128];
        int i, frames = backtrace(callstack, 128);
        char** strs = backtrace_symbols(callstack, frames);
        for (i = 0; i < frames; ++i) {
            printf("%s\n", strs[i]);
        }
        free(strs);
        printf("===============\n\n");
    }
}
