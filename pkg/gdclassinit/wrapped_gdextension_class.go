package gdclassinit

// #include "cgo_ptr.h"
import "C"
import (
	"runtime/cgo"
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
	// The binding slot is a cgo.Handle value, not the address of a Go
	// interface. Returning &inst here is what made engine-class bindings a
	// different shape from the user-defined ones written by SetConstructInfo
	// and WrappedPostInitialize, and the reader could only match one of them.
	// Every writer now stores a handle, which is also what the generated
	// setter's p_binding cgo.Handle type asserts.
	//
	// The handle is deliberately never deleted: the engine keeps the value for
	// the life of the object, and the previous shape pinned the interface
	// header permanently. See the never-delete-after-retrieval policy.
	handle := cgo.NewHandle(inst)
	// Packed through the C shim rather than unsafe.Pointer(uintptr(...)), which
	// go vet's unsafeptr check cannot know is pointer-stable. See
	// pkg/log/cgo_ptr.h.
	return (unsafe.Pointer)(C.cgo_handle_to_ptr(C.uintptr_t(handle)))
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
func GoCallback_GDExtensionBindingReference(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_reference bool) bool {
	return true
}
