package pkg

// In-engine retention and teardown gates for the class property list
// (openspec change fix-class-property-list-lifecycle).
//
// Before the change, one Go-side array was published on every
// get_property_list call and the paired free_property_list2 callback
// destroyed its element strings and discarded the list, so list-declared
// custom properties vanished from every query after the first. Post-change,
// each query owns a fresh C array (released by the free callback) and the
// class owns the string cells for the registration lifetime.
//
// The query arrays live in C memory, so runtime.ReadMemStats cannot see a
// missing release; like the binding-churn gate before it, this fixture
// asserts release accounting (pkg/core PropertyArrayAllocCount/FreeCount)
// and keeps a Go-heap band to prove the query path retains no Go memory
// (the pre-fix pnr.Pin-on-publish pattern accumulated pinner bookkeeping).

import (
	"fmt"
	"runtime"

	. "github.com/godot-go/godot-go/pkg/core"
)

// propListBandPerQuery bounds retained Go heap per query cycle. Nothing on
// the query path survives collection; the band only absorbs runtime noise,
// sized on the same headroom logic as churnBandPerInstance.
const propListBandPerQuery = 64

var (
	propChurnBaseHeap      uint64
	propChurnBaseImbalance int64
)

// TestPropListTeardownCheck exercises the cache lifecycle directly: build →
// zero / one / many query cycles → teardown, with allocator abort on any
// double-free (task 2.2 verification).
func (e *Example) TestPropListTeardownCheck() int32 {
	return objectArgGuard("prop_list_teardown_check", func() int32 {
		if err := RunPropertyCacheTeardownCheck(); err != nil {
			fmt.Printf("  prop_list_teardown: FAIL: cache lifecycle check unclean: %v\n", err)
			return -2
		}
		fmt.Println("  prop_list_teardown: zero/one/many-query teardown clean, arrays accounted, destroy idempotent")
		return 1
	})
}

// TestPropListChurnBegin records the baseline heap and array create/free
// imbalance before the GDScript query loop (which warms up first).
func (e *Example) TestPropListChurnBegin() int32 {
	return objectArgGuard("prop_list_churn_begin", func() int32 {
		runtime.GC()
		runtime.GC()
		propChurnBaseHeap = liveHeap()
		propChurnBaseImbalance = PropertyArrayAllocCount() - PropertyArrayFreeCount()
		return 1
	})
}

// TestPropListChurnEnd requires every churned query array to have been
// released by the paired free_property_list2 call, and the Go heap delta to
// stay inside the per-batch band.
func (e *Example) TestPropListChurnEnd(batch int64) int32 {
	return objectArgGuard("prop_list_churn_end", func() int32 {
		runtime.GC()
		runtime.GC()
		imbalance := PropertyArrayAllocCount() - PropertyArrayFreeCount()
		delta := int64(liveHeap()) - int64(propChurnBaseHeap)
		band := batch * propListBandPerQuery
		fmt.Printf("  prop_list_churn: arrays_outstanding=%+d (base %+d) heap_delta=%+d KiB band=%d KiB\n",
			imbalance, propChurnBaseImbalance, delta/1024, band/1024)
		if imbalance != propChurnBaseImbalance {
			fmt.Printf("  prop_list_churn: FAIL: %d query arrays were not released by free_property_list2\n",
				imbalance-propChurnBaseImbalance)
			return -2
		}
		if delta > band {
			fmt.Printf("  prop_list_churn: FAIL: churn retains %d KiB Go heap, band is %d KiB\n",
				delta/1024, band/1024)
			return -3
		}
		fmt.Println("  prop_list_churn: every query array released; no Go retention")
		return 1
	})
}
