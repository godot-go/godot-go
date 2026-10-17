package core

import (
	"strings"
	"testing"

	"github.com/godot-go/godot-go/pkg/constant"
)

// Unit tests for the PropertySpec helpers. These are pure authoring-time
// functions: no engine, no registries, no cgo calls.

func mustPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected panic containing %q, got none", want)
		}
		msg, ok := r.(string)
		if !ok {
			if err, isErr := r.(error); isErr {
				msg = err.Error()
			}
		}
		if !strings.Contains(msg, want) {
			t.Fatalf("expected panic containing %q, got %q", want, msg)
		}
	}()
	fn()
}

func TestDeriveClassNameResourceTypeHintWins(t *testing.T) {
	spec := PropertySpec{
		Hint:       constant.PROPERTY_HINT_RESOURCE_TYPE,
		HintString: "Texture2D",
		ClassName:  "Ignored",
	}
	if got := deriveClassName(spec, "Example"); got != "Texture2D" {
		t.Fatalf("resource-type hint must take class_name from hint string, got %q", got)
	}
}

func TestDeriveClassNameExplicitQualifier(t *testing.T) {
	spec := PropertySpec{
		Hint:      constant.PROPERTY_HINT_NONE,
		ClassName: "Node",
	}
	if got := deriveClassName(spec, "Example"); got != "Node" {
		t.Fatalf("explicit qualifier must win when no resource-type hint, got %q", got)
	}
}

func TestDeriveClassNameFallsBackToOwningClass(t *testing.T) {
	spec := PropertySpec{Hint: constant.PROPERTY_HINT_NONE}
	if got := deriveClassName(spec, "Example"); got != "Example" {
		t.Fatalf("no hint and no qualifier must keep the owning class, got %q", got)
	}
}

func TestDeriveClassNameRangeHintKeepsOwningClass(t *testing.T) {
	spec := PropertySpec{
		Hint:       constant.PROPERTY_HINT_RANGE,
		HintString: "0:100:1",
	}
	if got := deriveClassName(spec, "Example"); got != "Example" {
		t.Fatalf("non-resource-type hint must not consume the hint string, got %q", got)
	}
}

func TestValidatePropertySpecRejectsMissingGetter(t *testing.T) {
	mustPanic(t, "Getter method must be specified", func() {
		validatePropertySpec(PropertySpec{Name: "no_getter"})
	})
}

func TestValidatePropertySpecRejectsResourceTypeWithoutHintString(t *testing.T) {
	mustPanic(t, "PROPERTY_HINT_RESOURCE_TYPE", func() {
		validatePropertySpec(PropertySpec{
			Name:   "bad_resource",
			Getter: "get_bad_resource",
			Hint:   constant.PROPERTY_HINT_RESOURCE_TYPE,
		})
	})
}

func TestValidatePropertySpecAcceptsValidSpecs(t *testing.T) {
	validatePropertySpec(PropertySpec{
		Name:   "plain",
		Getter: "get_plain",
	})
	validatePropertySpec(PropertySpec{
		Name:       "typed_resource",
		Getter:     "get_typed_resource",
		Hint:       constant.PROPERTY_HINT_RESOURCE_TYPE,
		HintString: "Texture2D",
	})
}
