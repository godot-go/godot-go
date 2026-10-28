package core

import (
	"strings"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

// typed-method-dispatch task 6.1 unit tier (engine-free).
//
// The fixed-shape callback closures are plain Go funcs closing over the
// typed handler; the dispatch sites in classdb_callback.go / to_string.go
// prefer them when the dedicated binds attached one and fall back to the
// reflective Func.Call body otherwise. These tests drive the builders
// directly (the dispatch sites are cgo exports needing a live engine; the
// engine tier exercises the typed closures once task 7.1 binds demo classes
// through the dedicated API, and the four named suites cover the name-string
// fallback unchanged today).

type cbRecord struct {
	instName string
	what     int32
	reversed bool
	name     string
	sn       StringName
	value    Variant
}

// cbStubHandler returns a recorder wired to a spotStubClass method-style
// handler: each shape test asserts the receiver resolved through Self() is
// the instance handed to the closure, plus the argument pass-through.
func TestNotificationShapePassesThrough(t *testing.T) {
	stub := &spotStubClass{}
	var got cbRecord
	var called int

	cb := newNotificationShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, what int32, reversed bool) {
			if self != stub {
				t.Fatalf("handler receiver %p, want %p", self, stub)
			}
			got = cbRecord{what: what, reversed: reversed}
			called++
		})

	cb(stub, 16, true)

	if called != 1 || got.what != 16 || !got.reversed {
		t.Fatalf("dispatch got %+v called=%d, want {what:16 reversed:true} once", got, called)
	}
}

// plantedVariant builds an engine-free Variant whose bytes identify it for
// pass-through assertions (engine constructors need the live bindings table,
// which these engine-free tests never initialize).
func plantedVariant(tag byte) Variant {
	var v Variant
	v[0] = tag
	for i := 1; i < len(v); i++ {
		v[i] = tag ^ byte(i)
	}
	return v
}

func TestPropGetShapePassesThrough(t *testing.T) {
	stub := &spotStubClass{}
	planted := plantedVariant(42)

	cb := newPropGetShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name string) (Variant, bool) {
			if self != stub {
				t.Fatalf("handler receiver %p, want %p", self, stub)
			}
			if name != "dproperty_1" {
				t.Fatalf("name %q, want %q", name, "dproperty_1")
			}
			return planted, true
		})

	v, found := cb(stub, "dproperty_1")
	if !found {
		t.Fatal("found=false, want true")
	}
	if v != planted {
		t.Fatal("returned Variant is not the handler's value")
	}

	// The false-means-unhandled contract passes through unchanged.
	notFound := newPropGetShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name string) (Variant, bool) {
			return Variant{}, false
		})
	if _, found := notFound(stub, "missing"); found {
		t.Fatal("found=true, want the false-means-unhandled contract")
	}
}

func TestPropSetShapePassesThrough(t *testing.T) {
	stub := &spotStubClass{}
	planted := plantedVariant(7)

	cb := newPropSetShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name string, value Variant) bool {
			if self != stub {
				t.Fatalf("handler receiver %p, want %p", self, stub)
			}
			if name != "dproperty_2" || value != planted {
				t.Fatalf("args (%q, %v), want (\"dproperty_2\", %v)", name, value, planted)
			}
			return true
		})

	if !cb(stub, "dproperty_2", planted) {
		t.Fatal("consumed=false, want the handler's true")
	}
}

func TestPropCanRevertShapePassesThrough(t *testing.T) {
	stub := &spotStubClass{}
	planted := StringName{1, 2, 3, 4, 5, 6, 7, 8}

	cb := newPropCanRevertShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name StringName) bool {
			if self != stub {
				t.Fatalf("handler receiver %p, want %p", self, stub)
			}
			if name != planted {
				t.Fatalf("name %v, want the planted borrow bytes", planted)
			}
			return true
		})

	if !cb(stub, planted) {
		t.Fatal("can-revert=false, want the handler's true")
	}
}

func TestPropGetRevertShapePassesThrough(t *testing.T) {
	stub := &spotStubClass{}
	planted := plantedVariant(9)

	cb := newPropGetRevertShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name StringName) (Variant, bool) {
			if self != stub {
				t.Fatalf("handler receiver %p, want %p", self, stub)
			}
			return planted, true
		})

	v, found := cb(stub, StringName{})
	if !found || v != planted {
		t.Fatalf("got (%v, %v), want (planted, true)", v, found)
	}
}

func TestToStringShapePassesThrough(t *testing.T) {
	stub := &spotStubClass{}

	cb := newToStringShape[spotStubClass, *spotStubClass](func(self *spotStubClass) string {
		if self != stub {
			t.Fatalf("handler receiver %p, want %p", self, stub)
		}
		return "[ GDExtension::Stub <--> Instance ID:1 | é中 ]"
	})

	if s := cb(stub); !strings.Contains(s, "é中") {
		t.Fatalf("string %q did not pass through", s)
	}
}

// TestNotificationShapeIsZeroAlloc pins the hot-path budget: the closure is
// a concrete call through the selfTyped assertion, with no reflect boxes.
func TestNotificationShapeIsZeroAlloc(t *testing.T) {
	stub := &spotStubClass{}
	cb := newNotificationShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, what int32, reversed bool) {})

	if n := testing.AllocsPerRun(100, func() {
		cb(stub, 1, false)
	}); n > 0 {
		t.Fatalf("notification closure allocates %v per call, want 0", n)
	}
}

// TestCallbackBindsReachSharedValidation drives each dedicated bind against a
// class that resolves its qualified virtual but is not registered, so every
// bind must run the shared typed-bind gates and refuse at the class-
// existence check — the same refusal the generated overloads' tests pin. A
// bind that skipped validation would fail later (or panic differently).
func TestCallbackBindsReachSharedValidation(t *testing.T) {
	stub := &TypedBindStub{}
	cases := []struct {
		name string
		call func()
	}{
		{"notification", func() {
			ClassDBBindVirtualNotification(stub, (*TypedBindStub).V_TypedBindStub_Notification)
		}},
		{"get", func() {
			ClassDBBindVirtualGet(stub, (*TypedBindStub).V_TypedBindStub_Get)
		}},
		{"set", func() {
			ClassDBBindVirtualSet(stub, (*TypedBindStub).V_TypedBindStub_Set)
		}},
		{"can_revert", func() {
			ClassDBBindVirtualPropertyCanRevert(stub, (*TypedBindStub).V_TypedBindStub_PropertyCanRevert)
		}},
		{"get_revert", func() {
			ClassDBBindVirtualPropertyGetRevert(stub, (*TypedBindStub).V_TypedBindStub_PropertyGetRevert)
		}},
		{"to_string", func() {
			ClassDBBindVirtualToString(stub, (*TypedBindStub).V_TypedBindStub_ToString)
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Empty registry: the shared gates run and the bind refuses at
			// the class-existence check (a non-nil registry is required for
			// the lookup itself, which is why the case cannot run bare).
			withRegisteredClasses(t, map[string]*ClassInfo{}, func() {
				msg, fired := recoverFatal(tc.call)
				if !fired || !strings.Contains(msg, "Class doesn't exist.") {
					t.Fatalf("bind did not reach the shared class gate: fired=%v msg=%q", fired, msg)
				}
			})
		})
	}
}
