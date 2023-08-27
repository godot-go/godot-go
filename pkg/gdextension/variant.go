package gdextension

// #include <godot/gdextension_interface.h>
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

var (
	variantFromTypeConstructor [GDEXTENSION_VARIANT_TYPE_VARIANT_MAX]GDExtensionVariantFromTypeConstructorFunc
	variantToTypeConstructor   [GDEXTENSION_VARIANT_TYPE_VARIANT_MAX]GDExtensionTypeFromVariantConstructorFunc
)

func variantInitBindings() {
	log.Debug("variantInitBindings called")
	for i := GDExtensionVariantType(1); i < GDEXTENSION_VARIANT_TYPE_VARIANT_MAX; i++ {
		variantFromTypeConstructor[i] = CallFunc_GDExtensionInterfaceGetVariantFromTypeConstructor(i)
		variantToTypeConstructor[i] = CallFunc_GDExtensionInterfaceGetVariantToTypeConstructor(i)
	}

	builtinClassesInitBindings()
}

func ReflectTypeToGDExtensionVariantType(t reflect.Type) GDExtensionVariantType {
	var (
		ik reflect.Kind
		it reflect.Type
	)

	if t == nil {
		return GDEXTENSION_VARIANT_TYPE_NIL
	}

	ik = t.Kind()

	if ik == reflect.Pointer {
		it = t.Elem()
		ik = it.Kind()
	} else {
		it = t
	}

	switch ik {
	case reflect.Bool:
		return GDEXTENSION_VARIANT_TYPE_BOOL
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return GDEXTENSION_VARIANT_TYPE_INT
	case reflect.Float32, reflect.Float64:
		return GDEXTENSION_VARIANT_TYPE_FLOAT
	case reflect.Array, reflect.Slice:
		return GDEXTENSION_VARIANT_TYPE_ARRAY
	case reflect.String:
		return GDEXTENSION_VARIANT_TYPE_STRING
	case reflect.Struct:
		itv := reflect.Zero(it)
		itInst := itv.Interface()
		switch itInst.(type) {
		case String:
			return GDEXTENSION_VARIANT_TYPE_STRING
		case Vector2:
			return GDEXTENSION_VARIANT_TYPE_VECTOR2
		case Vector2i:
			return GDEXTENSION_VARIANT_TYPE_VECTOR2I
		case Rect2:
			return GDEXTENSION_VARIANT_TYPE_RECT2
		case Rect2i:
			return GDEXTENSION_VARIANT_TYPE_RECT2I
		case Vector3:
			return GDEXTENSION_VARIANT_TYPE_VECTOR3
		case Vector3i:
			return GDEXTENSION_VARIANT_TYPE_VECTOR3I
		case Vector4:
			return GDEXTENSION_VARIANT_TYPE_VECTOR4
		case Vector4i:
			return GDEXTENSION_VARIANT_TYPE_VECTOR4I
		case Transform2D:
			return GDEXTENSION_VARIANT_TYPE_TRANSFORM2D
		case Plane:
			return GDEXTENSION_VARIANT_TYPE_PLANE
		case Quaternion:
			return GDEXTENSION_VARIANT_TYPE_QUATERNION
		case AABB:
			return GDEXTENSION_VARIANT_TYPE_AABB
		case Basis:
			return GDEXTENSION_VARIANT_TYPE_BASIS
		case Transform3D:
			return GDEXTENSION_VARIANT_TYPE_TRANSFORM3D
		case Color:
			return GDEXTENSION_VARIANT_TYPE_COLOR
		case StringName:
			return GDEXTENSION_VARIANT_TYPE_STRING_NAME
		case NodePath:
			return GDEXTENSION_VARIANT_TYPE_NODE_PATH
		case RID:
			return GDEXTENSION_VARIANT_TYPE_RID
		case Object:
			return GDEXTENSION_VARIANT_TYPE_OBJECT
		case Callable:
			return GDEXTENSION_VARIANT_TYPE_CALLABLE
		case Signal:
			return GDEXTENSION_VARIANT_TYPE_SIGNAL
		case Dictionary:
			return GDEXTENSION_VARIANT_TYPE_DICTIONARY
		case Array:
			return GDEXTENSION_VARIANT_TYPE_ARRAY
		// case ByteArray:
		// 	return GDEXTENSION_VARIANT_TYPE_PACKED_BYTE_ARRAY
		// case Int32Array:
		// 	return GDEXTENSION_VARIANT_TYPE_PACKED_INT32_ARRAY
		// case Int64Array:
		// 	return GDEXTENSION_VARIANT_TYPE_PACKED_INT64_ARRAY
		// case Float32Array:
		// 	return GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY
		// case Float64Array:
		// 	return GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT64_ARRAY
		case PackedStringArray:
			return GDEXTENSION_VARIANT_TYPE_PACKED_STRING_ARRAY
		case PackedVector2Array:
			return GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR2_ARRAY
		case PackedVector3Array:
			return GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR3_ARRAY
		case PackedColorArray:
			return GDEXTENSION_VARIANT_TYPE_PACKED_COLOR_ARRAY
		default:
			if _, ok := itInst.(Ref); ok {
				log.Debug("detected Ref as GDEXTENSION_VARIANT_TYPE_OBJECT")
				return GDEXTENSION_VARIANT_TYPE_OBJECT
			}

			if _, ok := itInst.(GDClass); ok {
				log.Debug("detected GDClass as GDEXTENSION_VARIANT_TYPE_OBJECT")
				return GDEXTENSION_VARIANT_TYPE_OBJECT
			}

			if _, ok := itInst.(GDExtensionClass); ok {
				log.Debug("detected GDExtensionClass as GDEXTENSION_VARIANT_TYPE_OBJECT")
				return GDEXTENSION_VARIANT_TYPE_OBJECT
			}

			log.Panic("unhandled go struct", zap.Any("type", t), zap.Any("inner_type", it))
		}
	case reflect.Interface:
		tn := it.Name()

		if len(tn) > 0 {
			log.Debug("check object", zap.String("name", tn))

			if _, ok := gdRegisteredGDClasses.Get(tn); ok {
				return GDEXTENSION_VARIANT_TYPE_OBJECT
			}

			if _, ok := gdNativeConstructors.Get(tn); ok {
				return GDEXTENSION_VARIANT_TYPE_OBJECT
			}
		}

		log.Panic("unhandled go interface",
			zap.Any("type", t),
			zap.Any("inner_type", it),
		)

	case reflect.Map, reflect.Chan, reflect.Func, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.UnsafePointer:
		log.Panic("unhandled reflected go kind", zap.Any("type", t))
	default:
		log.Panic("unhandled go kind", zap.Any("type", t))
	}

	return GDEXTENSION_VARIANT_TYPE_VARIANT_MAX
}

func GDExtensionTypePtrFromReflectValue(value reflect.Value, rOut GDExtensionTypePtr) {
	k := value.Kind()

	switch k {
	case reflect.Bool:
		if value.Bool() {
			*(*C.GDExtensionBool)(rOut) = 1
		} else {
			*(*C.GDExtensionBool)(rOut) = 0
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v := (C.GDExtensionInt)(value.Int())
		*(*C.GDExtensionInt)(rOut) = v
	case reflect.Float32, reflect.Float64:
		v := value.Float()
		*(*float64)(rOut) = v
	case reflect.String:
		v := value.String()
		GDExtensionStringPtrWithUtf8Chars((GDExtensionStringPtr)(rOut), v)
	case reflect.Interface:
		log.Debug("returing interface",
			zap.String("name", value.Type().Name()),
		)
		switch inst := value.Interface().(type) {
		case Vector2:
			*(*C.GDExtensionTypePtr)(rOut) = (C.GDExtensionTypePtr)(inst.ptr())
		case Vector3:
			*(*C.GDExtensionTypePtr)(rOut) = (C.GDExtensionTypePtr)(inst.ptr())
		case Vector4:
			*(*C.GDExtensionTypePtr)(rOut) = (C.GDExtensionTypePtr)(inst.ptr())
		case Object:
			*(*C.GDExtensionObjectPtr)(rOut) = (C.GDExtensionObjectPtr)(unsafe.Pointer(inst.GetGodotObjectOwner()))
		default:
			log.Panic("unhandled go interface to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k))
		}
	default:
		log.Panic("unhandled native value to GDExtensionTypePtr",
			zap.Any("value", value),
			zap.Any("kind", k))
	}
}

func GDExtensionVariantPtrFromReflectValue(value reflect.Value, rOut GDExtensionVariantPtr) {
	log.Debug("GDExtensionVariantPtrFromReflectValue called",
		zap.String("type", value.Type().Name()),
	)
}

func NewVariantNil() Variant {
	ret := Variant{}
	ptr := (GDExtensionVariantPtr)(ret.ptr())
	GDExtensionVariantPtrWithNil(ptr)
	return ret
}

func GDExtensionVariantPtrWithNil(rOut GDExtensionVariantPtr) {
	CallFunc_GDExtensionInterfaceVariantNewNil(
		(GDExtensionUninitializedVariantPtr)(rOut),
	)
}

func NewVariantNativeCopy(native_ptr GDExtensionConstVariantPtr) Variant {
	ret := Variant{}
	CallFunc_GDExtensionInterfaceVariantNewCopy((GDExtensionUninitializedVariantPtr)(ret.ptr()), native_ptr)
	return ret
}

func NewVariantCopy(other Variant) Variant {
	ret := Variant{}
	CallFunc_GDExtensionInterfaceVariantNewCopy((GDExtensionUninitializedVariantPtr)(ret.ptr()), (GDExtensionConstVariantPtr)(other.ptr()))
	return ret
}

func NewVariantBool(v bool) Variant {
	ret := Variant{}
	ptr := (GDExtensionVariantPtr)(ret.ptr())
	GDExtensionVariantPtrFromBool(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromBool(v bool, rOut GDExtensionVariantPtr) {
	// MAKE_PTRARGCONV(bool, uint8_t);
	var encoded uint8
	if v {
		encoded = 1
	}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_BOOL]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(rOut),
		(GDExtensionTypePtr)(&encoded),
	)
}

func (c *Variant) ToBool() bool {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_BOOL]
	var v uint8
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		(GDExtensionVariantPtr)(c.ptr()),
	)
	return v != 0
}

func NewVariantInt64(v int64) Variant {
	ret := Variant{}
	ptr := (GDExtensionVariantPtr)(ret.ptr())
	GDExtensionVariantPtrFromInt64(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromInt64(v int64, rOut GDExtensionVariantPtr) {
	// MAKE_PTRARGCONV(uint8_t, int64_t);
	// MAKE_PTRARGCONV(int8_t, int64_t);
	// MAKE_PTRARGCONV(uint16_t, int64_t);
	// MAKE_PTRARGCONV(int16_t, int64_t);
	// MAKE_PTRARGCONV(uint32_t, int64_t);
	// MAKE_PTRARGCONV(int32_t, int64_t);
	// MAKE_PTRARG(int64_t);
	// MAKE_PTRARG(uint64_t);
	var encoded int64
	encoded = v
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(rOut),
		(GDExtensionTypePtr)(&encoded),
	)
}

func (c *Variant) ToInt64() int64 {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]
	var v int64
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		(GDExtensionVariantPtr)(c.ptr()),
	)
	return v
}

func NewVariantFloat64(v float64) Variant {
	ret := Variant{}
	ptr := (GDExtensionVariantPtr)(ret.ptr())
	GDExtensionVariantPtrFromFloat64(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromFloat64(v float64, rOut GDExtensionVariantPtr) {
	// MAKE_PTRARGCONV(float, double);
	// MAKE_PTRARG(double);
	var encoded float64
	encoded = v
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_FLOAT]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(ret.ptr()),
		(GDExtensionTypePtr)(&encoded),
	)
}

func (c *Variant) ToFloat64() float64 {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_FLOAT]
	var v float64
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&v),
		(GDExtensionVariantPtr)(c.ptr()),
	)
	return v
}

func NewVariantString(v String) Variant {
	ret := Variant{}
	ptr := (GDExtensionVariantPtr)(ret.ptr())
	GDExtensionVariantPtrFromString(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromString(v String, rOut GDExtensionVariantPtr) {
	// MAKE_PTRARG(String);
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(rOut),
		(GDExtensionTypePtr)(v.ptr()),
	)
}

func (c *Variant) ToString() String {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING]
	var v String
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.ptr()),
		(GDExtensionVariantPtr)(c.ptr()),
	)
	return v
}

func NewVariantStringName(v StringName) Variant {
	ret := Variant{}
	ptr := (GDExtensionVariantPtr)(ret.ptr())
	GDExtensionVariantPtrFromStringName(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromStringName(v StringName, rOut GDExtensionVariantPtr) {
	// MAKE_PTRARG(String);
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING_NAME]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(rOut),
		(GDExtensionTypePtr)(v.ptr()),
	)
}

func (c *Variant) ToStringName() StringName {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING_NAME]
	var (
		v StringName
	)
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.ptr()),
		(GDExtensionVariantPtr)(c.ptr()),
	)
	return v
}

func NewVariantObject(v Object) Variant {
	ret := Variant{}
	ptr := (GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr()))
	GDExtensionVariantPtrFromObject(v, ptr)
	return ret
}

func GDExtensionVariantPtrFromObject(v Object, rOut GDExtensionVariantPtr) {
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	owner := v.GetGodotObjectOwner()
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(rOut),
		(GDExtensionTypePtr)(&owner),
	)
}

func (c *Variant) ToObject() Object {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	var v Object
	owner := v.GetGodotObjectOwner()
	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(&owner),
		(GDExtensionVariantPtr)(c.ptr()),
	)
	return v
}

// func NewVariantWrapped(v Wrapped) Variant {
// 	ret := Variant{}
// 	ptr := (GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr()))
// 	GDExtensionVariantPtrFromWrapped(v, ptr)
// 	return ret
// }

// func GDExtensionVariantPtrFromWrapped(v Wrapped, rOut GDExtensionVariantPtr) {
// 	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
// 	owner := v.GetGodotObjectOwner()
// 	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
// 		(GDExtensionVariantFromTypeConstructorFunc)(fn),
// 		(GDExtensionUninitializedVariantPtr)(rOut),
// 		(GDExtensionTypePtr)(&owner),
// 	)
// }

// func (c *Variant) ToWrapped() Wrapped {
// 	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
// 	var v Wrapped
// 	owner := v.GetGodotObjectOwner()
// 	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
// 		(GDExtensionTypeFromVariantConstructorFunc)(fn),
// 		(GDExtensionUninitializedTypePtr)(&owner),
// 		(GDExtensionVariantPtr)(c.ptr()),
// 	)
// 	return v
// }

func NewVariantVector2(v Vector2) Variant {
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR2]
	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(ret.ptr()),
		(GDExtensionTypePtr)(v.ptr()),
	)
	return ret
}

func (c *Variant) ToVector2() Vector2 {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR2]

	var (
		v Vector2
	)

	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.ptr()),
		(GDExtensionVariantPtr)(c.ptr()),
	)

	return v
}

func NewVariantVector3(v Vector3) Variant {
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR3]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(ret.ptr()),
		(GDExtensionTypePtr)(v.ptr()),
	)

	return ret
}

func (c *Variant) ToVector3() Vector3 {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR3]

	var (
		v Vector3
	)

	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.ptr()),
		(GDExtensionVariantPtr)(c.ptr()),
	)

	return v
}

func NewVariantVector4(v Vector4) Variant {
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR4]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(ret.ptr()),
		(GDExtensionTypePtr)(v.ptr()),
	)

	return ret
}

func (c *Variant) ToVector4() Vector4 {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR4]

	var (
		v Vector4
	)

	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(v.ptr()),
		(GDExtensionVariantPtr)(c.ptr()),
	)

	return v
}

func NewVariantArray(v Array) Variant {
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_ARRAY]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(ret.ptr()),
		(GDExtensionTypePtr)(v.ptr()),
	)

	return ret
}

func (c *Variant) ToArray() Array {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_ARRAY]

	var (
		arr Array
	)

	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(arr.ptr()),
		(GDExtensionVariantPtr)(c.ptr()),
	)

	return arr
}

func NewVariantDictionary(v Dictionary) Variant {
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_DICTIONARY]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionUninitializedVariantPtr)(ret.ptr()),
		(GDExtensionTypePtr)(v.ptr()),
	)

	return ret
}

func (c *Variant) ToDictionary() Dictionary {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_DICTIONARY]

	var (
		dict Dictionary
	)

	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionUninitializedTypePtr)(dict.ptr()),
		(GDExtensionVariantPtr)(c.ptr()),
	)

	return dict
}

type Variant struct {
	// opaque size should be taken from extension_api.json
	opaque [24]uint8
}

func (c *Variant) ptr() GDExtensionConstVariantPtr {
	return (GDExtensionConstVariantPtr)(&c.opaque)
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

	sn := NewStringNameWithUtf8Chars(method)
	defer sn.Destroy()

	callArgs = AllocCopyVariantPtrSliceAsGDExtensionVariantPtrPtr(args)

	callArgCount := len(args)

	var err GDExtensionCallError

	CallFunc_GDExtensionInterfaceVariantCall(
		(GDExtensionVariantPtr)(c.ptr()),
		(GDExtensionConstStringNamePtr)(sn.ptr()),
		callArgs,
		(GDExtensionInt)(callArgCount),
		(GDExtensionUninitializedVariantPtr)(r_ret.ptr()),
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

	sn := NewStringNameWithUtf8Chars(method)
	defer sn.Destroy()

	callArgs = AllocCopyVariantPtrSliceAsGDExtensionVariantPtrPtr(args)

	callArgCount := len(args)

	var err GDExtensionCallError

	CallFunc_GDExtensionInterfaceVariantCallStatic(
		vt,
		(GDExtensionConstStringNamePtr)(sn.ptr()),
		callArgs,
		(GDExtensionInt)(callArgCount),
		(GDExtensionUninitializedVariantPtr)(r_ret.ptr()),
		&err,
	)

	if err.Ok() {
		return nil
	}

	return err
}

func (c *Variant) GetType() GDExtensionVariantType {
	return CallFunc_GDExtensionInterfaceVariantGetType((GDExtensionConstVariantPtr)(c.ptr()))
}

func (c *Variant) Clear() {
	if needsDeinit[(int)(c.GetType())] {
		CallFunc_GDExtensionInterfaceVariantDestroy((GDExtensionVariantPtr)(c.ptr()))
	}
	CallFunc_GDExtensionInterfaceVariantNewNil((GDExtensionUninitializedVariantPtr)(c.ptr()))
}

var (
	ErrOutOfBounds = fmt.Errorf("out of bounds")
	ErrInvalid     = fmt.Errorf("invalid")
)

func (c *Variant) Set(key Variant, value Variant) error {
	var valid GDExtensionBool

	CallFunc_GDExtensionInterfaceVariantSet(
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
		key.ptr(), value.ptr(), &valid)

	if !BoolEncoder.Decode(unsafe.Pointer(&valid)) {
		return ErrInvalid
	}

	return nil
}

func (c *Variant) SetNamed(name StringName, value Variant) error {
	var valid GDExtensionBool

	CallFunc_GDExtensionInterfaceVariantSetNamed(
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
		(GDExtensionConstStringNamePtr)(unsafe.Pointer(name.ptr())),
		(GDExtensionConstVariantPtr)(unsafe.Pointer(value.ptr())), &valid)

	if !BoolEncoder.Decode(unsafe.Pointer(&valid)) {
		return ErrInvalid
	}

	return nil
}

func (c *Variant) SetIndexed(index int, value Variant) error {
	var valid, oob GDExtensionBool

	CallFunc_GDExtensionInterfaceVariantSetIndexed(
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
		(GDExtensionInt)(index), value.ptr(), &valid, &oob)

	if !BoolEncoder.Decode(unsafe.Pointer(&valid)) {
		return ErrInvalid
	}

	if BoolEncoder.Decode(unsafe.Pointer(&oob)) {
		return ErrOutOfBounds
	}

	return nil
}

func (c *Variant) SetKeyed(key, value Variant) bool {
	var valid GDExtensionBool

	CallFunc_GDExtensionInterfaceVariantSetKeyed(
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
		key.ptr(),
		value.ptr(),
		&valid)

	return BoolEncoder.Decode(unsafe.Pointer(&valid))
}

func (c *Variant) GetIndexed(index int) (Variant, error) {
	var (
		result Variant
		valid  GDExtensionBool
		oob    GDExtensionBool
	)

	CallFunc_GDExtensionInterfaceVariantGetIndexed(
		c.ptr(), (GDExtensionInt)(index), (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(result.ptr())), &valid, &oob)

	if !BoolEncoder.Decode(unsafe.Pointer(&valid)) {
		return result, ErrInvalid
	}

	if BoolEncoder.Decode(unsafe.Pointer(&oob)) {
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
		c.ptr(), key.ptr(), (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(result.ptr())), &valid)

	if !BoolEncoder.Decode(unsafe.Pointer(&valid)) {
		return result, ErrInvalid
	}

	return result, nil
}

func (c *Variant) Destroy() {
	CallFunc_GDExtensionInterfaceVariantDestroy((GDExtensionVariantPtr)(c.ptr()))
}

func (c *Variant) ToReflectValue(inType GDExtensionVariantType, outType reflect.Type) reflect.Value {
	switch inType {
	case GDEXTENSION_VARIANT_TYPE_NIL:
		return reflect.ValueOf(nil)
	case GDEXTENSION_VARIANT_TYPE_BOOL:
		return reflect.ValueOf(c.ToBool())
	case GDEXTENSION_VARIANT_TYPE_INT:
		v := c.ToInt64()
		switch outType.Kind() {
		case reflect.Int:
			return reflect.ValueOf((int)(v))
		case reflect.Int8:
			return reflect.ValueOf((int8)(v))
		case reflect.Int16:
			return reflect.ValueOf((int16)(v))
		case reflect.Int32:
			return reflect.ValueOf((int32)(v))
		case reflect.Int64:
			return reflect.ValueOf((int64)(v))
		case reflect.Uint:
			return reflect.ValueOf((uint)(v))
		case reflect.Uint8:
			return reflect.ValueOf((uint8)(v))
		case reflect.Uint16:
			return reflect.ValueOf((uint16)(v))
		case reflect.Uint32:
			return reflect.ValueOf((uint32)(v))
		case reflect.Uint64:
			return reflect.ValueOf((uint64)(v))
		}
	case GDEXTENSION_VARIANT_TYPE_FLOAT:
		v := c.ToFloat64()
		switch outType.Kind() {
		case reflect.Float32:
			return reflect.ValueOf((float32)(v))
		case reflect.Float64:
			return reflect.ValueOf((float64)(v))
		}
	case GDEXTENSION_VARIANT_TYPE_STRING:
		gdstr := c.ToString()
		str := gdstr.ToAscii()
		return reflect.ValueOf(str)
	case GDEXTENSION_VARIANT_TYPE_VECTOR2:
		return reflect.ValueOf(c.ToVector2())
	case GDEXTENSION_VARIANT_TYPE_VECTOR3:
		return reflect.ValueOf(c.ToVector3())
	case GDEXTENSION_VARIANT_TYPE_VECTOR4:
		return reflect.ValueOf(c.ToVector4())
	case GDEXTENSION_VARIANT_TYPE_OBJECT:
		obj := c.ToObject()
		return reflect.ValueOf(obj)
	}

	log.Panic("unhandled GDExtension type", zap.Any("gdn_type", inType))

	return reflect.Zero(outType)
}
