package pkg

import (
	"fmt"
	"runtime"
	"time"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

// In-engine regression for object-return ownership (openspec change
// fix-object-return-reference-ownership).
//
// A generated ptrcall that returns a refcounted object fills a Go-allocated
// return slot through PtrToArg<Ref<T>>::convert, which constructs a Ref and so
// calls reference(). That +1 is transferred to whoever reads the slot. The
// wrapper used to be built with the borrowing NewRef, which sets no finalizer, so
// the transferred reference was never released: 200 get_shape() calls ended with
// the shape at reference count 200 and a leaked RID at exit.
//
// Every test keeps one long-lived handle (`held`) on the same object so the count
// stays readable while the wrapper under test comes and goes. Counts are read
// through that handle, never through the wrapper being released, so a freed
// object is never dereferenced.
//
// Return codes follow the existing convention: 1 on success, negative per failed
// assertion, -900 if the Go side panicked.

// stableCount collects until two consecutive reads agree, and reports whether it
// settled before the deadline. Used to read a baseline that earlier tests'
// pending finalizers could still be draining.
func stableCount(obj RefCounted, timeout time.Duration) (int32, bool) {
	deadline := time.Now().Add(timeout)
	prev := obj.GetReferenceCount()
	for {
		time.Sleep(2 * time.Millisecond)
		runtime.GC()
		runtime.GC()
		next := obj.GetReferenceCount()
		if next == prev {
			return next, true
		}
		prev = next
		if time.Now().After(deadline) {
			return prev, false
		}
	}
}

// waitCount collects repeatedly until the reference count reaches want, and
// reports whether it got there before the deadline.
//
// A single runtime.GC() is not enough to observe the settled count: finalizers
// run on the Go finalizer goroutine, which the scheduler is free to leave
// undrained when GC returns. Measured directly, the same build gave "3 -> 3" on
// one run and "3 -> 45" on the next. Polling toward the target rather than
// sampling once keeps the test fast when finalizers are prompt, and it cannot
// pass by catching a transient value -- the assertion only ever becomes true by
// actually reaching the target, and never true by luck.
func waitCount(obj RefCounted, want int32, timeout time.Duration) (int32, bool) {
	deadline := time.Now().Add(timeout)
	current := obj.GetReferenceCount()
	for current != want {
		if time.Now().After(deadline) {
			return current, false
		}
		runtime.GC()
		runtime.GC()
		current = obj.GetReferenceCount()
		time.Sleep(2 * time.Millisecond)
	}
	return want, true
}

// shapeHolder returns a long-lived reference to the shape owned by owner, plus
// the reference count once it has settled.
func shapeHolder(owner CollisionShape2D) (RefShape2D, int32, bool) {
	held := owner.GetShape()
	if isNullRef(held) {
		return nil, 0, false
	}
	base, ok := stableCount(held.ToObject(), 10*time.Second)
	if !ok {
		return nil, 0, false
	}
	return held, base, true
}

// TestReturnRefcountStability is the get_shape loop the object-argument tests
// deliberately left unexercised. N returned wrappers are dropped without being
// unref'd; once Go collects them the count must be back where it started.
// Before the fix each call leaked its transferred +1, so the count came back as
// base+N and never fell.
func (e *Example) TestReturnRefcountStability(owner CollisionShape2D, iterations int64) int32 {
	return objectArgGuard("return_refcount_stability", func() int32 {
		if isNullObject(owner) || iterations <= 0 {
			return -1
		}
		held, base, ok := shapeHolder(owner)
		if !ok {
			return -2
		}
		// held must stay live for the whole test. Go's liveness analysis makes it
		// dead after its last use, so without this the baseline handle's own
		// finalizer runs mid-test and the count drops by one that has nothing to
		// do with the wrapper under observation.
		defer runtime.KeepAlive(held)

		for i := int64(0); i < iterations; i++ {
			_ = owner.GetShape()
		}

		after, settled := waitCount(held.ToObject(), base, 15*time.Second)
		fmt.Printf("  return_stability: base=%d, +%d returns -> %d\n", base, iterations, after)
		if !settled {
			fmt.Printf("  return_stability: FAIL: %d references retained\n", after-base)
			return -3
		}
		return 1
	})
}

// TestReturnDroppedReleases proves the finalizer actually fires: holding an extra
// owned wrapper raises the count by exactly one, and dropping it plus collection
// takes that one back down.
func (e *Example) TestReturnDroppedReleases(owner CollisionShape2D) int32 {
	return objectArgGuard("return_dropped_releases", func() int32 {
		if isNullObject(owner) {
			return -1
		}
		held, base, ok := shapeHolder(owner)
		if !ok {
			return -2
		}
		// held must stay live for the whole test. Go's liveness analysis makes it
		// dead after its last use, so without this the baseline handle's own
		// finalizer runs mid-test and the count drops by one that has nothing to
		// do with the wrapper under observation.
		defer runtime.KeepAlive(held)

		extra := owner.GetShape()
		during := held.ToObject().GetReferenceCount()
		runtime.KeepAlive(extra)
		if during != base+1 {
			fmt.Printf("  return_dropped: FAIL: holding a returned ref gave %d, want %d\n", during, base+1)
			return -3
		}

		extra = nil
		after, settled := waitCount(held.ToObject(), base, 15*time.Second)
		fmt.Printf("  return_dropped: base=%d, held=%d, dropped=%d\n", base, during, after)
		if !settled {
			fmt.Printf("  return_dropped: FAIL: dropping the wrapper released %d references, want 1\n", base-after)
			return -4
		}
		return 1
	})
}

// TestReturnUnrefIdempotence pins the contract with the explicit release path:
// Unref() drops the count by one and clears the finalizer, so later collection
// must be a no-op rather than a second release.
func (e *Example) TestReturnUnrefIdempotence(owner CollisionShape2D) int32 {
	return objectArgGuard("return_unref_idempotence", func() int32 {
		if isNullObject(owner) {
			return -1
		}
		held, base, ok := shapeHolder(owner)
		if !ok {
			return -2
		}
		// held must stay live for the whole test. Go's liveness analysis makes it
		// dead after its last use, so without this the baseline handle's own
		// finalizer runs mid-test and the count drops by one that has nothing to
		// do with the wrapper under observation.
		defer runtime.KeepAlive(held)

		extra := owner.GetShape()
		if held.ToObject().GetReferenceCount() != base+1 {
			return -3
		}

		extra.Unref()
		afterUnref := held.ToObject().GetReferenceCount()
		if afterUnref != base {
			fmt.Printf("  return_idempotence: FAIL: Unref left %d, want %d\n", afterUnref, base)
			return -4
		}

		// Collection must not take it below base. Settle to base and stay there.
		afterGC, settled := waitCount(held.ToObject(), base, 15*time.Second)
		fmt.Printf("  return_idempotence: base=%d, after Unref=%d, after GC=%d\n", base, afterUnref, afterGC)
		if !settled {
			fmt.Printf("  return_idempotence: FAIL: GC moved the count off %d to %d\n", base, afterGC)
			return -5
		}
		return 1
	})
}

// TestBorrowedRefNeverReleases guards the other half of the split. A borrowing
// wrapper built by NewRef must not release on GC, or every AsRef-style caller
// would decrement a reference owned elsewhere.
func (e *Example) TestBorrowedRefNeverReleases(owner CollisionShape2D) int32 {
	return objectArgGuard("borrowed_ref_never_releases", func() int32 {
		if isNullObject(owner) {
			return -1
		}
		held, base, ok := shapeHolder(owner)
		if !ok {
			return -2
		}
		// held must stay live for the whole test. Go's liveness analysis makes it
		// dead after its last use, so without this the baseline handle's own
		// finalizer runs mid-test and the count drops by one that has nothing to
		// do with the wrapper under observation.
		defer runtime.KeepAlive(held)

		borrowed := NewRef[Shape2D](held.ToObject().(Shape2D))
		runtime.KeepAlive(borrowed)
		if held.ToObject().GetReferenceCount() != base {
			return -3
		}

		borrowed = nil
		runtime.GC()
		runtime.GC()

		after := held.ToObject().GetReferenceCount()
		fmt.Printf("  borrowed_ref: base=%d, after dropping borrower=%d\n", base, after)
		if after != base {
			fmt.Printf("  borrowed_ref: FAIL: borrowing wrapper released a reference it did not own\n")
			return -4
		}
		return 1
	})
}

// TestNullReturnSchedulesNoRelease covers the guard in NewRefTransfer. A
// get_shape() on an empty collision shape hands back a wrapper over no engine
// object; giving that wrapper a finalizer would issue an unreference ptrcall
// against a null owner. The assertion is that this survives collection.
func (e *Example) TestNullReturnSchedulesNoRelease(owner CollisionShape2D) int32 {
	return objectArgGuard("null_return_no_release", func() int32 {
		if isNullObject(owner) {
			return -1
		}
		owner.SetShape(nil)
		runtime.GC()
		runtime.GC()
		if !isNullRef(owner.GetShape()) {
			return -2
		}

		for i := 0; i < 50; i++ {
			_ = owner.GetShape()
		}
		runtime.GC()
		runtime.GC()

		fmt.Printf("  null_return: 50 null returns survived collection\n")
		return 1
	})
}

// TestReturnOwnershipRelease runs collection while the engine is still up so no
// Go-held reference survives into the engine's exit leak check, and so no
// finalizer races engine teardown.
func (e *Example) TestReturnOwnershipRelease() int32 {
	return objectArgGuard("return_ownership_release", func() int32 {
		runtime.GC()
		runtime.GC()
		return 1
	})
}
