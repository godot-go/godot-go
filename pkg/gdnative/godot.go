package gdnative

//revive:disable

// #include <godot/gdnative_interface.h>
// #include "gdnative_wrapper.gen.h"
// #include "gdnative_binding_wrapper.h"
// #include <stdio.h>
// #include <stdlib.h>
import "C"
import (
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type internalImpl struct {
	gdnInterface      *GDNativeInterface
	library           GDNativeExtensionClassLibraryPtr
	token             unsafe.Pointer
	gdNativeInstances *SyncMap[ObjectID, GDNativeClass]
	gdClassInstances  *SyncMap[GDObjectInstanceID, GDClass]
}

type GDExtensionBindingCallback func()

type GDNativeClassGoConstructorFromOwner func(*GodotObject) GDNativeClass

type GDClassGoConstructor func(data unsafe.Pointer) GDNativeObjectPtr

var (
	internal internalImpl

	nullptr = unsafe.Pointer(nil)

	gdNativeConstructors                               = NewSyncMap[TypeName, GDNativeClassGoConstructorFromOwner]()
	gdExtensionBindingGDNativeInstanceBindingCallbacks = NewSyncMap[TypeName, GDNativeInstanceBindingCallbacks]()
	gdRegisteredGDClasses                              = NewSyncMap[TypeName, *ClassInfo]()
	gdExtensionBindingInitCallbacks                    [GDNATIVE_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
	gdExtensionBindingTerminateCallbacks               [GDNATIVE_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
)

func _GDExtensionBindingInit(
	pInterface *GDNativeInterface,
	pLibrary GDNativeExtensionClassLibraryPtr,
	rInitialization *GDNativeInitialization,
) bool {
	internal.gdnInterface = pInterface
	internal.library = pLibrary
	internal.token = unsafe.Pointer(&pLibrary)
	internal.gdNativeInstances = NewSyncMap[ObjectID, GDNativeClass]()
	internal.gdClassInstances = NewSyncMap[GDObjectInstanceID, GDClass]()

	rInitialization.initialize = (*[0]byte)(C.cgo_wrapper_binding_initialize)
	rInitialization.deinitialize = (*[0]byte)(C.cgo_wrapper_binding_deinitialize)

	var hasInit bool

	for i := GDNativeInitializationLevel(0); i < GDNATIVE_MAX_INITIALIZATION_LEVEL; i++ {
		if gdExtensionBindingInitCallbacks[i] != nil {
			rInitialization.minimum_initialization_level = C.GDNativeInitializationLevel(i)
			hasInit = true
			break
		}
	}

	if !hasInit {
		panic("At least one initialization callback must be defined.")
	}

	variantInitBindings()

	return true
}

//export GDExtensionBindingInitializeLevel
func GDExtensionBindingInitializeLevel(userdata unsafe.Pointer, pLevel C.GDNativeInitializationLevel) {
	classdbCurrentLevel = (GDNativeInitializationLevel)(pLevel)

	if fn := gdExtensionBindingInitCallbacks[pLevel]; fn != nil {
		log.Debug("GDExtensionBindingInitializeLevel init", zap.Int32("level", (int32)(pLevel)))
		fn()
	}

	classDBInitialize(classdbCurrentLevel)
}

//export GDExtensionBindingDeinitializeLevel
func GDExtensionBindingDeinitializeLevel(userdata unsafe.Pointer, pLevel C.GDNativeInitializationLevel) {
	classdbCurrentLevel = (GDNativeInitializationLevel)(pLevel)
	classDBDeinitialize(classdbCurrentLevel)

	if gdExtensionBindingTerminateCallbacks[pLevel] != nil {
		gdExtensionBindingTerminateCallbacks[pLevel]()
	}
}

func GDExtensionBindingCreateInstanceCallback(pToken unsafe.Pointer, pInstance unsafe.Pointer) Wrapped {
	if pToken != unsafe.Pointer(internal.library) {
		panic("Asking for creating instance with invalid token.")
	}

	owner := (*GodotObject)(pInstance)

	id := GDNativeInterface_object_get_instance_id(internal.gdnInterface, (GDNativeObjectPtr)(owner))

	log.Debug("GDExtensionBindingCreateInstanceCallback called", zap.Any("id", id))

	w := NewWrappedFromGodotObject(owner)
	return w
}

func GDExtensionBindingFreeInstanceCallback(pToken unsafe.Pointer, pInstance unsafe.Pointer, pBinding unsafe.Pointer) {
	if pToken != unsafe.Pointer(internal.library) {
		panic("Asking for freeing instance with invalid token.")
	}

	w := (*wrapped)(pBinding)

	C.cgo_GDNativeInterface_object_destroy((*C.GDNativeInterface)(internal.gdnInterface), (C.GDNativeObjectPtr)(w.Owner))
}

type GDExtensionBinding struct {
}

type InitObject struct {
	gdnInterface   *GDNativeInterface
	library        GDNativeExtensionClassLibraryPtr
	initialization *GDNativeInitialization
}

func (o InitObject) RegisterCoreInitializer(pCoreInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDNATIVE_INITIALIZATION_CORE] = pCoreInit
}

func (o InitObject) RegisterServerInitializer(pServerInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDNATIVE_INITIALIZATION_SERVERS] = pServerInit
}

func (o InitObject) RegisterSceneInitializer(pSceneInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDNATIVE_INITIALIZATION_SCENE] = pSceneInit
}

func (o InitObject) RegisterEditorInitializer(pEditorInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDNATIVE_INITIALIZATION_EDITOR] = pEditorInit
}

func (o InitObject) RegisterCoreTerminator(pCoreTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDNATIVE_INITIALIZATION_CORE] = pCoreTerminate
}

func (o InitObject) RegisterServerTerminator(pServerTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDNATIVE_INITIALIZATION_SERVERS] = pServerTerminate
}

func (o InitObject) RegisterSceneTerminator(pSceneTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDNATIVE_INITIALIZATION_SCENE] = pSceneTerminate
}

func (o InitObject) RegisterEditorTerminator(pEditorTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDNATIVE_INITIALIZATION_EDITOR] = pEditorTerminate
}

func (o InitObject) Init() bool {
	return _GDExtensionBindingInit(o.gdnInterface, o.library, o.initialization)
}

func NewInitObject(
	gdnInterface *GDNativeInterface,
	library GDNativeExtensionClassLibraryPtr,
	initialization *GDNativeInitialization,
) *InitObject {
	return &InitObject{
		gdnInterface:   gdnInterface,
		library:        library,
		initialization: initialization,
	}
}
