package gdnative

// #include <godot/gdnative_interface.h>
// #include "gdnative_binding_wrapper.h"
import "C"
import (
	"reflect"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type GDNativeClass interface {
	Wrapped
	GetClassStatic() TypeName
	GetParentClassStatic() TypeName
}

func gdNativeClassInitializeClass(c GDNativeClass) {
}

func GoCallback_GDNativeBindingCreate[T GDNativeClass](p_token unsafe.Pointer, p_instance unsafe.Pointer) unsafe.Pointer {
	/*
		Notes:
		This is called when instances are created in GDScript. The code should return
		an object id to reference an instance in Go memory since we cannot return go pointers
		to C to use long term
	*/
	var t T
	tn := (TypeName)(reflect.TypeOf(t).Name())

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

func GoCallback_GDNativeBindingFree[C GDNativeClass](p_token unsafe.Pointer, p_instance unsafe.Pointer, p_binding unsafe.Pointer) {
	objId := ObjectID{Id: uint64(uintptr(p_binding))}

	if _, ok := internal.gdNativeInstances.Get(objId); !ok {
		log.Panic("GDNativeClass instance not found to free", zap.Uint64("id", objId.Id))
	}

	// inst := (*GDNativeClass)(p_binding)

	// inst.GetGodotObjectOwner()

	internal.gdNativeInstances.Delete(objId)
}

func GoCallback_GDNativeBindingReference[C GDNativeClass](p_token unsafe.Pointer, p_instance unsafe.Pointer, p_reference GDNativeBool) GDNativeBool {
	return 1
}
