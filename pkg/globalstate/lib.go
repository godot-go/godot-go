package globalstate

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/gdextension
#include <godot/gdextension_interface.h>
*/
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	. "github.com/godot-go/godot-go/pkg/util"
)

var (
	nullptr                                               = unsafe.Pointer(nil)
	GDNativeConstructors                                  = NewSyncMap[string, GDExtensionClassGoConstructorFromOwner]()
	GDClassRefConstructors                                = NewSyncMap[string, RefCountedConstructor]()
	GDExtensionBindingGDExtensionInstanceBindingCallbacks = NewSyncMap[string, GDExtensionInstanceBindingCallbacks]()
	GDRegisteredGDClassEncoders                           = NewSyncMap[string, ArgumentEncoder]()
	GDExtensionBindingInitCallbacks                       [GDEXTENSION_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
	GDExtensionBindingTerminateCallbacks                  [GDEXTENSION_MAX_INITIALIZATION_LEVEL]GDExtensionBindingCallback
)
