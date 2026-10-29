package core

import (
	"encoding/binary"
	"reflect"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// typed-method-dispatch task 7.2 — allocation budget matrix (engine-free tier).
//
// The dispatch-path promise of this change is that a compiled bind call runs
// with no reflection and no per-call heap traffic: bind-time factories pay any
// reflection (allowed, once per bind), and each call executes concrete loads,
// a direct handler call, and a concrete store. This file pins the per-call
// budget across a representative arity matrix — 0/2/5 arguments, every ptrcall
// return style (value, void, copy-return), instance/static/virtual binds —
// through testing.AllocsPerRun with a hard cap of zero allocations per call.
//
// Every cell builds its wrapper OUTSIDE the measured run, mirroring the
// production lifecycle where the factory instantiates once at bind time and
// the closure alone serves each call (measuring a factory call inside the
// loop would attribute bind-time boxing to the dispatch path).
//
// Falsification works in both directions:
//
//   - The reflective control below measures the same dispatch through the
//     name-string body and asserts a strictly positive floor (measured: 32
//     allocations per call at arity 1). If dispatch regressed to the
//     reflective body, the zero-cap cells would light up exactly like that
//     control.
//   - The seam cells drive GoMethodMetadata.Ptrcall with typedDispatch set
//     and Func left zero, so any reintroduced reflective hop inside the
//     dispatch seam panics on the zero Value instead of silently passing.
//
// What lives in the GODOT tier instead (test/pkg/dispatch_alloc_budget_tests.go):
// the varcall closures' call execution, whose decoders/encoders cross into
// the engine through Variant constructors and the RetCopyEncodable methods,
// the copy-return echo path with a real builtin return (the stub's echo
// compare is nil by contract), and the real engine→Go dispatch through
// registered classes. Build-phase allocation (the factories instantiating per
// bind) is by design nonzero-once and pinned engine-free in
// TestVarcallFactoriesBuildEngineFree.

// dispatchAllocCap is the per-call allocation cap for every compiled dispatch
// cell. Zero is the contract: the wrappers' hot bodies carry no interface
// boxes, no slices, no allocations ("no interface conversion and no
// allocation", loadTypedSlot/storeTypedRet comments).
const dispatchAllocCap = 0

// measureDispatch runs an already-built dispatch fn through
// testing.AllocsPerRun and fails when the per-call allocation count exceeds
// the budget cap. The warm-up run inside AllocsPerRun absorbs any lazily
// initialized one-time cost.
func measureDispatch(t *testing.T, name string, fn func()) {
	t.Helper()
	n := testing.AllocsPerRun(200, fn)
	if n > dispatchAllocCap {
		t.Fatalf("%s allocates %v per call, budget %v", name, n, dispatchAllocCap)
	}
}

// TestDispatchBudgetInstancePtrcall covers the instance matrix: arity 0/2/5
// for the value family, arity 0/2/5 for the void family, and copy-return at
// arity 0 and 2 (the copy family engine-free through the spotRetStub raw
// encode/counting-destroy pair).
func TestDispatchBudgetInstancePtrcall(t *testing.T) {
	stub := &spotStubClass{}
	var cell [8]byte
	binary.LittleEndian.PutUint64(cell[:], 3)
	arg := GDExtensionConstTypePtr(&cell[0])
	args0 := []GDExtensionConstTypePtr{}
	args2 := []GDExtensionConstTypePtr{arg, arg}
	args5 := []GDExtensionConstTypePtr{arg, arg, arg, arg, arg}
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])

	// Value family — raw load, direct call, raw store.
	value0 := wrapPtrcallValue0[spotStubClass, *spotStubClass](
		func(self *spotStubClass) int64 { return 7 })
	measureDispatch(t, "instance value-ret arity-0", func() {
		value0(stub, args0, retPtr)
	})
	value2 := wrapPtrcallValue2[spotStubClass, *spotStubClass](
		func(self *spotStubClass, n int64, f float64) int64 { return n })
	measureDispatch(t, "instance value-ret arity-2", func() {
		value2(stub, args2, retPtr)
	})
	value5 := wrapPtrcallValue5[spotStubClass, *spotStubClass](
		func(self *spotStubClass, a, b, c, d, e int64) int64 { return a + b + c + d + e })
	measureDispatch(t, "instance value-ret arity-5", func() {
		value5(stub, args5, retPtr)
	})

	// Void family — same loads, no store.
	none0 := wrapPtrcallNone0[spotStubClass, *spotStubClass](func(self *spotStubClass) {})
	measureDispatch(t, "instance void-ret arity-0", func() {
		none0(stub, args0, nil)
	})
	none2 := wrapPtrcallNone2[spotStubClass, *spotStubClass](func(self *spotStubClass, n int64, f float64) {})
	measureDispatch(t, "instance void-ret arity-2", func() {
		none2(stub, args2, nil)
	})
	none5 := wrapPtrcallNone5[spotStubClass, *spotStubClass](func(self *spotStubClass, a, b, c, d, e int64) {})
	measureDispatch(t, "instance void-ret arity-5", func() {
		none5(stub, args5, nil)
	})

	// Copy-return family — echo gate, RetEncodeInto, RetDestroy. The stub
	// methods are raw memory with a counter, and its echo compare is nil by
	// contract, so the cells exercise the always-destroy dispatch; the
	// echo-suppressed path with a real builtin return crosses into the engine
	// (copy ctor) and is budgeted in the GODOT tier.
	copy0 := wrapPtrcallCopyRet0[spotStubClass, *spotStubClass](
		func(self *spotStubClass) spotRetStub {
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], 9)
			return out
		})
	measureDispatch(t, "instance copy-ret arity-0", func() {
		copy0(stub, args0, retPtr)
	})
	copy2 := wrapPtrcallCopyRet2[spotStubClass, *spotStubClass, int64, float64, spotRetStub](
		func(self *spotStubClass, n int64, f float64) spotRetStub {
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], uint64(n))
			return out
		})
	measureDispatch(t, "instance copy-ret arity-2", func() {
		copy2(stub, args2, retPtr)
	})
}

// TestDispatchBudgetStaticPtrcall covers the static matrix (no receiver: the
// inst argument is nil and never asserted).
func TestDispatchBudgetStaticPtrcall(t *testing.T) {
	var cell [8]byte
	binary.LittleEndian.PutUint64(cell[:], 3)
	arg := GDExtensionConstTypePtr(&cell[0])
	args0 := []GDExtensionConstTypePtr{}
	args2 := []GDExtensionConstTypePtr{arg, arg}
	args5 := []GDExtensionConstTypePtr{arg, arg, arg, arg, arg}
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])

	value0 := wrapStaticPtrcallValue0(func() int64 { return 1 })
	measureDispatch(t, "static value-ret arity-0", func() {
		value0(nil, args0, retPtr)
	})
	value2 := wrapStaticPtrcallValue2(func(n int64, f float64) int64 { return n })
	measureDispatch(t, "static value-ret arity-2", func() {
		value2(nil, args2, retPtr)
	})
	value5 := wrapStaticPtrcallValue5(func(a, b, c, d, e int64) int64 { return a + b + c + d + e })
	measureDispatch(t, "static value-ret arity-5", func() {
		value5(nil, args5, retPtr)
	})
	none0 := wrapStaticPtrcallNone0(func() {})
	measureDispatch(t, "static void-ret arity-0", func() {
		none0(nil, args0, nil)
	})
	none2 := wrapStaticPtrcallNone2(func(n int64, f float64) {})
	measureDispatch(t, "static void-ret arity-2", func() {
		none2(nil, args2, nil)
	})
	none5 := wrapStaticPtrcallNone5(func(a, b, c, d, e int64) {})
	measureDispatch(t, "static void-ret arity-5", func() {
		none5(nil, args5, nil)
	})
	copy2 := wrapStaticPtrcallCopyRet2(func(n int64, f float64) spotRetStub {
		var out spotRetStub
		binary.LittleEndian.PutUint64(out[:], uint64(n))
		return out
	})
	measureDispatch(t, "static copy-ret arity-2", func() {
		copy2(nil, args2, retPtr)
	})
}

// TestDispatchBudgetVirtualPtrcall pins the virtual bind family. Virtual
// binds share the instance wrapper shapes (the registration flag, not the
// closure, is what differs); this cell proves the shared-shape claim stays
// allocation-free for the V_-bound path used by _get_maximum_size and
// friends.
func TestDispatchBudgetVirtualPtrcall(t *testing.T) {
	stub := &spotStubClass{}
	var cell [8]byte
	binary.LittleEndian.PutUint64(cell[:], 3)
	arg := GDExtensionConstTypePtr(&cell[0])
	argsEmpty := []GDExtensionConstTypePtr{}
	args1 := []GDExtensionConstTypePtr{arg}
	args2x := []GDExtensionConstTypePtr{arg, arg}
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])

	value0 := wrapPtrcallValue0[spotStubClass, *spotStubClass](
		func(self *spotStubClass) int64 { return 16 })
	measureDispatch(t, "virtual value-ret arity-0", func() {
		value0(stub, argsEmpty, retPtr)
	})
	value2 := wrapPtrcallValue2[spotStubClass, *spotStubClass](
		func(self *spotStubClass, n int64, f float64) int64 { return n })
	measureDispatch(t, "virtual value-ret arity-2", func() {
		value2(stub, args2x, retPtr)
	})
	copy1 := wrapPtrcallCopyRet1[spotStubClass, *spotStubClass, int64, spotRetStub](
		func(self *spotStubClass, n int64) spotRetStub {
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], uint64(n))
			return out
		})
	measureDispatch(t, "virtual copy-ret arity-1", func() {
		copy1(stub, args1, retPtr)
	})
}

// TestDispatchBudgetThroughMetadataSeam drives the budget through the real
// dispatch entry point (GoMethodMetadata.Ptrcall) for the arity matrix rather
// than the factories alone. Func stays the zero reflect.Value: a regression
// that routes the seam back through the reflective body panics on it, so the
// passing assertion also proves the compiled closure served every call.
func TestDispatchBudgetThroughMetadataSeam(t *testing.T) {
	stub := &spotStubClass{}
	var cell [8]byte
	binary.LittleEndian.PutUint64(cell[:], 3)
	arg := GDExtensionConstTypePtr(&cell[0])
	argsEmpty := []GDExtensionConstTypePtr{}
	args2 := []GDExtensionConstTypePtr{arg, arg}
	args5 := []GDExtensionConstTypePtr{arg, arg, arg, arg, arg}
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])

	md := &GoMethodMetadata{
		ClassName:    "SpotStub",
		GdMethodName: "budget",
		GoMethodName: "Budget",
	}

	md.typedDispatch = wrapPtrcallValue0[spotStubClass, *spotStubClass](func(self *spotStubClass) int64 { return 1 })
	measureDispatch(t, "seam value-ret arity-0", func() {
		md.Ptrcall(stub, argsEmpty, retPtr)
	})

	md.typedDispatch = wrapPtrcallValue2[spotStubClass, *spotStubClass](func(self *spotStubClass, n int64, f float64) int64 { return n })
	measureDispatch(t, "seam value-ret arity-2", func() {
		md.Ptrcall(stub, args2, retPtr)
	})

	md.typedDispatch = wrapPtrcallNone5[spotStubClass, *spotStubClass](func(self *spotStubClass, a, b, c, d, e int64) {})
	measureDispatch(t, "seam void-ret arity-5", func() {
		md.Ptrcall(stub, args5, nil)
	})

	md.typedDispatch = wrapPtrcallCopyRet2[spotStubClass, *spotStubClass, int64, float64, spotRetStub](
		func(self *spotStubClass, n int64, f float64) spotRetStub {
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], uint64(n))
			return out
		})
	measureDispatch(t, "seam copy-ret arity-2", func() {
		md.Ptrcall(stub, args2, retPtr)
	})
}

// TestDispatchBudgetNotificationClosure pins the D6 callback closure budget:
// an engine notification through a compiled callback is one interface
// assertion plus one direct call.
func TestDispatchBudgetNotificationClosure(t *testing.T) {
	stub := &spotStubClass{}
	cb := newNotificationShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, what int32, reversed bool) {})
	measureDispatch(t, "notification callback", func() {
		cb(stub, 1, false)
	})
}

// TestDispatchForSeam pins the observation seam the GODOT-tier budget test
// rides (compiled_dispatch_seam.go): resolution by class/gd-name, and the
// Compiled() report distinguishing a compiled bind from a reflective one. The
// dispatch methods forward verbatim to GoMethodMetadata.Ptrcall, already
// pinned by the seam cells above, so one call through the handle is enough
// here to prove the forwarding is live.
func TestDispatchForSeam(t *testing.T) {
	stub := &spotStubClass{}
	var cell [8]byte
	binary.LittleEndian.PutUint64(cell[:], 4)
	arg := GDExtensionConstTypePtr(&cell[0])
	args1 := []GDExtensionConstTypePtr{arg}
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])

	compiledMD := &GoMethodMetadata{
		ClassName:     "SpotStub",
		GdMethodName:  "compiled_probe",
		GoMethodName:  "CompiledProbe",
		typedDispatch: wrapPtrcallValue1[spotStubClass, *spotStubClass]((*spotStubClass).Triple),
		typedVarcall: wrapVarcallValue1[spotStubClass, *spotStubClass, int64, int64](
			(*spotStubClass).Triple, nil, "compiled_probe"),
	}
	reflectiveMD := &GoMethodMetadata{
		ClassName:    "SpotStub",
		GdMethodName: "reflective_probe",
		GoMethodName: "ReflectiveProbe",
	}
	registry := map[string]*ClassInfo{
		"SpotStub": {
			Name: "SpotStub",
			MethodMap: map[string]*MethodBindAndClassMethodInfo{
				"compiled_probe":   {GoMethodMetadata: compiledMD},
				"reflective_probe": {GoMethodMetadata: reflectiveMD},
			},
		},
	}

	withRegisteredClasses(t, registry, func() {
		d, found := DispatchFor("SpotStub", "compiled_probe")
		if !found {
			t.Fatal("compiled_probe not resolved through the seam")
		}
		if !d.Compiled() {
			t.Fatal("compiled_probe reports uncompiled with typedDispatch attached")
		}
		d.Ptrcall(stub, args1, retPtr)
		if got := binary.LittleEndian.Uint64(ret[:]); got != 12 {
			t.Fatalf("seam dispatch produced %d, want 12 (4*3 through the closure)", got)
		}

		d, found = DispatchFor("SpotStub", "reflective_probe")
		if !found {
			t.Fatal("reflective_probe not resolved through the seam")
		}
		if d.Compiled() {
			t.Fatal("reflective_probe reports compiled with no closures attached")
		}
		// The unattached dispatch falls through to the reflective body, which
		// panics on the zero Func — proving the handle forwards rather than
		// silently no-op'ing.
		if _, fired := recoverFatal(func() { d.Ptrcall(stub, args1, retPtr) }); !fired {
			t.Fatal("reflective handle did not forward to the reflective body")
		}

		if _, found := DispatchFor("SpotStub", "missing_method"); found {
			t.Fatal("missing method resolved")
		}
		if _, found := DispatchFor("NoSuchClass", "compiled_probe"); found {
			t.Fatal("unregistered class resolved")
		}
	})
}

// TestDispatchBudgetReflectiveControlAllocates is the falsification control:
// the same dispatch through the name-string reflective body must measure a
// strictly positive per-call cost. If someone reintroduces reflect machinery
// on the compiled path, this control is what proves the zero-cap cells above
// can actually detect it (a zero-cap assertion is only meaningful when the
// alternative is observably nonzero). Measured baseline: 9 allocations per
// call at arity 1 (32 before the lazy dispatch payloads of
// trim-dispatch-overhead); the floor asserts well below that to stay stable.
func TestDispatchBudgetReflectiveControlAllocates(t *testing.T) {
	InitPrimativeTypeEncodersForUnitTests()

	stub := &spotStubClass{}
	var cell [8]byte
	binary.LittleEndian.PutUint64(cell[:], 7)
	arg := GDExtensionConstTypePtr(&cell[0])
	args1 := []GDExtensionConstTypePtr{arg}
	var ret [8]byte

	md := &GoMethodMetadata{
		ClassName:        "SpotStub",
		GdMethodName:     "triple",
		GoMethodName:     "Triple",
		Func:             reflect.ValueOf((*spotStubClass).Triple),
		GoReturnType:     reflect.TypeFor[int64](),
		GoReturnStyle:    ValueReturnStyle,
		GoArgumentTypes:  []reflect.Type{reflect.TypeFor[int64]()},
		gdeArgumentTypes: []GDExtensionVariantType{GDEXTENSION_VARIANT_TYPE_INT},
	}

	n := testing.AllocsPerRun(200, func() {
		md.Ptrcall(stub, args1, GDExtensionUninitializedTypePtr(&ret[0]))
	})
	if n <= dispatchAllocCap {
		t.Fatal("reflective control measured zero allocations: the control cannot " +
			"detect regression if the reflective path itself stops allocating")
	}
}
