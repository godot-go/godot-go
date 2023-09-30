package gdextension

// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"golang.org/x/exp/constraints"
)

type Number interface {
	constraints.Integer | constraints.Float
}

func createNumberEncoder[T Number, E Number](t GDExtensionVariantType) argumentEncoder[T, E] {
	if t != GDEXTENSION_VARIANT_TYPE_INT && t != GDEXTENSION_VARIANT_TYPE_FLOAT {
		log.Panic("createNumberEncoder supports only INT and FLOAT")
	}
	tfn := typeFromVariantConstructor[t]
	vfn := variantFromTypeConstructor[t]
	decodeTypePtrArg := func(v GDExtensionConstTypePtr, pOut *T) {
		*pOut = (T)(*(*E)(v))
	}
	decodeTypePtr := func(v GDExtensionConstTypePtr) T {
		var out T
		decodeTypePtrArg(v, &out)
		return out
	}
	encodeTypePtrArg := func(in T, out GDExtensionUninitializedTypePtr) {
		*(*E)(out) = (E)(in)
	}
	encodeTypePtr := func(in T) GDExtensionTypePtr {
		var out E
		pOut := (GDExtensionTypePtr)(unsafe.Pointer(&out))
		encodeTypePtrArg(in, (GDExtensionUninitializedTypePtr)(pOut))
		return pOut
	}
	decodeVariantPtrArg := func(ptr GDExtensionConstVariantPtr, pOut *T) {
		var enc E
		pEnc := (GDExtensionTypePtr)(unsafe.Pointer(&enc))
		CallFunc_GDExtensionTypeFromVariantConstructorFunc(
			tfn,
			(GDExtensionUninitializedTypePtr)(pEnc),
			(GDExtensionVariantPtr)(ptr),
		)
		decodeTypePtrArg((GDExtensionConstTypePtr)(pEnc), pOut)
	}
	decodeVariantPtr := func(ptr GDExtensionConstVariantPtr) T {
		var out T
		decodeVariantPtrArg(ptr, &out)
		return out
	}
	encodeVariantPtrArg := func(in T, rOut GDExtensionUninitializedVariantPtr) {
		var enc E
		pEnc := (GDExtensionTypePtr)(unsafe.Pointer(&enc))
		encodeTypePtrArg(in, (GDExtensionUninitializedTypePtr)(pEnc))
		CallFunc_GDExtensionVariantFromTypeConstructorFunc(
			vfn,
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
	var encodeReflectTypePtrArg func(reflect.Value, GDExtensionUninitializedTypePtr)
	switch t {
	case GDEXTENSION_VARIANT_TYPE_INT:
		encodeReflectTypePtrArg = func(rv reflect.Value, pOut GDExtensionUninitializedTypePtr) {
			encodeTypePtrArg((T)(rv.Int()), pOut)
		}
	case GDEXTENSION_VARIANT_TYPE_FLOAT:
		encodeReflectTypePtrArg = func(rv reflect.Value, pOut GDExtensionUninitializedTypePtr) {
			encodeTypePtrArg((T)(rv.Float()), pOut)
		}
	default:
		log.Panic("createNumberEncoder supports only INT and FLOAT")
	}
	var encodeReflectTypePtr func(reflect.Value) GDExtensionTypePtr
	switch t {
	case GDEXTENSION_VARIANT_TYPE_INT:
		encodeReflectTypePtr = func(rv reflect.Value) GDExtensionTypePtr {
			var enc E
			pEnc := (GDExtensionTypePtr)(unsafe.Pointer(&enc))
			encodeReflectTypePtrArg(rv, (GDExtensionUninitializedTypePtr)(pEnc))
			return pEnc
		}
	case GDEXTENSION_VARIANT_TYPE_FLOAT:
		encodeReflectTypePtr = func(rv reflect.Value) GDExtensionTypePtr {
			var enc E
			pEnc := (GDExtensionTypePtr)(unsafe.Pointer(&enc))
			encodeReflectTypePtrArg(rv, (GDExtensionUninitializedTypePtr)(pEnc))
			return pEnc
		}
	default:
		log.Panic("createNumberEncoder supports only INT and FLOAT")
	}
	decodeReflectVariantPtr := func(ptr GDExtensionConstVariantPtr) reflect.Value {
		var enc E
		pEnc := (GDExtensionTypePtr)(unsafe.Pointer(&enc))
		CallFunc_GDExtensionTypeFromVariantConstructorFunc(
			tfn,
			(GDExtensionUninitializedTypePtr)(pEnc),
			(GDExtensionVariantPtr)(ptr),
		)
		var v T
		decodeTypePtrArg((GDExtensionConstTypePtr)(pEnc), &v)
		return reflect.ValueOf(v)
	}
	encodeReflectVariantPtrArg := func(rv reflect.Value, pOut GDExtensionUninitializedVariantPtr) {
		v := rv.Interface().(T)
		CallFunc_GDExtensionVariantFromTypeConstructorFunc(
			vfn,
			pOut,
			(GDExtensionTypePtr)(&v),
		)
	}
	encodeReflectVariantPtr := func(rv reflect.Value) GDExtensionVariantPtr {
		var out Variant
		pOut := (GDExtensionVariantPtr)(unsafe.Pointer(&out))
		encodeReflectVariantPtrArg(rv, (GDExtensionUninitializedVariantPtr)(pOut))
		return pOut
	}
	return argumentEncoder[T, E]{
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
