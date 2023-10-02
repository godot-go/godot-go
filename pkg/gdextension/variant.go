package gdextension

// #include <godot/gdextension_interface.h>
import "C"
import (
	"fmt"
	"strings"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

var (
	variantFromTypeConstructor [GDEXTENSION_VARIANT_TYPE_VARIANT_MAX]GDExtensionVariantFromTypeConstructorFunc
	typeFromVariantConstructor [GDEXTENSION_VARIANT_TYPE_VARIANT_MAX]GDExtensionTypeFromVariantConstructorFunc
)

func variantInitBindings() {
	log.Debug("variantInitBindings called")
	for i := GDExtensionVariantType(1); i < GDEXTENSION_VARIANT_TYPE_VARIANT_MAX; i++ {
		variantFromTypeConstructor[i] = CallFunc_GDExtensionInterfaceGetVariantFromTypeConstructor(i)
		typeFromVariantConstructor[i] = CallFunc_GDExtensionInterfaceGetVariantToTypeConstructor(i)
	}

	initPrimativeTypeEncoders()
	initBuiltinClassEncoders()
	builtinClassesInitBindings()
}

// copy funuction
func copyVariantWithGDExtensionTypePtr(dst GDExtensionUninitializedVariantPtr, src GDExtensionConstVariantPtr) {
	typedDst := (*[VariantSize]uint8)(dst)
	typedSrc := (*[VariantSize]uint8)(src)

	for i := 0; i < VariantSize; i++ {
		typedDst[i] = typedSrc[i]
	}
}

func NewVariantCopyWithGDExtensionConstVariantPtr(ptr GDExtensionConstVariantPtr) Variant {
	ret := Variant{}
	typedSrc := (*[VariantSize]uint8)(ptr)

	for i := 0; i < VariantSize; i++ {
		ret[i] = typedSrc[i]
	}
	return ret
}

func NewVariantNil() Variant {
	ret := Variant{}
	ptr := (GDExtensionUninitializedVariantPtr)(ret.nativePtr())
	GDExtensionVariantPtrWithNil(ptr)
	return ret
}

func GDExtensionVariantPtrWithNil(rOut GDExtensionUninitializedVariantPtr) {
	CallFunc_GDExtensionInterfaceVariantNewNil(rOut)
}

func NewVariantNativeCopy(nativeConstPtr GDExtensionConstVariantPtr) Variant {
	ret := Variant{}
	CallFunc_GDExtensionInterfaceVariantNewCopy((GDExtensionUninitializedVariantPtr)(ret.nativePtr()), nativeConstPtr)
	return ret
}

func NewVariantCopy(dst, src Variant) {
	CallFunc_GDExtensionInterfaceVariantNewCopy((GDExtensionUninitializedVariantPtr)(dst.nativePtr()), src.nativeConstPtr())
}

func NewVariantBool(v bool) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromBool(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromBool(v bool, rOut GDExtensionUninitializedVariantPtr) {
	// MAKE_PTRARGCONV(bool, uint8_t);
	var encoded uint8
	if v {
		encoded = 1
	}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_BOOL]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&encoded),
	)
}

func (c *Variant) ToBool() bool {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_BOOL]
	var v uint8
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v != 0
}

func NewVariantInt64(v int64) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromInt64(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromInt64(v int64, rOut GDExtensionUninitializedVariantPtr) {
	// MAKE_PTRARGCONV(uint8_t, int64_t);
	// MAKE_PTRARGCONV(int8_t, int64_t);
	// MAKE_PTRARGCONV(uint16_t, int64_t);
	// MAKE_PTRARGCONV(int16_t, int64_t);
	// MAKE_PTRARGCONV(uint32_t, int64_t);
	// MAKE_PTRARGCONV(int32_t, int64_t);
	// MAKE_PTRARG(int64_t);
	// MAKE_PTRARG(uint64_t);
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(rOut),
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToInt64() int64 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantFloat64(v float64) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromFloat64(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromFloat64(v float64, rOut GDExtensionUninitializedVariantPtr) {
	// MAKE_PTRARGCONV(float, double);
	// MAKE_PTRARG(double);
	var encoded float64
	encoded = v
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_FLOAT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&encoded),
	)
}

func (c *Variant) ToFloat64() float64 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_FLOAT]
	var v float64
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantNodePath(v NodePath) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromNodePath(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromNodePath(v NodePath, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_NODE_PATH]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToNodePath() NodePath {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_NODE_PATH]
	var v NodePath
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantRID(v RID) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromRID(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromRID(v RID, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_RID]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToRID() RID {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_RID]
	var v RID
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantColor(v Color) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromColor(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromColor(v Color, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_COLOR]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToColor() Color {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_COLOR]
	var v Color
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantQuaternion(v Quaternion) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromQuaternion(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromQuaternion(v Quaternion, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_QUATERNION]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToQuaternion() Quaternion {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_QUATERNION]
	var v Quaternion
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantGoString(v string) Variant {
	gdStr := NewStringWithUtf8Chars(v)
	defer gdStr.Destroy()
	ret := Variant{}
	GDExtensionVariantPtrFromString(gdStr, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func NewVariantString(v String) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromString(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromString(v String, rOut GDExtensionUninitializedVariantPtr) {
	// MAKE_PTRARG(String);
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToString() String {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_STRING]
	var v String
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.nativePtr()),
		c.nativePtr(),
	)
	return v
}

func (c *Variant) ToGoString() string {
	gdStr := c.ToString()
	defer gdStr.Destroy()
	return gdStr.ToUtf8()
}

func NewVariantStringName(v StringName) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromStringName(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromStringName(v StringName, rOut GDExtensionUninitializedVariantPtr) {
	// MAKE_PTRARG(String);
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING_NAME]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToStringName() StringName {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_STRING_NAME]
	var v StringName
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.nativePtr()),
		c.nativePtr(),
	)
	return v
}

func NewVariantObject(v Object) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromObject(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromObject(v Object, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.AsGDExtensionTypePtr(),
	)
}

func (c *Variant) ToObject() Object {
	if c.IsNil() {
		return nil
	}
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	var engineObject *GodotObject
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(unsafe.Pointer(&engineObject)),
		c.nativePtr(),
	)
	ret := getObjectInstanceBinding(engineObject)
	return ret
}

func NewVariantAABB(v AABB) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromAABB(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromAABB(v AABB, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_AABB]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToAABB() AABB {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_AABB]
	var v AABB
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantCallable(v Callable) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromCallable(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromCallable(v Callable, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_CALLABLE]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToCallable() Callable {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_CALLABLE]
	var v Callable
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantTransform2D(v Transform2D) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromTransform2D(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromTransform2D(v Transform2D, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_TRANSFORM2D]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToTransform2D() Transform2D {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_TRANSFORM2D]
	var v Transform2D
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantTransform3D(v Transform3D) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromTransform3D(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromTransform3D(v Transform3D, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_TRANSFORM3D]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToTransform3D() Transform3D {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_TRANSFORM3D]
	var v Transform3D
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantPlane(v Plane) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromPlane(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromPlane(v Plane, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PLANE]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToPlane() Plane {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PLANE]
	var v Plane
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantPackedByteArray(v PackedByteArray) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromPackedByteArray(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromPackedByteArray(v PackedByteArray, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToPackedByteArray() PackedByteArray {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY]
	var v PackedByteArray
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantPackedStringArray(v PackedStringArray) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromPackedStringArray(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromPackedStringArray(v PackedStringArray, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_STRING_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToPackedStringArray() PackedStringArray {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_STRING_ARRAY]
	var v PackedStringArray
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantPackedVector2Array(v PackedVector2Array) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromPackedVector2Array(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromPackedVector2Array(v PackedVector2Array, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR2_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToPackedVector2Array() PackedVector2Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR2_ARRAY]
	var v PackedVector2Array
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantPackedVector3Array(v PackedVector3Array) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromPackedVector3Array(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromPackedVector3Array(v PackedVector3Array, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR3_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToPackedVector3Array() PackedVector3Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR3_ARRAY]
	var v PackedVector3Array
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantPackedInt32Array(v PackedInt32Array) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromPackedInt32Array(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromPackedInt32Array(v PackedInt32Array, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_INT32_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToPackedInt32Array() PackedInt32Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_INT32_ARRAY]
	var v PackedInt32Array
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantPackedInt64Array(v PackedInt64Array) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromPackedInt64Array(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromPackedInt64Array(v PackedInt64Array, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_INT64_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToPackedInt64Array() PackedInt64Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_INT64_ARRAY]
	var v PackedInt64Array
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantPackedFloat32Array(v PackedFloat32Array) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromPackedFloat32Array(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromPackedFloat32Array(v PackedFloat32Array, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToPackedFloat32Array() PackedFloat32Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY]
	var v PackedFloat32Array
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantPackedFloat64Array(v PackedFloat64Array) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromPackedFloat64Array(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromPackedFloat64Array(v PackedFloat64Array, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT64_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToPackedFloat64Array() PackedFloat64Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT64_ARRAY]
	var v PackedFloat64Array
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantBasis(v Basis) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromBasis(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromBasis(v Basis, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_BASIS]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToBasis() Basis {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_BASIS]
	var v Basis
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantPackedColorArray(v PackedColorArray) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromPackedColorArray(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromPackedColorArray(v PackedColorArray, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_COLOR_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		(GDExtensionTypePtr)(&v),
	)
}

func (c *Variant) ToPackedColorArray() PackedColorArray {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_PACKED_COLOR_ARRAY]
	var v PackedColorArray
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		c.nativePtr(),
	)
	return v
}

func NewVariantVector2(v Vector2) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromVector2(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromVector2(v Vector2, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR2]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToVector2() Vector2 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR2]
	var v Vector2
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.nativePtr()),
		c.nativePtr(),
	)
	return v
}

func NewVariantVector2i(v Vector2i) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromVector2i(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromVector2i(v Vector2i, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR2I]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToVector2i() Vector2i {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR2I]
	var v Vector2i
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.nativePtr()),
		c.nativePtr(),
	)
	return v
}

func NewVariantRect2(v Rect2) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromRect2(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromRect2(v Rect2, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_RECT2]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToRect2() Rect2 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_RECT2]
	var v Rect2
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.nativePtr()),
		c.nativePtr(),
	)
	return v
}

func NewVariantRect2i(v Rect2i) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromRect2i(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromRect2i(v Rect2i, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_RECT2I]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToRect2i() Rect2i {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_RECT2I]
	var v Rect2i
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.nativePtr()),
		c.nativePtr(),
	)
	return v
}

func NewVariantVector3(v Vector3) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromVector3(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromVector3(v Vector3, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR3]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToVector3() Vector3 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR3]
	var v Vector3
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.nativePtr()),
		c.nativePtr(),
	)
	return v
}

func NewVariantVector3i(v Vector3i) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromVector3i(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromVector3i(v Vector3i, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR3I]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToVector3i() Vector3i {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR3I]
	var v Vector3i
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.nativePtr()),
		c.nativePtr(),
	)
	return v
}

func NewVariantVector4(v Vector4) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromVector4(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromVector4(v Vector4, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR4]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToVector4() Vector4 {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR4]
	var v Vector4
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.nativePtr()),
		c.nativePtr(),
	)
	return v
}

func NewVariantVector4i(v Vector4i) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromVector4i(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromVector4i(v Vector4i, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR4I]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToVector4i() Vector4i {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR4I]
	var v Vector4i
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.nativePtr()),
		c.nativePtr(),
	)
	return v
}

func NewVariantArray(v Array) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromArray(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromArray(v Array, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_ARRAY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		rOut,
		v.nativePtr(),
	)
}

func (c *Variant) ToArray() Array {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_ARRAY]
	var arr Array
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(arr.nativePtr()),
		c.nativePtr(),
	)
	return arr
}

func NewVariantDictionary(v Dictionary) Variant {
	ret := Variant{}
	GDExtensionVariantPtrFromDictionary(v, (GDExtensionUninitializedVariantPtr)(ret.nativePtr()))
	return ret
}

func GDExtensionVariantPtrFromDictionary(v Dictionary, rOut GDExtensionUninitializedVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_DICTIONARY]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(rOut),
		(GDExtensionTypePtr)(v.nativePtr()),
	)
}

func (c *Variant) ToDictionary() Dictionary {
	fn := typeFromVariantConstructor[GDEXTENSION_VARIANT_TYPE_DICTIONARY]
	var dict Dictionary
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(dict.nativePtr()),
		(GDExtensionVariantPtr)(c.nativePtr()),
	)
	return dict
}

const (
	VariantSize = 24
)

type Variant [VariantSize]uint8

func (c *Variant) nativeConstPtr() GDExtensionConstVariantPtr {
	return (GDExtensionConstVariantPtr)(c)
}

func (c *Variant) nativePtr() GDExtensionVariantPtr {
	return (GDExtensionVariantPtr)(c)
}

func (c *Variant) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return (GDExtensionConstTypePtr)(c)
}

func (c *Variant) AsGDExtensionTypePtr() GDExtensionTypePtr {
	return (GDExtensionTypePtr)(c)
}

var (
	needsDeinit = [GDEXTENSION_VARIANT_TYPE_VARIANT_MAX]bool{
		false, //NIL,
		false, //BOOL,
		false, //INT,
		false, //FLOAT,
		true,  //STRING,
		false, //VECTOR2,
		false, //VECTOR2I,
		false, //RECT2,
		false, //RECT2I,
		false, //VECTOR3,
		false, //VECTOR3I,
		true,  //TRANSFORM2D,
		false, //PLANE,
		false, //QUATERNION,
		true,  //AABB,
		true,  //BASIS,
		true,  //TRANSFORM,

		// misc types
		false, //COLOR,
		true,  //STRING_NAME,
		true,  //NODE_PATH,
		false, //RID,
		true,  //OBJECT,
		true,  //CALLABLE,
		true,  //SIGNAL,
		true,  //DICTIONARY,
		true,  //ARRAY,

		// typed arrays
		true, //PACKED_BYTE_ARRAY,
		true, //PACKED_INT32_ARRAY,
		true, //PACKED_INT64_ARRAY,
		true, //PACKED_FLOAT32_ARRAY,
		true, //PACKED_FLOAT64_ARRAY,
		true, //PACKED_STRING_ARRAY,
		true, //PACKED_VECTOR2_ARRAY,
		true, //PACKED_VECTOR3_ARRAY,
		true, //PACKED_COLOR_ARRAY,
	}
)

func (c *Variant) Call(
	method string,
	args []*Variant,
) (*Variant, error) {
	var (
		callArgs *GDExtensionConstVariantPtr
		r_ret    Variant
	)
	sn := NewStringNameWithLatin1Chars(method)
	defer sn.Destroy()
	callArgs = AllocCopyVariantPtrSliceAsGDExtensionVariantPtrPtr(args)
	callArgCount := len(args)
	var err GDExtensionCallError
	CallFunc_GDExtensionInterfaceVariantCall(
		(GDExtensionVariantPtr)(c.nativePtr()),
		(GDExtensionConstStringNamePtr)(sn.nativePtr()),
		callArgs,
		(GDExtensionInt)(callArgCount),
		(GDExtensionUninitializedVariantPtr)(r_ret.nativePtr()),
		&err,
	)
	if err.Ok() {
		return &r_ret, nil
	}
	return nil, err
}

func (c *Variant) CallStatic(
	vt GDExtensionVariantType,
	method string,
	args []*Variant,
	r_ret *Variant,
) error {
	var (
		callArgs *GDExtensionConstVariantPtr
	)
	sn := NewStringNameWithLatin1Chars(method)
	defer sn.Destroy()
	callArgs = AllocCopyVariantPtrSliceAsGDExtensionVariantPtrPtr(args)
	callArgCount := len(args)
	var err GDExtensionCallError
	CallFunc_GDExtensionInterfaceVariantCallStatic(
		vt,
		(GDExtensionConstStringNamePtr)(sn.nativePtr()),
		callArgs,
		(GDExtensionInt)(callArgCount),
		(GDExtensionUninitializedVariantPtr)(r_ret.nativePtr()),
		&err,
	)
	if err.Ok() {
		return nil
	}
	return err
}

func (c *Variant) GetType() GDExtensionVariantType {
	return CallFunc_GDExtensionInterfaceVariantGetType((GDExtensionConstVariantPtr)(c.nativePtr()))
}

func (c *Variant) Clear() {
	if needsDeinit[(int)(c.GetType())] {
		CallFunc_GDExtensionInterfaceVariantDestroy((GDExtensionVariantPtr)(c.nativePtr()))
	}
	CallFunc_GDExtensionInterfaceVariantNewNil((GDExtensionUninitializedVariantPtr)(c.nativePtr()))
}

var (
	ErrOutOfBounds = fmt.Errorf("out of bounds")
	ErrInvalid     = fmt.Errorf("invalid")
)

func (c *Variant) Set(key Variant, value Variant) error {
	var valid GDExtensionBool
	CallFunc_GDExtensionInterfaceVariantSet(
		c.nativePtr(),
		key.nativeConstPtr(), value.nativeConstPtr(), &valid)
	if valid != 0 {
		return ErrInvalid
	}
	return nil
}

func (c *Variant) SetNamed(name StringName, value Variant) error {
	var valid GDExtensionBool
	CallFunc_GDExtensionInterfaceVariantSetNamed(
		(GDExtensionVariantPtr)(unsafe.Pointer(c.nativePtr())),
		(GDExtensionConstStringNamePtr)(unsafe.Pointer(name.nativePtr())),
		(GDExtensionConstVariantPtr)(unsafe.Pointer(value.nativePtr())), &valid)
	if valid != 0 {
		return ErrInvalid
	}
	return nil
}

func (c *Variant) SetIndexed(index int, value Variant) error {
	var valid, oob GDExtensionBool
	CallFunc_GDExtensionInterfaceVariantSetIndexed(
		(GDExtensionVariantPtr)(unsafe.Pointer(c.nativePtr())),
		(GDExtensionInt)(index), value.nativeConstPtr(), &valid, &oob)
	if valid == 0 {
		return ErrInvalid
	}
	if oob != 0 {
		return ErrOutOfBounds
	}
	return nil
}

func (c *Variant) SetKeyed(key, value Variant) bool {
	var valid GDExtensionBool
	CallFunc_GDExtensionInterfaceVariantSetKeyed(
		(GDExtensionVariantPtr)(unsafe.Pointer(c.nativePtr())),
		key.nativeConstPtr(),
		value.nativeConstPtr(),
		&valid)
	return valid != 0
}

func (c *Variant) GetIndexed(index int) (Variant, error) {
	var (
		result Variant
		valid  GDExtensionBool
		oob    GDExtensionBool
	)
	CallFunc_GDExtensionInterfaceVariantGetIndexed(
		c.nativeConstPtr(), (GDExtensionInt)(index), (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(result.nativePtr())), &valid, &oob)
	if valid == 0 {
		return result, ErrInvalid
	}
	if oob != 0 {
		return result, ErrOutOfBounds
	}
	return result, nil
}

func (c *Variant) GetKeyed(key Variant) (Variant, error) {
	var (
		result Variant
		valid  GDExtensionBool
	)
	CallFunc_GDExtensionInterfaceVariantGetKeyed(
		c.nativeConstPtr(), key.nativeConstPtr(), (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(result.nativePtr())), &valid)
	if valid == 0 {
		return result, ErrInvalid
	}
	return result, nil
}

func (c *Variant) Destroy() {
	CallFunc_GDExtensionInterfaceVariantDestroy((GDExtensionVariantPtr)(c.nativePtr()))
}

func (c *Variant) Stringify() string {
	ret := NewString()
	defer ret.Destroy()
	CallFunc_GDExtensionInterfaceVariantStringify((GDExtensionConstVariantPtr)(c.nativePtr()), (GDExtensionStringPtr)(ret.nativePtr()))
	return ret.ToUtf8()
}

func (c *Variant) IsNil() bool {
	if c == nil {
		return true
	}
	for i := range c {
		if c[i] != 0 {
			return false
		}
	}
	return true
}

func Stringify(v Variant) string {
	return v.Stringify()
}

func VariantSliceToString(values []Variant) string {
	sb := strings.Builder{}
	sb.WriteString("[]Variant(")
	for i := range values {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(values[i].Stringify())
	}
	sb.WriteString(")")
	return sb.String()
}

func VariantPtrSliceToString(values []*Variant) string {
	sb := strings.Builder{}
	sb.WriteString("[]Variant(")
	for i := range values {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(values[i].Stringify())
	}
	sb.WriteString(")")
	return sb.String()
}

func ZapVariant(key string, v Variant) zap.Field {
	value := v.Stringify()
	return zap.String(key, value)
}

func ZapVector2(key string, v Vector2) zap.Field {
	variant := NewVariantVector2(v)
	defer variant.Destroy()
	value := variant.Stringify()
	return zap.String(key, value)
}

func ZapVector3(key string, v Vector3) zap.Field {
	variant := NewVariantVector3(v)
	defer variant.Destroy()
	value := variant.Stringify()
	return zap.String(key, value)
}

func ZapVector4(key string, v Vector4) zap.Field {
	variant := NewVariantVector4(v)
	defer variant.Destroy()
	value := variant.Stringify()
	return zap.String(key, value)
}
