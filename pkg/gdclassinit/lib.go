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
	. "github.com/godot-go/godot-go/pkg/util"
)

var (
	nullptr                     = unsafe.Pointer(nil)
	GDNativeConstructors        = NewSyncMap[string, GDExtensionClassGoConstructorFromOwner]()
	GDClassRefConstructors      = NewSyncMap[string, RefCountedConstructor]()
	GDRegisteredGDClassEncoders = NewSyncMap[string, ArgumentEncoder]()
	// pnr is for program-lifetime retention only: values the engine keeps a
	// pointer to after the call that registers them returns. Per-call scratch
	// must use a body-scoped runtime.Pinner with defer Unpin() instead.
	pnr runtime.Pinner
)
