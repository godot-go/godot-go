package builtin

// #include <godot/gdextension_interface.h>
import "C"
import (
	. "github.com/godot-go/godot-go/pkg/ffi"
)

func NewVariantBool(v bool) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromBool(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromBool(v bool, rOut GDExtensionUninitializedVariantPtr) {
	var encoded uint8
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	BoolEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_BOOL]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToBool() bool {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_BOOL]
	var v uint8
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded bool
	BoolEncoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantUint(v uint) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromUint(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromUint(v uint, rOut GDExtensionUninitializedVariantPtr) {
	var encoded int64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	UintEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToUint() uint {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded uint
	UintEncoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantInt(v int) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromInt(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromInt(v int, rOut GDExtensionUninitializedVariantPtr) {
	var encoded int64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	IntEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToInt() int {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded int
	IntEncoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantUint8(v uint8) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromUint8(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromUint8(v uint8, rOut GDExtensionUninitializedVariantPtr) {
	var encoded int64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	Uint8Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToUint8() uint8 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded uint8
	Uint8Encoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantInt8(v int8) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromInt8(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromInt8(v int8, rOut GDExtensionUninitializedVariantPtr) {
	var encoded int64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	Int8Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToInt8() int8 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded int8
	Int8Encoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantUint16(v uint16) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromUint16(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromUint16(v uint16, rOut GDExtensionUninitializedVariantPtr) {
	var encoded int64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	Uint16Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToUint16() uint16 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded uint16
	Uint16Encoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantInt16(v int16) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromInt16(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromInt16(v int16, rOut GDExtensionUninitializedVariantPtr) {
	var encoded int64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	Int16Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToInt16() int16 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded int16
	Int16Encoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantUint32(v uint32) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromUint32(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromUint32(v uint32, rOut GDExtensionUninitializedVariantPtr) {
	var encoded int64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	Uint32Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToUint32() uint32 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded uint32
	Uint32Encoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantInt32(v int32) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromInt32(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromInt32(v int32, rOut GDExtensionUninitializedVariantPtr) {
	var encoded int64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	Int32Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToInt32() int32 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded int32
	Int32Encoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantUint64(v uint64) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromUint64(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromUint64(v uint64, rOut GDExtensionUninitializedVariantPtr) {
	var encoded int64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	Uint64Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToUint64() uint64 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded uint64
	Uint64Encoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantInt64(v int64) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromInt64(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromInt64(v int64, rOut GDExtensionUninitializedVariantPtr) {
	var encoded int64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	Int64Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToInt64() int64 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded int64
	Int64Encoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantFloat32(v float32) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromFloat32(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromFloat32(v float32, rOut GDExtensionUninitializedVariantPtr) {
	var encoded float64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	Float32Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_FLOAT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToFloat32() float32 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_FLOAT]
	var v float64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded float32
	Float32Encoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantFloat64(v float64) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromFloat64(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromFloat64(v float64, rOut GDExtensionUninitializedVariantPtr) {
	var encoded float64
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	Float64Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_FLOAT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToFloat64() float64 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_FLOAT]
	var v float64
	ptr := (GDExtensionTypePtr)(&v)
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	var decoded float64
	Float64Encoder.DecodeTypePtrArg((GDExtensionConstTypePtr)(ptr), &decoded)
	return decoded
}

func NewVariantString(v String) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromString(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromString(v String, rOut GDExtensionUninitializedVariantPtr) {
	var encoded String
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	StringEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToString() String {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_STRING]
	var v String
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantVector2(v Vector2) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromVector2(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromVector2(v Vector2, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Vector2
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	Vector2Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR2]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToVector2() Vector2 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR2]
	var v Vector2
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantVector2i(v Vector2i) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromVector2i(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromVector2i(v Vector2i, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Vector2i
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	Vector2iEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR2I]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToVector2i() Vector2i {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR2I]
	var v Vector2i
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantVector3(v Vector3) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromVector3(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromVector3(v Vector3, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Vector3
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	Vector3Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR3]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToVector3() Vector3 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR3]
	var v Vector3
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantVector3i(v Vector3i) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromVector3i(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromVector3i(v Vector3i, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Vector3i
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	Vector3iEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR3I]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToVector3i() Vector3i {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR3I]
	var v Vector3i
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantTransform2D(v Transform2D) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromTransform2D(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromTransform2D(v Transform2D, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Transform2D
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	Transform2DEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_TRANSFORM2D]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToTransform2D() Transform2D {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_TRANSFORM2D]
	var v Transform2D
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantVector4(v Vector4) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromVector4(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromVector4(v Vector4, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Vector4
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	Vector4Encoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR4]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToVector4() Vector4 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR4]
	var v Vector4
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantVector4i(v Vector4i) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromVector4i(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromVector4i(v Vector4i, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Vector4i
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	Vector4iEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR4I]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToVector4i() Vector4i {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR4I]
	var v Vector4i
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantPlane(v Plane) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromPlane(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromPlane(v Plane, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Plane
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	PlaneEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PLANE]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToPlane() Plane {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PLANE]
	var v Plane
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantQuaternion(v Quaternion) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromQuaternion(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromQuaternion(v Quaternion, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Quaternion
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	QuaternionEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_QUATERNION]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToQuaternion() Quaternion {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_QUATERNION]
	var v Quaternion
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantAABB(v AABB) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromAABB(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromAABB(v AABB, rOut GDExtensionUninitializedVariantPtr) {
	var encoded AABB
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	AABBEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_AABB]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToAABB() AABB {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_AABB]
	var v AABB
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantBasis(v Basis) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromBasis(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromBasis(v Basis, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Basis
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	BasisEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_BASIS]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToBasis() Basis {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_BASIS]
	var v Basis
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantTransform3D(v Transform3D) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromTransform3D(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromTransform3D(v Transform3D, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Transform3D
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	Transform3DEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_TRANSFORM3D]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToTransform3D() Transform3D {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_TRANSFORM3D]
	var v Transform3D
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantProjection(v Projection) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromProjection(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromProjection(v Projection, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Projection
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	ProjectionEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PROJECTION]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToProjection() Projection {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PROJECTION]
	var v Projection
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantColor(v Color) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromColor(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromColor(v Color, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Color
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	ColorEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_COLOR]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToColor() Color {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_COLOR]
	var v Color
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantStringName(v StringName) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromStringName(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromStringName(v StringName, rOut GDExtensionUninitializedVariantPtr) {
	var encoded StringName
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	StringNameEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING_NAME]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToStringName() StringName {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_STRING_NAME]
	var v StringName
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantNodePath(v NodePath) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromNodePath(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromNodePath(v NodePath, rOut GDExtensionUninitializedVariantPtr) {
	var encoded NodePath
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	NodePathEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_NODE_PATH]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToNodePath() NodePath {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_NODE_PATH]
	var v NodePath
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantRID(v RID) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromRID(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromRID(v RID, rOut GDExtensionUninitializedVariantPtr) {
	var encoded RID
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	RIDEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_RID]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToRID() RID {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_RID]
	var v RID
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantCallable(v Callable) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromCallable(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromCallable(v Callable, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Callable
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	CallableEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_CALLABLE]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToCallable() Callable {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_CALLABLE]
	var v Callable
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantSignal(v Signal) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromSignal(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromSignal(v Signal, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Signal
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	SignalEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_SIGNAL]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToSignal() Signal {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_SIGNAL]
	var v Signal
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantDictionary(v Dictionary) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromDictionary(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromDictionary(v Dictionary, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Dictionary
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	DictionaryEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_DICTIONARY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToDictionary() Dictionary {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_DICTIONARY]
	var v Dictionary
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantArray(v Array) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromArray(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromArray(v Array, rOut GDExtensionUninitializedVariantPtr) {
	var encoded Array
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	ArrayEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToArray() Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_ARRAY]
	var v Array
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantPackedByteArray(v PackedByteArray) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromPackedByteArray(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromPackedByteArray(v PackedByteArray, rOut GDExtensionUninitializedVariantPtr) {
	var encoded PackedByteArray
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	PackedByteArrayEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_BYTE_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToPackedByteArray() PackedByteArray {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_BYTE_ARRAY]
	var v PackedByteArray
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantPackedInt32Array(v PackedInt32Array) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromPackedInt32Array(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromPackedInt32Array(v PackedInt32Array, rOut GDExtensionUninitializedVariantPtr) {
	var encoded PackedInt32Array
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	PackedInt32ArrayEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_INT32_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToPackedInt32Array() PackedInt32Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_INT32_ARRAY]
	var v PackedInt32Array
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantPackedInt64Array(v PackedInt64Array) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromPackedInt64Array(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromPackedInt64Array(v PackedInt64Array, rOut GDExtensionUninitializedVariantPtr) {
	var encoded PackedInt64Array
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	PackedInt64ArrayEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_INT64_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToPackedInt64Array() PackedInt64Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_INT64_ARRAY]
	var v PackedInt64Array
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantPackedFloat32Array(v PackedFloat32Array) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromPackedFloat32Array(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromPackedFloat32Array(v PackedFloat32Array, rOut GDExtensionUninitializedVariantPtr) {
	var encoded PackedFloat32Array
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	PackedFloat32ArrayEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToPackedFloat32Array() PackedFloat32Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY]
	var v PackedFloat32Array
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantPackedFloat64Array(v PackedFloat64Array) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromPackedFloat64Array(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromPackedFloat64Array(v PackedFloat64Array, rOut GDExtensionUninitializedVariantPtr) {
	var encoded PackedFloat64Array
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	PackedFloat64ArrayEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT64_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToPackedFloat64Array() PackedFloat64Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT64_ARRAY]
	var v PackedFloat64Array
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantPackedStringArray(v PackedStringArray) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromPackedStringArray(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromPackedStringArray(v PackedStringArray, rOut GDExtensionUninitializedVariantPtr) {
	var encoded PackedStringArray
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	PackedStringArrayEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_STRING_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToPackedStringArray() PackedStringArray {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_STRING_ARRAY]
	var v PackedStringArray
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantPackedVector2Array(v PackedVector2Array) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromPackedVector2Array(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromPackedVector2Array(v PackedVector2Array, rOut GDExtensionUninitializedVariantPtr) {
	var encoded PackedVector2Array
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	PackedVector2ArrayEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR2_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToPackedVector2Array() PackedVector2Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR2_ARRAY]
	var v PackedVector2Array
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantPackedVector3Array(v PackedVector3Array) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromPackedVector3Array(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromPackedVector3Array(v PackedVector3Array, rOut GDExtensionUninitializedVariantPtr) {
	var encoded PackedVector3Array
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	PackedVector3ArrayEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR3_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToPackedVector3Array() PackedVector3Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR3_ARRAY]
	var v PackedVector3Array
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}

func NewVariantPackedColorArray(v PackedColorArray) Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.NativePtr())
	pnr.Pin(ptr)
	GDExtensionVariantPtrFromPackedColorArray(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromPackedColorArray(v PackedColorArray, rOut GDExtensionUninitializedVariantPtr) {
	var encoded PackedColorArray
	encodedPtr := (GDExtensionTypePtr)(&encoded)
	pnr.Pin(encodedPtr)
	PackedColorArrayEncoder.EncodeTypePtrArg(v, (GDExtensionUninitializedTypePtr)(encodedPtr))
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_COLOR_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		encodedPtr,
	)
}

func (c Variant) ToPackedColorArray() PackedColorArray {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_COLOR_ARRAY]
	var v PackedColorArray
	ptr := v.NativePtr()
	pnr.Pin(ptr)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(ptr),
		c.NativePtr(),
	)
	return v
}
