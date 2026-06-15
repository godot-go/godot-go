package builtin

// #include <godot/gdextension_interface.h>
// #include "wrapped.h"
import "C"
import (
	"fmt"
	"runtime/cgo"
	"unsafe"

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
		Owner:    owner,
	}
	// Pin through the per-object pinner so we can selectively unpin on free.
	// The global pnr pins FFI.GodotVersion which must never be unpinned.
	inst.PinSelf()
	inst.pinner.Pin(obj)
	// Note: cnPtr is already pinned by AsGDExtensionConstStringNamePtr().
	// It is only needed for the ObjectSetInstance call below, and the StringName
	// is destroyed by defer snExtensionClassName.Destroy(). No need to pin it again.
	instHandle := cgo.NewHandle(inst)
	if cnPtr != nil {
		CallFunc_GDExtensionInterfaceObjectSetInstance(
			(GDExtensionObjectPtr)(owner),
			cnPtr,
			(GDExtensionClassInstancePtr)(instHandle),
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
func GoCallback_GDClassBindingFree(p_token unsafe.Pointer, p_instance unsafe.Pointer, p_binding unsafe.Pointer) {
	// p_binding is the cgo.Handle value for the WrappedClassInstance.
	// Unpin all pinned pointers and destroy the handle.
	if p_binding == nil {
		return
	}
	h := cgo.Handle(p_binding)
	wci := h.Value().(*WrappedClassInstance)

	// Unpin all objects that were pinned for this specific instance.
	// runtime.Pinner.Unpin() takes no args — it unpins everything pinned by this pinner.
	wci.pinner.Unpin()
	h.Delete()
}

//export GoCallback_GDClassBindingReference
func GoCallback_GDClassBindingReference(p_token unsafe.Pointer, p_instance unsafe.Pointer, p_reference C.GDExtensionBool) C.GDExtensionBool {
	wci := cgo.Handle(p_instance).Value().(*WrappedClassInstance)
	if wci == nil || wci.Instance == nil {
		return 1
	}
	if rc, ok := wci.Instance.(RefCounted); ok {
		if p_reference != 0 {
			rc.Reference()
		} else {
			rc.Unreference()
		}
	}
	return 1
}
