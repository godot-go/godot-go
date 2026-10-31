package builtin

import (
	"runtime"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/ffi"
)

// fakeRefCounted satisfies the full RefCounted method set through the embedded
// interface (unimplemented methods panic if ever called) and overrides exactly
// what NewRefAcquire and RefBase touch: the engine pointer accessor and the
// reference-count operations. This keeps the ownership tests free of cgo,
// which this project's test builds do not allow.
type fakeRefCounted struct {
	RefCounted
	ptr      GDExtensionObjectPtr
	refs     atomic.Int32
	released chan struct{}
}

func newFakeRefCounted() *fakeRefCounted {
	marker := byte(0)
	return &fakeRefCounted{
		ptr:      GDExtensionObjectPtr(unsafe.Pointer(&marker)),
		released: make(chan struct{}, 8),
	}
}

func (f *fakeRefCounted) AsGDExtensionObjectPtr() GDExtensionObjectPtr { return f.ptr }

// IsNil is declared so a typed nil answers without dereferencing the fake:
// the promotion through the embedded RefCounted interface would fault, the
// same contract every hand-written wrapper obeys.
func (f *fakeRefCounted) IsNil() bool { return f == nil }

func (f *fakeRefCounted) Reference() bool {
	f.refs.Add(1)
	return true
}

func (f *fakeRefCounted) Unreference() bool {
	n := f.refs.Add(-1)
	if n <= 0 {
		select {
		case f.released <- struct{}{}:
		default:
		}
	}
	return n == 0
}

func TestNewRefAcquiresItsOwnReference(t *testing.T) {
	obj := newFakeRefCounted()

	ref := NewRefAcquire[RefCounted](obj)

	if got := obj.refs.Load(); got != 1 {
		t.Fatalf("acquire took %d references, want 1", got)
	}
	if !ref.IsValid() {
		t.Fatal("acquired ref over a live object must be valid")
	}
	if ref.Ptr() != RefCounted(obj) {
		t.Fatal("acquired ref must hold the wrapped object")
	}
}

func TestRefAcquireUnrefReleasesExactlyOnce(t *testing.T) {
	obj := newFakeRefCounted()
	ref := NewRefAcquire[RefCounted](obj)

	ref.Unref()

	if got := obj.refs.Load(); got != 0 {
		t.Fatalf("after Unref the count is %d, want 0", got)
	}
	if ref.IsValid() {
		t.Fatal("released ref must not report valid")
	}

	// Releasing twice must not release a second reference (same contract the
	// ratified constructors pin).
	ref.Unref()
	if got := obj.refs.Load(); got != 0 {
		t.Fatalf("after a second Unref the count is %d, want 0", got)
	}
}

func TestRefAcquireOverNoObjectTakesNothing(t *testing.T) {
	// A typed nil reaches ObjectArgPtr's nil guard: Reference() must not
	// run against it (the fake would dereference nil) and no finalizer may be
	// installed.
	var obj *fakeRefCounted

	ref := NewRefAcquire[*fakeRefCounted](obj)

	if ref.IsValid() {
		t.Fatal("acquire over a null object must yield the zero ref")
	}
}

func TestRefAcquireFinalizerReleasesWhenUnreachable(t *testing.T) {
	obj := newFakeRefCounted()

	func() {
		ref := NewRefAcquire[RefCounted](obj)
		runtime.KeepAlive(ref)
	}()

	// The finalizer mirrors godot-cpp's destructor release; drive the GC and
	// wait briefly for it, as finalizers run asynchronously with collection.
	deadline := time.Now().Add(2 * time.Second)
	for {
		runtime.GC()
		select {
		case <-obj.released:
			if got := obj.refs.Load(); got != 0 {
				t.Fatalf("after the finalizer the count is %d, want 0", got)
			}
			return
		case <-time.After(10 * time.Millisecond):
		}
		if time.Now().After(deadline) {
			t.Fatalf("finalizer did not release the acquired reference; count is %d", obj.refs.Load())
		}
	}
}
