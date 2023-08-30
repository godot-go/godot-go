package gdextension

// #include <godot/gdextension_interface.h>
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"

	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/unicode/utf32"
)

func NewStringNameWithLatin1Chars(content string) StringName {
	cx := String{}
	defer cx.Destroy()
	ptr := (GDExtensionUninitializedStringPtr)(unsafe.Pointer(cx.ptr()))
	log.Debug("create string name",
		zap.String("ptr", fmt.Sprintf("%p", ptr)),
		zap.Uintptr("ptr_int", uintptr(unsafe.Pointer(ptr))),
		zap.Any("text", content),
	)
	CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars(ptr, content)
	return NewStringNameWithString(cx)
}

func NewStringNameWithUtf8Chars(content string) StringName {
	cx := String{}
	defer cx.Destroy()
	ptr := (GDExtensionUninitializedStringPtr)(unsafe.Pointer(cx.ptr()))
	log.Debug("create string name",
		zap.String("ptr", fmt.Sprintf("%p", ptr)),
		zap.Uintptr("ptr_int", uintptr(unsafe.Pointer(ptr))),
		zap.Any("text", content),
	)
	CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars(ptr, content)
	return NewStringNameWithString(cx)
}

func (cx *StringName) AsString() String {
	rt := NewStringWithUtf8Chars("")
	defer rt.Destroy()
	return cx.Add_String(rt)
}

func (cx *StringName) ToUtf8() string {
	gdStr := cx.AsString()
	return gdStr.ToUtf8()
}

func (cx StringName) AsGDExtensionConstStringNamePtr() GDExtensionConstStringNamePtr {
	return (GDExtensionConstStringNamePtr)(unsafe.Pointer(&cx))
}

func GDExtensionStringPtrWithUtf8Chars(ptr GDExtensionStringPtr, content string) {
	CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars((GDExtensionUninitializedStringPtr)(ptr), content)
}

func GDExtensionStringPtrWithLatin1Chars(ptr GDExtensionStringPtr, content string) {
	CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars((GDExtensionUninitializedStringPtr)(ptr), content)
}

func NewStringWithLatin1Chars(content string) String {
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(unsafe.Pointer(cx.ptr()))
	CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars(ptr, content)
	return cx
}

func NewStringWithUtf8Chars(content string) String {
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(unsafe.Pointer(cx.ptr()))
	CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars(ptr, content)
	return cx
}

func NewStringWithUtf32Char(content Char32T) String {
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(unsafe.Pointer(cx.ptr()))
	CallFunc_GDExtensionInterfaceStringNewWithUtf32Chars(ptr, &content)
	return cx
}

func (cx String) AsGDExtensionConstStringPtr() GDExtensionConstStringPtr {
	return (GDExtensionConstStringPtr)(unsafe.Pointer(&cx))
}

func (cx *String) ToAscii() string {
	size := CallFunc_GDExtensionInterfaceStringToLatin1Chars((GDExtensionConstStringPtr)(unsafe.Pointer(cx.ptr())), (*Char)(nullptr), (GDExtensionInt)(0))
	cstr := AllocArrayPtr[C.char](int(size) + 1)
	defer Free(unsafe.Pointer(cstr))
	CallFunc_GDExtensionInterfaceStringToLatin1Chars((GDExtensionConstStringPtr)(unsafe.Pointer(cx.ptr())), (*Char)(cstr), (GDExtensionInt)(size+1))
	ret := C.GoString(cstr)[:]
	return ret
}

func (cx *String) ToUtf8() string {
	size := CallFunc_GDExtensionInterfaceStringToUtf8Chars((GDExtensionConstStringPtr)(cx.ptr()), (*Char)(nullptr), (GDExtensionInt)(0))
	cstr := AllocArrayPtr[C.char](int(size) + 1)
	defer Free(unsafe.Pointer(cstr))
	CallFunc_GDExtensionInterfaceStringToUtf8Chars((GDExtensionConstStringPtr)(cx.ptr()), (*Char)(cstr), (GDExtensionInt)(size+1))
	ret := C.GoString(cstr)[:]
	return ret
}

var (
	utf32encoding = utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM)
)

func (cx *String) ToUtf32() string {
	size := CallFunc_GDExtensionInterfaceStringToUtf32Chars((GDExtensionConstStringPtr)(cx.ptr()), (*Char32T)(nullptr), (GDExtensionInt)(0))
	cstr := AllocArrayPtr[Char32T](int(size) + 1)
	defer Free(unsafe.Pointer(cstr))
	CallFunc_GDExtensionInterfaceStringToUtf32Chars((GDExtensionConstStringPtr)(cx.ptr()), (*Char32T)(cstr), (GDExtensionInt)(size+1))
	dec := utf32encoding.NewDecoder()
	bytesPtr := (*byte)(unsafe.Pointer(cstr))
	b := unsafe.Slice(bytesPtr, 4*(int(size)+1))
	bRet, err := dec.Bytes(b)
	if err != nil {
		log.Panic("unable to convert to utf32")
	}
	ret := string(bRet)
	log.Info("decoded utf32",
		zap.String("str", ret),
	)
	return ret
}
