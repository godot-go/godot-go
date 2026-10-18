package pkg

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
)

// In-engine fixtures for bind-side class static methods (openspec change
// support-static-methods). The static fixtures are package-level functions --
// no receiver anywhere -- bound to Example through ClassDBBindStaticMethod
// and dispatched by the engine with a null instance pointer. The GDScript
// driver (test/demo/main.gd test_static_methods) calls them through the
// class path and reads misuse verdicts from the recovery probes below, where
// 1 is success and -1 names the failure.

// ExampleStaticAdd is the typed static: a statically-typed GDScript call site
// reaches it through the ptrcall fast path with a null instance pointer.
func ExampleStaticAdd(a int64, b int64) int64 {
	return a + b
}

// ExampleStaticWithDefault covers the trailing-default fill on a static
// bind; default-filled calls travel the varcall channel.
func ExampleStaticWithDefault(base int64, bonus int64) int64 {
	return base + bonus
}

// ExampleStaticSumVarargs is the variadic static: raw Variants collected on
// the CallSlice path with no receiver prepended.
func ExampleStaticSumVarargs(args ...Variant) Variant {
	var sum int64
	for i := range args {
		sum += args[i].ToInt64()
	}
	return NewVariantInt64(sum)
}

// ExampleStaticObjectClass decodes an Object argument on a bind that has no
// receiver to attach it to, and reports the runtime class name.
func ExampleStaticObjectClass(node Object) string {
	return node.GetClassName()
}

// ExampleStaticEchoNode returns its argument through the standard
// object-return ownership path from a static bind.
func ExampleStaticEchoNode(node Node) Node {
	return node
}

// staticArityProbeRan records whether the arity-probe body executed, so a
// rejected call is distinguishable from a dispatched one on the same bind
// (the precedent is the instance-path callErrorProbeRan in example.go).
var staticArityProbeRan bool

// ExampleStaticArityProbe is the dedicated target for the call-error
// scenario: it reports its body ran and returns a + 1.
func ExampleStaticArityProbe(a int64) int64 {
	staticArityProbeRan = true
	return a + 1
}

// StaticBadArityRejected drives the static bind through the engine's generic
// call channel with the wrong argument count and reports the verdict
// (1 == pass): the reject must not run the Go body (the engine reports it as
// a call error, which from GDScript would abort the calling function instead
// of letting us observe it, so it is driven here), and a correctly-shaped
// call on the same bind must dispatch immediately afterwards.
func (e *Example) StaticBadArityRejected() int64 {
	sn := NewStringNameWithLatin1Chars("static_arity_probe")
	defer sn.Destroy()

	// Too few arguments: rejected, body never runs.
	staticArityProbeRan = false
	func() {
		defer func() { _ = recover() }()
		e.Call(sn)
	}()
	if staticArityProbeRan {
		return -1
	}

	// Correct arity on the same static bind dispatches and returns a + 1.
	staticArityProbeRan = false
	var ret Variant
	func() {
		defer func() { _ = recover() }()
		arg := NewVariantInt64(40)
		defer arg.Destroy()
		ret = e.Call(sn, arg)
	}()
	if !staticArityProbeRan {
		return -2
	}
	if ret.ToInt64() != 41 {
		return -3
	}
	return 1
}

// StaticInstanceCallObservation records what the engine does when a static
// bind is invoked on an instance through the generic Object call channel:
// it returns the method's result (3) if the engine dispatched it, or a
// sentinel code (-100) if the Go call died. No engine legality is asserted;
// the driver prints the number as an observation. Every outcome must leave
// the process alive -- the driver's next assertion is the liveness proof.
func (e *Example) StaticInstanceCallObservation() int64 {
	sn := NewStringNameWithLatin1Chars("static_add")
	defer sn.Destroy()
	a := NewVariantInt64(1)
	defer a.Destroy()
	b := NewVariantInt64(2)
	defer b.Destroy()
	var ret Variant
	dispatched := false
	func() {
		defer func() { _ = recover() }()
		ret = e.Call(sn, a, b)
		dispatched = true
	}()
	if !dispatched {
		return -100
	}
	return ret.ToInt64()
}

// StaticDuplicateBindRecovered rebinds an already-registered method name as
// a static and reports whether the bind failed fatally and was recovered.
// The driver proves afterwards that the original registration still answers.
func (e *Example) StaticDuplicateBindRecovered() int64 {
	recovered := false
	func() {
		defer func() { recovered = recover() != nil }()
		ClassDBBindStaticMethod(e, "static_add", ExampleStaticAdd, []string{"a", "b"}, nil)
	}()
	if recovered {
		return 1
	}
	return -1
}

// StaticNonCallableRecovered passes a non-function to the static bind API
// and reports whether it failed fatally and was recovered.
func (e *Example) StaticNonCallableRecovered() int64 {
	recovered := false
	func() {
		defer func() { recovered = recover() != nil }()
		ClassDBBindStaticMethod(e, "static_never_bound", 42, nil, nil)
	}()
	if recovered {
		return 1
	}
	return -1
}
