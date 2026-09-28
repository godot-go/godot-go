package builtin

// Unit tests for the instance-binding decode. These need no engine: the helper
// only reads a Go value out of a cgo.Handle, so every shape the binding slot can
// hold is exercised here rather than in a Godot run.

import (
	"errors"
	"runtime/cgo"
	"testing"
)

// fakeObject satisfies Object without an engine by embedding the interface. Only
// identity is under test here, so the promoted methods are never called.
type fakeObject struct {
	Object
}

func requireBindingError(t *testing.T, err error, want string) {
	t.Helper()
	if err == nil {
		t.Fatalf("%s: want an error, got nil", want)
	}
	var bindingErr *ObjectBindingError
	if !errors.As(err, &bindingErr) {
		t.Fatalf("%s: want *ObjectBindingError, got %T (%v)", want, err, err)
	}
}

func TestObjectFromBindingPtrNullSlot(t *testing.T) {
	obj, err := objectFromBindingPtr(0)
	if obj != nil {
		t.Fatalf("null slot resolved to %v", obj)
	}
	requireBindingError(t, err, "null slot")
}

func TestObjectFromBindingPtrWrappedClassInstance(t *testing.T) {
	want := &fakeObject{}
	h := cgo.NewHandle(&WrappedClassInstance{Instance: want})
	defer h.Delete()

	got, err := objectFromBindingPtr(uintptr(h))
	if err != nil {
		t.Fatalf("wrapped class instance: %v", err)
	}
	if got != Object(want) {
		t.Fatalf("wrapped class instance: got %v, want %v", got, want)
	}
}

func TestObjectFromBindingPtrWrappedClassInstanceWithoutObject(t *testing.T) {
	h := cgo.NewHandle(&WrappedClassInstance{})
	defer h.Delete()

	obj, err := objectFromBindingPtr(uintptr(h))
	if obj != nil {
		t.Fatalf("empty wrapped class instance resolved to %v", obj)
	}
	requireBindingError(t, err, "empty wrapped class instance")
}

func TestObjectFromBindingPtrPlainObject(t *testing.T) {
	want := &fakeObject{}
	h := cgo.NewHandle(Object(want))
	defer h.Delete()

	got, err := objectFromBindingPtr(uintptr(h))
	if err != nil {
		t.Fatalf("plain object: %v", err)
	}
	if got != Object(want) {
		t.Fatalf("plain object: got %v, want %v", got, want)
	}
}

func TestObjectFromBindingPtrUnrecognisedValue(t *testing.T) {
	// A handle carrying something that is not a wrapper at all. This is the
	// case the old code would have silently cast into an Object.
	h := cgo.NewHandle(42)
	defer h.Delete()

	obj, err := objectFromBindingPtr(uintptr(h))
	if obj != nil {
		t.Fatalf("unrecognised value resolved to %v", obj)
	}
	requireBindingError(t, err, "unrecognised value")
}

func TestObjectFromBindingPtrDeletedHandle(t *testing.T) {
	// cgo.Handle.Value() panics on a deleted handle. The helper must turn that
	// into the typed error instead of letting it take the process down.
	h := cgo.NewHandle(&WrappedClassInstance{Instance: &fakeObject{}})
	slot := uintptr(h)
	h.Delete()

	obj, err := objectFromBindingPtr(slot)
	if obj != nil {
		t.Fatalf("deleted handle resolved to %v", obj)
	}
	requireBindingError(t, err, "deleted handle")
}

func TestObjectBindingErrorMessagesNameTheClass(t *testing.T) {
	withClass := &ObjectBindingError{Class: "Node", Reason: "no callbacks"}
	if got := withClass.Error(); got != "instance binding unresolved for class Node: no callbacks" {
		t.Fatalf("class-bearing message: %q", got)
	}
	without := &ObjectBindingError{Reason: "binding slot is null"}
	if got := without.Error(); got != "instance binding unresolved: binding slot is null" {
		t.Fatalf("classless message: %q", got)
	}
}

func TestAnnotateClassFillsOnlyEmptyClass(t *testing.T) {
	blank := &ObjectBindingError{Reason: "r"}
	annotateClass(blank, "Filled")
	if blank.Class != "Filled" {
		t.Fatalf("blank error not annotated: %q", blank.Class)
	}

	taken := &ObjectBindingError{Class: "Original", Reason: "r"}
	annotateClass(taken, "Overwritten")
	if taken.Class != "Original" {
		t.Fatalf("existing class was overwritten: %q", taken.Class)
	}
}
