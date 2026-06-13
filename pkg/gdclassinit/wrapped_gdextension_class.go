package gdclassinit

import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

//export GoCallback_GDExtensionBindingCreate
func GoCallback_GDExtensionBindingCreate(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer) unsafe.Pointer {
	typeName := C.GoString(p_type_name)
	log.Debug("GoCallback_GDExtensionBindingCreate called",
		zap.String("class", typeName),
	)
	fn, ok := GDNativeConstructors.Get(typeName)
	if !ok {
		log.Panic("unable to find GDExtension constructor", zap.String("type", typeName))
	}
	owner := (*GodotObject)(p_instance)
	inst := fn(owner).(Object)
	if inst == nil {
		log.Panic("no instance returned")
	}
	// Heap-allocate the Object interface value so the pointer remains stable
	// after this function returns. Godot stores this pointer as the binding.
	ptr := new(Object)
	*ptr = inst
	pnr.Pin(ptr)
	return (unsafe.Pointer)(ptr)
}

//export GoCallback_GDExtensionBindingFree
func GoCallback_GDExtensionBindingFree(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_binding unsafe.Pointer) {
	// typeName := C.GoString(p_type_name)
	// log.Debug("GoCallback_GDExtensionBindingFree called",
	// 	zap.String("class", typeName),
	// )
	// GDNativeConstructors.Delete(typeName)
}

//export GoCallback_GDExtensionBindingReference
func GoCallback_GDExtensionBindingReference(p_type_name *C.char, p_token unsafe.Pointer, p_binding unsafe.Pointer, p_reference bool) bool {
	if p_binding == nil {
		return true
	}
	// p_binding is the heap-allocated *Object returned from GoCallback_GDExtensionBindingCreate.
	ptr := (*Object)(p_binding)
	obj := *ptr
	if obj == nil {
		return true
	}
	if rc, ok := obj.(RefCounted); ok {
		if p_reference {
			rc.Reference()
		} else {
			rc.Unreference()
		}
	}
	return true
}
