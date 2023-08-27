package gdextension

// #include <godot/gdextension_interface.h>
// #include "classdb_callback.h"
// #include "method_bind.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

//export GoCallback_ClassCreationInfoToString
func GoCallback_ClassCreationInfoToString(
	p_instance C.GDExtensionClassInstancePtr,
	r_is_valid *C.GDExtensionBool,
	p_out C.GDExtensionStringPtr) {
	log.Debug("GoCallback_ClassCreationInfoToString",
		zap.String("&p_instance", fmt.Sprintf("%p", p_instance)),
		zap.Reflect("p_instance", p_instance),
	)
	inst := ObjectClassFromGDExtensionClassInstancePtr((GDExtensionClassInstancePtr)(p_instance))
	className := inst.GetClassName()
	instanceId := inst.GetInstanceId()
	value := fmt.Sprintf("[ GDExtension::%s <--> Instance ID:%d ]", className, instanceId)
	GDExtensionStringPtrWithLatin1Chars((GDExtensionStringPtr)(p_out), value)
	var isValid C.uchar = 1
	r_is_valid = (*C.GDExtensionBool)(&isValid)
}

//export GoCallback_ClassCreationInfoGetVirtualCallWithData
func GoCallback_ClassCreationInfoGetVirtualCallWithData(pUserdata unsafe.Pointer, pName C.GDExtensionConstStringNamePtr) unsafe.Pointer {
	name := C.GoString((*C.char)(pUserdata))
	snMethodName := (*StringName)(unsafe.Pointer(pName))
	sMethodName := snMethodName.AsString()
	methodName := (&sMethodName).ToAscii()
	log.Info("GoCallback_ClassCreationInfoGetVirtualCallWithData called",
		zap.String("class_name_from_user_data", name),
		zap.String("method_name", methodName),
	)
	return pUserdata
}

//export GoCallback_ClassCreationInfoCallVirtualWithData
func GoCallback_ClassCreationInfoCallVirtualWithData(pInstance C.GDExtensionClassInstancePtr, pName C.GDExtensionConstStringNamePtr, pUserdata unsafe.Pointer, p_args *C.GDExtensionConstTypePtr, rRet C.GDExtensionTypePtr) {
	inst := ObjectClassFromGDExtensionClassInstancePtr((GDExtensionClassInstancePtr)(pInstance))
	if inst == nil {
		log.Panic("GDExtensionClassInstancePtr cannot be null")
	}
	className := inst.GetClassName()
	snMethodName := (*StringName)(unsafe.Pointer(pName))
	sMethodName := snMethodName.AsString()
	methodName := (&sMethodName).ToAscii()
	userData := C.GoString((*C.char)(pUserdata))
	log.Info("GoCallback_ClassCreationInfoCallVirtualWithData called",
		zap.String("type", className),
		zap.String("userData", userData),
		zap.String("method", methodName),
	)
	ci, ok := gdRegisteredGDClasses.Get(className)
	if !ok {
		log.Warn("class not found", zap.String("className", className))
		return
	}
	m, ok := ci.VirtualMethodMap[methodName]
	if !ok {
		methods := maps.Keys(ci.VirtualMethodMap)
		slices.Sort(methods)

		log.Info("no virtual method found",
			zap.String("className", className),
			zap.String("method", methodName),
			zap.Any("virtual_methods", methods),
		)
		return
	}
	m.Ptrcall(
		(GDExtensionClassInstancePtr)(pInstance),
		(*GDExtensionConstTypePtr)(p_args),
		(GDExtensionTypePtr)(rRet),
	)
}

// GoCallback_ClassCreationInfoCreateInstance is registered as a callback when a new GDScript instance is created.
//
//export GoCallback_ClassCreationInfoCreateInstance
func GoCallback_ClassCreationInfoCreateInstance(data unsafe.Pointer) C.GDExtensionObjectPtr {
	tn := C.GoString((*C.char)(data))
	inst := CreateGDClassInstance(tn)
	return (C.GDExtensionObjectPtr)(unsafe.Pointer(inst.GetGodotObjectOwner()))
}

//export GoCallback_ClassCreationInfoFreeInstance
func GoCallback_ClassCreationInfoFreeInstance(data unsafe.Pointer, ptr C.GDExtensionClassInstancePtr) {
	tn := C.GoString((*C.char)(data))
	// ptr is assigned in function WrappedPostInitialize as a (*Wrapped)
	inst := ObjectClassFromGDExtensionClassInstancePtr((GDExtensionClassInstancePtr)(ptr))
	goStr := inst.ToString()
	log.Info("GoCallback_ClassCreationInfoFreeInstance called",
		zap.String("type_name", tn),
		zap.String("ptr", fmt.Sprintf("%p", ptr)),
		zap.String("to_string", goStr.ToAscii()),
		zap.String("GodotObjectOwner()", fmt.Sprintf("%p", inst.GetGodotObjectOwner())),
	)
	id := CallFunc_GDExtensionInterfaceObjectGetInstanceId((GDExtensionConstObjectPtr)(unsafe.Pointer(inst.GetGodotObjectOwner())))
	if _, ok := internal.gdClassInstances.Get(id); !ok {
		log.Panic("GDClass instance not found to free", zap.Any("id", id))
	}
	internal.gdClassInstances.Delete(id)
	log.Info("GDClass instance freed", zap.Any("id", id))
}

//export GoCallback_ClassDBGetFunc
func GoCallback_ClassDBGetFunc(pInstance C.GDExtensionClassInstancePtr, pName C.GDExtensionConstStringNamePtr, rRet C.GDExtensionVariantPtr) C.GDExtensionBool {
	// wci := (*WrappedClassInstance)(unsafe.Pointer(pInstance))
	// if wci == nil {
	// 	return (C.GDExtensionBool)(0)
	// }
	// v := reflect.ValueOf(wci)
	// // TODO: get method and call
	// v.MethodByName("")
	log.Warn("TODO: GoCallback_ClassDBGetFunc not implemented")
	return (C.GDExtensionBool)(0)
}

//export GoCallback_ClassDBSetFunc
func GoCallback_ClassDBSetFunc(p_instance C.GDExtensionClassInstancePtr, p_name C.GDExtensionConstStringNamePtr, p_value C.GDExtensionConstVariantPtr) C.GDExtensionBool {
	// wci := (*WrappedClassInstance)(unsafe.Pointer(pInstance))
	// if wci == nil {
	// 	return (C.GDExtensionBool)(0)
	// }
	// v := reflect.ValueOf(wci)
	// // TODO: get method and call
	// v.MethodByName("")
	log.Warn("TODO: GoCallback_ClassDBSetFunc not implemented")
	return (C.GDExtensionBool)(0)
}
