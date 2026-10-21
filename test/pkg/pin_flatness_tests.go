package pkg

import (
	"fmt"
	"runtime"
)

// In-engine regression for call-scoped pinning (openspec change
// fix-generated-call-pin-leak).
//
// The pure-Go tests in pkg/gdclassimpl prove the runtime mechanism, but they do
// not drive real generated bodies. This does: it runs thousands of generated
// calls through the live engine and measures Go heap that survives a forced GC.
//
// Every generated call pins a return slot, an owner cell, an argument slice and
// its argument cells. Before the change those pins were taken on a package-global
// pinner that was never released, so live heap climbed with the iteration count
// and never came back down. With the pins scoped to the call body, the same loop
// returns to baseline.
//
// Only value-returning calls are used deliberately. A call that hands back a
// refcounted object or an engine-allocated string would exercise the separate
// reference-ownership question (see fix-object-return-reference-ownership) and
// could trip the engine's exit leak check for a reason this change does not
// govern.

// liveHeap reports bytes still allocated after the caller has forced collection.
func liveHeap() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc
}

// mixedGeneratedCalls drives a spread of generated ptrcall shapes: two integer
// returns and a boolean return, two of them carrying an argument, so the sample
// covers both the no-argument and one-argument pin layouts.
func (e *Example) mixedGeneratedCalls(n int64) {
	for i := int64(0); i < n; i++ {
		_ = e.GetChildCount(false)
		_ = e.IsInsideTree()
		_ = e.GetIndex(false)
	}
}

// TestPinScratchStaysFlat samples live Go heap after two differently sized bursts
// of generated calls. Retained pins make the delta grow with the iteration count;
// scoped pins keep it inside a constant band.
//
// Returns 1 on success, a negative code per failure mode, and -900 if the Go
// side panicked.
func (e *Example) TestPinScratchStaysFlat(small int64, large int64) int32 {
	return objectArgGuard("pin_flatness", func() int32 {
		if small <= 0 || large <= small {
			return -1
		}

		// measure runs one warm-up and one small/large pair, sampling live heap
		// after forced collection. The warm-up keeps first-call allocations out
		// of the baseline; the label distinguishes the confirming sample.
		measure := func(label string) (int64, int64) {
			e.mixedGeneratedCalls(250)
			runtime.GC()
			runtime.GC()
			base := int64(liveHeap())

			e.mixedGeneratedCalls(small)
			runtime.GC()
			runtime.GC()
			smallDelta := int64(liveHeap()) - base

			e.mixedGeneratedCalls(large)
			runtime.GC()
			runtime.GC()
			largeDelta := int64(liveHeap()) - base

			fmt.Printf("  pin_flatness: %sbase=%d KiB, +%d iters -> %+d KiB, +%d iters -> %+d KiB\n",
				label, base/1024, small, smallDelta/1024, large, largeDelta/1024)
			return smallDelta, largeDelta
		}

		// A permanently-pinned call retains on the order of a hundred bytes of Go
		// heap, so tens of thousands of calls drift by megabytes. Anything inside
		// this band is allocator and cache noise rather than retained pins.
		const band = 1 << 20

		// The absolute band alone could be passed by luck, so also require that the
		// larger burst does not scale with its iteration count relative to the
		// smaller one.
		const slack = band / 4

		smallDelta, largeDelta := measure("")

		// A heap can step once for reasons unrelated to the iteration count:
		// darwin/arm64 grew 416 KiB in one sample yet retained nothing over 40k
		// iterations on the next. Those steps survive the forced collections but
		// do not repeat, while retained pins accumulate in every burst. So an
		// out-of-band sample is confirmed by a second measurement before it fails.
		if smallDelta > band || largeDelta > band || largeDelta > smallDelta*3+slack {
			fmt.Println("  pin_flatness: first sample out of band; re-measuring before failing")
			smallDelta, largeDelta = measure("re-measured ")
		}

		if smallDelta > band || largeDelta > band {
			fmt.Printf("  pin_flatness: FAIL: drift exceeds %d KiB band, pins are being retained\n",
				band/1024)
			return -2
		}

		if largeDelta > smallDelta*3+slack {
			fmt.Printf("  pin_flatness: FAIL: drift scales with iterations (+%d KiB for %d iters vs +%d KiB for %d)\n",
				smallDelta/1024, small, largeDelta/1024, large)
			return -3
		}

		return 1
	})
}
