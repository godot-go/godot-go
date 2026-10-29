package builtin

import (
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// ObjectArgPtr encodes an engine-class object as the value of a ptrcall argument.
//
// Godot reads an object argument by dereferencing the argument slot:
// PtrToArg<T*>::convert is *reinterpret_cast<T*const*>(p_ptr)
// (godot/core/variant/method_ptrcall.h). The slot must therefore point at a cell
// whose *contents* are the object pointer. Passing the address of a Go interface
// instead hands Godot the {itab, dataptr} header, which it reads as an object
// pointer and crashes.
//
// The typed-nil test is load-bearing rather than defensive padding. Generated Impl
// types embed WrappedImpl by value (NodeImpl -> ObjectImpl -> WrappedImpl), so a
// call to a promoted method on a typed nil faults while Go addresses the embedded
// field, before any nil check inside the method body can run. A plain
// `obj != nil` test does not catch a typed nil, because a typed nil satisfies it.
// The contract's IsNil answers through a method declared on the outer pointer
// type of every wrapper (codegen for generated classes; hand-written wrappers
// must declare it too), so the check is an ordinary interface call instead of
// reflect.ValueOf.
func ObjectArgPtr(obj Wrapped) GDExtensionObjectPtr {
	if obj == nil {
		return nil
	}
	if obj.IsNil() {
		return nil
	}
	return obj.AsGDExtensionObjectPtr()
}

// RefArgPtr encodes a Ref[T] smart-pointer argument as the referenced object's
// pointer, or nil when no object is held.
//
// IsValid() is checked before ToObject() because RefBase.ToObject() returns the
// held value directly and so dereferences a nil *RefBase. IsValid() is safe on a
// nil receiver: it is declared on *RefBase[T] rather than promoted from an
// embedded value, and short-circuits on `r != nil`.
//
// No reference is acquired here. Godot's PtrToArg<Ref<T>>::convert constructs
// Ref<T>(ptr) itself, which takes the reference on the receiving side.
func RefArgPtr(ref Ref) GDExtensionObjectPtr {
	if ref == nil || !ref.IsValid() {
		return nil
	}
	return ObjectArgPtr(ref.ToObject())
}
