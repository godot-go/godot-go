package core

import (
	"reflect"
	"strings"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/gdclassimpl"
)

// TestVariantTypeRefInterfaceAcceptance pins the supported Ref spellings for
// binding (support-variant-reflect-completeness design D1): the generated
// RefXxx interfaces and bare builtin.Ref register as OBJECT. Bare Ref must
// not depend on the name strip, whose empty result made the old registry
// probe miss it.
func TestVariantTypeRefInterfaceAcceptance(t *testing.T) {
	accepted := []struct {
		name string
		typ  reflect.Type
	}{
		{"RefResource", reflect.TypeFor[RefResource]()},
		{"RefShape2D", reflect.TypeFor[RefShape2D]()},
		{"bare Ref", reflect.TypeFor[Ref]()},
	}
	for _, c := range accepted {
		t.Run(c.name, func(t *testing.T) {
			if got := ReflectTypeToGDExtensionVariantType(c.typ); got != GDEXTENSION_VARIANT_TYPE_OBJECT {
				t.Fatalf("variant type for %s: got %v, want OBJECT", c.name, got)
			}
		})
	}
}

// TestVariantTypeConcreteRefRejectionPanics pins bind-time rejection of the
// concrete reference forms: the decode branches that once claimed to handle
// them could never succeed (the assertions contradict what the registries
// hand out), so the mapper must refuse them with a diagnostic naming the
// offending type instead.
func TestVariantTypeConcreteRefRejectionPanics(t *testing.T) {
	rejected := []any{
		gdclassimpl.RefResourceImpl{},
		&gdclassimpl.RefResourceImpl{},
		(*RefBase[Resource])(nil),
		(*RefBase[RefCounted])(nil),
	}
	for _, sample := range rejected {
		refType := reflect.TypeOf(sample)
		t.Run(refType.String(), func(t *testing.T) {
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				ReflectTypeToGDExtensionVariantType(refType)
			}()
			if recovered == nil {
				t.Fatalf("expected bind-time panic for %s, got none", refType)
			}
			msg, ok := recovered.(string)
			if !ok || !strings.Contains(msg, "RefXxx interface or bare builtin.Ref") {
				t.Fatalf("expected panic naming the supported Ref forms, got: %v", recovered)
			}
			if !strings.Contains(msg, refType.String()) {
				t.Fatalf("expected panic message to name the offending type %s, got: %v", refType, recovered)
			}
		})
	}
}

// refBindFixture provides real reflect.Methods whose first parameter is a
// refused Ref form, so the bind path rejects before any engine-dependent
// metadata setup runs (the parameter loop maps variant types first).
type refBindFixture struct{}

func (refBindFixture) TakeRefImplParameter(r gdclassimpl.RefResourceImpl) {}

func (refBindFixture) TakeRefImplPointerParameter(r *gdclassimpl.RefResourceImpl) {}

func (refBindFixture) TakeRefBasePointerParameter(r *RefBase[Resource]) {}

func (refBindFixture) ReturnsRefImpl() gdclassimpl.RefResourceImpl {
	return gdclassimpl.RefResourceImpl{}
}

// TestBindRejectsConcreteRefForms verifies registration refuses concrete Ref
// forms on both parameters and the return, so no method with an undecodable
// signature ever becomes callable.
func TestBindRejectsConcreteRefForms(t *testing.T) {
	for _, methodName := range []string{
		"TakeRefImplParameter",
		"TakeRefImplPointerParameter",
		"TakeRefBasePointerParameter",
		"ReturnsRefImpl",
	} {
		t.Run(methodName, func(t *testing.T) {
			m, ok := reflect.TypeOf(&refBindFixture{}).MethodByName(methodName)
			if !ok {
				t.Fatalf("%s method not found", methodName)
			}
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				NewGoMethodMetadata(m, "refBindFixture", strings.ToLower(methodName), methodName, nil, nil, 0)
			}()
			if recovered == nil {
				t.Fatalf("expected bind-time rejection for %s, got none", methodName)
			}
			msg, ok := recovered.(string)
			if !ok || !strings.Contains(msg, "RefXxx interface or bare builtin.Ref") {
				t.Fatalf("expected panic naming the supported Ref forms, got: %v", recovered)
			}
		})
	}
}

// TestRefArgumentClassKeying pins the declared-class keying (design D2): the
// generated interface names its pointee, bare Ref keys RefCounted. Engine-side
// crossing of these classes is pinned by the demo fixtures.
func TestRefArgumentClassKeying(t *testing.T) {
	cases := []struct {
		typ  reflect.Type
		want string
	}{
		{reflect.TypeFor[RefResource](), "Resource"},
		{reflect.TypeFor[RefShape2D](), "Shape2D"},
		{reflect.TypeFor[Ref](), "RefCounted"},
	}
	for _, c := range cases {
		t.Run(c.typ.String(), func(t *testing.T) {
			if got := refArgumentClass(c.typ); got != c.want {
				t.Fatalf("refArgumentClass(%s) = %q, want %q", c.typ, got, c.want)
			}
		})
	}
}
