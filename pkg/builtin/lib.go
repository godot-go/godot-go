package builtin

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/builtin
#include <godot/gdextension_interface.h>
#include "wrapped.h"
*/
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
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
	cok := CallFunc_GDExtensionInterfaceObjectGetClassName(
		(GDExtensionConstObjectPtr)(engineObject),
		FFI.Library,
		(GDExtensionUninitializedStringNamePtr)(snClassName.NativePtr()),
	)
	if cok == 0 {
		log.Panic("failed to get class name",
			zap.Any("owner", engineObject),
		)
	}
	defer snClassName.Destroy()
	className := snClassName.ToUtf8()
	// const GDExtensionInstanceBindingCallbacks *binding_callbacks = nullptr;
	// Otherwise, try to look up the correct binding callbacks.
	cbs, ok := GDExtensionBindingGDExtensionInstanceBindingCallbacks.Get(className)
	if !ok {
		log.Warn("unable to find callbacks for Object")
		return nil
	}
	instPtr = (*Object)(CallFunc_GDExtensionInterfaceObjectGetInstanceBinding(
		(GDExtensionObjectPtr)(engineObject),
		FFI.Token,
		&cbs))
	if instPtr == nil || *instPtr == nil {
		log.Panic("unable to get instance")
		return nil
	}
	wrapperClassName := (*instPtr).GetClassName()
	gdStrClassName := (*instPtr).GetClass()
	defer gdStrClassName.Destroy()
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
	cbs := NewGDExtensionInstanceBindingCallbacks(
		(*[0]byte)(C.cgo_gdclass_binding_create_callback),
		(*[0]byte)(C.cgo_gdclass_binding_free_callback),
		(*[0]byte)(C.cgo_gdclass_binding_reference_callback),
	)

	_, ok := GDExtensionBindingGDExtensionInstanceBindingCallbacks.Get(tn)

	if ok {
		log.Panic("Class with the same name already initialized", zap.String("class", tn))
	}

	GDExtensionBindingGDExtensionInstanceBindingCallbacks.Set(tn, (GDExtensionInstanceBindingCallbacks)(cbs))
}
