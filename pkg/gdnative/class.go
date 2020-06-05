package gdnative

/*
#include <cgo_gateway_register_class.h>
#include <gdnative_api_struct.gen.h>
#include <nativescript.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"github.com/pcting/godot-go/pkg/log"
	"reflect"
	"strings"
	"unsafe"

	"github.com/pinzolo/casee"
)

type NativeScriptClass interface {
	Class
	OnClassRegistered(d ClassRegisteredEvent)
}

type Class interface {
	GetOwnerObject() *GodotObject
	GetTypeTag() TypeTag
	ClassName() string
	BaseClass() string
	Free()
}

type ClassRegisteredEvent struct {
	ClassName    string
	ClassType    reflect.Type
	ClassTypeTag TypeTag
	BaseTypeTag  TypeTag
}

type RegisterSignalArg struct {
	Name string
	Type VariantType
}

type CreateClassFunc func(owner *GodotObject, typeTag TypeTag) Class

// MethodData is used as the key for the internal identity map
type MethodData uint

type GoInstanceCreateFunc struct {
	InstanceCreateFunc C.godot_instance_create_func
	CreateFunc         CreateFunc
	FreeFunc           FreeFunc
}

type GoInstanceDestroyFunc struct {
	InstanceDestroyFunc C.godot_instance_destroy_func
	DestroyFunc         DestroyFunc
	FreeFunc            FreeFunc
}

type GoInstanceMethod struct {
	InstanceMethod C.godot_instance_method
	MethodFunc     MethodFunc
	FreeFunc       FreeFunc
}

// UserData must be unique to the instance
type UserData Class
type CreateFunc func(*GodotObject, MethodData) UserData
type DestroyFunc func(*GodotObject, MethodData, UserData)
type MethodFunc func(*GodotObject, MethodData, UserData, []*Variant) Variant
type FreeFunc func(MethodData)

var (
	SingletonConstructorMap       = map[UserData]CreateClassFunc{}
	ConstructorMap                = map[string]CreateClassFunc{}
	registerClassCreateCallbacks  = map[TypeTag]GoInstanceCreateFunc{}
	registerClassDestroyCallbacks = map[TypeTag]GoInstanceDestroyFunc{}
	methodCallbacks               = map[MethodTag]GoInstanceMethod{}

	classInstances = map[UserData]Class{}
	typeOfSignal   = reflect.TypeOf(Signal{})
)

func NewRegisteredClassCallbacks(methodData MethodData, base string, createClassFunc CreateClassFunc) (c GoInstanceCreateFunc, d GoInstanceDestroyFunc) {
	c.InstanceCreateFunc = C.godot_instance_create_func{}
	c.InstanceCreateFunc.create_func = (C.create_func)(unsafe.Pointer(C.cgo_gateway_create_func))
	c.InstanceCreateFunc.method_data = unsafe.Pointer(uintptr(uint(methodData)))
	c.InstanceCreateFunc.free_func = (C.free_func)(unsafe.Pointer(C.cgo_gateway_create_free_func))
	c.CreateFunc = func(o *GodotObject, md MethodData) UserData {
		tt := TypeTag(md)

		classInst := createClassFunc(o, tt)
		if classInst == nil {
			log.Panic("class should not be nil")
		}

		ud := classInst

		if _, ok := classInstances[ud]; ok {
			log.WithFields(WithRegisteredClassCB(tt, base, ud)).Panic("class instance already exists")
		} else {
			log.WithFields(WithRegisteredClassCB(tt, base, ud)).Trace("call create class")
		}

		classInstances[ud] = classInst

		return classInst
	}
	c.FreeFunc = func(md MethodData) {
		tt := TypeTag(md)

		log.WithFields(WithRegisteredClassFreeCB(tt, base)).Trace("call free create class")

		if _, ok := registerClassCreateCallbacks[tt]; !ok {
			log.WithFields(WithTypeTag(tt)).Panic("not found in registerClassCreateCallbacks")
		}

		delete(registerClassCreateCallbacks, tt)
	}

	d.InstanceDestroyFunc = C.godot_instance_destroy_func{}
	d.InstanceDestroyFunc.destroy_func = (C.destroy_func)(unsafe.Pointer(C.cgo_gateway_destroy_func))
	d.InstanceDestroyFunc.method_data = unsafe.Pointer(uintptr(uint(methodData)))
	d.InstanceDestroyFunc.free_func = (C.free_func)(unsafe.Pointer(C.cgo_gateway_destroy_free_func))
	d.DestroyFunc = func(o *GodotObject, md MethodData, ud UserData) {
		tt := TypeTag(md)

		if _, ok := classInstances[ud]; ok {
			log.WithFields(WithRegisteredClassCB(tt, base, ud)).Trace("call destroy class")
		} else {
			log.WithFields(WithRegisteredClassCB(tt, base, ud)).Panic("cannot find class to destroy")
		}

		delete(classInstances, ud)
	}
	d.FreeFunc = func(md MethodData) {
		tt := TypeTag(md)

		log.WithFields(WithRegisteredClassFreeCB(tt, base)).Trace("call free destroy class")

		if _, ok := registerClassDestroyCallbacks[tt]; !ok {
			log.WithFields(WithTypeTag(tt)).Panic("not found in registerClassDestroyCallbacks")
		}

		delete(registerClassDestroyCallbacks, tt)
	}

	return
}

func RegisterClass(instance NativeScriptClass, createClassFunc CreateClassFunc) {
	var (
		name string
	)

	base := instance.BaseClass()

	classType := reflect.TypeOf(instance)

	if classType.Kind() == reflect.Ptr {
		name = classType.Elem().Name()
	}

	if len(name) == 0 {
		log.WithField("class", name).Panic("invalid class name")
	}

	ctt, btt := RegisterState.TagDB.RegisterType(name, base)

	classMethodData := MethodData(ctt)

	createFunc, destroyFunc := NewRegisteredClassCallbacks(classMethodData, base, createClassFunc)

	if _, ok := registerClassCreateCallbacks[ctt]; ok {
		log.WithFields(WithRegisteredClass(name, base)).Panic("create callback with the same name already registered")
	}

	if _, ok := registerClassDestroyCallbacks[ctt]; ok {
		log.WithFields(WithRegisteredClass(name, base)).Panic("destroy callback with the same name already registered")
	}

	registerClassCreateCallbacks[ctt] = createFunc
	registerClassDestroyCallbacks[ctt] = destroyFunc

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cbase := C.CString(base)
	defer C.free(unsafe.Pointer(cbase))

	C.go_godot_nativescript_register_class(
		NativescriptApi,
		RegisterState.NativescriptHandle,
		cname,
		cbase,
		createFunc.InstanceCreateFunc,
		destroyFunc.InstanceDestroyFunc,
	)

	C.go_godot_nativescript_set_type_tag(Nativescript11Api, RegisterState.NativescriptHandle, cname, unsafe.Pointer(uintptr(ctt)))

	instance.OnClassRegistered(ClassRegisteredEvent{
		ClassName:    name,
		ClassType:    classType,
		ClassTypeTag: ctt,
		BaseTypeTag:  btt,
	})

	log.WithFields(WithRegisteredClass(name, base)).Info("class registered")
}

func (d ClassRegisteredEvent) RegisterSignal(signalName string, varargs ...RegisterSignalArg) {
	cClassName := C.CString(d.ClassName)
	defer C.free(unsafe.Pointer(cClassName))

	gsSignalName := NewStringFromGoString(signalName)
	defer gsSignalName.Destroy()

	size := len(varargs)

	signal := C.godot_signal{}
	signal.name = *(*C.godot_string)(unsafe.Pointer(&gsSignalName))
	signal.num_args = (C.int)(size)
	signal.num_default_args = 0

	if size > 0 {
		signal.args = (*C.godot_signal_argument)(unsafe.Pointer(AllocZeros(int32(unsafe.Sizeof(SignalArgument{})) * int32(size))))
		defer Free(unsafe.Pointer(signal.args))
	}

	argPtr := (*C.godot_signal_argument)(unsafe.Pointer(signal.args))

	for i, a := range varargs {
		str := NewStringFromGoString(a.Name)
		defer str.Destroy()

		curArgPtr := (*C.godot_signal_argument)(unsafe.Pointer(uintptr(unsafe.Pointer(argPtr)) + uintptr(i)*uintptr(C.sizeof_godot_signal_argument)))
		curArgPtr.name = *(*C.godot_string)(unsafe.Pointer(&str))
		curArgPtr._type = (C.godot_int)(a.Type)
	}

	C.go_godot_nativescript_register_signal(NativescriptApi, RegisterState.NativescriptHandle, cClassName, &signal)
}

func (d ClassRegisteredEvent) RegisterMethod(methodName string) {
	if !strings.HasPrefix(methodName, "X_") {
		log.WithField("method", methodName).Panic("method must be prefixed with 'X_'")
	}

	method, ok := d.ClassType.MethodByName(methodName)

	if !ok {
		log.WithField("method", methodName).Panic("method not found")
	}

	methodBindName := "_" + casee.ToSnakeCase(method.Name[2:])

	attribs := C.godot_method_attributes{}
	attribs.rpc_type = C.GODOT_METHOD_RPC_MODE_DISABLED

	mMethodTag := RegisterState.TagDB.RegisterMethod(d.ClassName, methodBindName)

	inst := C.godot_instance_method{}
	inst.method = (C.create_func)(unsafe.Pointer(C.cgo_gateway_method_func))
	inst.method_data = unsafe.Pointer(uintptr(uint(mMethodTag)))
	inst.free_func = (C.free_func)(unsafe.Pointer(C.cgo_gateway_method_free_func))

	regMethod := GoInstanceMethod{
		InstanceMethod: inst,
		MethodFunc: func(_ *GodotObject, md MethodData, ud UserData, args []*Variant) Variant {
			callArgs := make([]reflect.Value, len(args))

			for i, v := range args {
				// Convert the variant into its base type
				callArgs[i] = VariantToGoType(*v)
			}

			inst, ok := classInstances[ud]

			if !ok {
				log.WithField("identity map", fmt.Sprintf("%+v", classInstances)).WithFields(WithUserData(ud)).Panic("instance not found")
			}

			instValue := reflect.ValueOf(inst)

			instMethod := instValue.MethodByName(method.Name)

			result := instMethod.Call(callArgs)

			resultSize := len(result)

			if resultSize == 0 {
				return NewVariantNil()
			}

			if resultSize > 1 {
				log.Panic(fmt.Sprintf("only one value is expected: %v", result))
			}

			return GoTypeToVariant(result[0])
		},
		FreeFunc: func(md MethodData) {
			mt := MethodTag(md)
			log.WithFields(WithRegisteredMethodFreeCB(md, method.Name)).Trace("call free class method")

			if _, ok := methodCallbacks[mt]; !ok {
				log.WithFields(WithMethodTag(mt)).Panic("not found in methodCallbacks")
			}

			delete(methodCallbacks, mt)
		},
	}

	if _, ok := methodCallbacks[mMethodTag]; ok {
		log.WithFields(WithMethodTag(mMethodTag)).Panic("method callback already registered")
	}

	methodCallbacks[mMethodTag] = regMethod

	cclassName := C.CString(d.ClassName)
	defer C.free(unsafe.Pointer(cclassName))

	cmethodBindName := C.CString(methodBindName)
	defer C.free(unsafe.Pointer(cmethodBindName))

	C.go_godot_nativescript_register_method(
		NativescriptApi,
		RegisterState.NativescriptHandle,
		cclassName,
		cmethodBindName,
		attribs,
		regMethod.InstanceMethod,
	)

	log.WithFields(WithMethodTag(mMethodTag)).Trace("class method registered")
}

//export go_create_func
func go_create_func(godotObject *C.godot_object, methodData unsafe.Pointer) unsafe.Pointer {
	obj := (*GodotObject)(godotObject)
	tt := TypeTag(uint(uintptr(methodData)))

	rc, ok := registerClassCreateCallbacks[tt]

	if !ok {
		log.WithFields(WithTypeTag(tt)).Panic("create func callback not found")
	}

	ret := rc.CreateFunc(obj, MethodData(tt))

	pUserData := AllocCopy(unsafe.Pointer(&ret), int32(unsafe.Sizeof(ret)))

	return pUserData
}

//export go_destroy_func
func go_destroy_func(godotObject *C.godot_object, methodData unsafe.Pointer, userData unsafe.Pointer) {
	obj := (*GodotObject)(godotObject)
	tt := TypeTag(uint(uintptr(methodData)))

	rc, ok := registerClassDestroyCallbacks[tt]
	ud := *(*UserData)(userData)

	if !ok {
		log.WithFields(WithTypeTag(tt)).Panic("destroy func callback not found")
	}

	rc.DestroyFunc(obj, MethodData(tt), ud)

	Free(userData)
}

//export go_method_func
func go_method_func(godotObject *C.godot_object, methodData unsafe.Pointer, userData unsafe.Pointer, nargs C.int, args **C.godot_variant) C.godot_variant {
	obj := (*GodotObject)(godotObject)
	mt := MethodTag(uint(uintptr(methodData)))
	ud := *(*UserData)(userData)
	na := int(nargs)

	argArr := NewSliceFromCPtrPtrRef(na, unsafe.Pointer(args))

	if fmt.Sprintf("%p", args) != fmt.Sprintf("%p", argArr) {
		log.WithField(
			"arg", fmt.Sprintf("%p", args),
		).WithField(
			"argArr", fmt.Sprintf("%p", argArr),
		).Panic("wrong address for args slice")
	}

	as := make([]*Variant, na)

	for i := 0; i < na; i++ {
		as[i] = (*Variant)(argArr[i])
	}

	rc, ok := methodCallbacks[mt]

	if !ok {
		log.WithFields(WithMethodTag(mt)).Panic("method callback not found")
	}

	ret := rc.MethodFunc(obj, MethodData(mt), ud, as)

	return *(*C.godot_variant)(unsafe.Pointer(&ret))
}

//export go_create_free_func
func go_create_free_func(methodData unsafe.Pointer) {
	tt := TypeTag(uint(uintptr(methodData)))
	rc, ok := registerClassCreateCallbacks[tt]

	if !ok {
		log.WithFields(WithTypeTag(tt)).Panic("create free callback not found")
	}

	rc.FreeFunc(MethodData(tt))
}

//export go_destroy_free_func
func go_destroy_free_func(methodData unsafe.Pointer) {
	tt := TypeTag(uint(uintptr(methodData)))
	rc, ok := registerClassDestroyCallbacks[tt]

	if !ok {
		log.WithFields(WithTypeTag(tt)).Panic("destroy free callback not found")
	}

	rc.FreeFunc(MethodData(tt))
}

//export go_method_free_func
func go_method_free_func(methodData unsafe.Pointer) {
	mt := MethodTag(uint(uintptr(methodData)))
	rc, ok := methodCallbacks[mt]

	if !ok {
		log.WithFields(WithMethodTag(mt)).Panic("method free callback not found")
	}

	rc.FreeFunc(MethodData(mt))
}
