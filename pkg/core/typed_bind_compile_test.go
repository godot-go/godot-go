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
//   - the remaining fixtures must fail to build naming the offending type (and,
//     where the compiler reports a constraint failure rather than a signature
//     mismatch, the constraint too): float32 is outside nativeSlot by decoder
//     parity, StringName is outside retValueSlot by copy-ownership, and an
//     Object- or Ref-typed parameter is outside nativeSlot because decoding it
//     is per-call instance-binding work.
//
// Signatures the unions reject stay bindable through the name-string API —
// the reflective fallback the package doc promises — so the rejection is
// strictly about the typed surface.
func TestTypedBindCompileRejection(t *testing.T) {
	cases := []struct {
		name    string
		pkg     string
		want    []string // nil means the build must succeed
		message string
	}{
		{"well_typed_fixture_builds", "./testdata/typedbindok", nil, ""},
		{"float32_argument_rejected", "./testdata/typedbindbadarg",
			[]string{"float32"},
			"a float32-parameter bind via the typed overload must not compile"},
		{"copy_ret_through_value_family_rejected", "./testdata/typedbindbadret",
			[]string{"retValueSlot", "StringName"},
			"a StringName-return bind via the value overload must not compile"},
		{"object_argument_rejected", "./testdata/typedbindobjectarg",
			[]string{"nativeSlot", "Node2D"},
			"an object-parameter bind via the typed overload must not compile"},
		{"ref_argument_rejected", "./testdata/typedbindrefarg",
			[]string{"nativeSlot", "RefBase"},
			"a Ref-parameter bind via the typed overload must not compile"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cmd := exec.Command("go", "build", c.pkg)
			out, err := cmd.CombinedOutput()
			if c.want == nil {
				if err != nil {
					t.Fatalf("%s: %v\n%s", c.message, err, out)
				}
				return
			}
			if err == nil {
				t.Fatalf("%s\n%s", c.message, out)
			}
			for _, want := range c.want {
				if !strings.Contains(string(out), want) {
					t.Fatalf("expected %q in the build diagnostic, got:\n%s", want, out)
				}
			}
		})
	}
}
