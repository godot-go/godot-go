package core

// #include <godot/gdextension_interface.h>
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

func NewSimpleGDExtensionPropertyInfo(
	className string,
	variantType GDExtensionVariantType,
	name string,
) GDExtensionPropertyInfo {
	classNameStringName := NewStringNameWithLatin1Chars(className)
	classNamePtr := classNameStringName.AsGDExtensionConstStringNamePtr()
	nameStringName := NewStringNameWithLatin1Chars(name)
	namePtr := nameStringName.AsGDExtensionConstStringNamePtr()
	hintString := NewStringWithUtf8Chars("")
	hintPtr := hintString.AsGDExtensionConstStringPtr()
	ret := NewGDExtensionPropertyInfo(
		classNamePtr,
		variantType,
		namePtr,
		uint32(PROPERTY_HINT_NONE),
		hintPtr,
		uint32(PROPERTY_USAGE_DEFAULT),
	)
	ptr := unsafe.Pointer(&ret)
	pnr.Pin(ptr)
	return ret
}
