package pkg

import (
	"fmt"
	"strings"

	. "github.com/godot-go/godot-go/pkg/core"
	"github.com/godot-go/godot-go/pkg/gdclassimpl"
)

// Witness for the class-metadata-registration requirement "Metadata requests
// against an unregistered class fail loudly", group flavour. The group path
// panics on a missing registration before any engine call is made, so the
// fault is caught here and the run survives.

const (
	propertyGroupUnknownOK          int64 = 1
	propertyGroupUnknownNoFault     int64 = -1 // the request was not rejected
	propertyGroupUnknownTextMissing int64 = -2 // the fault does not name the class
)

// ghostClass carries a class name that was never registered. It exists only
// to drive the missing-registration path.
type ghostClass struct {
	gdclassimpl.ControlImpl
}

func (c *ghostClass) GetClassName() string { return "NeverRegisteredGhost" }

// TestPropertyGroupUnknownClass registers a property group for the ghost
// class and asserts the request fails loudly naming the class, before the
// engine could be called (the panic sits ahead of the ffi call in
// ClassDBAddPropertyGroup, so a fault proves the engine was not reached).
func (e *Example) TestPropertyGroupUnknownClass() int64 {
	var caught any
	func() {
		defer func() { caught = recover() }()
		ClassDBAddPropertyGroup(&ghostClass{}, "Ghost group", "ghost_")
	}()
	if caught == nil {
		return propertyGroupUnknownNoFault
	}
	message := fmt.Sprint(caught)
	if !strings.Contains(message, "NeverRegisteredGhost") || !strings.Contains(message, "property group") {
		return propertyGroupUnknownTextMissing
	}
	return propertyGroupUnknownOK
}
