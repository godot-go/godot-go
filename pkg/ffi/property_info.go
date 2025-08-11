package ffi

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/ffi
#include <godot/gdextension_interface.h>
#include "ffi_wrapper.gen.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"

	"github.com/godot-go/godot-go/pkg/constant"
)

func NewGDExtensionPropertyInfo(
	className GDExtensionConstStringNamePtr,
	propertyType GDExtensionVariantType,
	propertyName GDExtensionConstStringNamePtr,
	hint uint32,
	hintString GDExtensionConstStringPtr,
	usage uint32,
) GDExtensionPropertyInfo {
	// TODO: create StringName locally here
	return (GDExtensionPropertyInfo)(C.GDExtensionPropertyInfo{
		_type:       (C.GDExtensionVariantType)(propertyType),
		name:        (C.GDExtensionStringNamePtr)(unsafe.Pointer(propertyName)),
		class_name:  (C.GDExtensionStringNamePtr)(unsafe.Pointer(className)),
		hint:        (C.uint32_t)(hint),
		hint_string: (C.GDExtensionStringPtr)(unsafe.Pointer(hintString)),
		usage:       (C.uint32_t)(usage),
	})
}

func (p *GDExtensionPropertyInfo) SetUsage(usage constant.PropertyUsageFlags) {
	typed := (*C.GDExtensionPropertyInfo)(p)
	typed.usage = (C.uint32_t)(usage)
}

func (p *GDExtensionPropertyInfo) Name() GDExtensionStringNamePtr {
	typed := (*C.GDExtensionPropertyInfo)(p)
	return (GDExtensionStringNamePtr)(typed.name)
}

func (p *GDExtensionPropertyInfo) Destroy() {
	cp := (*C.GDExtensionPropertyInfo)(p)
	stringNameDestructor := (GDExtensionPtrDestructor)(CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING_NAME))
	stringDestructor := (GDExtensionPtrDestructor)(CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING))
	switch {
	case cp.name != nil:
		CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(unsafe.Pointer(cp.name)))
	case cp.class_name != nil:
		CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(unsafe.Pointer(cp.class_name)))
	case cp.hint_string != nil:
		CallFunc_GDExtensionPtrDestructor(stringDestructor, (GDExtensionTypePtr)(unsafe.Pointer(cp.hint_string)))
	}
}
