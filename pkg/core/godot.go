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

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/gdclassinit"
	"github.com/godot-go/godot-go/pkg/log"
	. "github.com/godot-go/godot-go/pkg/util"
	"go.uber.org/zap"
)

func _GDExtensionBindingInit(
	pGetProcAddress GDExtensionInterfaceGetProcAddress,
	pLibrary GDExtensionClassLibraryPtr,
	rInitialization *GDExtensionInitialization,
) bool {
	// uncomment to print out C stacktraces when logging at debug log level
	// C.enablePrintStacktrace = log.GetLevel() == log.DebugLevel

	Internal.GDClassInstances = NewSyncMap[GDObjectInstanceID, GDClass]()
	Internal.GDRegisteredGDClasses = NewSyncMap[string, *ClassInfo]()
	Internal.GDClassConstructors = NewSyncMap[string, GDClassGoConstructorFromOwner]()

	FFI.LoadProcAddresses(pGetProcAddress, pLibrary)

	// Load the Godot version.
	CallFunc_GDExtensionInterfaceGetGodotVersion(FFI.GodotVersion)

	log.Info("godot version",
		zap.Int32("major", FFI.GodotVersion.GetMajor()),
		zap.Int32("minor", FFI.GodotVersion.GetMinor()),
	)

	rInitialization.SetUserData(nil)

	rInitialization.SetCallbacks(
		(*[0]byte)(C.cgo_callfn_GDExtensionBindingInitializeLevel),
		(*[0]byte)(C.cgo_callfn_GDExtensionBindingDeinitializeLevel),
	)

	var hasInit bool

	for i := range GDEXTENSION_MAX_INITIALIZATION_LEVEL {
		level := GDExtensionInitializationLevel(i)
		if GDExtensionBindingInitCallbacks[level] != nil {
			rInitialization.SetInitializationLevel(level)
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

	// Mirror godot-cpp (godot.cpp initialize_level): hand the main loop
	// callbacks to the engine once, at the core level, after the level's
	// init callback has run, and only when at least one hook is assigned.
	if classdbCurrentLevel == GDEXTENSION_INITIALIZATION_CORE &&
		(GDExtensionBindingMainLoopStartupCallback != nil ||
			GDExtensionBindingMainLoopFrameCallback != nil ||
			GDExtensionBindingMainLoopShutdownCallback != nil) {
		registerMainLoopCallbacks()
	}
}

//export GDExtensionBindingDeinitializeLevel
func GDExtensionBindingDeinitializeLevel(userdata unsafe.Pointer, pLevel C.GDExtensionInitializationLevel) {
	classdbCurrentLevel = (GDExtensionInitializationLevel)(pLevel)
	deinitializeBindingLevel(classdbCurrentLevel)
}

// deinitializeBindingLevel runs the level's user terminator, then drains the
// editor plugin registry when the EDITOR level deinitializes, mirroring
// godot-cpp GDExtensionBinding::deinitialize_level (user terminate callback,
// then EditorPlugins::deinitialize). Extracted as a pure-Go seam so the
// terminator-before-drain ordering is unit-testable without an engine.
func deinitializeBindingLevel(level GDExtensionInitializationLevel) {
	if GDExtensionBindingTerminateCallbacks[level] != nil {
		GDExtensionBindingTerminateCallbacks[level]()
	}
	if level == GDEXTENSION_INITIALIZATION_EDITOR {
		editorPluginRegistryInstance.drain()
	}
}

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

// RegisterStartupCallback sets the hook invoked once after Godot is started
// and fully initialized, before scene processing begins. Assigning replaces
// any previously assigned startup hook. Hooks must be registered before
// Init() is called; registration with the engine happens at the core
// initialization level.
func (o InitObject) RegisterStartupCallback(pStartup GDExtensionBindingCallback) {
	GDExtensionBindingMainLoopStartupCallback = pStartup
}

// RegisterFrameCallback sets the hook invoked for every process frame of the
// main loop, after all `_process()` methods on Node and before
// ScriptServer::frame(). Assigning replaces any previously assigned frame
// hook. Hooks must be registered before Init() is called.
func (o InitObject) RegisterFrameCallback(pFrame GDExtensionBindingCallback) {
	GDExtensionBindingMainLoopFrameCallback = pFrame
}

// RegisterShutdownCallback sets the hook invoked when shutting down the main
// loop, before Godot is shut down while it is still fully initialized.
// Assigning replaces any previously assigned shutdown hook. Hooks must be
// registered before Init() is called.
func (o InitObject) RegisterShutdownCallback(pShutdown GDExtensionBindingCallback) {
	GDExtensionBindingMainLoopShutdownCallback = pShutdown
}

func (o InitObject) Init() bool {
	return _GDExtensionBindingInit(o.getProcAddress, o.library, o.initialization)
}

func NewInitObject(
	getProcAddress GDExtensionInterfaceGetProcAddress,
	library GDExtensionClassLibraryPtr,
	initialization *GDExtensionInitialization,
) *InitObject {
	ret := &InitObject{
		getProcAddress: getProcAddress,
		library:        library,
		initialization: initialization,
	}
	return ret
}
