package gdclassinit

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdclassinit
#include <godot/gdextension_interface.h>
*/
import "C"
import (
	"runtime"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	. "github.com/godot-go/godot-go/pkg/util"
	"go.uber.org/zap"
)

var (
	nullptr                     = unsafe.Pointer(nil)
	GDNativeConstructors        = NewSyncMap[string, GDExtensionClassGoConstructorFromOwner]()
	GDClassRefConstructors      = NewSyncMap[string, RefCountedConstructor]()
	GDRegisteredGDClassEncoders = NewSyncMap[string, ArgumentEncoder]()
	pnr                         runtime.Pinner
)

func init() {
	// Register the cache-miss factory so builtin can create wrappers for objects
	// that weren't created via the binding create callback. This avoids a cycle
	// import (builtin cannot import gdclassinit).
	GetObjectInstanceFactory = createWrapperFromOwner
}

// createWrapperFromOwner creates a Go wrapper for an engine object that wasn't
// routed through the binding create callback. Used as the cache-miss fallback
// in getObjectInstanceBinding (builtin/variant.go).
func createWrapperFromOwner(engineObject *GodotObject) Object {
	if engineObject == nil {
		return nil
	}

	// Get the class name from the engine object.
	snClassName := StringName{}
	snClassNamePtr := snClassName.NativePtr()
	pnr.Pin(snClassNamePtr)
	CallFunc_GDExtensionInterfaceObjectGetClassName(
		(GDExtensionConstObjectPtr)(engineObject),
		FFI.Library,
		(GDExtensionUninitializedStringNamePtr)(snClassNamePtr),
	)
	className := snClassName.ToUtf8()

	// Look up the constructor for this class.
	fn, ok := GDNativeConstructors.Get(className)
	if !ok {
		log.Warn("unable to find constructor for object",
			zap.String("class", className),
		)
		return nil
	}
	inst := fn(engineObject).(Object)
	if inst == nil {
		log.Panic("no instance returned from constructor")
		return nil
	}

	// Create wrapper, pin it, and store in cache.
	wci := &WrappedClassInstance{
		Instance: inst,
		Owner:    engineObject,
	}
	wci.PinSelf()
	SetBindingWrapper(unsafe.Pointer(engineObject), wci)

	log.Debug("createWrapperFromOwner created wrapper",
		zap.String("class", className),
	)
	return inst
}
