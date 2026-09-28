package pkg

import (
	"errors"
	"fmt"
	"reflect"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

// In-engine tests for resolving a user-defined extension class through a plain
// engine-class parameter (openspec change
// fix-user-defined-class-object-arg-decode).
//
// Before the change this SIGSEGV'd: the instance-binding slot held a cgo.Handle
// value for user-defined classes, and the reader dereferenced it as a *Object,
// reading a handle number as a pointer.

// TestUserDefinedNodeArg receives a user-defined extension class instance
// through a plain Node parameter. The wrapper must be the object the caller
// passed, not a re-wrap built from a class-name lookup that could not find
// TestHierarchicalDerived at all.
func (e *Example) TestUserDefinedNodeArg(node Node, expectedID int64) int32 {
	return objectArgGuard("user_defined_node_arg", func() int32 {
		if isNullObject(node) {
			return -1
		}
		if got := int64(node.GetInstanceId()); got != expectedID {
			fmt.Printf("  user_defined_node_arg: FAIL: instance id %d, want %d\n", got, expectedID)
			return -2
		}
		cls := node.GetClass()
		defer cls.Destroy()
		if cls.ToUtf8() != "TestHierarchicalDerived" {
			fmt.Printf("  user_defined_node_arg: FAIL: class %s, want TestHierarchicalDerived\n", cls.ToUtf8())
			return -3
		}
		fmt.Printf("  user_defined_node_arg: resolved %s instance %d through a Node parameter\n",
			cls.ToUtf8(), expectedID)
		return 1
	})
}

// TestUserDefinedArgStable resolves the same engine object twice through the
// binding and requires the identical Go wrapper both times. A decode that built
// a fresh wrapper per call would show up here as pointer inequality, which is
// what the pass-through in the varcall branch avoids.
func (e *Example) TestUserDefinedArgStable(node Node) int32 {
	return objectArgGuard("user_defined_arg_stable", func() int32 {
		if isNullObject(node) {
			return -1
		}
		owner := (*GodotObject)(node.AsGDExtensionObjectPtr())

		first, err := ObjectFromInstanceBinding(owner)
		if err != nil {
			fmt.Printf("  user_defined_arg_stable: FAIL: first resolve: %v\n", err)
			return -2
		}
		second, err := ObjectFromInstanceBinding(owner)
		if err != nil {
			fmt.Printf("  user_defined_arg_stable: FAIL: second resolve: %v\n", err)
			return -3
		}
		if reflect.ValueOf(first).Pointer() != reflect.ValueOf(second).Pointer() {
			fmt.Printf("  user_defined_arg_stable: FAIL: two resolves produced different wrappers\n")
			return -4
		}
		fmt.Printf("  user_defined_arg_stable: the same wrapper resolves repeatedly\n")
		return 1
	})
}

// TestEngineClassArgStillResolves guards the shape change to
// GoCallback_GDExtensionBindingCreate. Engine-class bindings used to be the
// address of a Go interface; they are now handles, so a plain Node argument has
// to survive the new shape.
func (e *Example) TestEngineClassArgStillResolves(node Node, expectedID int64) int32 {
	return objectArgGuard("engine_class_arg", func() int32 {
		if isNullObject(node) {
			return -1
		}
		if got := int64(node.GetInstanceId()); got != expectedID {
			fmt.Printf("  engine_class_arg: FAIL: instance id %d, want %d\n", got, expectedID)
			return -2
		}
		cls := node.GetClass()
		defer cls.Destroy()
		if cls.ToUtf8() != "Node" {
			fmt.Printf("  engine_class_arg: FAIL: class %s, want Node\n", cls.ToUtf8())
			return -3
		}
		fmt.Printf("  engine_class_arg: plain Node still resolves through the handle shape (instance %d)\n", expectedID)
		return 1
	})
}

// TestUnresolvableBindingIsTypedError requires that a lookup which cannot
// resolve yields an *ObjectBindingError rather than a panic or a crash. This is
// the failure mode the change replaces: it used to be a SIGSEGV or a
// log.Panic that took the process with it.
func (e *Example) TestUnresolvableBindingIsTypedError() int32 {
	return objectArgGuard("unresolvable_binding", func() int32 {
		_, err := ObjectFromInstanceBinding(nil)
		if err == nil {
			fmt.Printf("  unresolvable_binding: FAIL: a null object resolved without error\n")
			return -2
		}
		var bindingErr *ObjectBindingError
		if !errors.As(err, &bindingErr) {
			fmt.Printf("  unresolvable_binding: FAIL: error is %T, want *ObjectBindingError\n", err)
			return -3
		}
		fmt.Printf("  unresolvable_binding: typed error surfaced: %s\n", bindingErr.Error())
		return 1
	})
}
