#ifndef CGO_GODOT_GO_WRAPPED_H
#define CGO_GODOT_GO_WRAPPED_H

#include <godot/gdextension_interface.h>
#include <stdint.h>

// Packs a Go cgo.Handle into the void* instance slot that Godot stores for the
// lifetime of the bound instance. The cast lives in C so that go vet's
// unsafeptr analyzer (which cannot know a Handle is pointer-stable) stays
// silent; the value round-trips unchanged and is unpacked on the callback side
// as a cgo.Handle.
void *cgo_wrapped_instance_ptr(uintptr_t handle);

void *cgo_gdclass_binding_create_callback(void *p_token, void *p_instance);
void cgo_gdclass_binding_free_callback(void *p_token, void *p_instance, void *p_binding);
GDExtensionBool cgo_gdclass_binding_reference_callback(void *p_token, void *p_instance, GDExtensionBool p_reference);

#endif
