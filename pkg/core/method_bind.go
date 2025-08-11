package core

// #include <godot/gdextension_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"runtime/cgo"
	"strconv"
	"strings"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"github.com/godot-go/godot-go/pkg/util"
	"go.uber.org/zap"
)

type ReturnStyle uint8

const (
	NoneReturnStyle ReturnStyle = iota
	ValueReturnStyle
	ValueAndBoolReturnStyle
	// ValueAndErrorReturnStyle
	// ErrorReturnStyle
)

// GoMethodMetadata is used as method_userdata in callbacks called from godot into Go.
type GoMethodMetadata struct {
	ClassName              string
	GdMethodName           string
	GoMethodName           string
	Func                   reflect.Value
	GoReturnType           reflect.Type
	GoReturnStyle          ReturnStyle
	GoArgumentTypes        []reflect.Type
	DefaultArguments       []Variant
	IsVariadic             bool
	IsVirtual              bool
	MethodFlags            MethodFlags
	gdeReturnType          GDExtensionVariantType
	gdeReturnPropertyInfo  GDExtensionPropertyInfo
	gdeArgumentsInfo       []GDExtensionPropertyInfo
	gdeArgumentsMetadata   []GDExtensionClassMethodArgumentMetadata
	gdeArgumentTypes       []GDExtensionVariantType
	gdeDefaultArgumentPtrs []GDExtensionVariantPtr
}

func NewGoMethodMetadata(
	method reflect.Method,
	className string,
	gdMethodName string,
	goMethodName string,
	argumentNames []string,
	defaultArguments []Variant,
	methodFlags MethodFlags,
) *GoMethodMetadata {
	mt := method.Type
	fn := method.Func
	recv := mt.In(0)
	if recv.Kind() == reflect.Pointer {
		recv = recv.Elem()
	}
	if className != recv.Name() {
		log.Panic("class name did not match reciever type",
			zap.String("class", className),
			zap.String("method", gdMethodName),
			zap.String("reciover", recv.Name()),
		)
	}
	isVariadicTyped := mt.IsVariadic()
	isVariadicFlaged := (methodFlags & METHOD_FLAG_VARARG) == METHOD_FLAG_VARARG
	isVirtual := (methodFlags & METHOD_FLAG_VIRTUAL) == METHOD_FLAG_VIRTUAL
	if isVariadicTyped != isVariadicFlaged {
		log.Panic("go method and method flags are not variadic aligned",
			zap.Bool("is_variadic_type", isVariadicTyped),
			zap.Bool("is_variadic_flag", isVariadicFlaged),
		)
	}
	returnCount := mt.NumOut()
	if returnCount > 2 {
		log.Panic("method cannot return more than 1 type",
			zap.String("class", className),
			zap.String("method", gdMethodName),
		)
	}
	var (
		goReturnType       reflect.Type
		returnStyle        ReturnStyle
		returnPropertyInfo GDExtensionPropertyInfo
	)
	switch returnCount {
	case 0:
	case 1:
		// HACK: the second return value is ignored
		goReturnType = mt.Out(0)
		returnStyle = ValueReturnStyle
	case 2:
		goReturnType = mt.Out(0)
		returnStyle = ValueAndBoolReturnStyle
		if mt.Out(1).Kind() != reflect.Bool {
			log.Panic("method 2nd return value must be of type bool",
				zap.String("method", gdMethodName),
			)
		}
	default:
		log.Panic("method cannot return more than 1 value",
			zap.String("method", gdMethodName),
		)
	}
	log.Debug("return value type",
		zap.String("class", className),
		zap.String("method", gdMethodName),
		zap.Any("type", goReturnType),
	)
	returnType := ReflectTypeToGDExtensionVariantType(goReturnType)
	if returnType != GDEXTENSION_VARIANT_TYPE_NIL {
		returnPropertyInfo = NewSimpleGDExtensionPropertyInfo(className, returnType, goReturnType.Name())
	}
	argumentCount := mt.NumIn() - 1
	if len(argumentNames) > argumentCount {
		log.Panic(`Method definition has more arguments than the actual method.`,
			zap.String("method", gdMethodName),
			zap.Int("argument_count", argumentCount),
		)
	}
	defaultArgumentPtrs := make([]GDExtensionVariantPtr, len(defaultArguments))
	for i := range defaultArgumentPtrs {
		defaultArgumentPtrs[i] = (GDExtensionVariantPtr)(defaultArguments[i].NativePtr())
	}
	goArgumentTypes := make([]reflect.Type, argumentCount)
	variantTypes := make([]GDExtensionVariantType, argumentCount)
	argumentsInfo := make([]GDExtensionPropertyInfo, argumentCount)
	argumentsMetadata := make([]GDExtensionClassMethodArgumentMetadata, argumentCount)
	for i := 0; i < argumentCount; i++ {
		t := mt.In(i + 1)
		goArgumentTypes[i] = t
		variantTypes[i] = ReflectTypeToGDExtensionVariantType(t)
		argumentsInfo[i] = NewSimpleGDExtensionPropertyInfo(className, variantTypes[i], t.Name())
		argumentsMetadata[i] = GDEXTENSION_METHOD_ARGUMENT_METADATA_NONE
	}
	ret := &GoMethodMetadata{
		ClassName:              className,
		GdMethodName:           gdMethodName,
		GoMethodName:           goMethodName,
		Func:                   fn,
		GoReturnType:           goReturnType,
		GoReturnStyle:          returnStyle,
		GoArgumentTypes:        goArgumentTypes,
		DefaultArguments:       defaultArguments,
		IsVariadic:             isVariadicFlaged,
		IsVirtual:              isVirtual,
		MethodFlags:            methodFlags,
		gdeReturnType:          returnType,
		gdeReturnPropertyInfo:  returnPropertyInfo,
		gdeArgumentsInfo:       argumentsInfo,
		gdeArgumentsMetadata:   argumentsMetadata,
		gdeArgumentTypes:       variantTypes,
		gdeDefaultArgumentPtrs: defaultArgumentPtrs,
	}
	pnr.Pin(&returnPropertyInfo)
	pnr.Pin(ret)
	return ret
}

// VarargCallFunc is the function signature that can be called from GDScript
type VarargCallFunc func(GDClass, ...Variant) Variant

// Call is called by GDScript to call into Go
func (md *GoMethodMetadata) Call(inst GDClass, gdArgs ...Variant) Variant {
	gdArgsCount := len(gdArgs)
	defArgsCount := len(md.gdeDefaultArgumentPtrs)
	callArgs := make([]Variant, len(md.gdeArgumentTypes))
	for i := range callArgs {
		if i < gdArgsCount {
			callArgs[i] = gdArgs[i]
		} else if i < defArgsCount {
			callArgs[i] = md.DefaultArguments[i]
		} else {
			log.Panic("too few arguments",
				zap.String("bind", md.String()),
				zap.String("gd_args", VariantSliceToString(gdArgs)),
				zap.String("defaults", VariantSliceToString(md.DefaultArguments)),
			)
		}
	}
	exepctedTypes := md.GoArgumentTypes
	if md.IsVariadic {
		args := []reflect.Value{
			reflect.ValueOf(inst),
			reflect.ValueOf(gdArgs),
		}
		ret := md.Func.CallSlice(args)
		log.Info("Call Variadic",
			zap.String("bind", md.String()),
			zap.String("gd_args", VariantSliceToString(gdArgs)),
			zap.String("resolved_args", VariantSliceToString(callArgs)),
			zap.String("ret", util.ReflectValueSliceToString(ret)),
		)
		switch md.GoReturnStyle {
		case NoneReturnStyle:
		case ValueAndBoolReturnStyle:
			log.Warn("second return value ignored")
			fallthrough
		case ValueReturnStyle:
			// return ret[0].Interface().(Variant)
			v := Variant{}
			ptr := (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(v.NativePtr()))
			GDExtensionVariantPtrFromReflectValue(ret[0], ptr)
			return v
		default:
			log.Panic("unexpected MethodBindReturnStyle",
				zap.Any("value", ret),
			)
		}
		return NewVariantNil()
	} else {
		args := reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs(inst, callArgs, exepctedTypes)
		log.Debug("Calling",
			zap.String("bind", md.String()),
			zap.String("gd_args", VariantSliceToString(gdArgs)),
			zap.String("resolved_args", VariantSliceToString(callArgs)),
		)
		ret := md.Func.Call(args)
		log.Info("Call",
			zap.String("bind", md.String()),
			zap.String("gd_args", VariantSliceToString(gdArgs)),
			zap.String("resolved_args", VariantSliceToString(callArgs)),
			zap.String("ret", util.ReflectValueSliceToString(ret)),
		)
		switch md.GoReturnStyle {
		case NoneReturnStyle:
		case ValueAndBoolReturnStyle:
			log.Warn("second return value ignored")
			fallthrough
		case ValueReturnStyle:
			v := Variant{}
			ptr := (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(v.NativePtr()))
			pnr.Pin(ptr)
			GDExtensionVariantPtrFromReflectValue(ret[0], ptr)
			return v
		default:
			log.Panic("unexpected MethodBindReturnStyle",
				zap.Any("value", ret),
			)
		}
	}
	return NewVariantNil()
}

// Ptrcall is called by GDScript to call into Go
func (md *GoMethodMetadata) Ptrcall(inst GDClass, gdArgs []GDExtensionConstTypePtr, rReturn GDExtensionUninitializedTypePtr) {
	exepctedArgTypes := md.GoArgumentTypes
	args := reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs(inst, gdArgs, exepctedArgTypes)
	ret := md.Func.Call(args)
	log.Debug("Ptrcall",
		zap.String("bind", md.String()),
		zap.String("ret", util.ReflectValueSliceToString(ret)),
	)
	if err := validateReturnValues(ret, md.GoReturnStyle, md.GoReturnType); err != nil {
		log.Panic("return error",
			zap.Error(err),
		)
	}
	switch md.GoReturnStyle {
	case NoneReturnStyle:
	case ValueAndBoolReturnStyle:
		log.Warn("second return value ignored")
		fallthrough
	case ValueReturnStyle:
		GDExtensionTypePtrFromReflectValue(ret[0], rReturn)
	default:
		log.Panic("unexpected MethodBindReturnStyle",
			zap.Any("value", ret),
		)
	}
}

func (md *GoMethodMetadata) String() string {
	var sb strings.Builder
	sb.WriteString("MethodBind:")
	if md.ClassName != "" {
		sb.WriteString(md.ClassName)
		sb.WriteString(".")
	}
	sb.WriteString(md.GoMethodName)
	sb.WriteString("(")
	for i := range md.GoArgumentTypes {
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(md.GoArgumentTypes[i].Name())
		sb.WriteString("/")
		sb.WriteString(strconv.Itoa(int(md.gdeArgumentTypes[i])))

	}
	sb.WriteString(")")
	if md.GoReturnType != nil {
		sb.WriteString(" ")
		sb.WriteString(md.GoReturnType.Name())
	}

	return sb.String()
}

func NewGDExtensionClassMethodInfoFromMethodBind(md *GoMethodMetadata) *GDExtensionClassMethodInfo {
	var (
		argumentInfosPtr       *GDExtensionPropertyInfo
		argumentInfoCount      uint32
		argumentsMetadataPtr   *GDExtensionClassMethodArgumentMetadata
		defaultArgumentPtrsPtr *GDExtensionVariantPtr
		defaultArgumentCount   uint32
	)
	returnPropertyInfoPtr := &md.gdeReturnPropertyInfo
	pnr.Pin(returnPropertyInfoPtr)
	if md.IsVariadic {
		argumentsInfo := []GDExtensionPropertyInfo{
			NewSimpleGDExtensionPropertyInfo(md.ClassName, GDEXTENSION_VARIANT_TYPE_NIL, "varargs"),
		}
		argumentInfoCount = (uint32)(len(argumentsInfo))
		argumentInfosPtr = unsafe.SliceData(argumentsInfo)
		argumentsMetadata := []GDExtensionClassMethodArgumentMetadata{
			GDEXTENSION_METHOD_ARGUMENT_METADATA_NONE,
		}
		argumentsMetadataPtr = unsafe.SliceData(argumentsMetadata)
		defaultArgumentCount = (uint32)(len(md.gdeDefaultArgumentPtrs))
		defaultArgumentPtrsPtr = unsafe.SliceData(md.gdeDefaultArgumentPtrs)
		log.Debug("Create Variadic ClassMethodInfoFromMethodBind",
			zap.String("bind", md.String()),
		)
	} else {
		argumentInfoCount = (uint32)(len(md.gdeArgumentsInfo))
		argumentInfosPtr = unsafe.SliceData(md.gdeArgumentsInfo)
		argumentsMetadataPtr = unsafe.SliceData(md.gdeArgumentsMetadata)
		defaultArgumentCount = (uint32)(len(md.gdeDefaultArgumentPtrs))
		defaultArgumentPtrsPtr = unsafe.SliceData(md.gdeDefaultArgumentPtrs)
	}
	pnr.Pin(argumentInfosPtr)
	pnr.Pin(argumentsMetadataPtr)
	pnr.Pin(defaultArgumentPtrsPtr)
	for i := range md.gdeDefaultArgumentPtrs {
		pnr.Pin(md.gdeDefaultArgumentPtrs[i])
	}
	ret := NewGDExtensionClassMethodInfo(
		NewStringNameWithLatin1Chars(md.GdMethodName).AsGDExtensionConstStringNamePtr(),
		unsafe.Pointer(cgo.NewHandle(md)),
		(GDExtensionClassMethodCall)(C.cgo_method_bind_method_call),
		(GDExtensionClassMethodPtrCall)(C.cgo_method_bind_method_ptrcall),
		(uint32)(md.MethodFlags),
		md.GoReturnStyle != NoneReturnStyle,
		returnPropertyInfoPtr,
		GDEXTENSION_METHOD_ARGUMENT_METADATA_NONE,
		argumentInfoCount,
		(*GDExtensionPropertyInfo)(argumentInfosPtr),
		(*GDExtensionClassMethodArgumentMetadata)(argumentsMetadataPtr),
		defaultArgumentCount,
		(*GDExtensionVariantPtr)(defaultArgumentPtrsPtr),
	)
	pnr.Pin(ret)
	return ret
}
