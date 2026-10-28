package core

import (
	"encoding/binary"
	"reflect"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

// typed-method-dispatch task 5.2 unit tier (engine-free).
//
// Borrow-echo classification is a pure content comparison: the reflective
// path runs it with reflect.DeepEqual over the dynamic wrapper values, the
// typed path with the bind-time byte-compare closures from ret_release.gen.go.
// Builtins are fixed-size byte arrays and both classifiers reduce to element-
// wise equality, so every differential case below — echo, non-echo, COW
// shared-storage clone, StringName intern equality, receiver exclusion,
// cross-type pairs — runs without a live engine and asserts identical
// classification on both paths.
//
// The release side of the matrix (which path destroys a classified-owned
// source) is pinned structurally: ptrcall copy-ret wrappers always hand the
// return to RetDestroy when not an echo (asserted through the counting stub in
// typed_wrapper_test.go), newVarcallRetDestroyer encodes the varcall release
// set, and the variadic wrappers never consult classification. Refcount
// effects through a live engine arrive with the typed-bound demo examples
// (task 7.1), where the existing leak gates catch any drift.

// cellStringName plants an interned-pointer word, the way a decoded borrow
// carries the engine's interned StringName data pointer.
func cellStringName(p uint64) StringName {
	var s StringName
	binary.LittleEndian.PutUint64(s[:], p)
	return s
}

// cellStringNameText constructs a StringName independently from its cell
// bytes — two interned StringNames for the same text carry the same data
// pointer, which is exactly what this cell shape reproduces.
func cellStringNameText(b byte) StringName {
	var s StringName
	for i := range s {
		s[i] = b
	}
	return s
}

func cellNodePath(p uint64) NodePath {
	var n NodePath
	binary.LittleEndian.PutUint64(n[:], p)
	return n
}

func cellArray(p uint64) Array {
	var a Array
	binary.LittleEndian.PutUint64(a[:], p)
	return a
}

func cellPackedBytes(word0, word1 uint64) PackedByteArray {
	var p PackedByteArray
	binary.LittleEndian.PutUint64(p[:8], word0)
	binary.LittleEndian.PutUint64(p[8:], word1)
	return p
}

// typedEcho reproduces the classification the generated copy-ret wrappers
// perform: a bind-time type assertion yields the concrete byte-compare
// closure only for the matching type (the same same-type requirement the
// reflective classifier enforces with args[i].Type() == ret.Type()), and the
// wrapper ORs it over the decoded parameter slots. The receiver never joins —
// typed wrappers keep it in `self`, and the reflective path skips args[0] for
// instance binds.
func typedEcho[T RetCopyEncodable](ret T, params []any) bool {
	cmp, ok := newEchoCompare[T]().(func(T, T) bool)
	if !ok {
		return false
	}
	for _, p := range params {
		if a, ok := p.(T); ok && cmp(a, ret) {
			return true
		}
	}
	return false
}

// TestTypedEchoClassificationMatchesReflective drives the same fixtures
// through isPtrcallBorrowEcho (reflective) and typedEcho (typed wrappers).
func TestTypedEchoClassificationMatchesReflective(t *testing.T) {
	cases := []struct {
		name        string
		ret         any
		allArgs     []any // reflective slice, receiver included at index 0 when hasReceiver
		hasReceiver bool
		params      []any // typed slice, parameters only
		want        bool
	}{
		// --- StringName -------------------------------------------------
		{"stringname echo — borrowed argument returned verbatim",
			cellStringName(0x1000), []any{cellStringName(0x1000)}, false,
			[]any{cellStringName(0x1000)}, true},
		{"stringname distinct — freshly built value",
			cellStringName(0x1000), []any{cellStringName(0x2000)}, false,
			[]any{cellStringName(0x2000)}, false},
		{"stringname intern equality — same text, same interned pointer (accepted clone leak)",
			cellStringNameText(7), []any{cellStringNameText(7)}, false,
			[]any{cellStringNameText(7)}, true},
		{"stringname cross-type same bytes — NodePath argument cannot echo",
			cellStringName(0x1000), []any{cellNodePath(0x1000)}, false,
			[]any{cellNodePath(0x1000)}, false},
		{"stringname second slot echo",
			cellStringName(0x3000), []any{int64(5), cellStringName(0x3000)}, false,
			[]any{int64(5), cellStringName(0x3000)}, true},
		{"stringname receiver excluded from comparison (instance bind)",
			cellStringName(0x1000), []any{cellStringName(0x1000), cellStringName(0x2000)}, true,
			[]any{cellStringName(0x2000)}, false},

		// --- Array (COW container) ---------------------------------------
		{"array echo",
			cellArray(0xA000), []any{cellArray(0xA000)}, false,
			[]any{cellArray(0xA000)}, true},
		{"array COW clone unmutated — copy constructor shares storage (accepted clone leak)",
			cellArray(0xA000), []any{cellArray(0xA000)}, false,
			[]any{cellArray(0xA000)}, true},
		{"array mutated clone — storage pointer differs after CoW write",
			cellArray(0xB000), []any{cellArray(0xA000)}, false,
			[]any{cellArray(0xA000)}, false},
		{"array echo in second slot",
			cellArray(0xA000), []any{int64(1), cellArray(0xA000)}, false,
			[]any{int64(1), cellArray(0xA000)}, true},

		// --- PackedByteArray (two-word cell) -----------------------------
		{"packed echo — both words equal",
			cellPackedBytes(0xC000, 0x4), []any{cellPackedBytes(0xC000, 0x4)}, false,
			[]any{cellPackedBytes(0xC000, 0x4)}, true},
		{"packed one word differs — size change mutates the cell beyond the data pointer",
			cellPackedBytes(0xC000, 0x5), []any{cellPackedBytes(0xC000, 0x4)}, false,
			[]any{cellPackedBytes(0xC000, 0x4)}, false},

		// --- non-family return --------------------------------------------
		{"value-family return can never echo",
			Vector2{}, []any{Vector2{}}, false,
			[]any{Vector2{}}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			argVals := make([]reflect.Value, len(tc.allArgs))
			for i, a := range tc.allArgs {
				argVals[i] = reflect.ValueOf(a)
			}
			reflective := isPtrcallBorrowEcho(reflect.ValueOf(tc.ret), argVals, tc.hasReceiver)
			if reflective != tc.want {
				t.Fatalf("reflective classifier = %v, want %v", reflective, tc.want)
			}

			// Dispatch the typed side on the return's concrete type; only
			// the 16 copy-ret family members can reach a typed echo check
			// at all, mirroring the R constraint on the wrappers.
			var typed bool
			switch ret := tc.ret.(type) {
			case StringName:
				typed = typedEcho(ret, tc.params)
			case NodePath:
				typed = typedEcho(ret, tc.params)
			case Array:
				typed = typedEcho(ret, tc.params)
			case PackedByteArray:
				typed = typedEcho(ret, tc.params)
			default:
				// Outside RetCopyEncodable: no typed wrapper exists for
				// this return, the reflective default (never echo) is the
				// only behavior to pin.
				typed = false
			}
			if typed != reflective {
				t.Fatalf("typed classifier = %v, reflective classifier = %v", typed, reflective)
			}
		})
	}
}

// TestEchoCompareAssertsNilForForeignReturn documents the escape hatch: an
// R outside the generated emission list (only a test stub can produce one)
// gets no comparer, so no suppression is attempted.
func TestEchoCompareAssertsNilForForeignReturn(t *testing.T) {
	if c := newEchoCompare[spotRetStub](); c != nil {
		t.Fatalf("stub return type produced a comparer: %T", c)
	}
	if c := newEchoCompare[StringName](); c == nil {
		t.Fatal("family member produced no comparer")
	}
}

// TestVarcallRetDestroyerMatrix pins the varcall release matrix: only
// StringName/NodePath release their Go-side source after the return Variant
// is constructed; container varcall returns perform no return-path release
// (their lifecycle follows the owned-copy argument rules).
func TestVarcallRetDestroyerMatrix(t *testing.T) {
	if newVarcallRetDestroyer[StringName]() == nil {
		t.Fatal("StringName varcall return must release its source")
	}
	if newVarcallRetDestroyer[NodePath]() == nil {
		t.Fatal("NodePath varcall return must release its source")
	}
	// Checked per concrete closure type: erasing a nil func into `any` would
	// produce a non-nil interface and silently pass every assertion below.
	if d := newVarcallRetDestroyer[Callable](); d != nil {
		t.Fatal("Callable varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[Signal](); d != nil {
		t.Fatal("Signal varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[Array](); d != nil {
		t.Fatal("Array varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[Dictionary](); d != nil {
		t.Fatal("Dictionary varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[PackedByteArray](); d != nil {
		t.Fatal("PackedByteArray varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[PackedInt32Array](); d != nil {
		t.Fatal("PackedInt32Array varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[PackedInt64Array](); d != nil {
		t.Fatal("PackedInt64Array varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[PackedFloat32Array](); d != nil {
		t.Fatal("PackedFloat32Array varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[PackedFloat64Array](); d != nil {
		t.Fatal("PackedFloat64Array varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[PackedStringArray](); d != nil {
		t.Fatal("PackedStringArray varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[PackedVector2Array](); d != nil {
		t.Fatal("PackedVector2Array varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[PackedVector3Array](); d != nil {
		t.Fatal("PackedVector3Array varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[PackedVector4Array](); d != nil {
		t.Fatal("PackedVector4Array varcall return must not release its source")
	}
	if d := newVarcallRetDestroyer[PackedColorArray](); d != nil {
		t.Fatal("PackedColorArray varcall return must not release its source")
	}
}

// TestVariadicDispatchAppliesNoEchoSuppression pins the spec scenario: a
// variadic method receives raw Variants and never holds borrows, so nothing
// it returns can classify as an echo — the reflective variadic branch hard-
// codes destroySource=true (method_bind.go), and the typed variadic wrappers
// skip classification entirely (no arg decode exists). Here: even a byte-
// equal raw Variant cannot echo a StringName return.
func TestVariadicDispatchAppliesNoEchoSuppression(t *testing.T) {
	planted := cellStringName(0x1000)
	var rawVariant Variant
	copy(rawVariant[:], planted[:])
	if isPtrcallBorrowEcho(reflect.ValueOf(cellStringName(0x1000)),
		[]reflect.Value{reflect.ValueOf(rawVariant)}, false) {
		t.Fatal("a raw Variant argument must never classify as an echo")
	}
	// And the typed variadic wrapper's release is the unconditional varcall
	// destroyer — non-nil for StringName means "released per the matrix".
	if newVarcallRetDestroyer[StringName]() == nil {
		t.Fatal("variadic StringName return must release its source")
	}
}

// TestEchoCompareIsZeroAlloc pins the hot-path budget: the asserted closure
// is a concrete func(T, T) bool and `a == b` on the fixed-size cell compiles
// to an inline memequal — no boxing, no allocation per comparison.
func TestEchoCompareIsZeroAlloc(t *testing.T) {
	cmp := newEchoCompare[StringName]().(func(StringName, StringName) bool)
	a, b := cellStringName(0x1000), cellStringName(0x1000)
	if n := testing.AllocsPerRun(100, func() {
		if !cmp(a, b) {
			t.Fatal("planted cells differ")
		}
	}); n > 0 {
		t.Fatalf("echo compare allocates %v per call, want 0", n)
	}

	cmpArr := newEchoCompare[Array]().(func(Array, Array) bool)
	x, y := cellArray(0xA000), cellArray(0xA000)
	if n := testing.AllocsPerRun(100, func() {
		if !cmpArr(x, y) {
			t.Fatal("planted cells differ")
		}
	}); n > 0 {
		t.Fatalf("array echo compare allocates %v per call, want 0", n)
	}
}

// TestVarCopyRetEncoderBuildsEngineFree pins bind-time construction of the
// per-type Variant encoders for the whole copy-ret family (invoking them
// needs the engine; building them does not).
func TestVarCopyRetEncoderBuildsEngineFree(t *testing.T) {
	if newVarCopyRetEncoder[StringName]() == nil {
		t.Fatal("StringName varcall copy-ret encoder missing")
	}
	if newVarCopyRetEncoder[PackedVector4Array]() == nil {
		t.Fatal("PackedVector4Array varcall copy-ret encoder missing")
	}
}
