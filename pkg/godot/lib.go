package godot

/*
#cgo linux CFLAGS: -g -I${SRCDIR}/../../godot_headers
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
// #include "cgo_helpers.h"
*/
import "C"

var gdnative *C.godot_gdnative_core_api_struct

//export godot_gdnative_init
func godot_gdnative_init(options *C.godot_gdnative_init_options) {
	gdnative = (*options).api_struct
}

//export godot_gdnative_terminate
func godot_gdnative_terminate(options *C.godot_gdnative_terminate_options) {
	gdnative = nil
}
