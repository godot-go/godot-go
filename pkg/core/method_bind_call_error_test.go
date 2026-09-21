package core

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

func arityFixture(declared, defaults int, variadic bool) *GoMethodMetadata {
	return &GoMethodMetadata{
		gdeArgumentTypes:       make([]GDExtensionVariantType, declared),
		gdeDefaultArgumentPtrs: make([]GDExtensionVariantPtr, defaults),
		IsVariadic:             variadic,
	}
}

func TestClassifyVarcallArity(t *testing.T) {
	cases := []struct {
		name         string
		declared     int
		defaults     int
		variadic     bool
		supplied     int
		want         varcallArity
		wantExpected int
	}{
		{"exact match", 2, 0, false, 2, varcallArityOK, 0},
		{"too many", 2, 0, false, 3, varcallArityTooMany, 2},
		{"too many variadic ok", 1, 0, true, 5, varcallArityOK, 0},
		{"too few no defaults", 2, 0, false, 1, varcallArityTooFew, 2},
		{"too few zero args", 1, 0, false, 0, varcallArityTooFew, 1},
		{"all defaults satisfied by short call", 2, 2, false, 0, varcallArityOK, 0},
		{"all defaults partial supplied", 2, 2, false, 1, varcallArityOK, 0},
		{"partial defaults still too few", 3, 1, false, 1, varcallArityTooFew, 3},
		{"partial defaults satisfied by leading args", 3, 1, false, 2, varcallArityOK, 0},
		{"two defaults one supplied ok", 3, 2, false, 1, varcallArityOK, 0},
		{"variadic ignores too few", 1, 0, true, 0, varcallArityOK, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			md := arityFixture(tc.declared, tc.defaults, tc.variadic)
			got, expected := md.classifyVarcallArity(tc.supplied)
			if got != tc.want {
				t.Fatalf("classifyVarcallArity(%d) = %v, want %v", tc.supplied, got, tc.want)
			}
			if got != varcallArityOK && expected != tc.wantExpected {
				t.Fatalf("classifyVarcallArity(%d) expected=%d, want %d", tc.supplied, expected, tc.wantExpected)
			}
		})
	}
}

func TestClassifyVarcallArityMatchesCallFill(t *testing.T) {
	// classifyVarcallArity must reject exactly the calls that GoMethodMetadata.Call
	// would otherwise fail to fill under the trailing-default model: an unfilled
	// slot exists iff declared - supplied > defaults (for non-variadic methods).
	for declared := 0; declared <= 4; declared++ {
		for defaults := 0; defaults <= declared; defaults++ {
			for supplied := 0; supplied <= declared+2; supplied++ {
				md := arityFixture(declared, defaults, false)
				got, _ := md.classifyVarcallArity(supplied)
				// Call fills slot i iff i<supplied || i>=declared-defaults.
				defaultStart := declared - defaults
				callFillable := true
				for i := 0; i < declared; i++ {
					if !(i < supplied || i >= defaultStart) {
						callFillable = false
						break
					}
				}
				if callFillable && got == varcallArityTooFew {
					t.Fatalf("declared=%d defaults=%d supplied=%d: arity reports too-few but Call could fill", declared, defaults, supplied)
				}
				if !callFillable && got == varcallArityOK {
					t.Fatalf("declared=%d defaults=%d supplied=%d: arity reports OK but Call cannot fill", declared, defaults, supplied)
				}
			}
		}
	}
}

// TestCallPanicsOnUnfilledSlot verifies that a direct Go caller of the exported
// Call that under-supplies arguments (bypassing the varcall pre-validation) gets
// a loud panic naming the method, instead of the old silent zero-value fill.
func TestCallPanicsOnUnfilledSlot(t *testing.T) {
	md := &GoMethodMetadata{
		GdMethodName:           "too_short_method",
		gdeArgumentTypes:       make([]GDExtensionVariantType, 2),
		gdeDefaultArgumentPtrs: make([]GDExtensionVariantPtr, 0),
	}
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		md.Call(nil)
	}()
	if recovered == nil {
		t.Fatal("expected a panic for an unfilled slot, got none")
	}
	msg, ok := recovered.(string)
	if !ok || !strings.Contains(msg, "too few arguments") {
		t.Fatalf("expected panic message about too few arguments, got: %v", recovered)
	}
}

// bindGuardFixture provides a real reflect.Method with two parameters so the
// bind-time default-count guard can be exercised before any engine-dependent setup.
type bindGuardFixture struct{}

func (bindGuardFixture) TwoArgs(a, b int32) int32 { return a + b }

// TestBindRejectsExcessDefaults verifies that registering a method with more
// defaults than declared parameters panics at bind time.
func TestBindRejectsExcessDefaults(t *testing.T) {
	m, ok := reflect.TypeOf(&bindGuardFixture{}).MethodByName("TwoArgs")
	if !ok {
		t.Fatal("TwoArgs method not found")
	}
	excessDefaults := []Variant{{}, {}, {}} // 3 defaults for a 2-arg method
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		NewGoMethodMetadata(m, "bindGuardFixture", "two_args", "TwoArgs", nil, excessDefaults, 0)
	}()
	if recovered == nil {
		t.Fatal("expected a panic for excess defaults, got none")
	}
	msg, ok := recovered.(string)
	if !ok || !strings.Contains(msg, "more default arguments") {
		t.Fatalf("expected panic about excess default arguments, got: %v", recovered)
	}
}

// TestFillCallArgsVariadicSkipsFill verifies the D5 review fix: a variadic
// binding's declared slot is the slice itself, so the default fill (and its
// unfilled-slot panic) must not apply. fillCallArgs returns nil for variadic
// bindings even when they carry defaults; without the skip a zero-argument
// call would panic on the unfilled slice slot. The end-to-end dispatch with
// an empty slice is pinned by the engine demo (varargs_func with 0, 1 and 4
// args), since Call's return path requires live engine FFI pointers that bare
// unit tests lack.
func TestFillCallArgsVariadicSkipsFill(t *testing.T) {
	md := &GoMethodMetadata{
		GdMethodName:           "variadic_with_defaults",
		IsVariadic:             true,
		gdeArgumentTypes:       make([]GDExtensionVariantType, 1),
		DefaultArguments:       []Variant{{}},
		gdeDefaultArgumentPtrs: make([]GDExtensionVariantPtr, 1),
	}
	var got []Variant
	var recovered any
	func() {
		defer func() { recovered = recover() }()
		got = md.fillCallArgs(nil)
	}()
	if recovered != nil {
		t.Fatalf("variadic fillCallArgs with zero args must not panic, got: %v", recovered)
	}
	if got != nil {
		t.Fatalf("variadic fillCallArgs must return nil (dispatch uses gdArgs), got %v", got)
	}
}
