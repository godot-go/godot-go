package builtin

/*
	A-before/A-after benchmark for the typed-nil check switch
	(trim-dispatch-overhead task 3.3). The reflect form the contract check
	replaced is kept here as a same-run control: objectArgPtrReflectChecked
	is byte-for-byte what ObjectArgPtr did before the switch, so the two
	benchmarks measure the mechanism delta on one machine rather than
	comparing against a historical figure.

	Engine-free: typedNilProbe answers IsNil through its own declaration and
	the live case reads only Owner (an unsafe.Pointer conversion), so no
	benchmark iteration crosses into the engine. Run:

		go test ./pkg/builtin/ -run XXXnoTests -bench 'ObjectArgPtrCheck' -benchmem
*/

import (
	"reflect"
	"testing"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
)

type benchNilProbe struct {
	WrappedImpl
}

func (p *benchNilProbe) GetClassName() string       { return "benchNilProbe" }
func (p *benchNilProbe) GetParentClassName() string { return "Object" }
func (p *benchNilProbe) IsNil() bool                { return p == nil }

// objectArgPtrReflectChecked is the pre-switch implementation: a typed nil
// test through reflect.ValueOf before handing the accessor's result on.
func objectArgPtrReflectChecked(obj Wrapped) GDExtensionObjectPtr {
	if obj == nil {
		return nil
	}
	if v := reflect.ValueOf(obj); v.Kind() == reflect.Ptr && v.IsNil() {
		return nil
	}
	return obj.AsGDExtensionObjectPtr()
}

func BenchmarkObjectArgPtrCheckTypedNil(b *testing.B) {
	var obj Wrapped = (*benchNilProbe)(nil)
	for b.Loop() {
		if got := ObjectArgPtr(obj); got != nil {
			b.Fatalf("typed nil: got %v, want nil", got)
		}
	}
}

// BenchmarkObjectArgPtrCheckBareAccessor is the floor: the accessor call by
// itself, with no nil check, separating check cost from call cost.
func BenchmarkObjectArgPtrCheckBareAccessor(b *testing.B) {
	owner := sentinelOwner(0x22334455)
	p := &benchNilProbe{}
	p.Owner = owner
	want := (GDExtensionObjectPtr)(unsafe.Pointer(owner))
	obj := Wrapped(p)
	for b.Loop() {
		if got := obj.AsGDExtensionObjectPtr(); got != want {
			b.Fatalf("accessor: got %v, want %v", got, want)
		}
	}
}

func BenchmarkObjectArgPtrCheckTypedNilReflectOld(b *testing.B) {
	var obj Wrapped = (*benchNilProbe)(nil)
	for b.Loop() {
		if got := objectArgPtrReflectChecked(obj); got != nil {
			b.Fatalf("typed nil: got %v, want nil", got)
		}
	}
}

func BenchmarkObjectArgPtrCheckLive(b *testing.B) {
	owner := sentinelOwner(0x22334455)
	p := &benchNilProbe{}
	p.Owner = owner
	want := (GDExtensionObjectPtr)(unsafe.Pointer(owner))
	for b.Loop() {
		if got := ObjectArgPtr(p); got != want {
			b.Fatalf("live object: got %v, want %v", got, want)
		}
	}
}

func BenchmarkObjectArgPtrCheckLiveReflectOld(b *testing.B) {
	owner := sentinelOwner(0x22334455)
	p := &benchNilProbe{}
	p.Owner = owner
	want := (GDExtensionObjectPtr)(unsafe.Pointer(owner))
	for b.Loop() {
		if got := objectArgPtrReflectChecked(p); got != want {
			b.Fatalf("live object: got %v, want %v", got, want)
		}
	}
}
