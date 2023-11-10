package builtin

// #include <godot/gdextension_interface.h>
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextension/ffi"

	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/unicode/utf32"
)

func NewStringNameWithGDExtensionConstStringNamePtr(ptr GDExtensionConstStringNamePtr) StringName {
	cx := StringName{}
	typedSrc := (*[StringNameSize]uint8)(ptr)
	for i := 0; i < 8; i++ {
		cx[i] = typedSrc[i]
	}
	return cx
}

func NewStringNameWithLatin1Chars(content string) StringName {
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(unsafe.Pointer(&cx))
	log.Debug("create string name",
		zap.String("ptr", fmt.Sprintf("%p", ptr)),
		zap.Uintptr("ptr_int", uintptr(unsafe.Pointer(ptr))),
		zap.Any("text", content),
	)
	CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars(ptr, content)
	defer cx.Destroy()
	return NewStringNameWithString(cx)
}

func NewStringNameWithUtf8Chars(content string) StringName {
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(unsafe.Pointer(cx.NativePtr()))
	log.Debug("create string name",
		zap.String("ptr", fmt.Sprintf("%p", ptr)),
		zap.Uintptr("ptr_int", uintptr(unsafe.Pointer(ptr))),
		zap.Any("text", content),
	)
	CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars(ptr, content)
	defer cx.Destroy()
	return NewStringNameWithString(cx)
}

func (cx *StringName) AsString() String {
	buf := cx.ToUtf8Buffer()
	var str String
	defer str.Destroy()
	return buf.GetStringFromUtf8()
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
	ptr := (GDExtensionUninitializedStringPtr)(unsafe.Pointer(cx.NativePtr()))
	CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars(ptr, content)
	return cx
}

func NewStringWithUtf8Chars(content string) String {
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(unsafe.Pointer(cx.NativePtr()))
	CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars(ptr, content)
	return cx
}

func NewStringWithUtf32Char(content Char32T) String {
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(unsafe.Pointer(cx.NativePtr()))
	CallFunc_GDExtensionInterfaceStringNewWithUtf32Chars(ptr, &content)
	return cx
}

func (cx String) AsGDExtensionConstStringPtr() GDExtensionConstStringPtr {
	return (GDExtensionConstStringPtr)(unsafe.Pointer(&cx))
}

func (cx *String) ToAscii() string {
	size := CallFunc_GDExtensionInterfaceStringToLatin1Chars((GDExtensionConstStringPtr)(cx.NativeConstPtr()), (*Char)(nullptr), (GDExtensionInt)(0))
	cstrSlice := make([]C.char, int(size)+1)
	cstr := unsafe.SliceData(cstrSlice)
	CallFunc_GDExtensionInterfaceStringToLatin1Chars((GDExtensionConstStringPtr)(cx.NativeConstPtr()), (*Char)(cstr), (GDExtensionInt)(size+1))
	ret := C.GoString(cstr)[:]
	return ret
}

func (cx *String) ToUtf8() string {
	size := CallFunc_GDExtensionInterfaceStringToUtf8Chars((GDExtensionConstStringPtr)(cx.NativeConstPtr()), (*Char)(nullptr), (GDExtensionInt)(0))
	cstrSlice := make([]C.char, int(size)+1)
	cstr := unsafe.SliceData(cstrSlice)
	CallFunc_GDExtensionInterfaceStringToUtf8Chars((GDExtensionConstStringPtr)(cx.NativeConstPtr()), (*Char)(cstr), (GDExtensionInt)(size+1))
	ret := C.GoString(cstr)[:]
	return ret
}

var (
	utf32encoding = utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM)
)

func (cx *String) ToUtf32() string {
	size := CallFunc_GDExtensionInterfaceStringToUtf32Chars((GDExtensionConstStringPtr)(cx.NativeConstPtr()), (*Char32T)(nullptr), (GDExtensionInt)(0))
	cstrSlice := make([]Char32T, int(size)+1)
	cstr := unsafe.SliceData(cstrSlice)
	CallFunc_GDExtensionInterfaceStringToUtf32Chars((GDExtensionConstStringPtr)(cx.NativeConstPtr()), (*Char32T)(cstr), (GDExtensionInt)(size+1))
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
