package core

import (
	"testing"

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
		{"partial defaults still too few", 3, 1, false, 1, varcallArityTooFew, 2},
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
	// would otherwise fail to fill: an unsatisfiable slot exists iff
	// max(supplied, defaults) < declared (for non-variadic methods).
	for declared := 0; declared <= 4; declared++ {
		for defaults := 0; defaults <= declared; defaults++ {
			for supplied := 0; supplied <= declared+2; supplied++ {
				md := arityFixture(declared, defaults, false)
				got, _ := md.classifyVarcallArity(supplied)
				// Call fills slot i iff i<supplied || i<defaults.
				callFillable := true
				for i := 0; i < declared; i++ {
					if !(i < supplied || i < defaults) {
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
