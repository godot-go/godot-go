package pkg

// In-engine pins for the class callback prologues' teardown tolerance
// (openspec trim-dispatch-overhead task 4.1), written before the resolution
// basis changes so the behavior survives the move from the engine-reported
// class name to the instance-held identity (design D1/D2).
//
// What is pinned:
//
//  1. A notification for an instance whose reported class has no GDClass
//     registration — the presentation a destructing instance makes during
//     shutdown, when the extension class has left the class DB and the
//     engine answers get_class with the parent name — is absorbed: the call
//     returns with no panic and no log entry, even with the threshold
//     raised to the level that writes dispatch diagnostics.
//
//  2. The same-shaped revert queries (property_can_revert,
//     property_get_revert) answer negatively and silently.
//
//  3. The probes are not vacuously inert: the identical deliveries to the
//     registered Example class run the Go virtuals — the notification is
//     recorded and the revert answers arrive — so cells 1 and 2 measure a
//     decision, not a dead probe.
//
// The unregistered presentation is a plain engine Label: the prologues ask
// GDRegisteredGDClasses about the class name the instance reports, and
// "Label" is no more registered at runtime than the parent name is at
// shutdown. Deliveries go through the pkg/core probes
// (DeliverNotificationForInstance and friends), which run the real
// callbacks on a transient handle.

import (
	"bytes"
	"fmt"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	"github.com/godot-go/godot-go/pkg/core"
	"github.com/godot-go/godot-go/pkg/log"
)

// TestTeardownTolerance pins silent absorption and negative revert answers
// for an unregistered-class presentation, then re-checks the same deliveries
// against the registered Example class. instance is the unregistered
// subject (an engine Label); the fixture only reads its class name.
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
		fmt.Printf("  teardown_tolerance: live path still dispatches; absorption and decline are the unregistered branch alone\n")

		return 1
	})
}
