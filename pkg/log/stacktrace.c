#include "stacktrace.h"

bool enablePrintStacktrace = false;

void printStacktrace() {
    if (enablePrintStacktrace) {
    #ifdef _WIN64
    BOOL                result;
    HANDLE              process;
    HANDLE              thread;
    CONTEXT             context;
    STACKFRAME64        stack;
    ULONG               frame;
    IMAGEHLP_SYMBOL64   symbol;
    DWORD64             displacement;
    char name[ 256 ];

    RtlCaptureContext( &context );
    memset( &stack, 0, sizeof( STACKFRAME64 ) );

    process                = GetCurrentProcess();
    thread                 = GetCurrentThread();
    displacement           = 0;
    stack.AddrPC.Offset    = context.Eip;
    stack.AddrPC.Mode      = AddrModeFlat;
    stack.AddrStack.Offset = context.Esp;
    stack.AddrStack.Mode   = AddrModeFlat;
    stack.AddrFrame.Offset = context.Ebp;
    stack.AddrFrame.Mode   = AddrModeFlat;

    for( frame = 0; ; frame++ )
    {
        result = StackWalk64
        (
            IMAGE_FILE_MACHINE_I386,
            process,
            thread,
            &stack,
            &context,
            NULL,
            SymFunctionTableAccess64,
            SymGetModuleBase64,
            NULL
        );

        symbol.SizeOfStruct  = sizeof( IMAGEHLP_SYMBOL64 );
        symbol.MaxNameLength = 255;

        SymGetSymFromAddr64( process, ( ULONG64 )stack.AddrPC.Offset, &displacement, &symbol );
        UnDecorateSymbolName( symbol.Name, ( PSTR )name, 256, UNDNAME_COMPLETE );

        printf
        (
            "Frame %lu:\n"
            "    Symbol name:    %s\n"
            "    PC address:     0x%08LX\n"
            "    Stack address:  0x%08LX\n"
            "    Frame address:  0x%08LX\n"
            "\n",
            frame,
            symbol.Name,
            ( ULONG64 )stack.AddrPC.Offset,
            ( ULONG64 )stack.AddrStack.Offset,
            ( ULONG64 )stack.AddrFrame.Offset
        );

        if( !result )
        {
            break;
        }
    }
    #else
        printf("=[ start backtrace ]=============\n\n");
        void* callstack[128];
        int i, frames = backtrace(callstack, 128);
        char** strs = backtrace_symbols(callstack, frames);
        for (i = 0; i < frames; ++i) {
            printf("%s\n", strs[i]);
        }
        free(strs);
        printf("=[ end backtrace ]===============\n\n");
    #endif
    }
}
