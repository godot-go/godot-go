package core

// #include <godot/gdextension_interface.h>
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/util"
)

func NewSimpleGDExtensionPropertyInfo(
	className string,
	variantType GDExtensionVariantType,
	name string,
) GDExtensionPropertyInfo {
	classNamePtr := NewStringNameWithLatin1Chars(className).AsGDExtensionConstStringNamePtr()
	namePtr := NewStringNameWithLatin1Chars(name).AsGDExtensionConstStringNamePtr()
	hintPtr := NewStringWithUtf8Chars("").AsGDExtensionConstStringPtr()
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
	util.CgoTestCall(ptr)
	return ret
}
