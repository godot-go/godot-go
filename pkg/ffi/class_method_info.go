package ffi

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/ffi
#include <godot/gdextension_interface.h>
#include "ffi_wrapper.gen.h"
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
	"github.com/godot-go/godot-go/pkg/util"
)

func NewGDExtensionClassMethodInfo(
	name GDExtensionConstStringNamePtr,
	methodUserdata unsafe.Pointer,
	callFunc GDExtensionClassMethodCall,
	ptrcallFunc GDExtensionClassMethodPtrCall,
	methodFlags uint32,
	hasReturnValue bool,
	returnValueInfo *GDExtensionPropertyInfo,
	returnValueMetadata GDExtensionClassMethodArgumentMetadata,
	argumentCount uint32,
	argumentsInfo *GDExtensionPropertyInfo,
	argumentsMetadata *GDExtensionClassMethodArgumentMetadata,
	defaultArgumentCount uint32,
	defaultArguments *GDExtensionVariantPtr,
) *GDExtensionClassMethodInfo {
	// Allocate in C heap to avoid cgo "Go pointer to unpinned Go pointer" panic.
	// The struct contains C pointers to Go memory (StringName, PropertyInfo, etc.),
	// and the Go 1.26 cgo checker cannot track pinning across function boundaries.
	// By allocating in C heap, the cgo checker doesn't see Go pointers in the struct.
	cptr := C.malloc(C.sizeof_GDExtensionClassMethodInfo)
	if cptr == nil {
		panic("malloc failed: GDExtensionClassMethodInfo")
	}
	ret := (*GDExtensionClassMethodInfo)(cptr)
	(*C.GDExtensionClassMethodInfo)(ret).name = (C.GDExtensionStringNamePtr)(name)
	(*C.GDExtensionClassMethodInfo)(ret).method_userdata = methodUserdata
	(*C.GDExtensionClassMethodInfo)(ret).call_func = (C.GDExtensionClassMethodCall)(callFunc)
	(*C.GDExtensionClassMethodInfo)(ret).ptrcall_func = (C.GDExtensionClassMethodPtrCall)(ptrcallFunc)
	(*C.GDExtensionClassMethodInfo)(ret).method_flags = (C.uint32_t)(methodFlags)
	(*C.GDExtensionClassMethodInfo)(ret).has_return_value = (C.GDExtensionBool)(util.BoolToUint8(hasReturnValue))
	(*C.GDExtensionClassMethodInfo)(ret).return_value_info = (*C.GDExtensionPropertyInfo)(returnValueInfo)
	(*C.GDExtensionClassMethodInfo)(ret).return_value_metadata = (C.GDExtensionClassMethodArgumentMetadata)(returnValueMetadata)
	(*C.GDExtensionClassMethodInfo)(ret).argument_count = (C.uint32_t)(argumentCount)
	(*C.GDExtensionClassMethodInfo)(ret).arguments_info = (*C.GDExtensionPropertyInfo)(argumentsInfo)
	(*C.GDExtensionClassMethodInfo)(ret).arguments_metadata = (*C.GDExtensionClassMethodArgumentMetadata)(argumentsMetadata)
	(*C.GDExtensionClassMethodInfo)(ret).default_argument_count = (C.uint32_t)(defaultArgumentCount)
	(*C.GDExtensionClassMethodInfo)(ret).default_arguments = (*C.GDExtensionVariantPtr)(defaultArguments)
	return ret
}

func (m *GDExtensionClassMethodInfo) Destroy() {
	stringDestructor := (GDExtensionPtrDestructor)(CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING))
	if stringDestructor == nil {
		log.Panic("unable to get String Destructor")
	}
	stringNameDestructor := (GDExtensionPtrDestructor)(CallFunc_GDExtensionInterfaceVariantGetPtrDestructor(GDEXTENSION_VARIANT_TYPE_STRING_NAME))
	if stringNameDestructor == nil {
		log.Panic("unable to get StringName Destructor")
	}
	cm := (*C.GDExtensionClassMethodInfo)(m)
	if cm == nil {
		return
	}

	// Destroy the StringName stored in the struct.
	// Only call once — m.name and cm.name are the same pointer (type alias).
	// Calling twice would double-unref and corrupt Godot's refcounted StringName table.
	if cm.name != nil {
		CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(cm.name))
	}

	// Destroy StringName pointers in return_value_info
	if cm.return_value_info != nil {
		rv := (*GDExtensionPropertyInfo)(cm.return_value_info)
		if rv.name != nil {
			CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(rv.name))
		}
		if rv.class_name != nil {
			CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(rv.class_name))
		}
		if rv.hint_string != nil {
			CallFunc_GDExtensionPtrDestructor(stringDestructor, (GDExtensionTypePtr)(rv.hint_string))
		}
	}

	// Destroy StringName pointers in each argument's PropertyInfo
	argTypesSlice := unsafe.Slice(cm.arguments_info, cm.argument_count)
	for i := range argTypesSlice {
		CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(argTypesSlice[i].name))
		CallFunc_GDExtensionPtrDestructor(stringDestructor, (GDExtensionTypePtr)(argTypesSlice[i].hint_string))
		CallFunc_GDExtensionPtrDestructor(stringNameDestructor, (GDExtensionTypePtr)(argTypesSlice[i].class_name))
	}
	// defaultsSlice := unsafe.Slice(cm.default_arguments, cm.default_argument_count)
	// for i := range defaultsSlice {
	// 	builtin.NewVariant
	// }

	// NOTE: Do NOT call GDExtensionPropertyInfo.Destroy() on return_value_info or
	// arguments_info here — their internal StringName/String pointers are already
	// destroyed above in the argTypesSlice loop (for arguments_info) and in the
	// individual field cleanup. PropertyInfo.Destroy() uses switch/case and would
	// only partially destroy pointers, and the structs themselves are Go-owned
	// (stored in GoMethodMetadata), not C-allocated here.
	// The StringName lifecycle is managed by GoMethodMetadata (see Task 5.1).

	// Free the C heap-allocated struct itself.
	// Allocated via C.malloc() in NewGDExtensionClassMethodInfo.
	// Required for unload+reload safety — without this, the struct leaks on every
	// registration call when the extension is unloaded and reloaded.
	C.free(unsafe.Pointer(cm))
}
