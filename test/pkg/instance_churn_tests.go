package pkg

// In-engine regression gates for instance-scoped Go retention (openspec
// change fix-instance-handle-lifecycle, issue #152).
//
// Before the change, every engine-freed user-defined instance kept its
// cgo.Handle and package-pinner pins alive forever: weak references to the
// Go wrappers survived engine free with a 100% rate and live Go heap grew
// linearly with the churn count (~270 B per instance measured pre-fix,
// matching the issue report). Post-fix, only the runtime's own
// handle-table bookkeeping floor remains (~60 B per instance, also present
// in the issue's with-Delete reproduction), the wrappers are collectable,
// and each phase's delta stays inside the band.
//
// Two churn shapes are gated:
//
//  1. User-defined instances: GDScript `ExampleRef.new()` → engine create
//     callback (CreateInstance2 → WrappedPostInitialize) → the local drops
//     at the end of each loop iteration → refcount 0 → engine free →
//     FreeInstance (which now deletes the Handle). Weak survival must be
//     zero and heap deltas must stay in band.
//
//  2. Engine-class bindings: GDScript `RefCounted.new()` → Go resolves via
//     ObjectFromInstanceBinding (lazy binding create → Handle) → drop →
//     ~Object → GoCallback_GDExtensionBindingFree. Wrapper weak/heap
//     assertions cannot isolate this release because the generated
//     New*Owner constructors still pin wrappers on the package pinner
//     (explicit non-goal — follow-up change), so this gate asserts release
//     accounting instead: every binding created during churn must be freed
//     exactly once, leaving the create/free imbalance unchanged.

import (
	"fmt"
	"runtime"
	"weak"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
	. "github.com/godot-go/godot-go/pkg/gdclassimpl"
	"github.com/godot-go/godot-go/pkg/gdclassinit"
)

const (
	churnCaptureCap = 2048
	// churnBandPerInstance bounds a phase's retained Go heap by batch size.
	// Pre-fix retention measured ~270 B per created instance (wrapper graph
	// held by the leaked Handle); the post-fix floor is the runtime's own
	// handle-table bookkeeping at ~60-65 B per created handle, linear in
	// creations and unavoidable (the issue's with-Delete reproduction shows
	// the same slope). 200 B sits between them: ~3x headroom over the floor,
	// clearly below the leak rate.
	churnBandPerInstance = 200
)

var (
	churnBaseHeap    uint64
	churnLiveBase    int
	churnUserWeak    []weak.Pointer[ExampleRef]
	churnPhaseDeltas []int64
	churnBindBase    int64 // create-free imbalance before churn
	churnBindingWeak []weak.Pointer[RefCountedImpl]
)

// TestInstanceChurnBegin records the baseline live heap and live-instance
// count before the churn batches (GDScript warms the paths first).
func (e *Example) TestInstanceChurnBegin() int32 {
	return objectArgGuard("instance_churn_begin", func() int32 {
		runtime.GC()
		runtime.GC()
		churnBaseHeap = liveHeap()
		churnLiveBase = len(Internal.GDClassInstances.Keys())
		churnUserWeak = churnUserWeak[:0]
		churnPhaseDeltas = churnPhaseDeltas[:0]
		return 1
	})
}

// TestInstanceChurnCapture resolves the Go wrapper of a freshly created
// ExampleRef and keeps a weak reference to it (up to the capture cap).
func (e *Example) TestInstanceChurnCapture(ref RefCounted) int32 {
	if isNullObject(ref) {
		return -1
	}
	owner := (*GodotObject)(ref.AsGDExtensionObjectPtr())
	w, err := ObjectFromInstanceBinding(owner)
	if err != nil {
		fmt.Printf("  instance_churn: capture failed: %v\n", err)
		return -2
	}
	er, ok := w.(*ExampleRef)
	if !ok || er == nil {
		fmt.Printf("  instance_churn: capture got %T, want *ExampleRef\n", w)
		return -3
	}
	if len(churnUserWeak) < churnCaptureCap {
		churnUserWeak = append(churnUserWeak, weak.Make(er))
	}
	return 1
}

// TestInstanceChurnEnd forces collection after a batch and asserts the live
// heap delta stays within churnBandPerInstance times the batch size.
func (e *Example) TestInstanceChurnEnd(phase int64, batch int64) int32 {
	return objectArgGuard("instance_churn_end", func() int32 {
		runtime.GC()
		runtime.GC()
		delta := int64(liveHeap()) - int64(churnBaseHeap)
		churnPhaseDeltas = append(churnPhaseDeltas, delta)
		live := len(Internal.GDClassInstances.Keys())
		band := batch * churnBandPerInstance
		fmt.Printf("  instance_churn: phase=%d heap_delta=%+d KiB live_instances=%d (base %d) band=%d KiB\n",
			phase, delta/1024, live, churnLiveBase, band/1024)
		if delta > band {
			fmt.Printf("  instance_churn: FAIL: phase %d retains %d KiB, band is %d KiB\n",
				phase, delta/1024, band/1024)
			return -4
		}
		return 1
	})
}

// TestInstanceChurnVerify asserts the post-churn end state: every captured
// wrapper is collectable, the live-instance count returned to its baseline,
// and retention does not scale with the batch size.
func (e *Example) TestInstanceChurnVerify() int32 {
	return objectArgGuard("instance_churn_verify", func() int32 {
		runtime.GC()
		runtime.GC()
		if len(churnPhaseDeltas) < 2 {
			fmt.Println("  instance_churn: FAIL: fewer than two phases recorded")
			return -2
		}
		alive := 0
		for i := range churnUserWeak {
			if churnUserWeak[i].Value() != nil {
				alive++
			}
		}
		small, large := churnPhaseDeltas[0], churnPhaseDeltas[len(churnPhaseDeltas)-1]
		fmt.Printf("  instance_churn: verify weak_alive=%d/%d deltas=%+d KiB..%+d KiB live=%d base=%d\n",
			alive, len(churnUserWeak), small/1024, large/1024,
			len(Internal.GDClassInstances.Keys()), churnLiveBase)
		if alive > 0 {
			fmt.Printf("  instance_churn: FAIL: %d freed-instance wrappers still reachable (handle or pins retained)\n", alive)
			return -3
		}
		if live := len(Internal.GDClassInstances.Keys()); live != churnLiveBase {
			fmt.Printf("  instance_churn: FAIL: live-instance count %d, want baseline %d\n", live, churnLiveBase)
			return -4
		}
		fmt.Printf("  instance_churn: freed instances reclaim their Go wrappers; heap stays flat\n")
		return 1
	})
}

// TestBindingChurnBegin records the engine-class binding create/free
// imbalance before churn.
func (e *Example) TestBindingChurnBegin() int32 {
	return objectArgGuard("binding_churn_begin", func() int32 {
		churnBindBase = int64(gdclassinit.BindingCreateCount()) - int64(gdclassinit.BindingFreeCount())
		churnBindingWeak = churnBindingWeak[:0]
		return 1
	})
}

// TestBindingChurnCapture forces the lazy engine-class binding create for a
// fresh native object and keeps a weak reference for diagnostics only.
func (e *Example) TestBindingChurnCapture(ref RefCounted) int32 {
	if isNullObject(ref) {
		return -1
	}
	owner := (*GodotObject)(ref.AsGDExtensionObjectPtr())
	w, err := ObjectFromInstanceBinding(owner)
	if err != nil {
		fmt.Printf("  binding_churn: capture failed: %v\n", err)
		return -2
	}
	ri, ok := w.(*RefCountedImpl)
	if !ok || ri == nil {
		fmt.Printf("  binding_churn: capture got %T, want *RefCountedImpl\n", w)
		return -3
	}
	if len(churnBindingWeak) < churnCaptureCap {
		churnBindingWeak = append(churnBindingWeak, weak.Make(ri))
	}
	return 1
}

// TestBindingChurnVerify requires that every binding created during churn was
// released exactly once: the create-free imbalance after churn must equal the
// pre-churn imbalance (long-lived bindings from other fixtures are unchanged).
func (e *Example) TestBindingChurnVerify() int32 {
	return objectArgGuard("binding_churn_verify", func() int32 {
		imbalance := int64(gdclassinit.BindingCreateCount()) - int64(gdclassinit.BindingFreeCount())
		fmt.Printf("  binding_churn: verify imbalance_after=%d (base %d) weak_alive=%d/%d\n",
			imbalance, churnBindBase, countAliveWeak(churnBindingWeak), len(churnBindingWeak))
		if imbalance != churnBindBase {
			fmt.Printf("  binding_churn: FAIL: %d churned binding handles were not released\n", imbalance-churnBindBase)
			return -2
		}
		fmt.Printf("  binding_churn: every churned engine-class binding handle was released\n")
		return 1
	})
}

func countAliveWeak[T any](ws []weak.Pointer[T]) int {
	n := 0
	for i := range ws {
		if ws[i].Value() != nil {
			n++
		}
	}
	return n
}
