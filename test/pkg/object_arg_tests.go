package pkg

import (
	"fmt"
	"runtime"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
)

// Tests for the object-argument ptrcall encoding (openspec change
// fix-object-arg-ptrcall-encoding).
//
// Every method here passes an engine-class object *into* a Godot call, the
// direction that used to segfault: the generated wrapper handed Godot the address
// of the Go interface value, so Godot read the itab word as a T*. Each method
// returns 1 on success, a negative code for a specific failed assertion, or
// -900 if the Go side panicked at all.
//
// The panic guard is deliberate. The spec's guarantee is "no Go-side panic", so a
// panic has to be reported as a failing code rather than taking the engine down.

// objectArgGuard runs fn and converts a Go panic into the -900 failure code.
func objectArgGuard(label string, fn func() int32) (code int32) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("  object-arg test %s PANICKED: %v\n", label, r)
			code = -900
		}
	}()
	return fn()
}

// isNullObject reports whether obj is absent: nil interface, typed nil, or a
// wrapper with no owner. Reuses ObjectArgPtr's own nil semantics so the test asks
// the question the encoder answers.
//
// Do not use IsValid() for this: a generated object return always hands back a
// non-nil wrapper around a stack-allocated Impl, so IsValid() stays true even when
// Godot wrote a null owner into it.
func isNullObject(obj Wrapped) bool {
	return ObjectArgPtr(obj) == nil
}

// isNullRef is the Ref-shaped equivalent of isNullObject.
func isNullRef(r Ref) bool {
	return r == nil || !r.IsValid() || isNullObject(r.ToObject())
}

// TestObjectArgAddChild performs the call shape that used to crash: a Godot-owned
// node handed to Go, then passed back into Godot as an argument.
func (e *Example) TestObjectArgAddChild(child Node) int32 {
	return objectArgGuard("add_child", func() int32 {
		if isNullObject(child) {
			return -1
		}
		before := e.GetChildCount(false)
		e.AddChild(child, false, NODE_INTERNAL_MODE_INTERNAL_MODE_DISABLED)
		after := e.GetChildCount(false)
		if after != before+1 {
			fmt.Printf("  add_child: child count %d -> %d, want +1\n", before, after)
			return -2
		}
		parent := child.GetParent()
		if isNullObject(parent) {
			return -3
		}
		if parent.GetInstanceId() != e.GetInstanceId() {
			fmt.Printf("  add_child: parent id %d != example id %d\n", parent.GetInstanceId(), e.GetInstanceId())
			return -4
		}
		return 1
	})
}

// TestObjectArgIdentity wires two Godot-supplied nodes together and reads the
// relationship back out of the engine, proving both argument slots carried the
// right objects rather than merely surviving the call.
func (e *Example) TestObjectArgIdentity(parent Node, child Node) int32 {
	return objectArgGuard("identity", func() int32 {
		if isNullObject(parent) || isNullObject(child) {
			return -1
		}
		parent.AddChild(child, false, NODE_INTERNAL_MODE_INTERNAL_MODE_DISABLED)

		if child.GetParent().GetInstanceId() != parent.GetInstanceId() {
			return -2
		}
		// The child must be reachable through the parent's index, not just by pointer.
		viaIndex := parent.GetChild(0, false)
		if isNullObject(viaIndex) || viaIndex.GetInstanceId() != child.GetInstanceId() {
			return -3
		}
		return 1
	})
}

// TestObjectArgSetShape passes a Ref<Shape2D> into set_shape. The caller verifies
// in Godot that the engine really holds the object, which is the authoritative
// check that the argument arrived as the right object.
//
// The readback deliberately happens in GDScript rather than Go: the generated
// object-return path allocates a wrapper and pins it on every call, which leaves
// the returned resource reachable at shutdown and trips the leak check.
func (e *Example) TestObjectArgSetShape(owner CollisionShape2D, shape RefShape2D) int32 {
	return objectArgGuard("set_shape", func() int32 {
		if isNullObject(owner) || isNullRef(shape) {
			return -1
		}
		owner.SetShape(shape)
		return 1
	})
}

// TestObjectArgSetShapeTypedNil passes a typed-nil Ref, the shape a plain
// `!= nil` check lets through. The caller asserts the engine cleared the shape.
func (e *Example) TestObjectArgSetShapeTypedNil(owner CollisionShape2D) int32 {
	return objectArgGuard("set_shape_typed_nil", func() int32 {
		if isNullObject(owner) {
			return -1
		}
		var typedNil RefShape2D = (*RefBase[Shape2D])(nil)
		owner.SetShape(typedNil)
		return 1
	})
}

// TestObjectArgSetShapeInvalidRef passes a live Ref holding no object. The caller
// asserts the engine cleared the shape.
func (e *Example) TestObjectArgSetShapeInvalidRef(owner CollisionShape2D) int32 {
	return objectArgGuard("set_shape_invalid_ref", func() int32 {
		if isNullObject(owner) {
			return -1
		}
		owner.SetShape(NewRef[Shape2D](nil))
		return 1
	})
}

// TestObjectArgNilPlainObject covers a non-refcounted engine-class argument set to
// null. set_owner(null) is legal and clears the owner, so a real owner is set
// first to give the null something visible to undo. Godot requires the owner to be
// an ancestor in the tree, so the node is attached under this one before that.
func (e *Example) TestObjectArgNilPlainObject(node Node) int32 {
	return objectArgGuard("nil_plain_object", func() int32 {
		if isNullObject(node) {
			return -1
		}
		e.AddChild(node, false, NODE_INTERNAL_MODE_INTERNAL_MODE_DISABLED)
		node.SetOwner(e)
		if isNullObject(node.GetOwner()) {
			return -2
		}
		node.SetOwner(nil)
		if !isNullObject(node.GetOwner()) {
			return -3
		}
		return 1
	})
}

// TestObjectArgRelease forces the Go GC so finalizers on Ref values decoded from
// the object-argument calls run before the engine's exit-time leak check. Without
// it, references handed to Go by the decoder can still be held at shutdown and
// show up as leaked instances.
func (e *Example) TestObjectArgRelease() int32 {
	for i := 0; i < 20; i++ {
		runtime.GC()
		runtime.Gosched()
	}
	return 1
}

// TestObjectArgRefcountStability passes the same Ref across the boundary 50 times.
// Arguments are borrowed: Godot's PtrToArg<Ref<T>>::convert takes the receiving
// reference itself, so the count must not drift. A reinstated Reference() would
// show up here as +50.
func (e *Example) TestObjectArgRefcountStability(owner CollisionShape2D, shape RefShape2D) int32 {
	return objectArgGuard("refcount_stability", func() int32 {
		if isNullObject(owner) || shape == nil || !shape.IsValid() {
			return -1
		}
		owner.SetShape(shape)
		baseline := shape.Ptr().GetReferenceCount()

		const rounds = 50
		for i := 0; i < rounds; i++ {
			owner.SetShape(shape)
		}
		after := shape.Ptr().GetReferenceCount()
		if after != baseline {
			fmt.Printf("  refcount_stability: %d -> %d after %d calls (drift %d)\n",
				baseline, after, rounds, after-baseline)
			return -2
		}
		return 1
	})
}

// TestObjectArgBaseClassParam declares the base class and is called with a
// subclass instance. The argument metadata must advertise "Shape2D", not the
// owning class, or GDScript static typing rejects a typed CircleShape2D before
// the call is ever reached.
func (e *Example) TestObjectArgBaseClassParam(shape Shape2D) int32 {
	return objectArgGuard("base_class_param", func() int32 {
		if isNullObject(shape) {
			return -1
		}
		cls := shape.GetClass()
		defer cls.Destroy()
		if cls.ToUtf8() != "CircleShape2D" {
			fmt.Printf("  base_class_param: expected CircleShape2D, got %s\n", cls.ToUtf8())
			return -2
		}
		return 1
	})
}

// TestObjectArgReturnNode hands an engine class back to GDScript. The return
// PropertyInfo must advertise "Node"; if it still advertised the owning class the
// typed assignment in main.gd fails at parse time.
func (e *Example) TestObjectArgReturnNode(node Node) Node {
	return node
}
