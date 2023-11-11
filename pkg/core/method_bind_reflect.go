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

func reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs(reciever GDClass, suppliedArgs []Variant, expectedArgTypes []reflect.Type) []reflect.Value {
	argsCount := len(expectedArgTypes)
	args := make([]reflect.Value, argsCount+1)
	// add receiver instance as the first argument
	args[0] = reflect.ValueOf(reciever)
	for i := 0; i < argsCount; i++ {
		arg := suppliedArgs[i]
		t := expectedArgTypes[i]
		if t == nil {
			log.Panic("expectedArgType cannot be nil",
				zap.Any("type", t),
				zap.Int("arg_index", i),
			)
		}
		switch t.Kind() {
		case reflect.Bool:
			typedValue := arg.ToBool()
			log.Debug("ptrcall arg parsed",
				zap.Bool("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "bool"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Int:
			typedValue := int(arg.ToInt64())
			log.Debug("ptrcall arg parsed",
				zap.Int("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Int8:
			typedValue := int8(arg.ToInt64())
			log.Debug("ptrcall arg parsed",
				zap.Int8("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int8"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Int16:
			typedValue := int16(arg.ToInt64())
			log.Debug("ptrcall arg parsed",
				zap.Int16("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int16"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Int32:
			typedValue := int32(arg.ToInt64())
			log.Debug("ptrcall arg parsed",
				zap.Int32("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int32"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Int64:
			typedValue := arg.ToInt64()
			log.Debug("ptrcall arg parsed",
				zap.Int64("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "int64"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Uint:
			typedValue := uint(arg.ToInt64())
			log.Debug("ptrcall arg parsed",
				zap.Uint("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Uint8:
			typedValue := uint8(arg.ToInt64())
			log.Debug("ptrcall arg parsed",
				zap.Uint8("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint8"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Uint16:
			typedValue := uint16(arg.ToInt64())
			log.Debug("ptrcall arg parsed",
				zap.Uint16("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint16"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Uint32:
			typedValue := uint32(arg.ToInt64())
			log.Debug("ptrcall arg parsed",
				zap.Uint32("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint32"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Uint64:
			typedValue := uint64(arg.ToInt64())
			log.Debug("ptrcall arg parsed",
				zap.Uint64("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "uint64"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Float32:
			typedValue := float32(arg.ToFloat64())
			log.Debug("ptrcall arg parsed",
				zap.Float32("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "float32"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.Float64:
			typedValue := arg.ToFloat64()
			log.Debug("ptrcall arg parsed",
				zap.Float64("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "float64"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
		case reflect.String:
			// TODO: how to support native go strings, StringName, and Godot String?
			typedValue := arg.ToGoString()
			log.Debug("ptrcall arg parsed",
				zap.Any("value", typedValue),
				zap.Int("arg_index", i),
				zap.String("type", "string"),
			)
			args[i+1] = reflect.ValueOf(typedValue)
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
					zap.Int("arg_index", i),
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
					args[i+1] = reflect.Zero(t)
					break
				}
				ref := constructor(obj.(RefCounted))
				args[i+1] = reflect.ValueOf(ref)
			case t.Implements(gdObjectType):
				if arg.IsNil() {
					args[i+1] = reflect.Zero(t)
					break
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
						zap.Int("arg_index", i),
						zap.Any("type", t),
					)
				}
				inst := constructor(owner).(Object)
				log.Info("ptrcall arg parsed",
					zap.Int("arg_index", i),
					zap.String("class_name", className),
				)
				args[i+1] = reflect.ValueOf(inst)
			default:
				log.Panic("unsupported interface type",
					zap.Int("arg_index", i),
					zap.Any("type", t),
				)
			}
		case reflect.Array:
			v := reflect.Zero(t)
			inst := v.Interface()
			switch inst.(type) {
			case Variant:
				v := NewVariantCopyWithGDExtensionConstVariantPtr(arg.NativeConstPtr())
				args[i+1] = reflect.ValueOf(v)
			case Vector2:
				v := arg.ToVector2()
				args[i+1] = reflect.ValueOf(v)
			case Vector2i:
				v := arg.ToVector2i()
				args[i+1] = reflect.ValueOf(v)
			case Vector3:
				v := arg.ToVector3()
				args[i+1] = reflect.ValueOf(v)
			case Vector3i:
				v := arg.ToVector3i()
				args[i+1] = reflect.ValueOf(v)
			case Vector4:
				v := arg.ToVector4()
				args[i+1] = reflect.ValueOf(v)
			case Vector4i:
				v := arg.ToVector4i()
				args[i+1] = reflect.ValueOf(v)
			default:
				log.Panic("unsupported array type",
					zap.Int("arg_index", i),
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
					zap.Int("arg_index", i),
					zap.String("type", "Ref"),
				)
				args[i+1] = reflect.ValueOf(ref)
			case t.Implements(gdClassType):
				obj := arg.ToObject()
				// NOTE: add .Elem() if we want to support
				args[i+1] = reflect.ValueOf(obj)
				break
			default:
				log.Panic("unsupported pointer type",
					zap.Int("arg_index", i),
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
					zap.Int("arg_index", i),
					zap.String("type", "Ref"),
				)
				args[i+1] = reflect.ValueOf(ref)
			default:
				log.Panic("unsupported struct type",
					zap.Int("arg_index", i),
					zap.Any("type", t),
				)
			}
		default:
			log.Panic("unsupported type",
				zap.Int("arg_index", i),
				zap.Any("type", t),
			)
		}
	}
	log.Debug("argument converted",
		zap.String("args", spew.Sdump(args)),
	)
	return args
}

func reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs(inst GDClass, suppliedArgs []GDExtensionConstTypePtr, expectedArgTypes []reflect.Type) []reflect.Value {
	argsCount := len(expectedArgTypes)
	args := make([]reflect.Value, argsCount+1)
	// add receiver instance as the first argument
	args[0] = reflect.ValueOf(inst)
	for i := 0; i < argsCount; i++ {
		arg := suppliedArgs[i]
		t := expectedArgTypes[i]
		switch t.Kind() {
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
		case reflect.Struct:
			v := reflect.Zero(t)
			inst := v.Interface()
			switch inst.(type) {
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
			log.Panic(fmt.Sprintf("MethodBind.Ptrcall does not support type: %s", t.Name()))
		}
	}
	log.Debug("argument converted",
		zap.String("args", spew.Sdump(args)),
	)
	return args
}

func validateReturnValues(reflectedRet []reflect.Value, returnStyle ReturnStyle, expectedReturnType reflect.Type) error {
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
			if expectedReturnType.Name() != reflectedRet[0].Type().Name() {
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
