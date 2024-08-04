package core

// #include <godot/gdextension_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassinit"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

var (
	gdClassType          = reflect.TypeOf((*GDClass)(nil)).Elem()
	gdExtensionClassType = reflect.TypeOf((*GDExtensionClass)(nil)).Elem()
	gdObjectType         = reflect.TypeOf((*Object)(nil)).Elem()
	gdArrayType          = reflect.TypeOf((*Array)(nil)).Elem()
	gdVariantType        = reflect.TypeOf((*Variant)(nil)).Elem()
	errorType            = reflect.TypeOf((*error)(nil)).Elem()
	refType              = reflect.TypeOf((*Ref)(nil)).Elem()
)

// reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs is called for each
// function call argument that needs to be translated when GDScript calls into Go.
func reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs(reciever GDClass, suppliedArgs []Variant, expectedArgTypes []reflect.Type) []reflect.Value {
	argsCount := len(expectedArgTypes)
	args := make([]reflect.Value, argsCount+1)
	// add receiver instance as the first argument
	args[0] = reflect.ValueOf(reciever)
	for i := 0; i < argsCount; i++ {
		v, err := convertVariantToGoTypeReflectValue(suppliedArgs[i], expectedArgTypes[i])
		if err != nil {
			log.Panic("error converting variant to go type",
				zap.Int("arg_index", i),
				zap.Error(err),
			)
		}
		args[i+1] = v
	}
	log.Debug("argument converted",
		zap.String("args", spew.Sdump(args)),
	)
	return args
}

func convertVariantToGoTypeReflectValue(arg Variant, t reflect.Type) (reflect.Value, error) {
	var out reflect.Value
	if t == nil {
		return out, fmt.Errorf("type cannot be nil")
	}
	switch t.Kind() {
	case reflect.Bool:
		typedValue := arg.ToBool()
		log.Debug("ptrcall arg parsed",
			zap.Bool("value", typedValue),
			zap.String("type", "bool"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Int:
		typedValue := int(arg.ToInt64())
		log.Debug("ptrcall arg parsed",
			zap.Int("value", typedValue),
			zap.String("type", "int"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Int8:
		typedValue := int8(arg.ToInt64())
		log.Debug("ptrcall arg parsed",
			zap.Int8("value", typedValue),
			zap.String("type", "int8"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Int16:
		typedValue := int16(arg.ToInt64())
		log.Debug("ptrcall arg parsed",
			zap.Int16("value", typedValue),
			zap.String("type", "int16"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Int32:
		typedValue := int32(arg.ToInt64())
		log.Debug("ptrcall arg parsed",
			zap.Int32("value", typedValue),
			zap.String("type", "int32"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Int64:
		typedValue := arg.ToInt64()
		log.Debug("ptrcall arg parsed",
			zap.Int64("value", typedValue),
			zap.String("type", "int64"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Uint:
		typedValue := uint(arg.ToInt64())
		log.Debug("ptrcall arg parsed",
			zap.Uint("value", typedValue),
			zap.String("type", "uint"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Uint8:
		typedValue := uint8(arg.ToInt64())
		log.Debug("ptrcall arg parsed",
			zap.Uint8("value", typedValue),
			zap.String("type", "uint8"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Uint16:
		typedValue := uint16(arg.ToInt64())
		log.Debug("ptrcall arg parsed",
			zap.Uint16("value", typedValue),
			zap.String("type", "uint16"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Uint32:
		typedValue := uint32(arg.ToInt64())
		log.Debug("ptrcall arg parsed",
			zap.Uint32("value", typedValue),
			zap.String("type", "uint32"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Uint64:
		typedValue := uint64(arg.ToInt64())
		log.Debug("ptrcall arg parsed",
			zap.Uint64("value", typedValue),
			zap.String("type", "uint64"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Float32:
		typedValue := float32(arg.ToFloat64())
		log.Debug("ptrcall arg parsed",
			zap.Float32("value", typedValue),
			zap.String("type", "float32"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Float64:
		typedValue := arg.ToFloat64()
		log.Debug("ptrcall arg parsed",
			zap.Float64("value", typedValue),
			zap.String("type", "float64"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.String:
		// TODO: how to support native go strings, StringName, and Godot String?
		typedValue := arg.ToGoString()
		log.Debug("ptrcall arg parsed",
			zap.Any("value", typedValue),
			zap.String("type", "string"),
		)
		return reflect.ValueOf(typedValue), nil
	case reflect.Slice:
		log.Panic("slice not implemented",
			zap.Any("stringify", arg.Stringify()),
		)
	case reflect.Interface:
		switch {
		case t.Implements(refType):
			obj := arg.ToObject()
			log.Debug("ptrcall arg parsed",
				zap.String("value", arg.Stringify()),
				zap.String("type", t.Name()),
			)
			refTypeName := t.Name()
			constructor, ok := GDClassRefConstructors.Get(refTypeName[3:])
			if !ok {
				log.Fatal("unable to get ref constructor",
					zap.String("type", t.Name()),
				)
			}
			if obj == nil {
				return reflect.Zero(t), nil

			}
			ref := constructor(obj.(RefCounted))
			return reflect.ValueOf(ref), nil
		case t.Implements(gdObjectType):
			if arg.IsNil() {
				return reflect.Zero(t), nil
			}
			obj := arg.ToObject()
			gdsClass := obj.GetClass()
			className := gdsClass.ToUtf8()
			log.Debug("found object arg",
				zap.String("class", obj.GetClassName()),
				zap.String("class from gd", className),
			)
			gdObjPtr := obj.AsGDExtensionConstObjectPtr()
			// gdsn := StringName{}
			// ptr := (GDExtensionUninitializedStringNamePtr)(unsafe.Pointer(gdsn.NativePtr()))
			// cok := CallFunc_GDExtensionInterfaceObjectGetClassName(gdObjPtr, FFI.Library, ptr)
			// if cok == 0 {
			// 	log.Panic("failed to get class name",
			// 		zap.String("class", gdsn.ToUtf8()),
			// 	)
			// }
			// defer gdsn.Destroy()
			owner := (*GodotObject)(gdObjPtr)
			constructor, ok := GDNativeConstructors.Get(className)
			if !ok {
				log.Panic("unsupported interface class name",
					zap.String("class_name", className),
					zap.Any("type", t),
				)
			}
			inst := constructor(owner).(Object)
			log.Info("ptrcall arg parsed",
				zap.String("class_name", className),
			)
			return reflect.ValueOf(inst), nil
		default:
			log.Panic("unsupported interface type",
				zap.Any("type", t),
			)
		}
	case reflect.Array:
		v := reflect.Zero(t)
		inst := v.Interface()
		switch inst.(type) {
		case Variant:
			v := NewVariantCopyWithGDExtensionConstVariantPtr(arg.NativeConstPtr())
			return reflect.ValueOf(v), nil
		case Vector2:
			v := arg.ToVector2()
			return reflect.ValueOf(v), nil
		case Vector2i:
			v := arg.ToVector2i()
			return reflect.ValueOf(v), nil
		case Vector3:
			v := arg.ToVector3()
			return reflect.ValueOf(v), nil
		case Vector3i:
			v := arg.ToVector3i()
			return reflect.ValueOf(v), nil
		case Vector4:
			v := arg.ToVector4()
			return reflect.ValueOf(v), nil
		case Vector4i:
			v := arg.ToVector4i()
			return reflect.ValueOf(v), nil
		case PackedByteArray:
			v := arg.ToPackedByteArray()
			return reflect.ValueOf(v), nil
		case PackedInt32Array:
			v := arg.ToPackedInt32Array()
			return reflect.ValueOf(v), nil
		case PackedInt64Array:
			v := arg.ToPackedInt64Array()
			return reflect.ValueOf(v), nil
		case PackedFloat32Array:
			v := arg.ToPackedFloat32Array()
			return reflect.ValueOf(v), nil
		case PackedFloat64Array:
			v := arg.ToPackedFloat64Array()
			return reflect.ValueOf(v), nil
		case PackedStringArray:
			v := arg.ToPackedStringArray()
			return reflect.ValueOf(v), nil
		case PackedVector2Array:
			v := arg.ToPackedVector2Array()
			return reflect.ValueOf(v), nil
		case PackedVector3Array:
			v := arg.ToPackedVector3Array()
			return reflect.ValueOf(v), nil
		case PackedColorArray:
			v := arg.ToPackedColorArray()
			return reflect.ValueOf(v), nil
		case Dictionary:
			v := arg.ToDictionary()
			return reflect.ValueOf(v), nil
		case Signal:
			v := arg.ToSignal()
			return reflect.ValueOf(v), nil
		case Callable:
			v := arg.ToCallable()
			return reflect.ValueOf(v), nil
		default:
			log.Panic("unsupported array type",
				zap.Any("type", t),
			)
		}
	case reflect.Pointer:
		switch {
		case t.Implements(refType):
			// TODO: is this coming out as a Ref type here?
			obj := arg.ToObject()
			ref, ok := obj.(Ref)
			if !ok {
				log.Panic("not a ref instance",
					zap.String("value", arg.Stringify()),
				)
			}
			log.Debug("ptrcall arg parsed",
				zap.String("type", "Ref"),
			)
			return reflect.ValueOf(ref), nil
		case t.Implements(gdClassType):
			obj := arg.ToObject()
			// NOTE: add .Elem() if we want to support
			return reflect.ValueOf(obj), nil
		default:
			log.Panic("unsupported pointer type",
				zap.Any("type", t),
			)
		}
	case reflect.Struct:
		switch {
		case t.Implements(refType):
			// TODO: is this coming out as a Ref type here?
			obj := arg.ToObject()
			ref, ok := obj.(Ref)
			if !ok {
				log.Panic("not a ref instance",
					zap.String("value", arg.Stringify()),
				)
			}
			log.Debug("ptrcall arg parsed",
				zap.String("type", "Ref"),
			)
			return reflect.ValueOf(ref), nil
		default:
			log.Panic("unsupported struct type",
				zap.Any("type", t),
			)
		}
	default:
		log.Panic("unsupported type",
			zap.Any("type", t),
		)
	}
	return out, nil
}

func reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs(inst GDClass, suppliedArgs []GDExtensionConstTypePtr, expectedArgTypes []reflect.Type) []reflect.Value {
	argsCount := len(expectedArgTypes)
	args := make([]reflect.Value, argsCount+1)
	// add receiver instance as the first argument
	args[0] = reflect.ValueOf(inst)
	for i := 0; i < argsCount; i++ {
		arg := suppliedArgs[i]
		t := expectedArgTypes[i]
		k := t.Kind()
		switch k {
		case reflect.Bool:
			typedValue := *(*bool)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Bool("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "bool"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Int:
			typedValue := *(*int)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Int("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Int8:
			typedValue := *(*int8)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Int8("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int8"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Int16:
			typedValue := *(*int16)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Int16("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int16"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Int32:
			typedValue := *(*int32)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Int32("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int32"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Int64:
			typedValue := *(*int64)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Int64("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int64"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Uint:
			typedValue := *(*uint)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Uint("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Uint8:
			typedValue := *(*uint8)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Uint8("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint8"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Uint16:
			typedValue := *(*uint16)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Uint16("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint16"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Uint32:
			typedValue := *(*uint32)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Uint32("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint32"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Uint64:
			typedValue := *(*uint64)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Uint64("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint64"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Float32:
			rawTypedValue := *(*float64)(arg)
			typedValue := (float32)(rawTypedValue)
			log.Debug("ptrcall arg parsed",
				zap.Float32("value", typedValue),
				zap.Float64("raw_value", rawTypedValue),
				zap.Int("arg_index", i),
				zap.String("type", "float32"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Float64:
			typedValue := *(*float64)(arg)
			log.Debug("ptrcall arg parsed",
				zap.Float64("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "float64"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.String:
			typedValue := *(*String)(arg)
			str := typedValue.ToUtf8()
			log.Debug("ptrcall arg parsed",
				zap.Any("str", str),
				zap.Int("arg_index", i),
				zap.String("type", "string"),
			)
			args[i+1] = reflect.ValueOf(str)
		case reflect.Slice:
			slice := *(*[]unsafe.Pointer)(arg)
			log.Panic("slice not implemented",
				zap.Any("len", len(slice)),
				zap.Int("arg_index", i),
			)
		case reflect.Interface:
			v := reflect.Zero(t)
			inst := v.Interface()
			switch inst.(type) {
			case Object:
				gdObjPtr := (GDExtensionConstObjectPtr)(arg)

				// GDExtensionUninitializedStringNamePtr
				gdsn := NewStringName()
				defer gdsn.Destroy()
				ptr := (GDExtensionUninitializedStringNamePtr)(unsafe.Pointer(gdsn.NativePtr()))
				cok := CallFunc_GDExtensionInterfaceObjectGetClassName(gdObjPtr, FFI.Library, ptr)
				if cok == 0 {
					log.Panic("failed to get class name",
						zap.Any("gdObjPtr", gdObjPtr),
					)
				}
				owner := (*GodotObject)(gdObjPtr)
				gds := gdsn.AsString()
				defer gds.Destroy()
				className := gds.ToUtf8()
				constructor, ok := GDNativeConstructors.Get(className)
				if !ok {
					log.Panic("does not support gdextension class type",
						zap.String("class_name", className),
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				inst := constructor(owner)
				log.Debug("ptrcall arg parsed",
					zap.Int("arg_index", i),
					zap.String("class_name", className),
				)
				args[i+1] = reflect.ValueOf(inst)
			default:
				if t.Implements(refType) {
					gdRefPtr := (GDExtensionConstRefPtr)(arg)
					gdObjPtr := (GDExtensionConstObjectPtr)(CallFunc_GDExtensionInterfaceRefGetObject(gdRefPtr))

					// gdsn := NewStringName()
					// defer gdsn.Destroy()
					// ptr := (GDExtensionUninitializedStringNamePtr)(unsafe.Pointer(gdsn.NativePtr()))
					// cok := CallFunc_GDExtensionInterfaceObjectGetClassName(gdObjPtr, FFI.Library, ptr)
					// if cok == 0 {
					// 	log.Panic("failed to get class name",
					// 		zap.Any("gdObjPtr", gdObjPtr),
					// 	)
					// }
					// gds := gdsn.AsString()
					// defer gds.Destroy()
					// className := gds.ToUtf8()
					refClassName := t.Name()
					className := refClassName[3:]
					constructor, ok := GDNativeConstructors.Get(className)
					if !ok {
						log.Panic("does not support gdextension class type",
							zap.String("class_name", className),
							zap.Int("arg_index", i),
							zap.Any("type", t),
						)
					}
					owner := (*GodotObject)(gdObjPtr)
					inst := constructor(owner).(RefCounted)
					log.Debug("ptrcall arg parsed",
						zap.Int("arg_index", i),
						zap.String("type", "Ref"),
						zap.String("class_name", className),
					)
					refConstructor, ok := GDClassRefConstructors.Get(className)
					if !ok {
						log.Panic("unable to find ref for type",
							zap.String("class_name", className),
							zap.Int("arg_index", i),
							zap.Any("type", t),
						)
					}
					ref := refConstructor(inst)
					args[i+1] = reflect.ValueOf(ref)
					break
				}
				log.Panic("unsupported interface type",
					zap.Int("arg_index", i),
					zap.Any("type", t),
				)
			}
		case reflect.Array:
			v := reflect.Zero(t)
			inst := v.Interface()
			switch inst.(type) {
			case Vector4:
				v := NewVector4WithGDExtensionConstTypePtr((GDExtensionConstTypePtr)(arg))
				args[i+1] = reflect.ValueOf(v)
			case Vector2:
				v := NewVector2WithGDExtensionConstTypePtr((GDExtensionConstTypePtr)(arg))
				args[i+1] = reflect.ValueOf(v)
			case Variant:
				v := NewVariantCopyWithGDExtensionConstVariantPtr((GDExtensionConstVariantPtr)(arg))
				args[i+1] = reflect.ValueOf(v)
			case PackedInt64Array:
				v := NewPackedInt64ArrayWithGDExtensionConstTypePtr((GDExtensionConstTypePtr)(arg))
				log.Debug("reflect PackedInt64Array", zap.Any("v", Stringify(NewVariantPackedInt64Array(v))))
				args[i+1] = reflect.ValueOf(v)
			default:
				log.Panic(fmt.Sprintf("MethodBind.Ptrcall reflected as array does not support type: %s", t.Name()))
			}
		case reflect.Struct:
			v := reflect.Zero(t)
			inst := v.Interface()
			switch inst.(type) {
			case Vector4:
				v := NewVector4WithGDExtensionConstTypePtr((GDExtensionConstTypePtr)(arg))
				args[i+1] = reflect.ValueOf(v)
			case Vector2:
				v := NewVector2WithGDExtensionConstTypePtr((GDExtensionConstTypePtr)(arg))
				args[i+1] = reflect.ValueOf(v)
			case Variant:
				v := NewVariantCopyWithGDExtensionConstVariantPtr((GDExtensionConstVariantPtr)(arg))
				args[i+1] = reflect.ValueOf(v)
			default:
				if strings.HasPrefix(t.String(), "gdextension.Ref") {
					gdRefPtr := (GDExtensionConstRefPtr)(arg)
					gdObjPtr := (GDExtensionConstObjectPtr)(CallFunc_GDExtensionInterfaceRefGetObject(gdRefPtr))

					// GDExtensionUninitializedStringNamePtr
					gdsn := NewStringName()
					defer gdsn.Destroy()
					ptr := (GDExtensionUninitializedStringNamePtr)(unsafe.Pointer(gdsn.NativePtr()))
					cok := CallFunc_GDExtensionInterfaceObjectGetClassName(gdObjPtr, FFI.Library, ptr)
					if cok == 0 {
						log.Panic("failed to get class name",
							zap.Any("gdObjPtr", gdObjPtr),
							zap.Int("arg_index", i),
							zap.Any("type", t),
						)
					}
					owner := (*GodotObject)(gdObjPtr)
					gds := gdsn.AsString()
					defer gds.Destroy()
					className := gds.ToUtf8()
					constructor, ok := GDNativeConstructors.Get(className)
					if !ok {
						log.Panic("does not support gdextension class type",
							zap.String("class_name", className),
							zap.Int("arg_index", i),
							zap.Any("type", t),
						)
					}
					inst := constructor(owner).(RefCounted)
					log.Debug("ptrcall arg parsed",
						zap.Int("arg_index", i),
						zap.String("type", "Ref"),
						zap.String("class_name", className),
					)
					refConstructor, ok := GDClassRefConstructors.Get(className)
					if !ok {
						log.Panic("unable to find ref for type",
							zap.String("class_name", className),
							zap.Int("arg_index", i),
							zap.Any("type", t),
						)
					}
					ref := refConstructor(inst)
					args[i+1] = reflect.ValueOf(ref)
					break
				}
				log.Panic("unsupported struct type",
					zap.Int("arg_index", i),
					zap.Any("type", t),
				)
			}
		case reflect.Pointer:
			log.Panic("unsupported pointer type",
				zap.Int("arg_index", i),
				zap.Any("type", t),
			)
		default:
			log.Panic(fmt.Sprintf("MethodBind.Ptrcall does not support kind %s and type: %s", k, t.Name()))
		}
	}
	log.Debug("argument converted",
		zap.String("args", spew.Sdump(args)),
	)
	return args
}

func validateReturnValues(reflectedRet []reflect.Value, returnStyle ReturnStyle, expectedReturnType reflect.Type) error {
	log.Info("validateReturnValues called",
		zap.Any("return_style", returnStyle),
		zap.Any("expected_return_type", expectedReturnType),
	)
	switch len(reflectedRet) {
	case 0:
		if returnStyle != NoneReturnStyle {
			log.Panic("unexpected return style: expected none")
		}
		return nil
	case 1:
		if expectedReturnType == nil {
			log.Panic("no return value expected but 1 returned")
		}
		switch returnStyle {
		case NoneReturnStyle:
			log.Panic("expected no return value")
		case ValueReturnStyle:
			en := expectedReturnType.Name()
			if en != reflectedRet[0].Type().Name() {
				log.Panic("unexpected return type",
					zap.String("type", reflectedRet[0].Type().Name()),
				)
			}
		default:
			log.Panic("unexpected value returned")
		}
		return nil
	default:
		log.Panic("too many values returned", zap.Any("ret", reflectedRet))
	}
	return fmt.Errorf("unexpected code path reached")
}
