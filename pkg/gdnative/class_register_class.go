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
	registeredNativeScriptClassTypes = map[TypeTag]reflect.Type{}

	// TODO: do we want to nest this map to reduce the chance of hash collisions?
	nativeScriptInstanceMap UserDataMap = UserDataMap{}

	nativeScriptClassType = reflect.TypeOf(new(NativeScriptClass)).Elem()
)

func RegisterClass(instance NativeScriptClass) {
	// godot-cpp implementation:
	//
	// template <class T>
	// void register_class() {
	// 	godot_instance_create_func create = {};
	// 	create.create_func = _godot_class_instance_func<T>;

	// 	godot_instance_destroy_func destroy = {};
	// 	destroy.destroy_func = _godot_class_destroy_func<T>;

	// 	_TagDB::register_type(T::___get_id(), T::___get_base_id());

	// 	godot::nativescript_api->godot_nativescript_register_class(godot::_RegisterState::nativescript_handle, T::___get_type_name(), T::___get_base_type_name(), create, destroy);
	// 	godot::nativescript_1_1_api->godot_nativescript_set_type_tag(godot::_RegisterState::nativescript_handle, T::___get_type_name(), (const void *)typeid(T).hash_code());
	// 	T::_register_methods();
	// }

	baseName := instance.BaseClass()

	// classType should hold the type: *MyCustomClassStruct
	classType := reflect.TypeOf(instance)

	// all classTypes should implement gdnative.NativeScriptClass
	if !classType.Implements(nativeScriptClassType) {
		log.Panic("class type must implement NativeScriptClass")
	}

	className := classType.Elem().Name()

	if len(className) == 0 {
		log.Panic("invalid class name", StringField("class", className))
	}

	ctt, btt := RegisterState.TagDB.RegisterType(className, baseName)

	classMethodData := MethodData(ctt)

	createFunc := C.godot_instance_create_func{}
	createFunc.create_func = (C.create_func)(unsafe.Pointer(C.cgo_gateway_create_func))
	createFunc.method_data = unsafe.Pointer(uintptr(classMethodData))

	destroyFunc := C.godot_instance_destroy_func{}
	destroyFunc.destroy_func = (C.destroy_func)(unsafe.Pointer(C.cgo_gateway_destroy_func))
	destroyFunc.method_data = unsafe.Pointer(uintptr(classMethodData))

	if _, ok := registeredNativeScriptClassTypes[ctt]; ok {
		log.Panic("create class function with the same name already registered",
			StringField("class", className),
			StringField("base", baseName),
		)
	}

	registeredNativeScriptClassTypes[ctt] = classType

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

	log.Info("class registered", StringField("class", className))
}

//export go_create_func
func go_create_func(godotObject *C.godot_object, methodData unsafe.Pointer) unsafe.Pointer {
	// godot-cpp implementation:
	//
	// template <class T>
	// void *_godot_class_instance_func(godot_object *p, void *method_data) {
	// 	T *d = new T();
	// 	d->_owner = p;
	// 	d->_type_tag = typeid(T).hash_code();
	// 	d->_init();
	// 	return d;
	// }

	obj := (*GodotObject)(godotObject)
	tt := TypeTag(uintptr(methodData))

	classType, ok := registeredNativeScriptClassTypes[tt]

	if !ok {
		log.Info("create func callback not found",
			StringField("class", RegisterState.TagDB.GetRegisteredClassName(tt)),
		)
	}

	// returns an instance of *MyCustomClassStruct{}
	reflectedClassInst := reflect.New(classType.Elem())

	classInst := reflectedClassInst.Interface().(NativeScriptClass)

	if classInst == nil {
		log.Panic("cast failure", AnyField("inst", classInst))
	}

	if obj == nil {
		log.Panic("owner object cannot be nil", AnyField("owner", obj))
	}

	classInst.setOwnerObject(obj)
	classInst.setTypeTag(tt)
	classInst.setUserDataFromTypeTag(tt)

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

	log.Debug("go_create_func: created instance",
		StringField("class", RegisterState.TagDB.GetRegisteredClassName(tt)),
		GodotObjectField("owner", obj),
		NativeScriptClassField("inst", classInst),
	)

	nativeScriptInstanceMap[ud] = classInst

	return (unsafe.Pointer)(uintptr(ud))
}

//export go_destroy_func
func go_destroy_func(godotObject *C.godot_object, methodData unsafe.Pointer, userData unsafe.Pointer) {
	delete(nativeScriptInstanceMap, UserData(uintptr(userData)))
}
