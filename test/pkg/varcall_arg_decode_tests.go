package pkg

import (
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
)

// In-engine tests for the varcall argument-decode failure contract (openspec
// change fix-varcall-arg-error-reporting).
//
// Two layers are covered here, because neither can substitute for the other:
//
//   - The decoder, driven through DecodeVarcallArgs. This is where a caller-
//     induced failure has to turn into an error instead of a panic, and where the
//     owned prefix has to be released. The seam is used for the same reason
//     DecodePtrcallArgs is: GDScript cannot observe a GDExtensionCallError, so a
//     test written there cannot tell "the engine rejected my call" from "my test
//     broke".
//   - The real dispatch, driven from GDScript through example.call(). Only that
//     path can show the method body did not run and the process survived.
//
// Return codes follow the existing convention: 1 on success, negative per failed
// assertion, -900 on panic.

// objectVariant wraps obj the way the engine hands an OBJECT variant to the
// varcall decoder. The caller owns the returned Variant and must Destroy it.
func objectVariant(obj Wrapped) Variant {
	return NewVariantGodotObject((*GodotObject)(obj.AsGDExtensionObjectPtr()))
}

// TestVarcallDecodeWrongClass decodes a Label against a Sprite2D-declared
// parameter. The two classes are unrelated, so no wrapper satisfying Sprite2D
// exists and the decode must come back as a typed failure naming argument 0.
//
// This is the falsification test for the change: against the pre-fix decoder the
// same input ran a log.Panic off the cgo boundary and aborted the process.
func (e *Example) TestVarcallDecodeWrongClass(label Label) int32 {
	return objectArgGuard("varcall_wrong_class", func() int32 {
		if isNullObject(label) {
			return -1
		}
		v := objectVariant(label)
		defer v.Destroy()

		vals, err := DecodeVarcallArgs(e, []Variant{v}, []reflect.Type{reflect.TypeFor[Sprite2D]()})
		if err == nil {
			fmt.Printf("  varcall_wrong_class: FAIL: a Label decoded as Sprite2D (%d values)\n", len(vals))
			return -2
		}
		var decodeErr *VarcallArgDecodeError
		if !errors.As(err, &decodeErr) {
			fmt.Printf("  varcall_wrong_class: FAIL: error is %T, want *VarcallArgDecodeError: %v\n", err, err)
			return -3
		}
		if decodeErr.Index != 0 {
			fmt.Printf("  varcall_wrong_class: FAIL: failed index %d, want 0\n", decodeErr.Index)
			return -4
		}
		fmt.Printf("  varcall_wrong_class: rejected argument %d: %v\n", decodeErr.Index, err)
		return 1
	})
}

// TestVarcallDecodeSubclassStillAccepted is the control for the test above. A
// Label is a Node, so decoding it against a Node parameter must still succeed.
// Without this, a decoder that rejected every object would look like a pass.
func (e *Example) TestVarcallDecodeSubclassStillAccepted(label Label) int32 {
	return objectArgGuard("varcall_subclass_accepted", func() int32 {
		if isNullObject(label) {
			return -1
		}
		wantID := label.GetInstanceId()
		v := objectVariant(label)
		defer v.Destroy()

		vals, err := DecodeVarcallArgs(e, []Variant{v}, []reflect.Type{reflect.TypeFor[Node]()})
		if err != nil {
			fmt.Printf("  varcall_subclass_accepted: FAIL: %v\n", err)
			return -2
		}
		if len(vals) != 2 {
			fmt.Printf("  varcall_subclass_accepted: FAIL: %d values, want 2 (receiver + 1)\n", len(vals))
			return -3
		}
		got, ok := vals[1].Interface().(Node)
		if !ok || isNullObject(got) {
			return -4
		}
		if got.GetInstanceId() != wantID {
			fmt.Printf("  varcall_subclass_accepted: FAIL: id %d, want %d\n", got.GetInstanceId(), wantID)
			return -5
		}
		fmt.Printf("  varcall_subclass_accepted: Label accepted as Node instance %d\n", wantID)
		return 1
	})
}

// TestVarcallDecodeOwnedPrefixReleased decodes an owned Array copy, then fails on
// the object that follows it. The Array is decoded first, so it is the decoder's
// to release when the later argument fails; if it is not released the harness's
// leak check catches it at exit, and if it were released twice the run dies on
// the clobberfree GODEBUG rather than passing quietly.
//
// The decoder also returns no values on failure, which is what keeps the caller
// from destroying the same arguments a second time.
func (e *Example) TestVarcallDecodeOwnedPrefixReleased(arr Array, label Label) int32 {
	return objectArgGuard("varcall_prefix_released", func() int32 {
		if isNullObject(label) {
			return -1
		}
		objV := objectVariant(label)
		defer objV.Destroy()
		// Built with the real Variant constructor, not the raw byte-copying
		// NewVariantCopyWithGDExtensionConstVariantPtr: that helper copies the
		// Variant's bytes without taking a reference, so destroying one leaves
		// the original pointing at freed memory.
		arrV := NewVariantArray(arr)
		defer arrV.Destroy()

		vals, err := DecodeVarcallArgs(e, []Variant{arrV, objV}, []reflect.Type{
			reflect.TypeFor[Array](),
			reflect.TypeFor[Sprite2D](),
		})
		if err == nil {
			fmt.Printf("  varcall_prefix_released: FAIL: decode succeeded, want failure\n")
			return -2
		}
		var decodeErr *VarcallArgDecodeError
		if !errors.As(err, &decodeErr) {
			fmt.Printf("  varcall_prefix_released: FAIL: error is %T, want *VarcallArgDecodeError\n", err)
			return -3
		}
		if decodeErr.Index != 1 {
			fmt.Printf("  varcall_prefix_released: FAIL: failed index %d, want 1\n", decodeErr.Index)
			return -4
		}
		if vals != nil {
			fmt.Printf("  varcall_prefix_released: FAIL: values returned with an error, caller could double-free\n")
			return -5
		}
		fmt.Printf("  varcall_prefix_released: owned Array released, failure reported at argument 1\n")
		return 1
	})
}

// TestVarcallDecodeSuccessReleasesOwnedContainer is the success-path control: a
// well-formed Array argument still decodes and the caller still owns the value to
// release. The prefix-unwind path must not disturb this one.
func (e *Example) TestVarcallDecodeSuccessReleasesOwnedContainer(arr Array) int32 {
	return objectArgGuard("varcall_success_container", func() int32 {
		arrV := NewVariantArray(arr)
		defer arrV.Destroy()

		vals, err := DecodeVarcallArgs(e, []Variant{arrV}, []reflect.Type{reflect.TypeFor[Array]()})
		if err != nil {
			fmt.Printf("  varcall_success_container: FAIL: %v\n", err)
			return -2
		}
		if len(vals) != 2 {
			return -3
		}
		got, ok := vals[1].Interface().(Array)
		if !ok {
			return -4
		}
		if got.Size() != arr.Size() {
			fmt.Printf("  varcall_success_container: FAIL: size %d, want %d\n", got.Size(), arr.Size())
			return -5
		}
		// The seam hands the decoded values back without releasing them, which is
		// right: on the real path Call releases them once the return is encoded.
		// Driving the seam directly means this test owns them, so release here or
		// the harness's leak check fires.
		got.Destroy()
		fmt.Printf("  varcall_success_container: %d-element Array decoded intact\n", arr.Size())
		return 1
	})
}

// rejectProbeCounts how many times the probe body below actually ran. The driver
// reads it back after a deliberately rejected call: the only way to show a
// method body did not run is to have a body that reports running.
var rejectProbeCount int32

// TestVarcallRejectProbe declares a Sprite2D parameter. Called with a Sprite2D it
// runs and counts up; called with anything that cannot decode to Sprite2D the
// dispatch must reject before reaching this body.
func (e *Example) TestVarcallRejectProbe(target Sprite2D) int32 {
	atomic.AddInt32(&rejectProbeCount, 1)
	return 1
}

// TestVarcallRejectProbeCount exposes rejectProbeCount to the driver.
func (e *Example) TestVarcallRejectProbeCount() int32 {
	return atomic.LoadInt32(&rejectProbeCount)
}
