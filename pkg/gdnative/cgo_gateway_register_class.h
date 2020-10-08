#ifndef CGDNATIVE_CGO_GATEWAY_REGISTER_CLASS_H
#define CGDNATIVE_CGO_GATEWAY_REGISTER_CLASS_H

#include <gdnative_api_struct.gen.h>
#include <stdio.h>

// helper typedef for godot_instance_create_func and godot_instance_destroy_func

// owner, method_data - return user data
typedef void *(*create_func)(godot_object *, void *);

// owner, method data, user data
typedef void (*destroy_func)(godot_object *, void *, void *);

// owner, method data, user data, num args, args - return result as varaint
typedef void (*method_func)(godot_object *, void *, void *, int, godot_variant **);

// owner, method data, user data, property value
typedef void (*set_func)(godot_object *, void *, void *, godot_variant *);

// owner, method data, user data -> property value
typedef godot_variant (*get_func)(godot_object *, void *, void *);

// method data
typedef void (*free_func)(void *);

typedef void *(*alloc_instance_binding_data)(void *, const void *, godot_object *);

typedef void (*free_instance_binding_data)(void *, void *);

// cgo gateway / proxy: https://dev.to/mattn/call-go-function-from-c-function-1n3
// cgo_* functions are written in C. the cgo_* functions are assigned as callbacks
// for godot to call. These cgo_* functions will call the go_* functions.

void *cgo_gateway_create_func(godot_object *, void *);
void *go_create_func(godot_object *, void *);

void cgo_gateway_create_free_func(void *);
void go_create_free_func(void *);

void cgo_gateway_destroy_func(godot_object *, void *, void *);
void go_destroy_func(godot_object *, void *, void *);

void cgo_gateway_destroy_free_func(void *);
void go_destroy_free_func(void *);

godot_variant cgo_gateway_method_func(godot_object *, void *, void *, int, godot_variant **);
godot_variant go_method_func(godot_object *, void *, void *, int, godot_variant **);

void cgo_gateway_property_set_func(godot_object *, void *, void *, godot_variant *);
void go_property_set_func(godot_object *, void *, void *, godot_variant *);

godot_variant cgo_gateway_property_get_func(godot_object *, void *, void *);
godot_variant go_property_get_func(godot_object *, void *, void *);

char * get_class_name_from_type_tag(long);
char * get_method_name_from_method_tag(long);
char * get_property_name_from_method_tag(long);

#endif
