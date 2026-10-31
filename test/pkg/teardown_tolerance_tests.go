package pkg

// In-engine pins for the class callback prologues' teardown tolerance and
// loud-failure channel (openspec trim-dispatch-overhead tasks 4.1 and
// 4.4/4.5). The 4.1 cells were written before the resolution basis changed
// so the behavior survives the move from the engine-reported class name to
// the instance-held identity (design D1/D2); the decline and loud-fail cells
// pin the new basis itself.
//
// What is pinned:
//
//  1. A notification for an instance whose class has no GDClass
//     registration — the presentation a destructing instance makes during
//     shutdown, when the extension class has left the class DB — is
//     absorbed: no panic and no log entry, even with the threshold raised
//     to the level that writes dispatch diagnostics. The same-shaped revert
//     queries answer negatively and silently.
//
//  2. The probes are not vacuously inert: the identical deliveries to the
//     registered Example class run the Go virtuals — the notification is
//     recorded, the revert answers arrive, get reads and set writes the
//     property-from-list cell — so the declines measure a decision, not a
//     dead probe.
//
//  3. A wrapper marked freeing (the state the instance-free callback sets at
//     its re-entrancy guard point) declines every delivery class before the
//     registration lookup, silently and without running the handlers: the
//     teardown absorption no longer depends on any class name (design D2).
//
//  4. The loud-failure half stays reachable: a get, set, or validation
//     request for an unregistered class fails through the fatal channel with
//     the existing diagnostic and fabricates no return value (property-
//     access-dispatch spec).
//
// The unregistered presentation is a plain engine Label: the prologues ask
// GDRegisteredGDClasses about the name stored on the wrapper, and for a
// Label that is "Label" — no more registered at runtime than the parent
// name is at shutdown. Deliveries go through the pkg/core probes
// (DeliverNotificationForInstance and friends), which run the real
// callbacks on a transient handle and mark the freeing state on the probe's
// private wrapper only.

import (
	"bytes"
	"fmt"
	"strings"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	"github.com/godot-go/godot-go/pkg/core"
	"github.com/godot-go/godot-go/pkg/log"
)

// recoverLoudFail runs fn and reports the panic value it raised through the
// fatal channel, or "" when the call returned normally.
func recoverLoudFail(fn func()) (recovered string) {
	defer func() {
		if r := recover(); r != nil {
			recovered = fmt.Sprint(r)
		}
	}()
	fn()
	return ""
}

// vector3Equals compares the three float32 members of two Vector3 natives
// directly through their memory layout. The engine operator (Equal_Vector3)
// and member getter (MemberGetx) must not be used here: on the fixture's
// freshly constructed Go cells they hit the cgo unpinned-pointer check, while
// the engine-side paths in other fixtures borrow engine-managed cells.
func vector3Equals(a, b Vector3) bool {
	as := unsafe.Slice((*float32)(unsafe.Pointer(&a)), 3)
	bs := unsafe.Slice((*float32)(unsafe.Pointer(&b)), 3)
	return as[0] == bs[0] && as[1] == bs[1] && as[2] == bs[2]
}

// TestTeardownTolerance pins absorption, decline, live dispatch, and the loud
// failure half across the seven class callback prologues. instance is the
// unregistered subject (an engine Label); the fixture only reads its identity.
func (e *Example) TestTeardownTolerance(instance Node) int32 {
	return objectArgGuard("teardown_tolerance", func() int32 {
		if isNullObject(instance) {
			fmt.Println("  teardown_tolerance: FAIL: no unregistered instance supplied")
			return -1
		}

		// Force the permitting threshold for the whole window: tolerance is
		// tested where entries would actually be written, so a silent-absence
		// assertion cannot be an artifact of the default suppression.
		var buf bytes.Buffer
		prevWriter := log.SetWriteSyncer(&buf)
		prevLevel := log.GetLevel()
		log.SetLevel(log.InfoLevel)
		defer func() {
			log.SetLevel(prevLevel)
			log.SetWriteSyncer(prevWriter)
		}()

		// --- Cell 1: pre-delete-shaped notification, absorbed.
		core.DeliverNotificationForInstance(instance, OBJECT_NOTIFICATION_PREDELETE, false)
		if s := buf.String(); s != "" {
			fmt.Printf("  teardown_tolerance: FAIL: absorbed notification logged %q\n", s)
			return -2
		}
		fmt.Printf("  teardown_tolerance: absorbed pre-delete delivery for %T, no log\n", instance)

		// --- Cell 2: revert queries decline, silently.
		if core.QueryPropertyCanRevertForInstance(instance, "property_from_list") {
			fmt.Println("  teardown_tolerance: FAIL: revert query answered can-revert for an unregistered class")
			return -3
		}
		if v, found := core.QueryPropertyGetRevertForInstance(instance, "property_from_list"); found {
			v.Destroy()
			fmt.Println("  teardown_tolerance: FAIL: revert getter produced a value for an unregistered class")
			return -4
		}
		if s := buf.String(); s != "" {
			fmt.Printf("  teardown_tolerance: FAIL: declined revert queries logged %q\n", s)
			return -5
		}
		fmt.Printf("  teardown_tolerance: revert queries declined for the unregistered presentation\n")

		// --- Cell 3: the probes dispatch live for a registered class, so
		// the declines above are decisions and not dead plumbing.
		savedCodes := e.notificationCodes
		defer func() { e.notificationCodes = savedCodes }()
		const probeCode = 10043
		core.DeliverNotificationForInstance(e, probeCode, false)
		recorded := false
		for _, c := range e.notificationCodes {
			if c == probeCode {
				recorded = true
				break
			}
		}
		if !recorded {
			fmt.Println("  teardown_tolerance: FAIL: live notification did not reach the Go virtual")
			return -6
		}
		if !core.QueryPropertyCanRevertForInstance(e, "property_from_list") {
			fmt.Println("  teardown_tolerance: FAIL: live can-revert answered false for property_from_list")
			return -7
		}
		v, found := core.QueryPropertyGetRevertForInstance(e, "property_from_list")
		if !found {
			fmt.Println("  teardown_tolerance: FAIL: live revert getter declined property_from_list")
			return -8
		}
		v.Destroy()

		// Save the cell around the get/set pair so the fixture leaves the
		// property exactly as it found it.
		savedPL := e.propertyFromList
		defer func() { e.propertyFromList = savedPL }()

		vg, found := core.QueryPropertyGetForInstance(e, "property_from_list")
		if !found {
			fmt.Println("  teardown_tolerance: FAIL: live get declined property_from_list")
			return -9
		}
		gotVec := vg.ToVector3()
		if !vector3Equals(gotVec, savedPL) {
			fmt.Println("  teardown_tolerance: FAIL: live get returned the wrong value")
			vg.Destroy()
			return -10
		}
		vg.Destroy()

		setProbe := NewVector3WithFloat32Float32Float32(6.5, 7.5, 8.5)
		vs := NewVariantVector3(setProbe)
		if !core.SetPropertyForInstance(e, "property_from_list", vs) {
			vs.Destroy()
			fmt.Println("  teardown_tolerance: FAIL: live set declined property_from_list")
			return -11
		}
		if !vector3Equals(e.propertyFromList, setProbe) {
			vs.Destroy()
			fmt.Println("  teardown_tolerance: FAIL: live set did not reach the Go virtual")
			return -12
		}
		fmt.Printf("  teardown_tolerance: live path still dispatches get and set\n")

		// --- Cell 4: a freeing-marked wrapper declines every delivery
		// class, silently, before the registration lookup (design D2).
		// Live control: without the mark the same validate request passes
		// and the same notification reaches the virtual.
		const freeingCode = 10044
		core.DeliverNotificationForInstance(e, freeingCode, false)
		freeingRecorded := false
		for _, c := range e.notificationCodes {
			if c == freeingCode {
				freeingRecorded = true
				break
			}
		}
		if !freeingRecorded {
			vs.Destroy()
			fmt.Println("  teardown_tolerance: FAIL: unmarked notification did not reach the Go virtual")
			return -13
		}

		buf.Reset()
		core.DeliverNotificationForInstanceMarkedFreeing(e, 10045, false, true)
		for _, c := range e.notificationCodes {
			if c == 10045 {
				vs.Destroy()
				fmt.Println("  teardown_tolerance: FAIL: freeing-marked notification reached the Go virtual")
				return -14
			}
		}
		if core.QueryPropertyCanRevertForInstanceMarkedFreeing(e, "property_from_list", true) {
			vs.Destroy()
			fmt.Println("  teardown_tolerance: FAIL: freeing-marked can-revert answered true")
			return -15
		}
		if vr, found := core.QueryPropertyGetRevertForInstanceMarkedFreeing(e, "property_from_list", true); found {
			vr.Destroy()
			vs.Destroy()
			fmt.Println("  teardown_tolerance: FAIL: freeing-marked revert getter produced a value")
			return -16
		}
		if vg2, found := core.QueryPropertyGetForInstanceMarkedFreeing(e, "property_from_list", true); found {
			vg2.Destroy()
			vs.Destroy()
			fmt.Println("  teardown_tolerance: FAIL: freeing-marked get produced a value")
			return -17
		}
		if core.SetPropertyForInstanceMarkedFreeing(e, "property_from_list", vs, true) {
			vs.Destroy()
			fmt.Println("  teardown_tolerance: FAIL: freeing-marked set was accepted")
			return -18
		}
		if !vector3Equals(e.propertyFromList, setProbe) {
			vs.Destroy()
			fmt.Println("  teardown_tolerance: FAIL: freeing-marked set mutated the Go field")
			return -19
		}
		// The validate probe is exercised only on the decline path: with no
		// validation hook registered the accepted request reaches the class
		// hook's property-info dereference, which the probe would have to
		// populate exactly as the engine does. The decline fires before that
		// dereference, so a freeing-marked wrapper is safe to deliver to.
		if core.ValidatePropertyForInstance(e, true) {
			vs.Destroy()
			fmt.Println("  teardown_tolerance: FAIL: freeing-marked validate was accepted")
			return -20
		}
		if core.GetPropertyListForInstance(e, true) != 0 {
			vs.Destroy()
			fmt.Println("  teardown_tolerance: FAIL: freeing-marked property list was not empty")
			return -21
		}
		if s := buf.String(); s != "" {
			vs.Destroy()
			fmt.Printf("  teardown_tolerance: FAIL: freeing-marked deliveries logged %q\n", s)
			return -22
		}
		vs.Destroy()
		fmt.Printf("  teardown_tolerance: freeing-marked wrapper declines every delivery class, silently\n")

		// --- Cell 5: the loud half. A get, set, or validation request for
		// an unregistered class fails through the fatal channel with the
		// existing diagnostic and fabricates no value (the recovered calls
		// never wrote their out parameters).
		for _, c := range []struct {
			label string
			call  func()
		}{
			{"get", func() { core.QueryPropertyGetForInstance(instance, "property_from_list") }},
			{"set", func() {
				vv := NewVariantVector3(NewVector3WithFloat32Float32Float32(1, 2, 3))
				defer vv.Destroy()
				core.SetPropertyForInstance(instance, "property_from_list", vv)
			}},
			{"validate", func() { core.ValidatePropertyForInstance(instance, false) }},
			{"property_list", func() { core.GetPropertyListForInstance(instance, false) }},
		} {
			msg := recoverLoudFail(c.call)
			if msg == "" {
				fmt.Printf("  teardown_tolerance: FAIL: unregistered %s did not fail through the fatal channel\n", c.label)
				return -23
			}
			if !strings.Contains(msg, "invalid registered GDClass") {
				fmt.Printf("  teardown_tolerance: FAIL: unregistered %s failed with %q, want the existing diagnostic\n", c.label, msg)
				return -24
			}
		}
		fmt.Printf("  teardown_tolerance: unregistered get, set, validate, and property-list requests still fail loudly\n")

		return 1
	})
}
