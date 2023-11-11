package core

import (
	"reflect"
	"strings"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassinit"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// ReflectTypeToGDExtensionVariantType returns the correct GDExtensionVariantType given a reflect.Type.
func ReflectTypeToGDExtensionVariantType(t reflect.Type) GDExtensionVariantType {
	if t == nil {
		return GDEXTENSION_VARIANT_TYPE_NIL
	}

	switch t.Kind() {
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
		zero := reflect.Zero(t)
		inst := zero.Interface()
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
		case Callable:
			return GDEXTENSION_VARIANT_TYPE_CALLABLE
		case Signal:
			return GDEXTENSION_VARIANT_TYPE_SIGNAL
		case Dictionary:
			return GDEXTENSION_VARIANT_TYPE_DICTIONARY
		case Array:
			return GDEXTENSION_VARIANT_TYPE_ARRAY
		case PackedByteArray:
			return GDEXTENSION_VARIANT_TYPE_PACKED_BYTE_ARRAY
		case PackedInt32Array:
			return GDEXTENSION_VARIANT_TYPE_PACKED_INT32_ARRAY
		case PackedInt64Array:
			return GDEXTENSION_VARIANT_TYPE_PACKED_INT64_ARRAY
		case PackedFloat32Array:
			return GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY
		case PackedFloat64Array:
			return GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT64_ARRAY
		case PackedStringArray:
			return GDEXTENSION_VARIANT_TYPE_PACKED_STRING_ARRAY
		case PackedVector2Array:
			return GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR2_ARRAY
		case PackedVector3Array:
			return GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR3_ARRAY
		case PackedColorArray:
			return GDEXTENSION_VARIANT_TYPE_PACKED_COLOR_ARRAY
		case Variant:
			return GDEXTENSION_VARIANT_TYPE_VARIANT_MAX
		default:
			log.Panic("unhandled go struct", zap.Any("type", t))
		}
	case reflect.Pointer:
		zero := reflect.Zero(t)
		inst := zero.Interface()
		if _, ok := inst.(Ref); ok {
			log.Debug("detected Ref (type assertion) as GDEXTENSION_VARIANT_TYPE_OBJECT")
			return GDEXTENSION_VARIANT_TYPE_OBJECT
		}
		if _, ok := inst.(GDClass); ok {
			log.Debug("detected GDClass as GDEXTENSION_VARIANT_TYPE_OBJECT")
			return GDEXTENSION_VARIANT_TYPE_OBJECT
		}
		if t.Implements(gdClassType) {
			log.Debug("detected GDClass as GDEXTENSION_VARIANT_TYPE_OBJECT")
			return GDEXTENSION_VARIANT_TYPE_OBJECT
		}
		if _, ok := inst.(Variant); ok {
			log.Debug("detected Variant as GDEXTENSION_VARIANT_TYPE_VARIANT_MAX")
			return GDEXTENSION_VARIANT_TYPE_VARIANT_MAX
		}
		log.Panic("unhandled go pointer", zap.Any("type", t))
	case reflect.Interface:
		inst := reflect.Zero(t).Interface()
		switch inst.(type) {
		case Object:
			return GDEXTENSION_VARIANT_TYPE_OBJECT
		case GDExtensionClass:
			return GDEXTENSION_VARIANT_TYPE_OBJECT
		default:
			if t.Implements(refType) {
				return GDEXTENSION_VARIANT_TYPE_OBJECT
			}
			tn := t.Name()
			if _, ok := Internal.GDRegisteredGDClasses.Get(tn); ok {
				return GDEXTENSION_VARIANT_TYPE_OBJECT
			}
			if _, ok := GDNativeConstructors.Get(tn); ok {
				return GDEXTENSION_VARIANT_TYPE_OBJECT
			}
			if strings.HasPrefix(tn, "Ref") {
				if _, ok := GDClassRefConstructors.Get(tn[3:]); ok {
					return GDEXTENSION_VARIANT_TYPE_OBJECT
				}
			}
			log.Panic("unhandled go interface",
				zap.Any("type", t),
			)
		}
	case reflect.Map, reflect.Chan, reflect.Func, reflect.Uintptr, reflect.Complex64, reflect.Complex128, reflect.UnsafePointer:
		log.Panic("unhandled reflected go kind", zap.Any("type", t))
	default:
		log.Panic("unhandled go kind", zap.Any("type", t))
	}
	return GDEXTENSION_VARIANT_TYPE_VARIANT_MAX
}
