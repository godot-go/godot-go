package builtin

import (
	. "github.com/godot-go/godot-go/pkg/ffi"
)

func GetSingleton(name string) GDExtensionObjectPtr {
	snName := NewStringNameWithLatin1Chars(name)
	defer snName.Destroy()

	ret := CallFunc_GDExtensionInterfaceGlobalGetSingleton(
		snName.AsGDExtensionConstStringNamePtr(),
	)

	return (GDExtensionObjectPtr)(ret)
}
