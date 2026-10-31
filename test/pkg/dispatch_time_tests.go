package pkg

import (
	"fmt"
	"time"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// In-engine dispatch-timing fixture (trim-dispatch-overhead task 1.4).
//
// The dispatch-cost-baseline capability records per-shape wall-clock cost with
// the engine running: the engine-free tier (pkg/core/dispatch_cost_bench_test.go)
// cannot reach the FFI crossing, the Variant decoders/encoders, or the class
// callback prologues. This fixture prints one machine-readable line per shape:
//
//	dispatch_time: shape=<name> ns/call=<figure>
//
// and the output gate requires the PRESENCE of every shape's line — never a
// value (wall-clock figures are not gateable; the allocation budgets in
// dispatch_alloc_budget_tests.go and the engine-free gate carry that role).
//
// Shapes covered (the set check_test_output.sh pins — keep both in step):
//
//	static_add_ptrcall / static_add_varcall
//	value_arity0_ptrcall / value_arity0_varcall
//	value_arity2_ptrcall / value_arity2_varcall
//	value_arity5_ptrcall / value_arity5_varcall
//	virtual_arity0_ptrcall / virtual_arity0_varcall
//	container_arg_ptrcall / container_arg_varcall
//	copy_ret_ptrcall / copy_ret_varcall
//	reflective_value_arity0_ptrcall / reflective_value_arity0_varcall
//	reflective_value_arity2_ptrcall / reflective_value_arity2_varcall
//	reflective_value_arity5_ptrcall / reflective_value_arity5_varcall
//	prologue_notification / prologue_property_get
//
// The probe bindings are the same registered typed name strings the allocation
// budget fixture measures (alloc_probe_* and static_add); a bind that silently
// fell back to the reflective path would change what a timing record means, so
// the Compiled() report is verified before any timing line is printed.
//
// The two prologue cells ride live paths: a notification delivered through
// Object::notification lands on the notification prologue, and a property read
// of an unlisted "dproperty" lands on the Get prologue via the _get virtual.
// They therefore include the engine round trip around the prologue — constant
// before and after the resolution hoist, so the records still isolate the
// prologue delta. Example's notification handler records every code; this
// fixture restores the slice so downstream assertions are unaffected.
//
// Return codes follow the existing convention: 1 on success, negative per
// failed precondition, -900 on panic (objectArgGuard).

const (
	// timingRuns: enough iterations that the figure averages scheduler
	// noise, short enough that the whole fixture runs well under a second
	// per the demo's frame budget.
	timingRunsPtrcall    = 100000
	timingRunsVarcall    = 50000
	timingRunsReflective = 10000
	timingRunsPrologue   = 20000

	// timingWarmupFraction is the share of runs discarded before measuring,
	// absorbing lazy init (icache, method-bind cache) outside the figure.
	timingWarmupFraction = 8
)

// timeDispatchShape times fn for runs iterations and prints the machine-
// readable record. Nothing here compares the figure against anything: the
// line exists to be present, read, and diffed across runs.
func timeDispatchShape(shape string, runs int, fn func()) {
	warmup := runs / timingWarmupFraction
	if warmup < 64 {
		warmup = 64
	}
	for i := 0; i < warmup; i++ {
		fn()
	}
	start := time.Now()
	for i := 0; i < runs; i++ {
		fn()
	}
	elapsed := time.Since(start)
	fmt.Printf("dispatch_time: shape=%s ns/call=%.1f\n", shape,
		float64(elapsed.Nanoseconds())/float64(runs))
}

// TestDispatchTimes records the in-engine per-shape dispatch timing baseline.
func (e *Example) TestDispatchTimes() int32 {
	return objectArgGuard("dispatch_time", func() int32 {
		// --- Resolve the probe binds and verify compiled/reflective status
		// first: timings over a silently-fallen-back path would mislabel the
		// record rather than fail, so a mismatch aborts the fixture.
		compiledProbes := []struct {
			gdName string
			code   int32
		}{
			{"static_add", -70},
			{"alloc_probe_sum0", -71},
			{"alloc_probe_sum2", -72},
			{"alloc_probe_sum5", -73},
			{"alloc_probe_zero", -74},
			{"alloc_probe_array_arg", -75},
			{"alloc_probe_echo_sn", -76},
		}
		reflectiveProbes := []struct {
			gdName string
			code   int32
		}{
			{"alloc_probe_sum0_reflective", -77},
			{"alloc_probe_sum2_reflective", -78},
			{"alloc_probe_sum5_reflective", -79},
		}
		dispatchers := make(map[string]MethodDispatcher, len(compiledProbes)+len(reflectiveProbes))
		resolve := func(probes []struct {
			gdName string
			code   int32
		}) int32 {
			for _, p := range probes {
				d, found := DispatchFor("Example", p.gdName)
				if !found {
					fmt.Printf("  dispatch_time: FAIL: %s not registered\n", p.gdName)
					return p.code
				}
				dispatchers[p.gdName] = d
			}
			return 0
		}
		if code := resolve(compiledProbes); code != 0 {
			return code
		}
		if code := resolve(reflectiveProbes); code != 0 {
			return code
		}
		for _, p := range compiledProbes {
			if !dispatchers[p.gdName].Compiled() {
				fmt.Printf("  dispatch_time: FAIL: %s is not compiled\n", p.gdName)
				return p.code
			}
		}
		for _, p := range reflectiveProbes {
			if dispatchers[p.gdName].Compiled() {
				fmt.Printf("  dispatch_time: FAIL: %s unexpectedly compiled\n", p.gdName)
				return p.code
			}
		}

		// --- Argument cells and Variants, built once outside every loop.
		var argA, argB int64 = 2, 3
		twoCells := []GDExtensionConstTypePtr{
			(GDExtensionConstTypePtr)(&argA),
			(GDExtensionConstTypePtr)(&argB),
		}
		var args5 [5]int64
		for i := range args5 {
			args5[i] = int64(i + 1)
		}
		fiveCells := []GDExtensionConstTypePtr{
			(GDExtensionConstTypePtr)(&args5[0]),
			(GDExtensionConstTypePtr)(&args5[1]),
			(GDExtensionConstTypePtr)(&args5[2]),
			(GDExtensionConstTypePtr)(&args5[3]),
			(GDExtensionConstTypePtr)(&args5[4]),
		}
		emptyCells := []GDExtensionConstTypePtr{}
		var retCell [8]byte
		retPtr := GDExtensionUninitializedTypePtr(&retCell[0])

		twoVariants := []Variant{NewVariantInt64(argA), NewVariantInt64(argB)}
		defer func() {
			for _, v := range twoVariants {
				v.Destroy()
			}
		}()
		var fiveVariants []Variant
		for i := range args5 {
			fiveVariants = append(fiveVariants, NewVariantInt64(args5[i]))
		}
		defer func() {
			for _, v := range fiveVariants {
				v.Destroy()
			}
		}()
		emptyVariants := []Variant{}

		arr := NewPackedInt64Array()
		arr.Append(1)
		arr.Append(2)
		defer arr.Destroy()
		arrCells := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(arr.NativeConstPtr())}
		arrVariant := NewVariantPackedInt64Array(arr)
		defer arrVariant.Destroy()
		arrVarArgs := []Variant{arrVariant}

		sn := NewStringNameWithLatin1Chars("dispatch_time_name")
		defer sn.Destroy()
		snCells := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(sn.NativeConstPtr())}
		snVariant := NewVariantStringName(sn)
		defer snVariant.Destroy()
		snVarArgs := []Variant{snVariant}

		var snRet StringName
		snRetPtr := GDExtensionUninitializedTypePtr(unsafe.Pointer(&snRet))

		varCall := func(name string, inst GDClass, args []Variant) func() {
			d := dispatchers[name]
			return func() {
				ret, err := d.Varcall(inst, args)
				if err != nil {
					panic(err)
				}
				ret.Destroy()
			}
		}
		ptrCall := func(name string, inst GDClass, cells []GDExtensionConstTypePtr, ret GDExtensionUninitializedTypePtr) func() {
			d := dispatchers[name]
			return func() { d.Ptrcall(inst, cells, ret) }
		}

		// --- Compiled shapes.
		timeDispatchShape("static_add_ptrcall", timingRunsPtrcall,
			ptrCall("static_add", nil, twoCells, retPtr))
		timeDispatchShape("static_add_varcall", timingRunsVarcall,
			varCall("static_add", nil, twoVariants))
		timeDispatchShape("value_arity0_ptrcall", timingRunsPtrcall,
			ptrCall("alloc_probe_sum0", e, emptyCells, retPtr))
		timeDispatchShape("value_arity0_varcall", timingRunsVarcall,
			varCall("alloc_probe_sum0", e, emptyVariants))
		timeDispatchShape("value_arity2_ptrcall", timingRunsPtrcall,
			ptrCall("alloc_probe_sum2", e, twoCells, retPtr))
		timeDispatchShape("value_arity2_varcall", timingRunsVarcall,
			varCall("alloc_probe_sum2", e, twoVariants))
		timeDispatchShape("value_arity5_ptrcall", timingRunsPtrcall,
			ptrCall("alloc_probe_sum5", e, fiveCells, retPtr))
		timeDispatchShape("value_arity5_varcall", timingRunsVarcall,
			varCall("alloc_probe_sum5", e, fiveVariants))
		timeDispatchShape("virtual_arity0_ptrcall", timingRunsPtrcall,
			ptrCall("alloc_probe_zero", e, emptyCells, retPtr))
		timeDispatchShape("virtual_arity0_varcall", timingRunsVarcall,
			varCall("alloc_probe_zero", e, emptyVariants))
		timeDispatchShape("container_arg_ptrcall", timingRunsPtrcall,
			ptrCall("alloc_probe_array_arg", e, arrCells, retPtr))
		timeDispatchShape("container_arg_varcall", timingRunsVarcall,
			varCall("alloc_probe_array_arg", e, arrVarArgs))
		timeDispatchShape("copy_ret_ptrcall", timingRunsVarcall, func() {
			dispatchers["alloc_probe_echo_sn"].Ptrcall(e, snCells, snRetPtr)
			snRet.Destroy()
		})
		timeDispatchShape("copy_ret_varcall", timingRunsVarcall,
			varCall("alloc_probe_echo_sn", e, snVarArgs))

		// --- Reflective twins (name-string path): the record the fallback
		// change (compile-name-string-binds) will compare against.
		reflPairs := []struct {
			gdName  string
			shapeP  string
			shapeV  string
			cells   []GDExtensionConstTypePtr
			varArgs []Variant
		}{
			{"alloc_probe_sum0_reflective", "reflective_value_arity0_ptrcall",
				"reflective_value_arity0_varcall", emptyCells, emptyVariants},
			{"alloc_probe_sum2_reflective", "reflective_value_arity2_ptrcall",
				"reflective_value_arity2_varcall", twoCells, twoVariants},
			{"alloc_probe_sum5_reflective", "reflective_value_arity5_ptrcall",
				"reflective_value_arity5_varcall", fiveCells, fiveVariants},
		}
		for _, pair := range reflPairs {
			timeDispatchShape(pair.shapeP, timingRunsReflective,
				ptrCall(pair.gdName, e, pair.cells, retPtr))
			timeDispatchShape(pair.shapeV, timingRunsReflective,
				varCall(pair.gdName, e, pair.varArgs))
		}

		// --- Class callback prologues (live paths). The notification handler
		// appends to Example's recorded codes; restore them so downstream
		// notification assertions see the pre-fixture state.
		savedCodes := e.notificationCodes
		timeDispatchShape("prologue_notification", timingRunsPrologue, func() {
			e.Notification(1000, false)
		})
		e.notificationCodes = savedCodes

		propName := NewStringNameWithLatin1Chars("dproperty_1")
		defer propName.Destroy()
		timeDispatchShape("prologue_property_get", timingRunsPrologue, func() {
			v := e.Get(propName)
			v.Destroy()
		})

		return 1
	})
}
