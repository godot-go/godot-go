package gdextension

// #include <godot/gdextension_interface.h>
// #include "wrapped.h"
import "C"
import (
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type GDExtensionClass interface {
	Wrapped
}

//export GoCallback_GDExtensionBindingCreate
func GoCallback_GDExtensionBindingCreate(typeName string, p_token unsafe.Pointer, p_instance unsafe.Pointer) unsafe.Pointer {
	fn, ok := gdNativeConstructors.Get(typeName)

	if !ok {
		log.Panic("unable to find GDExtension constructor", zap.String("type", typeName))
	}

	owner := (*GodotObject)(p_instance)

	inst := fn(owner)

	objId := NewObjectID()

	internal.gdNativeInstances.Set(objId, inst)

	return (unsafe.Pointer)(uintptr(objId.Id))
}

//export GoCallback_GDExtensionBindingFree
func GoCallback_GDExtensionBindingFree(typeName string, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_binding unsafe.Pointer) {
	objId := ObjectID{Id: uint64(uintptr(p_binding))}

	if _, ok := internal.gdNativeInstances.Get(objId); !ok {
		log.Panic("GDExtensionClass instance not found to free", zap.Uint64("id", objId.Id))
	}

	// inst := (*GDExtensionClass)(p_binding)

	// inst.GetGodotObjectOwner()

	internal.gdNativeInstances.Delete(objId)
}

//export GoCallback_GDExtensionBindingReference
func GoCallback_GDExtensionBindingReference(typeName string, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_reference bool) bool {
	return true
}
