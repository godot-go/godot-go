package core

import (
	"reflect"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassinit"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

func GDExtensionTypePtrFromReflectValue(value reflect.Value, rOut GDExtensionUninitializedTypePtr) {
	k := value.Kind()
	switch k {
	case reflect.Bool:
		BoolEncoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Int:
		IntEncoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Int8:
		Int8Encoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Int16:
		Int16Encoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Int32:
		Int32Encoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Int64:
		Int64Encoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Uint:
		UintEncoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Uint8:
		Uint8Encoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Uint16:
		Uint16Encoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Uint32:
		Uint32Encoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Uint64:
		Uint64Encoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Float32:
		Float32Encoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Float64:
		Float64Encoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.String:
		GoStringUtf8Encoder.EncodeReflectTypePtrArg(value, rOut)
	case reflect.Interface:
		log.Debug("returing interface",
			zap.String("name", value.Type().Name()),
		)
		switch inst := value.Interface().(type) {
		case Object:
			ObjectEncoder.EncodeTypePtrArg(inst, rOut)
			// *(*C.GDExtensionObjectPtr)(rOut) = (C.GDExtensionObjectPtr)(inst.AsGDExtensionObjectPtr())
		default:
			log.Panic("unhandled go interface to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k))
		}
	case reflect.Struct:
		// TODO: add all types supported by variant
		switch inst := value.Interface().(type) {
		case Variant:
			VariantEncoder.EncodeTypePtrArg(inst, rOut)
		case String:
			StringEncoder.EncodeTypePtrArg(inst, rOut)
		case Vector2:
			Vector2Encoder.EncodeTypePtrArg(inst, rOut)
		case Vector2i:
			Vector2iEncoder.EncodeTypePtrArg(inst, rOut)
		case Rect2:
			Rect2Encoder.EncodeTypePtrArg(inst, rOut)
		case Rect2i:
			Rect2iEncoder.EncodeTypePtrArg(inst, rOut)
		case Vector3:
			Vector3Encoder.EncodeTypePtrArg(inst, rOut)
		case Vector3i:
			Vector3iEncoder.EncodeTypePtrArg(inst, rOut)
		case Transform2D:
			Transform2DEncoder.EncodeTypePtrArg(inst, rOut)
		case Vector4:
			Vector4Encoder.EncodeTypePtrArg(inst, rOut)
		case Vector4i:
			Vector4iEncoder.EncodeTypePtrArg(inst, rOut)
		case Plane:
			PlaneEncoder.EncodeTypePtrArg(inst, rOut)
		case Quaternion:
			QuaternionEncoder.EncodeTypePtrArg(inst, rOut)
		case AABB:
			AABBEncoder.EncodeTypePtrArg(inst, rOut)
		case Basis:
			BasisEncoder.EncodeTypePtrArg(inst, rOut)
		case Transform3D:
			Transform3DEncoder.EncodeTypePtrArg(inst, rOut)
		case Projection:
			ProjectionEncoder.EncodeTypePtrArg(inst, rOut)
		case Color:
			ColorEncoder.EncodeTypePtrArg(inst, rOut)
		case StringName:
			StringNameEncoder.EncodeTypePtrArg(inst, rOut)
		case NodePath:
			NodePathEncoder.EncodeTypePtrArg(inst, rOut)
		case RID:
			RIDEncoder.EncodeTypePtrArg(inst, rOut)
		case Callable:
			CallableEncoder.EncodeTypePtrArg(inst, rOut)
		case Signal:
			SignalEncoder.EncodeTypePtrArg(inst, rOut)
		case Dictionary:
			DictionaryEncoder.EncodeTypePtrArg(inst, rOut)
		case Array:
			ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedByteArray:
			PackedByteArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedInt32Array:
			PackedInt32ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedInt64Array:
			PackedInt64ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedFloat32Array:
			PackedFloat32ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedFloat64Array:
			PackedFloat64ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedStringArray:
			PackedStringArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedVector2Array:
			PackedVector2ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedVector3Array:
			PackedVector3ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedColorArray:
			PackedColorArrayEncoder.EncodeTypePtrArg(inst, rOut)
		default:
			log.Panic("unhandled go struct to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k))
		}
	case reflect.Pointer:
		switch {
		case value.Type().Implements(gdObjectType):
			inst := value.Interface().(Object)
			ObjectEncoder.EncodeTypePtrArg(inst, rOut)
		default:
			log.Panic("unhandled pointer type to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k),
			)
		}
	case reflect.Array:
		switch inst := value.Interface().(type) {
		case Variant:
			VariantEncoder.EncodeTypePtrArg(inst, rOut)
		case String:
			StringEncoder.EncodeTypePtrArg(inst, rOut)
		case Vector2:
			Vector2Encoder.EncodeTypePtrArg(inst, rOut)
		case Vector2i:
			Vector2iEncoder.EncodeTypePtrArg(inst, rOut)
		case Rect2:
			Rect2Encoder.EncodeTypePtrArg(inst, rOut)
		case Rect2i:
			Rect2iEncoder.EncodeTypePtrArg(inst, rOut)
		case Vector3:
			Vector3Encoder.EncodeTypePtrArg(inst, rOut)
		case Vector3i:
			Vector3iEncoder.EncodeTypePtrArg(inst, rOut)
		case Transform2D:
			Transform2DEncoder.EncodeTypePtrArg(inst, rOut)
		case Vector4:
			Vector4Encoder.EncodeTypePtrArg(inst, rOut)
		case Vector4i:
			Vector4iEncoder.EncodeTypePtrArg(inst, rOut)
		case Plane:
			PlaneEncoder.EncodeTypePtrArg(inst, rOut)
		case Quaternion:
			QuaternionEncoder.EncodeTypePtrArg(inst, rOut)
		case AABB:
			AABBEncoder.EncodeTypePtrArg(inst, rOut)
		case Basis:
			BasisEncoder.EncodeTypePtrArg(inst, rOut)
		case Transform3D:
			Transform3DEncoder.EncodeTypePtrArg(inst, rOut)
		case Projection:
			ProjectionEncoder.EncodeTypePtrArg(inst, rOut)
		case Color:
			ColorEncoder.EncodeTypePtrArg(inst, rOut)
		case StringName:
			StringNameEncoder.EncodeTypePtrArg(inst, rOut)
		case NodePath:
			NodePathEncoder.EncodeTypePtrArg(inst, rOut)
		case RID:
			RIDEncoder.EncodeTypePtrArg(inst, rOut)
		case Callable:
			CallableEncoder.EncodeTypePtrArg(inst, rOut)
		case Signal:
			SignalEncoder.EncodeTypePtrArg(inst, rOut)
		case Dictionary:
			DictionaryEncoder.EncodeTypePtrArg(inst, rOut)
		case Array:
			ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedByteArray:
			PackedByteArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedInt32Array:
			PackedInt32ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedInt64Array:
			PackedInt64ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedFloat32Array:
			PackedFloat32ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedFloat64Array:
			PackedFloat64ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedStringArray:
			PackedStringArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedVector2Array:
			PackedVector2ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedVector3Array:
			PackedVector3ArrayEncoder.EncodeTypePtrArg(inst, rOut)
		case PackedColorArray:
			PackedColorArrayEncoder.EncodeTypePtrArg(inst, rOut)
		default:
			log.Panic("unhandled array value type to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k),
				zap.Any("inst", inst),
			)
		}
	default:
		log.Panic("unhandled native value to GDExtensionTypePtr",
			zap.Any("value", value),
			zap.Any("kind", k),
		)
	}
}

func GDExtensionVariantPtrFromReflectValue(value reflect.Value, rOut GDExtensionUninitializedVariantPtr) {
	log.Debug("GDExtensionVariantPtrFromReflectValue called",
		zap.String("type", value.Type().Name()),
	)
	k := value.Kind()
	switch k {
	case reflect.Bool:
		BoolEncoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Int:
		IntEncoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Int8:
		Int8Encoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Int16:
		Int16Encoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Int32:
		Int32Encoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Int64:
		Int64Encoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Uint:
		UintEncoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Uint8:
		Uint8Encoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Uint16:
		Uint16Encoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Uint32:
		Uint32Encoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Uint64:
		Uint64Encoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Float32:
		Float32Encoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Float64:
		Float64Encoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.String:
		GoStringUtf8Encoder.EncodeReflectVariantPtrArg(value, rOut)
	case reflect.Interface:
		switch inst := value.Interface().(type) {
		case Object:
			ObjectEncoder.EncodeVariantPtrArg(inst, rOut)
		default:
			log.Panic("unhandled go interface to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k),
			)
		}
	case reflect.Array:
		switch inst := value.Interface().(type) {
		case Variant:
			VariantEncoder.EncodeVariantPtrArg(inst, rOut)
		case Vector2:
			Vector2Encoder.EncodeVariantPtrArg(inst, rOut)
		case String:
			StringEncoder.EncodeVariantPtrArg(inst, rOut)
		case Vector2i:
			Vector2iEncoder.EncodeVariantPtrArg(inst, rOut)
		case Rect2:
			Rect2Encoder.EncodeVariantPtrArg(inst, rOut)
		case Rect2i:
			Rect2iEncoder.EncodeVariantPtrArg(inst, rOut)
		case Vector3:
			Vector3Encoder.EncodeVariantPtrArg(inst, rOut)
		case Vector3i:
			Vector3iEncoder.EncodeVariantPtrArg(inst, rOut)
		case Transform2D:
			Transform2DEncoder.EncodeVariantPtrArg(inst, rOut)
		case Vector4:
			Vector4Encoder.EncodeVariantPtrArg(inst, rOut)
		case Vector4i:
			Vector4iEncoder.EncodeVariantPtrArg(inst, rOut)
		case Plane:
			PlaneEncoder.EncodeVariantPtrArg(inst, rOut)
		case Quaternion:
			QuaternionEncoder.EncodeVariantPtrArg(inst, rOut)
		case AABB:
			AABBEncoder.EncodeVariantPtrArg(inst, rOut)
		case Basis:
			BasisEncoder.EncodeVariantPtrArg(inst, rOut)
		case Transform3D:
			Transform3DEncoder.EncodeVariantPtrArg(inst, rOut)
		case Projection:
			ProjectionEncoder.EncodeVariantPtrArg(inst, rOut)
		case Color:
			ColorEncoder.EncodeVariantPtrArg(inst, rOut)
		case StringName:
			StringNameEncoder.EncodeVariantPtrArg(inst, rOut)
		case NodePath:
			NodePathEncoder.EncodeVariantPtrArg(inst, rOut)
		case RID:
			RIDEncoder.EncodeVariantPtrArg(inst, rOut)
		case Callable:
			CallableEncoder.EncodeVariantPtrArg(inst, rOut)
		case Signal:
			SignalEncoder.EncodeVariantPtrArg(inst, rOut)
		case Dictionary:
			DictionaryEncoder.EncodeVariantPtrArg(inst, rOut)
		case Array:
			ArrayEncoder.EncodeVariantPtrArg(inst, rOut)
		case PackedByteArray:
			PackedByteArrayEncoder.EncodeVariantPtrArg(inst, rOut)
		case PackedInt32Array:
			PackedInt32ArrayEncoder.EncodeVariantPtrArg(inst, rOut)
		case PackedInt64Array:
			PackedInt64ArrayEncoder.EncodeVariantPtrArg(inst, rOut)
		case PackedFloat32Array:
			PackedFloat32ArrayEncoder.EncodeVariantPtrArg(inst, rOut)
		case PackedFloat64Array:
			PackedFloat64ArrayEncoder.EncodeVariantPtrArg(inst, rOut)
		case PackedStringArray:
			PackedStringArrayEncoder.EncodeVariantPtrArg(inst, rOut)
		case PackedVector2Array:
			PackedVector2ArrayEncoder.EncodeVariantPtrArg(inst, rOut)
		case PackedVector3Array:
			PackedVector3ArrayEncoder.EncodeVariantPtrArg(inst, rOut)
		case PackedColorArray:
			PackedColorArrayEncoder.EncodeVariantPtrArg(inst, rOut)
		default:
			log.Panic("unhandled array type to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k))
		}
	case reflect.Pointer:
		switch {
		case value.Type().Implements(gdObjectType):
			ObjectEncoder.EncodeReflectVariantPtrArg(value, rOut)
		default:
			log.Panic("unhandled pointer type to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k),
			)
		}
	case reflect.Struct:
		className := value.Type().Name()
		encoder, ok := GDRegisteredGDClassEncoders.Get(className)
		if ok {
			encoder.EncodeReflectVariantPtrArg(value, rOut)
			return
		}
		log.Panic("unhandled go struct to GDExtensionTypePtr",
			zap.Any("value", value),
			zap.Any("kind", k))
	default:
		log.Panic("unhandled native value to GDExtensionTypePtr",
			zap.Any("value", value),
			zap.Any("kind", k))
	}
}
