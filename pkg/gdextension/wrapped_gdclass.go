package gdextension

// #include <godot/gdextension_interface.h>
// #include "wrapped.h"
import "C"
import (
	"fmt"
	"reflect"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type GetPropertyFunc func(string, *Variant)
type SetPropertyFunc func(string, *Variant)

type GDClass interface {
	Wrapped
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

// func GDClassFromGDExtensionClassInstancePtr(p_instance GDExtensionClassInstancePtr) GDClass {
// 	// if (C.GDExtensionClassInstancePtr)(p_instance) == (C.GDExtensionClassInstancePtr)(nullptr) {
// 	// 	log.Info("GDClass To GDExtensionClassInstancePtr: new instance",
// 	// 		zap.String("class_info", p_classinfo.String()),
// 	// 	)
// 	// 	v := reflect.New(p_classinfo.ClassType)
// 	// 	return v.Interface().(GDClass)
// 	// }

// 	// log.Info("GDClass To GDExtensionClassInstancePtr: casted",
// 	// 	zap.String("class_info", p_classinfo.String()),
// 	// )
// 	// return *(*GDClass)(p_instance)

// 	wci := (*WrappedClassInstance)(unsafe.Pointer(p_instance))

// 	return wci.Instance
// }

func ObjectClassFromGDExtensionClassInstancePtr(p_instance GDExtensionClassInstancePtr) Object {
	if p_instance == nil {
		return nil
	}

	wci := (*WrappedClassInstance)(unsafe.Pointer(p_instance))

	if wci.Instance == nil {
		log.Panic("unexpected nil instance")
		return nil
	}

	return wci.Instance
}

// WrappedPostInitialize is equivalent to Wrapped::_postinitialize in godot-cpp
// this should only be called for GDClasses and not GDExtensionClasses
func WrappedPostInitialize(extensionClassName string, w Wrapped) {
	owner := w.GetGodotObjectOwner()
	if len(extensionClassName) == 0 {
		log.Panic("extension class name cannot be empty",
			zap.String("w", fmt.Sprintf("%p", w)),
			zap.String("w.GetGodotObjectOwner()", fmt.Sprintf("%p", owner)),
		)
	}

	snExtensionClassName := NewStringNameWithLatin1Chars(extensionClassName)
	defer snExtensionClassName.Destroy()

	callbacks, ok := gdExtensionBindingGDExtensionInstanceBindingCallbacks.Get(extensionClassName)

	if !ok {
		log.Panic("unable to retrieve binding callbacks", zap.String("type", extensionClassName))
	}

	cnPtr := snExtensionClassName.AsGDExtensionConstStringNamePtr()

	obj, ok := w.(Object)

	if !ok {
		log.Panic("unable to cast to Object")
	}

	inst := &WrappedClassInstance{
		Instance: obj,
	}

	if cnPtr != nil {
		CallFunc_GDExtensionInterfaceObjectSetInstance(
			(GDExtensionObjectPtr)(owner),
			cnPtr,
			(GDExtensionClassInstancePtr)(unsafe.Pointer(inst)),
		)
	}

	CallFunc_GDExtensionInterfaceObjectSetInstanceBinding(
		(GDExtensionObjectPtr)(owner),
		unsafe.Pointer(FFI.Token),
		unsafe.Pointer(inst),
		&callbacks,
	)
}

func CreateGDClassInstance(tn string) GDClass {
	ci, ok := gdRegisteredGDClasses.Get(tn)

	if !ok {
		log.Panic("type not found",
			zap.String("name", tn),
			zap.String("dump", spew.Sdump(gdRegisteredGDClasses)),
		)
	}

	log.Debug("CreateGDClassInstance called",
		zap.String("class_name", tn),
		zap.Any("parent_name", ci.ParentName),
	)

	snParentName := NewStringNameWithLatin1Chars(ci.ParentName)
	defer snParentName.Destroy()

	// create inherited GDExtensionClass first
	owner := CallFunc_GDExtensionInterfaceClassdbConstructObject(
		snParentName.AsGDExtensionConstStringNamePtr(),
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

	id := CallFunc_GDExtensionInterfaceObjectGetInstanceId((GDExtensionConstObjectPtr)(unsafe.Pointer(owner)))

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
