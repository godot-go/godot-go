package gdextension

// #include <godot/gdextension_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"strings"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
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

type MethodMetadata struct {
	Func                reflect.Value
	GoReturnType        reflect.Type
	GoReturnStyle       ReturnStyle
	GoArgumentTypes     []reflect.Type
	ReturnType          GDExtensionVariantType
	ReturnPropertyInfo  GDExtensionPropertyInfo
	ArgumentsInfo       []GDExtensionPropertyInfo
	ArgumentsMetadata   []GDExtensionClassMethodArgumentMetadata
	ArgumentTypes       []GDExtensionVariantType
	DefaultArguments    []Variant
	DefaultArgumentPtrs []GDExtensionVariantPtr
	IsVariadic          bool
	IsVirtual           bool
	MethodFlags         MethodFlags
}

func NewMethodMetadata(
	method reflect.Method,
	className string,
	methodName string,
	argumentNames []string,
	defaultArguments []Variant,
	methodFlags MethodFlags,
) *MethodMetadata {
	mt := method.Type
	fn := method.Func
	recv := mt.In(0)
	if recv.Kind() == reflect.Pointer {
		recv = recv.Elem()
	}
	if className != recv.Name() {
		log.Panic("class name did not match reciever type",
			zap.String("class", className),
			zap.String("method", methodName),
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
			zap.String("method", methodName),
		)
	}
	var (
		goReturnType       reflect.Type
		returnStyle        ReturnStyle
		returnPropertyInfo GDExtensionPropertyInfo
	)
	switch returnCount {
	case 0:
	case 1, 2:
		// HACK: the second return value is ignored
		rt0 := mt.Out(0)
		goReturnType = rt0
		returnStyle = ValueReturnStyle
	default:
		log.Panic("method cannot return more than 1 value",
			zap.String("method", methodName),
		)
	}
	log.Debug("return value type",
		zap.Any("type", goReturnType),
	)
	returnType := ReflectTypeToGDExtensionVariantType(goReturnType)
	if returnType != GDEXTENSION_VARIANT_TYPE_NIL {
		returnPropertyInfo = NewSimpleGDExtensionPropertyInfo(className, returnType, goReturnType.Name())
	}
	argumentCount := mt.NumIn() - 1
	if len(argumentNames) > argumentCount {
		log.Panic(`Method definition has more arguments than the actual method.`,
			zap.String("method", methodName),
		)
	}
	defaultArgumentPtrs := make([]GDExtensionVariantPtr, len(defaultArguments))
	for i := range defaultArgumentPtrs {
		defaultArgumentPtrs[i] = (GDExtensionVariantPtr)(defaultArguments[i].nativePtr())
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
	return &MethodMetadata{
		Func:                fn,
		GoReturnType:        goReturnType,
		GoReturnStyle:       returnStyle,
		GoArgumentTypes:     goArgumentTypes,
		ReturnType:          returnType,
		ReturnPropertyInfo:  returnPropertyInfo,
		ArgumentsInfo:       argumentsInfo,
		ArgumentsMetadata:   argumentsMetadata,
		ArgumentTypes:       variantTypes,
		DefaultArguments:    defaultArguments,
		DefaultArgumentPtrs: defaultArgumentPtrs,
		IsVariadic:          isVariadicFlaged,
		IsVirtual:           isVirtual,
		MethodFlags:         methodFlags,
	}
}

// CallFunc is the function signature that can be called from GDScript
type VarargCallFunc func(GDClass, ...Variant) Variant

type MethodBindImpl struct {
	ClassName      string
	MethodName     string
	GoMethodName   string
	MethodMetadata MethodMetadata
	PtrcallFunc    reflect.Value
}

// Call is to be called by GDScript
func (b *MethodBindImpl) Call(
	inst GDClass,
	gdArgs []Variant,
) Variant {
	md := b.MethodMetadata
	gdArgsCount := len(gdArgs)
	defArgsCount := len(md.DefaultArgumentPtrs)
	callArgs := make([]Variant, len(md.ArgumentTypes))
	for i := range callArgs {
		if i < gdArgsCount {
			callArgs[i] = gdArgs[i]
		} else if i < defArgsCount {
			callArgs[i] = md.DefaultArguments[i]
		} else {
			log.Panic("too few arguments",
				zap.String("bind", b.String()),
				zap.String("gd_args", VariantSliceToString(gdArgs)),
				zap.String("defaults", VariantSliceToString(md.DefaultArguments)),
			)
		}
	}
	exepctedTypes := b.MethodMetadata.GoArgumentTypes
	if md.IsVariadic {
		args := []reflect.Value{
			reflect.ValueOf(inst),
			reflect.ValueOf(gdArgs),
		}
		ret := b.PtrcallFunc.CallSlice(args)
		log.Info("Call Variadic",
			zap.String("bind", b.String()),
			zap.String("gd_args", VariantSliceToString(gdArgs)),
			zap.String("resolved_args", VariantSliceToString(callArgs)),
			zap.String("ret", util.ReflectValueSliceToString(ret)),
		)
		switch b.MethodMetadata.GoReturnStyle {
		case NoneReturnStyle:
		case ValueAndBoolReturnStyle:
			log.Warn("second return value ignored")
			fallthrough
		case ValueReturnStyle:
			// return ret[0].Interface().(Variant)
			v := Variant{}
			ptr := (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(v.nativePtr()))
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
			zap.String("bind", b.String()),
			zap.String("gd_args", VariantSliceToString(gdArgs)),
			zap.String("resolved_args", VariantSliceToString(callArgs)),
		)
		ret := b.PtrcallFunc.Call(args)
		log.Info("Call",
			zap.String("bind", b.String()),
			zap.String("gd_args", VariantSliceToString(gdArgs)),
			zap.String("resolved_args", VariantSliceToString(callArgs)),
			zap.String("ret", util.ReflectValueSliceToString(ret)),
		)
		switch b.MethodMetadata.GoReturnStyle {
		case NoneReturnStyle:
		case ValueAndBoolReturnStyle:
			log.Warn("second return value ignored")
			fallthrough
		case ValueReturnStyle:
			v := Variant{}
			ptr := (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(v.nativePtr()))
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

func (b *MethodBindImpl) Ptrcall(
	inst GDClass,
	gdArgs []GDExtensionConstTypePtr,
	rReturn GDExtensionUninitializedTypePtr,
) {
	exepctedTypes := b.MethodMetadata.GoArgumentTypes
	args := reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs(inst, gdArgs, exepctedTypes)
	ret := b.PtrcallFunc.Call(args)
	// log.Info("Ptrcall",
	// 	zap.String("bind", b.String()),
	// 	zap.String("ret", util.ReflectValueSliceToString(ret)),
	// )
	if err := validateReturnValues(ret, b.MethodMetadata.GoReturnStyle, b.MethodMetadata.GoReturnType); err != nil {
		log.Panic("return error",
			zap.Error(err),
		)
	}
	switch b.MethodMetadata.GoReturnStyle {
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

func (b *MethodBindImpl) String() string {
	var sb strings.Builder
	sb.WriteString("MethodBind:")
	if b.ClassName != "" {
		sb.WriteString(b.ClassName)
		sb.WriteString(".")
	}
	sb.WriteString(b.GoMethodName)
	sb.WriteString("(")
	for i := range b.MethodMetadata.GoArgumentTypes {
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(b.MethodMetadata.GoArgumentTypes[i].Name())

	}
	sb.WriteString(")")
	if b.MethodMetadata.GoReturnType != nil {
		sb.WriteString(" ")
		sb.WriteString(b.MethodMetadata.GoReturnType.Name())
	}

	return sb.String()
}

func NewMethodBind(
	className string,
	methodName string,
	goMethodName string,
	methodMetadata MethodMetadata,
	ptrcallFunc reflect.Value,
) *MethodBindImpl {
	return &MethodBindImpl{
		ClassName:      className,
		MethodName:     methodName,
		GoMethodName:   goMethodName,
		MethodMetadata: methodMetadata,
		PtrcallFunc:    ptrcallFunc,
	}
}

func NewGDExtensionClassMethodInfoFromMethodBind(mb *MethodBindImpl) *GDExtensionClassMethodInfo {
	md := mb.MethodMetadata
	if md.IsVariadic {
		argumentsInfo := []GDExtensionPropertyInfo{
			NewSimpleGDExtensionPropertyInfo(mb.ClassName, GDEXTENSION_VARIANT_TYPE_NIL, "varargs"),
		}
		argumentsMetadata := []GDExtensionClassMethodArgumentMetadata{
			GDEXTENSION_METHOD_ARGUMENT_METADATA_NONE,
		}
		log.Info("Create Variadic ClassMethodInfoFromMethodBind",
			zap.String("bind", mb.String()),
		)
		return NewGDExtensionClassMethodInfo(
			NewStringNameWithLatin1Chars(mb.MethodName).AsGDExtensionConstStringNamePtr(),
			unsafe.Pointer(mb),
			(GDExtensionClassMethodCall)(C.cgo_method_bind_method_call),
			(GDExtensionClassMethodPtrCall)(C.cgo_method_bind_method_ptrcall),
			(uint32)(md.MethodFlags),
			md.GoReturnStyle != NoneReturnStyle,
			&md.ReturnPropertyInfo,
			GDEXTENSION_METHOD_ARGUMENT_METADATA_NONE,
			(uint32)(1),
			(*GDExtensionPropertyInfo)(unsafe.SliceData(argumentsInfo)),
			(*GDExtensionClassMethodArgumentMetadata)(unsafe.SliceData(argumentsMetadata)),
			(uint32)(len(md.DefaultArgumentPtrs)),
			(*GDExtensionVariantPtr)(unsafe.Pointer((unsafe.SliceData(md.DefaultArgumentPtrs)))),
		)
	}
	log.Info("Create Normal ClassMethodInfoFromMethodBind",
		zap.String("bind", mb.String()),
	)
	return NewGDExtensionClassMethodInfo(
		NewStringNameWithLatin1Chars(mb.MethodName).AsGDExtensionConstStringNamePtr(),
		unsafe.Pointer(mb),
		(GDExtensionClassMethodCall)(C.cgo_method_bind_method_call),
		(GDExtensionClassMethodPtrCall)(C.cgo_method_bind_method_ptrcall),
		(uint32)(md.MethodFlags),
		md.GoReturnStyle != NoneReturnStyle,
		&md.ReturnPropertyInfo,
		GDEXTENSION_METHOD_ARGUMENT_METADATA_NONE,
		(uint32)(len(md.ArgumentTypes)),
		(*GDExtensionPropertyInfo)(unsafe.SliceData(md.ArgumentsInfo)),
		(*GDExtensionClassMethodArgumentMetadata)(unsafe.SliceData(md.ArgumentsMetadata)),
		(uint32)(len(md.DefaultArgumentPtrs)),
		(*GDExtensionVariantPtr)(unsafe.Pointer((unsafe.SliceData(md.DefaultArgumentPtrs)))),
	)
}
