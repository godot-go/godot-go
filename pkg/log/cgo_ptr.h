#ifndef CGO_GODOT_GO_CGO_PTR_H
#define CGO_GODOT_GO_CGO_PTR_H

#include <stdint.h>

/* Packs a Go cgo.Handle into a void* slot that Godot stores for the lifetime
   of the bound object (instance, binding, method userdata, extension class
   binding). The cast lives in C so that go vet's unsafeptr analyzer (which
   cannot know a Handle is pointer-stable) stays silent; the value round-trips
   unchanged and is unpacked on the callback side as a cgo.Handle.

   Shared by pkg/builtin, pkg/core, and pkg/ffi via the -I.../pkg/log cgo
   include path; static inline keeps it collision-free across packages. */
static inline void *cgo_handle_to_ptr(uintptr_t handle) {
	return (void *)handle;
}

#endif
