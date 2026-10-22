package pkg

import (
	"fmt"
	"strings"

	. "github.com/godot-go/godot-go/pkg/core"
)

// Witnesses for the class-metadata-registration requirement "A signal name
// already carried by the class or an ancestor is rejected". A fatal aborts
// the run, so the collision attempts happen inside recover() at the fixture
// call site (the pattern of AwaitMisuseRecovered) and GDScript asserts the
// verdict code.

const (
	dupSignalOK                     int64 = 1
	dupSignalInheritedSlipped       int64 = -1 // registering an inherited name did not fault
	dupSignalInheritedClassMissing  int64 = -2 // caught diagnostic does not name the class
	dupSignalInheritedSignalMissing int64 = -3 // caught diagnostic does not name the signal
	dupSignalOwnSlipped             int64 = -4 // registering a second copy of an own name did not fault
	dupSignalOwnSignalMissing       int64 = -5 // caught diagnostic does not name the signal
	dupSignalRejectedConsumed       int64 = -6 // a rejected name entered the class's own record
	dupSignalSaysConstant           int64 = -7 // the diagnostic still describes a constant
)

// TestDuplicateSignalRejection attempts two colliding signal registrations —
// a name carried only by inheritance ("ready", from Node) and a second copy
// of a name the class registered at start-up ("custom_signal") — and asserts
// each fails fatally with a diagnostic naming the class and the signal, that
// neither rejection describes a constant, and that the rejected inherited
// name was not retained as one the class owns.
func (e *Example) TestDuplicateSignalRejection() int64 {
	inheritedFault := recoverSignalRegistration(e, "ready")
	if inheritedFault == "" {
		return dupSignalInheritedSlipped
	}
	if !strings.Contains(inheritedFault, "Example") {
		return dupSignalInheritedClassMissing
	}
	if !strings.Contains(inheritedFault, "ready") {
		return dupSignalInheritedSignalMissing
	}
	if strings.Contains(inheritedFault, "onstant") {
		return dupSignalSaysConstant
	}

	ownFault := recoverSignalRegistration(e, "custom_signal")
	if ownFault == "" {
		return dupSignalOwnSlipped
	}
	if !strings.Contains(ownFault, "custom_signal") {
		return dupSignalOwnSignalMissing
	}
	if strings.Contains(ownFault, "onstant") {
		return dupSignalSaysConstant
	}

	ci, ok := Internal.GDRegisteredGDClasses.Get("Example")
	if !ok {
		return dupSignalRejectedConsumed
	}
	if _, owned := ci.SignalNameSet["ready"]; owned {
		return dupSignalRejectedConsumed
	}
	return dupSignalOK
}

// recoverSignalRegistration runs one ClassDBAddSignal and returns the fault
// message it produced, or the empty string when the registration was not
// rejected.
func recoverSignalRegistration(e *Example, signalName string) string {
	var caught any
	func() {
		defer func() { caught = recover() }()
		ClassDBAddSignal(e, signalName)
	}()
	if caught == nil {
		return ""
	}
	return fmt.Sprint(caught)
}
