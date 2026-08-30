package pkg

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// TestDelegationRepro is the recursion-repro class for the delegation trap
// (task 1.2): its _get_maximum_size override delegates through the plain
// GetMaximumSize wrapper, which re-enters the registered virtual today.
// delegationDepth caps each dispatch so the trap is proven deterministically
// instead of exhausting the goroutine stack.
type TestDelegationRepro struct {
	ControlImpl

	delegationDepth int32
}

func (t *TestDelegationRepro) GetClassName() string {
	return "TestDelegationRepro"
}

func (t *TestDelegationRepro) GetParentClassName() string {
	return "Control"
}

const delegationDepthCap = 25

func (t *TestDelegationRepro) V_TestDelegationRepro_GetMaximumSize() Vector2 {
	t.delegationDepth++
	defer func() { t.delegationDepth-- }()
	if t.delegationDepth > delegationDepthCap {
		log.Panic("delegation recursion confirmed",
			zap.Int32("depth", t.delegationDepth),
		)
	}
	return t.GetMaximumSize()
}

func NewTestDelegationReproFromOwnerObject(owner *GodotObject) GDClass {
	obj := &TestDelegationRepro{}
	obj.SetGodotObjectOwner(owner)
	return obj
}

func RegisterClassTestDelegationRepro() {
	ClassDBRegisterClass(NewTestDelegationReproFromOwnerObject, nil, nil, func(t *TestDelegationRepro) {
		ClassDBBindMethodVirtual(t, "V_TestDelegationRepro_GetMaximumSize", "_get_maximum_size", nil, nil)
	})
	log.Debug("TestDelegationRepro registered")
}
