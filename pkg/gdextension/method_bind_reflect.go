package gdextension

// #include <godot/gdextension_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

var (
	gdArrayType   = reflect.TypeOf((*Array)(nil)).Elem()
	gdVariantType = reflect.TypeOf((*Variant)(nil)).Elem()
	errorType     = reflect.TypeOf((*error)(nil)).Elem()
)

func reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs(inst GDClass, suppliedArgs []Variant, expectedArgTypes []reflect.Type) []reflect.Value {
	argsCount := len(expectedArgTypes)
	args := make([]reflect.Value, argsCount+1)
	// add receiver instance as the first argument
	args[0] = reflect.ValueOf(inst)
	for i := 0; i < argsCount; i++ {
		arg := suppliedArgs[i]
		t := expectedArgTypes[i]
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
			log.Panic("MethodBind.Ptrcall slice not implemented",
				zap.Any("stringify", arg.Stringify()),
			)
		case reflect.Interface:
			v := reflect.Zero(t)
			inst := v.Interface()
			switch inst.(type) {
			case Object:
				obj := arg.ToObject()
				gdObjPtr := (GDExtensionConstObjectPtr)(unsafe.Pointer(obj.GetGodotObjectOwner()))
				gdsn := NewStringName()
				defer gdsn.Destroy()
				ptr := (GDExtensionUninitializedStringNamePtr)(unsafe.Pointer(gdsn.ptr()))
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
				constructor, ok := gdNativeConstructors.Get(className)
				if !ok {
					log.Panic(fmt.Sprintf("MethodBind.Ptrcall does not support gdextension class type: %s", className))
				}
				inst := constructor(owner).(RefCounted)
				log.Debug("ptrcall arg parsed",
					zap.Int("arg_index", i),
					zap.String("type", "Ref"),
					zap.String("class_name", className),
				)
				args[i+1] = reflect.ValueOf(inst)
			default:
				log.Panic(fmt.Sprintf("MethodBind.Ptrcall does not support interface type: %s", t.Name()))
			}
		case reflect.Struct:
			v := reflect.Zero(t)
			inst := v.Interface()
			switch inst.(type) {
			case Ref:
				inst := arg.ToObject().(RefCounted)
				log.Debug("ptrcall arg parsed",
					zap.Int("arg_index", i),
					zap.String("type", "Ref"),
				)
				ref := NewRef(inst)
				args[i+1] = reflect.ValueOf(*ref)
			case Variant:
				v := NewVariantCopyWithGDExtensionConstVariantPtr((GDExtensionConstVariantPtr)(arg.ptr()))
				args[i+1] = reflect.ValueOf(v)
			default:
				log.Panic(fmt.Sprintf("MethodBind.Ptrcall does not support struct type: %s", t.Name()))
			}
		default:
			log.Panic(fmt.Sprintf("MethodBind.Ptrcall does not support type: %s", t.Name()))
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
			asciiValue := typedValue.ToUtf8()
			log.Debug("ptrcall arg parsed",
				zap.Any("ascii", asciiValue),
				zap.Int("arg_index", i),
				zap.String("type", "string"),
			)
			args[i+1] = reflect.ValueOf(asciiValue)
		case reflect.Slice:
			slice := *(*[]unsafe.Pointer)(arg)
			log.Panic("MethodBind.Ptrcall slice not implemented",
				zap.Any("len", len(slice)),
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
				ptr := (GDExtensionUninitializedStringNamePtr)(unsafe.Pointer(gdsn.ptr()))
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
				constructor, ok := gdNativeConstructors.Get(className)
				if !ok {
					log.Panic(fmt.Sprintf("MethodBind.Ptrcall does not support gdextension class type: %s", className))
				}
				inst := constructor(owner).(RefCounted)
				log.Debug("ptrcall arg parsed",
					zap.Int("arg_index", i),
					zap.String("type", "Ref"),
					zap.String("class_name", className),
				)
				args[i+1] = reflect.ValueOf(inst)
			default:
				log.Panic(fmt.Sprintf("MethodBind.Ptrcall does not support interface type: %s", t.Name()))
			}
		case reflect.Struct:
			v := reflect.Zero(t)
			inst := v.Interface()
			switch inst.(type) {
			case Ref:
				gdRefPtr := (GDExtensionConstRefPtr)(arg)
				gdObjPtr := (GDExtensionConstObjectPtr)(CallFunc_GDExtensionInterfaceRefGetObject(gdRefPtr))

				// GDExtensionUninitializedStringNamePtr
				gdsn := NewStringName()
				defer gdsn.Destroy()
				ptr := (GDExtensionUninitializedStringNamePtr)(unsafe.Pointer(gdsn.ptr()))
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
				constructor, ok := gdNativeConstructors.Get(className)
				if !ok {
					log.Panic(fmt.Sprintf("MethodBind.Ptrcall does not support gdextension class type: %s", className))
				}
				inst := constructor(owner).(RefCounted)
				log.Debug("ptrcall arg parsed",
					zap.Int("arg_index", i),
					zap.String("type", "Ref"),
					zap.String("class_name", className),
				)
				ref := NewRef(inst)
				args[i+1] = reflect.ValueOf(*ref)
			case Variant:
				v := NewVariantCopyWithGDExtensionConstVariantPtr((GDExtensionConstVariantPtr)(arg))
				args[i+1] = reflect.ValueOf(v)
			default:
				log.Panic(fmt.Sprintf("MethodBind.Ptrcall does not support struct type: %s", t.Name()))
			}
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
