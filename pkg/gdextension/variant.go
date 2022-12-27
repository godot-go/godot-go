package gdextension

// #include <godot/gdextension_interface.h>
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdnative"
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
		variantFromTypeConstructor[i] = GDExtensionInterface_get_variant_from_type_constructor(internal.gdnInterface, i)
		variantToTypeConstructor[i] = GDExtensionInterface_get_variant_to_type_constructor(internal.gdnInterface, i)
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

	if ik == reflect.Ptr {
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
		inst := reflect.Zero(it).Interface()
		switch inst.(type) {
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
			inst := reflect.New(it).Interface()

			log.Debug("reflected inst", zap.Any("inst", inst), zap.Any("type", reflect.TypeOf(inst)))

			if classInst, ok := inst.(GDClass); ok {
				if _, ok := gdRegisteredGDClasses.Get(classInst.GetClassName()); ok {
					return GDEXTENSION_VARIANT_TYPE_OBJECT
				}
			}

			if nativeClassInst, ok := inst.(GDExtensionClass); ok {
				if _, ok := gdNativeConstructors.Get(nativeClassInst.GetClassName()); ok {
					return GDEXTENSION_VARIANT_TYPE_OBJECT
				}
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

func NewVariantFromReflectValue(value reflect.Value) Variant {
	k := value.Kind()

	switch k {
	case reflect.Bool:
		return NewVariantBool(value.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return NewVariantInt64(value.Int())
	case reflect.Float32, reflect.Float64:
		return NewVariantFloat64(value.Float())
	case reflect.String:
		gdnStr := NewStringWithLatin1Chars(value.String())
		defer gdnStr.Destroy()
		return NewVariantString(gdnStr)
	case reflect.Ptr:
		iv := value.Elem()
		ik := iv.Kind()
		switch ik {
		case reflect.Struct:
			switch inst := value.Interface().(type) {
			case *Vector4:
				return NewVariantVector4(*inst)
			case *Array:
				return NewVariantArray(*inst)
			case *Dictionary:
				return NewVariantDictionary(*inst)
			case Wrapped:
				if nativeClassInst, ok := inst.(GDExtensionClass); ok {
					log.Debug("create GDExtensionClass variant object", zap.Any("inst", nativeClassInst.GetClassName()))
				} else if classInst, ok := inst.(GDClass); ok {
					log.Debug("create GDClass variant object", zap.Any("inst", classInst.GetClassName()))
				} else {
					log.Debug("create new wrapped variant object", zap.Any("inst", inst))
				}

				return NewVariantWrapped(inst.(Wrapped))

			default:
				log.Panic("unhandled reflected ptr struct to variant",
					zap.Any("value", value),
					zap.Any("kind", k))
			}
		default:
			log.Panic("unhandled reflected ptr value to variant",
				zap.Any("value", value),
				zap.Any("kind", k))
		}
	case reflect.Struct:
		ptr := reflect.New(value.Type())
		ptr.Elem().Set(value)
		switch inst := ptr.Interface().(type) {
		case *Vector2:
			return NewVariantVector2(*inst)
		case *Vector4:
			return NewVariantVector4(*inst)
		case *Array:
			return NewVariantArray(*inst)
		case *Dictionary:
			return NewVariantDictionary(*inst)
		case Wrapped:
			if nativeClassInst, ok := inst.(GDExtensionClass); ok {
				log.Debug("create GDExtensionClass variant object", zap.Any("inst", nativeClassInst.GetClassName()))
			} else if classInst, ok := inst.(GDClass); ok {
				log.Debug("create GDClass variant object", zap.Any("inst", classInst.GetClassName()))
			} else {
				log.Debug("create new wrapped variant object", zap.Any("inst", inst))
			}

			return NewVariantWrapped(inst.(Wrapped))

		default:
			log.Panic("unhandled reflected struct to variant",
				zap.Any("value", value),
				zap.Any("kind", k))
		}
	case reflect.Interface:
		switch w := value.Interface().(type) {
		case Viewport:
			return NewVariantWrapped(w)
		case Wrapped:
			return NewVariantWrapped(w)
		default:
			log.Warn("unhandled reflected interface to variant",
				zap.Any("value", value),
				zap.Any("value_type", value.Type()),
				zap.Any("w", w),
			)
		}
	default:
		log.Panic("unhandled native value to variant",
			zap.Any("value", value),
			zap.Any("kind", k))
	}

	return Variant{}
}

func NewVariantNil() Variant {
	v := Variant{}
	GDExtensionInterface_variant_new_nil(internal.gdnInterface, (GDExtensionVariantPtr)(unsafe.Pointer(v.ptr())))
	return v
}

func NewVariantNativeCopy(native_ptr GDExtensionConstVariantPtr) Variant {
	ret := Variant{}
	GDExtensionInterface_variant_new_copy(internal.gdnInterface, (GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())), native_ptr)
	return ret
}

func NewVariantCopy(other Variant) Variant {
	ret := Variant{}
	GDExtensionInterface_variant_new_copy(internal.gdnInterface, (GDExtensionVariantPtr)(ret.ptr()), (GDExtensionConstVariantPtr)(other.ptr()))
	return ret
}

func NewVariantBool(v bool) Variant {
	// MAKE_PTRARGCONV(bool, uint8_t);
	var encoded uint8
	if v {
		encoded = 1
	}
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_BOOL]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())),
		(GDExtensionTypePtr)(&encoded),
	)
	return ret
}

func (c *Variant) ToBool() bool {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_BOOL]

	var (
		v uint8
	)

	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionTypePtr)(unsafe.Pointer(&v)),
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
	)

	return v != 0
}

func NewVariantInt64(v int64) Variant {
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
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())),
		(GDExtensionTypePtr)(&encoded),
	)
	return ret
}

func (c *Variant) ToInt64() int64 {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_INT]

	var (
		v int64
	)

	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionTypePtr)(unsafe.Pointer(&v)),
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
	)

	return v
}

func NewVariantFloat64(v float64) Variant {
	// MAKE_PTRARGCONV(float, double);
	// MAKE_PTRARG(double);
	var encoded float64
	encoded = v
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_FLOAT]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())),
		(GDExtensionTypePtr)(&encoded),
	)
	return ret
}

func (c *Variant) ToFloat64() float64 {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_FLOAT]

	var (
		v float64
	)

	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionTypePtr)(unsafe.Pointer(&v)),
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
	)

	return v
}

func NewVariantString(v String) Variant {
	// MAKE_PTRARG(String);
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())),
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
	)

	return ret
}

func (c *Variant) ToString() String {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING]

	var (
		v String
	)

	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
	)

	return v
}

func NewVariantStringName(v StringName) Variant {
	// MAKE_PTRARG(String);
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING_NAME]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())),
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
	)

	return ret
}

func (c *Variant) ToStringName() StringName {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_STRING_NAME]

	var (
		v StringName
	)

	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
	)

	return v
}

func NewVariantWrapped(w Wrapped) Variant {
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]
	owner := w.GetGodotObjectOwner()

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())),
		(GDExtensionTypePtr)(unsafe.Pointer(&owner)),
	)

	return ret
}

func (c *Variant) ToWrapped() Wrapped {
	fn := variantToTypeConstructor[GDEXTENSION_VARIANT_TYPE_OBJECT]

	var (
		v Wrapped
	)

	owner := v.GetGodotObjectOwner()

	CallFunc_GDExtensionTypeFromVariantConstructorFunc(
		(GDExtensionTypeFromVariantConstructorFunc)(fn),
		(GDExtensionTypePtr)(unsafe.Pointer(&owner)),
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
	)

	return v
}

func NewVariantVector2(v Vector2) Variant {
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR2]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())),
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
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
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
	)

	return v
}

func NewVariantVector3(v Vector3) Variant {
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR3]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())),
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
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
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
	)

	return v
}

func NewVariantVector4(v Vector4) Variant {
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_VECTOR4]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())),
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
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
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
	)

	return v
}

func NewVariantArray(v Array) Variant {
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_ARRAY]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())),
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
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
		(GDExtensionTypePtr)(unsafe.Pointer(arr.ptr())),
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
	)

	return arr
}

func NewVariantDictionary(v Dictionary) Variant {
	ret := Variant{}
	fn := variantFromTypeConstructor[GDEXTENSION_VARIANT_TYPE_DICTIONARY]

	CallFunc_GDExtensionVariantFromTypeConstructorFunc(
		(GDExtensionVariantFromTypeConstructorFunc)(fn),
		(GDExtensionVariantPtr)(unsafe.Pointer(ret.ptr())),
		(GDExtensionTypePtr)(unsafe.Pointer(v.ptr())),
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
		(GDExtensionTypePtr)(unsafe.Pointer(dict.ptr())),
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
	)

	return dict
}

type Variant struct {
	// opaque size should be taken from extension_api.json
	opaque [24]uint8
}

func (c *Variant) ptr() GDExtensionConstVariantPtr {
	return (GDExtensionConstVariantPtr)(unsafe.Pointer(&c.opaque))
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

	callArgs = AllocCopyVariantPtrSliceAsGDExtensionVariantPtrPtr(args)

	callArgCount := len(args)

	var err GDExtensionCallError

	GDExtensionInterface_variant_call(
		internal.gdnInterface,
		(GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
		(GDExtensionConstStringNamePtr)(unsafe.Pointer(sn.ptr())),
		callArgs,
		(GDExtensionInt)(callArgCount),
		(GDExtensionVariantPtr)(r_ret.ptr()),
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

	callArgs = AllocCopyVariantPtrSliceAsGDExtensionVariantPtrPtr(args)

	callArgCount := len(args)

	var err GDExtensionCallError

	GDExtensionInterface_variant_call_static(
		internal.gdnInterface,
		vt,
		(GDExtensionConstStringNamePtr)(sn.ptr()),
		callArgs,
		(GDExtensionInt)(callArgCount),
		(GDExtensionVariantPtr)(r_ret.ptr()),
		&err,
	)

	if err.Ok() {
		return nil
	}

	return err
}

func (c *Variant) GetType() GDExtensionVariantType {
	return GDExtensionInterface_variant_get_type(internal.gdnInterface, (GDExtensionConstVariantPtr)(c.ptr()))
}

func (c *Variant) Clear() {
	if needsDeinit[(int)(c.GetType())] {
		GDExtensionInterface_variant_destroy(internal.gdnInterface, (GDExtensionVariantPtr)(c.ptr()))
	}
	GDExtensionInterface_variant_new_nil(internal.gdnInterface, (GDExtensionVariantPtr)(c.ptr()))
}

var (
	OutOfBoundsError = fmt.Errorf("out of bounds")
	InvalidError     = fmt.Errorf("invalid")
)

func (c *Variant) Set(key Variant, value Variant) error {
	var valid GDExtensionBool

	GDExtensionInterface_variant_set(
		internal.gdnInterface, (GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
		key.ptr(), value.ptr(), &valid)

	if !BoolEncoder.Decode(unsafe.Pointer(&valid)) {
		return InvalidError
	}

	return nil
}

func (c *Variant) SetNamed(name StringName, value Variant) error {
	var valid GDExtensionBool

	GDExtensionInterface_variant_set_named(
		internal.gdnInterface, (GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
		(GDExtensionConstStringNamePtr)(unsafe.Pointer(name.ptr())),
		(GDExtensionConstVariantPtr)(unsafe.Pointer(value.ptr())), &valid)

	if !BoolEncoder.Decode(unsafe.Pointer(&valid)) {
		return InvalidError
	}

	return nil
}

func (c *Variant) SetIndexed(index int, value Variant) error {
	var valid, oob GDExtensionBool

	GDExtensionInterface_variant_set_indexed(
		internal.gdnInterface, (GDExtensionVariantPtr)(unsafe.Pointer(c.ptr())),
		(GDExtensionInt)(index), value.ptr(), &valid, &oob)

	if !BoolEncoder.Decode(unsafe.Pointer(&valid)) {
		return InvalidError
	}

	if BoolEncoder.Decode(unsafe.Pointer(&oob)) {
		return OutOfBoundsError
	}

	return nil
}

func (c *Variant) SetKeyed(key, value Variant) bool {
	var valid GDExtensionBool

	GDExtensionInterface_variant_set_keyed(internal.gdnInterface,
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

	GDExtensionInterface_variant_get_indexed(internal.gdnInterface,
		c.ptr(), (GDExtensionInt)(index), (GDExtensionVariantPtr)(unsafe.Pointer(result.ptr())), &valid, &oob)

	if !BoolEncoder.Decode(unsafe.Pointer(&valid)) {
		return result, InvalidError
	}

	if BoolEncoder.Decode(unsafe.Pointer(&oob)) {
		return result, OutOfBoundsError
	}

	return result, nil
}

func (c *Variant) GetKeyed(key Variant) (Variant, error) {
	var (
		result Variant
		valid  GDExtensionBool
	)

	GDExtensionInterface_variant_get_keyed(internal.gdnInterface,
		c.ptr(), key.ptr(), (GDExtensionVariantPtr)(unsafe.Pointer(result.ptr())), &valid)

	if !BoolEncoder.Decode(unsafe.Pointer(&valid)) {
		return result, InvalidError
	}

	return result, nil
}

func (c *Variant) Destroy() {
	GDExtensionInterface_variant_destroy(internal.gdnInterface, (GDExtensionVariantPtr)(c.ptr()))
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
		w := c.ToWrapped()
		return reflect.ValueOf(w)
	}

	log.Panic("unhandled GDExtension type", zap.Any("gdn_type", inType))

	return reflect.Zero(outType)
}
