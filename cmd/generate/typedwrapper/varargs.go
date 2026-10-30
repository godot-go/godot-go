package typedwrapper

// varArgType is one varcall slot type. Kind names the reflect.Kind of the
// scalar families (matching the reflective varcall decoder's outer kind
// switch); it is empty for the array-kind builtins, which the decoder
// resolves on exact type and so are dispatched on exact type here too.
// Decode replicates the reflective varcall decoder's per-type case verbatim
// (method_bind_reflect.go convertVariantToGoTypeReflectValue), Borrow marks
// the convert-construct → byte-borrow → Destroy shape, Release marks
// destroyOwnedContainerArgs membership (the owned copy the dispatch releases
// after the return value has been encoded), and RetEncoder names the encoder
// whose EncodeVariantPtrArg call is exactly what the reflective Variant
// return matrix performs for that type
// (GDExtensionVariantPtrFromReflectValue).
//
// Membership rules (design D4):
//   - Decode covers the ptrcall nativeSlot family plus the two types the
//     varcall decoder serves that ptrcall cannot (float32: ToFloat64 then
//     convert; Go string: retained text copy). Everything nativeSlot carries
//     must appear here because the shared bind overloads constrain one
//     parameter list for both dispatch shapes.
//   - RetEncoder covers the retValueSlot family plus float32 and Go string.
//     Copy-constructor return types are absent: their varcall encode lives in
//     the dedicated copy-ret emission family (ret_release.gen.go), which
//     carries its own destroy-source decision (design D5).
//   - Excluded on purpose: Variant (decoder view semantics), slices (deep
//     copy), Object/Ref/GDClass parameters (decode-error seam and instance
//     binding resolution stay with the reflective path).
//
// The generated surface is a bind-time factory trio (newVarArgDecoder,
// newVarArgReleaser, newVarRetEncoder): each resolves the type argument once
// at bind time — bind-time reflection is allowed, per-call reflection is not
// — and returns the concrete closure the wrapper captures. Per-type method
// families are impossible here: Go forbids declaring methods on predeclared
// types (bool, int, ...).
type varArgType struct {
	Type       string
	Kind       string // scalar family: reflect.Kind string; empty: exact array-kind type
	Decode     string
	Borrow     bool
	Release    bool
	RetEncoder string
	RetFirst   bool // set in generateVarArgs: first varRetSlot member
}

var varArgTypes = []varArgType{
	{Type: "bool", Kind: "reflect.Bool", Decode: "v.ToBool()", RetEncoder: "BoolEncoder"},

	{Type: "int", Kind: "reflect.Int", Decode: "int(v.ToInt64())", RetEncoder: "IntEncoder"},
	{Type: "int8", Kind: "reflect.Int8", Decode: "int8(v.ToInt64())", RetEncoder: "Int8Encoder"},
	{Type: "int16", Kind: "reflect.Int16", Decode: "int16(v.ToInt64())", RetEncoder: "Int16Encoder"},
	{Type: "int32", Kind: "reflect.Int32", Decode: "int32(v.ToInt64())", RetEncoder: "Int32Encoder"},
	{Type: "int64", Kind: "reflect.Int64", Decode: "v.ToInt64()", RetEncoder: "Int64Encoder"},

	{Type: "uint", Kind: "reflect.Uint", Decode: "v.ToUint()", RetEncoder: "UintEncoder"},
	{Type: "uint8", Kind: "reflect.Uint8", Decode: "v.ToUint8()", RetEncoder: "Uint8Encoder"},
	{Type: "uint16", Kind: "reflect.Uint16", Decode: "v.ToUint16()", RetEncoder: "Uint16Encoder"},
	{Type: "uint32", Kind: "reflect.Uint32", Decode: "v.ToUint32()", RetEncoder: "Uint32Encoder"},
	{Type: "uint64", Kind: "reflect.Uint64", Decode: "v.ToUint64()", RetEncoder: "Uint64Encoder"},

	{Type: "float32", Kind: "reflect.Float32", Decode: "float32(v.ToFloat64())", RetEncoder: "Float32Encoder"},
	{Type: "float64", Kind: "reflect.Float64", Decode: "v.ToFloat64()", RetEncoder: "Float64Encoder"},

	// Go string: kind-dispatched like the reflective decoder's String kind.
	{Type: "string", Kind: "reflect.String", Decode: "v.ToGoString()", RetEncoder: "GoStringUtf8Encoder"},

	{Type: "Vector2", Decode: "v.ToVector2()", RetEncoder: "Vector2Encoder"},
	{Type: "Vector2i", Decode: "v.ToVector2i()", RetEncoder: "Vector2iEncoder"},
	{Type: "Vector3", Decode: "v.ToVector3()", RetEncoder: "Vector3Encoder"},
	{Type: "Vector3i", Decode: "v.ToVector3i()", RetEncoder: "Vector3iEncoder"},
	{Type: "Vector4", Decode: "v.ToVector4()", RetEncoder: "Vector4Encoder"},
	{Type: "Vector4i", Decode: "v.ToVector4i()", RetEncoder: "Vector4iEncoder"},
	{Type: "Rect2", Decode: "Rect2Encoder.DecodeVariantPtr(v.NativeConstPtr())", RetEncoder: "Rect2Encoder"},
	{Type: "Rect2i", Decode: "Rect2iEncoder.DecodeVariantPtr(v.NativeConstPtr())", RetEncoder: "Rect2iEncoder"},
	{Type: "Transform2D", Decode: "v.ToTransform2D()", RetEncoder: "Transform2DEncoder"},
	{Type: "Transform3D", Decode: "v.ToTransform3D()", RetEncoder: "Transform3DEncoder"},
	{Type: "Basis", Decode: "v.ToBasis()", RetEncoder: "BasisEncoder"},
	{Type: "AABB", Decode: "v.ToAABB()", RetEncoder: "AABBEncoder"},
	{Type: "Plane", Decode: "v.ToPlane()", RetEncoder: "PlaneEncoder"},
	{Type: "Quaternion", Decode: "v.ToQuaternion()", RetEncoder: "QuaternionEncoder"},
	{Type: "Projection", Decode: "v.ToProjection()", RetEncoder: "ProjectionEncoder"},
	{Type: "Color", Decode: "v.ToColor()", RetEncoder: "ColorEncoder"},

	{Type: "RID", Decode: "v.ToRID()", RetEncoder: "RIDEncoder"},
	// String is an owned copy like the containers below: v.ToString()
	// constructs through the variant constructor (acquire), and
	// destroyOwnedContainerArgs now releases it after the return encoding.
	{Type: "String", Decode: "v.ToString()", Release: true, RetEncoder: "StringEncoder"},

	// StringName/NodePath varcall decode constructs through the engine
	// conversion, keeps a byte borrow of the result, and destroys the
	// constructed value -- the handler receives a call-scoped borrow, exactly
	// like the reflective decoder's StringName/NodePath cases. No return
	// encoder here: their varcall encode lives in the copy-ret emission
	// family with its own destroy-source decision (design D5).
	{Type: "StringName", Borrow: true},
	{Type: "NodePath", Borrow: true},

	// Owned container copies: released after the return encoding, mirroring
	// destroyOwnedContainerArgs (whose String case joins the set above).
	{Type: "Callable", Decode: "v.ToCallable()", Release: true},
	{Type: "Signal", Decode: "v.ToSignal()", Release: true},
	{Type: "Array", Decode: "v.ToArray()", Release: true},
	{Type: "Dictionary", Decode: "v.ToDictionary()", Release: true},
	{Type: "PackedByteArray", Decode: "v.ToPackedByteArray()", Release: true},
	{Type: "PackedInt32Array", Decode: "v.ToPackedInt32Array()", Release: true},
	{Type: "PackedInt64Array", Decode: "v.ToPackedInt64Array()", Release: true},
	{Type: "PackedFloat32Array", Decode: "v.ToPackedFloat32Array()", Release: true},
	{Type: "PackedFloat64Array", Decode: "v.ToPackedFloat64Array()", Release: true},
	{Type: "PackedStringArray", Decode: "v.ToPackedStringArray()", Release: true},
	{Type: "PackedVector2Array", Decode: "v.ToPackedVector2Array()", Release: true},
	{Type: "PackedVector3Array", Decode: "v.ToPackedVector3Array()", Release: true},
	{Type: "PackedVector4Array", Decode: "v.ToPackedVector4Array()", Release: true},
	{Type: "PackedColorArray", Decode: "v.ToPackedColorArray()", Release: true},
}
