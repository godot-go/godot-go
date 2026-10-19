/* mingw's headers gate the SRWLOCK primitives (INIT_SRWLOCK and friends)
   behind _WIN32_WINNT >= Vista, and some toolchains default below that.
   Pin Windows 7 semantics for this translation unit before any system
   header reads the value; Godot 4.x already requires Windows 10+ at
   runtime, so this adds no practical constraint. */
#ifdef _WIN32
#undef _WIN32_WINNT
#define _WIN32_WINNT 0x0601
#undef WINVER
#define WINVER 0x0601
#endif

#include "stacktrace.h"

#ifdef _WIN32
#include <string.h>
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

bool stacktraceCaptureAvailable(void) {
#if defined(_WIN32) || defined(GODOT_GO_HAVE_EXECINFO)
	return true;
#else
	return false;
#endif
}

void printStacktrace() {
    if (enablePrintStacktrace) {
    #ifdef _WIN32
        /* DbgHelp is process-global and not thread-safe; callback fire on
           whatever engine thread runs them, so serialize every walk.
           INIT_SRWLOCK needs no init call, keeping this valid gnu90. */
        static SRWLOCK traceLock = INIT_SRWLOCK;

        AcquireSRWLockExclusive(&traceLock);
        printf("___ C start backtrace ___________\n");

#if defined(_M_X64) || defined(__x86_64__) || defined(_M_IX86) || defined(__i386__)
        {
        HANDLE process;
        CONTEXT context;
        STACKFRAME64 stackframe;
        char symbolBuffer[sizeof(SYMBOL_INFO) + MAX_SYM_NAME * sizeof(CHAR)];
        PSYMBOL_INFO symbol = (PSYMBOL_INFO)symbolBuffer;
        DWORD64 displacement;
        DWORD64 moduleBase;
        HMODULE module;
        char moduleName[MAX_PATH];
        char *basename;
        int frames;

        process = GetCurrentProcess();
        memset(&context, 0, sizeof(CONTEXT));
        context.ContextFlags = CONTEXT_FULL;
        RtlCaptureContext(&context);

        SymSetOptions(SYMOPT_UNDNAME | SYMOPT_DEFERRED_LOADS | SYMOPT_LOAD_LINES);
        SymInitialize(process, NULL, TRUE);

        ZeroMemory(&stackframe, sizeof(STACKFRAME64));
#if defined(_M_X64) || defined(__x86_64__)
        stackframe.AddrPC.Offset = context.Rip;
        stackframe.AddrPC.Mode = AddrModeFlat;
        stackframe.AddrFrame.Offset = context.Rbp;
        stackframe.AddrFrame.Mode = AddrModeFlat;
        stackframe.AddrStack.Offset = context.Rsp;
        stackframe.AddrStack.Mode = AddrModeFlat;
#elif defined(_M_IX86) || defined(__i386__)
        stackframe.AddrPC.Offset = context.Eip;
        stackframe.AddrPC.Mode = AddrModeFlat;
        stackframe.AddrFrame.Offset = context.Ebp;
        stackframe.AddrFrame.Mode = AddrModeFlat;
        stackframe.AddrStack.Offset = context.Esp;
        stackframe.AddrStack.Mode = AddrModeFlat;
#endif
        for (frames = 0; frames < 25; frames++) {
            BOOL ok = StackWalk64(
#if defined(_M_X64) || defined(__x86_64__)
                IMAGE_FILE_MACHINE_AMD64,
#else
                IMAGE_FILE_MACHINE_I386,
#endif
                process, GetCurrentThread(),
                &stackframe, &context, NULL,
                SymFunctionTableAccess64, SymGetModuleBase64, NULL);
            if (!ok) {
                break;
            }
            symbol->SizeOfStruct = sizeof(SYMBOL_INFO);
            symbol->MaxNameLen = MAX_SYM_NAME;
            displacement = 0;
            if (SymFromAddr(process, stackframe.AddrPC.Offset, &displacement, symbol)) {
                printf("[%i] %s+0x%llx\n", frames, symbol->Name,
                    (unsigned long long)displacement);
            } else {
                /* System dbghelp.dll cannot read mingw DWARF, so names often
                   resolve to nothing; module+offset+ip is the form addr2line
                   and gdb can still resolve offline. */
                module = 0;
                moduleBase = SymGetModuleBase64(process, stackframe.AddrPC.Offset);
                if (moduleBase != 0) {
                    module = (HMODULE)(DWORD_PTR)moduleBase;
                }
                if (module != 0 && GetModuleFileNameA(module, moduleName, MAX_PATH) > 0) {
                    basename = strrchr(moduleName, '\\');
                    if (basename == NULL) {
                        basename = moduleName;
                    } else {
                        basename++;
                    }
                    printf("[%i] %s+0x%llx (ip 0x%llx)\n", frames, basename,
                        (unsigned long long)(stackframe.AddrPC.Offset - moduleBase),
                        (unsigned long long)stackframe.AddrPC.Offset);
                } else {
                    printf("[%i] ??? (ip 0x%llx)\n", frames,
                        (unsigned long long)stackframe.AddrPC.Offset);
                }
            }
        }

        SymCleanup(process);
        }
#endif
        printf("___ C end backtrace   ___________\n\n");
        ReleaseSRWLockExclusive(&traceLock);
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
