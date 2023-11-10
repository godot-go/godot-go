package builtin

// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextension/ffi"
)

type ArgumentEncoder interface {
	DecodeReflectTypePtr(GDExtensionConstTypePtr) reflect.Value
	EncodeReflectTypePtrArg(reflect.Value, GDExtensionUninitializedTypePtr)
	EncodeReflectTypePtr(reflect.Value) GDExtensionTypePtr
	DecodeReflectVariantPtr(GDExtensionConstVariantPtr) reflect.Value
	EncodeReflectVariantPtrArg(reflect.Value, GDExtensionUninitializedVariantPtr)
	EncodeReflectVariantPtr(reflect.Value) GDExtensionVariantPtr
}

type argumentEncoder[T any, E any] struct {
	DecodeTypePtrArg           func(GDExtensionConstTypePtr, *T)
	DecodeTypePtr              func(GDExtensionConstTypePtr) T
	EncodeTypePtrArg           func(T, GDExtensionUninitializedTypePtr)
	EncodeTypePtr              func(T) GDExtensionTypePtr
	DecodeVariantPtrArg        func(GDExtensionConstVariantPtr, *T)
	DecodeVariantPtr           func(GDExtensionConstVariantPtr) T
	EncodeVariantPtrArg        func(T, GDExtensionUninitializedVariantPtr)
	EncodeVariantPtr           func(T) GDExtensionVariantPtr
	decodeReflectTypePtr       func(GDExtensionConstTypePtr) reflect.Value
	encodeReflectTypePtrArg    func(reflect.Value, GDExtensionUninitializedTypePtr)
	encodeReflectTypePtr       func(reflect.Value) GDExtensionTypePtr
	decodeReflectVariantPtr    func(GDExtensionConstVariantPtr) reflect.Value
	encodeReflectVariantPtrArg func(reflect.Value, GDExtensionUninitializedVariantPtr)
	encodeReflectVariantPtr    func(reflect.Value) GDExtensionVariantPtr
}

func (e argumentEncoder[T, E]) DecodeReflectTypePtr(ptr GDExtensionConstTypePtr) reflect.Value {
	return e.decodeReflectTypePtr(ptr)
}

func (e argumentEncoder[T, E]) EncodeReflectTypePtrArg(rv reflect.Value, pOut GDExtensionUninitializedTypePtr) {
	e.encodeReflectTypePtrArg(rv, pOut)
}

func (e argumentEncoder[T, E]) EncodeReflectTypePtr(rv reflect.Value) GDExtensionTypePtr {
	return e.encodeReflectTypePtr(rv)
}

func (e argumentEncoder[T, E]) DecodeReflectVariantPtr(ptr GDExtensionConstVariantPtr) reflect.Value {
	return e.decodeReflectVariantPtr(ptr)
}

func (e argumentEncoder[T, E]) EncodeReflectVariantPtrArg(rv reflect.Value, pOut GDExtensionUninitializedVariantPtr) {
	e.encodeReflectVariantPtrArg(rv, pOut)
}

func (e argumentEncoder[T, E]) EncodeReflectVariantPtr(rv reflect.Value) GDExtensionVariantPtr {
	return e.encodeReflectVariantPtr(rv)
}

type objectArgumentEncoder[T Object] struct {
	DecodeTypePtrArg           func(GDExtensionConstTypePtr, T)
	DecodeTypePtr              func(GDExtensionConstTypePtr) T
	EncodeTypePtrArg           func(T, GDExtensionUninitializedTypePtr)
	EncodeTypePtr              func(T) GDExtensionTypePtr
	DecodeVariantPtrArg        func(GDExtensionConstVariantPtr, T)
	DecodeVariantPtr           func(GDExtensionConstVariantPtr) T
	EncodeVariantPtrArg        func(T, GDExtensionUninitializedVariantPtr)
	EncodeVariantPtr           func(T) GDExtensionVariantPtr
	decodeReflectTypePtr       func(GDExtensionConstTypePtr) reflect.Value
	encodeReflectTypePtrArg    func(reflect.Value, GDExtensionUninitializedTypePtr)
	encodeReflectTypePtr       func(reflect.Value) GDExtensionTypePtr
	decodeReflectVariantPtr    func(GDExtensionConstVariantPtr) reflect.Value
	encodeReflectVariantPtrArg func(reflect.Value, GDExtensionUninitializedVariantPtr)
	encodeReflectVariantPtr    func(reflect.Value) GDExtensionVariantPtr
}

func (e objectArgumentEncoder[T]) DecodeReflectTypePtr(ptr GDExtensionConstTypePtr) reflect.Value {
	return e.decodeReflectTypePtr(ptr)
}

func (e objectArgumentEncoder[T]) EncodeReflectTypePtrArg(rv reflect.Value, pOut GDExtensionUninitializedTypePtr) {
	e.encodeReflectTypePtrArg(rv, pOut)
}

func (e objectArgumentEncoder[T]) EncodeReflectTypePtr(rv reflect.Value) GDExtensionTypePtr {
	return e.encodeReflectTypePtr(rv)
}

func (e objectArgumentEncoder[T]) DecodeReflectVariantPtr(ptr GDExtensionConstVariantPtr) reflect.Value {
	return e.decodeReflectVariantPtr(ptr)
}

func (e objectArgumentEncoder[T]) EncodeReflectVariantPtrArg(rv reflect.Value, pOut GDExtensionUninitializedVariantPtr) {
	e.encodeReflectVariantPtrArg(rv, pOut)
}

func (e objectArgumentEncoder[T]) EncodeReflectVariantPtr(rv reflect.Value) GDExtensionVariantPtr {
	return e.encodeReflectVariantPtr(rv)
}

const (
	ObjectSize = (int)(unsafe.Sizeof((*int)(nil)))
)

var (
	BoolEncoder           argumentEncoder[bool, uint8]
	UintEncoder           argumentEncoder[uint, int64]
	IntEncoder            argumentEncoder[int, int64]
	Uint8Encoder          argumentEncoder[uint8, int64]
	Int8Encoder           argumentEncoder[int8, int64]
	Uint16Encoder         argumentEncoder[uint16, int64]
	Int16Encoder          argumentEncoder[int16, int64]
	Uint32Encoder         argumentEncoder[uint32, int64]
	Int32Encoder          argumentEncoder[int32, int64]
	Uint64Encoder         argumentEncoder[uint64, int64]
	Int64Encoder          argumentEncoder[int64, int64]
	Float32Encoder        argumentEncoder[float32, float64]
	Float64Encoder        argumentEncoder[float64, float64]
	GoStringUtf8Encoder   argumentEncoder[string, String]
	GoStringLatin1Encoder argumentEncoder[string, String]
	ObjectEncoder         objectArgumentEncoder[Object]
	VariantEncoder        argumentEncoder[Variant, Variant]
)

func initPrimativeTypeEncoders() {
	// bool type
	BoolEncoder = createBoolEncoder[bool, uint8]()

	// integer type
	UintEncoder = createNumberEncoder[uint, int64](GDEXTENSION_VARIANT_TYPE_INT)
	IntEncoder = createNumberEncoder[int, int64](GDEXTENSION_VARIANT_TYPE_INT)
	Uint8Encoder = createNumberEncoder[uint8, int64](GDEXTENSION_VARIANT_TYPE_INT)
	Int8Encoder = createNumberEncoder[int8, int64](GDEXTENSION_VARIANT_TYPE_INT)
	Uint16Encoder = createNumberEncoder[uint16, int64](GDEXTENSION_VARIANT_TYPE_INT)
	Int16Encoder = createNumberEncoder[int16, int64](GDEXTENSION_VARIANT_TYPE_INT)
	Uint32Encoder = createNumberEncoder[uint32, int64](GDEXTENSION_VARIANT_TYPE_INT)
	Int32Encoder = createNumberEncoder[int32, int64](GDEXTENSION_VARIANT_TYPE_INT)
	Uint64Encoder = createNumberEncoder[uint64, int64](GDEXTENSION_VARIANT_TYPE_INT)
	Int64Encoder = createNumberEncoder[int64, int64](GDEXTENSION_VARIANT_TYPE_INT)

	// float types
	Float32Encoder = createNumberEncoder[float32, float64](GDEXTENSION_VARIANT_TYPE_FLOAT)
	Float64Encoder = createNumberEncoder[float64, float64](GDEXTENSION_VARIANT_TYPE_FLOAT)

	// native go string to String
	GoStringUtf8Encoder = createGoStringEncoder(Utf8GoStringFormat)
	GoStringLatin1Encoder = createGoStringEncoder(Latin1GoStringFormat)

	ObjectEncoder = CreateObjectEncoder[Object]()
	VariantEncoder = createVariantEncoder()
}
