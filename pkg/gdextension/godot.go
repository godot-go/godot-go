package gdextension

//revive:disable

// #include <godot/gdextension_interface.h>
// #include <stdio.h>
// #include <stdlib.h>
// #include "gdextension_binding_init.h"
// #include "stacktrace.h"
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextensionffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

type internalImpl struct {
	gdnInterface      *GDExtensionInterface
	library           GDExtensionClassLibraryPtr
	token             unsafe.Pointer
	gdNativeInstances *SyncMap[ObjectID, GDExtensionClass]
	gdClassInstances  *SyncMap[GDObjectInstanceID, GDClass]
}

type GDExtensionBindingCallback func()

type GDExtensionClassGoConstructorFromOwner func(*GodotObject) GDExtensionClass

type GDClassGoConstructor func(data unsafe.Pointer) GDExtensionObjectPtr

var (
	internal internalImpl

	nullptr = unsafe.Pointer(nil)

	gdNativeConstructors                                  = NewSyncMap[string, GDExtensionClassGoConstructorFromOwner]()
	gdExtensionBindingGDExtensionInstanceBindingCallbacks = NewSyncMap[string, GDExtensionInstanceBindingCallbacks]()
	gdRegisteredGDClasses                                 = NewSyncMap[string, *ClassInfo]()
	gdExtensionBindingInitCallbacks                       [GDEXTENSION_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
	gdExtensionBindingTerminateCallbacks                  [GDEXTENSION_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
)

func _GDExtensionBindingInit(
	pInterface *GDExtensionInterface,
	pLibrary GDExtensionClassLibraryPtr,
	rInitialization *GDExtensionInitialization,
) bool {
	C.enablePrintStacktrace = log.GetLevel() == log.DebugLevel

	internal.gdnInterface = pInterface
	internal.library = pLibrary
	internal.token = unsafe.Pointer(&pLibrary)
	internal.gdNativeInstances = NewSyncMap[ObjectID, GDExtensionClass]()
	internal.gdClassInstances = NewSyncMap[GDObjectInstanceID, GDClass]()

	rInitialization.SetCallbacks(
		(*[0]byte)(C.cgo_callfn_GDExtensionBindingInitializeLevel),
		(*[0]byte)(C.cgo_callfn_GDExtensionBindingDeinitializeLevel),
	)

	var hasInit bool

	for i := GDExtensionInitializationLevel(0); i < GDEXTENSION_MAX_INITIALIZATION_LEVEL; i++ {
		if gdExtensionBindingInitCallbacks[i] != nil {
			rInitialization.SetInitializationLevel(i)
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
func GDExtensionBindingInitializeLevel(userdata unsafe.Pointer, pLevel C.GDExtensionInitializationLevel) {
	classdbCurrentLevel = (GDExtensionInitializationLevel)(pLevel)

	if fn := gdExtensionBindingInitCallbacks[pLevel]; fn != nil {
		log.Debug("GDExtensionBindingInitializeLevel init", zap.Int32("level", (int32)(pLevel)))
		fn()
	}

	classDBInitialize(classdbCurrentLevel)
}

//export GDExtensionBindingDeinitializeLevel
func GDExtensionBindingDeinitializeLevel(userdata unsafe.Pointer, pLevel C.GDExtensionInitializationLevel) {
	classdbCurrentLevel = (GDExtensionInitializationLevel)(pLevel)
	// classDBDeinitialize(classdbCurrentLevel)

	if gdExtensionBindingTerminateCallbacks[pLevel] != nil {
		gdExtensionBindingTerminateCallbacks[pLevel]()
	}
}

func GDExtensionBindingCreateInstanceCallback(pToken unsafe.Pointer, pInstance unsafe.Pointer) Wrapped {
	if pToken != unsafe.Pointer(internal.library) {
		panic("Asking for creating instance with invalid token.")
	}

	owner := (*GodotObject)(pInstance)

	id := GDExtensionInterface_object_get_instance_id(internal.gdnInterface, (GDExtensionConstObjectPtr)(owner))

	log.Debug("GDExtensionBindingCreateInstanceCallback called", zap.Any("id", id))

	obj := NewGDExtensionClassFromObjectOwner(owner).(Object)

	strClass := obj.GetClass()

	cn := strClass.ToAscii()

	w := obj.CastTo(cn)
	return w
}

func GDExtensionBindingFreeInstanceCallback(pToken unsafe.Pointer, pInstance unsafe.Pointer, pBinding unsafe.Pointer) {
	if pToken != unsafe.Pointer(internal.library) {
		panic("Asking for freeing instance with invalid token.")
	}

	w := (*WrappedImpl)(pBinding)

	GDExtensionInterface_object_destroy(internal.gdnInterface, (GDExtensionObjectPtr)(w.Owner))
}

type GDExtensionBinding struct {
}

type InitObject struct {
	gdnInterface   *GDExtensionInterface
	library        GDExtensionClassLibraryPtr
	initialization *GDExtensionInitialization
}

func (o InitObject) RegisterCoreInitializer(pCoreInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_CORE] = pCoreInit
}

func (o InitObject) RegisterServerInitializer(pServerInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_SERVERS] = pServerInit
}

func (o InitObject) RegisterSceneInitializer(pSceneInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_SCENE] = pSceneInit
}

func (o InitObject) RegisterEditorInitializer(pEditorInit GDExtensionBindingCallback) {
	gdExtensionBindingInitCallbacks[GDEXTENSION_INITIALIZATION_EDITOR] = pEditorInit
}

func (o InitObject) RegisterCoreTerminator(pCoreTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_CORE] = pCoreTerminate
}

func (o InitObject) RegisterServerTerminator(pServerTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_SERVERS] = pServerTerminate
}

func (o InitObject) RegisterSceneTerminator(pSceneTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_SCENE] = pSceneTerminate
}

func (o InitObject) RegisterEditorTerminator(pEditorTerminate GDExtensionBindingCallback) {
	gdExtensionBindingTerminateCallbacks[GDEXTENSION_INITIALIZATION_EDITOR] = pEditorTerminate
}

func (o InitObject) Init() bool {
	return _GDExtensionBindingInit(o.gdnInterface, o.library, o.initialization)
}

func NewInitObject(
	gdnInterface *GDExtensionInterface,
	library GDExtensionClassLibraryPtr,
	initialization *GDExtensionInitialization,
) *InitObject {
	return &InitObject{
		gdnInterface:   gdnInterface,
		library:        library,
		initialization: initialization,
	}
}
