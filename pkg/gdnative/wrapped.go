package gdnative

/*
#include <cgo_gateway_class.h>
#include <nativescript.gen.h>
#include <gdnative.gen.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"github.com/pcting/godot-go/pkg/log"
	"unsafe"
)

var (
	wrappedNativeScript                       *C.godot_object
	wrappedNativeScriptSetClassNameMethodBind *C.godot_method_bind
	wrappedNativeScriptSetLibraryMethodBind   *C.godot_method_bind
	wrappedNativeScriptSetScriptMethodBind    *C.godot_method_bind
	nilptr                                    = unsafe.Pointer(uintptr(0))
)

func init() {
	RegisterInitCallback(wrappedInitCallback)
}

type Wrapped struct {
	Owner   *GodotObject
	TypeTag TypeTag
	Name    string
}

func (w Wrapped) GetOwner() *GodotObject {
	return w.Owner
}

func (w Wrapped) GetTypeTag() TypeTag {
	return w.TypeTag
}

func (w Wrapped) GetWrappedName() string {
	return w.Name
}

func wrappedInitCallback() {
	// Ported from godot-cpp: https://github.com/godotengine/godot-cpp/blob/master/include/core/Godot.hpp#L39
	// these are static members in create_custom_class_instance()

	strNativeScript := C.CString("NativeScript")
	defer C.free(unsafe.Pointer(strNativeScript))
	wrappedNativeScript = (*C.godot_object)(C.go_godot_get_class_constructor_new(CoreApi, strNativeScript))

	strSetClassName := C.CString("set_class_name")
	defer C.free(unsafe.Pointer(strSetClassName))
	wrappedNativeScriptSetClassNameMethodBind = (*C.godot_method_bind)(unsafe.Pointer(C.go_godot_method_bind_get_method(CoreApi, strNativeScript, strSetClassName)))

	if wrappedNativeScriptSetClassNameMethodBind == nil {
		log.Debug(errors.New("failed to initialize method bind wrappedNativeScriptSetClassNameMethodBind"))
	}

	strSetLibrary := C.CString("set_library")
	defer C.free(unsafe.Pointer(strSetLibrary))
	wrappedNativeScriptSetLibraryMethodBind = (*C.godot_method_bind)(unsafe.Pointer(C.go_godot_method_bind_get_method(CoreApi, strNativeScript, strSetLibrary)))

	if wrappedNativeScriptSetLibraryMethodBind == nil {
		log.Debug(errors.New("failed to initialize method bind wrappedNativeScriptSetLibraryMethodBind"))
	}

	strObject := C.CString("Object")
	defer C.free(unsafe.Pointer(strObject))
	strSetScript := C.CString("set_script")
	defer C.free(unsafe.Pointer(strSetScript))
	wrappedNativeScriptSetScriptMethodBind = (*C.godot_method_bind)(unsafe.Pointer(C.go_godot_method_bind_get_method(CoreApi, strObject, strSetScript)))
}

func wrappedTerminateCallback() {
	Free(unsafe.Pointer(wrappedNativeScript))
}

func GetWrapper(owner *GodotObject) Wrapped {
	return *(*Wrapped)(unsafe.Pointer(C.go_godot_nativescript_get_instance_binding_data(Nativescript11Api, RegisterState.LanguageIndex, unsafe.Pointer(owner))))
}

func GetCustomClassInstance(owner *GodotObject) Class {
	if owner == nil {
		log.Panic("cannot cast null owner as NativeScriptClass")
	}

	ud := *(*UserData)(C.go_godot_nativescript_get_userdata(NativescriptApi, unsafe.Pointer(owner)))

	if inst, ok := classInstances[ud]; ok {
		return inst
	}

	return nil
}

func CreateCustomClassInstance(className string, baseClassName string) Class {
	// Ported from godot-cpp: https://github.com/godotengine/godot-cpp/blob/master/include/core/Godot.hpp#L39

	setLibraryArgs := [...]unsafe.Pointer{unsafe.Pointer(GDNativeLibObject)}

	C.go_godot_method_bind_ptrcall(
		CoreApi,
		wrappedNativeScriptSetLibraryMethodBind,
		unsafe.Pointer(wrappedNativeScript),
		(*unsafe.Pointer)(unsafe.Pointer(&setLibraryArgs[0])),
		nilptr,
	)

	strClassName := C.CString(className)
	defer C.free(unsafe.Pointer(strClassName))

	setClassNameArgs := [...]unsafe.Pointer{unsafe.Pointer(&strClassName)}

	C.go_godot_method_bind_ptrcall(
		CoreApi,
		wrappedNativeScriptSetClassNameMethodBind,
		unsafe.Pointer(wrappedNativeScript),
		(*unsafe.Pointer)(unsafe.Pointer(&setClassNameArgs[0])),
		nilptr,
	)

	strBaseClassName := C.CString(baseClassName)
	defer C.free(unsafe.Pointer(strBaseClassName))

	baseObject := (*C.godot_object)(C.go_godot_get_class_constructor_new(CoreApi, strBaseClassName))

	setScriptArgs := [...]unsafe.Pointer{unsafe.Pointer(wrappedNativeScript)}

	C.go_godot_method_bind_ptrcall(
		CoreApi,
		wrappedNativeScriptSetScriptMethodBind,
		unsafe.Pointer(baseObject),
		(*unsafe.Pointer)(unsafe.Pointer(&setScriptArgs[0])),
		nilptr,
	)

	ud := *(*UserData)(C.go_godot_nativescript_get_userdata(
		NativescriptApi,
		unsafe.Pointer(baseObject),
	))

	inst, ok := classInstances[ud]

	if !ok {
		log.Panic(fmt.Sprintf("unable to find class instance %s", ud))
	}

	return inst
}

// func NewClass(class string) Class {
// 	c, ok := ConstructorMap[class]

// 	if !ok {
// 		log.WithField("class", class).Panic("unable to find constructor")
// 	}

// 	cClass := C.CString(class)
// 	defer C.free(unsafe.Pointer(cClass))

// 	owner := (*GodotObject)(C.go_godot_get_class_constructor_new(CoreApi, cClass))
// 	wrapped := *(*Wrapped)(C.go_godot_nativescript_get_instance_binding_data(Nativescript11Api, RegisterState.LanguageIndex, unsafe.Pointer(owner)))

// 	return c(wrapped.Owner, wrapped.TypeTag)
// }
