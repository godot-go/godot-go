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

func (p *GDExtensionPropertyInfo) SetUsage(usage uint32) {
	typed := (*C.GDExtensionPropertyInfo)(p)
	typed.usage = (C.uint32_t)(usage)
}

func (p *GDExtensionPropertyInfo) Name() GDExtensionStringNamePtr {
	typed := (*C.GDExtensionPropertyInfo)(p)
	return (GDExtensionStringNamePtr)(typed.name)
}

// func (p *GDExtensionPropertyInfo) Name() string {
	// stringNameConstructorWithString := CallFunc_GDExtensionInterfaceVariantGetPtrConstructor(GDEXTENSION_VARIANT_TYPE_STRING_NAME, 2)
	// stringNameConstructor := CallFunc_GDExtensionInterfaceVariantGetPtrConstructor(GDEXTENSION_VARIANT_TYPE_STRING_NAME)
	// stringNameDestructor := CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING_NAME)
	// stringDestructor := CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING)

	// cstringNameToUtf8BufferMethodName := C.CString("to_utf8_buffer")
	// defer C.free(unsafe.Pointer(cstringNameToUtf8BufferMethodName))

	// // TODO: conditional 4/8 on 32-bit/64-bit machines respectively
	// var gdStringToUtf8BufferMethodName [8]uint8
	// var gdStringToUtf8BufferMethodNamePtr = (GDExtensionStringPtr)(unsafe.Pointer(&gdStringToUtf8BufferMethodName))
	// CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars(
	// 	(GDExtensionUninitializedStringPtr)(gdStringToUtf8BufferMethodNamePtr),
	// 	cstringNameToUtf8BufferMethodName,
	// )
	// defer CallFunc_GDExtensionPtrDestructor(stringDestructor,
	// 	(GDExtensionTypePtr)(gdStringToUtf8BufferMethodNamePtr))

	// var gdStringNameToUtf8BufferMethodName [8]uint8
	// var gdStringNameToUtf8BufferMethodNamePtr = (GDExtensionStringNamePtr)(unsafe.Pointer(&gdStringNameToUtf8BufferMethodName))
	// CallBuiltinConstructor(
	// 	stringNameConstructorWithString,
	// 	(GDExtensionUninitializedTypePtr)(gdStringNameToUtf8BufferMethodNamePtr),
	// 	(GDExtensionConstTypePtr)(gdStringToUtf8BufferMethodNamePtr),
	// )
	// defer CallFunc_GDExtensionPtrDestructor(stringNameDestructor,
	// 	(GDExtensionTypePtr)(gdStringNameToUtf8BufferMethodNamePtr))

	// stringNameToUtf8BufferMethodBind := CallFunc_GDExtensionInterfaceVariantGetPtrBuiltinMethod(
	// 	GDEXTENSION_VARIANT_TYPE_STRING_NAME,
	// 	gdStringNameToUtf8BufferMethodNamePtr, 247621236)

	// // var packedByteArray [16]uint8
	// // var packedByteArrayPtr = (GDExtensionTypePtr)(unsafe.Pointer(&packedByteArray))
	// cp := (*C.GDExtensionPropertyInfo)(p)
	// utf8Buffer := CallBuiltinMethodPtrRet[[16]uint8](stringNameToUtf8BufferMethodBind, cp.name, nil)

	// // setup StringName instance of "get_string_from_utf8"
	// var gdStringToUtf8BufferMethodName [8]uint8
	// var gdStringToUtf8BufferMethodNamePtr = (GDExtensionStringPtr)(unsafe.Pointer(&gdStringToUtf8BufferMethodName))
	// CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars(
	// 	(GDExtensionUninitializedStringPtr)(gdStringToUtf8BufferMethodNamePtr),
	// 	cstringNameToUtf8BufferMethodName,
	// )
	// defer CallFunc_GDExtensionPtrDestructor(stringDestructor,
	// 	(GDExtensionTypePtr)(gdStringToUtf8BufferMethodNamePtr))

	// var gdStringNameToUtf8BufferMethodName [8]uint8
	// var gdStringNameToUtf8BufferMethodNamePtr = (GDExtensionStringNamePtr)(unsafe.Pointer(&gdStringNameToUtf8BufferMethodName))
	// CallBuiltinConstructor(
	// 	stringNameConstructorWithString,
	// 	(GDExtensionUninitializedTypePtr)(gdStringNameToUtf8BufferMethodNamePtr),
	// 	(GDExtensionConstTypePtr)(gdStringToUtf8BufferMethodNamePtr),
	// )
	// defer CallFunc_GDExtensionPtrDestructor(stringNameDestructor,
	// 	(GDExtensionTypePtr)(gdStringNameToUtf8BufferMethodNamePtr))

	// stringNameToUtf8BufferMethodBind := CallFunc_GDExtensionInterfaceVariantGetPtrBuiltinMethod(
	// 	GDEXTENSION_VARIANT_TYPE_STRING_NAME,
	// 	gdStringNameToUtf8BufferMethodNamePtr, 247621236)

	// methodName92 := NewStringNameWithLatin1Chars("to_utf8_buffer")
	// method_to_utf8_buffer := CallFunc_GDExtensionInterfaceVariantGetPtrBuiltinMethod(GDEXTENSION_VARIANT_TYPE_STRING_NAME, methodName92.AsGDExtensionConstStringNamePtr(), 247621236)
	// if method_to_utf8_buffer == nil {
	// 	log.Panic("unable to globalStringNameMethodBindings.method_to_utf8_buffer")
	// }
	// mb := globalStringNameMethodBindings.method_to_utf8_buffer

	// if mb == nil {
	// 	log.Panic("method bind cannot be nil")
	// }

	// bx := cx.nativePtr()
	// if bx == nil {
	// 	log.Panic("object cannot be nil")
	// }

	// ret := CallBuiltinMethodPtrRet[PackedByteArray](mb, bx, nil)
	// return ret

	// cp := (*C.GDExtensionPropertyInfo)(p)

	// CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars(ptr, content)
	// return "" //className
// }

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
