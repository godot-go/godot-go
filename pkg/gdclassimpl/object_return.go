package gdclassimpl

// The generated object-return branch hands the engine a pointer-sized slot for
// the object it returns: `var ret <Declared>Impl`, ptrcall writes the object
// pointer into it, and the old code returned `&ret`. That wrapper is typed to
// the method's declared class, so every caller that wants the object's runtime
// class had to re-resolve the instance binding first, and a plain Go type
// assertion to the runtime class failed.
//
// This file owns the resolution that closes that gap. It is unexported on
// purpose: the behaviour belongs to the generated methods, and application code
// should see neither a cast entry point nor a helper call. See openspec change
// resolve-object-returns-through-binding.

import (
	"reflect"

	"github.com/godot-go/godot-go/pkg/builtin"
)

// resolveGoStructObjectReturn returns what a generated method should return for
// a non-refcounted engine object: the wrapper bound for the object's runtime
// class when one resolves and satisfies T, otherwise the declared-class wrapper
// the method built.
//
// T is the method's declared return type, passed explicitly by the generated
// call (`resolveGoStructObjectReturn[Node](&ret)`). It must not be left to
// inference: the argument's static type is the declared *struct* pointer
// (*NodeImpl), and a runtime wrapper for a subclass (*LabelImpl) is not that
// type, so an inferred T would make every call take the fallback branch.
//
// Requiring the bound wrapper to satisfy T is what keeps the fallback safe. A
// binding whose class model disagrees with the engine's cannot replace a
// working declared-class wrapper, so no return path that worked before starts
// failing.
//
// A return slot holding a null object pointer resolves to an error inside
// builtin.ObjectFromInstanceBinding and falls back here, which is exactly the
// wrapper the method returned before this change.
func resolveGoStructObjectReturn[T builtin.Object](declared T) T {
	// The generated call site always passes &ret, but the helper is general:
	// a typed-nil wrapper must not reach GetGodotObjectOwner, which would fault
	// while addressing the embedded WrappedImpl.
	if isNilObjectWrapper(declared) {
		var zero T
		return zero
	}
	bound, err := builtin.ObjectFromInstanceBinding(declared.GetGodotObjectOwner())
	return resolveObjectReturnFrom(declared, bound, err)
}

// resolveObjectReturnFrom is the engine-free half of the resolution: given the
// declared wrapper and the result of the instance-binding lookup, it chooses
// between the bound wrapper and the declared one.
//
// Splitting it out is what lets the four branches be unit-tested with fake
// wrappers, the same reason builtin.objectFromBindingPtr takes a handle rather
// than an engine pointer.
func resolveObjectReturnFrom[T builtin.Object](declared T, bound builtin.Object, err error) T {
	if err == nil && bound != nil {
		if resolved, ok := bound.(T); ok {
			return resolved
		}
	}
	return declared
}

// isNilObjectWrapper reports whether an object-typed value holds no wrapper.
//
// A plain `obj == nil` test does not catch a typed nil, because an interface
// holding a nil *NodeImpl satisfies it. The reflect check mirrors
// builtin.ObjectArgPtr, which carries the fault this guards against.
func isNilObjectWrapper(obj builtin.Object) bool {
	if obj == nil {
		return true
	}
	v := reflect.ValueOf(obj)
	return v.Kind() == reflect.Ptr && v.IsNil()
}
