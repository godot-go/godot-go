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
	ptr := &inst
	pnr.Pin(ptr)
	return (unsafe.Pointer)(ptr)
}

//export GoCallback_GDExtensionBindingFree
func GoCallback_GDExtensionBindingFree(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_binding unsafe.Pointer) {
}

//export GoCallback_GDExtensionBindingReference
func GoCallback_GDExtensionBindingReference(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_reference bool) bool {
	return true
}
