#include "stacktrace.h"

#ifdef _WIN32
#include <windows.h>
#include <dbghelp.h>
#elif defined(GODOT_GO_ENABLE_EXECINFO)
/* User-forced path: link an external execinfo implementation
   (header + backtrace/backtrace_symbols), e.g. on Alpine/musl. See README. */
#include <execinfo.h>
#define GODOT_GO_HAVE_EXECINFO 1
#elif defined(__has_include)
#if __has_include(<execinfo.h>)
/* glibc and Apple platforms ship execinfo natively. */
#include <execinfo.h>
#define GODOT_GO_HAVE_EXECINFO 1
#endif
#elif defined(__GLIBC__) || defined(__APPLE__)
/* Fallback for compilers without __has_include (pre GCC 5 / clang 3.3). */
#include <execinfo.h>
#define GODOT_GO_HAVE_EXECINFO 1
#endif

bool enablePrintStacktrace = false;

void printStacktrace() {
    if (enablePrintStacktrace) {
    #ifdef _WIN32
        // HANDLE process = GetCurrentProcess();
        // HANDLE thread = GetCurrentThread();

        // CONTEXT context;
        // memset(&context, 0, sizeof(CONTEXT));
        // context.ContextFlags = CONTEXT_FULL;
        // RtlCaptureContext(&context);

        // SymInitialize(process, NULL, TRUE);

        // DWORD image;
        // STACKFRAME64 stackframe;
        // ZeroMemory(&stackframe, sizeof(STACKFRAME64));

        // #ifdef _M_IX86
        // image = IMAGE_FILE_MACHINE_I386;
        // stackframe.AddrPC.Offset = context.Eip;
        // stackframe.AddrPC.Mode = AddrModeFlat;
        // stackframe.AddrFrame.Offset = context.Ebp;
        // stackframe.AddrFrame.Mode = AddrModeFlat;
        // stackframe.AddrStack.Offset = context.Esp;
        // stackframe.AddrStack.Mode = AddrModeFlat;
        // #elif _M_X64
        // image = IMAGE_FILE_MACHINE_AMD64;
        // stackframe.AddrPC.Offset = context.Rip;
        // stackframe.AddrPC.Mode = AddrModeFlat;
        // stackframe.AddrFrame.Offset = context.Rsp;
        // stackframe.AddrFrame.Mode = AddrModeFlat;
        // stackframe.AddrStack.Offset = context.Rsp;
        // stackframe.AddrStack.Mode = AddrModeFlat;
        // #elif _M_IA64
        // image = IMAGE_FILE_MACHINE_IA64;
        // stackframe.AddrPC.Offset = context.StIIP;
        // stackframe.AddrPC.Mode = AddrModeFlat;
        // stackframe.AddrFrame.Offset = context.IntSp;
        // stackframe.AddrFrame.Mode = AddrModeFlat;
        // stackframe.AddrBStore.Offset = context.RsBSP;
        // stackframe.AddrBStore.Mode = AddrModeFlat;
        // stackframe.AddrStack.Offset = context.IntSp;
        // stackframe.AddrStack.Mode = AddrModeFlat;
        // #endif

        // for (size_t i = 0; i < 25; i++) {

        //     BOOL result = StackWalk64(
        //     image, process, thread,
        //     &stackframe, &context, NULL,
        //     SymFunctionTableAccess64, SymGetModuleBase64, NULL);

        //     if (!result) { break; }

        //     char buffer[sizeof(SYMBOL_INFO) + MAX_SYM_NAME * sizeof(TCHAR)];
        //     PSYMBOL_INFO symbol = (PSYMBOL_INFO)buffer;
        //     symbol->SizeOfStruct = sizeof(SYMBOL_INFO);
        //     symbol->MaxNameLen = MAX_SYM_NAME;

        //     DWORD64 displacement = 0;
        //     if (SymFromAddr(process, stackframe.AddrPC.Offset, &displacement, symbol)) {
        //     printf("[%i] %s\n", i, symbol->Name);
        //     } else {
        //     printf("[%i] ???\n", i);
        //     }

        // }

        // SymCleanup(process);
    #elif defined(GODOT_GO_HAVE_EXECINFO)
        printf("___ C start backtrace ___________\n");
        void* callstack[128];
        int i, frames = backtrace(callstack, 128);
        char** strs = backtrace_symbols(callstack, frames);
        for (i = 0; i < frames; ++i) {
            printf("%s\n", strs[i]);
        }
        free(strs);
        printf("___ C end backtrace   ___________\n\n");
    #else
        /* Platform provides no execinfo backtrace interface (e.g. musl);
           capture is compiled out. Define GODOT_GO_ENABLE_EXECINFO and link
           an implementation to restore it. */
    #endif
    }
}
