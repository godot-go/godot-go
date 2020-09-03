package gdnative

/*
#include <nativescript.wrapper.gen.h>
#include <cgo_gateway_register_class.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"reflect"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
)

// CreateNativeScriptClassFunc are functions defined by the user to return
// an instance of their Class
type CreateNativeScriptClassFunc func(owner *GodotObject, typeTag TypeTag) NativeScriptClass

var (
	// TODO: this currently doesn't empty out since there is no "unregister" method
	registeredCreateClassInstanceFuncs = map[TypeTag]CreateNativeScriptClassFunc{}

	// TODO: do we want to nest this map to reduce the chance of hash collisions?
	nativeScriptInstanceMap UserDataMap = UserDataMap{}
)

func RegisterClass(instance NativeScriptClass, classFunc CreateNativeScriptClassFunc) {
	baseName := instance.BaseClass()

	classType := reflect.TypeOf(instance)

	if classType.Kind() != reflect.Ptr {
		log.Panic("instance should be a pointer to a struct")
	}

	className := classType.Elem().Name()

	if len(className) == 0 {
		log.WithField("class", className).Panic("invalid class name")
	}

	ctt, btt := RegisterState.TagDB.RegisterType(className, baseName)

	classMethodData := MethodData(ctt)

	createFunc := C.godot_instance_create_func{}
	createFunc.create_func = (C.create_func)(unsafe.Pointer(C.cgo_gateway_create_func))
	createFunc.method_data = unsafe.Pointer(uintptr(classMethodData))
	createFunc.free_func = (C.free_func)(unsafe.Pointer(C.cgo_gateway_create_free_func))

	destroyFunc := C.godot_instance_destroy_func{}
	destroyFunc.destroy_func = (C.destroy_func)(unsafe.Pointer(C.cgo_gateway_destroy_func))
	destroyFunc.method_data = unsafe.Pointer(uintptr(classMethodData))
	destroyFunc.free_func = (C.free_func)(unsafe.Pointer(C.cgo_gateway_destroy_free_func))

	if _, ok := registeredCreateClassInstanceFuncs[ctt]; ok {
		log.WithFields(WithRegisteredClass(className, baseName)).Panic("create class function with the same name already registered")
	}

	registeredCreateClassInstanceFuncs[ctt] = classFunc

	event := NewClassRegisteredEvent(
		className,
		classType,
		ctt,
		baseName,
		btt,
	)
	defer event.Destroy()

	C.go_godot_nativescript_register_class(
		NativescriptApi,
		RegisterState.NativescriptHandle,
		event._pCharClassName,
		event._pCharBaseName,
		createFunc,
		destroyFunc,
	)

	C.go_godot_nativescript_set_type_tag(
		Nativescript11Api,
		RegisterState.NativescriptHandle,
		event._pCharClassName,
		unsafe.Pointer(uintptr(ctt)),
	)

	// notify the class that class registration has completed so that
	// it can begin registering methods, signals, and properties
	instance.OnClassRegistered(event)

	log.WithFields(WithRegisteredClass(className, baseName)).Info("class registered")
}

//export go_create_func
func go_create_func(godotObject *C.godot_object, methodData unsafe.Pointer) unsafe.Pointer {
	obj := (*GodotObject)(godotObject)
	tt := TypeTag(uintptr(methodData))

	createClassFunc, ok := registeredCreateClassInstanceFuncs[tt]

	if !ok {
		log.WithFields(WithTypeTag(tt)).Panic("create func callback not found")
	}

	classInst := createClassFunc(obj, tt)

	classInst.generateUserData(tt)

	if classInst.GetUserData() == UserData(0) {
		log.Panic("class must have a user data identifier")
	}

	if classInst == nil {
		log.Panic("class must not be nil")
	}

	classInst.Init()

	ud := classInst.GetUserData()

	if _, ok := nativeScriptInstanceMap[ud]; ok {
		log.Panic("user data error: collision found")
	}

	nativeScriptInstanceMap[ud] = classInst

	return (unsafe.Pointer)(uintptr(ud))
}

//export go_destroy_func
func go_destroy_func(godotObject *C.godot_object, methodData unsafe.Pointer, userData unsafe.Pointer) {
	delete(nativeScriptInstanceMap, UserData(uintptr(userData)))
}

//export go_create_free_func
func go_create_free_func(methodData unsafe.Pointer) {

}

//export go_destroy_free_func
func go_destroy_free_func(methodData unsafe.Pointer) {

}
