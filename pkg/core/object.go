package core

// #include <godot/gdextension_interface.h>
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// NewSimpleGDExtensionPropertyInfo is DEPRECATED — use NewGDExtensionPropertyInfoFromNames instead.
// This function creates StringName objects internally and returns pointers to them,
// which become dangling when the function returns. The pointers happen to work due to
// escape analysis keeping StringNames on the heap, but the pattern is fragile.
// See Task 5.1 in fix-orphan-stringname change.
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

// NewGDExtensionPropertyInfoFromNames constructs a GDExtensionPropertyInfo
// using pre-created StringName and String objects. The caller is responsible
// for managing the lifecycle of the StringName/String objects (they must persist
// for the duration of the GDExtensionPropertyInfo's use).
// This eliminates the dangling pointer issue in NewSimpleGDExtensionPropertyInfo.
func NewGDExtensionPropertyInfoFromNames(
	className *StringName,
	variantType GDExtensionVariantType,
	name *StringName,
	hintString *String,
) GDExtensionPropertyInfo {
	classNamePtr := className.AsGDExtensionConstStringNamePtr()
	namePtr := name.AsGDExtensionConstStringNamePtr()
	hintPtr := hintString.AsGDExtensionConstStringPtr()
	return NewGDExtensionPropertyInfo(
		classNamePtr,
		variantType,
		namePtr,
		uint32(PROPERTY_HINT_NONE),
		hintPtr,
		uint32(PROPERTY_USAGE_DEFAULT),
	)
}
