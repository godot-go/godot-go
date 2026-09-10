#ifndef CGO_GODOT_GO_METHOD_BIND_H
#define CGO_GODOT_GO_METHOD_BIND_H

#include <godot/gdextension_interface.h>
#include <stdint.h>

// Packs a Go cgo.Handle into the void* method_userdata slot that Godot stores
// for the lifetime of the method bind. The cast lives in C so that go vet's
// unsafeptr analyzer (which cannot know a Handle is pointer-stable) stays
// silent; the value round-trips unchanged and is unpacked on the callback side
// as a cgo.Handle.
void *cgo_method_bind_userdata(uintptr_t handle);

void cgo_method_bind_method_call(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionVariantPtr *p_args, const GDExtensionInt p_argument_count, GDExtensionVariantPtr r_return, GDExtensionCallError *r_error);
void cgo_method_bind_method_ptrcall(void *method_userdata, GDExtensionClassInstancePtr p_instance, const GDExtensionTypePtr *p_args, GDExtensionTypePtr r_ret);

#endif
