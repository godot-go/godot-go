#include <stdbool.h>
#include <godot/gdextension_interface.h>
#include "main_loop_callback.h"
#include "stacktrace.h"

extern void GoCallback_MainLoopStartup(void);
extern void GoCallback_MainLoopFrame(void);
extern void GoCallback_MainLoopShutdown(void);

void cgo_main_loop_startup(void) {
	printStacktrace();
	GoCallback_MainLoopStartup();
}

void cgo_main_loop_frame(void) {
	printStacktrace();
	GoCallback_MainLoopFrame();
}

void cgo_main_loop_shutdown(void) {
	printStacktrace();
	GoCallback_MainLoopShutdown();
}

GDExtensionMainLoopCallbacks cgo_main_loop_callbacks_build(bool p_startup, bool p_frame, bool p_shutdown) {
	GDExtensionMainLoopCallbacks callbacks;
	callbacks.startup_func = p_startup ? cgo_main_loop_startup : NULL;
	callbacks.frame_func = p_frame ? cgo_main_loop_frame : NULL;
	callbacks.shutdown_func = p_shutdown ? cgo_main_loop_shutdown : NULL;
	return callbacks;
}
