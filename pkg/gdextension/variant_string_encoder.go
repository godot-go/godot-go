package gdextension

// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
)

type GoStringFormat uint8

const (
	Latin1GoStringFormat GoStringFormat = iota
	Utf8GoStringFormat
)

func createGoStringEncoder(format GoStringFormat) argumentEncoder[string, String] {
	tfn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_STRING]
	vfn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING]
	var fmtfn func(GDExtensionConstStringPtr, *Char, GDExtensionInt) GDExtensionInt
	var createfn func(GDExtensionUninitializedStringPtr, string)
	switch format {
	case Latin1GoStringFormat:
		fmtfn = CallFunc_GDExtensionInterfaceStringToLatin1Chars
		createfn = CallFunc_GDExtensionInterfaceStringNewWithLatin1Chars
	case Utf8GoStringFormat:
		fmtfn = CallFunc_GDExtensionInterfaceStringToUtf8Chars
		createfn = CallFunc_GDExtensionInterfaceStringNewWithUtf8Chars
	default:
		log.Panic("unexpected go string encoder format")
	}
	decodeTypePtrArg := func(ptr GDExtensionConstTypePtr, pOut *string) {
		size := fmtfn((GDExtensionConstStringPtr)(ptr),
			(*Char)(nullptr),
			(GDExtensionInt)(0),
		)
		cstr := AllocArrayPtr[C.char](int(size) + 1)
		defer Free(unsafe.Pointer(cstr))
		fmtfn(
			(GDExtensionConstStringPtr)(ptr),
			(*Char)(unsafe.Pointer(cstr)),
			(GDExtensionInt)(size+1),
		)
		*pOut = C.GoString(cstr)[:]
	}
	decodeTypePtr := func(ptr GDExtensionConstTypePtr) string {
		var out string
		decodeTypePtrArg(ptr, &out)
		return out
	}
	encodeTypePtrArg := func(in string, out GDExtensionUninitializedTypePtr) {
		createfn((GDExtensionUninitializedStringPtr)(out), in)
	}
	encodeTypePtr := func(in string) GDExtensionTypePtr {
		var out String
		pOut := (GDExtensionTypePtr)(&out)
		encodeTypePtrArg(in, (GDExtensionUninitializedTypePtr)(pOut))
		return pOut
	}
	decodeVariantPtrArg := func(ptr GDExtensionConstVariantPtr, pOut *string) {
		var enc String
		pEnc := (GDExtensionTypePtr)(unsafe.Pointer(&enc))
		CallFunc_GDExtensionTypeFromVariantConstructorFunc(tfn, (GDExtensionUninitializedTypePtr)(pEnc), (GDExtensionVariantPtr)(ptr))
		defer enc.Destroy()
		decodeTypePtrArg((GDExtensionConstTypePtr)(pEnc), pOut)
	}
	decodeVariantPtr := func(ptr GDExtensionConstVariantPtr) string {
		var out string
		decodeVariantPtrArg(ptr, &out)
		return out
	}
	encodeVariantPtrArg := func(in string, pOut GDExtensionUninitializedVariantPtr) {
		var enc String
		pEnc := (GDExtensionTypePtr)(unsafe.Pointer(&enc))
		encodeTypePtrArg(in, (GDExtensionUninitializedTypePtr)(pEnc))
		CallFunc_GDExtensionVariantFromTypeConstructorFunc(vfn, pOut, pEnc)
		defer enc.Destroy()
	}
	encodeVariantPtr := func(in string) GDExtensionVariantPtr {
		var out Variant
		pOut := (GDExtensionVariantPtr)(unsafe.Pointer(&out))
		encodeVariantPtrArg(in, (GDExtensionUninitializedVariantPtr)(pOut))
		return pOut
	}
	decodeReflectTypePtr := func(ptr GDExtensionConstTypePtr) reflect.Value {
		content := decodeTypePtr(ptr)
		return reflect.ValueOf(content)
	}
	encodeReflectTypePtrArg := func(rv reflect.Value, pOut GDExtensionUninitializedTypePtr) {
		content := rv.String()
		encodeTypePtrArg(content, pOut)
	}
	encodeReflectTypePtr := func(rv reflect.Value) GDExtensionTypePtr {
		var out String
		pOut := (GDExtensionTypePtr)(&out)
		encodeReflectTypePtrArg(rv, (GDExtensionUninitializedTypePtr)(pOut))
		return pOut
	}
	decodeReflectVariantPtr := func(ptr GDExtensionConstVariantPtr) reflect.Value {
		var v String
		CallFunc_GDExtensionTypeFromVariantConstructorFunc(tfn, (GDExtensionUninitializedTypePtr)(v.nativePtr()), (GDExtensionVariantPtr)(ptr))
		return reflect.ValueOf(v)
	}
	encodeReflectVariantPtrArg := func(rv reflect.Value, pOut GDExtensionUninitializedVariantPtr) {
		content := rv.String()
		var enc String
		pEnc := (GDExtensionTypePtr)(unsafe.Pointer(&enc))
		encodeTypePtrArg(content, (GDExtensionUninitializedTypePtr)(pEnc))
		CallFunc_GDExtensionVariantFromTypeConstructorFunc(vfn, pOut, pEnc)
	}
	encodeReflectVariantPtr := func(rv reflect.Value) GDExtensionVariantPtr {
		var out Variant
		pOut := (GDExtensionVariantPtr)(unsafe.Pointer(&out))
		encodeReflectVariantPtrArg(rv, (GDExtensionUninitializedVariantPtr)(pOut))
		return pOut
	}
	return argumentEncoder[string, String]{
		decodeTypePtrArg:           decodeTypePtrArg,
		decodeTypePtr:              decodeTypePtr,
		encodeTypePtrArg:           encodeTypePtrArg,
		encodeTypePtr:              encodeTypePtr,
		decodeVariantPtrArg:        decodeVariantPtrArg,
		decodeVariantPtr:           decodeVariantPtr,
		encodeVariantPtrArg:        encodeVariantPtrArg,
		encodeVariantPtr:           encodeVariantPtr,
		decodeReflectTypePtr:       decodeReflectTypePtr,
		encodeReflectTypePtrArg:    encodeReflectTypePtrArg,
		encodeReflectTypePtr:       encodeReflectTypePtr,
		decodeReflectVariantPtr:    decodeReflectVariantPtr,
		encodeReflectVariantPtrArg: encodeReflectVariantPtrArg,
		encodeReflectVariantPtr:    encodeReflectVariantPtr,
	}
}
