package gdextension

// #include <godot/gdnative_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"github.com/godot-go/godot-go/pkg/gdnative"
)

func getSingleton(name string) gdnative.GDNativeObjectPtr {
	ret := gdnative.GDNativeInterface_global_get_singleton(
		internal.gdnInterface,
		name,
	)

	return (gdnative.GDNativeObjectPtr)(ret)
}

func GetInputSingleton() Input {
	return *(*Input)(getSingleton("input"))
}
