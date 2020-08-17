package gdnative

/*
#include <nativescript.wrappergen.h>
#include <cgo_gateway_instance_binding.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"github.com/pcting/godot-go/pkg/log"
	"unsafe"
)

var (
	sizeofWrapped = int32(unsafe.Sizeof(Wrapped{}))
)

func RegisterInstanceBindingFunctions() {
	ibf := C.godot_instance_binding_functions{}
	ibf.alloc_instance_binding_data = (C.alloc_instance_binding_data)(unsafe.Pointer(C.cgo_gateway_alloc_instance_binding_data))
	ibf.free_instance_binding_data = (C.free_instance_binding_data)(unsafe.Pointer(C.cgo_gateway_free_instance_binding_data))

	idx := C.go_godot_nativescript_register_instance_binding_data_functions(Nativescript11Api, ibf)
	RegisterState.LanguageIndex = idx
}

func UnregisterInstanceBindingFunctions() {
	C.go_godot_nativescript_unregister_instance_binding_data_functions(Nativescript11Api, RegisterState.LanguageIndex)
}

//export go_alloc_instance_binding_data
func go_alloc_instance_binding_data(data unsafe.Pointer, typeTag unsafe.Pointer, instance *C.godot_object) unsafe.Pointer {
	w := (*Wrapped)(Alloc(sizeofWrapped))

	if w == nil {
		log.Panic("memory allocation for Wrapped failed")
	}

	tt := TypeTag(uintptr(typeTag))

	w.Owner = (*GodotObject)(instance)
	w.TypeTag = tt
	w.generateUserData(tt)

	name := RegisterState.TagDB.GetRegisteredClassName(w.TypeTag)

	log.WithField("tag_name", name).Info("alloc instance binding data")

	return unsafe.Pointer(w)
}

//export go_free_instance_binding_data
func go_free_instance_binding_data(data unsafe.Pointer, wrapper unsafe.Pointer) {
	if wrapper != nil {
		Free(wrapper)
	}
}
