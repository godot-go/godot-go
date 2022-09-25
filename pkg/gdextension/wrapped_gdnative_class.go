package gdextension

// #include <godot/gdnative_interface.h>
// #include "wrapped.h"
import "C"
import (
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type GDNativeClass interface {
	Wrapped
}

//export GoCallback_GDNativeBindingCreate
func GoCallback_GDNativeBindingCreate(typeName string, p_token unsafe.Pointer, p_instance unsafe.Pointer) unsafe.Pointer {
	tn := (TypeName)(typeName)

	fn, ok := gdNativeConstructors.Get(tn)

	if !ok {
		log.Panic("unable to find GDNative constructor", zap.String("type", (string)(tn)))
	}

	owner := (*GodotObject)(p_instance)

	inst := fn(owner)

	objId := NewObjectID()

	internal.gdNativeInstances.Set(objId, inst)

	return (unsafe.Pointer)(uintptr(objId.Id))
}

//export GoCallback_GDNativeBindingFree
func GoCallback_GDNativeBindingFree(typeName string, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_binding unsafe.Pointer) {
	objId := ObjectID{Id: uint64(uintptr(p_binding))}

	if _, ok := internal.gdNativeInstances.Get(objId); !ok {
		log.Panic("GDNativeClass instance not found to free", zap.Uint64("id", objId.Id))
	}

	// inst := (*GDNativeClass)(p_binding)

	// inst.GetGodotObjectOwner()

	internal.gdNativeInstances.Delete(objId)
}

//export GoCallback_GDNativeBindingReference
func GoCallback_GDNativeBindingReference(typeName string, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_reference bool) bool {
	return true
}
