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
	// Create the wrapper and pin it so GC won't move it while Godot holds the binding.
	wci := &WrappedClassInstance{
		Instance: inst,
		Owner:    owner,
	}
	wci.PinSelf()
	// Store in cache keyed by the engine object pointer.
	SetBindingWrapper(p_instance, wci)
	// Return the engine object pointer itself — NOT a Go pointer.
	// Godot stores this value and passes it back in the free callback.
	// Returning a Go pointer across the CGO boundary triggers:
	// "panic: runtime error: cgo result is unpinned Go pointer or points to unpinned Go pointer"
	return p_instance
}

//export GoCallback_GDExtensionBindingFree
func GoCallback_GDExtensionBindingFree(p_type_name *C.char, p_token unsafe.Pointer, p_instance unsafe.Pointer, p_binding unsafe.Pointer) {
	if p_binding == nil {
		return
	}
	// p_binding is the engine object pointer returned from GoCallback_GDExtensionBindingCreate.
	// Look up the wrapper from the cache and clean it up.
	wci := GetBindingWrapper(p_binding)
	if wci == nil {
		return
	}
	// Remove from cache before unpinning so concurrent lookups won't find a stale entry.
	DeleteBindingWrapper(p_binding)
	// Unpin all objects pinned by this wrapper's pinner.
	wci.Unpin()
}

//export GoCallback_GDExtensionBindingReference
func GoCallback_GDExtensionBindingReference(p_type_name *C.char, p_token unsafe.Pointer, p_binding unsafe.Pointer, p_reference bool) bool {
	if p_binding == nil {
		return true
	}
	// p_binding is the engine object pointer. Look up the wrapper from the cache.
	wci := GetBindingWrapper(p_binding)
	if wci == nil {
		return true
	}
	obj := wci.Instance
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
