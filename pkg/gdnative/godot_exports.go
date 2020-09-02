package gdnative

/*
#include <cgo_gateway_register_class.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	"github.com/godot-go/godot-go/pkg/log"
	"strconv"
	"unsafe"
)

//GodotGoVersion holds the relese version
var GodotGoVersion = "0.1"

// This NOOP init() seems to get this file evaluated first
// so that the #cgo directive get evaluated first
func init() {
}

func apiStructOffset(extensions **C.godot_gdnative_api_struct, i int) *C.godot_gdnative_api_struct {
	ptrPtrExts := (**C.godot_gdnative_api_struct)(unsafe.Pointer(extensions))

	ptrPtr := (**C.godot_gdnative_api_struct)(unsafe.Pointer(uintptr(unsafe.Pointer(ptrPtrExts)) + uintptr(i)*unsafe.Sizeof(*ptrPtrExts)))
	return *ptrPtr
}

func GodotGdnativeInit(options *GdnativeInitOptions) {
	RegisterState.InitCount += 1

	log.WithField("InitCount", fmt.Sprintf("%d", RegisterState.InitCount)).Debug("GodotGdnativeInit called")

	if CoreApi != nil {
		log.Panic("godot gdnative is already initialized!")
	}

	RegisterState.TagDB = tagDB{
		parentTo:        map[TypeTag]TypeTag{},
		typeTags:        map[TypeTag]string{},
		methodTags:      map[MethodTag]classMethod{},
		propertySetTags: map[PropertySetTag]classPropertySet{},
		propertyGetTags: map[PropertyGetTag]classPropertyGet{},
	}

	cOpts := (*C.godot_gdnative_init_options)(unsafe.Pointer(options))

	CoreApi = cOpts.api_struct
	GDNativeLibObject = (*C.godot_object)(unsafe.Pointer(cOpts.gd_native_library))

	coreExtension := cOpts.api_struct.next

	for coreExtension != nil {
		if coreExtension.version.major == 1 && coreExtension.version.minor == 1 {
			Core11Api = (*C.godot_gdnative_core_1_1_api_struct)(unsafe.Pointer(coreExtension))
		} else if coreExtension.version.major == 1 && coreExtension.version.minor == 2 {
			Core12Api = (*C.godot_gdnative_core_1_2_api_struct)(unsafe.Pointer(coreExtension))
		}
		coreExtension = coreExtension.next
	}

	// output library path
	activeLibraryPath := *(*String)(unsafe.Pointer(cOpts.active_library_path))
	charString := activeLibraryPath.Ascii()
	log.Info("active_library_path: ", charString.GetData())

	// output extension versions
	for i := 0; i < int(cOpts.api_struct.num_extensions); i++ {
		ext := apiStructOffset(cOpts.api_struct.extensions, i)

		switch ext._type {
		case C.GDNATIVE_EXT_NATIVESCRIPT: // TODO: codegen enum in go
			NativescriptApi = (*C.godot_gdnative_ext_nativescript_api_struct)(unsafe.Pointer(ext))

			extension := ext.next

			for extension != nil {
				if extension.version.major == 1 && extension.version.minor == 1 {
					Nativescript11Api = (*C.godot_gdnative_ext_nativescript_1_1_api_struct)(unsafe.Pointer(extension))
				}
				extension = extension.next
			}
		case C.GDNATIVE_EXT_PLUGINSCRIPT:
			PluginscriptApi = (*C.godot_gdnative_ext_pluginscript_api_struct)(unsafe.Pointer(ext))
		case C.GDNATIVE_EXT_ANDROID:
			AndroidApi = (*C.godot_gdnative_ext_android_api_struct)(unsafe.Pointer(ext))
		case C.GDNATIVE_EXT_ARVR:
			ARVRApi = (*C.godot_gdnative_ext_arvr_api_struct)(unsafe.Pointer(ext))
		case C.GDNATIVE_EXT_VIDEODECODER:
			VideodecoderApi = (*C.godot_gdnative_ext_videodecoder_api_struct)(unsafe.Pointer(ext))
		case C.GDNATIVE_EXT_NET:
			NetApi = (*C.godot_gdnative_ext_net_api_struct)(unsafe.Pointer(ext))

			extension := ext.next

			for extension != nil {
				if extension.version.major == 3 && extension.version.minor == 2 {
					Net32Api = (*C.godot_gdnative_ext_net_3_2_api_struct)(unsafe.Pointer(extension))
				}
				extension = extension.next
			}
		}
	}

	if NativescriptApi == nil {
		log.Panic("unable to find nativescript extension")
	}

	log.Info(fmt.Sprintf("init %d type tag(s)...", len(registerTypeTagCallbacks)))
	for _, cb := range registerTypeTagCallbacks {
		cb()
	}

	log.Info(fmt.Sprintf("init %d class(es) method binds...", len(registerMethodBindsCallbacks)))
	for _, cb := range registerMethodBindsCallbacks {
		cb()
	}

	log.Debug("GodotGdnativeInit finished")
}

// func GodotGdnativeProfilingAddData(signature *byte, time uint64) {
// 	sig := (*C.char)(unsafe.Pointer(signature))
// 	t := (C.uint64_t)(time)
// 	log.Trace("godot_gdnative_profiling_add_data: %s, %d", C.GoString(sig), uint(time))
// 	C.go_godot_nativescript_profiling_add_data(Nativescript11Api, sig, t)
// }

func GodotGdnativeTerminate(options *GdnativeTerminateOptions) {
	log.Trace("GodotGdnativeTerminate called")

	if Nativescript11Api == nil {
		log.Panic("godot extension nativescript 1.1 api is already nil")
	}

	Nativescript11Api = nil

	if NativescriptApi == nil {
		log.Panic("godot extension nativescript api is already nil")
	}

	NativescriptApi = nil

	if CoreApi == nil {
		log.Panic("godot core api is already nil")
	}

	CoreApi = nil

	log.Trace("GodotGdnativeTerminate finished")
}

func GodotNativescriptInit(handle unsafe.Pointer) {
	log.Trace("GodotNativescriptInit called")

	if len(initNativescriptCallbacks) == 0 {
		log.Warn("no gdnative init callbacks registered gdnative.RegisterInitCallback ")
	}

	if Nativescript11Api == nil {
		log.Panic("godot extension nativescript 1.1 is not initialized!")
	}

	if NativescriptApi == nil {
		log.Panic("godot extension nativescript is not initialized!")
	}

	if CoreApi == nil {
		log.Panic("godot core api is not initialized!")
	}

	RegisterState.NativescriptHandle = handle

	if RegisterState.NativescriptHandle == nil {
		log.Panic("godot nativescript handle is nil!")
	}

	RegisterInstanceBindingFunctions()
	log.WithField("language_index", strconv.Itoa(int(RegisterState.LanguageIndex))).Info("language index assigned")

	log.Info(fmt.Sprintf("init %d callback(s)...", len(initNativescriptCallbacks)))
	for _, cb := range initNativescriptCallbacks {
		cb()
	}

	log.Trace("GodotNativescriptInit finished")
}

func GodotNativescriptTerminate(handle unsafe.Pointer) {
	log.Trace("GodotNativescriptTerminate called")

	wrappedTerminateCallback()

	for _, cb := range terminateCallbacks {
		cb()
	}

	UnregisterInstanceBindingFunctions()

	if RegisterState.NativescriptHandle == nil {
		log.Panic("godot nativescript handle is already nil")
	}

	RegisterState.NativescriptHandle = nil

	log.Trace("GodotNativescriptTerminate finished")
}
