package gdextension

// #include <godot/gdnative_interface.h>
// #include "wrapped.h"
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
	. "github.com/godot-go/godot-go/pkg/gdnative"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type GDClass interface {
	Wrapped
}

func gdClassInitializeClass(c GDClass) {
	gdClassRegisterVirtuals(c)
}

func gdClassRegisterInstanceBindingCallbacks(tn TypeName) {
	// substitute for:
	// static constexpr GDNativeInstanceBindingCallbacks ___binding_callbacks = {
	// 	___binding_create_callback,
	// 	___binding_free_callback,
	// 	___binding_reference_callback,
	// };
	cbs := NewGDNativeInstanceBindingCallbacks(
		(*[0]byte)(C.cgo_gdclass_binding_create_callback),
		(*[0]byte)(C.cgo_gdclass_binding_free_callback),
		(*[0]byte)(C.cgo_gdclass_binding_reference_callback),
	)

	_, ok := gdExtensionBindingGDNativeInstanceBindingCallbacks.Get(tn)

	if ok {
		log.Panic("Class with the same name already initialized", zap.String("class", (string)(tn)))
	}

	gdExtensionBindingGDNativeInstanceBindingCallbacks.Set(tn, (GDNativeInstanceBindingCallbacks)(cbs))
}

func GDClassFromGDExtensionClassInstancePtr(p_classinfo *ClassInfo, p_instance GDExtensionClassInstancePtr) GDClass {
	if (C.GDExtensionClassInstancePtr)(p_instance) == (C.GDExtensionClassInstancePtr)(nullptr) {
		v := reflect.New(p_classinfo.ClassType)
		return v.Interface().(GDClass)
	}

	return *(*GDClass)(p_instance)
}

// WrappedPostInitialize is equivalent to Wrapped::_postinitialize in godot-cpp
// this should only be called for GDClasses and not GDNativeClasses
func WrappedPostInitialize(tn TypeName, w Wrapped) {
	extensionClassName := (string)(tn)
	owner := w.GetGodotObjectOwner()
	inst := (GDExtensionClassInstancePtr)(unsafe.Pointer(&w))
	if len(extensionClassName) == 0 {
		log.Panic("extension class name cannot be empty",
			zap.String("w", fmt.Sprintf("%p", w)),
			zap.String("w.GetGodotObjectOwner()", fmt.Sprintf("%p", w.GetGodotObjectOwner())),
		)
	}

	GDNativeInterface_object_set_instance(internal.gdnInterface, (GDNativeObjectPtr)(owner), extensionClassName, inst)

	callbacks, ok := gdExtensionBindingGDNativeInstanceBindingCallbacks.Get(tn)

	if !ok {
		log.Panic("unable to retrieve binding callbacks", zap.String("type", (string)(tn)))
	}

	GDNativeInterface_object_set_instance_binding(internal.gdnInterface, (GDNativeObjectPtr)(owner), internal.token, (unsafe.Pointer)(inst), &callbacks)
}

// GoCallback_GDNativeExtensionClassCreateInstance is registered as a callback when a new GDScript instance is created.
//
//export GoCallback_GDNativeExtensionClassCreateInstance
func GoCallback_GDNativeExtensionClassCreateInstance(data unsafe.Pointer) C.GDNativeObjectPtr {
	tn := (TypeName)(C.GoString((*C.char)(data)))

	inst := CreateGDClassInstance(tn)

	return (C.GDNativeObjectPtr)(unsafe.Pointer(inst.GetGodotObjectOwner()))
}

func CreateGDClassInstance(tn TypeName) GDClass {
	ci, ok := gdRegisteredGDClasses.Get(tn)

	if !ok {
		log.Panic("type not found",
			zap.String("name", (string)(tn)),
			zap.String("dump", spew.Sdump(gdRegisteredGDClasses)),
		)
	}

	log.Debug("GoCallback_GDNativeExtensionClassCreateInstance called",
		zap.String("class_name", (string)(tn)),
		zap.Any("parent_name", ci.ParentName),
	)

	// create inherited GDNativeClass first
	owner := GDNativeInterface_classdb_construct_object(internal.gdnInterface, string(ci.ParentName))

	if owner == nil {
		log.Panic("owner is nil", zap.String("type_name", (string)(tn)))
	}

	// create GDClass
	reflectedInst := reflect.New(ci.ClassType)

	inst, ok := reflectedInst.Interface().(GDClass)

	if !ok {
		log.Panic("instance not a GDClass", zap.String("type_name", (string)(tn)))
	}

	object := (*GodotObject)(unsafe.Pointer(owner))

	id := GDNativeInterface_object_get_instance_id(internal.gdnInterface, owner)

	inst.SetGodotObjectOwner(object)

	WrappedPostInitialize(tn, inst)

	internal.gdClassInstances.Set(id, inst)

	log.Info("GDClass instance created",
		zap.Any("object_id", id),
		zap.String("class_name", (string)(tn)),
		zap.Any("parent_name", ci.ParentName),
		zap.String("inst", fmt.Sprintf("%p", inst)),
		zap.String("owner", fmt.Sprintf("%p", owner)),
		zap.String("object", fmt.Sprintf("%p", object)),
		zap.String("inst.GetGodotObjectOwner", fmt.Sprintf("%p", inst.GetGodotObjectOwner())),
	)

	return inst
}

//export GoCallback_GDNativeExtensionClassFreeInstance
func GoCallback_GDNativeExtensionClassFreeInstance(data unsafe.Pointer, ptr C.GDExtensionClassInstancePtr) {
	tn := (TypeName)(C.GoString((*C.char)(data)))

	// ptr is assigned in function WrappedPostInitialize as a (*Wrapped)
	w := *(*Wrapped)(unsafe.Pointer(ptr))

	log.Info("GoCallback_GDNativeExtensionClassFreeInstance called",
		zap.String("type_name", (string)(tn)),
		zap.String("ptr", fmt.Sprintf("%p", ptr)),
		zap.String("w", fmt.Sprintf("%p", w)),
		zap.String("w.GetGodotObjectOwner()", fmt.Sprintf("%p", w.GetGodotObjectOwner())),
	)

	id := GDNativeInterface_object_get_instance_id(internal.gdnInterface, (GDNativeObjectPtr)(unsafe.Pointer(w.GetGodotObjectOwner())))

	if _, ok := internal.gdClassInstances.Get(id); !ok {
		log.Panic("GDClass instance not found to free", zap.Any("id", id))
	}

	internal.gdClassInstances.Delete(id)

	log.Info("GDClass instance freed", zap.Any("id", id))
}

//export GoCallback_GDClassBindingCreate
func GoCallback_GDClassBindingCreate(p_token unsafe.Pointer, p_instance unsafe.Pointer) unsafe.Pointer {
	return nullptr
}

//export GoCallback_GDClassBindingFree
func GoCallback_GDClassBindingFree(p_token unsafe.Pointer, p_instance unsafe.Pointer, p_binding unsafe.Pointer) {
}

//export GoCallback_GDClassBindingReference
func GoCallback_GDClassBindingReference(p_token unsafe.Pointer, p_instance unsafe.Pointer, p_reference C.GDNativeBool) C.GDNativeBool {
	return 1
}

func gdClassRegisterVirtuals(c GDClass) {
	// TODO: figure out how to approach this
	// P.RegisterVirtuals()
}
