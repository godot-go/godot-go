package core

// #include <godot/gdextension_interface.h>
import "C"
import (
	. "github.com/godot-go/godot-go/pkg/gdextension/builtin"
	. "github.com/godot-go/godot-go/pkg/gdextension/constant"
	. "github.com/godot-go/godot-go/pkg/gdextension/ffi"
)

func NewSimpleGDExtensionPropertyInfo(
	className string,
	variantType GDExtensionVariantType,
	name string,
) GDExtensionPropertyInfo {
	return NewGDExtensionPropertyInfo(
		NewStringNameWithLatin1Chars(className).AsGDExtensionConstStringNamePtr(),
		variantType,
		NewStringNameWithLatin1Chars(name).AsGDExtensionConstStringNamePtr(),
		uint32(PROPERTY_HINT_NONE),
		NewStringWithUtf8Chars("").AsGDExtensionConstStringPtr(),
		uint32(PROPERTY_USAGE_DEFAULT),
	)
}
