package core

import (
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"
	"testing"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// typed-method-dispatch task 2.2 — per-wrapper-kind differential and budget
// baseline (engine-free unit tier).
//
// Every generated wrapper kind runs twice against the SAME handler: once through
// the compiled closure, once through the reflective body the name-string API
// dispatches with (Func plus the metadata that path builds). Both must hand the
// handler identical argument values and write identical return bytes; the typed
// closure must then hold the zero-allocation budget. The handler is shared on
// purpose — a difference between the two runs can only come from dispatch.
//
// Engine-free reach. The reflective body decodes builtin arguments through their
// engine copy constructors (NewVector2WithVector2 and friends) and encodes
// copy-return results through the copy-ctor family, so the reflective twin can
// execute only for the raw-load scalar kinds with value or void returns. The
// kinds that cross into the engine are pinned where they can be observed:
//
//   - copy-return wrappers — encoded bytes plus the destroy gate in
//     TestTypedWrapperInstanceCopyRet and TestTypedWrapperStaticCopyRetArity0,
//     borrow-echo classification equality in method_bind_borrow_echo_typed_test.go,
//     and refcount effects under the shutdown leak gates by the demo's typed-bound
//     echoes (tasks 1.4, 5.2, 7.1);
//   - builtin and String arguments/returns — the demo's typed-bound vectors,
//     containers and StringName/NodePath echoes run the unchanged GDScript
//     suites, which passed identically before and after the typed path existed,
//     and task 7.2's compiled-vs-reflective margin keeps a reflective hop on a
//     compiled path from passing silently.
//
// Budgets are asserted through measureDispatch (cap dispatchAllocCap, shared
// with the task 7.2 arity matrix); each kind's reflective figure is logged
// beside it, which is the contrast that makes the zero meaningful.

// observed records what the most recent handler call received, as raw bit
// patterns in a fixed array. No formatting, no boxing, no append: the recording
// must cost the zero-allocation budget nothing, or the probe would book its own
// formatting work to the dispatch path (the attribution trap task 7.2 hit).
var (
	observed      [5]uint64
	observedCount int
)

func observeBool(slot int, v bool) {
	var bits uint64
	if v {
		bits = 1
	}
	observeSlot(slot, bits)
}

func observeInt(slot int, v int64) { observeSlot(slot, uint64(v)) }

func observeFloat(slot int, v float64) { observeSlot(slot, math.Float64bits(v)) }

func observeSlot(slot int, bits uint64) {
	observed[slot] = bits
	if slot+1 > observedCount {
		observedCount = slot + 1
	}
}

func resetObserved() {
	observed = [5]uint64{}
	observedCount = 0
}

func snapshotObserved() string {
	return fmt.Sprint(observed[:observedCount])
}

// The recording handlers below are bound through both paths, so an observed
// difference is a dispatch difference. Values are chosen so a mis-decoded slot
// (wrong width, wrong sign, wrong type) changes the recording.

func (s *spotStubClass) kindValue0() int64 {
	return 7
}

func (s *spotStubClass) kindValue1(n int64) int64 {
	observeInt(0, n)
	return n * 3
}

func (s *spotStubClass) kindValue2(a int64, b float64) float64 {
	observeInt(0, a)
	observeFloat(1, b)
	return float64(a) + b
}

func (s *spotStubClass) kindValue5(a bool, b int32, c uint64, d float64, e int64) int64 {
	observeBool(0, a)
	observeInt(1, int64(b))
	observeInt(2, int64(c))
	observeFloat(3, d)
	observeInt(4, e)
	if a {
		return int64(b) + int64(c) + int64(d) + e
	}
	return -int64(b)
}

func (s *spotStubClass) kindVoid0() {}

func (s *spotStubClass) kindVoid1(n int64) { observeInt(0, n) }

func (s *spotStubClass) kindVoid2(a int64, b bool) {
	observeInt(0, a)
	observeBool(1, b)
}

func (s *spotStubClass) kindVoid5(a bool, b int32, c uint64, d float64, e int64) {
	observeBool(0, a)
	observeInt(1, int64(b))
	observeInt(2, int64(c))
	observeFloat(3, d)
	observeInt(4, e)
}

// The static twins carry no receiver, matching the static wrapper family.
func kindStaticValue0() float64 { return -1.5 }

func kindStaticValue2(a int64, b bool) int64 {
	observeInt(0, a)
	observeBool(1, b)
	if b {
		return a + 1
	}
	return a - 1
}

func kindStaticVoid0() {}

func kindStaticVoid2(d float64, c uint64) {
	observeFloat(0, d)
	observeInt(1, int64(c))
}

// kindCopyRet1 returns the stub copy-encodable value; its reflective twin would
// need the engine copy constructor, so this kind is budgeted here and its
// encode/destroy contract is pinned by the copy-ret tests named above.
func (s *spotStubClass) kindCopyRet1(n int64) spotRetStub {
	observeInt(0, n)
	var out spotRetStub
	binary.LittleEndian.PutUint64(out[:], uint64(n))
	return out
}

// Argument cells, planted once: every kind reads the same bytes on both paths.
var (
	cellTrue = true
	cellI32  = int32(-7)
	cellU64  = uint64(1) << 40
	cellF64  = float64(2.5)
	cellI64  = int64(11)
)

// wrapperKind pairs one generated wrapper kind with the reflective metadata the
// name-string path would build for the same handler. A nil reflective twin means
// the kind's decode or encode crosses into the engine (see the file header).
type wrapperKind struct {
	name    string
	typed   ptrcallFn
	reflect *GoMethodMetadata
	inst    GDClass
	args    []GDExtensionConstTypePtr
}

func instanceMD(goName string, fn any, retType reflect.Type, style ReturnStyle, argTypes ...reflect.Type) *GoMethodMetadata {
	// The reflective body's log line reads the advertised Variant type per
	// argument, so the twin carries the mapping NewGoMethodMetadata computes.
	gdeArgumentTypes := make([]GDExtensionVariantType, len(argTypes))
	for i, argType := range argTypes {
		gdeArgumentTypes[i] = ReflectTypeToGDExtensionVariantType(argType)
	}
	return &GoMethodMetadata{
		ClassName:        "SpotStub",
		GdMethodName:     strings.ToLower(goName),
		GoMethodName:     goName,
		Func:             reflect.ValueOf(fn),
		GoReturnType:     retType,
		GoReturnStyle:    style,
		GoArgumentTypes:  argTypes,
		gdeArgumentTypes: gdeArgumentTypes,
	}
}

func staticMD(goName string, fn any, retType reflect.Type, style ReturnStyle, argTypes ...reflect.Type) *GoMethodMetadata {
	md := instanceMD(goName, fn, retType, style, argTypes...)
	md.MethodFlags = METHOD_FLAG_STATIC
	return md
}

func wrapperKinds() []wrapperKind {
	boolT := reflect.TypeFor[bool]()
	i32T := reflect.TypeFor[int32]()
	u64T := reflect.TypeFor[uint64]()
	f64T := reflect.TypeFor[float64]()
	i64T := reflect.TypeFor[int64]()

	args0 := []GDExtensionConstTypePtr{}
	args1 := []GDExtensionConstTypePtr{ptrTo(&cellI64)}
	args2if := []GDExtensionConstTypePtr{ptrTo(&cellI64), ptrTo(&cellF64)}
	args2ib := []GDExtensionConstTypePtr{ptrTo(&cellI64), ptrTo(&cellTrue)}
	args5 := []GDExtensionConstTypePtr{ptrTo(&cellTrue), ptrTo(&cellI32), ptrTo(&cellU64), ptrTo(&cellF64), ptrTo(&cellI64)}
	var stub GDClass = &spotStubClass{}

	return []wrapperKind{
		// Instance value-return family.
		{"instance value arity-0",
			wrapPtrcallValue0[spotStubClass, *spotStubClass]((*spotStubClass).kindValue0),
			instanceMD("KindValue0", (*spotStubClass).kindValue0, i64T, ValueReturnStyle),
			stub, args0},
		{"instance value arity-1",
			wrapPtrcallValue1[spotStubClass, *spotStubClass]((*spotStubClass).kindValue1),
			instanceMD("KindValue1", (*spotStubClass).kindValue1, i64T, ValueReturnStyle, i64T),
			stub, args1},
		{"instance value arity-2",
			wrapPtrcallValue2[spotStubClass, *spotStubClass]((*spotStubClass).kindValue2),
			instanceMD("KindValue2", (*spotStubClass).kindValue2, f64T, ValueReturnStyle, i64T, f64T),
			stub, args2if},
		{"instance value arity-5 mixed kinds",
			wrapPtrcallValue5[spotStubClass, *spotStubClass]((*spotStubClass).kindValue5),
			instanceMD("KindValue5", (*spotStubClass).kindValue5, i64T, ValueReturnStyle, boolT, i32T, u64T, f64T, i64T),
			stub, args5},

		// Instance void-return family.
		{"instance void arity-0",
			wrapPtrcallNone0[spotStubClass, *spotStubClass]((*spotStubClass).kindVoid0),
			instanceMD("KindVoid0", (*spotStubClass).kindVoid0, nil, NoneReturnStyle),
			stub, args0},
		{"instance void arity-1",
			wrapPtrcallNone1[spotStubClass, *spotStubClass]((*spotStubClass).kindVoid1),
			instanceMD("KindVoid1", (*spotStubClass).kindVoid1, nil, NoneReturnStyle, i64T),
			stub, args1},
		{"instance void arity-2",
			wrapPtrcallNone2[spotStubClass, *spotStubClass]((*spotStubClass).kindVoid2),
			instanceMD("KindVoid2", (*spotStubClass).kindVoid2, nil, NoneReturnStyle, i64T, boolT),
			stub, args2ib},
		{"instance void arity-5 mixed kinds",
			wrapPtrcallNone5[spotStubClass, *spotStubClass]((*spotStubClass).kindVoid5),
			instanceMD("KindVoid5", (*spotStubClass).kindVoid5, nil, NoneReturnStyle, boolT, i32T, u64T, f64T, i64T),
			stub, args5},

		// Static value/void families (no receiver on either path).
		{"static value arity-0",
			wrapStaticPtrcallValue0(kindStaticValue0),
			staticMD("KindStaticValue0", kindStaticValue0, f64T, ValueReturnStyle),
			nil, args0},
		{"static value arity-2",
			wrapStaticPtrcallValue2(kindStaticValue2),
			staticMD("KindStaticValue2", kindStaticValue2, i64T, ValueReturnStyle, i64T, boolT),
			nil, args2ib},
		{"static void arity-0",
			wrapStaticPtrcallNone0(kindStaticVoid0),
			staticMD("KindStaticVoid0", kindStaticVoid0, nil, NoneReturnStyle),
			nil, args0},
		{"static void arity-2",
			wrapStaticPtrcallNone2(kindStaticVoid2),
			staticMD("KindStaticVoid2", kindStaticVoid2, nil, NoneReturnStyle, f64T, u64T),
			nil, []GDExtensionConstTypePtr{ptrTo(&cellF64), ptrTo(&cellU64)}},

		// Copy-return family: budget only (no engine-free reflective twin).
		{"instance copy-ret arity-1",
			wrapPtrcallCopyRet1[spotStubClass, *spotStubClass, int64, spotRetStub]((*spotStubClass).kindCopyRet1),
			nil, stub, args1},
	}
}

// ptrTo plants a Go value as an engine argument cell. The values it points at
// are package-level, so the cells stay alive for every measured run.
func ptrTo[T any](v *T) GDExtensionConstTypePtr {
	return GDExtensionConstTypePtr(unsafe.Pointer(v))
}

// TestTypedWrapperKindDifferential is the per-kind half of task 2.2: the compiled
// closure and the reflective body must agree on the arguments the handler sees
// and on the bytes written to the return cell.
func TestTypedWrapperKindDifferential(t *testing.T) {
	// The reflective return encode goes through the primitive encoder globals
	// that VariantInitBindings normally primes at engine startup.
	InitPrimativeTypeEncodersForUnitTests()

	for _, k := range wrapperKinds() {
		if k.reflect == nil {
			continue
		}
		t.Run(k.name, func(t *testing.T) {
			var typedRet, reflectRet [16]byte
			retPtr := GDExtensionUninitializedTypePtr(&typedRet[0])

			resetObserved()
			k.typed(k.inst, k.args, retPtr)
			typedSeen, typedCount := snapshotObserved(), observedCount

			resetObserved()
			k.reflect.Ptrcall(k.inst, k.args, GDExtensionUninitializedTypePtr(&reflectRet[0]))
			reflectSeen := snapshotObserved()

			if observedCount != typedCount {
				t.Fatalf("%s: handler saw %d arguments reflectively, %d through the closure",
					k.name, observedCount, typedCount)
			}
			if typedSeen != reflectSeen {
				t.Fatalf("%s: handler saw typed [%s] but reflective [%s]", k.name, typedSeen, reflectSeen)
			}
			if typedRet != reflectRet {
				t.Fatalf("%s: return bytes typed %x != reflective %x", k.name, typedRet, reflectRet)
			}
		})
	}
}

// TestTypedWrapperKindBudgets records the per-kind allocation baseline: the
// compiled closure holds the zero cap for every kind, and each kind's reflective
// figure is logged as the contrast the caps are measured against.
func TestTypedWrapperKindBudgets(t *testing.T) {
	InitPrimativeTypeEncodersForUnitTests()

	for _, k := range wrapperKinds() {
		t.Run(k.name, func(t *testing.T) {
			var ret [16]byte
			binary.LittleEndian.PutUint64(ret[:], 0)
			retPtr := GDExtensionUninitializedTypePtr(&ret[0])
			measureDispatch(t, k.name+" compiled closure", func() {
				k.typed(k.inst, k.args, retPtr)
			})
			if k.reflect == nil {
				t.Logf("%s: reflective twin needs the engine copy constructor", k.name)
				return
			}
			n := testing.AllocsPerRun(50, func() {
				k.reflect.Ptrcall(k.inst, k.args, retPtr)
			})
			t.Logf("%s: reflective body %v allocs/call", k.name, n)
		})
	}
}
