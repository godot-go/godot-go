package globalstate

/*
#cgo CFLAGS: -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/log -I${SRCDIR}/../../pkg/globalstate
#include <godot/gdextension_interface.h>
*/
import "C"
import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdextension/builtin"
	. "github.com/godot-go/godot-go/pkg/util"
)

var (
	nullptr                     = unsafe.Pointer(nil)
	GDNativeConstructors        = NewSyncMap[string, GDExtensionClassGoConstructorFromOwner]()
	GDClassRefConstructors      = NewSyncMap[string, RefCountedConstructor]()
	GDRegisteredGDClassEncoders = NewSyncMap[string, ArgumentEncoder]()
)
