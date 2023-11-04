package builtin

// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
)

func CreateObjectEncoder[T Object]() objectArgumentEncoder[T] {
	tfn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	if tfn == nil {
		log.Panic("could not find type from variant constructor GDEXTENSION_VARIANT_TYPE_OBJECT")
	}
	vfn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	if vfn == nil {
		log.Panic("could not find variant from type constructor GDEXTENSION_VARIANT_TYPE_OBJECT")
	}
	decodeTypePtrArg := func(ptr GDExtensionConstTypePtr, pOut T) {
		dst := (**GodotObject)(unsafe.Pointer(pOut.AsGDExtensionTypePtr()))
		src := (**GodotObject)(unsafe.Pointer(ptr))
		if dst == src {
			// noop
			return
		}
		*dst = *src
	}
	decodeTypePtr := func(ptr GDExtensionConstTypePtr) T {
		var out T
		decodeTypePtrArg(ptr, out)
		return out
	}
	encodeTypePtrArg := func(in T, pOut GDExtensionUninitializedTypePtr) {
		pEnc := in.AsGDExtensionTypePtr()
		*(**GodotObject)(unsafe.Pointer(pOut)) = *(**GodotObject)(unsafe.Pointer(pEnc))
	}
	encodeTypePtr := func(in T) GDExtensionTypePtr {
		var out T
		pOut := (GDExtensionTypePtr)(unsafe.Pointer(&out))
		encodeTypePtrArg(in, (GDExtensionUninitializedTypePtr)(pOut))
		return pOut
	}
	decodeVariantPtrArg := func(ptr GDExtensionConstVariantPtr, pOut T) {
		pEnc := pOut.AsGDExtensionTypePtr()
		CallFunc_GDExtensionTypeFromVariantConstructorFunc(
			tfn,
			(GDExtensionUninitializedTypePtr)(pEnc),
			(GDExtensionVariantPtr)(ptr),
		)
	}
	decodeVariantPtr := func(ptr GDExtensionConstVariantPtr) T {
		var out T
		decodeVariantPtrArg(ptr, out)
		return out
	}
	encodeVariantPtrArg := func(in T, pOut GDExtensionUninitializedVariantPtr) {
		enc := make([]uint8, ObjectSize)
		pEnc := (GDExtensionTypePtr)(unsafe.SliceData(enc))
		encodeTypePtrArg(in, (GDExtensionUninitializedTypePtr)(pEnc))
		CallFunc_GDExtensionVariantFromTypeConstructorFunc(vfn, pOut, pEnc)
	}
	encodeVariantPtr := func(in T) GDExtensionVariantPtr {
		var out Variant
		pOut := (GDExtensionVariantPtr)(unsafe.Pointer(&out))
		encodeVariantPtrArg(
			in,
			(GDExtensionUninitializedVariantPtr)(pOut),
		)
		return pOut
	}
	decodeReflectTypePtr := func(ptr GDExtensionConstTypePtr) reflect.Value {
		v := decodeTypePtr(ptr)
		return reflect.ValueOf(v)
	}
	encodeReflectTypePtrArg := func(rv reflect.Value, pOut GDExtensionUninitializedTypePtr) {
		v := rv.Interface().(T)
		if (Object)(v) == nil {
			return
		}
		encodeTypePtrArg(v, pOut)
	}
	encodeReflectTypePtr := func(rv reflect.Value) GDExtensionTypePtr {
		var out T
		pOut := (GDExtensionTypePtr)(unsafe.Pointer(&out))
		encodeReflectTypePtrArg(rv, (GDExtensionUninitializedTypePtr)(pOut))
		return pOut
	}
	decodeReflectVariantPtr := func(ptr GDExtensionConstVariantPtr) reflect.Value {
		v := decodeVariantPtr(ptr)
		return reflect.ValueOf(v)
	}
	encodeReflectVariantPtrArg := func(rv reflect.Value, pOut GDExtensionUninitializedVariantPtr) {
		v := rv.Interface().(T)
		if (Object)(v) == nil {
			CallFunc_GDExtensionInterfaceVariantNewNil(pOut)
		}
		encodeVariantPtrArg(v, pOut)
	}
	encodeReflectVariantPtr := func(rv reflect.Value) GDExtensionVariantPtr {
		var out Variant
		pOut := (GDExtensionVariantPtr)(unsafe.Pointer(&out))
		encodeReflectVariantPtrArg(rv, (GDExtensionUninitializedVariantPtr)(pOut))
		return pOut
	}
	return objectArgumentEncoder[T]{
		DecodeTypePtrArg:           decodeTypePtrArg,
		DecodeTypePtr:              decodeTypePtr,
		EncodeTypePtrArg:           encodeTypePtrArg,
		EncodeTypePtr:              encodeTypePtr,
		DecodeVariantPtrArg:        decodeVariantPtrArg,
		DecodeVariantPtr:           decodeVariantPtr,
		EncodeVariantPtrArg:        encodeVariantPtrArg,
		EncodeVariantPtr:           encodeVariantPtr,
		decodeReflectTypePtr:       decodeReflectTypePtr,
		encodeReflectTypePtrArg:    encodeReflectTypePtrArg,
		encodeReflectTypePtr:       encodeReflectTypePtr,
		decodeReflectVariantPtr:    decodeReflectVariantPtr,
		encodeReflectVariantPtrArg: encodeReflectVariantPtrArg,
		encodeReflectVariantPtr:    encodeReflectVariantPtr,
	}
}
