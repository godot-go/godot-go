package gdclassimpl

import (
	"runtime"
	"testing"
	"unsafe"
	"weak"
)

// Collectability tests for the pin-leak fix. They exercise the runtime mechanism
// the whole change rests on, without needing a live engine: a value pinned on the
// package pinner stays reachable forever (that is the leak), and the same value
// pinned on a body-scoped pinner becomes collectable once the body returns and
// the deferred Unpin runs (that is the fix).

type pinProbe struct {
	payload [128]byte
}

func collect(wp weak.Pointer[pinProbe]) bool {
	// Two cycles: the first clears the weak reference, the second makes the
	// result stable regardless of finaliser ordering.
	runtime.GC()
	runtime.GC()
	return wp.Value() == nil
}

func TestScopedPinnerReleasesValueForCollection(t *testing.T) {
	var wp weak.Pointer[pinProbe]

	func() {
		var pinner runtime.Pinner
		defer pinner.Unpin()
		v := &pinProbe{}
		pinner.Pin(v)
		wp = weak.Make(v)
		if wp.Value() == nil {
			t.Fatal("weak pointer was empty while the value was still live")
		}
	}()

	if !collect(wp) {
		t.Fatal("value pinned on a scoped pinner was still reachable after the " +
			"deferred Unpin ran; per-call pins are not being released")
	}
}

func TestPackagePinnerRetainsValue(t *testing.T) {
	var wp weak.Pointer[pinProbe]

	func() {
		v := &pinProbe{}
		pnr.Pin(v)
		wp = weak.Make(v)
	}()

	if collect(wp) {
		t.Fatal("value pinned on the package pinner was collected; the retention " +
			"mechanism this change reasons about has changed, and the design's " +
			"call-scoped-vs-program-lifetime split needs revisiting")
	}
}

// TestScopedPinnerReleasesGeneratedStyleBody mirrors the shape the templates emit
// -- return slot, owner cell, argument slice and argument cells all pinned -- and
// asserts none of it survives the call.
func TestScopedPinnerReleasesGeneratedStyleBody(t *testing.T) {
	var probes [4]weak.Pointer[pinProbe]

	func() {
		var pinner runtime.Pinner
		defer pinner.Unpin()

		returnSlot := &pinProbe{}
		ownerCell := &pinProbe{}
		argSlice := []*pinProbe{{}, {}}

		pinner.Pin(returnSlot)
		pinner.Pin(ownerCell)
		pinner.Pin(unsafe.Pointer(unsafe.SliceData(argSlice)))
		pinner.Pin(argSlice[0])
		pinner.Pin(argSlice[1])

		probes[0] = weak.Make(returnSlot)
		probes[1] = weak.Make(ownerCell)
		probes[2] = weak.Make(argSlice[0])
		probes[3] = weak.Make(argSlice[1])
	}()

	for i, p := range probes {
		if !collect(p) {
			t.Errorf("call-scratch probe %d survived the scoped pinner", i)
		}
	}
}
