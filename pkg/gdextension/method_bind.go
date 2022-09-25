package gdextension

// #include <godot/gdnative_interface.h>
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

type MethodBind struct {
	GoName        MethodName
	GDName        MethodName
	InstanceClass TypeName
	ArgumentCount uint32
	HintFlags     MethodFlags

	IsConst   bool
	HasReturn bool

	GDNameAsStringName *StringName
	Func               reflect.Value

	ArgumentNames         []string
	ArgumentTypes         []GDNativeVariantType
	ArgumentGoTypes       []reflect.Type
	ArgumentPropertyInfos []GDNativePropertyInfo
	ArgumentMetadatas     []GDNativeExtensionClassMethodArgumentMetadata
	DefaultArguments      []*Variant
}

func NewMethodBind(
	method reflect.Method,
	gdName MethodName,
	argNames []string,
	defaultArguments []*Variant,
	hintFlags MethodFlags,
) *MethodBind {

	mt := method.Type

	fn := method.Func

	recv := mt.In(0)

	if recv.Kind() == reflect.Pointer {
		recv = recv.Elem()
	}

	name := (MethodName)(method.Name)
	instanceClass := (TypeName)(recv.Name())
	argumentCount := (uint32)(mt.NumIn() - 1)
	isConst := false
	returnCount := mt.NumOut()
	hasReturn := returnCount > 0
	argumentNames := make([]string, argumentCount)
	argumentTypes := make([]GDNativeVariantType, argumentCount+1)
	argumentGoTypes := make([]reflect.Type, argumentCount+1)

	if returnCount > 2 {
		log.Panic("method cannot return more than 1 type", zap.String("name", (string)(name)))
	}

	if (uint32)(len(argNames)) > argumentCount {
		log.Panic(`Method definition has more arguments than the actual method.`, zap.String("name", (string)(name)))
		return nil
	}

	switch returnCount {
	case 0:
		argumentGoTypes[0] = nil
		argumentTypes[0] = GDNATIVE_VARIANT_TYPE_NIL
	case 1:
		goRt := mt.Out(0)
		gdRt := ReflectTypeToGDNativeVariantType(goRt)

		argumentGoTypes[0] = goRt
		argumentTypes[0] = gdRt
	case 2:
		goRt := mt.Out(0)
		gdRt := ReflectTypeToGDNativeVariantType(goRt)

		argumentGoTypes[0] = goRt
		argumentTypes[0] = gdRt

		errRt := mt.Out(1)
		_, ok := reflect.New(errRt).Interface().(*error)

		if !ok {
			log.Panic("second return type can only be error",
				zap.String("name", (string)(name)),
				zap.Any("type", errRt))
		}
	default:
		log.Panic("method cannot return more than 2 values", zap.String("name", (string)(name)))
	}

	for i := 0; i < (int)(argumentCount); i++ {
		t := mt.In(i + 1)
		if i < len(argNames) {
			argumentNames[i] = argNames[i]
		}
		argumentGoTypes[i+1] = t
		argumentTypes[i+1] = ReflectTypeToGDNativeVariantType(t)
	}

	argumentPropertyInfos := make([]GDNativePropertyInfo, argumentCount+1)
	argumentMetadatas := make([]GDNativeExtensionClassMethodArgumentMetadata, argumentCount+1)

	gdsnName := NewStringNameWithLatin1Chars((string)(gdName))

	return &MethodBind{
		GoName:                name,
		GDName:                gdName,
		InstanceClass:         instanceClass,
		Func:                  fn,
		ArgumentCount:         argumentCount,
		HintFlags:             hintFlags,
		IsConst:               isConst,
		HasReturn:             hasReturn,
		GDNameAsStringName:    &gdsnName,
		ArgumentNames:         argumentNames,
		ArgumentTypes:         argumentTypes,
		ArgumentGoTypes:       argumentGoTypes,
		ArgumentPropertyInfos: argumentPropertyInfos,
		ArgumentMetadatas:     argumentMetadatas,
		DefaultArguments:      defaultArguments,
	}
}

// virtual Variant call(GDExtensionClassInstancePtr p_instance, const GDNativeVariantPtr *p_args, const GDNativeInt p_argument_count, GDNativeCallError &r_error) const {
// 	(static_cast<T *>(p_instance)->*MethodBindVarArgBase<MethodBindVarArgT<T>, T, void, false>::method)((const Variant **)p_args, p_argument_count, r_error);
// 	return {};
// }

/*
template <class T, class... P>
void call_with_variant_args_dv(T *p_instance, void (T::*p_method)(P...), const GDNativeVariantPtr *p_args, int p_argcount, GDNativeCallError &r_error, const std::vector<Variant> &default_values) {
#ifdef DEBUG_ENABLED
	if ((size_t)p_argcount > sizeof...(P)) {
		r_error.error = GDNATIVE_CALL_ERROR_TOO_MANY_ARGUMENTS;
		r_error.argument = (int32_t)sizeof...(P);
		return;
	}
#endif

	int32_t missing = (int32_t)sizeof...(P) - (int32_t)p_argcount;

	int32_t dvs = (int32_t)default_values.size();
#ifdef DEBUG_ENABLED
	if (missing > dvs) {
		r_error.error = GDNATIVE_CALL_ERROR_TOO_FEW_ARGUMENTS;
		r_error.argument = (int32_t)sizeof...(P);
		return;
	}
#endif

	Variant args[sizeof...(P) == 0 ? 1 : sizeof...(P)]; // Avoid zero sized array.
	std::array<const Variant *, sizeof...(P)> argsp;
	for (int32_t i = 0; i < (int32_t)sizeof...(P); i++) {
		if (i < p_argcount) {
			args[i] = Variant(p_args[i]);
		} else {
			args[i] = default_values[i - p_argcount + (dvs - missing)];
		}
		argsp[i] = &args[i];
	}

	call_with_variant_args_helper(p_instance, p_method, argsp.data(), r_error, BuildIndexSequence<sizeof...(P)>{});
}
*/

func (b *MethodBind) GetClassInfo() *ClassInfo {
	tn := b.InstanceClass
	ci, ok := gdRegisteredGDClasses.Get(tn)

	if !ok {
		log.Panic("instance class not found", zap.String("instance_class", (string)(tn)))
	}

	return ci
}

func (b *MethodBind) Call(
	p_instance GDExtensionClassInstancePtr,
	p_args *GDNativeVariantPtr,
	p_argument_count GDNativeInt,
	r_error *GDNativeCallError,
) Variant {
	if (uint32)(p_argument_count) > b.ArgumentCount {
		r_error.SetErrorFields(
			GDNATIVE_CALL_ERROR_TOO_MANY_ARGUMENTS,
			int32(p_argument_count),
			int32(b.ArgumentCount),
		)
		return Variant{}
	}

	w := *(*Wrapped)(unsafe.Pointer(&p_instance))
	vInstance := NewVariantWrapped(w)

	vReturn := NewVariantNil()
	var (
		r_return GDNativeVariantPtr = (GDNativeVariantPtr)(vReturn.ptr())
	)

	GDNativeInterface_variant_call(
		internal.gdnInterface,
		(GDNativeVariantPtr)(&vInstance),
		(GDNativeStringNamePtr)(b.GDNameAsStringName),
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
	p_args *GDNativeVariantPtr,
	p_argument_count GDNativeInt,
	r_error *GDNativeCallError,
) Variant {
	if (uint32)(p_argument_count) > b.ArgumentCount {
		r_error.SetErrorFields(
			GDNATIVE_CALL_ERROR_TOO_MANY_ARGUMENTS,
			int32(p_argument_count),
			int32(b.ArgumentCount),
		)
		return Variant{}
	}

	w := *(*Wrapped)(unsafe.Pointer(&p_instance))
	vInstance := NewVariantWrapped(w)
	vt := vInstance.GetType()

	vReturn := NewVariantNil()
	var (
		r_return GDNativeVariantPtr = (GDNativeVariantPtr)(vReturn.ptr())
	)

	GDNativeInterface_variant_call_static(
		internal.gdnInterface,
		vt,
		(GDNativeStringNamePtr)(b.GDNameAsStringName),
		p_args,
		p_argument_count,
		r_return,
		r_error,
	)

	return vReturn
}

func (b *MethodBind) Ptrcall(
	p_instance GDExtensionClassInstancePtr,
	p_args *GDNativeTypePtr,
	r_ret *GDNativeTypePtr,
) {
	ci := b.GetClassInfo()

	pArgArray := *(*[MAX_ARG_COUNT]unsafe.Pointer)(unsafe.Pointer(p_args))

	// add an extra parameter for the receiver instance
	args := make([]reflect.Value, b.ArgumentCount+1)

	inst := GDClassFromGDExtensionClassInstancePtr(ci, p_instance)

	args[0] = reflect.ValueOf(inst)

	for i := 0; i < (int)(b.ArgumentCount); i++ {
		arg := pArgArray[i]

		t := b.ArgumentGoTypes[i+1]

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
		if b.HasReturn {
			log.Panic("no return value expected")
		}
	case 1:
		if !b.HasReturn {
			log.Panic("return value expected but none provided")
		}

		valueInterface := reflectedRet[0].Interface()

		*r_ret = *(*GDNativeTypePtr)(unsafe.Pointer(&valueInterface))
	case 2:
		if !b.HasReturn {
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
			*r_ret = *(*GDNativeTypePtr)(unsafe.Pointer(&valueInterface))
		}
	default:
		log.Panic("too many values returned", zap.Any("ret", reflectedRet))
	}
}

func (b *MethodBind) Destroy() {
	b.GDNameAsStringName.Destroy()
}

//export GoCallback_MethodBindBindGetArgumentType
func GoCallback_MethodBindBindGetArgumentType(p_method_userdata unsafe.Pointer, p_argument int32) C.GDNativeVariantType {
	bind := (*MethodBind)(p_method_userdata)
	return (C.GDNativeVariantType)(bind.ArgumentTypes[p_argument+1])
}

//export GoCallback_MethodBindBindGetArgumentInfo
func GoCallback_MethodBindBindGetArgumentInfo(p_method_userdata unsafe.Pointer, p_argument int32, r_info unsafe.Pointer) {
	bind := (*MethodBind)(p_method_userdata)
	*((*GDNativePropertyInfo)(r_info)) = bind.ArgumentPropertyInfos[p_argument+1]
}

//export GoCallback_MethodBindBindGetArgumentMetadata
func GoCallback_MethodBindBindGetArgumentMetadata(p_method_userdata unsafe.Pointer, p_argument int32) C.GDNativeExtensionClassMethodArgumentMetadata {
	bind := (*MethodBind)(p_method_userdata)
	return (C.GDNativeExtensionClassMethodArgumentMetadata)(bind.ArgumentMetadatas[p_argument+1])
}

// this is called when methods in Go callback into GDScript
//
//export GoCallback_MethodBindBindCall
func GoCallback_MethodBindBindCall(p_method_userdata unsafe.Pointer, p_instance C.GDExtensionClassInstancePtr, p_args *C.GDNativeVariantPtr, p_argument_count C.GDNativeInt, r_return C.GDNativeVariantPtr, r_error *C.GDNativeCallError) {
	b := (*MethodBind)(p_method_userdata)

	ci := b.GetClassInfo()

	var vArgs *[MAX_ARG_COUNT]*Variant

	if p_args != nil {
		vArgs = (*[MAX_ARG_COUNT]*Variant)(unsafe.Pointer(p_args))
	}

	if (uint32)(p_argument_count) > b.ArgumentCount {
		((*GDNativeCallError)(unsafe.Pointer(r_error))).SetErrorFields(
			GDNATIVE_CALL_ERROR_TOO_MANY_ARGUMENTS,
			int32(p_argument_count),
			int32(b.ArgumentCount),
		)
		return
	}

	// add an extra parameter for the receiver instance
	args := make([]reflect.Value, b.ArgumentCount+1)

	inst := GDClassFromGDExtensionClassInstancePtr(ci, (GDExtensionClassInstancePtr)(p_instance))

	args[0] = reflect.ValueOf(inst)

	for i := 0; i < (int)(b.ArgumentCount); i++ {
		var varg *Variant
		if i < (int)(p_argument_count) {
			varg = vArgs[i]
		} else {
			varg = b.DefaultArguments[i]
		}

		args[i+1] = varg.ToReflectValue(
			b.ArgumentTypes[i+1],
			b.ArgumentGoTypes[i+1],
		)
	}

	log.Debug("argument converted",
		zap.String("method_name", (string)(b.GoName)),
		zapGDNativeVariantPtrp("vargs", (*GDNativeVariantPtr)(p_args), (int)(p_argument_count)),
		zap.String("args", spew.Sdump(args)),
	)

	reflectedRet := b.Func.Call(args)

	var ret Variant

	if len(reflectedRet) == 0 {
		if b.HasReturn {
			log.Panic("no return value expected")
		}

		ret = NewVariantNil()
		defer ret.Destroy()
	} else {
		if !b.HasReturn {
			log.Panic("return value expected but none provided")
		}

		log.Debug("returned", zap.String("value", spew.Sdump(reflectedRet[0])))


		switch len(reflectedRet) {
		case 0:
			if b.HasReturn {
				log.Panic("no return value expected")
			}
		case 1:
			if !b.HasReturn {
				log.Panic("return value expected but none provided")
			}

			ret = NewVariantFromReflectValue(reflectedRet[0])
			defer ret.Destroy()
		case 2:
			if !b.HasReturn {
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
				ret = NewVariantNil()
				defer ret.Destroy()
			}

			ret = NewVariantFromReflectValue(reflectedRet[0])
			defer ret.Destroy()
		default:
			log.Panic("too many values returned", zap.Any("ret", reflectedRet))
		}
	}

	// This assumes the return value is an empty Variant, so it doesn't need to call the destructor first.
	// Since only NativeExtensionMethodBind calls this from the Godot side, it should always be the case.
	GDNativeInterface_variant_new_copy(internal.gdnInterface, (GDNativeVariantPtr)(r_return), (GDNativeVariantPtr)(ret.ptr()))
}

//export GoCallback_MethodBindBindPtrcall
func GoCallback_MethodBindBindPtrcall(p_method_userdata unsafe.Pointer, p_instance C.GDExtensionClassInstancePtr, p_args *C.GDNativeTypePtr, r_ret C.GDNativeTypePtr) {
	bind := (*MethodBind)(p_method_userdata)
	bind.Ptrcall(
		(GDExtensionClassInstancePtr)(p_instance),
		(*GDNativeTypePtr)(p_args),
		(*GDNativeTypePtr)(&r_ret),
	)
}
