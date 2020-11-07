package gdnative

/*
#include <nativescript.wrapper.gen.h>
#include <cgo_gateway_instance_binding.h>
#include <cgo_gateway_register_class.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
)

var (
	sizeofWrapped = int32(unsafe.Sizeof(WrappedImpl{}))
)

// RegisterInstanceBindingFunctions should be called from
func registerInstanceBindingFunctions() {
	ibf := C.godot_instance_binding_functions{}
	ibf.alloc_instance_binding_data = (C.alloc_instance_binding_data)(unsafe.Pointer(C.cgo_gateway_alloc_instance_binding_data))
	ibf.free_instance_binding_data = (C.free_instance_binding_data)(unsafe.Pointer(C.cgo_gateway_free_instance_binding_data))

	idx := C.go_godot_nativescript_register_instance_binding_data_functions(Nativescript11Api, ibf)
	RegisterState.LanguageIndex = idx
}

func unregisterInstanceBindingFunctions() {
	C.go_godot_nativescript_unregister_instance_binding_data_functions(Nativescript11Api, RegisterState.LanguageIndex)
}

//export go_alloc_instance_binding_data
func go_alloc_instance_binding_data(data unsafe.Pointer, typeTag unsafe.Pointer, owner *C.godot_object) unsafe.Pointer {
	w := (*WrappedImpl)(Alloc(sizeofWrapped))

	if w == nil {
		log.Panic("memory allocation for WrappedImpl failed")
	}

	tt := TypeTag(uintptr(typeTag))

	w.Owner = (*GodotObject)(owner)
	w.TypeTag = tt
	w.setUserDataFromTypeTag(tt)

	log.Debug("alloc instance binding WrappedImpl", TypeTagField("tag", w.TypeTag))

	return unsafe.Pointer(w)
}

//export go_free_instance_binding_data
func go_free_instance_binding_data(data unsafe.Pointer, wrapper unsafe.Pointer) {
	if wrapper == nil {
		log.Error("cannot free nil WrappedImpl")
		return
	}

	w := (*WrappedImpl)(wrapper)

	log.Debug("free instance binding WrappedImpl", TypeTagField("tag", w.TypeTag))

	Free(wrapper)
}
