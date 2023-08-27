package gdextension

// #include <godot/gdextension_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
)

func getSingleton(name string) GDExtensionObjectPtr {
	snName := NewStringNameWithUtf8Chars(name)
	defer snName.Destroy()

	ret := CallFunc_GDExtensionInterfaceGlobalGetSingleton(
		snName.AsGDExtensionConstStringNamePtr(),
	)

	return (GDExtensionObjectPtr)(ret)
}

func GetInputSingleton() Input {
	owner := (*GodotObject)(unsafe.Pointer(getSingleton("Input")))
	return NewInputWithGodotOwnerObject(owner)
}
