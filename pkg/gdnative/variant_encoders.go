package gdnative

import (
	"unsafe"

	"golang.org/x/exp/constraints"
)

type ArgumentEncoder[I any, O any] struct {
	Decode    func(unsafe.Pointer) I
	Encode    func(I, unsafe.Pointer)
	EncodeArg func(I) O
}

type ArgumentReferenceEncoder[I any, O any] struct {
	Decode    func(unsafe.Pointer) I
	Encode    func(*I, unsafe.Pointer)
	EncodeArg func(*I) O
}

var (
	// bool type
	BoolEncoder = createBoolEncoder[bool, uint8]()

	// integer type
	UintEncoder   = createNumberEncoder[uint, int64]()
	IntEncoder    = createNumberEncoder[int, int64]()
	Uint8Encoder  = createNumberEncoder[uint8, int64]()
	Int8Encoder   = createNumberEncoder[int8, int64]()
	Uint16Encoder = createNumberEncoder[uint16, int64]()
	Int16Encoder  = createNumberEncoder[int16, int64]()
	Uint32Encoder = createNumberEncoder[uint32, int64]()
	Int32Encoder  = createNumberEncoder[int32, int64]()
	Uint64Encoder = createNumberEncoder[uint64, int64]()
	Int64Encoder  = createEncoder[int64]()

	// float types
	Float32Encoder = createNumberEncoder[float32, float64]()
	Float64Encoder = createEncoder[float64]()

	// native go string to String
	GoStringEncoder = createGoStringEncoder()

	// built-in types
	StringEncoder             = createEncoder[String]()
	Vector2Encoder            = createEncoder[Vector2]()
	Vector2iEncoder           = createEncoder[Vector2i]()
	Rect2Encoder              = createEncoder[Rect2]()
	Rect2iEncoder             = createEncoder[Rect2i]()
	Vector3Encoder            = createReferenceEncoder[Vector3]()
	Vector3iEncoder           = createReferenceEncoder[Vector3i]()
	Transform2DEncoder        = createEncoder[Transform2D]()
	Vector4Encoder            = createReferenceEncoder[Vector4]()
	Vector4iEncoder           = createReferenceEncoder[Vector4i]()
	PlaneEncoder              = createReferenceEncoder[Plane]()
	QuaternionEncoder         = createEncoder[Quaternion]()
	AABBEncoder               = createReferenceEncoder[AABB]()
	BasisEncoder              = createReferenceEncoder[Basis]()
	Transform3DEncoder        = createReferenceEncoder[Transform3D]()
	ProjectionEncoder         = createReferenceEncoder[Projection]()
	ColorEncoder              = createReferenceEncoder[Color]()
	StringNameEncoder         = createEncoder[StringName]()
	NodePathEncoder           = createEncoder[NodePath]()
	RIDEncoder                = createEncoder[RID]()
	CallableEncoder           = createEncoder[Callable]()
	SignalEncoder             = createEncoder[Signal]()
	DictionaryEncoder         = createEncoder[Dictionary]()
	ArrayEncoder              = createEncoder[Array]()
	PackedByteArrayEncoder    = createEncoder[PackedByteArray]()
	PackedInt32ArrayEncoder   = createEncoder[PackedInt32Array]()
	PackedInt64ArrayEncoder   = createEncoder[PackedInt64Array]()
	PackedFloat32ArrayEncoder = createEncoder[PackedFloat32Array]()
	PackedFloat64ArrayEncoder = createEncoder[PackedFloat64Array]()
	PackedStringArrayEncoder  = createEncoder[PackedStringArray]()
	PackedVector2ArrayEncoder = createEncoder[PackedVector2Array]()
	PackedVector3ArrayEncoder = createEncoder[PackedVector3Array]()
	PackedColorArrayEncoder   = createEncoder[PackedColorArray]()
	VariantEncoder            = createReferenceEncoder[Variant]()
)

func createBoolEncoder[T ~bool, E ~uint8]() ArgumentEncoder[T, E] {
	return ArgumentEncoder[T, E]{
		Decode: func(v unsafe.Pointer) T {
			return (*(*E)(v)) > 0
		},
		Encode: func(in T, out unsafe.Pointer) {
			tOut := (*E)(out)
			if in {
				*tOut = 1
			} else {
				*tOut = 0
			}

		},
		EncodeArg: func(in T) E {
			if in {
				return 1
			} else {
				return 0
			}
		},
	}
}

type Number interface {
	constraints.Integer | constraints.Float
}

func createNumberEncoder[T Number, E Number]() ArgumentEncoder[T, E] {
	return ArgumentEncoder[T, E]{
		Decode: func(v unsafe.Pointer) T {
			return (T)(*(*E)(v))
		},
		Encode: func(in T, out unsafe.Pointer) {
			tOut := (*E)(out)
			*tOut = (E)(in)
		},
		EncodeArg: func(in T) E {
			return (E)(in)
		},
	}
}

func createGoStringEncoder() ArgumentEncoder[string, String] {
	return ArgumentEncoder[string, String]{
		Decode: func(v unsafe.Pointer) string {
			gdnstr := (*String)(v)
			return gdnstr.ToAscii()
		},
		Encode: func(in string, out unsafe.Pointer) {
			tOut := (*String)(out)
			*tOut = NewStringWithLatin1Chars(in)
		},
		EncodeArg: func(in string) String {
			return NewStringWithLatin1Chars(in)
		},
	}
}

// createPointerEncoder reiles on the fact that the opaque data is the first field in the struct:
// a := &Example{}
// iface := (interface{})(a)
// fmt.Printf("%p\n", a)
// fmt.Printf("%p\n", a.ptr())
// fmt.Printf("%p\n", iface)
// all 3 addresses should be the same
func createEncoder[T any]() ArgumentEncoder[T, unsafe.Pointer] {
	return ArgumentEncoder[T, unsafe.Pointer]{
		Decode: func(v unsafe.Pointer) T {
			return *(*T)(v)
		},
		Encode: func(in T, out unsafe.Pointer) {
			tOut := (*T)(out)
			*tOut = in
		},
		EncodeArg: nil,
	}
}

func createReferenceEncoder[T any]() ArgumentReferenceEncoder[T, unsafe.Pointer] {
	return ArgumentReferenceEncoder[T, unsafe.Pointer]{
		Decode: func(v unsafe.Pointer) T {
			return *(*T)(v)
		},
		Encode: func(in *T, out unsafe.Pointer) {
			tOut := (**T)(out)
			*tOut = in
		},
		EncodeArg: nil,
	}
}
