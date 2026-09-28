package core

import (
	"reflect"
	"strings"
	"testing"

	"github.com/godot-go/godot-go/pkg/util"
)

// Unit tests for the to_string dispatch rules. They drive resolveToStringFor with
// synthetic registration entries rather than a live engine object, so the branch
// that matters — a bound virtual's value reaching the caller instead of the
// hardcoded default — is covered without Godot running.

// markerRecv is the receiver type the synthetic to_string virtuals expect.
type markerRecv struct {
	label string
}

// withRegisteredClasses installs a fresh class registry for the duration of fn so
// tests neither depend on nor pollute the real one.
func withRegisteredClasses(t *testing.T, entries map[string]*ClassInfo, fn func()) {
	t.Helper()

	original := Internal.GDRegisteredGDClasses
	registry := util.NewSyncMap[string, *ClassInfo]()
	for k, v := range entries {
		registry.Set(k, v)
	}
	Internal.GDRegisteredGDClasses = registry
	defer func() { Internal.GDRegisteredGDClasses = original }()

	fn()
}

// classWithVirtual builds a ClassInfo whose to_string virtual is fn.
func classWithVirtual(name string, fn any, goMethodName string) *ClassInfo {
	return &ClassInfo{
		Name: name,
		VirtualMethodMap: map[string]*MethodBindAndClassMethodInfo{
			toStringGdName: {
				GoMethodMetadata: &GoMethodMetadata{
					ClassName:    name,
					GdMethodName: toStringGdName,
					GoMethodName: goMethodName,
					Func:         reflect.ValueOf(fn),
					IsVirtual:    true,
					GoReturnType: reflect.TypeOf(""),
					GoArgumentTypes: []reflect.Type{
						reflect.TypeFor[markerRecv](),
					},
				},
			},
		},
	}
}

func TestToStringDispatchesBoundVirtual(t *testing.T) {
	const marker = "MARKER: dispatched to_string ran"
	withRegisteredClasses(t, map[string]*ClassInfo{
		"MarkerClass": classWithVirtual("MarkerClass",
			func(r markerRecv) string { return marker + ":" + r.label },
			"V_MarkerClass_ToString"),
	}, func() {
		got := resolveToStringFor(markerRecv{label: "abc"}, "MarkerClass", 99)

		if got != marker+":abc" {
			t.Fatalf("expected the dispatched virtual's string, got %q", got)
		}
		if strings.Contains(got, "GDExtension::") {
			t.Fatalf("default format leaked into a dispatched result: %q", got)
		}
	})
}

func TestToStringFallsBackWhenNoVirtualBound(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{
		"NoVirtualClass": {
			Name:             "NoVirtualClass",
			VirtualMethodMap: map[string]*MethodBindAndClassMethodInfo{},
		},
	}, func() {
		got := resolveToStringFor(markerRecv{}, "NoVirtualClass", 7)
		want := "[ GDExtension::NoVirtualClass <--> Instance ID:7 ]"
		if got != want {
			t.Fatalf("expected default format %q, got %q", want, got)
		}
	})
}

func TestToStringFallsBackForUnregisteredClass(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{}, func() {
		got := resolveToStringFor(markerRecv{}, "NotRegistered", 11)
		want := "[ GDExtension::NotRegistered <--> Instance ID:11 ]"
		if got != want {
			t.Fatalf("expected default format %q, got %q", want, got)
		}
	})
}

// A mis-bound override must degrade to the default rather than panic: to_string is
// display-only and must not take the process down.
func TestToStringFallsBackOnNonStringReturn(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{
		"IntToStringClass": classWithVirtual("IntToStringClass",
			func(r markerRecv) int { return 42 },
			"V_IntToStringClass_ToString"),
	}, func() {
		got := resolveToStringFor(markerRecv{}, "IntToStringClass", 3)
		want := "[ GDExtension::IntToStringClass <--> Instance ID:3 ]"
		if got != want {
			t.Fatalf("expected default format %q for a non-string return, got %q", want, got)
		}
	})
}

func TestToStringFallsBackOnMultipleReturns(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{
		"PairToStringClass": classWithVirtual("PairToStringClass",
			func(r markerRecv) (string, bool) { return "first", true },
			"V_PairToStringClass_ToString"),
	}, func() {
		got := resolveToStringFor(markerRecv{}, "PairToStringClass", 5)
		want := "[ GDExtension::PairToStringClass <--> Instance ID:5 ]"
		if got != want {
			t.Fatalf("expected default format %q for a multi-value return, got %q", want, got)
		}
	})
}

// The default format is the contract the demo asserts against; pin its shape so a
// casual edit cannot silently change what users see.
func TestDefaultToStringFormatShape(t *testing.T) {
	got := defaultToStringFormat("Example", 1234)
	want := "[ GDExtension::Example <--> Instance ID:1234 ]"
	if got != want {
		t.Fatalf("default format changed: got %q want %q", got, want)
	}
}
