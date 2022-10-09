package gdextension

// #include <godot/gdextension_interface.h>
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"reflect"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
	. "github.com/godot-go/godot-go/pkg/gdnative"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

const (
	sizeOfGDExtensionPropertyInfo = int(unsafe.Sizeof(GDExtensionPropertyInfo{}))
)

type MethodBind struct {
	ClassMethodInfo    GDExtensionClassMethodInfo
	Name               string
	InstanceClass      string
	GoName             string
	GoReturnType       reflect.Type
	GoArgumentTypes  []reflect.Type
	ArgumentTypes    []GDExtensionVariantType
	DefaultArguments []*Variant
	Func               reflect.Value
}

func NewMethodBind(
	p_method reflect.Method,
	p_name string,
	p_argumentNames []string,
	p_defaultArguments []*Variant,
	p_methodFlags MethodFlags,
) *MethodBind {

	mt := p_method.Type

	fn := p_method.Func

	recv := mt.In(0)

	if recv.Kind() == reflect.Pointer {
		recv = recv.Elem()
	}

	instanceClass := recv.Name()

	returnCount := mt.NumOut()
	hasReturnValue := returnCount > 0

	argumentCount := mt.NumIn() - 1

	if returnCount > 2 {
		log.Panic("method cannot return more than 1 type", zap.String("name", p_name))
	}

	if len(p_argumentNames) > argumentCount {
		log.Panic(`Method definition has more arguments than the actual method.`, zap.String("name", p_name))
		return nil
	}

	var (
		goReturnValueType reflect.Type
		returnValueType GDExtensionVariantType
		returnValueInfo GDExtensionPropertyInfo
	)

	switch returnCount {
	case 0:
		goReturnValueType = nil
	case 1:
		goReturnValueType = mt.Out(0)
	case 2:
		goReturnValueType = mt.Out(0)

		// if there are 2 return values, the second one must be an error
		errRt := mt.Out(1)
		_, ok := reflect.New(errRt).Interface().(*error)

		if !ok {
			log.Panic("second return type can only be error",
				zap.String("name", p_name),
				zap.Any("type", errRt),
			)
		}
	default:
		log.Panic("method cannot return more than 2 values", zap.String("name", p_name))
	}

	returnValueType = ReflectTypeToGDExtensionVariantType(goReturnValueType)

	if returnValueType != GDEXTENSION_VARIANT_TYPE_NIL {
		returnValueInfo = NewSimpleGDExtensionPropertyInfo(instanceClass, returnValueType, goReturnValueType.Name())
	}

	goArgumentTypes := make([]reflect.Type, argumentCount)
	variantTypes := make([]GDExtensionVariantType, argumentCount)
	argumentsInfo := AllocArrayPtr[GDExtensionPropertyInfo](argumentCount)
	argumentsMetadata := AllocArrayPtr[GDExtensionClassMethodArgumentMetadata](argumentCount)

	argumentsInfoSlice := unsafe.Slice(argumentsInfo, argumentCount)
	argumentsMetadataSlice := unsafe.Slice(argumentsMetadata, argumentCount)

	for i := 0; i < argumentCount; i++ {
		t := mt.In(i + 1)

		goArgumentTypes[i] = t

		variantTypes[i] = ReflectTypeToGDExtensionVariantType(t)

		argumentsInfoSlice[i] = NewSimpleGDExtensionPropertyInfo(instanceClass, variantTypes[i], t.Name())
		argumentsMetadataSlice[i] = GDEXTENSION_METHOD_ARGUMENT_METADATA_NONE

		// *((*GDExtensionPropertyInfo)(unsafe.Add(unsafe.Pointer(argumentsInfo), uintptr(sizeOfGDExtensionPropertyInfo * i)))) = NewSimpleGDExtensionPropertyInfo(instanceClass, variantTypes[i], t.Name())
		// *((*GDExtensionClassMethodArgumentMetadata)(unsafe.Add(unsafe.Pointer(argumentsMetadata), uintptr(sizeOfGDExtensionPropertyInfo * i)))) = GDEXTENSION_EXTENSION_METHOD_ARGUMENT_METADATA_NONE
	}

	methodBind := &MethodBind{
		Name:               p_name,
		InstanceClass:      instanceClass,
		GoName:             p_method.Name,
		GoReturnType:       goReturnValueType,
		GoArgumentTypes:    goArgumentTypes,
		ArgumentTypes:      variantTypes,
		DefaultArguments:   p_defaultArguments,
		Func:               fn,
	}

	classMethodInfo := NewGDExtensionClassMethodInfo(
		NewStringNameWithLatin1Chars(p_name).AsGDExtensionStringNamePtr(),
		unsafe.Pointer(methodBind),
		(GDExtensionClassMethodCall)(C.cgo_method_bind_method_call),
		(GDExtensionClassMethodPtrCall)(C.cgo_method_bind_method_ptrcall),
		(uint32)(p_methodFlags),
		hasReturnValue,
		&returnValueInfo,
	    GDEXTENSION_METHOD_ARGUMENT_METADATA_NONE,
		(uint32)(argumentCount),
		argumentsInfo,
		argumentsMetadata,
		(uint32)(len(p_defaultArguments)),
		SliceHeaderDataPtr[Variant, GDExtensionVariantPtr](p_defaultArguments),
	)

	methodBind.ClassMethodInfo = classMethodInfo

	return methodBind
}

func (b *MethodBind) GetClassInfo() *ClassInfo {
	ci, ok := gdRegisteredGDClasses.Get(b.InstanceClass)

	if !ok {
		log.Panic("instance class not found", zap.String("instance_class", b.InstanceClass))
	}

	return ci
}

func (b *MethodBind) Call(
	p_instance GDExtensionClassInstancePtr,
	p_args *GDExtensionConstVariantPtr,
	p_argument_count GDExtensionInt,
	r_error *GDExtensionCallError,
) Variant {
	if int(p_argument_count) > len(b.GoArgumentTypes) {
		r_error.SetErrorFields(
			GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS,
			int32(p_argument_count),
			int32(len(b.GoArgumentTypes)),
		)
		return Variant{}
	}

	w := *(*Wrapped)(unsafe.Pointer(&p_instance))
	vInstance := NewVariantWrapped(w)

	vReturn := NewVariantNil()
	var (
		r_return GDExtensionVariantPtr = (GDExtensionVariantPtr)(vReturn.ptr())
	)

	GDExtensionInterface_variant_call(
		internal.gdnInterface,
		(GDExtensionVariantPtr)(&vInstance),
		NewStringNameWithLatin1Chars(b.Name).AsGDExtensionStringNamePtr(),
		p_args,
		p_argument_count,
		r_return,
		r_error,
	)

	return vReturn
}

// NOTE: i think this was meant for built-in functions calls
func (b *MethodBind) CallStatic(
	p_instance GDExtensionClassInstancePtr,
	p_args *GDExtensionConstVariantPtr,
	p_argument_count GDExtensionInt,
	r_error *GDExtensionCallError,
) Variant {
	if int(p_argument_count) > len(b.GoArgumentTypes) {
		r_error.SetErrorFields(
			GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS,
			int32(p_argument_count),
			int32(len(b.GoArgumentTypes)),
		)
		return Variant{}
	}

	w := *(*Wrapped)(unsafe.Pointer(&p_instance))
	vInstance := NewVariantWrapped(w)
	vt := vInstance.GetType()

	vReturn := NewVariantNil()
	var (
		r_return GDExtensionVariantPtr = (GDExtensionVariantPtr)(vReturn.ptr())
	)

	GDExtensionInterface_variant_call_static(
		internal.gdnInterface,
		vt,
		NewStringNameWithLatin1Chars(b.Name).AsGDExtensionStringNamePtr(),
		p_args,
		p_argument_count,
		r_return,
		r_error,
	)

	return vReturn
}

func (b *MethodBind) Ptrcall(
	p_instance GDExtensionClassInstancePtr,
	p_args *GDExtensionConstTypePtr,
	r_ret *GDExtensionTypePtr,
) {
	ci := b.GetClassInfo()

	pArgArraySlice := unsafe.Slice(p_args, len(b.GoArgumentTypes))

	// add an extra parameter for the receiver instance
	args := make([]reflect.Value, len(b.GoArgumentTypes)+1)

	inst := GDClassFromGDExtensionClassInstancePtr(ci, p_instance)

	args[0] = reflect.ValueOf(inst)

	for i := 0; i < len(b.GoArgumentTypes); i++ {
		arg := pArgArraySlice[i]

		t := b.GoArgumentTypes[i]

		switch t.Kind() {
		case reflect.Bool:
			args[i+1] = reflect.ValueOf(*(*bool)(arg))
		case reflect.Int:
			args[i+1] = reflect.ValueOf(*(*int)(arg))
		case reflect.Int8:
			args[i+1] = reflect.ValueOf(*(*int8)(arg))
		case reflect.Int16:
			args[i+1] = reflect.ValueOf(*(*int16)(arg))
		case reflect.Int32:
			args[i+1] = reflect.ValueOf(*(*int32)(arg))
		case reflect.Int64:
			args[i+1] = reflect.ValueOf(*(*int64)(arg))
		case reflect.Uint:
			args[i+1] = reflect.ValueOf(*(*uint)(arg))
		case reflect.Uint8:
			args[i+1] = reflect.ValueOf(*(*uint8)(arg))
		case reflect.Uint16:
			args[i+1] = reflect.ValueOf(*(*uint16)(arg))
		case reflect.Uint32:
			args[i+1] = reflect.ValueOf(*(*uint32)(arg))
		case reflect.Uint64:
			args[i+1] = reflect.ValueOf(*(*uint64)(arg))
		case reflect.Float32:
			args[i+1] = reflect.ValueOf(*(*float32)(arg))
		case reflect.Float64:
			args[i+1] = reflect.ValueOf(*(*float64)(arg))
		case reflect.Struct:
			switch reflect.New(t).Interface().(type) {
			case String:
				args[i+1] = reflect.ValueOf(*(*String)(arg))
			}
		case reflect.String:
			log.Panic("MethodBind.Ptrcall does not support native go string type")
		}
	}

	log.Debug("argument converted",
		zap.String("args", spew.Sdump(args)),
	)

	reflectedRet := b.Func.Call(args)

	switch len(reflectedRet) {
	case 0:
		if b.GoReturnType != nil {
			log.Panic("no return value expected")
		}
	case 1:
		if b.GoReturnType == nil {
			log.Panic("return value expected but none provided")
		}

		valueInterface := reflectedRet[0].Interface()

		*r_ret = *(*GDExtensionTypePtr)(unsafe.Pointer(&valueInterface))
	case 2:
		if b.GoReturnType == nil {
			log.Panic("return value expected but none provided")
		}

		valueInterface := reflectedRet[0].Interface()

		// 2nd value can only be of type error
		if !reflectedRet[1].IsNil() {
			err, ok := reflectedRet[1].Interface().(error)
			if !ok {
				log.Panic("second return value must be of type error",
					zap.String("type", reflectedRet[1].Type().Name()),
				)
			}

			if err != nil {
				log.Panic("error returned", zap.Error(err))
			}
		}

		if reflectedRet[0].IsNil() {
			log.Warn("returning nil value",
				zap.String("type", reflectedRet[0].Type().Name()),
			)
			*r_ret = *(*GDExtensionTypePtr)(unsafe.Pointer(&valueInterface))
		}
	default:
		log.Panic("too many values returned", zap.Any("ret", reflectedRet))
	}
}

func (b *MethodBind) Destroy() {
	// TODO: wire this up
	// b.ClassMethodInfo.Destroy()
}

// //export GoCallback_MethodBindBindGetArgumentType
// func GoCallback_MethodBindBindGetArgumentType(p_method_userdata unsafe.Pointer, p_argument int32) C.GDExtensionVariantType {
// 	bind := (*MethodBind)(p_method_userdata)
// 	return (C.GDExtensionVariantType)(bind.GoArgumentTypes[p_argument])
// }

// //export GoCallback_MethodBindBindGetArgumentInfo
// func GoCallback_MethodBindBindGetArgumentInfo(p_method_userdata unsafe.Pointer, p_argument int32, r_info unsafe.Pointer) {
// 	bind := (*MethodBind)(p_method_userdata)
// 	*((*GDExtensionPropertyInfo)(r_info)) = bind.GoArgumentsInfo[p_argument].ToGDExtensionPropertyInfo()
// }

// //export GoCallback_MethodBindBindGetArgumentMetadata
// func GoCallback_MethodBindBindGetArgumentMetadata(p_method_userdata unsafe.Pointer, p_argument int32) C.GDExtensionClassMethodArgumentMetadata {
// 	bind := (*MethodBind)(p_method_userdata)
// 	return (C.GDExtensionClassMethodArgumentMetadata)(bind.ArgumentsMetadata[p_argument])
// }

// this is called when methods in Go callback into GDScript
//
//export GoCallback_MethodBindBindCall
func GoCallback_MethodBindBindCall(p_method_userdata unsafe.Pointer, p_instance C.GDExtensionClassInstancePtr, p_args *C.GDExtensionVariantPtr, p_argument_count C.GDExtensionInt, r_return C.GDExtensionVariantPtr, r_error *C.GDExtensionCallError) {
	b := (*MethodBind)(p_method_userdata)

	ci := b.GetClassInfo()

	var vArgs *[MAX_ARG_COUNT]*Variant

	if p_args != nil {
		vArgs = (*[MAX_ARG_COUNT]*Variant)(unsafe.Pointer(p_args))
	}

	if int(p_argument_count) > len(b.GoArgumentTypes) {
		((*GDExtensionCallError)(unsafe.Pointer(r_error))).SetErrorFields(
			GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS,
			int32(p_argument_count),
			int32(len(b.GoArgumentTypes)),
		)
		return
	}

	// add an extra parameter for the receiver instance
	args := make([]reflect.Value, len(b.GoArgumentTypes)+1)

	inst := GDClassFromGDExtensionClassInstancePtr(ci, (GDExtensionClassInstancePtr)(p_instance))

	args[0] = reflect.ValueOf(inst)

	for i := 0; i < len(b.GoArgumentTypes); i++ {
		var varg *Variant
		if i < (int)(p_argument_count) {
			varg = vArgs[i]
		} else {
			varg = b.DefaultArguments[i]
		}

		args[i+1] = varg.ToReflectValue(
			b.ArgumentTypes[i],
			b.GoArgumentTypes[i],
		)
	}

	log.Debug("argument converted",
		zap.String("method_name", (string)(b.GoName)),
		zap.Int("args_cap", cap(args)),
		zap.Int("args_len", len(args)),
		zap.Any("func", spew.Sdump(b.Func)),
		// zapGDExtensionVariantPtrp("vargs", (*GDExtensionVariantPtr)(p_args), (int)(p_argument_count)),
		zap.String("args", spew.Sdump(args)),
	)

	reflectedRet := b.Func.Call(args)

	log.Debug("returned value", zap.Any("reflectedRet", reflectedRet))

	var ret Variant

	if len(reflectedRet) == 0 {
		if b.GoReturnType != nil {
			log.Panic("no return value expected")
		}

		ret = NewVariantNil()
		defer ret.Destroy()
	} else {
		if b.GoReturnType == nil {
			log.Panic("return value expected but none provided")
		}

		log.Debug("returned", zap.String("value", spew.Sdump(reflectedRet[0])))

		switch len(reflectedRet) {
		case 0:
			if b.GoReturnType != nil {
				log.Panic("no return value expected")
			}
		case 1:
			if b.GoReturnType == nil {
				log.Panic("return value expected but none provided")
			}

			ret = NewVariantFromReflectValue(reflectedRet[0])
			defer ret.Destroy()
		case 2:
			if b.GoReturnType == nil {
				log.Panic("return value expected but none provided")
			}

			// 2nd value can only be of type error
			// panic if an error is returned
			if !reflectedRet[1].IsNil() {
				err, ok := reflectedRet[1].Interface().(error)
				if !ok {
					log.Panic("second return value must be of type error",
						zap.String("type", reflectedRet[1].Type().Name()),
					)
				}

				if err != nil {
					log.Panic("error returned", zap.Error(err))
				}
			}

			if reflectedRet[0].IsNil() {
				log.Warn("returning nil value",
					zap.String("type", reflectedRet[0].Type().Name()),
				)
				ret = NewVariantNil()
				defer ret.Destroy()
			}

			ret = NewVariantFromReflectValue(reflectedRet[0])
			defer ret.Destroy()
		default:
			log.Panic("too many values returned",
				zap.Any("go_name", b.GoName),
				zap.Any("gd_name", b.Name),
			)
		}
	}
	log.Debug("pre-copy")

	// This assumes the return value is an empty Variant, so it doesn't need to call the destructor first.
	// Since only NativeExtensionMethodBind calls this from the Godot side, it should always be the case.
	GDExtensionInterface_variant_new_copy(internal.gdnInterface, (GDExtensionVariantPtr)(r_return), (GDExtensionConstVariantPtr)(ret.ptr()))

	log.Debug("end GoCallback_MethodBindBindCall")
}

//export GoCallback_MethodBindBindPtrcall
func GoCallback_MethodBindBindPtrcall(p_method_userdata unsafe.Pointer, p_instance C.GDExtensionClassInstancePtr, p_args *C.GDExtensionConstTypePtr, r_ret C.GDExtensionTypePtr) {
	bind := (*MethodBind)(p_method_userdata)
	bind.Ptrcall(
		(GDExtensionClassInstancePtr)(p_instance),
		(*GDExtensionConstTypePtr)(p_args),
		(*GDExtensionTypePtr)(&r_ret),
	)
}
