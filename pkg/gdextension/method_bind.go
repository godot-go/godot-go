package gdextension

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
	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

const (
	sizeOfGDExtensionPropertyInfo = int(unsafe.Sizeof(GDExtensionPropertyInfo{}))
)

type MethodBind struct {
	ClassMethodInfo  GDExtensionClassMethodInfo
	Name             string
	InstanceClass    string
	GoName           string
	GoReturnType     reflect.Type
	GoArgumentTypes  []reflect.Type
	ArgumentTypes    []GDExtensionVariantType
	DefaultArguments []*Variant
	Func             reflect.Value
}

func (b *MethodBind) String() string {
	var sb strings.Builder

	if b.InstanceClass != "" {
		sb.WriteString(b.InstanceClass)
		sb.WriteString(".")
	}
	sb.WriteString(b.GoName)
	sb.WriteString("(")
	for i := range b.GoArgumentTypes {
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(b.GoArgumentTypes[i].Name())

	}
	sb.WriteString(")")
	if b.GoReturnType != nil {
		sb.WriteString(" ")
		sb.WriteString(b.GoReturnType.Name())
	}

	return sb.String()
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
		returnValueType   GDExtensionVariantType
		returnValueInfo   GDExtensionPropertyInfo
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
		Name:             p_name,
		InstanceClass:    instanceClass,
		GoName:           p_method.Name,
		GoReturnType:     goReturnValueType,
		GoArgumentTypes:  goArgumentTypes,
		ArgumentTypes:    variantTypes,
		DefaultArguments: p_defaultArguments,
		Func:             fn,
	}

	classMethodInfo := NewGDExtensionClassMethodInfo(
		NewStringNameWithUtf8Chars(p_name).AsGDExtensionConstStringNamePtr(),
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
		(*GDExtensionVariantPtr)(unsafe.Pointer((unsafe.SliceData(p_defaultArguments)))),
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

	w := *(*Object)(unsafe.Pointer(&p_instance))
	vInstance := NewVariantObject(w)

	vReturn := NewVariantNil()
	var (
		r_return GDExtensionUninitializedVariantPtr = (GDExtensionUninitializedVariantPtr)(vReturn.ptr())
		bindName                                    = NewStringNameWithUtf8Chars(b.Name)
	)

	defer bindName.Destroy()

	CallFunc_GDExtensionInterfaceVariantCall(
		(GDExtensionVariantPtr)(&vInstance),
		bindName.AsGDExtensionConstStringNamePtr(),
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

	w := *(*Object)(unsafe.Pointer(&p_instance))
	vInstance := NewVariantObject(w)
	vt := vInstance.GetType()

	vReturn := NewVariantNil()
	var (
		r_return GDExtensionUninitializedVariantPtr = (GDExtensionUninitializedVariantPtr)(vReturn.ptr())
	)

	snName := NewStringNameWithUtf8Chars(b.Name)
	defer snName.Destroy()

	CallFunc_GDExtensionInterfaceVariantCallStatic(
		vt,
		snName.AsGDExtensionConstStringNamePtr(),
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
	r_ret GDExtensionTypePtr,
) {
	inst := ObjectClassFromGDExtensionClassInstancePtr(p_instance)
	if inst == nil {
		log.Panic("p_instance GDExtensionClassInstancePtr canoot be null")
	}
	cn := inst.GetClass()
	log.Info("MethodBind.Ptrcall called",
		zap.String("class_name", cn.ToUtf8()),
		zap.String("bind", b.String()),
	)
	args := b.reflectCallArgsFromGDExtensionArgs(inst, p_args)
	// call the go function
	reflectedRet := b.Func.Call(args)
	log.Info("reflect method called",
		zap.String("ret", valuesToString(reflectedRet)),
	)
	b.writeReturnValueFromReflectReturnValues(reflectedRet, r_ret)
}

func valuesToString(values []reflect.Value) string {
	var sb strings.Builder
	sb.WriteString("(")
	for i := range values {
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(values[i].Type().Name())
	}
	sb.WriteString(")")
	return sb.String()
}

func (b *MethodBind) reflectCallArgsFromGDExtensionArgs(inst Object, p_args *GDExtensionConstTypePtr) []reflect.Value {
	pArgArraySlice := unsafe.Slice(p_args, len(b.GoArgumentTypes))
	// add an extra parameter for the receiver instance
	args := make([]reflect.Value, len(b.GoArgumentTypes)+1)
	args[0] = reflect.ValueOf(inst)
	for i := 0; i < len(b.GoArgumentTypes); i++ {
		arg := pArgArraySlice[i]
		t := b.GoArgumentTypes[i]
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
			asciiValue := typedValue.ToAscii()
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

func (b *MethodBind) writeReturnValueFromReflectReturnValues(reflectedRet []reflect.Value, r_ret GDExtensionTypePtr) {
	switch len(reflectedRet) {
	case 0:
		if b.GoReturnType != nil {
			log.Panic("no return value expected")
		}
	case 1:
		if b.GoReturnType == nil {
			log.Panic("return value expected but none provided")
		}
		GDExtensionTypePtrFromReflectValue(reflectedRet[0], r_ret)
	case 2:
		if b.GoReturnType == nil {
			log.Panic("return value expected but none provided")
		}
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
			return
		}
		if b.GoReturnType.Name() != reflectedRet[0].Type().Name() {
			log.Panic("unexpected return type",
				zap.String("type", reflectedRet[0].Type().Name()),
			)
		}
		GDExtensionTypePtrFromReflectValue(reflectedRet[0], r_ret)
	default:
		log.Panic("too many values returned", zap.Any("ret", reflectedRet))
	}
}

func (b *MethodBind) Destroy() {
	b.ClassMethodInfo.Destroy()
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

// GoCallback_MethodBindBindCall is called when methods in Go callback into GDScript
//
//export GoCallback_MethodBindBindCall
func GoCallback_MethodBindBindCall(p_method_userdata unsafe.Pointer, p_instance C.GDExtensionClassInstancePtr, p_args *C.GDExtensionVariantPtr, p_argument_count C.GDExtensionInt, r_return C.GDExtensionVariantPtr, r_error *C.GDExtensionCallError) {
	b := (*MethodBind)(p_method_userdata)
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
	// TODO: deal with static method calls as inst will be nil
	inst := ObjectClassFromGDExtensionClassInstancePtr((GDExtensionClassInstancePtr)(p_instance))
	args[0] = reflect.ValueOf(inst)
	argsToString := make([]string, len(b.GoArgumentTypes))
	for i := 0; i < len(b.GoArgumentTypes); i++ {
		var varg *Variant
		if i < (int)(p_argument_count) {
			varg = vArgs[i]
		} else {
			varg = b.DefaultArguments[i]
		}
		str := varg.ToString()
		argsToString[i] = str.ToUtf8()
		args[i+1] = varg.ToReflectValue(
			b.ArgumentTypes[i],
			b.GoArgumentTypes[i],
		)
	}
	log.Info("argument converted",
		zap.String("method_name", (string)(b.GoName)),
		zap.Int("args_cap", cap(args)),
		zap.Int("args_len", len(args)),
		zap.Any("func", spew.Sdump(b.Func)),
		zap.String("args", strings.Join(argsToString[:], ",")),
	)
	reflectedRet := b.Func.Call(args)
	retToString := make([]string, len(reflectedRet))
	for i := 0; i < len(reflectedRet); i++ {
		retToString[i] = fmt.Sprintf("%v", reflectedRet[i].Interface())
	}
	log.Info("returned value",
		zap.String("ret", strings.Join(retToString[:], ",")),
	)
	if len(reflectedRet) == 0 {
		if b.GoReturnType != nil {
			log.Panic("no return value expected")
		}
		GDExtensionVariantPtrWithNil((GDExtensionVariantPtr)(r_return))
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
			GDExtensionVariantPtrFromReflectValue(reflectedRet[0], (GDExtensionVariantPtr)(r_return))
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

				GDExtensionVariantPtrWithNil((GDExtensionVariantPtr)(r_return))
			}
			GDExtensionVariantPtrFromReflectValue(reflectedRet[0], (GDExtensionVariantPtr)(r_return))
		default:
			log.Panic("too many values returned",
				zap.Any("go_name", b.GoName),
				zap.Any("gd_name", b.Name),
			)
		}
	}
	// log.Debug("pre-copy")
	// // This assumes the return value is an empty Variant, so it doesn't need to call the destructor first.
	// // Since only NativeExtensionMethodBind calls this from the Godot side, it should always be the case.
	// CallFunc_GDExtensionInterfaceVariantNewCopy((GDExtensionUninitializedVariantPtr)(r_return), (GDExtensionConstVariantPtr)(ret.ptr()))
	log.Debug("end GoCallback_MethodBindBindCall")
}

// called when godot calls into golang code
//
//export GoCallback_MethodBindBindPtrcall
func GoCallback_MethodBindBindPtrcall(p_method_userdata unsafe.Pointer, p_instance C.GDExtensionClassInstancePtr, p_args *C.GDExtensionConstTypePtr, r_ret C.GDExtensionTypePtr) {
	bind := (*MethodBind)(p_method_userdata)
	log.Debug("GoCallback_MethodBindBindPtrcall called",
		zap.String("bind", bind.String()),
	)
	bind.Ptrcall(
		(GDExtensionClassInstancePtr)(p_instance),
		(*GDExtensionConstTypePtr)(p_args),
		(GDExtensionTypePtr)(r_ret),
	)
}
