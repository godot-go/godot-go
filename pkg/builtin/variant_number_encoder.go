package builtin

// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
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
			switch rv.Kind() {
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				encodeTypePtrArg((T)(rv.Uint()), pOut)
			default:
				encodeTypePtrArg((T)(rv.Int()), pOut)
			}
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
		var v T
		switch rv.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v = (T)(rv.Uint())
		case reflect.Float32, reflect.Float64:
			v = (T)(rv.Float())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v = (T)(rv.Int())
		default:
			v = rv.Interface().(T)
		}
		var enc E
		pEnc := (GDExtensionTypePtr)(unsafe.Pointer(&enc))
		encodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(pEnc))
		CallFunc_GDExtensionVariantFromTypeConstructorFunc(
			vfn,
			pOut,
			pEnc,
		)
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
