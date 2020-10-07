#ifndef CGDNATIVE_CGO_GATEWAY_INSTANCE_BINDING_H
#define CGDNATIVE_CGO_GATEWAY_INSTANCE_BINDING_H

#include <gdnative_api_struct.gen.h>

// cgo gateway / proxy: https://dev.to/mattn/call-go-function-from-c-function-1n3
// cgo_* functions are written in C. the cgo_* functions are assigned as callbacks
// for godot to call. These cgo_* functions will call the go_* functions.

void *cgo_gateway_alloc_instance_binding_data(void *, const void *, godot_object *);
void *go_alloc_instance_binding_data(void *, void *, godot_object *);

void cgo_gateway_free_instance_binding_data(void *, void *);
void go_free_instance_binding_data(void *, void *);
#endif
