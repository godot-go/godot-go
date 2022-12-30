package gdextension

// #include <godot/gdextension_interface.h>
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"

	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

func NewStringNameWithLatin1Chars(content string) StringName {
	cx := String{}
	defer cx.Destroy()

	// cx.opaque[0] = 88

	ptr := (GDExtensionStringPtr)(unsafe.Pointer(cx.ptr()))

	log.Debug("create string name",
		zap.Uintptr("gdnInterface", uintptr(unsafe.Pointer(internal.gdnInterface))),
		zap.Uintptr("ptr", uintptr(unsafe.Pointer(ptr))),
	)

	GDExtensionInterface_string_new_with_latin1_chars(internal.gdnInterface, ptr, content)

	return NewStringNameWithString(cx)
}

func NewStringNameWithUtf8Chars(content string) StringName {
	cx := String{}
	defer cx.Destroy()

	ptr := cx.ptr()

	GDExtensionInterface_string_new_with_utf8_chars(internal.gdnInterface, (GDExtensionStringPtr)(ptr), content)

	return NewStringNameWithString(cx)
}

// func NewGDExtensionStringNamePtrWithLatin1Chars(content string) GDExtensionStringNamePtr {
// 	strName := NewStringNameWithLatin1Chars(content)
// 	return (GDExtensionStringNamePtr)(unsafe.Pointer(&strName))
// }

// func NewGDExtensionStringNamePtrWithUtf8Chars(content string) GDExtensionStringNamePtr {
// 	strName := NewStringNameWithUtf8Chars(content)
// 	return (GDExtensionStringNamePtr)(unsafe.Pointer(&strName))
// }

func (cx StringName) AsGDExtensionStringNamePtr() GDExtensionConstStringNamePtr {
	return (GDExtensionConstStringNamePtr)(unsafe.Pointer(&cx))
}

func NewStringWithLatin1Chars(content string) String {
	cx := String{}

	ptr := (GDExtensionStringPtr)(unsafe.Pointer(cx.ptr()))

	GDExtensionInterface_string_new_with_latin1_chars(internal.gdnInterface, ptr, content)

	return cx
}

func NewStringWithUtf8Chars(content string) String {
	cx := String{}

	ptr := (GDExtensionStringPtr)(unsafe.Pointer(cx.ptr()))

	GDExtensionInterface_string_new_with_utf8_chars(internal.gdnInterface, ptr, content)

	return cx
}

func (cx String) AsGDExtensionStringPtr() GDExtensionConstStringPtr {
	return (GDExtensionConstStringPtr)(unsafe.Pointer(&cx))
}

func (cx *String) ToAscii() string {
	size := GDExtensionInterface_string_to_latin1_chars(internal.gdnInterface, (GDExtensionConstStringPtr)(unsafe.Pointer(cx.ptr())), (*Char)(nullptr), (GDExtensionInt)(0))

	cstr := AllocArrayPtr[C.char](int(size) + 1)
	defer Free(unsafe.Pointer(cstr))

	GDExtensionInterface_string_to_latin1_chars(internal.gdnInterface, (GDExtensionConstStringPtr)(unsafe.Pointer(cx.ptr())), (*Char)(cstr), (GDExtensionInt)(size+1))

	// *unsafe.Add(unsafe.Pointer(cstr), size + 1) = (C.char)('\000')

	ret := C.GoString(cstr)

	return ret
}

func (cx *String) ToUtf8() string {
	size := GDExtensionInterface_string_to_utf8_chars(internal.gdnInterface, (GDExtensionConstStringPtr)(cx.ptr()), (*Char)(nullptr), (GDExtensionInt)(0))

	cstr := AllocArrayPtr[C.char](int(size) + 1)
	defer Free(unsafe.Pointer(cstr))

	GDExtensionInterface_string_to_utf8_chars(internal.gdnInterface, (GDExtensionConstStringPtr)(cx.ptr()), (*Char)(cstr), (GDExtensionInt)(size+1))

	// *unsafe.Add(unsafe.Pointer(cstr), size + 1) = (C.char)('\000')

	ret := C.GoString(cstr)

	return ret
}
