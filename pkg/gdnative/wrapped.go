package gdnative

/*
#include <cgo_gateway_class.h>
#include <nativescript.wrapper.gen.h>
#include <gdnative.wrapper.gen.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
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

	UserDataIdentifiableImpl
}

func (w Wrapped) GetOwner() *GodotObject {
	return w.Owner
}

func (w Wrapped) GetTypeTag() TypeTag {
	return w.TypeTag
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

func GetCustomClassInstance(owner *GodotObject) NativeScriptClass {
	if owner == nil {
		log.Panic("cannot cast null owner as NativeScriptClass")
	}

	ud := (UserData)(uintptr(C.go_godot_nativescript_get_userdata(NativescriptApi, unsafe.Pointer(owner))))

	classInst, ok := nativeScriptInstanceMap[ud]

	if !ok {
		log.Panic("unable o find NativeScriptClass instance")
	}

	return classInst
}

func CreateCustomClassInstance(className string, baseClassName string) NativeScriptClass {
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

	ud := (UserData)(uintptr(C.go_godot_nativescript_get_userdata(
		NativescriptApi,
		unsafe.Pointer(baseObject),
	)))

	classInst, ok := nativeScriptInstanceMap[ud]

	if !ok {
		log.Panic("unable to find NativeScriptClass instance")
	}

	return classInst
}
