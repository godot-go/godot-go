package pkg

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
)

// TestVirtualsConformance is the compile-time signature-verification idiom
// (design Decision 2): a convention-correct qualified override checked against
// a per-method anonymous interface. There is deliberately no registration and
// no runtime behavior — the assertion below fails at build time if the
// override's signature drifts from the one the generated catalog
// (ControlVirtuals in pkg/gdclassimpl/virtuals.gen.go) declares for
// _get_minimum_size.
type TestVirtualsConformance struct {
	ControlImpl
}

// V_TestVirtualsConformance_GetMinimumSize mirrors ControlVirtuals'
// V_Control_GetMinimumSize (no arguments, returns Vector2) under this level's
// qualified name.
func (t *TestVirtualsConformance) V_TestVirtualsConformance_GetMinimumSize() Vector2 {
	return NewVector2WithFloat32Float32(1, 2)
}

var _ interface{ V_TestVirtualsConformance_GetMinimumSize() Vector2 } = (*TestVirtualsConformance)(nil)
