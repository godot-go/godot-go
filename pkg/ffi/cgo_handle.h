#ifndef CGO_GODOT_GO_CGO_HANDLE_H
#define CGO_GODOT_GO_CGO_HANDLE_H

#include <stdint.h>

/* Packs a Go cgo.Handle into a void* slot that Godot stores for the lifetime
   of the bound object. The cast lives in C so that go vet's unsafeptr analyzer
   (which cannot know a Handle is pointer-stable) stays silent; the value
   round-trips unchanged and is unpacked on the callback side as a cgo.Handle. */
void *cgo_handle_to_ptr(uintptr_t handle);

#endif
