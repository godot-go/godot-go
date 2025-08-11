package builtin

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/builtin
#include <godot/gdextension_interface.h>
#include "wrapped.h"
*/
import "C"
import (
	"runtime"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"github.com/godot-go/godot-go/pkg/util"
	. "github.com/godot-go/godot-go/pkg/util"
	"go.uber.org/zap"
)

type GodotObject unsafe.Pointer
type GDExtensionBindingCallback func()
type GDExtensionClassGoConstructorFromOwner func(*GodotObject) GDExtensionClass
type GDClassGoConstructorFromOwner func(*GodotObject) GDClass
type RefCountedConstructor func(reference RefCounted) Ref

// type GDClassGoConstructor func(data unsafe.Pointer) GDExtensionObjectPtr

var (
	variantFromTypeConstructor                            [GDEXTENSION_VARIANT_TYPE_VARIANT_MAX]GDExtensionVariantFromTypeConstructorFunc
	typeFromVariantConstructor                            [GDEXTENSION_VARIANT_TYPE_VARIANT_MAX]GDExtensionTypeFromVariantConstructorFunc
	nullptr                                               = unsafe.Pointer(nil)
	GDExtensionBindingGDExtensionInstanceBindingCallbacks = NewSyncMap[string, GDExtensionInstanceBindingCallbacks]()
	pnr                                                   = new(runtime.Pinner)
)

func GetObjectInstanceBinding(engineObject *GodotObject) Object {
	if engineObject == nil {
		return nil
	}
	// Get existing instance binding, if one already exists.
	instPtr := (*Object)(CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		(GDExtensionObjectPtr)(engineObject),
		FFI.Token,
		nil))
	if instPtr != nil && *instPtr != nil {
		return *instPtr
	}
	snClassName := StringName{}
	snClassNamePtr := snClassName.NativePtr()
	pnr.Pin(snClassNamePtr)
	cok := CallFunc_GDExtensionInterfaceObjectGetClassName(
		(GDExtensionConstObjectPtr)(engineObject),
		FFI.Library,
		(GDExtensionUninitializedStringNamePtr)(snClassNamePtr),
	)
	if cok == 0 {
		log.Panic("failed to get class name",
			zap.Any("owner", engineObject),
		)
	}
	pnr.Pin(snClassNamePtr)
	// defer snClassName.Destroy()
	className := snClassName.ToUtf8()
	// const GDExtensionInstanceBindingCallbacks *binding_callbacks = nullptr;
	// Otherwise, try to look up the correct binding callbacks.
	cbs, ok := GDExtensionBindingGDExtensionInstanceBindingCallbacks.Get(className)
	if !ok {
		log.Warn("unable to find callbacks for Object")
		return nil
	}
	cbsPtr := &cbs
	pnr.Pin(cbsPtr)
	pnr.Pin(engineObject)
	pnr.Pin(FFI.Token)

	util.CgoTestCall(unsafe.Pointer(cbsPtr))
	util.CgoTestCall(unsafe.Pointer(engineObject))
	util.CgoTestCall(FFI.Token)
	instPtr = (*Object)(CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		(GDExtensionObjectPtr)(engineObject),
		FFI.Token,
		cbsPtr))
	runtime.KeepAlive(engineObject)
	runtime.KeepAlive(FFI.Token)
	runtime.KeepAlive(cbsPtr)
	if instPtr == nil || *instPtr == nil {
		log.Panic("unable to get instance")
		return nil
	}
	pnr.Pin(instPtr)
	wrapperClassName := (*instPtr).GetClassName()
	gdStrClassName := (*instPtr).GetClass()
	// defer gdStrClassName.Destroy()
	log.Info("GetObjectInstanceBinding casted",
		zap.String("class", gdStrClassName.ToUtf8()),
		zap.String("className", wrapperClassName),
	)
	return *instPtr
}

func GDClassRegisterInstanceBindingCallbacks(tn string) {
	// substitute for:
	// static constexpr GDExtensionInstanceBindingCallbacks ___binding_callbacks = {
	// 	___binding_create_callback,
	// 	___binding_free_callback,
	// 	___binding_reference_callback,
	// };
	createPtr := (*[0]byte)(C.cgo_gdclass_binding_create_callback)
	freePtr := (*[0]byte)(C.cgo_gdclass_binding_free_callback)
	referencePtr := (*[0]byte)(C.cgo_gdclass_binding_reference_callback)
	pnr.Pin(createPtr)
	pnr.Pin(freePtr)
	pnr.Pin(referencePtr)
	cbs := NewGDExtensionInstanceBindingCallbacks(
		createPtr,
		freePtr,
		referencePtr,
	)
	_, ok := GDExtensionBindingGDExtensionInstanceBindingCallbacks.Get(tn)
	if ok {
		log.Panic("Class with the same name already initialized", zap.String("class", tn))
	}
	GDExtensionBindingGDExtensionInstanceBindingCallbacks.Set(tn, (GDExtensionInstanceBindingCallbacks)(cbs))
}
