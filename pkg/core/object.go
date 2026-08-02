package core

// #include <godot/gdextension_interface.h>
import "C"
import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

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
