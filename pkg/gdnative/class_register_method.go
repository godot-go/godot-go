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

// MethodData is used as the key for the internal identity map
type MethodData uint

type MethodFunc func(*GodotObject, MethodData, UserData, []*Variant) Variant

func (d ClassRegisteredEvent) RegisterMethod(bindName string, methodName string) {
	_, ok := d.ClassType.MethodByName(methodName)

	if !ok {
		log.Panic("method not found",
			StringField("bind", bindName),
			StringField("method", methodName),
		)
	}

	attribs := C.godot_method_attributes{}
	attribs.rpc_type = C.GODOT_METHOD_RPC_MODE_DISABLED

	tag := RegisterState.TagDB.RegisterMethod(d.ClassName, bindName, methodName)

	inst := C.godot_instance_method{}
	inst.method = (C.create_func)(unsafe.Pointer(C.cgo_gateway_method_func))
	inst.method_data = unsafe.Pointer(uintptr(tag))

	cClassName := C.CString(d.ClassName)
	defer C.free(unsafe.Pointer(cClassName))

	cBindName := C.CString(bindName)
	defer C.free(unsafe.Pointer(cBindName))

	C.go_godot_nativescript_register_method(
		NativescriptApi,
		RegisterState.NativescriptHandle,
		cClassName,
		cBindName,
		attribs,
		inst,
	)

	log.Debug("class method registered",
		StringField("class", d.ClassName),
		StringField("bind", bindName),
		StringField("method", methodName),
		MethodTagField("methodTag", tag),
	)
}

//export go_method_func
func go_method_func(godotObject *C.godot_object, methodData unsafe.Pointer, userData unsafe.Pointer, nargs C.int, args **C.godot_variant) C.godot_variant {
	tag := MethodTag(uintptr(methodData))
	ud := UserData(uintptr(userData))
	na := int(nargs)

	argArr := WrapUnsafePointerAsSlice(int(na), unsafe.Pointer(args))

	if fmt.Sprintf("%p", args) != fmt.Sprintf("%p", argArr) {
		log.Panic("wrong address for args slice",
			StringField("arg", fmt.Sprintf("%p", args)),
			StringField("argArr", fmt.Sprintf("%p", argArr)),
		)
	}

	as := make([]*Variant, na)

	for i := 0; i < na; i++ {
		as[i] = (*Variant)(argArr[i])
	}

	callArgs := make([]reflect.Value, na)

	// Unwrap the Variants
	for i, v := range as {
		x := VariantToGoType(*v)
		// if x.IsValid() && x.Kind() == reflect.Int64 {
		// 	int32v := int32(x.Int())
		// 	callArgs[i] = reflect.ValueOf(int32v)
		// } else {
		// 	callArgs[i] = x
		// }
		callArgs[i] = x
	}

	classInst, ok := nativeScriptInstanceMap[ud]

	if !ok {
		log.Panic("unable to find UserData instance")
	}

	instValue := reflect.ValueOf(classInst)

	methodName := RegisterState.TagDB.GetRegisteredMethodName(tag)

	instMethod := instValue.MethodByName(methodName)

	if instMethod == (reflect.Value{}) {
		log.Panic("method not found", StringField("method", methodName))
	}

	if methodName != "PhysicsProcess" {
		log.Info("call", StringField("method", methodName), AnyField("args", callArgs))
	}

	result := instMethod.Call(callArgs)

	resultSize := len(result)

	if resultSize == 0 {
		ret := NewVariantNil()
		return *(*C.godot_variant)(unsafe.Pointer(&ret))
	}

	if resultSize > 1 {
		log.Panic("unexpected multiple results", AnyField("result", result))
	}

	valueInterface := result[0].Interface()
	switch v := valueInterface.(type) {
	case Variant:
		return *(*C.godot_variant)(unsafe.Pointer(&v))
	}

	ret := NewVariantNil()
	return *(*C.godot_variant)(unsafe.Pointer(&ret))
}
