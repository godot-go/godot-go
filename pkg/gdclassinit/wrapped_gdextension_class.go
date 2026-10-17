package gdclassinit

// #include "cgo_ptr.h"
import "C"
import (
	"runtime/cgo"
	"sync/atomic"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// Observability counters for the engine-class binding lifecycle (openspec
// fix-instance-handle-lifecycle): the binding-churn gate asserts that every
// created binding handle is released exactly once.
var (
	bindingCreateCount atomic.Uint64
	bindingFreeCount   atomic.Uint64
)

// BindingCreateCount returns how many engine-class binding handles this
// extension has created since start-up.
func BindingCreateCount() uint64 { return bindingCreateCount.Load() }

// BindingFreeCount returns how many engine-class binding releases have run.
func BindingFreeCount() uint64 { return bindingFreeCount.Load() }

//export GoCallback_GDExtensionBindingCreate
func GoCallback_GDExtensionBindingCreate(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer) uintptr {
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
	// The Handle is the lifetime anchor for the wrapper and is released by
	// GoCallback_GDExtensionBindingFree, the engine's free notification for
	// this binding shape (openspec fix-instance-handle-lifecycle).
	handle := cgo.NewHandle(inst)
	bindingCreateCount.Add(1)
	// Returned as an integer; the C thunk casts to the void* the engine
	// stores. Keeping the Go value integer-typed means the packed Handle
	// never occupies a GC pointer slot (openspec
	// fix-instance-handle-lifecycle; see cmd/generate/ffi/handle_slots.go
	// for the same rule on the call side).
	return uintptr(handle)
}

//export GoCallback_GDExtensionBindingFree
func GoCallback_GDExtensionBindingFree(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_binding uintptr) {
	// typeName := C.GoString(p_type_name)
	// log.Debug("GoCallback_GDExtensionBindingFree called",
	// 	zap.String("class", typeName),
	// )
	// The engine runs this exactly once per created binding slot (it removes
	// the slot before notifying, so explicit frees, ~Object, and TOOLS
	// hot-reload clear paths cannot double-notify). p_binding round-trips
	// the Handle create returned as an integer (openspec
	// fix-instance-handle-lifecycle).
	bindingFreeCount.Add(1)
	if p_binding == 0 {
		return
	}
	cgo.Handle(p_binding).Delete()
}

//export GoCallback_GDExtensionBindingReference
func GoCallback_GDExtensionBindingReference(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_reference bool) bool {
	return true
}
