package gdextension

// #include <godot/gdextension_interface.h>
// #include "classdb_callback.h"
// #include "method_bind.h"
// #include <stdint.h>
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"github.com/godot-go/godot-go/pkg/util"
	"go.uber.org/zap"
)

//export GoCallback_ClassCreationInfoToString
func GoCallback_ClassCreationInfoToString(
	p_instance C.GDExtensionClassInstancePtr,
	r_is_valid *C.GDExtensionBool,
	p_out C.GDExtensionStringPtr) {
	log.Debug("GoCallback_ClassCreationInfoToString",
		zap.String("&p_instance", fmt.Sprintf("%p", p_instance)),
		// zap.Reflect("p_instance", p_instance),
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
	// name := C.GoString((*C.char)(pUserdata))
	// snMethodName := (*StringName)(unsafe.Pointer(pName))
	// sMethodName := snMethodName.AsString()
	// methodName := (&sMethodName).ToUft8()
	// log.Debug("GoCallback_ClassCreationInfoGetVirtualCallWithData called",
	// 	zap.String("class_name_from_user_data", name),
	// 	zap.String("method_name", methodName),
	// )
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
	log.Debug("GoCallback_ClassCreationInfoCallVirtualWithData called",
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
		log.Debug("no virtual method found",
			zap.String("className", className),
			zap.String("method", methodName),
		)
		return
	}
	mb := m.MethodBind
	args := unsafe.Slice(
		(*GDExtensionConstTypePtr)(unsafe.Pointer(p_args)),
		len(mb.MethodMetadata.GoArgumentTypes),
	)
	mb.Ptrcall(
		inst,
		args,
		(GDExtensionUninitializedTypePtr)(rRet),
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

//export GoCallback_ClassCreationInfoGetPropertyList
func GoCallback_ClassCreationInfoGetPropertyList(pInstance C.GDExtensionClassInstancePtr, rCount *C.uint32_t) *C.GDExtensionPropertyInfo {
	wci := (*WrappedClassInstance)(unsafe.Pointer(pInstance))
	if wci == nil {
		*rCount = (C.uint32_t)(0)
		return (*C.GDExtensionPropertyInfo)(nil)
	}

	gdStrClass := wci.Instance.GetClass()
	className := gdStrClass.ToUtf8()
	log.Debug("GoCallback_ClassCreationInfoGetPropertyList called",
		zap.String("class", className),
	)
	ci, ok := gdRegisteredGDClasses.Get(className)
	if !ok {
		log.Panic("invalid registered GDClass",
			zap.String("class", className),
		)
	}
	if ci.PropertyList == nil {
		*rCount = (C.uint32_t)(0)
		return (*C.GDExtensionPropertyInfo)(nil)
	}
	*rCount = (C.uint32_t)(len(ci.PropertyList))
	return (*C.GDExtensionPropertyInfo)(unsafe.Pointer(unsafe.SliceData(ci.PropertyList)))
}

//export GoCallback_ClassCreationInfoFreePropertyList
func GoCallback_ClassCreationInfoFreePropertyList(pInstance C.GDExtensionClassInstancePtr, pList *C.GDExtensionPropertyInfo) {

}

//export GoCallback_ClassCreationInfoPropertyCanRevert
func GoCallback_ClassCreationInfoPropertyCanRevert(p_instance C.GDExtensionClassInstancePtr, p_name C.GDExtensionConstStringNamePtr) C.GDExtensionBool {
	return 0
}

//export GoCallback_ClassCreationInfoPropertyGetRevert
func GoCallback_ClassCreationInfoPropertyGetRevert(p_instance C.GDExtensionClassInstancePtr, p_name C.GDExtensionConstStringNamePtr, r_ret C.GDExtensionVariantPtr) C.GDExtensionBool {
	return 0
}

//export GoCallback_ClassCreationInfoValidateProperty
func GoCallback_ClassCreationInfoValidateProperty(pInstance C.GDExtensionClassInstancePtr, pProperty *C.GDExtensionPropertyInfo) C.GDExtensionBool {
	wci := (*WrappedClassInstance)(unsafe.Pointer(pInstance))
	if wci == nil {
		return 0
	}

	gdStrClass := wci.Instance.GetClass()
	className := gdStrClass.ToUtf8()
	log.Debug("GoCallback_ClassCreationInfoValidateProperty called",
		zap.String("class", className),
	)
	ci, ok := gdRegisteredGDClasses.Get(className)
	if !ok {
		log.Panic("invalid registered GDClass",
			zap.String("class", className),
		)
	}
	prop := (*GDExtensionPropertyInfo)(unsafe.Pointer(pProperty))
	if ci.ValidateProperty == nil {
		return 0
	}

	ci.ValidateProperty(prop)
	return 1
}

//export GoCallback_ClassCreationInfoNotification
func GoCallback_ClassCreationInfoNotification(p_instance C.GDExtensionClassInstancePtr, p_what C.int32_t, p_reversed C.GDExtensionBool) {

}

//export GoCallback_ClassCreationInfoGet
func GoCallback_ClassCreationInfoGet(pInstance C.GDExtensionClassInstancePtr, pName C.GDExtensionConstStringNamePtr, rRet C.GDExtensionVariantPtr) C.GDExtensionBool {
	wci := (*WrappedClassInstance)(unsafe.Pointer(pInstance))
	if wci == nil {
		return 0
	}
	gdStrClass := wci.Instance.GetClass()
	className := gdStrClass.ToUtf8()
	gdName := NewStringNameWithGDExtensionConstStringNamePtr((GDExtensionConstStringNamePtr)(pName))
	name := gdName.ToUtf8()
	log.Debug("GoCallback_ClassCreationInfoGet called",
		zap.String("class", className),
		zap.String("method_name", name),
	)
	ci, ok := gdRegisteredGDClasses.Get(className)
	if !ok {
		log.Panic("invalid registered GDClass",
			zap.String("class", className),
			zap.String("method_name", name),
		)
	}
	mcmi, ok := ci.VirtualMethodMap["_get"]
	if !ok {
		log.Info("no V_Get method registered",
			zap.String("class", className),
			zap.String("method_name", name),
		)
		return 0
	}
	args := []reflect.Value{
		reflect.ValueOf(wci.Instance),
		reflect.ValueOf(name),
	}
	reflectedRet := mcmi.MethodBind.PtrcallFunc.Call(args)
	v, ok := reflectedRet[0].Interface().(Variant)
	if !ok {
		log.Panic("invalid return value: expected Variant",
			zap.String("class", name),
		)
	}
	if !reflectedRet[1].Bool() {
		log.Debug("_get call returned false")
		return 0
	}
	gdStrV := v.ToString()
	defer gdStrV.Destroy()
	log.Info("reflect method called",
		zap.String("ret", util.ReflectValueSliceToString(reflectedRet)),
		zap.String("v", gdStrV.ToUtf8()),
	)
	copyVariantWithGDExtensionTypePtr((GDExtensionUninitializedVariantPtr)(rRet), v.nativeConstPtr())
	return 1
}

//export GoCallback_ClassCreationInfoSet
func GoCallback_ClassCreationInfoSet(pInstance C.GDExtensionClassInstancePtr, pName C.GDExtensionConstStringNamePtr, pValue C.GDExtensionConstVariantPtr) C.GDExtensionBool {
	wci := (*WrappedClassInstance)(unsafe.Pointer(pInstance))
	if wci == nil {
		return 0
	}
	gdStrClass := wci.Instance.GetClass()
	className := gdStrClass.ToUtf8()
	gdName := NewStringNameWithGDExtensionConstStringNamePtr((GDExtensionConstStringNamePtr)(pName))
	name := gdName.ToUtf8()
	v := NewVariantCopyWithGDExtensionConstVariantPtr((GDExtensionConstVariantPtr)(pValue))
	log.Info("GoCallback_ClassCreationInfoSet called",
		zap.String("class", className),
		zap.String("name", name),
		zap.String("value", v.ToGoString()),
	)
	ci, ok := gdRegisteredGDClasses.Get(className)
	if !ok {
		log.Panic("invalid registered GDClass",
			zap.String("class", name),
		)
	}
	mcmi, ok := ci.VirtualMethodMap["_set"]
	if !ok {
		log.Info("no V_Set method registered",
			zap.String("class", name),
		)
		return 0
	}
	args := []reflect.Value{
		reflect.ValueOf(wci.Instance),
		reflect.ValueOf(name),
		reflect.ValueOf(v),
	}
	reflectedRet := mcmi.MethodBind.PtrcallFunc.Call(args)
	log.Info("reflect method called",
		zap.String("ret", util.ReflectValueSliceToString(reflectedRet)),
	)
	if !reflectedRet[0].Bool() {
		return 0
	}
	return 1
}
