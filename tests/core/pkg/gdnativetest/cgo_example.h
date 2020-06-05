#ifndef CGO_EXAMPLE_TEST_H
#define CGO_EXAMPLE_TEST_H

#include <gdnative/gdnative.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
#include <stdio.h>

typedef struct {
	godot_string name;
	float f;
	int i;
	bool b;
} example_struct;

example_struct cgo_example_struct(godot_gdnative_core_api_struct *api, char *name, godot_real f, godot_int i, godot_bool b);

void cgo_example_struct_from_p_args(godot_gdnative_core_api_struct *api, const void **p_args, void *p_ret);

#endif
