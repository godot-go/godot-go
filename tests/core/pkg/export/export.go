package export

import "C"
import (
	"github.com/pcting/godot-go/pkg/gdnative"
	"unsafe"
)

//export godot_gdnative_init
func godot_gdnative_init(options unsafe.Pointer) {
	gdnative.GodotGdnativeInit((*gdnative.GdnativeInitOptions)(options))
}

//export godot_gdnative_terminate
func godot_gdnative_terminate(options unsafe.Pointer) {
	gdnative.GodotGdnativeTerminate((*gdnative.GdnativeTerminateOptions)(options))
}

//export godot_nativescript_init
func godot_nativescript_init(handle unsafe.Pointer) {
	gdnative.GodotNativescriptInit(handle)
}

//export godot_nativescript_terminate
func godot_nativescript_terminate(handle unsafe.Pointer) {
	gdnative.GodotNativescriptTerminate(handle)
}
