package core

import (
	"os/exec"
	"strings"
	"testing"
)

// TestTypedBindCompileRejection is the task 3.2 fixture tier: the typed bind
// overloads must reject mismatched handlers at COMPILE time, not at bind time.
// The fixtures live under testdata/ (invisible to package wildcards) and are
// built here with the go tool:
//
//   - typedbindok builds, proving the harness and the constraint unions admit
//     well-typed binds (so a rejection assertion cannot pass vacuously),
//   - typedbindbadarg and typedbindbadret must fail to build with the named
//     constraint in the diagnostic: float32 is outside nativeSlot by decoder
//     parity, and StringName is outside retValueSlot by copy-ownership.
//
// Signatures the unions reject stay bindable through the name-string API —
// the reflective fallback the package doc promises — so the rejection is
// strictly about the typed surface.
func TestTypedBindCompileRejection(t *testing.T) {
	cases := []struct {
		name    string
		pkg     string
		want    string // "" means the build must succeed
		message string
	}{
		{"well_typed_fixture_builds", "./testdata/typedbindok", "", ""},
		{"float32_argument_rejected", "./testdata/typedbindbadarg", "float32",
			"a float32-parameter bind via the typed overload must not compile"},
		{"copy_ret_through_value_family_rejected", "./testdata/typedbindbadret", "retValueSlot",
			"a StringName-return bind via the value overload must not compile"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmd := exec.Command("go", "build", c.pkg)
			out, err := cmd.CombinedOutput()
			if c.want == "" {
				if err != nil {
					t.Fatalf("%s: %v\n%s", c.message, err, out)
				}
				return
			}
			if err == nil {
				t.Fatalf("%s\n%s", c.message, out)
			}
			if !strings.Contains(string(out), c.want) {
				t.Fatalf("expected %q in the build diagnostic, got:\n%s", c.want, out)
			}
		})
	}
}
