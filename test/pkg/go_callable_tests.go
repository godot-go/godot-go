package pkg

import (
	"errors"
	"sync/atomic"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
)

// In-engine fixtures for the Go func -> Callable bridge (openspec change
// support-go-callable-bridge).
//
// Convention follows ref_method_boundary_tests.go: bound probe methods return
// 1 on success and negative values per failed assertion. The callable bodies
// themselves record into shared atomics that the probe methods read back,
// because GDScript can observe return values but not Go captures.

// goCallableInvocations counts executions of the probe callable body. Signal
// dedup scenarios distinguish one connection from two by the per-emission
// delta.
var goCallableInvocations int64

// goCallableStoredCopy holds the copy a callable made from a call argument at
// package scope so an explicit release, mirroring TestRefBoundaryRelease,
// drops it before the shutdown leak snapshot.
var (
	goCallableStoredCopy    Variant
	goCallableStoredCopySet bool
)

// CreateCounterCallable hands the script a Callable whose every invocation
// returns the next counter value. The counter lives in the closure capture,
// proving the handle resolves to the same entry across calls.
func (e *Example) CreateCounterCallable() Callable {
	var counter int64
	return NewGoCallable(func(args []Variant) (Variant, error) {
		n := atomic.AddInt64(&counter, 1)
		return NewVariantInt(int(n)), nil
	})
}

// CreateProbeCallable counts invocations into goCallableInvocations and
// returns the running total.
func (e *Example) CreateProbeCallable() Callable {
	return NewGoCallable(func(args []Variant) (Variant, error) {
		n := atomic.AddInt64(&goCallableInvocations, 1)
		return NewVariantInt(int(n)), nil
	})
}

// CreateErrorCallable always fails with a Go error; the engine must see a
// call error while the process stays up.
func (e *Example) CreateErrorCallable() Callable {
	return NewGoCallable(func(args []Variant) (Variant, error) {
		return Variant{}, errors.New("go callable fixture failure")
	})
}

// CreatePanicCallable always panics; the barrier must recover and report the
// same call error without aborting the process.
func (e *Example) CreatePanicCallable() Callable {
	return NewGoCallable(func(args []Variant) (Variant, error) {
		panic("go callable fixture panic")
	})
}

// CreateCopyStoredCallable pins the borrowed-argument contract: a call with
// arguments stores an explicit copy of args[0]; a call without arguments
// returns the stored copy. Reading back the original value after the
// originating call ended proves the copy outlived the borrow.
func (e *Example) CreateCopyStoredCallable() Callable {
	return NewGoCallable(func(args []Variant) (Variant, error) {
		if len(args) == 0 {
			if !goCallableStoredCopySet {
				return NewVariantGoString("<nothing stored>"), nil
			}
			return goCallableStoredCopy, nil
		}
		if goCallableStoredCopySet {
			goCallableStoredCopy.Destroy()
		}
		goCallableStoredCopy = NewVariantNativeCopy(args[0].NativeConstPtr())
		goCallableStoredCopySet = true
		return NewVariantInt(1), nil
	})
}

// ReleaseGoCallableStoredCopy drops the stored copy so the retained reference
// cannot read as a leak at shutdown. Called last in the callable suite.
func (e *Example) ReleaseGoCallableStoredCopy() int64 {
	if goCallableStoredCopySet {
		goCallableStoredCopy.Destroy()
		goCallableStoredCopySet = false
	}
	return 1
}

// CreateReentrantCallable returns a Callable whose body re-enters the engine
// through Object.call, which varcalls the Go-bound target method below: the
// full engine -> Go -> engine -> Go nesting inside one invocation.
func (e *Example) CreateReentrantCallable() Callable {
	return NewGoCallable(func(args []Variant) (Variant, error) {
		sn := NewStringNameWithLatin1Chars("callable_bind_target")
		defer sn.Destroy()
		v := e.Call(sn, NewVariantInt(41))
		n := v.ToInt64()
		v.Destroy()
		return NewVariantInt(int(n) + 1), nil
	})
}

// CallableBindTarget is the varcall target for CreateReentrantCallable.
func (e *Example) CallableBindTarget(value int64) int64 {
	return value
}

// GoCallableInvocationCount reads the shared probe counter.
func (e *Example) GoCallableInvocationCount() int64 {
	return atomic.LoadInt64(&goCallableInvocations)
}

// ResetGoCallableInvocationCount zeroes the shared probe counter.
func (e *Example) ResetGoCallableInvocationCount() int64 {
	atomic.StoreInt64(&goCallableInvocations, 0)
	return 1
}

// GoCallableRegistryCount exposes the live registry entry count.
func (e *Example) GoCallableRegistryCount() int64 {
	return GoCallableRegistryCount()
}
