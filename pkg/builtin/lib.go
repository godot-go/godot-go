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
	pnr                                                   = runtime.Pinner{}
)

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
