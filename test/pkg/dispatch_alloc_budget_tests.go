package pkg

import (
	"fmt"
	"testing"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// In-engine allocation-budget test for the compiled dispatch paths
// (typed-method-dispatch task 7.2).
//
// The engine-free tier (pkg/core/typed_dispatch_budget_test.go) measures the
// ptrcall wrapper bodies on raw byte cells at a zero-allocation cap. What
// cannot run without a live engine — and what this file measures — is
// everything whose closure body crosses into the engine through the published
// function-pointer table:
//
//   - varcall execution (Variant decoders and return encoders), including the
//     copy-return echo family with a real builtin,
//   - dispatch through classes actually registered in the engine, resolved
//     via the DispatchFor observation seam,
//   - the static and virtual bind shapes (static_add, alloc_probe_zero).
//
// The probe handlers are allocation-neutral by construction: their bodies do
// no engine work and allocate nothing, so every counted allocation belongs to
// the dispatch path plus the FFI boundary it crosses. That boundary — not
// dispatch — sets the varcall allowances: each Variant decode/encode leg pins
// through a fresh runtime.Pinner inside pkg/builtin (measured ~2 per scalar
// leg) and the shared trailing-default fill adds one slice. A copy-return
// pays the copy-constructor encode and the classified release. The ptrcall
// path crosses no boundary for value-family returns, so its cap stays zero.
//
// What the caps cannot express — "no reflect machinery came back" — the
// reflective twins assert: the same signatures through the name-string API
// measure 8-13 ptrcall / 16-37 varcall allocations per call against the
// compiled 0 / 3-14 (before the lazy dispatch payloads of
// trim-dispatch-overhead: 27-56 / 38-249 — the eager diagnostic strings were
// the bulk of the reflective cost), so the margin check (reflective ≥
// reflectMargin × compiled, plus an absolute floor) fails loudly if a
// compiled path is ever swapped back to Func.Call.
//
// Varcall probes destroy each measured return Variant, and the copy-return
// ptrcall probe destroys each encoded cell: the encoding transfers an engine
// reference to whoever owns the cell, so the probe must release it or the
// shutdown leak gates would book the probe itself as a leak.
//
// Every cell is measured and printed even after a breach so one engine run
// reports the whole matrix. Return codes follow the existing convention: 1 on
// success, negative per failed assertion, -900 on panic.

const (
	// budgetRuns per cell: enough for AllocsPerRun to average away noise,
	// small enough that the whole matrix runs in well under a second.
	budgetRuns = 100

	// capPtrcall: zero — the compiled direct-call contract for
	// allocation-neutral handlers. Value-family ptrcall crosses no engine
	// boundary; a container argument is a raw handle load.
	capPtrcall = 0

	// capVarcallNullary: measured 3 (fill slice + return-encode pins).
	capVarcallNullary = 5
	// capVarcallScalar: measured 8 at arity 2, 3 at arity 0.
	capVarcallScalar = 12
	// capVarcallWide: measured 14 at arity 5.
	capVarcallWide = 20
	// capVarcallContainerArg: the decode copy-construct plus release on top
	// of the scalar baseline.
	capVarcallContainerArg = 12
	// capCopyRetPtrcall: measured 3 — the copy-ctor encode and the release
	// gate; no decode legs (arity-1 echo borrows its argument cell).
	capCopyRetPtrcall = 5
	// capCopyRetVarcall: measured 7.
	capCopyRetVarcall = 11

	// reflectFloor: the reflective control must reach at least this many
	// allocations per call (measured engine-free: 32 at arity 1).
	reflectFloor = 4

	// reflectMargin: the reflective twin must measure at least this many
	// times the compiled cell on the same signature. Reflection costs dozens
	// of boxes per call, so the margin catches any reintroduction while
	// staying far from the measurement noise floor.
	reflectMargin = 2
)

// budgetCell is one measured dispatch path.
type budgetCell struct {
	label string
	cap   float64
	code  int32
	run   func()
}

// runBudgetCells measures every cell, prints the whole matrix, and reports
// the first breach code (0 when all cells are within budget).
func runBudgetCells(runs int, cells []budgetCell) (int32, map[string]float64) {
	var breach int32
	measured := make(map[string]float64, len(cells))
	for _, c := range cells {
		n := testing.AllocsPerRun(runs, c.run)
		measured[c.label] = n
		if n > c.cap {
			fmt.Printf("  dispatch_budget: FAIL: %s allocates %v per call, budget %v\n", c.label, n, c.cap)
			if breach == 0 {
				breach = c.code
			}
		} else {
			fmt.Printf("  dispatch_budget: %s: %v allocs/call (budget %v)\n", c.label, n, c.cap)
		}
	}
	return breach, measured
}

// TestCompiledDispatchAllocs runs the compiled-dispatch allocation matrix
// through the registered demo bindings:
//
//	shape                      | style    | arity | cap
//	---------------------------+----------+-------+--------------------
//	static_add                 | value    | 2     | p 0 / v scalar
//	alloc_probe_sum0           | value    | 0     | p 0 / v nullary
//	alloc_probe_sum2           | value    | 2     | p 0 / v scalar
//	alloc_probe_sum5           | value    | 5     | p 0 / v wide
//	alloc_probe_zero (virtual) | value    | 0     | p 0 / v nullary
//	alloc_probe_array_arg      | value    | 1     | p 0 / v container
//	alloc_probe_echo_sn        | copy-ret | 1     | p copy / v copy
//
// plus the name-string reflective twins of the arity 0/2/5 value family,
// which must measure above reflectFloor and by at least reflectMargin over
// their compiled twins.
func (e *Example) TestCompiledDispatchAllocs() int32 {
	return objectArgGuard("dispatch_budget", func() int32 {
		// --- Resolve and verify the compiled binds first: a silent
		// fallback to the reflective path would otherwise measure as a
		// mere budget breach.
		probes := []struct {
			gdName string
			code   int32
		}{
			{"static_add", -10},
			{"alloc_probe_sum0", -11},
			{"alloc_probe_sum2", -12},
			{"alloc_probe_sum5", -13},
			{"alloc_probe_zero", -14},
			{"alloc_probe_array_arg", -15},
			{"alloc_probe_echo_sn", -16},
			{"alloc_probe_sum0_reflective", -17},
			{"alloc_probe_sum2_reflective", -18},
			{"alloc_probe_sum5_reflective", -19},
		}
		dispatchers := make(map[string]MethodDispatcher, len(probes))
		for _, p := range probes {
			d, found := DispatchFor("Example", p.gdName)
			if !found {
				fmt.Printf("  dispatch_budget: FAIL: %s not registered\n", p.gdName)
				return p.code
			}
			dispatchers[p.gdName] = d
		}
		for _, p := range probes[:7] {
			if !dispatchers[p.gdName].Compiled() {
				fmt.Printf("  dispatch_budget: FAIL: %s is not fully compiled\n", p.gdName)
				return p.code
			}
		}
		for _, p := range probes[7:] {
			if dispatchers[p.gdName].Compiled() {
				fmt.Printf("  dispatch_budget: FAIL: %s unexpectedly compiled\n", p.gdName)
				return p.code
			}
		}

		// --- Argument cells, built once outside every measured loop.
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

		sn := NewStringNameWithLatin1Chars("alloc_budget_name")
		defer sn.Destroy()
		snCells := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(sn.NativeConstPtr())}
		snVariant := NewVariantStringName(sn)
		defer snVariant.Destroy()
		snVarArgs := []Variant{snVariant}

		// The copy-return encode hands the cell a fresh engine reference; a
		// real engine caller destroys it after reading, so the probe destroys
		// it each iteration.
		var snRet StringName
		snRetPtr := GDExtensionUninitializedTypePtr(unsafe.Pointer(&snRet))

		compiled := func(name string) MethodDispatcher { return dispatchers[name] }
		varCall := func(d MethodDispatcher, inst GDClass, args []Variant) func() {
			return func() {
				ret, err := d.Varcall(inst, args)
				if err != nil {
					panic(err)
				}
				ret.Destroy()
			}
		}

		cells := []budgetCell{
			{"static_add ptrcall", capPtrcall, -30, func() {
				compiled("static_add").Ptrcall(nil, twoCells, retPtr)
			}},
			{"static_add varcall", capVarcallScalar, -31,
				varCall(compiled("static_add"), nil, twoVariants)},
			{"alloc_probe_sum0 ptrcall", capPtrcall, -32, func() {
				compiled("alloc_probe_sum0").Ptrcall(e, emptyCells, retPtr)
			}},
			{"alloc_probe_sum0 varcall", capVarcallNullary, -33,
				varCall(compiled("alloc_probe_sum0"), e, emptyVariants)},
			{"alloc_probe_sum2 ptrcall", capPtrcall, -34, func() {
				compiled("alloc_probe_sum2").Ptrcall(e, twoCells, retPtr)
			}},
			{"alloc_probe_sum2 varcall", capVarcallScalar, -35,
				varCall(compiled("alloc_probe_sum2"), e, twoVariants)},
			{"alloc_probe_sum5 ptrcall", capPtrcall, -36, func() {
				compiled("alloc_probe_sum5").Ptrcall(e, fiveCells, retPtr)
			}},
			{"alloc_probe_sum5 varcall", capVarcallWide, -37,
				varCall(compiled("alloc_probe_sum5"), e, fiveVariants)},
			{"alloc_probe_zero ptrcall (virtual)", capPtrcall, -38, func() {
				compiled("alloc_probe_zero").Ptrcall(e, emptyCells, retPtr)
			}},
			{"alloc_probe_zero varcall (virtual)", capVarcallNullary, -39,
				varCall(compiled("alloc_probe_zero"), e, emptyVariants)},
			{"alloc_probe_array_arg ptrcall", capPtrcall, -40, func() {
				compiled("alloc_probe_array_arg").Ptrcall(e, arrCells, retPtr)
			}},
			{"alloc_probe_array_arg varcall", capVarcallContainerArg, -41,
				varCall(compiled("alloc_probe_array_arg"), e, arrVarArgs)},
			{"alloc_probe_echo_sn ptrcall", capCopyRetPtrcall, -42, func() {
				compiled("alloc_probe_echo_sn").Ptrcall(e, snCells, snRetPtr)
				snRet.Destroy()
			}},
			{"alloc_probe_echo_sn varcall", capCopyRetVarcall, -43,
				varCall(compiled("alloc_probe_echo_sn"), e, snVarArgs)},
		}

		breach, measured := runBudgetCells(budgetRuns, cells)

		// --- Reflective twins: the same signatures through the name-string
		// API. They stay uncompiled (checked above) and must measure above
		// the floor and by the margin over their compiled twins — the
		// falsification control proving the caps can detect reflect
		// machinery on a dispatch path.
		reflPairs := []struct {
			gdName     string
			compiledAt string
			cells      []GDExtensionConstTypePtr
			varArgs    []Variant
			code       int32
		}{
			{"alloc_probe_sum0_reflective", "alloc_probe_sum0 varcall", emptyCells, emptyVariants, -50},
			{"alloc_probe_sum2_reflective", "alloc_probe_sum2 varcall", twoCells, twoVariants, -51},
			{"alloc_probe_sum5_reflective", "alloc_probe_sum5 varcall", fiveCells, fiveVariants, -52},
		}
		for _, pair := range reflPairs {
			refl := dispatchers[pair.gdName]
			nPtr := testing.AllocsPerRun(budgetRuns, func() {
				refl.Ptrcall(e, pair.cells, retPtr)
			})
			nVar := testing.AllocsPerRun(budgetRuns, func() {
				ret, err := refl.Varcall(e, pair.varArgs)
				if err != nil {
					panic(err)
				}
				ret.Destroy()
			})
			fmt.Printf("  dispatch_budget: %s: %v ptrcall / %v varcall allocs/call (floor %v, margin x%v)\n",
				pair.gdName, nPtr, nVar, reflectFloor, reflectMargin)

			compiledVar := measured[pair.compiledAt]
			if nPtr < reflectFloor || nVar < reflectFloor {
				fmt.Printf("  dispatch_budget: FAIL: %s below reflect floor — the control cannot detect regression\n", pair.gdName)
				if breach == 0 {
					breach = pair.code
				}
				continue
			}
			if nVar < compiledVar*reflectMargin {
				fmt.Printf("  dispatch_budget: FAIL: %s varcall %v not >= %vx compiled %v — the caps cannot detect regression\n",
					pair.gdName, nVar, reflectMargin, compiledVar)
				if breach == 0 {
					breach = pair.code
				}
			}
		}

		if breach != 0 {
			return breach
		}
		return 1
	})
}
