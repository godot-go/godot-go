#include <gdnative/gdnative.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
#include <stdio.h>
#include "cgo_example.h"

example_struct cgo_example_struct(godot_gdnative_core_api_struct *api, char *name, godot_real f, godot_int i, godot_bool b) {
	example_struct ret;
	godot_string gs;
	api->godot_string_new(&gs);
	godot_bool error = api->godot_string_parse_utf8(&gs, name);

	if (error) {
		return ret;
	}

	godot_char_string gcs = api->godot_string_utf8(&gs);
	const char* cname = api->godot_char_string_get_data(&gcs);
	
	ret.name = gs;
	ret.f = f;
	ret.i = i;
	ret.b = b;

	// printf("===> name: %s, f: %.4f, i: %d, ret: %p\n", cname, f, i, &ret);

	return ret;
}

void cgo_example_struct_from_p_args(godot_gdnative_core_api_struct *api, const void **p_args, void *p_ret) {
	example_struct *example = (example_struct*)p_ret;

	char ** p_n = (char**)(p_args[0]);
	godot_real * p_f = (godot_real*)(p_args[1]);
	godot_int * p_i = (godot_int*)(p_args[2]);
	godot_bool * p_b = (godot_bool*)(p_args[3]);

	*example = cgo_example_struct(api, *p_n, *p_f, *p_i, *p_b);
}
