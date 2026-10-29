package core

// #include <godot/gdextension_interface.h>
//
// // cgo_probe_instance_ptr casts a cgo.Handle token to the class instance
// // pointer type. The handle is an opaque table index that the prologues
// // convert back with cgo.Handle; it is never dereferenced as memory. The
// // cast lives on the C side so the Go integer-to-pointer rule (and vet's
// // unsafeptr check) is not engaged by the probes below.
// static inline GDExtensionClassInstancePtr cgo_probe_instance_ptr(uintptr_t h) {
// 	return (GDExtensionClassInstancePtr)h;
// }
import "C"
import (
	"runtime/cgo"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

// Instance-callback probes for the runtime fixture suite (openspec
// trim-dispatch-overhead task 4.1).
//
// The teardown tolerance of the class callback prologues — absorbing a
// notification silently and answering revert queries negatively for an
// instance whose reported class has no GDClass registration — is a
// presentation the engine makes during shutdown, after the extension class
// has left the class DB and a destructing instance reports its parent class
// name. A runtime fixture cannot wait for that window, so these probes
// deliver the same callbacks through a transient handle: each prologue runs
// unchanged, reading the wrapper exactly as an engine delivery hands it to
// them. The handle exists only across the single call and is deleted before
// the probe returns; nothing else can observe it, and the prologues register
// no state of their own.
//
// The handle payload is a *WrappedClassInstance holding inst, built directly
// rather than through builtin.WrappedPostInitialize: that helper also stores
// the handle on the live Godot object, and a probe whose handle the engine
// also holds would have two owners of one release. Here the probe is the
// sole owner across the call, which is what makes the defer Delete safe.

// DeliverNotificationForInstance routes a notification through the class
// notification callback for inst, as the engine delivers it to a live
// extension instance.
func DeliverNotificationForInstance(inst Object, what int32, reversed bool) {
	wci := &WrappedClassInstance{Instance: inst}
	h := cgo.NewHandle(wci)
	defer h.Delete()
	var rev C.GDExtensionBool
	if reversed {
		rev = 1
	}
	GoCallback_ClassCreationInfoNotification(
		C.cgo_probe_instance_ptr((C.uintptr_t)(h)),
		C.int32_t(what),
		rev,
	)
}

// QueryPropertyCanRevertForInstance asks the class revert query, through the
// same callback the engine calls, whether the named property can revert for
// inst. A declined query answers false.
func QueryPropertyCanRevertForInstance(inst Object, propertyName string) bool {
	wci := &WrappedClassInstance{Instance: inst}
	h := cgo.NewHandle(wci)
	defer h.Delete()
	sn := NewStringNameWithLatin1Chars(propertyName)
	defer sn.Destroy()
	return GoCallback_ClassCreationInfoPropertyCanRevert(
		C.cgo_probe_instance_ptr((C.uintptr_t)(h)),
		(C.GDExtensionConstStringNamePtr)(unsafe.Pointer(sn.AsGDExtensionConstStringNamePtr())),
	) != 0
}

// QueryPropertyGetRevertForInstance asks the class revert getter, through the
// same callback the engine calls, for the default value of the named
// property. A true answer hands back the written Variant, which the caller
// owns and must destroy.
func QueryPropertyGetRevertForInstance(inst Object, propertyName string) (Variant, bool) {
	wci := &WrappedClassInstance{Instance: inst}
	h := cgo.NewHandle(wci)
	defer h.Delete()
	sn := NewStringNameWithLatin1Chars(propertyName)
	defer sn.Destroy()
	var ret Variant
	found := GoCallback_ClassCreationInfoPropertyGetRevert(
		C.cgo_probe_instance_ptr((C.uintptr_t)(h)),
		(C.GDExtensionConstStringNamePtr)(unsafe.Pointer(sn.AsGDExtensionConstStringNamePtr())),
		(C.GDExtensionVariantPtr)(unsafe.Pointer(&ret)),
	) != 0
	if !found {
		return Variant{}, false
	}
	return ret, true
}
