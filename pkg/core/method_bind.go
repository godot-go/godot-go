package core

// #include <godot/gdextension_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"runtime/cgo"
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
	// Lifecycle-managed StringName objects for PropertyInfo structs.
	// These persist for the lifetime of GoMethodMetadata to prevent
	// dangling pointers in GDExtensionPropertyInfo.name/class_name fields.
	// See Task 5.1 in fix-orphan-stringname change.
	gdeReturnPropNameStringName    StringName
	gdeReturnPropClassNameStringName StringName
	gdeReturnPropHintString        String
	gdeArgPropNameStringNames      []StringName
	gdeArgPropClassNameStringNames []StringName
	gdeArgPropHintStrings          []String
	// Lifecycle-managed StringName for the GD method name.
	gdeMethodNameStringName StringName
	// Lifecycle-managed StringName/String for variadic (varargs) argument PropertyInfo.
	// Only set when IsVariadic is true.
	gdeVarArgPropClassNameStringName StringName
	gdeVarArgPropNameStringName      StringName
	gdeVarArgPropHintString          String
}

// Destroy cleans up all StringName and String objects stored in this metadata.
// Must be called once during class unregistration to prevent orphan StringNames.
// IMPORTANT: Each StringName/String is copied to a local variable and pinned
// before passing to cgo functions. This satisfies the cgo "Go pointer to unpinned
// Go pointer" check, since GoMethodMetadata contains reflect.Value (Go pointer).
func (md *GoMethodMetadata) Destroy() {

	// Only destroy return value PropertyInfo if it was created (non-NIL return type).
	if md.gdeReturnType != GDEXTENSION_VARIANT_TYPE_NIL {
		retClassName := md.gdeReturnPropClassNameStringName
		retName := md.gdeReturnPropNameStringName
		retHint := md.gdeReturnPropHintString
		pnr.Pin(&retClassName)
		pnr.Pin(&retName)
		pnr.Pin(&retHint)
		retClassName.Destroy()
		retName.Destroy()
		retHint.Destroy()
	}

	// Destroy argument PropertyInfo StringName/String objects.
	for i := range md.gdeArgPropClassNameStringNames {
		className := md.gdeArgPropClassNameStringNames[i]
		pnr.Pin(&className)
		className.Destroy()
	}
	md.gdeArgPropClassNameStringNames = nil
	for i := range md.gdeArgPropNameStringNames {
		name := md.gdeArgPropNameStringNames[i]
		pnr.Pin(&name)
		name.Destroy()
	}
	md.gdeArgPropNameStringNames = nil
	for i := range md.gdeArgPropHintStrings {
		hint := md.gdeArgPropHintStrings[i]
		pnr.Pin(&hint)
		hint.Destroy()
	}
	md.gdeArgPropHintStrings = nil

	// Destroy the GD method name StringName.
	methodName := md.gdeMethodNameStringName
	pnr.Pin(&methodName)
	methodName.Destroy()

	// Destroy variadic argument StringName/String objects (if variadic method).
	if md.IsVariadic {
		varClassName := md.gdeVarArgPropClassNameStringName
		varName := md.gdeVarArgPropNameStringName
		varHint := md.gdeVarArgPropHintString
		pnr.Pin(&varClassName)
		pnr.Pin(&varName)
		pnr.Pin(&varHint)
		varClassName.Destroy()
		varName.Destroy()
		varHint.Destroy()
	}
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
	// Track StringName objects for return value PropertyInfo lifecycle management.
	// These persist in GoMethodMetadata to prevent dangling pointers.
	var (
		returnPropNameStringName    StringName
		returnPropClassNameStringName StringName
		returnPropHintString        String
	)
	if returnType != GDEXTENSION_VARIANT_TYPE_NIL {
		returnPropClassNameStringName = NewStringNameWithLatin1Chars(className)
		returnPropNameStringName = NewStringNameWithLatin1Chars(goReturnType.Name())
		returnPropHintString = NewStringWithUtf8Chars("")
		returnPropertyInfo = NewGDExtensionPropertyInfoFromNames(
			&returnPropClassNameStringName,
			returnType,
			&returnPropNameStringName,
			&returnPropHintString,
		)
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
	// Track StringName objects for argument PropertyInfo lifecycle management.
	// These persist in GoMethodMetadata to prevent dangling pointers.
	argPropNameStringNames := make([]StringName, argumentCount)
	argPropClassNameStringNames := make([]StringName, argumentCount)
	argPropHintStrings := make([]String, argumentCount)
	for i := 0; i < argumentCount; i++ {
		t := mt.In(i + 1)
		goArgumentTypes[i] = t
		variantTypes[i] = ReflectTypeToGDExtensionVariantType(t)
		argPropClassNameStringNames[i] = NewStringNameWithLatin1Chars(className)
		argPropNameStringNames[i] = NewStringNameWithLatin1Chars(t.Name())
		argPropHintStrings[i] = NewStringWithUtf8Chars("")
		argumentsInfo[i] = NewGDExtensionPropertyInfoFromNames(
			&argPropClassNameStringNames[i],
			variantTypes[i],
			&argPropNameStringNames[i],
			&argPropHintStrings[i],
		)
		argumentsMetadata[i] = GDEXTENSION_METHOD_ARGUMENT_METADATA_NONE
	}
	ret := &GoMethodMetadata{
		ClassName:                    className,
		GdMethodName:                 gdMethodName,
		GoMethodName:                 goMethodName,
		Func:                         fn,
		GoReturnType:                 goReturnType,
		GoReturnStyle:                returnStyle,
		GoArgumentTypes:              goArgumentTypes,
		DefaultArguments:             defaultArguments,
		IsVariadic:                   isVariadicFlaged,
		IsVirtual:                    isVirtual,
		MethodFlags:                  methodFlags,
		gdeReturnType:                returnType,
		gdeReturnPropertyInfo:        returnPropertyInfo,
		gdeArgumentsInfo:             argumentsInfo,
		gdeArgumentsMetadata:         argumentsMetadata,
		gdeArgumentTypes:             variantTypes,
		gdeDefaultArgumentPtrs:       defaultArgumentPtrs,
		gdeReturnPropNameStringName:  returnPropNameStringName,
		gdeReturnPropClassNameStringName: returnPropClassNameStringName,
		gdeReturnPropHintString:      returnPropHintString,
		gdeArgPropNameStringNames:    argPropNameStringNames,
		gdeArgPropClassNameStringNames: argPropClassNameStringNames,
		gdeArgPropHintStrings:        argPropHintStrings,
		gdeMethodNameStringName:      NewStringNameWithLatin1Chars(gdMethodName),
	}
	// Create variadic argument PropertyInfo StringNames for variadic methods.
	// These persist in GoMethodMetadata for lifecycle management.
	if isVariadicFlaged {
		ret.gdeVarArgPropClassNameStringName = NewStringNameWithLatin1Chars(className)
		ret.gdeVarArgPropNameStringName = NewStringNameWithLatin1Chars("varargs")
		ret.gdeVarArgPropHintString = NewStringWithUtf8Chars("")
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
		for _, arg := range args {
			switch v := arg.Interface().(type) {
			case StringName:
				v.Destroy()
			case NodePath:
				v.Destroy()
			}
		}
		log.Info("Call",
			zap.String("bind", md.String()),
			zap.String("gd_args", VariantSliceToString(gdArgs)),
			zap.String("resolved_args", VariantSliceToString(callArgs)),
			zap.String("ret", util.ReflectValueSliceToString(ret)),
		)
		var retVariant Variant
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
			retVariant = v
		default:
			log.Panic("unexpected MethodBindReturnStyle",
				zap.Any("value", ret),
			)
		}
		// Release the owned container arguments now that the return value (which
		// may encode an echoed container argument into a new Variant) has been
		// produced; the encoded return holds its own reference.
		destroyOwnedContainerArgs(args)
		if md.GoReturnStyle == NoneReturnStyle {
			return NewVariantNil()
		}
		return retVariant
	}
}

// Ptrcall is called by GDScript to call into Go
func (md *GoMethodMetadata) Ptrcall(inst GDClass, gdArgs []GDExtensionConstTypePtr, rReturn GDExtensionUninitializedTypePtr) {
	exepctedArgTypes := md.GoArgumentTypes
	args := reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs(inst, gdArgs, exepctedArgTypes)
	ret := md.Func.Call(args)
	log.Info("Ptrcall",
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
		GDExtensionTypePtrFromReflectValue(ret[0], rReturn, !isPtrcallBorrowEcho(ret[0], args))
	default:
		log.Panic("unexpected MethodBindReturnStyle",
			zap.Any("value", ret),
		)
	}
}

// isPtrcallBorrowEcho reports whether the ptrcall return value is a byte-for-byte
// echo of a decoded call argument. Decoded refcounted container arguments are
// borrows holding no refcount of their own, so echoing one back must not release
// a reference through the return path.
//
// The comparison runs on the wrapped dynamic values (args[i].Interface() and
// ret.Interface()), not on the reflect.Value wrappers themselves. If
// reflect.DeepEqual were applied to the two reflect.Values, it would compare
// the wrapper struct fields — type pointer, flag word, and data pointer — and
// those differ between a decoded argument slot and a reflected return value
// even when both wrap identical bytes. Comparing the wrappers would therefore
// never detect an echo, and the borrow would be wrongly destroyed.
func isPtrcallBorrowEcho(ret reflect.Value, args []reflect.Value) bool {
	// Only the container types whose ptrcall return encoding destroys the
	// source (see GDExtensionTypePtrFromReflectValue) can be borrow echoes.
	switch ret.Interface().(type) {
	case Array, PackedByteArray, PackedInt32Array, PackedInt64Array,
		PackedFloat32Array, PackedFloat64Array, PackedStringArray,
		PackedVector2Array, PackedVector3Array, PackedColorArray:
	default:
		return false
	}
	for i := 1; i < len(args); i++ {
		if args[i].Type() == ret.Type() && reflect.DeepEqual(args[i].Interface(), ret.Interface()) {
			return true
		}
	}
	return false
}

func (md *GoMethodMetadata) String() string {
	if md == nil {
		log.Panic("GoMethodMetadata cannot be null")
		return ""
	}
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
		sb.WriteString(GDExtensionVariantTypeStringMap[md.gdeArgumentTypes[i]])

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

	// Pin StringName objects used in return PropertyInfo to prevent GC movement
	// during the cgo call. These are stored in GoMethodMetadata and persist.
	if md.gdeReturnType != GDEXTENSION_VARIANT_TYPE_NIL {
		pnr.Pin(&md.gdeReturnPropNameStringName)
		pnr.Pin(&md.gdeReturnPropClassNameStringName)
		pnr.Pin(&md.gdeReturnPropHintString)
	}

	if md.IsVariadic {
		// Use variadic argument PropertyInfo StringNames from GoMethodMetadata.
		// Their lifecycle is managed by GoMethodMetadata.Destroy().
		pnr.Pin(&md.gdeVarArgPropClassNameStringName)
		pnr.Pin(&md.gdeVarArgPropNameStringName)
		pnr.Pin(&md.gdeVarArgPropHintString)
		argumentsInfo := []GDExtensionPropertyInfo{
			NewGDExtensionPropertyInfoFromNames(&md.gdeVarArgPropClassNameStringName, GDEXTENSION_VARIANT_TYPE_NIL, &md.gdeVarArgPropNameStringName, &md.gdeVarArgPropHintString),
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
		// Pin StringName objects used in argument PropertyInfo structs
		for i := range md.gdeArgPropNameStringNames {
			pnr.Pin(&md.gdeArgPropNameStringNames[i])
			pnr.Pin(&md.gdeArgPropClassNameStringNames[i])
			pnr.Pin(&md.gdeArgPropHintStrings[i])
		}
	}
	pnr.Pin(argumentInfosPtr)
	pnr.Pin(argumentsMetadataPtr)
	pnr.Pin(defaultArgumentPtrsPtr)
	for i := range md.gdeDefaultArgumentPtrs {
		pnr.Pin(md.gdeDefaultArgumentPtrs[i])
	}
	// Use the tracked method name StringName from GoMethodMetadata.
	// Its lifecycle is managed by GoMethodMetadata.Destroy().
	pnr.Pin(&md.gdeMethodNameStringName)
	ret := NewGDExtensionClassMethodInfo(
		md.gdeMethodNameStringName.AsGDExtensionConstStringNamePtr(),
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
