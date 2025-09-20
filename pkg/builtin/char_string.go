package builtin

// #include <godot/gdextension_interface.h>
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"

	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/unicode/utf32"
)

func NewStringNameWithGDExtensionConstStringNamePtr(ptr GDExtensionConstStringNamePtr) StringName {
	cx := StringName{}
	pnr.Pin(&cx)
	typedSrc := (*[StringNameSize]uint8)(ptr)
	for i := 0; i < StringNameSize; i++ {
		cx[i] = typedSrc[i]
	}
	return cx
}

func NewStringNameWithLatin1Chars(content string) StringName {
	cx := String{}
	defer cx.Destroy()
	ptr := (GDExtensionUninitializedStringPtr)(cx.NativePtr())
	pnr.Pin(ptr)
	// log.Debug("create string name",
	// 	zap.String("ptr", fmt.Sprintf("%p", ptr)),
	// 	zap.Uintptr("ptr_int", uintptr(ptr)),
	// 	zap.Any("text", content),
	// 	zap.Any("cx", cx),
	// )
	CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars(ptr, content)
	// log.Debug("create string name after",
	// 	zap.String("ptr", fmt.Sprintf("%p", ptr)),
	// 	zap.Uintptr("ptr_int", uintptr(ptr)),
	// 	zap.Any("text", content),
	// 	zap.Any("cx", cx),
	// )
	return NewStringNameWithString(cx)
}

func NewStringNameWithUtf8Chars(content string) StringName {
	cx := String{}
	defer cx.Destroy()
	ptr := (GDExtensionUninitializedStringPtr)(cx.NativePtr())
	pnr.Pin(ptr)
	log.Debug("create string name",
		zap.String("ptr", fmt.Sprintf("%p", ptr)),
		zap.Uintptr("ptr_int", uintptr(unsafe.Pointer(ptr))),
		zap.Any("text", content),
	)
	CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars(ptr, content)
	return NewStringNameWithString(cx)
}

func (cx StringName) AsString() String {
	buf := cx.ToUtf8Buffer()
	// defer buf.Destroy()
	return buf.GetStringFromUtf8()
}

func (cx StringName) ToUtf8() string {
	str := cx.AsString()
	defer str.Destroy()
	return str.ToUtf8()
}

func (cx *StringName) AsGDExtensionConstStringNamePtr() GDExtensionConstStringNamePtr {
	ptr := (GDExtensionConstStringNamePtr)(cx)
	pnr.Pin(ptr)
	return ptr
}

func GDExtensionStringPtrWithUtf8Chars(ptr GDExtensionStringPtr, content string) {
	CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars((GDExtensionUninitializedStringPtr)(ptr), content)
}

func GDExtensionStringPtrWithLatin1Chars(ptr GDExtensionStringPtr, content string) {
	CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars((GDExtensionUninitializedStringPtr)(ptr), content)
}

func NewStringWithLatin1Chars(content string) String {
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(cx.NativePtr())
	pnr.Pin(ptr)
	CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars(ptr, content)
	return cx
}

func NewStringWithUtf8Chars(content string) String {
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(cx.NativePtr())
	pnr.Pin(ptr)
	CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars(ptr, content)
	return cx
}

func NewStringWithUtf32Char(content Char32T) String {
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(cx.NativePtr())
	pnr.Pin(ptr)
	CallFunc_GDExtensionInterfaceStringNewWithUtf32Chars(ptr, &content)
	return cx
}

func (cx *String) AsGDExtensionConstStringPtr() GDExtensionConstStringPtr {
	ptr := (GDExtensionConstStringPtr)(cx)
	pnr.Pin(ptr)
	return ptr
}

func (cx String) ToAscii() string {
	ptr := (GDExtensionConstStringPtr)(cx.NativeConstPtr())
	pnr.Pin(ptr)
	size := CallFunc_GDExtensionInterfaceStringToLatin1Chars(ptr, (*Char)(nullptr), (GDExtensionInt)(0))
	cstrSlice := make([]C.char, int(size)+1)
	cstr := unsafe.SliceData(cstrSlice)
	pnr.Pin(cstr)
	CallFunc_GDExtensionInterfaceStringToLatin1Chars((GDExtensionConstStringPtr)(ptr), (*Char)(cstr), (GDExtensionInt)(size+1))
	ret := C.GoString(cstr)[:]
	return ret
}

func (cx String) ToUtf8() string {
	ptr := (GDExtensionConstStringPtr)(cx.NativeConstPtr())
	pnr.Pin(ptr)
	size := CallFunc_GDExtensionInterfaceStringToUtf8Chars(ptr, (*Char)(nullptr), (GDExtensionInt)(0))
	cstrSlice := make([]C.char, int(size)+1)
	cstr := unsafe.SliceData(cstrSlice)
	// defer func() {
	// 	stringDestructor := (GDExtensionPtrDestructor)(CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING))
	// 	if stringDestructor == nil {
	// 		log.Panic("unable to get String Destructor")
	// 	}
	// 	CallFunc_GDExtensionPtrDestructor(stringDestructor, (GDExtensionTypePtr)(cstr))
	// }()
	pnr.Pin(cstr)
	CallFunc_GDExtensionInterfaceStringToUtf8Chars(ptr, (*Char)(cstr), (GDExtensionInt)(size+1))
	ret := C.GoString(cstr)[:]
	return ret
}

var (
	utf32encoding = utf32.UTF32(utf32.LittleEndian, utf32.IgnoreBOM)
)

func (cx String) ToUtf32() string {
	ptr := (GDExtensionConstStringPtr)(cx.NativeConstPtr())
	size := CallFunc_GDExtensionInterfaceStringToUtf32Chars(ptr, (*Char32T)(nullptr), (GDExtensionInt)(0))
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
