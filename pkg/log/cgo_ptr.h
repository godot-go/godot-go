#ifndef CGO_GODOT_GO_CGO_PTR_H
#define CGO_GODOT_GO_CGO_PTR_H

#include <stdint.h>

/* Writes a packed cgo.Handle into a pointer-sized slot living in C memory
   (e.g. a GDExtensionClassMethodInfo.method_userdata or
   GDExtensionCallableCustomInfo2.callable_userdata field). The cast lives in
   C so the integer never transits a Go pointer-typed variable that a GC
   stack scan would reject (openspec fix-instance-handle-lifecycle). */
static inline void cgo_set_handle_slot(uintptr_t *p_slot, uintptr_t p_handle) {
	*p_slot = p_handle;
}

#endif
