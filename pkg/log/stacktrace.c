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
#include <stdint.h>
#include <string.h>
#include <windows.h>

/* PE/COFF libbacktrace bundled by Godot 4.7-stable. On Windows it resolves
   frames in the extension DLL from the DWARF embedded by mingw, so capture
   no longer depends on dbghelp.dll, CodeView, or a PDB. */
#include "libbacktrace/backtrace.h"
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

#ifdef _WIN32
/* libbacktrace has no free API and caches DWARF data in the state, so the
   state lives for the process lifetime. That is safe here because the Go
   extension module is pinned in the process on Windows (issue #172): it is
   never unloaded, and the static filename buffer stays valid. */
static SRWLOCK traceLock = {0};
static struct backtrace_state *backtraceState = NULL;
static uintptr_t backtraceRuntimeBase = 0;
static uintptr_t backtraceImageBase = 0;
static size_t backtraceImageSize = 0;
static int backtraceStateInitialized = 0;
static char backtraceModulePath[MAX_PATH];

typedef struct {
	int frame;
	uintptr_t pc;
	int printed;
} resolve_context;

static void backtraceError(void *data, const char *msg, int errnum) {
	(void)data;
	(void)msg;
	(void)errnum;
}

/* Windows rewrites OptionalHeader.ImageBase in the mapped image to the
   actual load address. The DWARF addresses in the file use the preferred
   base, so read that value back from the DLL on disk. */
static uintptr_t getFileImageBase(const char *path) {
	FILE *file;
	IMAGE_DOS_HEADER dos;
	IMAGE_NT_HEADERS nt;
	uintptr_t imageBase = 0;

	file = fopen(path, "rb");
	if (file == NULL) {
		return 0;
	}
	if (fread(&dos, sizeof(dos), 1, file) == 1 &&
			dos.e_magic == IMAGE_DOS_SIGNATURE && dos.e_lfanew >= 0 &&
			fseek(file, (long)dos.e_lfanew, SEEK_SET) == 0 &&
			fread(&nt, sizeof(nt), 1, file) == 1 &&
			nt.Signature == IMAGE_NT_SIGNATURE) {
		imageBase = (uintptr_t)nt.OptionalHeader.ImageBase;
	}
	fclose(file);
	return imageBase;
}

static void getModuleImageInfo(HMODULE module, const char *path, uintptr_t *imageBase, size_t *imageSize) {
	const IMAGE_DOS_HEADER *dos;
	const IMAGE_NT_HEADERS *nt;

	*imageBase = 0;
	*imageSize = 0;
	if (module == NULL) {
		return;
	}
	dos = (const IMAGE_DOS_HEADER *)(const void *)module;
	if (dos->e_magic != IMAGE_DOS_SIGNATURE || dos->e_lfanew < 0) {
		return;
	}
	nt = (const IMAGE_NT_HEADERS *)((const BYTE *)module + dos->e_lfanew);
	if (nt->Signature != IMAGE_NT_SIGNATURE) {
		return;
	}
	*imageSize = (size_t)nt->OptionalHeader.SizeOfImage;
	*imageBase = getFileImageBase(path);
}

static void ensureBacktraceState(void) {
	HMODULE module = NULL;
	DWORD pathLength;

	if (backtraceStateInitialized) {
		return;
	}
	backtraceStateInitialized = 1;

	/* Use an address in this module's data, not a function pointer: the
	   extension DLL owns both it and the capture code calling this. */
	if (!GetModuleHandleExA(GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS | GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT,
			(LPCSTR)(const void *)&enablePrintStacktrace, &module) ||
			module == NULL) {
		return;
	}
	pathLength = GetModuleFileNameA(module, backtraceModulePath, (DWORD)sizeof(backtraceModulePath));
	if (pathLength == 0 || pathLength >= (DWORD)sizeof(backtraceModulePath)) {
		return;
	}
	getModuleImageInfo(module, backtraceModulePath, &backtraceImageBase, &backtraceImageSize);
	if (backtraceImageBase == 0 || backtraceImageSize == 0) {
		return;
	}

	/* threaded=1: callbacks can fire on more than one engine thread. The
	   capture lock below serializes the printed window, but the state
	   itself must still be safe if that policy changes later. */
	backtraceState = backtrace_create_state(backtraceModulePath, 1, backtraceError, NULL);
	if (backtraceState == NULL) {
		return;
	}
	backtraceRuntimeBase = (uintptr_t)module;
}

static int resolveFrame(void *data, uintptr_t pc, const char *filename, int lineno, const char *function) {
	resolve_context *ctx = (resolve_context *)data;
	(void)pc;

	/* Only treat a frame as resolved when both pieces the spec promises
	   are present. A function name without a line still falls back to the
	   module+offset form, which addr2line and gdb can resolve offline. */
	if (filename != NULL && lineno > 0 && function != NULL) {
		printf("[%i] %s (%s:%d) (ip 0x%llx)\n", ctx->frame, function, filename, lineno,
			(unsigned long long)ctx->pc);
		ctx->printed = 1;
	}
	return 0;
}

static void printFrameFallback(int frame, uintptr_t pc) {
	HMODULE module = NULL;
	char moduleName[MAX_PATH];
	char *basename;
	DWORD length;

	if (GetModuleHandleExA(GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS | GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT,
			(LPCSTR)(uintptr_t)pc, &module) &&
			module != NULL) {
		length = GetModuleFileNameA(module, moduleName, (DWORD)sizeof(moduleName));
		if (length > 0 && length < (DWORD)sizeof(moduleName)) {
			basename = strrchr(moduleName, '\\');
			if (basename == NULL) {
				basename = strrchr(moduleName, '/');
			}
			if (basename == NULL) {
				basename = moduleName;
			} else {
				basename++;
			}
			printf("[%i] %s+0x%llx (ip 0x%llx)\n", frame, basename,
				(unsigned long long)(pc - (uintptr_t)module),
				(unsigned long long)pc);
			return;
		}
	}
	printf("[%i] ??? (ip 0x%llx)\n", frame, (unsigned long long)pc);
}
#endif

bool stacktraceCaptureAvailable(void) {
#if defined(_WIN32)
	return true;
#elif defined(GODOT_GO_HAVE_EXECINFO)
	return true;
#else
	return false;
#endif
}

void printStacktrace() {
    if (enablePrintStacktrace) {
    #if defined(_WIN32)
        /* The print window is serialized so concurrent callbacks cannot
           interleave banners and frames. The libbacktrace state is also
           created under this lock. */
        void *callstack[25];
        unsigned short count;
        int frame;
        intptr_t delta;

        AcquireSRWLockExclusive(&traceLock);
        printf("___ C start backtrace ___________\n");

        ensureBacktraceState();
        delta = (intptr_t)(backtraceRuntimeBase - backtraceImageBase);
        count = CaptureStackBackTrace(0, 25, callstack, NULL);

        for (frame = 0; frame < (int)count; frame++) {
            uintptr_t pc = (uintptr_t)callstack[frame];
            uintptr_t lookupPc;
            resolve_context ctx;

            if (backtraceState == NULL ||
                    pc < backtraceRuntimeBase ||
                    pc - backtraceRuntimeBase >= backtraceImageSize) {
                printFrameFallback(frame, pc);
                continue;
            }

            /* CaptureStackBackTrace returns return addresses. Step back one
               byte for symbolization so the line belongs to the call, not
               the instruction after it. The DWARF addresses are preferred
               image VAs, so subtract the ASLR delta first. */
            lookupPc = (uintptr_t)((intptr_t)pc - delta);
            if (lookupPc > backtraceImageBase) {
                lookupPc -= 1;
            }

            ctx.frame = frame;
            ctx.pc = pc;
            ctx.printed = 0;
            backtrace_pcinfo(backtraceState, lookupPc, resolveFrame, backtraceError, &ctx);
            if (!ctx.printed) {
                printFrameFallback(frame, pc);
            }
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
        /* Compiled out: a libc with no execinfo interface (e.g. musl). See
           the include block at the top for how to opt into an external
           implementation. */
    #endif
    }
}
