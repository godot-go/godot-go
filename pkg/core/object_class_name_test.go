package core

import (
	"reflect"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/gdclassinit"
	"github.com/godot-go/godot-go/pkg/util"
)

// Unit tests for godotClassNameForObjectType, the resolution behind every
// object-typed argument and return in method metadata.
//
// The registries are populated by extension init, which does not run under a
// plain `go test`, so each case seeds exactly the entries it needs.

// Local named types stand in for classes the resolver looks up by name.
type FixtureEngineClass struct{}

type FixtureExtensionClass struct{}

type RefFixturePointee struct{}

type RefFoo struct{}

type foo struct{}

type TotallyUnknown struct{}

// withSeededRegistries adds the given entries for the duration of fn and removes
// them afterwards, so tests neither depend on nor pollute whatever the real
// init left behind.
//
// GDNativeConstructors and GDClassRefConstructors belong to another package and
// are read-only here, so they are seeded and unseeded in place.
// Internal.GDRegisteredGDClasses is nil until extension init runs, so the test
// owns an instance for the duration and puts the nil back afterwards.
func withSeededRegistries(t *testing.T, native []string, registered []string, refPointees []string, fn func()) {
	t.Helper()

	originalRegistered := Internal.GDRegisteredGDClasses
	if originalRegistered == nil {
		Internal.GDRegisteredGDClasses = util.NewSyncMap[string, *ClassInfo]()
	}

	for _, n := range native {
		GDNativeConstructors.Set(n, nil)
	}
	for _, n := range registered {
		Internal.GDRegisteredGDClasses.Set(n, &ClassInfo{Name: n})
	}
	for _, n := range refPointees {
		GDClassRefConstructors.Set(n, nil)
	}
	defer func() {
		for _, n := range native {
			GDNativeConstructors.Delete(n)
		}
		for _, n := range registered {
			Internal.GDRegisteredGDClasses.Delete(n)
		}
		for _, n := range refPointees {
			GDClassRefConstructors.Delete(n)
		}
		Internal.GDRegisteredGDClasses = originalRegistered
	}()

	fn()
}

func TestResolvePlainEngineClass(t *testing.T) {
	withSeededRegistries(t, []string{"FixtureEngineClass"}, nil, nil, func() {
		got, ok := godotClassNameForObjectType(reflect.TypeOf(FixtureEngineClass{}))
		if !ok || got != "FixtureEngineClass" {
			t.Fatalf("want FixtureEngineClass, got %q ok=%v", got, ok)
		}
	})
}

func TestResolveRegisteredExtensionClass(t *testing.T) {
	withSeededRegistries(t, nil, []string{"FixtureExtensionClass"}, nil, func() {
		got, ok := godotClassNameForObjectType(reflect.TypeOf(FixtureExtensionClass{}))
		if !ok || got != "FixtureExtensionClass" {
			t.Fatalf("want FixtureExtensionClass, got %q ok=%v", got, ok)
		}
	})
}

func TestResolveRefAdvertisesPointee(t *testing.T) {
	withSeededRegistries(t, nil, nil, []string{"FixturePointee"}, func() {
		got, ok := godotClassNameForObjectType(reflect.TypeOf(RefFixturePointee{}))
		if !ok || got != "FixturePointee" {
			t.Fatalf("want FixturePointee, got %q ok=%v", got, ok)
		}
	})
}

// A class whose own name starts with "Ref" must keep it. Plain-name lookup runs
// before the prefix strip for exactly this reason.
func TestResolveRefPrefixedClassNameNotStripped(t *testing.T) {
	withSeededRegistries(t, []string{"RefFoo"}, nil, []string{"Foo"}, func() {
		got, ok := godotClassNameForObjectType(reflect.TypeOf(RefFoo{}))
		if !ok {
			t.Fatal("RefFoo did not resolve")
		}
		if got != "RefFoo" {
			t.Fatalf("RefFoo was stripped to %q; plain-name lookup must win", got)
		}
	})
}

func TestResolveGenericObjectInterfacesAdvertiseBase(t *testing.T) {
	withSeededRegistries(t, nil, nil, nil, func() {
		for _, rt := range []reflect.Type{
			reflect.TypeFor[Object](),
			reflect.TypeFor[Ref](),
			reflect.TypeFor[GDClass](),
			reflect.TypeFor[GDExtensionClass](),
		} {
			got, ok := godotClassNameForObjectType(rt)
			if !ok || got != "Object" {
				t.Errorf("%s: want Object, got %q ok=%v", rt.Name(), got, ok)
			}
		}
	})
}

// A bare Ref must not resolve to the empty string its prefix strip would give.
func TestResolveBareRefIsNotEmptyString(t *testing.T) {
	withSeededRegistries(t, nil, nil, []string{""}, func() {
		got, _ := godotClassNameForObjectType(reflect.TypeFor[Ref]())
		if got == "" {
			t.Fatal("bare Ref resolved to an empty class name")
		}
		if got != "Object" {
			t.Fatalf("bare Ref want Object, got %q", got)
		}
	})
}

func TestResolveUnknownTypeFails(t *testing.T) {
	withSeededRegistries(t, nil, nil, nil, func() {
		if got, ok := godotClassNameForObjectType(reflect.TypeOf(TotallyUnknown{})); ok {
			t.Fatalf("unknown type resolved to %q; must report unresolved", got)
		}
	})
}

// Concrete Go class implementations are pointers, and a pointer's Name() is
// empty. Without dereferencing, a method returning *FixtureExtensionClass looked
// unresolvable and panicked at bind time.
func TestResolvePointerToRegisteredClass(t *testing.T) {
	withSeededRegistries(t, nil, []string{"FixtureExtensionClass"}, nil, func() {
		got, ok := godotClassNameForObjectType(reflect.TypeFor[*FixtureExtensionClass]())
		if !ok || got != "FixtureExtensionClass" {
			t.Fatalf("want FixtureExtensionClass for a pointer return, got %q ok=%v", got, ok)
		}
	})
}

func TestResolvePointerToEngineClass(t *testing.T) {
	withSeededRegistries(t, []string{"FixtureEngineClass"}, nil, nil, func() {
		got, ok := godotClassNameForObjectType(reflect.TypeFor[*FixtureEngineClass]())
		if !ok || got != "FixtureEngineClass" {
			t.Fatalf("want FixtureEngineClass for a pointer return, got %q ok=%v", got, ok)
		}
	})
}
