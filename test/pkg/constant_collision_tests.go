package pkg

import (
	"fmt"
	"strings"

	. "github.com/godot-go/godot-go/pkg/core"
)

// Witness for the class-metadata-registration requirement "A constant name is
// unique within its class": binding a name a second time under a different
// enum must fail fatally before the engine is asked. As with the signal
// collision probe, the fault is caught here so the run survives.

const (
	constantCollisionOK           int64 = 1
	constantCollisionSlipped      int64 = -1 // re-binding the taken name did not fault
	constantCollisionClassMissing int64 = -2 // caught diagnostic does not name the class
	constantCollisionNameMissing  int64 = -3 // caught diagnostic does not name the constant
)

// TestConstantCollisionRejection binds FIRST — already owned by ExampleEnum —
// a second time under the SecondBitfield enum and asserts the collision fails
// fatally naming the class and the constant. The first binding keeps its
// value, which GDScript asserts after this returns.
func (e *Example) TestConstantCollisionRejection() int64 {
	var caught any
	func() {
		defer func() { caught = recover() }()
		ClassDBBindEnumConstant(e, "Example.SecondBitfield", "FIRST", 999)
	}()
	if caught == nil {
		return constantCollisionSlipped
	}
	message := fmt.Sprint(caught)
	if !strings.Contains(message, "Example") {
		return constantCollisionClassMissing
	}
	if !strings.Contains(message, "FIRST") {
		return constantCollisionNameMissing
	}
	return constantCollisionOK
}
