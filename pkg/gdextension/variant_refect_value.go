package gdextension

// #include <godot/gdextension_interface.h>
import "C"
import (
	"reflect"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
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
			ObjectEncoder.encodeTypePtrArg(inst, rOut)
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
			VariantEncoder.encodeTypePtrArg(inst, rOut)
		case String:
			StringEncoder.encodeTypePtrArg(inst, rOut)
		case Vector2:
			Vector2Encoder.encodeTypePtrArg(inst, rOut)
		case Vector2i:
			Vector2iEncoder.encodeTypePtrArg(inst, rOut)
		case Rect2:
			Rect2Encoder.encodeTypePtrArg(inst, rOut)
		case Rect2i:
			Rect2iEncoder.encodeTypePtrArg(inst, rOut)
		case Vector3:
			Vector3Encoder.encodeTypePtrArg(inst, rOut)
		case Vector3i:
			Vector3iEncoder.encodeTypePtrArg(inst, rOut)
		case Transform2D:
			Transform2DEncoder.encodeTypePtrArg(inst, rOut)
		case Vector4:
			Vector4Encoder.encodeTypePtrArg(inst, rOut)
		case Vector4i:
			Vector4iEncoder.encodeTypePtrArg(inst, rOut)
		case Plane:
			PlaneEncoder.encodeTypePtrArg(inst, rOut)
		case Quaternion:
			QuaternionEncoder.encodeTypePtrArg(inst, rOut)
		case AABB:
			AABBEncoder.encodeTypePtrArg(inst, rOut)
		case Basis:
			BasisEncoder.encodeTypePtrArg(inst, rOut)
		case Transform3D:
			Transform3DEncoder.encodeTypePtrArg(inst, rOut)
		case Projection:
			ProjectionEncoder.encodeTypePtrArg(inst, rOut)
		case Color:
			ColorEncoder.encodeTypePtrArg(inst, rOut)
		case StringName:
			StringNameEncoder.encodeTypePtrArg(inst, rOut)
		case NodePath:
			NodePathEncoder.encodeTypePtrArg(inst, rOut)
		case RID:
			RIDEncoder.encodeTypePtrArg(inst, rOut)
		case Callable:
			CallableEncoder.encodeTypePtrArg(inst, rOut)
		case Signal:
			SignalEncoder.encodeTypePtrArg(inst, rOut)
		case Dictionary:
			DictionaryEncoder.encodeTypePtrArg(inst, rOut)
		case Array:
			ArrayEncoder.encodeTypePtrArg(inst, rOut)
		case PackedByteArray:
			PackedByteArrayEncoder.encodeTypePtrArg(inst, rOut)
		case PackedInt32Array:
			PackedInt32ArrayEncoder.encodeTypePtrArg(inst, rOut)
		case PackedInt64Array:
			PackedInt64ArrayEncoder.encodeTypePtrArg(inst, rOut)
		case PackedFloat32Array:
			PackedFloat32ArrayEncoder.encodeTypePtrArg(inst, rOut)
		case PackedFloat64Array:
			PackedFloat64ArrayEncoder.encodeTypePtrArg(inst, rOut)
		case PackedStringArray:
			PackedStringArrayEncoder.encodeTypePtrArg(inst, rOut)
		case PackedVector2Array:
			PackedVector2ArrayEncoder.encodeTypePtrArg(inst, rOut)
		case PackedVector3Array:
			PackedVector3ArrayEncoder.encodeTypePtrArg(inst, rOut)
		case PackedColorArray:
			PackedColorArrayEncoder.encodeTypePtrArg(inst, rOut)
		default:
			log.Panic("unhandled go struct to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k))
		}
	case reflect.Pointer:
		switch {
		case value.Type().Implements(gdObjectType):
			inst := value.Interface().(Object)
			ObjectEncoder.encodeTypePtrArg(inst, rOut)
		default:
			log.Panic("unhandled pointer type to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k),
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
		BoolEncoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Int:
		IntEncoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Int8:
		Int8Encoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Int16:
		Int16Encoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Int32:
		Int32Encoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Int64:
		Int64Encoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Uint:
		UintEncoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Uint8:
		Uint8Encoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Uint16:
		Uint16Encoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Uint32:
		Uint32Encoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Uint64:
		Uint64Encoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Float32:
		Float32Encoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Float64:
		Float64Encoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.String:
		GoStringUtf8Encoder.encodeReflectVariantPtrArg(value, rOut)
	case reflect.Interface:
		switch inst := value.Interface().(type) {
		case Object:
			ObjectEncoder.encodeVariantPtrArg(inst, rOut)
			// if inst == nil {
			// 		GDExtensionVariantPtrWithNil(rOut)
			// 		return
			// }
			// v := NewVariantObject(inst)
			// copyVariantWithGDExtensionTypePtr(rOut, v.nativeConstPtr())
		default:
			log.Panic("unhandled go interface to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k),
			)
		}
	case reflect.Array:
		switch inst := value.Interface().(type) {
		case Variant:
			VariantEncoder.encodeVariantPtrArg(inst, rOut)
		case Vector2:
			Vector2Encoder.encodeVariantPtrArg(inst, rOut)
		case String:
			StringEncoder.encodeVariantPtrArg(inst, rOut)
		case Vector2i:
			Vector2iEncoder.encodeVariantPtrArg(inst, rOut)
		case Rect2:
			Rect2Encoder.encodeVariantPtrArg(inst, rOut)
		case Rect2i:
			Rect2iEncoder.encodeVariantPtrArg(inst, rOut)
		case Vector3:
			Vector3Encoder.encodeVariantPtrArg(inst, rOut)
		case Vector3i:
			Vector3iEncoder.encodeVariantPtrArg(inst, rOut)
		case Transform2D:
			Transform2DEncoder.encodeVariantPtrArg(inst, rOut)
		case Vector4:
			Vector4Encoder.encodeVariantPtrArg(inst, rOut)
		case Vector4i:
			Vector4iEncoder.encodeVariantPtrArg(inst, rOut)
		case Plane:
			PlaneEncoder.encodeVariantPtrArg(inst, rOut)
		case Quaternion:
			QuaternionEncoder.encodeVariantPtrArg(inst, rOut)
		case AABB:
			AABBEncoder.encodeVariantPtrArg(inst, rOut)
		case Basis:
			BasisEncoder.encodeVariantPtrArg(inst, rOut)
		case Transform3D:
			Transform3DEncoder.encodeVariantPtrArg(inst, rOut)
		case Projection:
			ProjectionEncoder.encodeVariantPtrArg(inst, rOut)
		case Color:
			ColorEncoder.encodeVariantPtrArg(inst, rOut)
		case StringName:
			StringNameEncoder.encodeVariantPtrArg(inst, rOut)
		case NodePath:
			NodePathEncoder.encodeVariantPtrArg(inst, rOut)
		case RID:
			RIDEncoder.encodeVariantPtrArg(inst, rOut)
		case Callable:
			CallableEncoder.encodeVariantPtrArg(inst, rOut)
		case Signal:
			SignalEncoder.encodeVariantPtrArg(inst, rOut)
		case Dictionary:
			DictionaryEncoder.encodeVariantPtrArg(inst, rOut)
		case Array:
			ArrayEncoder.encodeVariantPtrArg(inst, rOut)
		case PackedByteArray:
			PackedByteArrayEncoder.encodeVariantPtrArg(inst, rOut)
		case PackedInt32Array:
			PackedInt32ArrayEncoder.encodeVariantPtrArg(inst, rOut)
		case PackedInt64Array:
			PackedInt64ArrayEncoder.encodeVariantPtrArg(inst, rOut)
		case PackedFloat32Array:
			PackedFloat32ArrayEncoder.encodeVariantPtrArg(inst, rOut)
		case PackedFloat64Array:
			PackedFloat64ArrayEncoder.encodeVariantPtrArg(inst, rOut)
		case PackedStringArray:
			PackedStringArrayEncoder.encodeVariantPtrArg(inst, rOut)
		case PackedVector2Array:
			PackedVector2ArrayEncoder.encodeVariantPtrArg(inst, rOut)
		case PackedVector3Array:
			PackedVector3ArrayEncoder.encodeVariantPtrArg(inst, rOut)
		case PackedColorArray:
			PackedColorArrayEncoder.encodeVariantPtrArg(inst, rOut)
		default:
			log.Panic("unhandled array type to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k))
		}
	case reflect.Pointer:
		switch {
		case value.Type().Implements(gdObjectType):
			ObjectEncoder.encodeReflectVariantPtrArg(value, rOut)
		default:
			log.Panic("unhandled pointer type to GDExtensionTypePtr",
				zap.Any("value", value),
				zap.Any("kind", k),
			)
		}
	case reflect.Struct:
		className := value.Type().Name()
		encoder, ok := gdRegisteredGDClassEncoders.Get(className)
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
