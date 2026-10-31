package pkg

/*
#include <godot/gdextension_interface.h>

// dispatchDiagCountingStringify is a //exported Go function below; cgo gives
// it a C symbol of the same name. This helper hands that symbol back as a
// function pointer with the engine's typedef so the fixture can install it
// into the loaded interface-function table.
extern void dispatchDiagCountingStringify(GDExtensionConstVariantPtr p_self, GDExtensionStringPtr r_ret);

static GDExtensionInterfaceVariantStringify dispatchDiagStringifyPtr(void) {
	return (GDExtensionInterfaceVariantStringify)(&dispatchDiagCountingStringify);
}
*/
import "C"

import (
	"bytes"
	"fmt"
	"strings"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	"github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
)

// In-engine diagnostics-cost fixture (trim-dispatch-overhead tasks 2.2, 2.3).
//
// The dispatch entries ("Call", "Call Variadic", "Ptrcall", ...) describe a
// call's arguments by asking the ENGINE to stringify each Variant. When the
// entry threshold discards the entry, that work — and every engine crossing
// it makes — is pure loss. This fixture counts the crossings directly: it
// swaps the variant_stringify entry of the loaded interface-function table
// for a Go counter (a cgo-exported function installed as the table pointer),
// dispatches the reflective fallback varcalls, and restores the real pointer.
// Not a log grep: the counter is the measurement.
//
// Safety of the swap window: within this binding the table's
// variant_stringify is read only by Variant.Stringify, which is diagnostic
// formatting only (the engine stringifies its own values through its own C++
// methods, never through this extension's table). The swap lives only across
// the fixture's own dispatch calls on the single-threaded main loop, and
// defer-restores even on panic. The stub never writes the return String, so
// stringified values read as "" while it is installed — the count never
// depends on the value.
//
// Thresholds are forced by the fixture (log.SetLevel), never inherited from
// the environment, so the measurement means the same thing under the suite's
// default warning threshold and under a raised-threshold run alike.
//
// Records printed, then asserted:
//
//	dispatch_diag: threshold=suppress shape=<name> arity=<n> crossings=<n>  -> all zero, flat across arities
//	dispatch_diag: threshold=raised  shape=<name> arity=<n> crossings=<n>  -> >0 when arity>0 (lazy, not deleted)
//
// The capture section then points the log at a buffer, raises the threshold,
// and drives a two-argument reflective varcall ("Call" entry: must still
// carry bind / gd_args / ret identifying the call) and a two-argument
// variadic call ("Call Variadic" entry: the argument collection must appear
// exactly once with real values — the duplicate resolved_args stringification
// is gone outright — and the "Call" entry's resolved_args must record the
// identity "same as gd_args" when the fill changed nothing).
//
// Return codes follow the existing convention: 1 success, negative per failed
// assertion, -900 on panic. Every figure is printed even after a breach so
// one run reports the whole matrix.

// dispatchDiagStringifyCounts counts variant_stringify crossings while the
// stub is installed.
//
//export dispatchDiagCountingStringify
func dispatchDiagCountingStringify(self C.GDExtensionConstVariantPtr, ret C.GDExtensionStringPtr) {
	dispatchDiagStringifyCount++
	// Deliberately does not write r_ret: stringified text reads as empty
	// while the stub is installed, acceptable because the fixture asserts
	// only on the count inside the swap window.
}

var dispatchDiagStringifyCount int

// TestDispatchDiagnostics measures what a discarded dispatch entry costs and
// what a written entry must still carry.
func (e *Example) TestDispatchDiagnostics() int32 {
	return objectArgGuard("dispatch_diag", func() int32 {
		// --- Resolve the reflective fallback varcalls (name-string binds).
		probes := []struct {
			gdName string
			arity  int
			code   int32
		}{
			{"alloc_probe_sum0_reflective", 0, -20},
			{"alloc_probe_sum2_reflective", 2, -21},
			{"alloc_probe_sum5_reflective", 5, -22},
		}
		dispatchers := make(map[string]MethodDispatcher, len(probes))
		variants := make(map[string][]Variant, len(probes))
		for _, p := range probes {
			d, found := DispatchFor("Example", p.gdName)
			if !found {
				fmt.Printf("  dispatch_diag: FAIL: %s not registered\n", p.gdName)
				return p.code
			}
			if d.Compiled() {
				fmt.Printf("  dispatch_diag: FAIL: %s unexpectedly compiled\n", p.gdName)
				return p.code
			}
			dispatchers[p.gdName] = d
			var vs []Variant
			for i := 0; i < p.arity; i++ {
				vs = append(vs, NewVariantInt64(int64(i+1)))
			}
			variants[p.gdName] = vs
		}
		defer func() {
			for _, vs := range variants {
				for _, v := range vs {
					v.Destroy()
				}
			}
		}()

		varargs, found := DispatchFor("Example", "varargs_func")
		if !found {
			fmt.Printf("  dispatch_diag: FAIL: varargs_func not registered\n")
			return -27
		}

		callOnce := func(d MethodDispatcher, args []Variant) {
			ret, err := d.Varcall(e, args)
			if err != nil {
				panic(fmt.Sprintf("dispatch_diag: varcall failed: %v", err))
			}
			ret.Destroy()
		}

		previousLevel := log.GetLevel()
		var breach int32
		failWith := func(code int32, format string, a ...any) {
			fmt.Printf("  dispatch_diag: FAIL: "+format+"\n", a...)
			if breach == 0 {
				breach = code
			}
		}

		// --- Suppressed threshold: warning discards the info entries, so
		// no payload may touch the engine. Crossings must be zero at every
		// arity, hence flat: "0, 4, 10" style growth with arity is the
		// eager-payload signature this change removes.
		realStringifySave := ffi.FFI.VariantStringify
		ffi.FFI.VariantStringify = ffi.GDExtensionInterfaceVariantStringify(C.dispatchDiagStringifyPtr())
		defer func() { ffi.FFI.VariantStringify = realStringifySave }()
		log.SetLevel(log.WarnLevel)
		for _, p := range probes {
			dispatchDiagStringifyCount = 0
			callOnce(dispatchers[p.gdName], variants[p.gdName])
			n := dispatchDiagStringifyCount
			fmt.Printf("  dispatch_diag: threshold=suppress shape=%s arity=%d crossings=%d\n", p.gdName, p.arity, n)
			if n != 0 {
				failWith(p.code, "%s asked the engine to stringify %d value(s) of a discarded call", p.gdName, n)
			}
		}

		// --- Raised threshold: the entries are written, so lazy payloads
		// marshal and crossings happen again. Zero at arity>0 would mean the
		// entries were deleted rather than made lazy.
		log.SetLevel(log.InfoLevel)
		for _, p := range probes {
			dispatchDiagStringifyCount = 0
			callOnce(dispatchers[p.gdName], variants[p.gdName])
			n := dispatchDiagStringifyCount
			fmt.Printf("  dispatch_diag: threshold=raised shape=%s arity=%d crossings=%d\n", p.gdName, p.arity, n)
			if p.arity > 0 && n == 0 {
				failWith(p.code-100, "%s described no arguments at a permitting threshold", p.gdName)
			}
		}

		// --- Capture section: field names and content of the written
		// entries, with the real stringify restored so values are real.
		ffi.FFI.VariantStringify = realStringifySave
		var buf bytes.Buffer
		prevWriter := log.SetWriteSyncer(&buf)
		log.SetLevel(log.InfoLevel)
		callOnce(dispatchers["alloc_probe_sum2_reflective"], variants["alloc_probe_sum2_reflective"])
		a := NewVariantInt64(2)
		b := NewVariantInt64(3)
		callOnce(varargs, []Variant{a, b})
		a.Destroy()
		b.Destroy()
		log.SetLevel(previousLevel)
		log.SetWriteSyncer(prevWriter)

		captured := buf.String()
		for _, line := range strings.Split(captured, "\n") {
			if strings.Contains(line, "AllocProbeSum2Reflective") || strings.Contains(line, "Variadic") {
				fmt.Printf("  dispatch_diag: captured | %s\n", line)
			}
		}
		if !strings.Contains(captured, "bind") || !strings.Contains(captured, "gd_args") ||
			!strings.Contains(captured, "ret") {
			failWith(-24, "captured entry lost a field name (bind/gd_args/ret)")
		}
		// Per-entry content, asserted on the written rendering (spaces
		// flattened: the console encoder pads after colons).
		seenCall, seenVariadic := false, false
		for _, line := range strings.Split(captured, "\n") {
			flat := strings.ReplaceAll(line, " ", "")
			switch {
			case strings.Contains(line, "Call Variadic"):
				seenVariadic = true
				// The collected Variants go to the handler unchanged, so the
				// duplicate resolved_args field is gone outright...
				if strings.Contains(flat, "resolved_args") {
					failWith(-26, "variadic entry still carries the duplicate resolved_args field")
				}
				// ...and the collection is described exactly once (one
				// "values" array carrying both stringified arguments).
				if c := strings.Count(flat, `"values":[`); c != 1 {
					failWith(-25, "variadic entry rendered the argument collection %d times (want 1)", c)
				}
				if !strings.Contains(flat, `"2"`) || !strings.Contains(flat, `"3"`) {
					failWith(-25, "variadic entry lost the stringified argument values: %s", line)
				}
			case strings.Contains(flat, `"method":"AllocProbeSum2Reflective"`):
				seenCall = true
				// The trailing-default fill copies the supplied cells
				// unchanged, so resolved_args records the identity instead of
				// re-describing every value a second time in one entry.
				if !strings.Contains(line, `"resolution": "same as gd_args"`) {
					failWith(-27, "fully-supplied call re-described its arguments instead of recording the identity: %s", line)
				}
				if !strings.Contains(flat, `"1"`) || !strings.Contains(flat, `"2"`) {
					failWith(-27, "Call entry lost the stringified argument values: %s", line)
				}
			}
		}
		if !seenVariadic {
			failWith(-28, "capture saw no \"Call Variadic\" entry")
		}
		if !seenCall {
			failWith(-28, "capture saw no \"Call\" entry for the reflective probe")
		}

		if breach != 0 {
			return breach
		}
		return 1
	})
}
