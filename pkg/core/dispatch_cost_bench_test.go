package core

import (
	"encoding/binary"
	"runtime/cgo"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/util"
)

// trim-dispatch-overhead task 1.1 — engine-free wall-clock baseline (Benchmark tier).
//
// These are the timing counterpart to typed_dispatch_budget_test.go: that file
// gates allocations, this one records ns/op for the same dispatch shapes so the
// change has a before/after to measure against (the repository had zero
// Benchmark* functions before this change). Per the dispatch-cost-baseline
// capability, timing is recorded, never gated — run these explicitly with:
//
//	go test ./pkg/core/ -bench 'Dispatch' -benchmem
//
// Every shape here is engine-free, matching the budget tier: compiled ptrcall
// closures run concrete loads, a direct handler call, and a concrete store over
// the spotStubClass/spotRetStub pair. What the engine-free tier cannot reach —
// the varcall closures' argument decode and Variant return encode, the
// echo-suppressed copy-return with a real builtin return, and the real
// engine→Go crossings — is recorded in-engine by the task 1.4 fixture instead
// (see the split documented in typed_dispatch_budget_test.go's header).
//
// The seam cells drive GoMethodMetadata.Ptrcall / CallWithError with Func left
// the zero reflect.Value, mirroring the budget seam: if the seam ever routes
// back through the reflective body the benchmark panics on the zero Func
// instead of silently measuring the old path. The varcall seam measures the
// CallWithError selection and hand-off — the generated varcall closure itself
// dereferences the engine table on its final Variant encode, so the hand-off is
// timed through a shape-identical stand-in closure.

// benchSink absorbs handler results so the compiler cannot delete the calls it
// is supposed to be timing.
var (
	benchSinkI64  int64
	benchSinkStr  string
	benchSinkBool bool
	benchSinkVar  Variant
)

// benchDispatchCells builds the shared raw argument cells once per benchmark;
// the engine hands dispatch a stable pointer array per call site, so the
// benchmark lifecycle matches production (everything built outside the loop).
func benchDispatchCells() (args0 []GDExtensionConstTypePtr, args2, args5 []GDExtensionConstTypePtr) {
	var cell [8]byte
	binary.LittleEndian.PutUint64(cell[:], 3)
	arg := GDExtensionConstTypePtr(&cell[0])
	return []GDExtensionConstTypePtr{},
		[]GDExtensionConstTypePtr{arg, arg},
		[]GDExtensionConstTypePtr{arg, arg, arg, arg, arg}
}

// --- instance ptrcall family ---

func BenchmarkDispatchInstanceValueRet0(b *testing.B) {
	args0, _, _ := benchDispatchCells()
	stub := &spotStubClass{}
	value0 := wrapPtrcallValue0[spotStubClass, *spotStubClass](
		func(self *spotStubClass) int64 { return 7 })
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value0(stub, args0, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

func BenchmarkDispatchInstanceValueRet2(b *testing.B) {
	_, args2, _ := benchDispatchCells()
	stub := &spotStubClass{}
	value2 := wrapPtrcallValue2[spotStubClass, *spotStubClass](
		func(self *spotStubClass, n int64, f float64) int64 { return n + int64(f) })
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value2(stub, args2, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

func BenchmarkDispatchInstanceValueRet5(b *testing.B) {
	_, _, args5 := benchDispatchCells()
	stub := &spotStubClass{}
	value5 := wrapPtrcallValue5[spotStubClass, *spotStubClass](
		func(self *spotStubClass, a, c, d, e, f int64) int64 { return a + c + d + e + f })
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value5(stub, args5, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

func BenchmarkDispatchInstanceVoidRet0(b *testing.B) {
	args0, _, _ := benchDispatchCells()
	stub := &spotStubClass{}
	none0 := wrapPtrcallNone0[spotStubClass, *spotStubClass](
		func(self *spotStubClass) { benchSinkI64++ })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		none0(stub, args0, nil)
	}
}

func BenchmarkDispatchInstanceVoidRet2(b *testing.B) {
	_, args2, _ := benchDispatchCells()
	stub := &spotStubClass{}
	none2 := wrapPtrcallNone2[spotStubClass, *spotStubClass](
		func(self *spotStubClass, n int64, f float64) { benchSinkI64 += n + int64(f) })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		none2(stub, args2, nil)
	}
}

func BenchmarkDispatchInstanceVoidRet5(b *testing.B) {
	_, _, args5 := benchDispatchCells()
	stub := &spotStubClass{}
	none5 := wrapPtrcallNone5[spotStubClass, *spotStubClass](
		func(self *spotStubClass, a, c, d, e, f int64) { benchSinkI64 += a + c + d + e + f })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		none5(stub, args5, nil)
	}
}

// --- instance copy-return family (always-destroy path; echo compare nil) ---

func BenchmarkDispatchInstanceCopyRet0(b *testing.B) {
	args0, _, _ := benchDispatchCells()
	stub := &spotStubClass{}
	copy0 := wrapPtrcallCopyRet0[spotStubClass, *spotStubClass](
		func(self *spotStubClass) spotRetStub {
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], 9)
			return out
		})
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy0(stub, args0, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

func BenchmarkDispatchInstanceCopyRet1(b *testing.B) {
	stub := &spotStubClass{}
	var cell [8]byte
	binary.LittleEndian.PutUint64(cell[:], 3)
	args1 := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(&cell[0])}
	copy1 := wrapPtrcallCopyRet1[spotStubClass, *spotStubClass, int64, spotRetStub](
		func(self *spotStubClass, n int64) spotRetStub {
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], uint64(n))
			return out
		})
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy1(stub, args1, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

func BenchmarkDispatchInstanceCopyRet2(b *testing.B) {
	_, args2, _ := benchDispatchCells()
	stub := &spotStubClass{}
	copy2 := wrapPtrcallCopyRet2[spotStubClass, *spotStubClass, int64, float64, spotRetStub](
		func(self *spotStubClass, n int64, f float64) spotRetStub {
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], uint64(n))
			return out
		})
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy2(stub, args2, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

// --- static ptrcall family ---

func BenchmarkDispatchStaticValueRet0(b *testing.B) {
	args0, _, _ := benchDispatchCells()
	value0 := wrapStaticPtrcallValue0(func() int64 { return 1 })
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value0(nil, args0, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

func BenchmarkDispatchStaticValueRet2(b *testing.B) {
	_, args2, _ := benchDispatchCells()
	value2 := wrapStaticPtrcallValue2(func(n int64, f float64) int64 { return n + int64(f) })
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value2(nil, args2, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

func BenchmarkDispatchStaticValueRet5(b *testing.B) {
	_, _, args5 := benchDispatchCells()
	value5 := wrapStaticPtrcallValue5(func(a, c, d, e, f int64) int64 { return a + c + d + e + f })
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		value5(nil, args5, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

func BenchmarkDispatchStaticVoidRet0(b *testing.B) {
	args0, _, _ := benchDispatchCells()
	none0 := wrapStaticPtrcallNone0(func() { benchSinkI64++ })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		none0(nil, args0, nil)
	}
}

func BenchmarkDispatchStaticVoidRet2(b *testing.B) {
	_, args2, _ := benchDispatchCells()
	none2 := wrapStaticPtrcallNone2(func(n int64, f float64) { benchSinkI64 += n + int64(f) })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		none2(nil, args2, nil)
	}
}

func BenchmarkDispatchStaticVoidRet5(b *testing.B) {
	_, _, args5 := benchDispatchCells()
	none5 := wrapStaticPtrcallNone5(func(a, c, d, e, f int64) { benchSinkI64 += a + c + d + e + f })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		none5(nil, args5, nil)
	}
}

func BenchmarkDispatchStaticCopyRet0(b *testing.B) {
	args0, _, _ := benchDispatchCells()
	copy0 := wrapStaticPtrcallCopyRet0(func() spotRetStub {
		var out spotRetStub
		binary.LittleEndian.PutUint64(out[:], 9)
		return out
	})
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy0(nil, args0, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

func BenchmarkDispatchStaticCopyRet1(b *testing.B) {
	var cell [8]byte
	binary.LittleEndian.PutUint64(cell[:], 3)
	args1 := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(&cell[0])}
	copy1 := wrapStaticPtrcallCopyRet1(func(n int64) spotRetStub {
		var out spotRetStub
		binary.LittleEndian.PutUint64(out[:], uint64(n))
		return out
	})
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy1(nil, args1, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

func BenchmarkDispatchStaticCopyRet2(b *testing.B) {
	_, args2, _ := benchDispatchCells()
	copy2 := wrapStaticPtrcallCopyRet2(func(n int64, f float64) spotRetStub {
		var out spotRetStub
		binary.LittleEndian.PutUint64(out[:], uint64(n))
		return out
	})
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		copy2(nil, args2, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

// --- the six callback closure shapes (D6 dedicated binds) ---

func BenchmarkDispatchNotificationClosure(b *testing.B) {
	stub := &spotStubClass{}
	cb := newNotificationShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, what int32, reversed bool) { benchSinkI64 += int64(what) })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb(stub, 16, false)
	}
}

func BenchmarkDispatchPropGetClosure(b *testing.B) {
	stub := &spotStubClass{}
	planted := plantedVariant(42)
	cb := newPropGetShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name string) (Variant, bool) { return planted, true })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, ok := cb(stub, "dproperty_1")
		benchSinkVar = v
		benchSinkBool = ok
	}
}

func BenchmarkDispatchPropSetClosure(b *testing.B) {
	stub := &spotStubClass{}
	planted := plantedVariant(7)
	cb := newPropSetShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name string, value Variant) bool { return true })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkBool = cb(stub, "dproperty_2", planted)
	}
}

func BenchmarkDispatchPropCanRevertClosure(b *testing.B) {
	stub := &spotStubClass{}
	name := StringName{1, 2, 3, 4, 5, 6, 7, 8}
	cb := newPropCanRevertShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, n StringName) bool { return true })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkBool = cb(stub, name)
	}
}

func BenchmarkDispatchPropGetRevertClosure(b *testing.B) {
	stub := &spotStubClass{}
	planted := plantedVariant(9)
	cb := newPropGetRevertShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass, name StringName) (Variant, bool) { return planted, true })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, ok := cb(stub, StringName{})
		benchSinkVar = v
		benchSinkBool = ok
	}
}

func BenchmarkDispatchToStringClosure(b *testing.B) {
	stub := &spotStubClass{}
	cb := newToStringShape[spotStubClass, *spotStubClass](
		func(self *spotStubClass) string { return "[ GDExtension::Stub <--> Instance ID:1 ]" })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchSinkStr = cb(stub)
	}
}

// --- callback prologue resolve-and-dispatch (task 1.2) ---

// benchPrologueObject satisfies Object for the bench instance without an
// engine: every engine method stays on the embedded nil interface, so an
// accidental crossing faults fast instead of silently measuring nothing. The
// prologue path exercises only GetClassName (the Go-side identity, standing in
// for today's engine GetClass() crossing — see the cells' caveat) and the
// compiled callback's self assertion.
type benchPrologueObject struct {
	Object
}

// Self is the selfTyped one-liner the wrapper contract requires.
func (b *benchPrologueObject) Self() *benchPrologueObject { return b }

func (b *benchPrologueObject) GetClassName() string { return "SpotStub" }

var _ Object = (*benchPrologueObject)(nil)

// benchPrologueStub installs a SpotStub registration carrying compiled
// notification and property-get callbacks and returns the registry restore.
// Mirrors the shape the dedicated binds attach (cbNotification / cbPropGet on
// the metadata, VirtualMethodMap routing under "_notification" / "_get").
func benchPrologueStub() (restore func()) {
	original := Internal.GDRegisteredGDClasses
	registry := util.NewSyncMap[string, *ClassInfo]()
	planted := plantedVariant(42)
	registry.Set("SpotStub", &ClassInfo{
		Name: "SpotStub",
		VirtualMethodMap: map[string]*MethodBindAndClassMethodInfo{
			"_notification": {GoMethodMetadata: &GoMethodMetadata{
				ClassName:    "SpotStub",
				GdMethodName: "_notification",
				GoMethodName: "V_Notification",
				cbNotification: newNotificationShape[benchPrologueObject, *benchPrologueObject](
					func(self *benchPrologueObject, what int32, reversed bool) { benchSinkI64 += int64(what) }),
			}},
			"_get": {GoMethodMetadata: &GoMethodMetadata{
				ClassName:    "SpotStub",
				GdMethodName: "_get",
				GoMethodName: "V_Get",
				cbPropGet: newPropGetShape[benchPrologueObject, *benchPrologueObject](
					func(self *benchPrologueObject, name string) (Variant, bool) { return planted, true }),
			}},
		},
	})
	Internal.GDRegisteredGDClasses = registry
	return func() { Internal.GDRegisteredGDClasses = original }
}

// BenchmarkDispatchPrologueNotification reproduces the resolve-and-dispatch
// part of GoCallback_ClassCreationInfoNotification's prologue engine-free:
// the cgo handle dereference, class-name acquisition, the registry lookup,
// the VirtualMethodMap lookup, and the compiled callback dispatch.
//
// Engine-free reach caveat (recorded in the task evidence): the name is read
// from the Go-side GetClassName() — the instance-held identity this change
// plans to provide — because the CURRENT prologue's engine GetClass() +
// ToUtf8() + Destroy() ceremony cannot execute without a live engine. These
// cells therefore measure the shared floor of before/after, NOT today's full
// prologue; the engine ceremony itself is recorded in-engine by task 1.4.
func BenchmarkDispatchPrologueNotification(b *testing.B) {
	restore := benchPrologueStub()
	defer restore()
	inst := &benchPrologueObject{}
	wci := &WrappedClassInstance{Instance: inst}
	handle := cgo.NewHandle(wci)
	defer handle.Delete()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := handle.Value().(*WrappedClassInstance)
		ci, ok := Internal.GDRegisteredGDClasses.Get(w.Instance.GetClassName())
		if !ok {
			b.Fatal("SpotStub not resolvable in the bench registry")
		}
		mcmi, ok := ci.VirtualMethodMap["_notification"]
		if !ok {
			b.Fatal("_notification not registered")
		}
		cb := mcmi.GoMethodMetadata.cbNotification
		if cb == nil {
			b.Fatal("compiled notification callback missing")
		}
		cb(w.Instance, 16, false)
	}
}

// BenchmarkDispatchProloguePropGet reproduces the resolve-and-dispatch part of
// GoCallback_ClassCreationInfoGet's prologue with the same engine-free reach
// caveat as the notification cell (the current prologue's StringName
// ToUtf8() and engine GetClass() crossings are recorded in-engine, task 1.4;
// the resolved name is supplied as a Go string, which is what cbPropGet
// receives either way).
func BenchmarkDispatchProloguePropGet(b *testing.B) {
	restore := benchPrologueStub()
	defer restore()
	inst := &benchPrologueObject{}
	wci := &WrappedClassInstance{Instance: inst}
	handle := cgo.NewHandle(wci)
	defer handle.Delete()
	const name = "dproperty_1"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := handle.Value().(*WrappedClassInstance)
		ci, ok := Internal.GDRegisteredGDClasses.Get(w.Instance.GetClassName())
		if !ok {
			b.Fatal("SpotStub not resolvable in the bench registry")
		}
		mcmi, ok := ci.VirtualMethodMap["_get"]
		if !ok {
			b.Fatal("_get not registered")
		}
		cb := mcmi.GoMethodMetadata.cbPropGet
		if cb == nil {
			b.Fatal("compiled prop-get callback missing")
		}
		v, found := cb(w.Instance, name)
		benchSinkVar = v
		benchSinkBool = found
	}
}

// --- metadata seam (Func stays the zero reflect.Value) ---

func BenchmarkDispatchSeamValueRet2(b *testing.B) {
	_, args2, _ := benchDispatchCells()
	stub := &spotStubClass{}
	md := &GoMethodMetadata{
		ClassName:    "SpotStub",
		GdMethodName: "bench_value",
		GoMethodName: "BenchValue",
	}
	md.typedDispatch = wrapPtrcallValue2[spotStubClass, *spotStubClass](
		func(self *spotStubClass, n int64, f float64) int64 { return n + int64(f) })
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		md.Ptrcall(stub, args2, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

func BenchmarkDispatchSeamVoidRet5(b *testing.B) {
	_, _, args5 := benchDispatchCells()
	stub := &spotStubClass{}
	md := &GoMethodMetadata{
		ClassName:    "SpotStub",
		GdMethodName: "bench_void",
		GoMethodName: "BenchVoid",
	}
	md.typedDispatch = wrapPtrcallNone5[spotStubClass, *spotStubClass](
		func(self *spotStubClass, a, c, d, e, f int64) { benchSinkI64 += a + c + d + e + f })
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		md.Ptrcall(stub, args5, nil)
	}
}

func BenchmarkDispatchSeamCopyRet2(b *testing.B) {
	_, args2, _ := benchDispatchCells()
	stub := &spotStubClass{}
	md := &GoMethodMetadata{
		ClassName:    "SpotStub",
		GdMethodName: "bench_copy",
		GoMethodName: "BenchCopy",
	}
	md.typedDispatch = wrapPtrcallCopyRet2[spotStubClass, *spotStubClass, int64, float64, spotRetStub](
		func(self *spotStubClass, n int64, f float64) spotRetStub {
			var out spotRetStub
			binary.LittleEndian.PutUint64(out[:], uint64(n))
			return out
		})
	var ret [8]byte
	retPtr := GDExtensionUninitializedTypePtr(&ret[0])
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		md.Ptrcall(stub, args2, retPtr)
	}
	benchSinkI64 = int64(ret[0])
}

// BenchmarkDispatchSeamVarcallValueRet0 times the CallWithError varcall seam:
// the nil-check and hand-off to the compiled varcall closure. The closure here
// is a shape-identical stand-in because every generated varcall wrapper ends in
// a Variant return encode that dereferences the engine table (documented at
// method_bind_varcall_typed_test.go); the full typed-varcall cost per call is
// recorded by the in-engine timing fixture. Func stays zero: a seam that routed
// back to the reflective body would panic here instead of measuring.
func BenchmarkDispatchSeamVarcallValueRet0(b *testing.B) {
	stub := &spotStubClass{}
	var sentinel Variant
	md := &GoMethodMetadata{
		ClassName:    "SpotStub",
		GdMethodName: "bench_varcall",
		GoMethodName: "BenchVarcall",
		typedVarcall: func(inst GDClass, args []Variant) (Variant, error) {
			return sentinel, nil
		},
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		v, _ := md.CallWithError(stub)
		benchSinkVar = v
	}
}
