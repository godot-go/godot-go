#include <godot/gdextension_interface.h>
#include <stdint.h>
#include "wrapped.h"
#include "stacktrace.h"

// GDClass
extern void *GoCallback_GDClassBindingCreate(void *p_token, void *p_instance);
extern void GoCallback_GDClassBindingFree(void *p_token, void *p_instance, uintptr_t p_binding);
extern GDExtensionBool GoCallback_GDClassBindingReference(void *p_token, void *p_instance, GDExtensionBool p_reference);

void *cgo_gdclass_binding_create_callback(void *p_token, void *p_instance) {
	printStacktrace();
	return GoCallback_GDClassBindingCreate(p_token, p_instance);
}

void cgo_gdclass_binding_free_callback(void *p_token, void *p_instance, void *p_binding) {
	printStacktrace();
	// p_binding carries a packed cgo.Handle (see WrappedPostInitialize), so
	// it crosses into Go as an integer: a small handle value in a pointer
	// slot is fatal under GC stack scans (openspec
	// fix-instance-handle-lifecycle).
	GoCallback_GDClassBindingFree(p_token, p_instance, (uintptr_t)p_binding);
}

GDExtensionBool cgo_gdclass_binding_reference_callback(void *p_token, void *p_instance, GDExtensionBool p_reference) {
	printStacktrace();
	return GoCallback_GDClassBindingReference(p_token, p_instance, p_reference);
}
