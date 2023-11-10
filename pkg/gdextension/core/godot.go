package core

//revive:disable

// #include <godot/gdextension_interface.h>
// #include <stdio.h>
// #include <stdlib.h>
// #include "gdextension_binding_init.h"
// #include "stacktrace.h"
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextension/builtin"
	. "github.com/godot-go/godot-go/pkg/gdextension/ffi"
	. "github.com/godot-go/godot-go/pkg/globalstate"
	. "github.com/godot-go/godot-go/pkg/gdextension/nativestructure"
	. "github.com/godot-go/godot-go/pkg/util"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

func _GDExtensionBindingInit(
	pGetProcAddress GDExtensionInterfaceGetProcAddress,
	pLibrary GDExtensionClassLibraryPtr,
	rInitialization *GDExtensionInitialization,
) bool {
	// uncomment to print out C stacktraces when logging at debug log level
	// C.enablePrintStacktrace = log.GetLevel() == log.DebugLevel

	Internal.GDNativeInstances = NewSyncMap[ObjectID, GDExtensionClass]()
	Internal.GDClassInstances = NewSyncMap[GDObjectInstanceID, GDClass]()
	Internal.GDRegisteredGDClasses = NewSyncMap[string, *ClassInfo]()

	FFI.LoadProcAddresses(pGetProcAddress, pLibrary)

	// Load the Godot version.
	CallFunc_GDExtensionInterfaceGetGodotVersion(&FFI.GodotVersion)

	log.Info("godot version",
		zap.Int32("major", FFI.GodotVersion.GetMajor()),
		zap.Int32("minor", FFI.GodotVersion.GetMinor()),
	)

	rInitialization.SetCallbacks(
		(*[0]byte)(C.cgo_callfn_GDExtensionBindingInitializeLevel),
		(*[0]byte)(C.cgo_callfn_GDExtensionBindingDeinitializeLevel),
	)

	var hasInit bool

	for i := GDExtensionInitializationLevel(0); i < GDEXTENSION_MAX_INITIALIZATION_LEVEL; i++ {
		if GDExtensionBindingInitCallbacks[i] != nil {
			rInitialization.SetInitializationLevel(i)
			hasInit = true
			break
		}
	}

	if !hasInit {
		panic("At least one initialization callback must be defined.")
	}

	VariantInitBindings()
	RegisterEngineClasses()
	RegisterEngineClassRefs()

	return true
}

//export GDExtensionBindingInitializeLevel
func GDExtensionBindingInitializeLevel(userdata unsafe.Pointer, pLevel C.GDExtensionInitializationLevel) {
	classdbCurrentLevel = (GDExtensionInitializationLevel)(pLevel)

	if fn := GDExtensionBindingInitCallbacks[pLevel]; fn != nil {
		log.Debug("GDExtensionBindingInitializeLevel init", zap.Int32("level", (int32)(pLevel)))
		fn()
	}

	classDBInitialize(classdbCurrentLevel)
}

//export GDExtensionBindingDeinitializeLevel
func GDExtensionBindingDeinitializeLevel(userdata unsafe.Pointer, pLevel C.GDExtensionInitializationLevel) {
	classdbCurrentLevel = (GDExtensionInitializationLevel)(pLevel)
	classDBDeinitialize(classdbCurrentLevel)

	if GDExtensionBindingTerminateCallbacks[pLevel] != nil {
		GDExtensionBindingTerminateCallbacks[pLevel]()
	}
}

// func GDExtensionBindingCreateInstanceCallback(pToken unsafe.Pointer, pInstance unsafe.Pointer) Wrapped {
// 	if pToken != unsafe.Pointer(FFI.Library) {
// 		panic("Asking for creating instance with invalid token.")
// 	}

// 	owner := (*GodotObject)(pInstance)

// 	id := CallFunc_GDExtensionInterfaceObjectGetInstanceId((GDExtensionConstObjectPtr)(owner))

// 	log.Debug("GDExtensionBindingCreateInstanceCallback called", zap.Any("id", id))

// 	obj := NewGDExtensionClassFromObjectOwner(owner).(Object)

// 	strClass := obj.GetClass()

// 	cn := strClass.ToAscii()

// 	w := obj.CastTo(cn)
// 	return w
// }

// func GDExtensionBindingFreeInstanceCallback(pToken unsafe.Pointer, pInstance unsafe.Pointer, pBinding unsafe.Pointer) {
// 	if pToken != unsafe.Pointer(FFI.Library) {
// 		panic("Asking for freeing instance with invalid token.")
// 	}

// 	w := (*WrappedImpl)(pBinding)

// 	CallFunc_GDExtensionInterfaceObjectDestroy((GDExtensionObjectPtr)(w.Owner))
// }

type InitObject struct {
	getProcAddress GDExtensionInterfaceGetProcAddress
	library        GDExtensionClassLibraryPtr
	initialization *GDExtensionInitialization
}

func (o InitObject) RegisterCoreInitializer(pCoreInit GDExtensionBindingCallback) {
	GDExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_CORE] = pCoreInit
}

func (o InitObject) RegisterServerInitializer(pServerInit GDExtensionBindingCallback) {
	GDExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_SERVERS] = pServerInit
}

func (o InitObject) RegisterSceneInitializer(pSceneInit GDExtensionBindingCallback) {
	GDExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_SCENE] = pSceneInit
}

func (o InitObject) RegisterEditorInitializer(pEditorInit GDExtensionBindingCallback) {
	GDExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_EDITOR] = pEditorInit
}

func (o InitObject) RegisterCoreTerminator(pCoreTerminate GDExtensionBindingCallback) {
	GDExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_CORE] = pCoreTerminate
}

func (o InitObject) RegisterServerTerminator(pServerTerminate GDExtensionBindingCallback) {
	GDExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_SERVERS] = pServerTerminate
}

func (o InitObject) RegisterSceneTerminator(pSceneTerminate GDExtensionBindingCallback) {
	GDExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_SCENE] = pSceneTerminate
}

func (o InitObject) RegisterEditorTerminator(pEditorTerminate GDExtensionBindingCallback) {
	GDExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_EDITOR] = pEditorTerminate
}

func (o InitObject) Init() bool {
	return _GDExtensionBindingInit(o.getProcAddress, o.library, o.initialization)
}

func NewInitObject(
	getProcAddress GDExtensionInterfaceGetProcAddress,
	library GDExtensionClassLibraryPtr,
	initialization *GDExtensionInitialization,
) *InitObject {
	return &InitObject{
		getProcAddress: getProcAddress,
		library:        library,
		initialization: initialization,
	}
}
