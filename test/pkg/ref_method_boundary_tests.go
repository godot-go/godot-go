package pkg

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

// In-engine fixtures for Ref crossing at method boundaries (openspec change
// support-variant-reflect-completeness).
//
// Two dispatch paths share one decode: GDScript reaches them differently and
// the suite pins both. A typed GDScript local (var res: Resource = ...) lets
// the compiler emit the ptrcall fast path; example.call(...) pins varcall. The
// ptrcall argument cells here are additionally driven through DecodePtrcallArgs
// by hand -- same shape as ptrcall_object_decode_tests.go -- because the
// fast-path failure modes (wrong class, bare Ref against a plain object) have
// no error slot and must not abort the process from a fixture call site.
//
// Return codes follow the existing convention: 1 on success, negative per
// failed assertion, -900 on panic.

// refBoundaryStash holds one acquired Ref across calls to prove retention;
// TestRefBoundaryRelease drops it and collects before the shutdown leak
// snapshot, so the acquired reference cannot read as an engine leak.
var refBoundaryStash RefResource

// refBoundaryRejectProbeRan counts executions of the mismatch probe body. A
// rejected call must leave it untouched: containment means the body never runs.
var refBoundaryRejectProbeRan int64

// TestRefBoundaryValid reports whether the Ref argument arrived valid. A nil
// argument is the zero Ref and the body still runs, so the script side expects
// 0 rather than a rejection.
func (e *Example) TestRefBoundaryValid(r RefResource) int64 {
	if r == nil || !r.IsValid() {
		return 0
	}
	obj := r.ToObject()
	if isNullObject(obj) {
		return 0
	}
	return 1
}

// TestRefBoundaryEcho hands the Ref straight back out. The instance id the
// script compares on both ends is what proves the decoded Ref identifies the
// caller's object and the return encoder carried it, not a substitute.
func (e *Example) TestRefBoundaryEcho(r RefResource) RefResource {
	return r
}

// TestRefBoundaryNullReturn hands back the zero Ref. An invalid Ref returns
// null, the same way godot-cpp's empty Ref<T> converts.
func (e *Example) TestRefBoundaryNullReturn() RefResource {
	return nil
}

// TestRefBoundaryBareRefValid is the bare builtin.Ref spelling: it stands for
// Ref<RefCounted>, so any reference-counted object decodes and anything else
// is a caller error the varcall path contains.
func (e *Example) TestRefBoundaryBareRefValid(r Ref) int64 {
	if r == nil || !r.IsValid() {
		return 0
	}
	if isNullObject(r.ToObject()) {
		return 0
	}
	return 1
}

// TestRefBoundarySubclassSliced takes a RefShape2D while the script passes a
// RectangleShape2D. The Ref slices to the declared pointee (Ptr() yields the
// declared class) while the object keeps its runtime identity.
func (e *Example) TestRefBoundarySubclassSliced(shape RefShape2D) int64 {
	if shape == nil || !shape.IsValid() {
		return -1
	}
	obj := shape.ToObject()
	if isNullObject(obj) {
		return -2
	}
	cls := obj.GetClass()
	defer cls.Destroy()
	if cls.ToUtf8() != "RectangleShape2D" {
		fmt.Printf("  ref_subclass_sliced: FAIL: runtime class %s, want RectangleShape2D\n", cls.ToUtf8())
		return -3
	}
	var declared Resource = shape.Ptr()
	if declared == nil || isNullObject(declared) {
		return -4
	}
	fmt.Printf("  ref_subclass_sliced: RefShape2D holds instance %d as %s\n", obj.GetInstanceId(), cls.ToUtf8())
	return 1
}

// TestRefBoundaryRefcountAtLeastTwo makes the acquire observable: by the time
// the body runs, the object count is at least the script's own reference plus
// the one this call took (godot-cpp's PtrToArg<Ref<T>>::convert equivalent).
func (e *Example) TestRefBoundaryRefcountAtLeastTwo(r RefResource) int64 {
	if r == nil || !r.IsValid() {
		return -1
	}
	if n := r.ToObject().GetReferenceCount(); n < 2 {
		fmt.Printf("  ref_refcount: FAIL: reference count %d, want at least 2\n", n)
		return -2
	}
	return 1
}

// TestRefBoundaryStash keeps the decoded Ref past this call. The script then
// drops its own reference; only the acquired one can keep the object alive.
// The decode's +1 is released deterministically at call end (releaseDecodedRefArgs),
// so the stash must copy -- not alias -- to hold a reference of its own.
func (e *Example) TestRefBoundaryStash(r RefResource) int64 {
	if r == nil || !r.IsValid() {
		return -1
	}
	refBoundaryStash = NewRefAcquire[Resource](r.Ptr())
	return 1
}

// TestRefBoundaryStashAlive reads through the stashed Ref. Reading a freed
// object faults rather than returning false, so a pass here is proof of
// retention, not just of a non-nil field.
func (e *Example) TestRefBoundaryStashAlive() int64 {
	r := refBoundaryStash
	if r == nil || !r.IsValid() {
		return -1
	}
	obj := r.ToObject()
	if isNullObject(obj) {
		return -2
	}
	cls := obj.GetClass()
	defer cls.Destroy()
	if cls.ToUtf8() != "RectangleShape2D" {
		fmt.Printf("  ref_stash_alive: FAIL: runtime class %s, want RectangleShape2D\n", cls.ToUtf8())
		return -3
	}
	fmt.Printf("  ref_stash_alive: stashed Ref survived the script dropping its reference\n")
	return 1
}

// TestRefBoundaryRelease drops the stash and releases the reference it
// acquired, deterministically rather than relying on a GC finalizer, so the
// engine's exit leak snapshot sees a clean reference count.
func (e *Example) TestRefBoundaryRelease() int64 {
	if refBoundaryStash != nil {
		refBoundaryStash.Unref()
		refBoundaryStash = nil
	}
	runtime.GC()
	runtime.GC()
	return 1
}

// TestRefBoundaryRejectProbe's body executing at all is the failure a
// contained rejection must avoid. The script asserts the count never moved.
func (e *Example) TestRefBoundaryRejectProbe(r RefResource) int64 {
	refBoundaryRejectProbeRan++
	return 1
}

// TestRefBoundaryRejectProbeCount reports how many times the probe body ran.
func (e *Example) TestRefBoundaryRejectProbeCount() int64 {
	return refBoundaryRejectProbeRan
}

// TestRefBoundaryUserRefId reads the Id field off a decoded bare Ref that the
// script passed in. ExampleRef is a user-defined extension class whose Go
// wrapper is the only place the Id lives, so returning the value the script
// just set proves the decode resolved to that live wrapper -- not a fresh
// engine-only copy -- through whichever path the call took (task 6.2).
func (e *Example) TestRefBoundaryUserRefId(r Ref) int64 {
	if r == nil || !r.IsValid() {
		return -1
	}
	ref, ok := r.ToObject().(*ExampleRef)
	if !ok {
		return -2
	}
	return int64(ref.GetId())
}

// TestRefBoundaryBareRefEcho hands the bare Ref back out unchanged: the bare
// builtin.Ref return signature, exercised end to end so the spec's acceptance
// scenario is pinned by a live call and not only by the bind-time unit test.
func (e *Example) TestRefBoundaryBareRefEcho(r Ref) Ref {
	return r
}

// TestRefBoundaryPtrcallCell decodes a Ref-shape cell over a live object and
// observes the acquire in the reference count.
func (e *Example) TestRefBoundaryPtrcallCell(shape Shape2D) int32 {
	return objectArgGuard("ptrcall_ref_cell", func() int32 {
		if isNullObject(shape) {
			return -1
		}
		before := shape.GetReferenceCount()

		decoded, ok := decodeOne(e, ObjectArgPtr(shape), reflect.TypeFor[RefShape2D]()).Interface().(RefShape2D)
		if !ok || decoded == nil || !decoded.IsValid() {
			return -2
		}
		defer decoded.Unref() // release the acquire the seam took; bypasses the dispatch cleanup
		if got := decoded.ToObject().GetInstanceId(); got != shape.GetInstanceId() {
			fmt.Printf("  ptrcall_ref_cell: FAIL: ref points at %d, want %d\n", got, shape.GetInstanceId())
			return -3
		}
		after := shape.GetReferenceCount()
		if after < before+1 {
			fmt.Printf("  ptrcall_ref_cell: FAIL: reference %d before decode, %d after; the decode must acquire\n", before, after)
			return -4
		}
		fmt.Printf("  ptrcall_ref_cell: RefShape2D decoded instance %d; reference %d -> %d\n", shape.GetInstanceId(), before, after)
		return 1
	})
}

// TestRefBoundaryPtrcallBareRefCell decodes a bare builtin.Ref against the
// protocol key RefCounted: any reference-counted object passes.
func (e *Example) TestRefBoundaryPtrcallBareRefCell(shape Shape2D) int32 {
	return objectArgGuard("ptrcall_ref_bare_cell", func() int32 {
		if isNullObject(shape) {
			return -1
		}
		decoded, ok := decodeOne(e, ObjectArgPtr(shape), reflect.TypeFor[Ref]()).Interface().(Ref)
		if !ok || decoded == nil || !decoded.IsValid() {
			return -2
		}
		defer decoded.Unref() // release the acquire the seam took; bypasses the dispatch cleanup
		if got := decoded.ToObject().GetInstanceId(); got != shape.GetInstanceId() {
			return -3
		}
		fmt.Printf("  ptrcall_ref_bare_cell: bare Ref decoded instance %d\n", shape.GetInstanceId())
		return 1
	})
}

// TestRefBoundaryPtrcallNilCell requires a null Ref cell to decode to the zero
// Ref with no engine call against the null pointer.
func (e *Example) TestRefBoundaryPtrcallNilCell() int32 {
	return objectArgGuard("ptrcall_ref_nil_cell", func() int32 {
		v := decodeOne(e, nil, reflect.TypeFor[RefResource]())
		if !v.IsValid() {
			return -2
		}
		if decoded, ok := v.Interface().(RefResource); ok && decoded != nil && decoded.IsValid() {
			fmt.Printf("  ptrcall_ref_nil_cell: FAIL: null cell decoded to a live Ref\n")
			return -3
		}
		fmt.Printf("  ptrcall_ref_nil_cell: null cell decoded to the zero Ref without touching the engine\n")
		return 1
	})
}

// TestRefBoundaryPtrcallMismatchCell feeds a Label against a RefResource
// declaration. The fast path has no CallError slot, so the refusal is a loud
// panic naming the argument index and both class names; the test catches it
// and announces the expected PANIC log first so make test output reads as the
// fixture doing its job.
func (e *Example) TestRefBoundaryPtrcallMismatchCell(node Node) int32 {
	return objectArgGuard("ptrcall_ref_mismatch", func() int32 {
		if isNullObject(node) {
			return -1
		}
		fmt.Printf("  ptrcall_ref_mismatch: about to decode a Label against RefResource; one expected PANIC log follows\n")
		var msg string
		func() {
			msg = capturePanic(func() {
				decodeOne(e, ObjectArgPtr(node), reflect.TypeFor[RefResource]())
			})
		}()
		if msg == "" {
			fmt.Printf("  ptrcall_ref_mismatch: FAIL: wrong-class argument decoded without error\n")
			return -2
		}
		if !strings.Contains(msg, "ref argument 0") || !strings.Contains(msg, "Label") || !strings.Contains(msg, "Resource") {
			fmt.Printf("  ptrcall_ref_mismatch: FAIL: message lacks the index or class names: %s\n", firstLine(msg))
			return -3
		}
		fmt.Printf("  ptrcall_ref_mismatch: PASS - refused loudly as required: %s\n", firstLine(msg))
		return 1
	})
}
