package builtin

import (
	"testing"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
)

// Godot reads an object argument by dereferencing the slot
// (PtrToArg<T*>::convert == *reinterpret_cast<T*const*>(p_ptr)), so a nil object
// must arrive as a null pointer instead of panicking on the way there.

// sentinelOwner returns a distinct addressable value usable as a fake GodotObject.
func sentinelOwner(marker int) *GodotObject {
	v := marker
	return (*GodotObject)(unsafe.Pointer(&v))
}

// typedProbe models a generated Impl type: it embeds WrappedImpl by value and
// completes the Wrapped contract, so a typed nil of it addresses the embedded
// field exactly as (*NodeImpl)(nil) would.
type typedProbe struct {
	WrappedImpl
}

func (p *typedProbe) GetClassName() string       { return "typedProbe" }
func (p *typedProbe) GetParentClassName() string { return "Object" }

// fakeRef implements the Ref contract without real refcounting, so the helper's
// IsValid-before-ToObject gate can be exercised in isolation. Real Ref behaviour is
// covered by the Godot-side integration tests.
type fakeRef struct {
	valid bool
	obj   RefCounted
}

func (f *fakeRef) ToObject() RefCounted { return f.obj }
func (f *fakeRef) Ref(from Ref)         {}
func (f *fakeRef) Unref()              {}

// IsValid mirrors RefBase.IsValid's nil-receiver safety (`r != nil && r.m_ref != zero`)
// rather than dereferencing unconditionally.
func (f *fakeRef) IsValid() bool { return f != nil && f.valid }

// --- accessor hardening (inner layer) ---

func TestWrappedImplAsGDExtensionObjectPtrNilReceiver(t *testing.T) {
	var w *WrappedImpl
	if got := w.AsGDExtensionObjectPtr(); got != nil {
		t.Fatalf("nil receiver: got %v, want nil", got)
	}
	if got := w.AsGDExtensionConstObjectPtr(); got != nil {
		t.Fatalf("nil receiver (const): got %v, want nil", got)
	}
}

func TestWrappedImplAsGDExtensionObjectPtrNilOwner(t *testing.T) {
	w := &WrappedImpl{}
	if got := w.AsGDExtensionObjectPtr(); got != nil {
		t.Fatalf("nil owner: got %v, want nil", got)
	}
	if got := w.AsGDExtensionConstObjectPtr(); got != nil {
		t.Fatalf("nil owner (const): got %v, want nil", got)
	}
}

func TestWrappedImplAsGDExtensionObjectPtrReturnsOwnerUnchanged(t *testing.T) {
	owner := sentinelOwner(0x11223344)
	w := &WrappedImpl{Owner: owner}

	if got := w.AsGDExtensionObjectPtr(); got != (GDExtensionObjectPtr)(unsafe.Pointer(owner)) {
		t.Fatalf("object ptr: got %v, want %v", got, unsafe.Pointer(owner))
	}
	if got := w.AsGDExtensionConstObjectPtr(); got != (GDExtensionConstObjectPtr)(unsafe.Pointer(owner)) {
		t.Fatalf("const object ptr: got %v, want %v", got, unsafe.Pointer(owner))
	}
}

func TestWrappedClassInstanceAsGDExtensionObjectPtrNilReceiver(t *testing.T) {
	var w *WrappedClassInstance
	if got := w.AsGDExtensionObjectPtr(); got != nil {
		t.Fatalf("nil receiver: got %v, want nil", got)
	}
	if got := w.AsGDExtensionConstObjectPtr(); got != nil {
		t.Fatalf("nil receiver (const): got %v, want nil", got)
	}
}

func TestWrappedClassInstanceAsGDExtensionObjectPtrNilInstance(t *testing.T) {
	w := &WrappedClassInstance{}
	if got := w.AsGDExtensionObjectPtr(); got != nil {
		t.Fatalf("nil instance: got %v, want nil", got)
	}
	if got := w.AsGDExtensionConstObjectPtr(); got != nil {
		t.Fatalf("nil instance (const): got %v, want nil", got)
	}
}

// --- ObjectArgPtr: the shape the generated code actually uses ---

func TestObjectArgPtrNilInterface(t *testing.T) {
	var obj Wrapped
	if got := ObjectArgPtr(obj); got != nil {
		t.Fatalf("nil interface: got %v, want nil", got)
	}
}

// The case receiver hardening cannot reach: a typed nil faults while Go addresses
// the embedded WrappedImpl, so only the helper's reflect check makes it safe.
func TestObjectArgPtrTypedNil(t *testing.T) {
	var obj Wrapped = (*typedProbe)(nil)
	if got := ObjectArgPtr(obj); got != nil {
		t.Fatalf("typed nil: got %v, want nil", got)
	}
}

func TestObjectArgPtrLiveObject(t *testing.T) {
	owner := sentinelOwner(0x55667788)
	p := &typedProbe{}
	p.Owner = owner

	if got := ObjectArgPtr(p); got != (GDExtensionObjectPtr)(unsafe.Pointer(owner)) {
		t.Fatalf("live object: got %v, want %v", got, unsafe.Pointer(owner))
	}
}

// --- RefArgPtr ---

func TestRefArgPtrNilRef(t *testing.T) {
	var ref Ref
	if got := RefArgPtr(ref); got != nil {
		t.Fatalf("nil ref: got %v, want nil", got)
	}
}

func TestRefArgPtrInvalidRef(t *testing.T) {
	ref := &fakeRef{valid: false}
	if got := RefArgPtr(ref); got != nil {
		t.Fatalf("invalid ref: got %v, want nil", got)
	}
}

func TestRefArgPtrTypedNilRef(t *testing.T) {
	var ref Ref = (*fakeRef)(nil)
	if got := RefArgPtr(ref); got != nil {
		t.Fatalf("typed-nil ref: got %v, want nil", got)
	}
}

// The valid-Ref path is a direct delegation to ObjectArgPtr once IsValid() passes,
// and needs a real refcounted object to assert against. It is covered by the
// Godot-side integration tests (SetShape round-trip and reference-count
// stability), not by this unit test.
