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
           The static zero pattern IS an initialized SRWLOCK (SRWLOCK_INIT
           and mingw's INIT_SRWLOCK both expand to it); spelling it
           directly avoids mingw toolchains that ship no such macro. */
        static SRWLOCK traceLock = {0};

        AcquireSRWLockExclusive(&traceLock);
        printf("___ C start backtrace ___________\n");

        {
        void *callstack[25];
        unsigned short count;
        int frames;
        HANDLE process;
        char symbolBuffer[sizeof(SYMBOL_INFO) + MAX_SYM_NAME * sizeof(CHAR)];
        PSYMBOL_INFO symbol = (PSYMBOL_INFO)symbolBuffer;
        IMAGEHLP_LINE64 sourceLine;
        DWORD lineDisplacement;
        DWORD64 displacement;
        DWORD64 moduleBase;
        HMODULE module;
        char moduleName[MAX_PATH];
        char *basename;

        /* CaptureStackBackTrace (ntdll) yields a plain array of return
           addresses. Do NOT reintroduce a STACKFRAME64 local here:
           dwarf2codeview (verified on MinGW-Builds GCC 15.2.0) segfaults
           emitting that type's nested anonymous-union graph under
           -gcodeview (see openspec/changes/windows-pdb-line-info design
           D6). The return-address form also survives optimized frames
           where an RBP chain would stall the walk. */
        count = CaptureStackBackTrace(0, 25, callstack, NULL);

        process = GetCurrentProcess();
        SymSetOptions(SYMOPT_UNDNAME | SYMOPT_DEFERRED_LOADS | SYMOPT_LOAD_LINES);
        SymInitialize(process, NULL, TRUE);

        for (frames = 0; frames < (int)count; frames++) {
            symbol->SizeOfStruct = sizeof(SYMBOL_INFO);
            symbol->MaxNameLen = MAX_SYM_NAME;
            displacement = 0;
            if (SymFromAddr(process, (DWORD64)callstack[frames], &displacement, symbol)) {
                /* The build links a CodeView PDB next to the library
                   (-gcodeview + ld --pdb), so SYMOPT_LOAD_LINES can resolve
                   the exact source line; Go-runtime and engine frames have
                   no PDB and simply keep the symbol-only form. */
                memset(&sourceLine, 0, sizeof(sourceLine));
                sourceLine.SizeOfStruct = sizeof(IMAGEHLP_LINE64);
                lineDisplacement = 0;
                if (SymGetLineFromAddr64(process, (DWORD64)callstack[frames], &lineDisplacement, &sourceLine)) {
                    printf("[%i] %s+0x%llx %s:%lu\n", frames, symbol->Name,
                        (unsigned long long)displacement, sourceLine.FileName,
                        sourceLine.LineNumber);
                } else {
                    printf("[%i] %s+0x%llx\n", frames, symbol->Name,
                        (unsigned long long)displacement);
                }
            } else {
                /* System dbghelp.dll cannot read mingw DWARF, so names often
                   resolve to nothing; module+offset+ip is the form addr2line
                   and gdb can still resolve offline. */
                module = 0;
                moduleBase = SymGetModuleBase64(process, (DWORD64)callstack[frames]);
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
                        (unsigned long long)((DWORD64)callstack[frames] - moduleBase),
                        (unsigned long long)(DWORD64)callstack[frames]);
                } else {
                    printf("[%i] ??? (ip 0x%llx)\n", frames,
                        (unsigned long long)(DWORD64)callstack[frames]);
                }
            }
        }

        SymCleanup(process);
        }
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
