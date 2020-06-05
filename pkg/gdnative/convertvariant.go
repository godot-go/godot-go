package gdnative

import (
	"fmt"
	"reflect"

	"github.com/pcting/godot-go/pkg/log"
)

// VariantToGoType will check the given variant type and convert it to its
// actual type. The value is returned as a reflect.Value.
func VariantToGoType(variant Variant) reflect.Value {
	v := variant.GetType()
	switch v {
	case GODOT_VARIANT_TYPE_NIL:
		return reflect.ValueOf(nil)
	case GODOT_VARIANT_TYPE_BOOL:
		return reflect.ValueOf(variant.AsBool())
	case GODOT_VARIANT_TYPE_INT:
		return reflect.ValueOf(variant.AsInt())
	case GODOT_VARIANT_TYPE_REAL:
		return reflect.ValueOf(variant.AsReal())
	case GODOT_VARIANT_TYPE_STRING:
		return reflect.ValueOf(variant.AsString())
	case GODOT_VARIANT_TYPE_VECTOR2:
		return reflect.ValueOf(variant.AsVector2())
	case GODOT_VARIANT_TYPE_RECT2:
		return reflect.ValueOf(variant.AsRect2())
	case GODOT_VARIANT_TYPE_VECTOR3:
		return reflect.ValueOf(variant.AsVector3())
	case GODOT_VARIANT_TYPE_TRANSFORM2D:
		return reflect.ValueOf(variant.AsTransform2D())
	case GODOT_VARIANT_TYPE_PLANE:
		return reflect.ValueOf(variant.AsPlane())
	case GODOT_VARIANT_TYPE_QUAT:
		return reflect.ValueOf(variant.AsQuat())
	case GODOT_VARIANT_TYPE_AABB:
		return reflect.ValueOf(variant.AsAABB())
	case GODOT_VARIANT_TYPE_BASIS:
		return reflect.ValueOf(variant.AsBasis())
	case GODOT_VARIANT_TYPE_TRANSFORM:
		return reflect.ValueOf(variant.AsTransform())
	case GODOT_VARIANT_TYPE_COLOR:
		return reflect.ValueOf(variant.AsColor())
	case GODOT_VARIANT_TYPE_NODE_PATH:
		return reflect.ValueOf(variant.AsNodePath())
	case GODOT_VARIANT_TYPE_RID:
		return reflect.ValueOf(variant.AsRID())
	case GODOT_VARIANT_TYPE_OBJECT:
		return reflect.ValueOf(variant.AsObject())
	case GODOT_VARIANT_TYPE_DICTIONARY:
		return reflect.ValueOf(variant.AsDictionary())
	case GODOT_VARIANT_TYPE_ARRAY:
		return reflect.ValueOf(variant.AsArray())
	case GODOT_VARIANT_TYPE_POOL_BYTE_ARRAY:
		return reflect.ValueOf(variant.AsPoolByteArray())
	case GODOT_VARIANT_TYPE_POOL_INT_ARRAY:
		return reflect.ValueOf(variant.AsPoolIntArray())
	case GODOT_VARIANT_TYPE_POOL_REAL_ARRAY:
		return reflect.ValueOf(variant.AsPoolRealArray())
	case GODOT_VARIANT_TYPE_POOL_STRING_ARRAY:
		return reflect.ValueOf(variant.AsPoolStringArray())
	case GODOT_VARIANT_TYPE_POOL_VECTOR2_ARRAY:
		return reflect.ValueOf(variant.AsPoolVector2Array())
	case GODOT_VARIANT_TYPE_POOL_VECTOR3_ARRAY:
		return reflect.ValueOf(variant.AsPoolVector3Array())
	case GODOT_VARIANT_TYPE_POOL_COLOR_ARRAY:
		return reflect.ValueOf(variant.AsPoolColorArray())
	}
	log.WithField("type", fmt.Sprintf("%d", variant.GetType())).Panic("variant to native built-in type version unhandled")
	return reflect.ValueOf(nil)
}

// GoTypeToVariant will check the given Go type and convert it to its
// Variant type. The value is returned as a Variant.
func GoTypeToVariant(value reflect.Value) Variant {
	if !value.IsValid() {
		return NewVariantNil()
	}

	valueInterface := value.Interface()
	switch v := valueInterface.(type) {
	case bool:
		return NewVariantBool(v)
	case int:
		return NewVariantInt(int64(v))
	case int16:
		return NewVariantInt(int64(v))
	case int32:
		return NewVariantInt(int64(v))
	case int64:
		return NewVariantInt(v)
	case float32:
		return NewVariantReal(float64(v))
	case float64:
		return NewVariantReal(v)
	case string:
		log.
			WithField("value", fmt.Sprintf("%v", value)).
			Panic("unable to handle native go string. please wrap string in a Godot String with gdnative.NewStringFromGoString")
	case String:
		return NewVariantString(v)
	case Vector2:
		return NewVariantVector2(v)
	case Rect2:
		return NewVariantRect2(v)
	case Vector3:
		return NewVariantVector3(v)
	case Transform2D:
		return NewVariantTransform2D(v)
	case Plane:
		return NewVariantPlane(v)
	case Quat:
		return NewVariantQuat(v)
	case AABB:
		return NewVariantAABB(v)
	case Basis:
		return NewVariantBasis(v)
	case Transform:
		return NewVariantTransform(v)
	case Color:
		return NewVariantColor(v)
	case NodePath:
		return NewVariantNodePath(v)
	case RID:
		return NewVariantRID(v)
	case *GodotObject:
		return NewVariantObject(v)
	case Dictionary:
		return NewVariantDictionary(v)
	case Array:
		return NewVariantArray(v)
	case PoolByteArray:
		return NewVariantPoolByteArray(v)
	case PoolIntArray:
		return NewVariantPoolIntArray(v)
	case PoolRealArray:
		return NewVariantPoolRealArray(v)
	case PoolStringArray:
		return NewVariantPoolStringArray(v)
	case PoolVector2Array:
		return NewVariantPoolVector2Array(v)
	case PoolVector3Array:
		return NewVariantPoolVector3Array(v)
	case PoolColorArray:
		return NewVariantPoolColorArray(v)
	}
	log.WithField("value", fmt.Sprintf("%v", value)).Panic("value not handled")
	return NewVariantNil()
}
