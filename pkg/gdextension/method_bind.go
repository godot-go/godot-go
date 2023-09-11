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
	Func               reflect.Value
	GoReturnType       reflect.Type
	GoReturnStyle      ReturnStyle
	GoArgumentTypes    []reflect.Type
	ReturnType         GDExtensionVariantType
	ReturnPropertyInfo GDExtensionPropertyInfo
	ArgumentsInfo      []GDExtensionPropertyInfo
	ArgumentsMetadata  []GDExtensionClassMethodArgumentMetadata
	ArgumentTypes      []GDExtensionVariantType
	DefaultArguments   []Variant // TODO: switch this to []*Variant to align with Godot
	IsVariadic         bool
}

func NewMethodMetadata(
	method reflect.Method,
	className string,
	methodName string,
	argumentNames []string,
	defaultArguments []Variant,
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
		Func:               fn,
		GoReturnType:       goReturnType,
		GoReturnStyle:      returnStyle,
		GoArgumentTypes:    goArgumentTypes,
		ReturnType:         returnType,
		ReturnPropertyInfo: returnPropertyInfo,
		ArgumentsInfo:      argumentsInfo,
		ArgumentsMetadata:  argumentsMetadata,
		ArgumentTypes:      variantTypes,
		DefaultArguments:   defaultArguments,
		IsVariadic:         mt.IsVariadic(),
	}
}

// CallFunc is the function signature that can be called from GDScript
type VarargCallFunc func(GDClass, ...Variant) Variant

type MethodBindImpl struct {
	ClassName      string
	MethodName     string
	GoMethodName   string
	MethodMetadata MethodMetadata
	CallFunc       reflect.Value
	PtrcallFunc    reflect.Value
}

// Call is to be called by GDScript
func (b *MethodBindImpl) Call(
	inst GDClass,
	gdArgs []Variant,
) Variant {
	md := b.MethodMetadata
	gdArgsCount := len(gdArgs)
	defArgsCount := len(md.DefaultArguments)
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
	args := reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs(inst, callArgs, exepctedTypes)
	ret := b.PtrcallFunc.Call(args)
	log.Info("Call",
		zap.String("bind", b.String()),
		zap.String("gd_args", VariantSliceToString(gdArgs)),
		zap.String("resolved_args", VariantSliceToString(callArgs)),
		zap.String("ret", util.ReflectValueSliceToString(ret)),
	)
	v := NewVariantNil()
	ptr := (GDExtensionVariantPtr)(unsafe.Pointer(v.ptr()))
	switch b.MethodMetadata.GoReturnStyle {
	case NoneReturnStyle:
	case ValueReturnStyle:
		GDExtensionVariantPtrFromReflectValue(ret[0], ptr)
		return v
	case ValueAndBoolReturnStyle:
		log.Warn("second return value ignored")
		GDExtensionVariantPtrFromReflectValue(ret[0], ptr)
	default:
		log.Panic("unexpected MethodBindReturnStyle",
			zap.Any("value", ret),
		)
	}
	return v
}

func (b *MethodBindImpl) Ptrcall(
	inst GDClass,
	gdArgs []GDExtensionConstTypePtr,
	rReturn GDExtensionTypePtr,
) {
	exepctedTypes := b.MethodMetadata.GoArgumentTypes
	args := reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs(inst, gdArgs, exepctedTypes)
	ret := b.PtrcallFunc.Call(args)
	log.Info("Ptrcall",
		zap.String("bind", b.String()),
		zap.String("ret", util.ReflectValueSliceToString(ret)),
	)
	if err := validateReturnValues(ret, b.MethodMetadata.GoReturnStyle, b.MethodMetadata.GoReturnType); err != nil {
		log.Panic("return error",
			zap.Error(err),
		)
	}
	switch b.MethodMetadata.GoReturnStyle {
	case NoneReturnStyle:
	case ValueReturnStyle:
		GDExtensionTypePtrFromReflectValue(ret[0], rReturn)
	case ValueAndBoolReturnStyle:
		log.Warn("second return value ignored")
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
	callFunc reflect.Value,
	ptrcallFunc reflect.Value,
) *MethodBindImpl {
	return &MethodBindImpl{
		ClassName:      className,
		MethodName:     methodName,
		GoMethodName:   goMethodName,
		MethodMetadata: methodMetadata,
		CallFunc:       callFunc,
		PtrcallFunc:    ptrcallFunc,
	}
}

func NewGDExtensionClassMethodInfoFromMethodBind(mb *MethodBindImpl) *GDExtensionClassMethodInfo {
	md := mb.MethodMetadata
	argumentCount := len(md.ArgumentTypes)
	// argumentsInfo := AllocArrayPtr[GDExtensionPropertyInfo](argumentCount)
	// argumentsMetadata := AllocArrayPtr[GDExtensionClassMethodArgumentMetadata](argumentCount)
	// argumentsInfo := make([]GDExtensionPropertyInfo, argumentCount)
	defArgPtrs := make([]GDExtensionVariantPtr, len(md.DefaultArguments))
	for i := range defArgPtrs {
		defArgPtrs[i] = (GDExtensionVariantPtr)(md.DefaultArguments[i].ptr())
	}
	classMethodInfo := NewGDExtensionClassMethodInfo(
		NewStringNameWithLatin1Chars(mb.MethodName).AsGDExtensionConstStringNamePtr(),
		unsafe.Pointer(mb),
		(GDExtensionClassMethodCall)(C.cgo_method_bind_method_call),
		(GDExtensionClassMethodPtrCall)(C.cgo_method_bind_method_ptrcall),
		(uint32)(METHOD_FLAGS_DEFAULT),
		md.GoReturnStyle != NoneReturnStyle,
		&md.ReturnPropertyInfo,
		GDEXTENSION_METHOD_ARGUMENT_METADATA_NONE,
		(uint32)(argumentCount),
		(*GDExtensionPropertyInfo)(unsafe.SliceData(md.ArgumentsInfo)),
		(*GDExtensionClassMethodArgumentMetadata)(unsafe.SliceData(md.ArgumentsMetadata)),
		(uint32)(len(defArgPtrs)),
		(*GDExtensionVariantPtr)(unsafe.Pointer((unsafe.SliceData(defArgPtrs)))),
	)
	return &classMethodInfo
}
