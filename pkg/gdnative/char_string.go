package gdnative

// #include <godot/gdnative_interface.h>
// #include "gdnative_wrapper.gen.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"
)

func NewStringNameWithLatin1Chars(content string) StringName {
	cx := String{}
	defer cx.Destroy()

	ptr := cx.ptr()

	GDNativeInterface_string_new_with_latin1_chars(internal.gdnInterface, (GDNativeStringPtr)(ptr), content)

	return NewStringNameWithString(cx)
}

func NewStringWithLatin1Chars(content string) String {
	cx := String{}

	ptr := cx.ptr()

	GDNativeInterface_string_new_with_latin1_chars(internal.gdnInterface, (GDNativeStringPtr)(ptr), content)

	return cx
}

func NewStringWithUtf8Chars(content string) String {
	cx := String{}

	ptr := cx.ptr()

	GDNativeInterface_string_new_with_utf8_chars(internal.gdnInterface, (GDNativeStringPtr)(ptr), content)

	return cx
}

func (cx *String) ToAscii() string {
	size := GDNativeInterface_string_to_latin1_chars(internal.gdnInterface, (GDNativeStringPtr)(cx.ptr()), (*Char)(nullptr), (GDNativeInt)(0))

	arr := AllocArray[C.char]((uint64)(size + 1))

	cstr := (*C.char)(unsafe.Pointer(arr))

	GDNativeInterface_string_to_latin1_chars(internal.gdnInterface, (GDNativeStringPtr)(cx.ptr()), (*Char)(cstr), (GDNativeInt)(size+1))

	arr[size] = (C.char)('\000')

	ret := C.GoString(cstr)

	return ret
}
