#ifndef CGO_GODOT_GO_MAIN_LOOP_CALLBACK_H
#define CGO_GODOT_GO_MAIN_LOOP_CALLBACK_H

#include <godot/gdextension_interface.h>
#include <stdbool.h>

void cgo_main_loop_startup(void);
void cgo_main_loop_frame(void);
void cgo_main_loop_shutdown(void);

// Builds the engine callback struct, wiring a trampoline only for phases the
// caller requested; unrequested phases stay NULL exactly like godot-cpp's
// InitData.main_loop_callbacks struct.
GDExtensionMainLoopCallbacks cgo_main_loop_callbacks_build(bool p_startup, bool p_frame, bool p_shutdown);

#endif
