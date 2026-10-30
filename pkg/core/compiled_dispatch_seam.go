package core

// Dispatch observation seam for the in-engine allocation-budget test
// (typed-method-dispatch task 7.2, test/pkg/dispatch_alloc_budget_tests.go).
//
// The budget assertion has to measure the dispatch paths the engine itself
// uses — compiled closures where a bind attached one, the reflective body
// otherwise — and the in-engine Go tier drives registered classes, where the
// closures live in unexported GoMethodMetadata fields. This seam hands out a
// handle whose methods run exactly the engine-facing entry points
// (GoMethodMetadata.Ptrcall, GoMethodMetadata.CallWithError), so the
// measurement attributes one call to one dispatch path and the reflective
// control exercises the same seam as the compiled probes.
//
// Like the DecodePtrcallArgs and DecodeVarcallArgs seams, it adds no policy
// of its own and changes nothing about how engine-facing dispatch behaves.

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// MethodDispatcher runs one bound method's engine-facing dispatch entry
// points. Which path serves each call (compiled closure or reflective body)
// is decided by the bind, exactly as in engine dispatch.
type MethodDispatcher struct {
	md *GoMethodMetadata
}

// Ptrcall dispatches the direct native-call shape: the compiled closure when
// the bind attached one, the reflective body otherwise.
func (d MethodDispatcher) Ptrcall(inst GDClass, args []GDExtensionConstTypePtr, ret GDExtensionUninitializedTypePtr) {
	d.md.Ptrcall(inst, args, ret)
}

// Varcall dispatches the Variant-call shape (the CallWithError entry the
// varcall callback uses): the compiled closure when attached, the reflective
// body otherwise.
func (d MethodDispatcher) Varcall(inst GDClass, args []Variant) (Variant, error) {
	return d.md.CallWithError(inst, args...)
}

// Compiled reports whether both dispatch shapes of the bind run through
// compiled closures. A name-string registration (or an untypeable
// signature) reports false and its calls run the reflective body.
func (d MethodDispatcher) Compiled() bool {
	return d.md.typedDispatch != nil && d.md.typedVarcall != nil
}

// DispatchFor resolves a bound method of a registered class by the Godot
// name it resolves under — first among the instance/static methods, then the
// virtuals (e.g. "test_echo_string_name_arg" or "_notification") — and
// returns the dispatch handle. It reports found=false when the class or
// method is not registered, so the caller distinguishes a missing bind from a
// reflective one.
func DispatchFor(className string, gdMethodName string) (d MethodDispatcher, found bool) {
	ci, ok := Internal.GDRegisteredGDClasses.Get(className)
	if !ok {
		return MethodDispatcher{}, false
	}
	mcmi, ok := ci.MethodMap[gdMethodName]
	if !ok {
		mcmi, ok = ci.VirtualMethodMap[gdMethodName]
	}
	if !ok {
		return MethodDispatcher{}, false
	}
	return MethodDispatcher{md: mcmi.GoMethodMetadata}, true
}
