package core

import (
	"strings"
	"testing"
)

// namespaceProbe is the synthetic registration the namespace tests mutate.
func namespaceProbe() *ClassInfo {
	return &ClassInfo{
		Name: "NamespaceProbe",
		MethodMap: map[string]*MethodBindAndClassMethodInfo{
			"dup_method": {},
		},
		VirtualMethodMap: map[string]*MethodBindAndClassMethodInfo{
			"dup_virtual": {},
		},
	}
}

// TestEnsureMethodNameAvailableRejectsCollisions pins the pre-allocation guard:
// a name already bound as a method or as a virtual is refused before any engine
// metadata is allocated.
func TestEnsureMethodNameAvailableRejectsCollisions(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{
		"NamespaceProbe": namespaceProbe(),
	}, func() {
		cases := []struct {
			name   string
			method string
			want   string
		}{
			{"duplicate method", "dup_method", "Binding duplicate method"},
			{"name used by a virtual", "dup_virtual", "Method already bound as virtual"},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				msg, fired := recoverFatal(func() {
					ensureMethodNameAvailable("NamespaceProbe", c.method, "GoMethod")
				})
				if !fired {
					t.Fatalf("expected bind-time rejection for %q, got none", c.method)
				}
				if !strings.Contains(msg, c.want) {
					t.Fatalf("expected %q in the diagnostic, got: %s", c.want, msg)
				}
			})
		}
	})
}

// TestEnsureMethodNameAvailableRejectsUnregisteredClass pins that a bind for a
// class with no registration cannot create a method namespace.
func TestEnsureMethodNameAvailableRejectsUnregisteredClass(t *testing.T) {
	withRegisteredClasses(t, map[string]*ClassInfo{}, func() {
		msg, fired := recoverFatal(func() {
			ensureMethodNameAvailable("MissingClass", "some_method", "GoMethod")
		})
		if !fired {
			t.Fatal("an unregistered class must fail the bind")
		}
		if !strings.Contains(msg, "Class doesn't exist") {
			t.Fatalf("expected an unregistered-class diagnostic, got: %s", msg)
		}
	})
}
