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
	"github.com/godot-go/godot-go/pkg/ffi"
)

// Instance-callback probes for the runtime fixture suite (openspec
// trim-dispatch-overhead tasks 4.1 and 4.4/4.5).
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
// The *MarkedFreeing variants additionally set the wrapper's freeing state
// before delivery — the state the instance-free callback sets at its
// re-entrancy guard point — so the decline half of design D2 is observable
// at runtime against a live, registered instance. The mark lives on the
// probe's private wrapper only: the engine's own binding for inst is a
// different wrapper and keeps reporting not-freeing.
//
// The handle payload is a *WrappedClassInstance holding inst, built directly
// rather than through builtin.WrappedPostInitialize: that helper also stores
// the handle on the live Godot object, and a probe whose handle the engine
// also holds would have two owners of one release. Here the probe is the
// sole owner across the call, which is what makes the defer Delete safe.
//
// Each probe stores the Go-side class name on the wrapper exactly as the
// construction sites would, so the prologues answer the same registration
// question an engine delivery does on the instance-held identity basis.

// newProbeInstanceHandle builds the transient wrapper+handle pair shared by
// every probe below; the caller owns the handle and must Delete it.
func newProbeInstanceHandle(inst Object, freeing bool) *WrappedClassInstance {
	wci := &WrappedClassInstance{Instance: inst, ClassName: inst.GetClassName()}
	if freeing {
		wci.MarkFreeing()
	}
	return wci
}

// DeliverNotificationForInstance routes a notification through the class
// notification callback for inst, as the engine delivers it to a live
// extension instance.
func DeliverNotificationForInstance(inst Object, what int32, reversed bool) {
	DeliverNotificationForInstanceMarkedFreeing(inst, what, reversed, false)
}

// DeliverNotificationForInstanceMarkedFreeing delivers the notification with
// the wrapper's freeing state set to freeing. A delivery to a marked
// instance is absorbed by the freeing guard before any registration lookup.
func DeliverNotificationForInstanceMarkedFreeing(inst Object, what int32, reversed bool, freeing bool) {
	wci := newProbeInstanceHandle(inst, freeing)
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
	return QueryPropertyCanRevertForInstanceMarkedFreeing(inst, propertyName, false)
}

// QueryPropertyCanRevertForInstanceMarkedFreeing asks the revert query with
// the wrapper's freeing state set to freeing; a marked instance declines.
func QueryPropertyCanRevertForInstanceMarkedFreeing(inst Object, propertyName string, freeing bool) bool {
	wci := newProbeInstanceHandle(inst, freeing)
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
	return QueryPropertyGetRevertForInstanceMarkedFreeing(inst, propertyName, false)
}

// QueryPropertyGetRevertForInstanceMarkedFreeing asks the revert getter with
// the wrapper's freeing state set to freeing; a marked instance declines.
func QueryPropertyGetRevertForInstanceMarkedFreeing(inst Object, propertyName string, freeing bool) (Variant, bool) {
	wci := newProbeInstanceHandle(inst, freeing)
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

// QueryPropertyGetForInstance routes a property get through the class get
// callback for inst. A handled name answers true and hands back the written
// Variant, which the caller owns and must destroy.
func QueryPropertyGetForInstance(inst Object, propertyName string) (Variant, bool) {
	return QueryPropertyGetForInstanceMarkedFreeing(inst, propertyName, false)
}

// QueryPropertyGetForInstanceMarkedFreeing routes the get with the wrapper's
// freeing state set to freeing; a marked instance declines without running
// the _get virtual.
func QueryPropertyGetForInstanceMarkedFreeing(inst Object, propertyName string, freeing bool) (Variant, bool) {
	wci := newProbeInstanceHandle(inst, freeing)
	h := cgo.NewHandle(wci)
	defer h.Delete()
	sn := NewStringNameWithLatin1Chars(propertyName)
	defer sn.Destroy()
	var ret Variant
	found := GoCallback_ClassCreationInfoGet(
		C.cgo_probe_instance_ptr((C.uintptr_t)(h)),
		(C.GDExtensionConstStringNamePtr)(unsafe.Pointer(sn.AsGDExtensionConstStringNamePtr())),
		(C.GDExtensionVariantPtr)(unsafe.Pointer(&ret)),
	) != 0
	if !found {
		return Variant{}, false
	}
	return ret, true
}

// SetPropertyForInstance routes a property set through the class set callback
// for inst with the supplied value. The value stays owned by the caller.
func SetPropertyForInstance(inst Object, propertyName string, value Variant) bool {
	return SetPropertyForInstanceMarkedFreeing(inst, propertyName, value, false)
}

// SetPropertyForInstanceMarkedFreeing routes the set with the wrapper's
// freeing state set to freeing; a marked instance declines without running
// the _set virtual.
func SetPropertyForInstanceMarkedFreeing(inst Object, propertyName string, value Variant, freeing bool) bool {
	wci := newProbeInstanceHandle(inst, freeing)
	h := cgo.NewHandle(wci)
	defer h.Delete()
	sn := NewStringNameWithLatin1Chars(propertyName)
	defer sn.Destroy()
	return GoCallback_ClassCreationInfoSet(
		C.cgo_probe_instance_ptr((C.uintptr_t)(h)),
		(C.GDExtensionConstStringNamePtr)(unsafe.Pointer(sn.AsGDExtensionConstStringNamePtr())),
		(C.GDExtensionConstVariantPtr)(unsafe.Pointer(&value)),
	) != 0
}

// ValidatePropertyForInstance routes a property validation request through
// the class validate callback for inst. A class with no validation hook
// answers false; a wrapper whose class has no registration fails through the
// fatal channel.
func ValidatePropertyForInstance(inst Object, freeing bool) bool {
	wci := newProbeInstanceHandle(inst, freeing)
	h := cgo.NewHandle(wci)
	defer h.Delete()
	var prop ffi.GDExtensionPropertyInfo
	return GoCallback_ClassCreationInfoValidateProperty(
		C.cgo_probe_instance_ptr((C.uintptr_t)(h)),
		(*C.GDExtensionPropertyInfo)(unsafe.Pointer(&prop)),
	) != 0
}

// GetPropertyListForInstance routes a property-list request through the class
// callback for inst and reports the returned entry count. A wrapper whose
// class has no registration fails through the fatal channel. No fixture cell
// currently receives a non-nil table (the loud-fail cells panic and the
// decline cells answer empty); if one ever does, the paired free_property_list
// callback owns those entries and the probe must route them there.
func GetPropertyListForInstance(inst Object, freeing bool) uint32 {
	wci := newProbeInstanceHandle(inst, freeing)
	h := cgo.NewHandle(wci)
	defer h.Delete()
	var count C.uint32_t
	GoCallback_ClassCreationInfoGetPropertyList(
		C.cgo_probe_instance_ptr((C.uintptr_t)(h)),
		&count,
	)
	return uint32(count)
}
