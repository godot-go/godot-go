package builtin

// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
)

func createBoolEncoder[T ~bool, E ~uint8]() argumentEncoder[T, E] {
	vfn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_BOOL]
	tfn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_BOOL]
	decodeTypePtrArg := func(v GDExtensionConstTypePtr, pOut *T) {
		*pOut = (*(*E)(v)) > 0
	}
	decodeTypePtr := func(v GDExtensionConstTypePtr) T {
		var out T
		decodeTypePtrArg(v, &out)
		return out
	}
	encodeTypePtrArg := func(in T, pOut GDExtensionUninitializedTypePtr) {
		pEnc := (*E)(pOut)
		if in {
			*pEnc = 1
		} else {
			*pEnc = 0
		}
	}
	encodeTypePtr := func(in T) GDExtensionTypePtr {
		var out E
		pOut := (GDExtensionTypePtr)(unsafe.Pointer(&out))
		encodeTypePtrArg(in, (GDExtensionUninitializedTypePtr)(pOut))
		return pOut
	}
	decodeVariantPtrArg := func(pVariant GDExtensionConstVariantPtr, pOut *T) {
		var enc E
		pEnc := (GDExtensionTypePtr)(unsafe.Pointer(&enc))
		CallFunc_GDExtensionTypeFromVariantConstructorFunc(
			(GDExtensionTypeFromVariantConstructorFunc)(tfn),
			(GDExtensionUninitializedTypePtr)(pEnc),
			(GDExtensionVariantPtr)(pVariant),
		)
		decodeTypePtrArg((GDExtensionConstTypePtr)(pEnc), pOut)
	}
	decodeVariantPtr := func(pVariant GDExtensionConstVariantPtr) T {
		var out T
		decodeVariantPtrArg(pVariant, &out)
		return out
	}
	encodeVariantPtrArg := func(in T, rOut GDExtensionUninitializedVariantPtr) {
		var enc E
		pEnc := (GDExtensionTypePtr)(unsafe.Pointer(&enc))
		encodeTypePtrArg(in, (GDExtensionUninitializedTypePtr)(pEnc))
		CallFunc_GDExtensionVariantFromTypeConstructorFunc(
			(GDExtensionVariantFromTypeConstructorFunc)(vfn),
			rOut,
			pEnc,
		)
	}
	encodeVariantPtr := func(in T) GDExtensionVariantPtr {
		var out Variant
		pOut := (GDExtensionVariantPtr)(unsafe.Pointer(&out))
		encodeVariantPtrArg(in, (GDExtensionUninitializedVariantPtr)(pOut))
		return pOut
	}
	decodeReflectTypePtr := func(ptr GDExtensionConstTypePtr) reflect.Value {
		v := decodeTypePtr(ptr)
		return reflect.ValueOf(v)
	}
	encodeReflectTypePtrArg := func(rv reflect.Value, rOut GDExtensionUninitializedTypePtr) {
		v := (T)(rv.Bool())
		encodeTypePtrArg(v, rOut)
	}
	encodeReflectTypePtr := func(rv reflect.Value) GDExtensionTypePtr {
		var out E
		pOut := (GDExtensionTypePtr)(unsafe.Pointer(&out))
		encodeReflectTypePtrArg(rv, (GDExtensionUninitializedTypePtr)(pOut))
		return pOut
	}
	decodeReflectVariantPtr := func(ptr GDExtensionConstVariantPtr) reflect.Value {
		v := decodeVariantPtr(ptr)
		return reflect.ValueOf(v)
	}
	encodeReflectVariantPtrArg := func(rv reflect.Value, rOut GDExtensionUninitializedVariantPtr) {
		v := rv.Interface().(T)
		encodeVariantPtrArg(v, rOut)
	}
	encodeReflectVariantPtr := func(rv reflect.Value) GDExtensionVariantPtr {
		var out Variant
		pOut := (GDExtensionVariantPtr)(unsafe.Pointer(&out))
		encodeReflectVariantPtrArg(rv, (GDExtensionUninitializedVariantPtr)(pOut))
		return pOut
	}
	return argumentEncoder[T, E]{
		DecodeTypePtrArg:           decodeTypePtrArg,
		DecodeTypePtr:              decodeTypePtr,
		EncodeTypePtr:              encodeTypePtr,
		EncodeTypePtrArg:           encodeTypePtrArg,
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
