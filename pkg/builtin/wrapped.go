package builtin

import (
	"runtime"
	"sync"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// bindingWrapperCache maps engine object pointers to their Go wrapper instances.
// This avoids returning Go pointers across the CGO boundary in
// GoCallback_GDExtensionBindingCreate, which would trigger:
// "panic: runtime error: cgo result is unpinned Go pointer or points to unpinned Go pointer"
//
// The create callback returns the engine object pointer (p_instance) itself.
// Godot stores that value and passes it back in the free callback.
// We use the engine pointer as a stable key to retrieve the Go wrapper.
var bindingWrapperCache sync.Map // map[unsafe.Pointer]*WrappedClassInstance

// GetBindingWrapper retrieves the Go wrapper for an engine object pointer.
func GetBindingWrapper(engineObjPtr unsafe.Pointer) *WrappedClassInstance {
	if engineObjPtr == nil {
		return nil
	}
	val, ok := bindingWrapperCache.Load(engineObjPtr)
	if !ok {
		return nil
	}
	return val.(*WrappedClassInstance)
}

// SetBindingWrapper registers a Go wrapper for an engine object pointer.
func SetBindingWrapper(engineObjPtr unsafe.Pointer, wci *WrappedClassInstance) {
	bindingWrapperCache.Store(engineObjPtr, wci)
}

// DeleteBindingWrapper removes the Go wrapper for an engine object pointer.
func DeleteBindingWrapper(engineObjPtr unsafe.Pointer) {
	bindingWrapperCache.Delete(engineObjPtr)
}

type WrappedImpl struct {
	// Owner must be public but do not directly modify.
	Owner *GodotObject
}

func (w *WrappedImpl) GetGodotObjectOwner() *GodotObject {
	return w.Owner
}

func (w *WrappedImpl) AsGDExtensionObjectPtr() GDExtensionObjectPtr {
	return (GDExtensionObjectPtr)(unsafe.Pointer(w.Owner))
}

func (w *WrappedImpl) AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr {
	return (GDExtensionConstObjectPtr)(unsafe.Pointer(w.Owner))
}

func (w *WrappedImpl) AsGDExtensionTypePtr() GDExtensionTypePtr {
	return (GDExtensionTypePtr)(unsafe.Pointer(&w.Owner))
}

func (w *WrappedImpl) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return (GDExtensionConstTypePtr)(unsafe.Pointer(&w.Owner))
}

func (w *WrappedImpl) SetGodotObjectOwner(owner *GodotObject) {
	w.Owner = owner
}

func (w *WrappedImpl) IsNil() bool {
	return w == nil || w.Owner == nil
}

func (w *WrappedImpl) Destroy() {
}

func ObjectCastTo(obj Object, className string) Object {
	if obj == nil {
		return nil
	}
	gdStrCn := obj.GetClass()
	defer gdStrCn.Destroy()
	log.Info("ObjectCastTo called",
		zap.String("class", gdStrCn.ToUtf8()),
		zap.String("className", obj.GetClassName()),
		zap.String("otherClassName", className),
	)
	owner := obj.GetGodotObjectOwner()
	cn := NewStringNameWithLatin1Chars(className)
	defer cn.Destroy()
	tag := CallFunc_GDExtensionInterfaceClassdbGetClassTag(
		cn.AsGDExtensionConstStringNamePtr(),
	)
	if tag == nil {
		log.Panic("classTag unexpectedly came back nil", zap.String("type", className))
	}
	casted := CallFunc_GDExtensionInterfaceObjectCastTo(
		(GDExtensionConstObjectPtr)(owner),
		tag,
	)
	if casted == nil {
		return nil
	}
	// Look up the Go wrapper from the cache using the engine object pointer.
	// This avoids calling ObjectGetInstanceBinding which triggers CGO pointer
	// validation panics when the create callback returns a Go pointer.
	wci := GetBindingWrapper(unsafe.Pointer(casted))
	if wci == nil {
		log.Warn("unable to find binding wrapper for casted object",
			zap.String("name", className),
		)
		return nil
	}
	wrapperClassName := wci.Instance.GetClassName()
	gdStrClassName := wci.Instance.GetClass()
	defer gdStrClassName.Destroy()
	log.Info("ObjectCastTo casted",
		zap.String("class", gdStrClassName.ToUtf8()),
		zap.String("className", wrapperClassName),
	)
	return wci.Instance
}

type WrappedClassInstance struct {
	Instance Object
	Owner    *GodotObject
	pinner   runtime.Pinner
}

// Unpin releases all pointers pinned by this instance's pinner.
// Call during binding cleanup to allow the GC to move/finalize objects.
func (wci *WrappedClassInstance) Unpin() {
	if wci != nil {
		wci.pinner.Unpin()
	}
}

// PinSelf pins this instance and its owner so the GC won't move them
// while Godot holds the binding. Must be called exactly once after creation.
func (wci *WrappedClassInstance) PinSelf() {
	if wci != nil {
		wci.pinner.Pin(wci)
		wci.pinner.Pin(wci.Owner)
	}
}
