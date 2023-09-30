package gdextension

// #include <godot/gdextension_interface.h>
// #include "wrapped.h"
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type GDExtensionClass interface {
	Wrapped
}

type HasDestructor interface {
	Destroy()
}

//export GoCallback_GDExtensionBindingCreate
func GoCallback_GDExtensionBindingCreate(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer) unsafe.Pointer {
	typeName := C.GoString(p_type_name)
	log.Debug("GoCallback_GDExtensionBindingCreate called",
		zap.String("class", typeName),
	)
	fn, ok := gdNativeConstructors.Get(typeName)

	if !ok {
		log.Panic("unable to find GDExtension constructor", zap.String("type", typeName))
	}

	owner := (*GodotObject)(p_instance)

	inst := fn(owner).(Object)

	if inst == nil {
		log.Panic("no instance returned")
	}

	// objId := NewObjectID()

	// internal.gdNativeInstances.Set(objId, inst)

	return (unsafe.Pointer)(&inst)
}

//export GoCallback_GDExtensionBindingFree
func GoCallback_GDExtensionBindingFree(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_binding unsafe.Pointer) {
	// inst := *(*GDExtensionClass)(p_binding)

	// objId := ObjectID{Id: uint64(uintptr(p_binding))}

	// if _, ok := internal.gdNativeInstances.Get(objId); !ok {
	// 	log.Panic("GDExtensionClass instance not found to free", zap.Uint64("id", objId.Id))
	// }

	// inst := (*GDExtensionClass)(p_binding)

	// inst.GetGodotObjectOwner()

	// internal.gdNativeInstances.Delete(objId)
}

//export GoCallback_GDExtensionBindingReference
func GoCallback_GDExtensionBindingReference(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_reference bool) bool {
	return true
}

func getObjectInstanceBinding(engineObject *GodotObject) Object {
	if engineObject == nil {
		return nil
	}
	// Get existing instance binding, if one already exists.
	instPtr := (*Object)(CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		(GDExtensionObjectPtr)(engineObject),
		FFI.Token,
		nil))
	if instPtr != nil && *instPtr != nil {
		return *instPtr
	}
	snClassName := StringName{}
	cok := CallFunc_GDExtensionInterfaceObjectGetClassName(
		(GDExtensionConstObjectPtr)(engineObject),
		FFI.Library,
		(GDExtensionUninitializedStringNamePtr)(snClassName.nativePtr()),
	)
	if cok == 0 {
		log.Panic("failed to get class name",
			zap.Any("owner", engineObject),
		)
	}
	defer snClassName.Destroy()
	className := snClassName.ToUtf8()
	// const GDExtensionInstanceBindingCallbacks *binding_callbacks = nullptr;
	// Otherwise, try to look up the correct binding callbacks.
	cbs, ok := gdExtensionBindingGDExtensionInstanceBindingCallbacks.Get(className)
	if !ok {
		log.Warn("unable to find callbacks for Object")
		return nil
	}
	instPtr = (*Object)(CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		(GDExtensionObjectPtr)(engineObject),
		FFI.Token,
		&cbs))
	if instPtr == nil || *instPtr == nil {
		log.Panic("unable to get instance")
		return nil
	}
	wrapperClassName := (*instPtr).GetClassName()
	gdStrClassName := (*instPtr).GetClass()
	defer gdStrClassName.Destroy()
	log.Info("getObjectInstanceBinding casted",
		zap.String("class", gdStrClassName.ToUtf8()),
		zap.String("className", wrapperClassName),
	)
	return *instPtr
}
