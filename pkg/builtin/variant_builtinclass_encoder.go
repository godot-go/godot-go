package builtin

// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

func createBuiltinClassEncoder[T any](t GDExtensionVariantType, opaqueSize int) argumentEncoder[T, T] {
	switch t {
	case
		GDEXTENSION_VARIANT_TYPE_VECTOR2, GDEXTENSION_VARIANT_TYPE_VECTOR2I,
		GDEXTENSION_VARIANT_TYPE_RECT2, GDEXTENSION_VARIANT_TYPE_RECT2I,
		GDEXTENSION_VARIANT_TYPE_VECTOR3, GDEXTENSION_VARIANT_TYPE_VECTOR3I,
		GDEXTENSION_VARIANT_TYPE_TRANSFORM2D, GDEXTENSION_VARIANT_TYPE_VECTOR4,
		GDEXTENSION_VARIANT_TYPE_VECTOR4I, GDEXTENSION_VARIANT_TYPE_PLANE,
		GDEXTENSION_VARIANT_TYPE_QUATERNION, GDEXTENSION_VARIANT_TYPE_AABB,
		GDEXTENSION_VARIANT_TYPE_BASIS, GDEXTENSION_VARIANT_TYPE_TRANSFORM3D,
		GDEXTENSION_VARIANT_TYPE_PROJECTION, GDEXTENSION_VARIANT_TYPE_COLOR,
		GDEXTENSION_VARIANT_TYPE_STRING_NAME, GDEXTENSION_VARIANT_TYPE_NODE_PATH,
		GDEXTENSION_VARIANT_TYPE_RID, GDEXTENSION_VARIANT_TYPE_STRING,
		GDEXTENSION_VARIANT_TYPE_CALLABLE, GDEXTENSION_VARIANT_TYPE_SIGNAL,
		GDEXTENSION_VARIANT_TYPE_DICTIONARY, GDEXTENSION_VARIANT_TYPE_ARRAY,
		GDEXTENSION_VARIANT_TYPE_PACKED_BYTE_ARRAY, GDEXTENSION_VARIANT_TYPE_PACKED_INT32_ARRAY,
		GDEXTENSION_VARIANT_TYPE_PACKED_INT64_ARRAY, GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY,
		GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT64_ARRAY, GDEXTENSION_VARIANT_TYPE_PACKED_STRING_ARRAY,
		GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR2_ARRAY, GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR3_ARRAY,
		GDEXTENSION_VARIANT_TYPE_PACKED_COLOR_ARRAY, GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR4_ARRAY:
	default:
		log.Panic("createEncoder does not support GDEXTENSION_VARIANT_TYPE", zap.Any("type_id", t))
	}
	tfn := typeFromVariantConstructor[t]
	if tfn == nil {
		log.Panic("could not find type from variant constructor", zap.Any("type_id", t))
	}
	vfn := variantFromTypeConstructor[t]
	if vfn == nil {
		log.Panic("could not find variant from type constructor", zap.Any("type_id", t))
	}
	decodeTypePtrArg := func(ptr GDExtensionConstTypePtr, pOut *T) {
		dst := unsafe.Pointer(pOut)
		src := unsafe.Pointer(ptr)
		if dst == src {
			// noop
			return
		}
		*(*T)(unsafe.Pointer(pOut)) = *(*T)(unsafe.Pointer(ptr))
	}
	decodeTypePtr := func(ptr GDExtensionConstTypePtr) T {
		var out T
		decodeTypePtrArg(ptr, &out)
		return out
	}
	encodeTypePtrArg := func(in T, pOut GDExtensionUninitializedTypePtr) {
		*(*T)(unsafe.Pointer(pOut)) = in
	}
	encodeTypePtr := func(in T) GDExtensionTypePtr {
		var out T
		pOut := (GDExtensionTypePtr)(unsafe.Pointer(&out))
		encodeTypePtrArg(in, (GDExtensionUninitializedTypePtr)(pOut))
		return pOut
	}
	decodeVariantPtrArg := func(ptr GDExtensionConstVariantPtr, pOut *T) {
		CallFunc_GDExtensionTypeFromVariantConstructorFunc(
			tfn,
			(GDExtensionUninitializedTypePtr)(unsafe.Pointer(pOut)),
			(GDExtensionVariantPtr)(ptr),
		)
	}
	decodeVariantPtr := func(ptr GDExtensionConstVariantPtr) T {
		var out T
		decodeVariantPtrArg(ptr, &out)
		return out
	}
	encodeVariantPtrArg := func(in T, pOut GDExtensionUninitializedVariantPtr) {
		enc := make([]uint8, opaqueSize)
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
		encodeVariantPtrArg(v, pOut)
	}
	encodeReflectVariantPtr := func(rv reflect.Value) GDExtensionVariantPtr {
		var out Variant
		pOut := (GDExtensionVariantPtr)(unsafe.Pointer(&out))
		encodeReflectVariantPtrArg(rv, (GDExtensionUninitializedVariantPtr)(pOut))
		return pOut
	}
	return argumentEncoder[T, T]{
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
