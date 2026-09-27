package builtin

// #include <godot/gdextension_interface.h>
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"runtime"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"

	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/unicode/utf32"
)

func StringNameCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalStringNameMethodBindings.constructor_1, out, src)
}

func NodePathCopyConstructor(out GDExtensionUninitializedTypePtr, src GDExtensionConstTypePtr) {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	pinner.Pin(unsafe.Pointer(src))
	CallBuiltinConstructor(globalNodePathMethodBindings.constructor_1, out, src)
}

func NewStringNameWithLatin1Chars(content string) StringName {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	cx := String{}
	defer cx.Destroy()
	ptr := (GDExtensionUninitializedStringPtr)(cx.NativePtr())
	pinner.Pin(ptr)
	CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars(ptr, content)
	return NewStringNameWithString(cx)
}

func NewStringNameWithUtf8Chars(content string) StringName {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	cx := String{}
	defer cx.Destroy()
	ptr := (GDExtensionUninitializedStringPtr)(cx.NativePtr())
	pinner.Pin(ptr)
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

// AsGDExtensionConstStringNamePtr returns the receiver's address for use as a C
// call argument and pins it on the package pinner.
//
// The pin is required, not defensive: registration code stores this pointer inside
// Go structs (GDExtensionPropertyInfo and friends) that are then handed to C, and
// cgocheck rejects a cgo argument whose pointee contains an unpinned Go pointer.
// Callers that pass the pointer *directly* to a synchronous call should use
// AsGDExtensionConstStringNamePtrPinned so the pin ends with that call instead of
// lasting for the program.
func (cx *StringName) AsGDExtensionConstStringNamePtr() GDExtensionConstStringNamePtr {
	ptr := (GDExtensionConstStringNamePtr)(cx)
	pnr.Pin(ptr)
	return ptr
}

// AsGDExtensionConstStringNamePtrPinned is the call-scoped counterpart of
// AsGDExtensionConstStringNamePtr: the pin is taken on the caller's pinner, so it
// is released when the caller's body returns. Use only where the pointer is a
// direct argument to a synchronous engine call that copies it, never where it is
// stored in memory the engine keeps.
func (cx *StringName) AsGDExtensionConstStringNamePtrPinned(pinner *runtime.Pinner) GDExtensionConstStringNamePtr {
	pinner.Pin(cx)
	return (GDExtensionConstStringNamePtr)(cx)
}

func GDExtensionStringPtrWithUtf8Chars(ptr GDExtensionStringPtr, content string) {
	CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars((GDExtensionUninitializedStringPtr)(ptr), content)
}

func GDExtensionStringPtrWithLatin1Chars(ptr GDExtensionStringPtr, content string) {
	CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars((GDExtensionUninitializedStringPtr)(ptr), content)
}

func NewStringWithLatin1Chars(content string) String {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(cx.NativePtr())
	pinner.Pin(ptr)
	CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars(ptr, content)
	return cx
}

func NewStringWithUtf8Chars(content string) String {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(cx.NativePtr())
	pinner.Pin(ptr)
	CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars(ptr, content)
	return cx
}

func NewStringWithUtf32Char(content Char32T) String {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	cx := String{}
	ptr := (GDExtensionUninitializedStringPtr)(cx.NativePtr())
	pinner.Pin(ptr)
	CallFunc_GDExtensionInterfaceStringNewWithUtf32Chars(ptr, &content)
	return cx
}

// AsGDExtensionConstStringPtr is the String counterpart of
// AsGDExtensionConstStringNamePtr and carries the same required package pin.
func (cx *String) AsGDExtensionConstStringPtr() GDExtensionConstStringPtr {
	ptr := (GDExtensionConstStringPtr)(cx)
	pnr.Pin(ptr)
	return ptr
}

// AsGDExtensionConstStringPtrPinned is the call-scoped counterpart of
// AsGDExtensionConstStringPtr.
func (cx *String) AsGDExtensionConstStringPtrPinned(pinner *runtime.Pinner) GDExtensionConstStringPtr {
	pinner.Pin(cx)
	return (GDExtensionConstStringPtr)(cx)
}

func (cx String) ToAscii() string {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	ptr := (GDExtensionConstStringPtr)(cx.NativeConstPtr())
	pinner.Pin(ptr)
	size := CallFunc_GDExtensionInterfaceStringToLatin1Chars(ptr, (*Char)(nullptr), (GDExtensionInt)(0))
	cstrSlice := make([]C.char, int(size)+1)
	cstr := unsafe.SliceData(cstrSlice)
	pinner.Pin(cstr)
	CallFunc_GDExtensionInterfaceStringToLatin1Chars((GDExtensionConstStringPtr)(ptr), (*Char)(cstr), (GDExtensionInt)(size+1))
	ret := C.GoString(cstr)[:]
	return ret
}

func (cx String) ToUtf8() string {
	var pinner runtime.Pinner
	defer pinner.Unpin()
	ptr := (GDExtensionConstStringPtr)(cx.NativeConstPtr())
	pinner.Pin(ptr)
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
	pinner.Pin(cstr)
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
