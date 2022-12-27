package gdextension

// #include <godot/gdextension_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"

	"github.com/godot-go/godot-go/pkg/gdnative"
)

func getSingleton(name string) gdnative.GDExtensionObjectPtr {
	ret := gdnative.GDExtensionInterface_global_get_singleton(
		internal.gdnInterface,
		NewStringNameWithLatin1Chars(name).AsGDExtensionStringNamePtr(),
	)

	return (gdnative.GDExtensionObjectPtr)(ret)
}

func GetInputSingleton() Input {
	owner := (*GodotObject)(unsafe.Pointer(getSingleton("Input")))
	return NewInputWithGodotOwnerObject(owner)
}
