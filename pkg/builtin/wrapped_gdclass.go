package builtin

// #include <godot/gdextension_interface.h>
// #include "wrapped.h"
import "C"
import (
	"fmt"
	"runtime"
	"runtime/cgo"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type GetPropertyFunc func(string, *Variant)
type SetPropertyFunc func(string, *Variant)

type GDClass interface {
	Wrapped
}

func ObjectClassFromGDExtensionClassInstancePtr(p_instance GDExtensionClassInstancePtr) Object {
	if p_instance == nil {
		return nil
	}
	wci := cgo.Handle(p_instance).Value().(*WrappedClassInstance)
	if wci.Instance == nil {
		log.Panic("unexpected nil instance")
		return nil
	}
	return wci.Instance
}

// SetConstructInfo is equivalent to Wrapped::_set_construct_info in godot-cpp
func SetConstructInfo(w Wrapped, extensionClassName string, cbs ffi.GDExtensionInstanceBindingCallbacks) {
	log.Debug("SetConstructInfo called", zap.String("extensionClassName", extensionClassName))
	owner := w.GetGodotObjectOwner()
	if len(extensionClassName) == 0 {
		log.Panic("extension class name cannot be empty",
			zap.String("w", fmt.Sprintf("%p", w)),
			zap.String("w.GetGodotObjectOwner()", fmt.Sprintf("%p", owner)),
		)
	}
	obj, ok := w.(Object)
	if !ok {
		log.Panic("unable to cast to Object")
	}
	// The handle payload is a *WrappedClassInstance so the shared
	// instance-free callback (GoCallback_ClassCreationInfoFreeInstance)
	// resolves and releases it — the single release owner for every shape
	// that stores this instance (openspec fix-instance-handle-lifecycle).
	inst := &WrappedClassInstance{
		Instance: obj,
	}
	snExtensionClassName := NewStringNameWithLatin1Chars(extensionClassName)
	defer snExtensionClassName.Destroy()
	cnPtr := snExtensionClassName.AsGDExtensionConstStringNamePtr()
	// Call-scoped pins: cnPtr and &cbs cross into C for the duration of the
	// calls below and nowhere longer. The engine copies the callbacks struct
	// out by value (Object::set_instance_binding), so nothing it retains
	// outlives this call. The Handle keeps inst reachable for as long as the
	// engine can call back; releasing the Handle at instance-free time makes
	// the wrapped graph collectable.
	var scratch runtime.Pinner
	defer scratch.Unpin()
	scratch.Pin(unsafe.Pointer(cnPtr))
	scratch.Pin(unsafe.Pointer(&cbs))
	instHandle := cgo.NewHandle(inst)

	// Set the extension instance in the native Godot object.
	CallFunc_GDExtensionInterfaceObjectSetInstance(
		(GDExtensionObjectPtr)(owner),
		(GDExtensionConstStringNamePtr)(cnPtr),
		instHandle,
	)
	CallFunc_GDExtensionInterfaceObjectSetInstanceBinding(
		(GDExtensionObjectPtr)(owner),
		unsafe.Pointer(FFI.Token),
		instHandle,
		&cbs,
	)
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
	callbacks, ok := GDExtensionBindingGDExtensionInstanceBindingCallbacks.Get(extensionClassName)
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
	// No per-instance value is retained on the program-lifetime pinner
	// (openspec fix-instance-handle-lifecycle, issue #152). The Handle below
	// is the sole lifetime anchor: it keeps inst and obj reachable for
	// exactly as long as the engine can call back, and
	// GoCallback_ClassCreationInfoFreeInstance releases it. owner is a C
	// pointer and Pin would be a no-op. cnPtr and &callbacks only cross the
	// two calls below, so call-scoped pins cover them; the engine copies the
	// callbacks struct out by value (Object::set_instance_binding).
	var scratch runtime.Pinner
	defer scratch.Unpin()
	scratch.Pin(unsafe.Pointer(cnPtr))
	scratch.Pin(unsafe.Pointer(&callbacks))
	instHandle := cgo.NewHandle(inst)
	if cnPtr != nil {
		CallFunc_GDExtensionInterfaceObjectSetInstance(
			(GDExtensionObjectPtr)(owner),
			cnPtr,
			instHandle,
		)
	}
	CallFunc_GDExtensionInterfaceObjectSetInstanceBinding(
		(GDExtensionObjectPtr)(owner),
		unsafe.Pointer(FFI.Token),
		instHandle,
		&callbacks,
	)
}

//export GoCallback_GDClassBindingCreate
func GoCallback_GDClassBindingCreate(p_token unsafe.Pointer, p_instance unsafe.Pointer) unsafe.Pointer {
	return nullptr
}

//export GoCallback_GDClassBindingFree
func GoCallback_GDClassBindingFree(p_token unsafe.Pointer, p_instance unsafe.Pointer, p_binding uintptr) {
	// Deliberate no-op. The p_binding value is the same cgo.Handle that
	// WrappedPostInitialize stored as the extension instance, and the
	// instance-free callback (GoCallback_ClassCreationInfoFreeInstance) is
	// its single release owner — it fires strictly earlier, from
	// Object::_predelete, and deletes the Handle there (openspec
	// fix-instance-handle-lifecycle). Deleting here as well would be a
	// double-delete, which panics in runtime/cgo.
}

//export GoCallback_GDClassBindingReference
func GoCallback_GDClassBindingReference(p_token unsafe.Pointer, p_instance unsafe.Pointer, p_reference C.GDExtensionBool) C.GDExtensionBool {
	return 1
}
