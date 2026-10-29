package core

import (
	"encoding/binary"
	"reflect"
	"runtime/cgo"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// trim-dispatch-overhead task 1.3 — gating allocation budget over the whole
// benchmark shape set (dispatch-cost-baseline capability).
//
// typed_dispatch_budget_test.go gates the compiled ptrcall families; this file
// extends the same gate to every shape the benchmark set measures (instance +
// static value/void/copy-ret, the six callback closures, the prologue
// resolve-and-dispatch cells, and the metadata seam) so a reintroduced
// per-call allocation on ANY dispatched path fails the target — not just the
// ones the previous change happened to enumerate.
//
// Gate semantics (per the capability's requirements):
//
//   - Each shape carries a budget (measured floor + noise margin) and FAILS
//     when its measured per-call allocations exceed it. The compiled shapes
//     hold a zero budget — allocation counts from testing.AllocsPerRun are
//     exact, so the margin is zero; the floor-plus-margin language comes into
//     play on the reflective control, whose expected figure can drift with Go
//     versions.
//   - Every run reports each shape's measured figure beside its budget.
//   - Vacuity guard: each shape's handler increments a witness once per
//     dispatch; if the witness does not advance in step with the measured
//     calls, the budget would pass on a shape that stopped measuring anything,
//     so the target fails naming the vacuous shape.
//   - The reflective control must keep allocating (floor assertion): if the
//     name-string path ever stops allocating, the zero-cap rows lose their
//     falsification partner and the run fails.

// budgetGate runs one shape through testing.AllocsPerRun and applies the gate:
// budget comparison, measured-beside-budget report, and the vacuity guard on
// the handler witness (which must advance at least once per measured call).
func budgetGate(t *testing.T, name string, budget int, fn func(), witness *int64) {
	t.Helper()
	const runs = 200
	before := *witness
	n := testing.AllocsPerRun(runs, fn)
	after := *witness
	t.Logf("dispatch budget %-38s measured=%g budget=%d handler_calls=%d", name, n, budget, after-before)
	if after-before < runs {
		t.Errorf("%s: vacuous budget — handler ran %d times across %d measured "+
			"calls; the shape stopped dispatching, so the budget cannot observe it",
			name, after-before, runs)
		return
	}
	if n > float64(budget) {
		t.Errorf("%s allocates %g per call, budget %d", name, n, budget)
	}
}

// TestDispatchBudgetGateInstancePtrcall gates the instance families: value and
// void at arity 0/2/5, copy-return at arity 0/1/2.
func TestDispatchBudgetGateInstancePtrcall(t *testing.T) {
	stub := &spotStubClass{}
	args0, args2, args5 := benchDispatchCells()
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])

	var w int64
	value0 := wrapPtrcallValue0[spotStubClass, *spotStubClass](
		func(self *spotStubClass) int64 { w++; return 7 })
	budgetGate(t, "instance value-ret arity-0", 0, func() { value0(stub, args0, retPtr) }, &w)
	value2 := wrapPtrcallValue2[spotStubClass, *spotStubClass](
		func(self *spotStubClass, n int64, f float64) int64 { w++; return n })
	budgetGate(t, "instance value-ret arity-2", 0, func() { value2(stub, args2, retPtr) }, &w)
	value5 := wrapPtrcallValue5[spotStubClass, *spotStubClass](
		func(self *spotStubClass, a, b, c, d, e int64) int64 { w++; return a + b + c + d + e })
	budgetGate(t, "instance value-ret arity-5", 0, func() { value5(stub, args5, retPtr) }, &w)

	none0 := wrapPtrcallNone0[spotStubClass, *spotStubClass](func(self *spotStubClass) { w++ })
	budgetGate(t, "instance void-ret arity-0", 0, func() { none0(stub, args0, nil) }, &w)
	none2 := wrapPtrcallNone2[spotStubClass, *spotStubClass](
		func(self *spotStubClass, n int64, f float64) { w++ })
	budgetGate(t, "instance void-ret arity-2", 0, func() { none2(stub, args2, nil) }, &w)
	none5 := wrapPtrcallNone5[spotStubClass, *spotStubClass](
		func(self *spotStubClass, a, b, c, d, e int64) { w++ })
	budgetGate(t, "instance void-ret arity-5", 0, func() { none5(stub, args5, nil) }, &w)

	copy0 := wrapPtrcallCopyRet0[spotStubClass, *spotStubClass](
		func(self *spotStubClass) spotRetStub {
			w++
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], 9)
			return out
		})
	budgetGate(t, "instance copy-ret arity-0", 0, func() { copy0(stub, args0, retPtr) }, &w)
	var cell1 [8]byte
	binary.LittleEndian.PutUint64(cell1[:], 3)
	args1 := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(&cell1[0])}
	copy1 := wrapPtrcallCopyRet1[spotStubClass, *spotStubClass, int64, spotRetStub](
		func(self *spotStubClass, n int64) spotRetStub {
			w++
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], uint64(n))
			return out
		})
	budgetGate(t, "instance copy-ret arity-1", 0, func() { copy1(stub, args1, retPtr) }, &w)
	copy2 := wrapPtrcallCopyRet2[spotStubClass, *spotStubClass, int64, float64, spotRetStub](
		func(self *spotStubClass, n int64, f float64) spotRetStub {
			w++
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], uint64(n))
			return out
		})
	budgetGate(t, "instance copy-ret arity-2", 0, func() { copy2(stub, args2, retPtr) }, &w)
}

// TestDispatchBudgetGateStaticPtrcall gates the static families.
func TestDispatchBudgetGateStaticPtrcall(t *testing.T) {
	args0, args2, args5 := benchDispatchCells()
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])

	var w int64
	value0 := wrapStaticPtrcallValue0(func() int64 { w++; return 1 })
	budgetGate(t, "static value-ret arity-0", 0, func() { value0(nil, args0, retPtr) }, &w)
	value2 := wrapStaticPtrcallValue2(func(n int64, f float64) int64 { w++; return n })
	budgetGate(t, "static value-ret arity-2", 0, func() { value2(nil, args2, retPtr) }, &w)
	value5 := wrapStaticPtrcallValue5(func(a, b, c, d, e int64) int64 { w++; return a + b + c + d + e })
	budgetGate(t, "static value-ret arity-5", 0, func() { value5(nil, args5, retPtr) }, &w)

	none0 := wrapStaticPtrcallNone0(func() { w++ })
	budgetGate(t, "static void-ret arity-0", 0, func() { none0(nil, args0, nil) }, &w)
	none2 := wrapStaticPtrcallNone2(func(n int64, f float64) { w++ })
	budgetGate(t, "static void-ret arity-2", 0, func() { none2(nil, args2, nil) }, &w)
	none5 := wrapStaticPtrcallNone5(func(a, b, c, d, e int64) { w++ })
	budgetGate(t, "static void-ret arity-5", 0, func() { none5(nil, args5, nil) }, &w)

	copy0 := wrapStaticPtrcallCopyRet0(func() spotRetStub {
		w++
		var out spotRetStub
		binary.LittleEndian.PutUint64(out[:], 9)
		return out
	})
	budgetGate(t, "static copy-ret arity-0", 0, func() { copy0(nil, args0, retPtr) }, &w)
	var cell1 [8]byte
	binary.LittleEndian.PutUint64(cell1[:], 3)
	args1 := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(&cell1[0])}
	copy1 := wrapStaticPtrcallCopyRet1(func(n int64) spotRetStub {
		w++
		var out spotRetStub
		binary.LittleEndian.PutUint64(out[:], uint64(n))
		return out
	})
	budgetGate(t, "static copy-ret arity-1", 0, func() { copy1(nil, args1, retPtr) }, &w)
	copy2 := wrapStaticPtrcallCopyRet2(func(n int64, f float64) spotRetStub {
		w++
		var out spotRetStub
		binary.LittleEndian.PutUint64(out[:], uint64(n))
		return out
	})
	budgetGate(t, "static copy-ret arity-2", 0, func() { copy2(nil, args2, retPtr) }, &w)
}

// TestDispatchBudgetGateCallbackShapes gates the six D6 callback closures.
func TestDispatchBudgetGateCallbackShapes(t *testing.T) {
	stub := &spotStubClass{}
	planted := plantedVariant(42)
	var w int64

	notif := newNotificationShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, what int32, reversed bool) { w++ })
	budgetGate(t, "notification closure", 0, func() { notif(stub, 16, false) }, &w)

	propGet := newPropGetShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name string) (Variant, bool) { w++; return planted, true })
	budgetGate(t, "prop-get closure", 0, func() { propGet(stub, "dproperty_1") }, &w)

	propSet := newPropSetShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name string, value Variant) bool { w++; return true })
	budgetGate(t, "prop-set closure", 0, func() { propSet(stub, "dproperty_2", planted) }, &w)

	canRevert := newPropCanRevertShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name StringName) bool { w++; return true })
	budgetGate(t, "prop-can-revert closure", 0,
		func() { canRevert(stub, StringName{1, 2, 3, 4, 5, 6, 7, 8}) }, &w)

	getRevert := newPropGetRevertShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name StringName) (Variant, bool) { w++; return planted, true })
	budgetGate(t, "prop-get-revert closure", 0, func() { getRevert(stub, StringName{}) }, &w)

	toStr := newToStringShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass) string { w++; return "[ GDExtension::Stub <--> Instance ID:1 ]" })
	budgetGate(t, "to-string closure", 0, func() { toStr(stub) }, &w)
}

// TestDispatchBudgetGateSeam gates dispatch through the real entry points with
// Func left the zero reflect.Value: a regression routing the seam back through
// the reflective body panics on it instead of passing silently.
func TestDispatchBudgetGateSeam(t *testing.T) {
	stub := &spotStubClass{}
	_, args2, args5 := benchDispatchCells()
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])

	var w int64
	md := &GoMethodMetadata{
		ClassName:    "SpotStub",
		GdMethodName: "budget_gate",
		GoMethodName: "BudgetGate",
	}
	md.typedDispatch = wrapPtrcallValue2[spotStubClass, *spotStubClass](
		func(self *spotStubClass, n int64, f float64) int64 { w++; return n })
	budgetGate(t, "seam value-ret arity-2", 0, func() { md.Ptrcall(stub, args2, retPtr) }, &w)
	md.typedDispatch = wrapPtrcallNone5[spotStubClass, *spotStubClass](
		func(self *spotStubClass, a, b, c, d, e int64) { w++ })
	budgetGate(t, "seam void-ret arity-5", 0, func() { md.Ptrcall(stub, args5, nil) }, &w)
	md.typedDispatch = wrapPtrcallCopyRet2[spotStubClass, *spotStubClass, int64, float64, spotRetStub](
		func(self *spotStubClass, n int64, f float64) spotRetStub {
			w++
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], uint64(n))
			return out
		})
	budgetGate(t, "seam copy-ret arity-2", 0, func() { md.Ptrcall(stub, args2, retPtr) }, &w)

	md.typedVarcall = func(inst GDClass, args []Variant) (Variant, error) {
		w++
		return Variant{}, nil
	}
	budgetGate(t, "seam varcall value-ret arity-0", 0,
		func() { md.CallWithError(stub) }, &w)
}

// TestDispatchBudgetGatePrologue gates the prologue resolve-and-dispatch cells
// (cgo handle, class-name acquisition, registry lookup, virtual-map lookup,
// compiled callback). Same engine-free reach caveat as the benchmark cells:
// the current prologues' engine GetClass()/ToUtf8()/Destroy() ceremony cannot
// run without a live engine and is gated in-engine by the task 1.4/1.5 pair.
func TestDispatchBudgetGatePrologue(t *testing.T) {
	restore := benchPrologueStub()
	defer restore()
	inst := &benchPrologueObject{}
	wci := &WrappedClassInstance{Instance: inst}
	handle := cgo.NewHandle(wci)
	defer handle.Delete()

	var w int64
	// Re-witness: the installed registry callbacks must feed THIS counter.
	// Rebuild the registration with a counter-carrying handler instead of the
	// bench stub's sink handler.
	registry := Internal.GDRegisteredGDClasses
	ci, ok := registry.Get("SpotStub")
	if !ok {
		t.Fatal("SpotStub missing from the gate registry")
	}
	ci.VirtualMethodMap["_notification"].GoMethodMetadata.cbNotification =
		newNotificationShape[benchPrologueObject, *benchPrologueObject](
			func(self *benchPrologueObject, what int32, reversed bool) { w++ })
	ci.VirtualMethodMap["_get"].GoMethodMetadata.cbPropGet =
		newPropGetShape[benchPrologueObject, *benchPrologueObject](
			func(self *benchPrologueObject, name string) (Variant, bool) { w++; return Variant{}, true })

	budgetGate(t, "prologue notification resolve-and-dispatch", 0, func() {
		wp := handle.Value().(*WrappedClassInstance)
		c, ok := Internal.GDRegisteredGDClasses.Get(wp.Instance.GetClassName())
		if !ok {
			t.Fatal("SpotStub not resolvable")
		}
		cb := c.VirtualMethodMap["_notification"].GoMethodMetadata.cbNotification
		cb(wp.Instance, 16, false)
	}, &w)

	budgetGate(t, "prologue prop-get resolve-and-dispatch", 0, func() {
		wp := handle.Value().(*WrappedClassInstance)
		c, ok := Internal.GDRegisteredGDClasses.Get(wp.Instance.GetClassName())
		if !ok {
			t.Fatal("SpotStub not resolvable")
		}
		cb := c.VirtualMethodMap["_get"].GoMethodMetadata.cbPropGet
		cb(wp.Instance, "dproperty_1")
	}, &w)
}

// TestDispatchBudgetGateReflectiveControlFloor is the falsification partner:
// the name-string reflective body must keep measuring allocations (floor 16,
// baseline 32 at arity 1 per typed_dispatch_budget_test.go — the floor sits
// low so Go-version drift cannot make the control vacuous). If this fails,
// every zero-budget row above has lost its proof that the gate can observe a
// regression at all.
func TestDispatchBudgetGateReflectiveControlFloor(t *testing.T) {
	InitPrimativeTypeEncodersForUnitTests()

	stub := &spotStubClass{}
	var cell [8]byte
	binary.LittleEndian.PutUint64(cell[:], 7)
	args1 := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(&cell[0])}
	var ret [8]byte

	var w int64
	md := &GoMethodMetadata{
		ClassName:        "SpotStub",
		GdMethodName:     "gate_triple",
		GoMethodName:     "GateTriple",
		Func:             reflect.ValueOf(func(s *spotStubClass, n int64) int64 { w++; return n }),
		GoReturnType:     reflect.TypeFor[int64](),
		GoReturnStyle:    ValueReturnStyle,
		GoArgumentTypes:  []reflect.Type{reflect.TypeFor[int64]()},
		gdeArgumentTypes: []GDExtensionVariantType{GDEXTENSION_VARIANT_TYPE_INT},
	}

	const floor = 16
	const runs = 200
	n := testing.AllocsPerRun(runs, func() {
		md.Ptrcall(stub, args1, GDExtensionUninitializedTypePtr(&ret[0]))
	})
	t.Logf("dispatch budget %-38s measured=%g floor=%d handler_calls=%d",
		"reflective control arity-1", n, floor, w)
	if w < runs {
		t.Errorf("reflective control: handler ran %d times across %d calls — vacuous control", w, runs)
	}
	if n < float64(floor) {
		t.Errorf("reflective control allocates %g per call, floor %d: the control can no "+
			"longer prove the zero-budget shapes are falsifiable", n, floor)
	}
}
