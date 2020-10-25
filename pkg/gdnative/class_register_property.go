package gdnative

/*
#include <nativescript.wrapper.gen.h>
#include <cgo_gateway_register_class.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
)

type GoPropertySetFunc func(Variant)
type GoPropertyGetFunc func() Variant

func (d ClassRegisteredEvent) RegisterProperty(
	name, setFunc, getFunc string, defaultValue Variant,
) {
	var (
		ok bool
	)

	if _, ok = d.ClassType.MethodByName(setFunc); !ok {
		log.Panic("setFunc not found", StringField("setFunc", setFunc))
	}

	if _, ok = d.ClassType.MethodByName(getFunc); !ok {
		log.Panic("getFunc not found", StringField("getFunc", getFunc))
	}

	pst := RegisterState.TagDB.RegisterPropertySet(d.ClassName, name, setFunc)
	pgt := RegisterState.TagDB.RegisterPropertyGet(d.ClassName, name, getFunc)

	rpcMode := C.GODOT_METHOD_RPC_MODE_DISABLED
	usage := C.GODOT_PROPERTY_USAGE_DEFAULT
	hint := C.GODOT_PROPERTY_HINT_NONE
	hintString := NewStringFromGoString("")
	defer hintString.Destroy()

	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	cClassName := C.CString(d.ClassName)
	defer C.free(unsafe.Pointer(cClassName))

	attr := C.godot_property_attributes{}

	valueType := defaultValue.GetType()

	if valueType == GODOT_VARIANT_TYPE_NIL {
		attr._type = C.GODOT_VARIANT_TYPE_OBJECT
	} else {
		attr._type = (C.int)(valueType)
		attr.default_value = *(*C.godot_variant)(unsafe.Pointer(&defaultValue))
	}
	attr.hint = (C.godot_property_hint)(hint)
	attr.rset_type = (C.godot_method_rpc_mode)(rpcMode)
	attr.usage = (C.godot_property_usage_flags)(usage)
	attr.hint_string = *(*C.godot_string)(unsafe.Pointer(&hintString))

	propSetFunc := C.godot_property_set_func{}
	propSetFunc.method_data = unsafe.Pointer(uintptr(uint(pst)))
	propSetFunc.set_func = (C.set_func)(unsafe.Pointer(C.cgo_gateway_property_set_func))

	propGetFunc := C.godot_property_get_func{}
	propGetFunc.method_data = unsafe.Pointer(uintptr(uint(pgt)))
	propGetFunc.get_func = (C.get_func)(unsafe.Pointer(C.cgo_gateway_property_get_func))

	C.go_godot_nativescript_register_property(NativescriptApi, RegisterState.NativescriptHandle, cClassName, cName, &attr, propSetFunc, propGetFunc)

	log.Debug("class property registered",
		StringField("class", d.ClassName),
		StringField("name", name),
		StringField("property", name),
	)
}

//export go_property_set_func
func go_property_set_func(owner *C.godot_object, methodData unsafe.Pointer, userData unsafe.Pointer, value *C.godot_variant) {
	tag := PropertySetTag(uint(uintptr(methodData)))
	ud := (UserData)(uintptr(userData))

	callArgs := []reflect.Value{
		reflect.ValueOf(*(*Variant)(unsafe.Pointer(value))),
	}

	classInst, ok := nativeScriptInstanceMap[ud]

	if !ok {
		log.Panic("unable to find instance")
	}

	// assert this is a valid instance
	log.Info("assert class and base names",
		StringField("className", classInst.ClassName()),
		StringField("baseName", classInst.BaseClass()),
	)

	valueInst := reflect.ValueOf(classInst)

	methodName := RegisterState.TagDB.GetRegisteredPropertySet(tag)

	instMethod := valueInst.MethodByName(methodName)

	if instMethod == (reflect.Value{}) {
		log.Panic("unable to find property set method", StringField("methodName", methodName))
	}

	instMethod.Call(callArgs)
}

//export go_property_get_func
func go_property_get_func(owner *C.godot_object, methodData unsafe.Pointer, userData unsafe.Pointer) C.godot_variant {
	tag := PropertyGetTag(uint(uintptr(methodData)))
	ud := (UserData)(uintptr(userData))

	callArgs := []reflect.Value{}

	classInst, ok := nativeScriptInstanceMap[ud]

	if !ok {
		log.Panic("unable to find instance")
	}

	// assert this is a valid instance
	log.Info("assert class and base names",
		StringField("className", classInst.ClassName()),
		StringField("baseName", classInst.BaseClass()),
	)

	valueInst := reflect.ValueOf(classInst)

	methodName := RegisterState.TagDB.GetRegisteredPropertyGet(tag)

	instMethod := valueInst.MethodByName(methodName)

	if instMethod == (reflect.Value{}) {
		log.Panic("unable to find property get method", StringField("methodName", methodName))
	}

	result := instMethod.Call(callArgs)

	resultSize := len(result)

	if resultSize == 0 {
		ret := NewVariantNil()
		return *(*C.godot_variant)(unsafe.Pointer(&ret))
	}

	if resultSize > 1 {
		log.Panic(fmt.Sprintf("only one value is expected: %v", result))
	}

	ret := GoTypeToVariant(result[0])

	return *(*C.godot_variant)(unsafe.Pointer(&ret))
}
