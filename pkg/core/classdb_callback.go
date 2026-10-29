package core

/*
Godot-Owned StringName Pointer Safety

Several callback functions in this file receive `GDExtensionConstStringNamePtr`
parameters from Godot (e.g., pName in Get/Set/ValidateProperty callbacks).
These pointers reference StringName objects owned and managed by Godot's internal
refcounted interned hash table.

CRITICAL RULE: NEVER call .Destroy() on these Godot-owned StringName pointers.
Doing so would decrement Godot's internal refcount, potentially corrupting the
shared StringName table and causing use-after-free or double-free errors.

Safe operations on Godot-owned StringName pointers:
  - Casting to *StringName and calling read-only methods (AsString(), ToUtf8(), etc.)
  - Passing to Godot API functions that expect GDExtensionConstStringNamePtr

The result of .AsString() returns a NEW String object that IS owned by the caller
and should be destroyed with defer when done.

See: Task 5.4 in fix-orphan-stringname change.
*/

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
	"runtime/cgo"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"github.com/godot-go/godot-go/pkg/util"
	"go.uber.org/zap"
)

//export GoCallback_ClassCreationInfoToString
func GoCallback_ClassCreationInfoToString(
	p_instance C.GDExtensionClassInstancePtr,
	r_is_valid *C.GDExtensionBool,
	p_out C.GDExtensionStringPtr) {
	wci := cgo.Handle(p_instance).Value().(*WrappedClassInstance)
	if wci == nil {
		log.Panic("wci should not be null")
	}
	value := resolveToString(wci.Instance)
	// Go strings are UTF-8. The Latin-1 constructor reinterprets each byte as a
	// Latin-1 code point, which corrupts anything outside ASCII; the old
	// ASCII-only default masked that.
	GDExtensionStringPtrWithUtf8Chars((GDExtensionStringPtr)(p_out), value)
	// Write the flag *through* the out pointer. The previous code rebound this
	// function's local copy of the pointer parameter, so the engine's bool never
	// changed, stayed false, and discarded the string we just constructed in
	// favour of the built-in Object::to_string().
	*r_is_valid = C.GDExtensionBool(1)
}

//export GoCallback_ClassCreationInfoGetVirtualCallWithData
func GoCallback_ClassCreationInfoGetVirtualCallWithData(pUserdata unsafe.Pointer, pName C.GDExtensionConstStringNamePtr) unsafe.Pointer {
	name := C.GoString((*C.char)(pUserdata))
	// pName is Godot-owned — safe to read via AsString(), NEVER Destroy() pName itself
	snMethodName := (*StringName)(pName)
	sMethodName := snMethodName.AsString()
	defer sMethodName.Destroy()
	methodName := sMethodName.ToUtf8()
	ci, ok := Internal.GDRegisteredGDClasses.Get(name)
	if !ok {
		log.Warn("class not found",
			zap.String("className", name),
			zap.String("method_name", methodName),
		)
		return nil
	}
	if _, ok := ci.VirtualMethodMap[methodName]; !ok {
		log.Debug("no virtual method found",
			zap.String("className", name),
			zap.String("method", methodName),
		)
		return nil
	}
	log.Info("GoCallback_ClassCreationInfoGetVirtualCallWithData called",
		zap.String("class_name_from_user_data", name),
		zap.String("method_name", methodName),
	)
	return pUserdata
}

//export GoCallback_ClassCreationInfoCallVirtualWithData
func GoCallback_ClassCreationInfoCallVirtualWithData(pInstance C.GDExtensionClassInstancePtr, pName C.GDExtensionConstStringNamePtr, pUserdata unsafe.Pointer, p_args *C.GDExtensionConstTypePtr, rRet C.GDExtensionTypePtr) {
	// inst := ObjectClassFromGDExtensionClassInstancePtr((GDExtensionClassInstancePtr)(pInstance))
	// if inst == nil {
	// 	log.Panic("GDExtensionClassInstancePtr cannot be null")
	// }
	wci := cgo.Handle(pInstance).Value().(*WrappedClassInstance)
	if wci == nil {
		log.Panic("wci should not be null")
	}
	inst := wci.Instance
	className := inst.GetClassName()
	// pName is Godot-owned — safe to read via AsString(), NEVER Destroy() pName itself
	snMethodName := (*StringName)(pName)
	sMethodName := snMethodName.AsString()
	defer sMethodName.Destroy()
	methodName := (&sMethodName).ToAscii()
	var userData string
	if pUserdata != nil {
		userData = C.GoString((*C.char)(pUserdata))
	} else {
		log.Warn("user data is nil")
	}
	log.Info("GoCallback_ClassCreationInfoCallVirtualWithData called",
		zap.String("type", className),
		zap.String("userData", userData),
		zap.String("method", methodName),
	)
	if len(methodName) == 0 {
		log.Panic("call virtual with data method was empty")
	}
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
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
	mb := m.GoMethodMetadata
	args := unsafe.Slice(
		(*GDExtensionConstTypePtr)(p_args),
		len(mb.GoArgumentTypes),
	)
	mb.Ptrcall(
		inst,
		args,
		(GDExtensionUninitializedTypePtr)(rRet),
	)
}

type ClassUserdata struct {
	callbacks  GDExtensionInstanceBindingCallbacks
	className  string
	createFunc GDClassGoConstructor
	typeData   reflect.Type
}

func (d *ClassUserdata) HasError() error {
	if d == nil {
		return fmt.Errorf("ClassUserdata is null")
	}
	if len(d.className) == 0 {
		return fmt.Errorf("className is empty")
	}
	if d.createFunc == nil {
		return fmt.Errorf("createFunc is nil")
	}
	if d.typeData == nil {
		return fmt.Errorf("typeData is null")
	}
	return nil
}

// GoCallback_ClassCreationInfoCreateInstance2 is registered as a callback when a new GDScript instance is created.
//
//export GoCallback_ClassCreationInfoCreateInstance2
func GoCallback_ClassCreationInfoCreateInstance2(data unsafe.Pointer, p_notify_postinitialize C.GDExtensionBool) C.GDExtensionObjectPtr {
	tn := C.GoString((*C.char)(data))
	inst := CreateGDClassInstance2(tn)
	return (C.GDExtensionObjectPtr)(unsafe.Pointer(inst.GetGodotObjectOwner()))
}

//export GoCallback_ClassCreationInfoFreeInstance
func GoCallback_ClassCreationInfoFreeInstance(data unsafe.Pointer, ptr C.GDExtensionClassInstancePtr) {
	if ptr == nil {
		return
	}
	ptrHandle := cgo.Handle(unsafe.Pointer(ptr))
	wrapped := ptrHandle.Value().(*WrappedClassInstance)
	// Deferred calls run LIFO: Delete is registered first so it runs LAST —
	// after Destroy and after the SyncMap guard below. Any re-entrant callback
	// issued while Destroy() executes still resolves a live Handle (openspec
	// fix-instance-handle-lifecycle, issue #152). Godot clears its copies of
	// _extension/_extension_instance immediately after free_instance returns
	// (Object::_predelete), so the engine issues no further callback with the
	// stale pointer; binding-slot readers already degrade a dead handle to a
	// typed error (objectFromBindingPtr).
	defer ptrHandle.Delete()
	// Defer Destroy so the SyncMap guard below catches re-entrant calls.
	defer wrapped.Instance.Destroy()
	id := CallFunc_GDExtensionInterfaceObjectGetInstanceId((GDExtensionConstObjectPtr)(unsafe.Pointer(wrapped.Instance.GetGodotObjectOwner())))
	if _, ok := Internal.GDClassInstances.Get(id); !ok {
		log.Panic("GDClass instance not found to free", zap.Any("id", id))
	}
	Internal.GDClassInstances.Delete(id)
}

//export GoCallback_ClassCreationInfoGetPropertyList
func GoCallback_ClassCreationInfoGetPropertyList(pInstance C.GDExtensionClassInstancePtr, rCount *C.uint32_t) *C.GDExtensionPropertyInfo {
	wci := cgo.Handle(pInstance).Value().(*WrappedClassInstance)
	if wci == nil {
		*rCount = (C.uint32_t)(0)
		return (*C.GDExtensionPropertyInfo)(nil)
	}
	gdStrClass := wci.Instance.GetClass()
	defer gdStrClass.Destroy()
	className := gdStrClass.ToUtf8()
	log.Debug("GoCallback_ClassCreationInfoGetPropertyList called",
		zap.String("class", className),
	)
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
	if !ok {
		log.Panic("invalid registered GDClass",
			zap.String("class", className),
		)
	}
	// Query-owned C array (openspec fix-class-property-list-lifecycle): built
	// fresh per query from the class-owned spec cells, so no Go memory is
	// exposed and nothing is pinned. The engine deep-copies the elements and
	// calls FreePropertyList2 in the same scope (Object::get_property_list),
	// which releases the array. Repeated queries always see the full list.
	ptr := allocPropertyArray(ci.propertySpecs)
	if ptr == nil {
		*rCount = (C.uint32_t)(0)
		return (*C.GDExtensionPropertyInfo)(nil)
	}
	*rCount = (C.uint32_t)(len(ci.propertySpecs))
	return ptr
}

//export GoCallback_ClassCreationInfoFreePropertyList2
func GoCallback_ClassCreationInfoFreePropertyList2(pInstance C.GDExtensionClassInstancePtr, pList *C.GDExtensionPropertyInfo, pCount C.uint32_t) {
	// Releases only the query-owned C array allocated by
	// GoCallback_ClassCreationInfoGetPropertyList. The class-level string
	// cells are owned by ClassInfo and destroyed with the registration, so
	// query timing cannot double-free anything
	// (openspec fix-class-property-list-lifecycle).
	freePropertyArray(pList)
}

//export GoCallback_ClassCreationInfoPropertyCanRevert
func GoCallback_ClassCreationInfoPropertyCanRevert(p_instance C.GDExtensionClassInstancePtr, p_name C.GDExtensionConstStringNamePtr) C.GDExtensionBool {
	wci := cgo.Handle(p_instance).Value().(*WrappedClassInstance)
	if wci == nil {
		return 0
	}
	gdStrClass := wci.Instance.GetClass()
	defer gdStrClass.Destroy()
	className := gdStrClass.ToUtf8()
	// p_name is Godot-owned — NEVER Destroy() it
	gdName := (*StringName)(p_name)
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
	if !ok {
		// During teardown an instance can report its parent class name
		// (e.g. "Control") which is not a registered GDClass; treat it
		// as absence of support instead of failing.
		return 0
	}
	mcmi, ok := ci.VirtualMethodMap["_property_can_revert"]
	if !ok {
		// Revert queries are frequent (the inspector asks per property);
		// report absence immediately without logging.
		return 0
	}
	if cb := mcmi.GoMethodMetadata.cbPropCanRevert; cb != nil {
		// Compiled callback shape (design D6): the StringName is the same
		// Godot-owned borrow the reflective path boxes, passed by value.
		if cb(wci.Instance, *gdName) {
			return 1
		}
		return 0
	}
	args := []reflect.Value{
		reflect.ValueOf(wci.Instance),
		reflect.ValueOf(*gdName),
	}
	reflectedRet := mcmi.GoMethodMetadata.Func.Call(args)
	if reflectedRet[0].Bool() {
		return 1
	}
	return 0
}

//export GoCallback_ClassCreationInfoPropertyGetRevert
func GoCallback_ClassCreationInfoPropertyGetRevert(p_instance C.GDExtensionClassInstancePtr, p_name C.GDExtensionConstStringNamePtr, r_ret C.GDExtensionVariantPtr) C.GDExtensionBool {
	wci := cgo.Handle(p_instance).Value().(*WrappedClassInstance)
	if wci == nil {
		return 0
	}
	gdStrClass := wci.Instance.GetClass()
	defer gdStrClass.Destroy()
	className := gdStrClass.ToUtf8()
	// p_name is Godot-owned — NEVER Destroy() it
	gdName := (*StringName)(p_name)
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
	if !ok {
		// During teardown an instance can report its parent class name
		// (e.g. "Control") which is not a registered GDClass; treat it
		// as absence of support instead of failing.
		return 0
	}
	mcmi, ok := ci.VirtualMethodMap["_property_get_revert"]
	if !ok {
		// Revert queries are frequent (the inspector asks per property);
		// report absence immediately without logging.
		return 0
	}
	if cb := mcmi.GoMethodMetadata.cbPropGetRevert; cb != nil {
		// Compiled callback shape (design D6): the concrete (Variant, bool)
		// return makes the reflective path's interface-cast guard
		// structurally guaranteed.
		v, found := cb(wci.Instance, *gdName)
		if !found {
			return 0
		}
		*(*Variant)(unsafe.Pointer(r_ret)) = v
		return 1
	}
	args := []reflect.Value{
		reflect.ValueOf(wci.Instance),
		reflect.ValueOf(*gdName),
	}
	reflectedRet := mcmi.GoMethodMetadata.Func.Call(args)
	v, ok := reflectedRet[0].Interface().(Variant)
	if !ok {
		log.Panic("invalid return value: expected Variant",
			zap.String("class", className),
		)
	}
	if !reflectedRet[1].Bool() {
		return 0
	}
	*(*Variant)(unsafe.Pointer(r_ret)) = v
	return 1
}

//export GoCallback_ClassCreationInfoValidateProperty
func GoCallback_ClassCreationInfoValidateProperty(pInstance C.GDExtensionClassInstancePtr, pProperty *C.GDExtensionPropertyInfo) C.GDExtensionBool {
	wci := cgo.Handle(pInstance).Value().(*WrappedClassInstance)
	if wci == nil {
		return 0
	}
	gdStrClass := wci.Instance.GetClass()
	defer gdStrClass.Destroy()
	className := gdStrClass.ToUtf8()
	log.Debug("GoCallback_ClassCreationInfoValidateProperty called",
		zap.String("class", className),
	)
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
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
	wci := cgo.Handle(p_instance).Value().(*WrappedClassInstance)
	if wci == nil {
		return
	}
	gdStrClass := wci.Instance.GetClass()
	defer gdStrClass.Destroy()
	className := gdStrClass.ToUtf8()
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
	if !ok {
		// During teardown an instance can report its parent class name
		// (e.g. "Control") which is not a registered GDClass; treat it
		// as a silent no-op instead of failing.
		return
	}
	mcmi, ok := ci.VirtualMethodMap["_notification"]
	if !ok {
		// Notifications are high-frequency; missing registration is a
		// silent no-op without logging.
		return
	}
	if cb := mcmi.GoMethodMetadata.cbNotification; cb != nil {
		// Compiled callback shape (design D6): the same int32 what and
		// GDExtensionBool-to-Go-bool conversion the reflective boxes carry.
		cb(wci.Instance, int32(p_what), p_reversed != 0)
		return
	}
	args := []reflect.Value{
		reflect.ValueOf(wci.Instance),
		reflect.ValueOf(int32(p_what)),
		reflect.ValueOf(p_reversed != 0),
	}
	mcmi.GoMethodMetadata.Func.Call(args)
}

//export GoCallback_ClassCreationInfoGet
func GoCallback_ClassCreationInfoGet(pInstance C.GDExtensionClassInstancePtr, pName C.GDExtensionConstStringNamePtr, rRet C.GDExtensionVariantPtr) C.GDExtensionBool {
	wci := cgo.Handle(pInstance).Value().(*WrappedClassInstance)
	if wci == nil {
		return 0
	}
	gdStrClass := wci.Instance.GetClass()
	defer gdStrClass.Destroy()
	className := gdStrClass.ToUtf8()
	// pName is Godot-owned — safe to read via ToUtf8(), NEVER Destroy() pName itself
	gdName := (*StringName)(pName)
	name := gdName.ToUtf8()
	log.Debug("GoCallback_ClassCreationInfoGet called",
		zap.String("class", className),
		zap.String("method_name", name),
	)
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
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
	if cb := mcmi.GoMethodMetadata.cbPropGet; cb != nil {
		// Compiled callback shape (design D6): UTF-8 name in exactly as the
		// reflective box, (value, exists) out with the same
		// false-means-unhandled contract; no per-call diagnostic allocates
		// here, so the reflective path's ToString logging stays reflective-
		// only.
		v, found := cb(wci.Instance, name)
		if !found {
			log.Debug("_get call returned false")
			return 0
		}
		*(*Variant)(unsafe.Pointer(rRet)) = v
		return 1
	}
	args := []reflect.Value{
		reflect.ValueOf(wci.Instance),
		reflect.ValueOf(name),
	}
	reflectedRet := mcmi.GoMethodMetadata.Func.Call(args)
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
	log.Info("reflect method called",
		zap.Stringer("ret", lazyText(func() string { return util.ReflectValueSliceToString(reflectedRet) })),
		// Variant-to-String conversion is an engine crossing, so it belongs
		// to the lazy payload: a discarded entry must not touch the engine.
		// v stays valid until the rRet assignment below copies the cell out,
		// and marshaling runs before this function returns.
		zap.Stringer("v", lazyText(func() string {
			s := v.ToString()
			defer s.Destroy()
			return s.ToUtf8()
		})),
	)
	*(*Variant)(unsafe.Pointer(rRet)) = v
	return 1
}

//export GoCallback_ClassCreationInfoSet
func GoCallback_ClassCreationInfoSet(pInstance C.GDExtensionClassInstancePtr, pName C.GDExtensionConstStringNamePtr, pValue C.GDExtensionConstVariantPtr) C.GDExtensionBool {
	wci := cgo.Handle(pInstance).Value().(*WrappedClassInstance)
	if wci == nil {
		return 0
	}
	gdStrClass := wci.Instance.GetClass()
	defer gdStrClass.Destroy()
	className := gdStrClass.ToUtf8()
	// pName is Godot-owned — safe to read via ToUtf8(), NEVER Destroy() pName itself
	gdName := (*StringName)(pName)
	name := gdName.ToUtf8()
	v := VariantViewFromConstPtr((GDExtensionConstVariantPtr)(pValue))
	log.Info("GoCallback_ClassCreationInfoSet called",
		zap.String("class", className),
		zap.String("name", name),
		zap.String("value", v.ToGoString()),
	)
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
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
	if cb := mcmi.GoMethodMetadata.cbPropSet; cb != nil {
		// Compiled callback shape (design D6): the same UTF-8 name and
		// engine-owned Variant view the reflective boxes carry.
		if cb(wci.Instance, name, v) {
			return 1
		}
		return 0
	}
	args := []reflect.Value{
		reflect.ValueOf(wci.Instance),
		reflect.ValueOf(name),
		reflect.ValueOf(v),
	}
	reflectedRet := mcmi.GoMethodMetadata.Func.Call(args)
	log.Info("reflect method called",
		zap.Stringer("ret", lazyText(func() string { return util.ReflectValueSliceToString(reflectedRet) })),
	)
	if !reflectedRet[0].Bool() {
		return 0
	}
	return 1
}
