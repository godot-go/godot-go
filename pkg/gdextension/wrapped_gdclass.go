package gdextension

// #include <godot/gdextension_interface.h>
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

func gdClassRegisterInstanceBindingCallbacks(tn string) {
	// substitute for:
	// static constexpr GDExtensionInstanceBindingCallbacks ___binding_callbacks = {
	// 	___binding_create_callback,
	// 	___binding_free_callback,
	// 	___binding_reference_callback,
	// };
	cbs := NewGDExtensionInstanceBindingCallbacks(
		(*[0]byte)(C.cgo_gdclass_binding_create_callback),
		(*[0]byte)(C.cgo_gdclass_binding_free_callback),
		(*[0]byte)(C.cgo_gdclass_binding_reference_callback),
	)

	_, ok := gdExtensionBindingGDExtensionInstanceBindingCallbacks.Get(tn)

	if ok {
		log.Panic("Class with the same name already initialized", zap.String("class", tn))
	}

	gdExtensionBindingGDExtensionInstanceBindingCallbacks.Set(tn, (GDExtensionInstanceBindingCallbacks)(cbs))
}

func GDClassFromGDExtensionClassInstancePtr(p_classinfo *ClassInfo, p_instance GDExtensionClassInstancePtr) GDClass {
	if (C.GDExtensionClassInstancePtr)(p_instance) == (C.GDExtensionClassInstancePtr)(nullptr) {
		v := reflect.New(p_classinfo.ClassType)
		return v.Interface().(GDClass)
	}

	return *(*GDClass)(p_instance)
}

// WrappedPostInitialize is equivalent to Wrapped::_postinitialize in godot-cpp
// this should only be called for GDClasses and not GDExtensionClasses
func WrappedPostInitialize(extensionClassName string, w Wrapped) {
	owner := w.GetGodotObjectOwner()
	inst := (GDExtensionClassInstancePtr)(unsafe.Pointer(&w))
	if len(extensionClassName) == 0 {
		log.Panic("extension class name cannot be empty",
			zap.String("w", fmt.Sprintf("%p", w)),
			zap.String("w.GetGodotObjectOwner()", fmt.Sprintf("%p", w.GetGodotObjectOwner())),
		)
	}

	GDExtensionInterface_object_set_instance(
		internal.gdnInterface,
		(GDExtensionObjectPtr)(owner),
		NewStringNameWithLatin1Chars(extensionClassName).AsGDExtensionStringNamePtr(),
		inst,
	)

	callbacks, ok := gdExtensionBindingGDExtensionInstanceBindingCallbacks.Get(extensionClassName)

	if !ok {
		log.Panic("unable to retrieve binding callbacks", zap.String("type", extensionClassName))
	}

	GDExtensionInterface_object_set_instance_binding(internal.gdnInterface, (GDExtensionObjectPtr)(owner), internal.token, (unsafe.Pointer)(inst), &callbacks)
}

// GoCallback_GDExtensionClassCreateInstance is registered as a callback when a new GDScript instance is created.
//
//export GoCallback_GDExtensionClassCreateInstance
func GoCallback_GDExtensionClassCreateInstance(data unsafe.Pointer) C.GDExtensionObjectPtr {
	tn := C.GoString((*C.char)(data))

	inst := CreateGDClassInstance(tn)

	return (C.GDExtensionObjectPtr)(unsafe.Pointer(inst.GetGodotObjectOwner()))
}

func CreateGDClassInstance(tn string) GDClass {
	ci, ok := gdRegisteredGDClasses.Get(tn)

	if !ok {
		log.Panic("type not found",
			zap.String("name", tn),
			zap.String("dump", spew.Sdump(gdRegisteredGDClasses)),
		)
	}

	log.Debug("GoCallback_GDExtensionClassCreateInstance called",
		zap.String("class_name", tn),
		zap.Any("parent_name", ci.ParentName),
	)

	// create inherited GDExtensionClass first
	owner := GDExtensionInterface_classdb_construct_object(
		internal.gdnInterface,
		NewStringNameWithLatin1Chars(ci.ParentName).AsGDExtensionStringNamePtr(),
	)

	if owner == nil {
		log.Panic("owner is nil", zap.String("type_name", tn))
	}

	// create GDClass
	reflectedInst := reflect.New(ci.ClassType)

	inst, ok := reflectedInst.Interface().(GDClass)

	if !ok {
		log.Panic("instance not a GDClass", zap.String("type_name", tn))
	}

	object := (*GodotObject)(unsafe.Pointer(owner))

	id := GDExtensionInterface_object_get_instance_id(internal.gdnInterface, (GDExtensionConstObjectPtr)(unsafe.Pointer(owner)))

	inst.SetGodotObjectOwner(object)

	WrappedPostInitialize(tn, inst)

	internal.gdClassInstances.Set(id, inst)

	log.Info("GDClass instance created",
		zap.Any("object_id", id),
		zap.String("class_name", tn),
		zap.Any("parent_name", ci.ParentName),
		zap.String("inst", fmt.Sprintf("%p", inst)),
		zap.String("owner", fmt.Sprintf("%p", owner)),
		zap.String("object", fmt.Sprintf("%p", object)),
		zap.String("inst.GetGodotObjectOwner", fmt.Sprintf("%p", inst.GetGodotObjectOwner())),
	)

	return inst
}

//export GoCallback_GDExtensionClassFreeInstance
func GoCallback_GDExtensionClassFreeInstance(data unsafe.Pointer, ptr C.GDExtensionClassInstancePtr) {
	tn := C.GoString((*C.char)(data))

	// ptr is assigned in function WrappedPostInitialize as a (*Wrapped)
	w := *(*Wrapped)(unsafe.Pointer(ptr))

	log.Info("GoCallback_GDExtensionClassFreeInstance called",
		zap.String("type_name", tn),
		zap.String("ptr", fmt.Sprintf("%p", ptr)),
		zap.String("w", fmt.Sprintf("%p", w)),
		zap.String("w.GetGodotObjectOwner()", fmt.Sprintf("%p", w.GetGodotObjectOwner())),
	)

	id := GDExtensionInterface_object_get_instance_id(internal.gdnInterface, (GDExtensionConstObjectPtr)(unsafe.Pointer(w.GetGodotObjectOwner())))

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
func GoCallback_GDClassBindingReference(p_token unsafe.Pointer, p_instance unsafe.Pointer, p_reference C.GDExtensionBool) C.GDExtensionBool {
	return 1
}

func gdClassRegisterVirtuals(c GDClass) {
	// TODO: figure out how to approach this
	// P.RegisterVirtuals()
}
