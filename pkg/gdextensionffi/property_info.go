package gdextensionffi

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextensionffi
#include <godot/gdextension_interface.h>
#include "ffi_wrapper.gen.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"
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

func (p *GDExtensionPropertyInfo) Destroy() {
	cp := (*C.GDExtensionPropertyInfo)(p)
	stringNameDestructor := (GDExtensionPtrDestructor)(CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING_NAME))
	stringDestructor := (GDExtensionPtrDestructor)(CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING))
	if cp.name != nil {
		CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(unsafe.Pointer(cp.name)))
	}
	if cp.class_name != nil {
		CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(unsafe.Pointer(cp.class_name)))
	}
	if cp.hint_string != nil {
		CallFunc_GDExtensionPtrDestructor(stringDestructor, (GDExtensionTypePtr)(unsafe.Pointer(cp.hint_string)))
	}
}
