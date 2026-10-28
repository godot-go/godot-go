package core

import (
	"encoding/binary"
	"math"
	"reflect"
	"testing"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// typed-method-dispatch task 2.1 spot-check (engine-free unit tier).
//
// The generated wrapper family (wrappers.gen.go) resolves the receiver
// through the selfTyped contract and decodes/encodes through the constraint
// unions. These tests drive representative wrappers as pure closures in Go —
// engine cells are synthesized byte arrays and the receiver is a stub
// GDClass — asserting:
//   - the instance handoff lands on the handler with the exact concrete *T
//     (the selfTyped itab assert plus the Self() one-liner),
//   - each argument slot is a raw load of the planted cell bytes,
//   - value-family returns raw-store their bytes into the engine ret cell,
//   - void-return wrappers leave the ret cell untouched,
//   - static binds receive no receiver and arity 0 works at both edges.
//
// The receivers here satisfy the same contract bound user classes will
// (one Self() one-liner next to the GetClassName/GetParentClassName pair
// they already write); a misbound closure fails loudly on the itab check.
// Engine-tier parity (typed-vs-reflective through a real dispatch) is the
// task 2.2 fixture tier, run under GODOT.

type spotStubClass struct {
	owner *GodotObject
}

// Self is the selfTyped one-liner the wrapper contract requires.
func (s *spotStubClass) Self() *spotStubClass { return s }

func (s *spotStubClass) Destroy()                                     {}
func (s *spotStubClass) GetGodotObjectOwner() *GodotObject            { return s.owner }
func (s *spotStubClass) SetGodotObjectOwner(owner *GodotObject)       { s.owner = owner }
func (s *spotStubClass) GetClassName() string                         { return "SpotStub" }
func (s *spotStubClass) GetParentClassName() string                   { return "Object" }
func (s *spotStubClass) AsGDExtensionObjectPtr() GDExtensionObjectPtr { return nil }
func (s *spotStubClass) AsGDExtensionConstObjectPtr() GDExtensionConstObjectPtr {
	return nil
}
func (s *spotStubClass) AsGDExtensionTypePtr() GDExtensionTypePtr { return nil }
func (s *spotStubClass) AsGDExtensionConstTypePtr() GDExtensionConstTypePtr {
	return nil
}

// Triple is a pure int64→int64 method bound both ways (typed closure and
// reflective method expression) in the task 4.2 byte-parity test.
func (s *spotStubClass) Triple(n int64) int64 { return n * 3 }

func TestTypedWrapperInstanceValueRet(t *testing.T) {
	stub := &spotStubClass{}
	var arg1 [8]byte
	binary.LittleEndian.PutUint64(arg1[:], 42)
	var arg2 [8]byte
	binary.LittleEndian.PutUint64(arg2[:], math.Float64bits(2.5))
	args := []GDExtensionConstTypePtr{
		GDExtensionConstTypePtr(&arg1[0]),
		GDExtensionConstTypePtr(&arg2[0]),
	}
	var ret [8]byte

	// S is not inferable from the handler alone; bind-time instantiations
	// pass the class pointer and its self-typed twin explicitly.
	fn := wrapPtrcallValue2[spotStubClass, *spotStubClass](func(self *spotStubClass, n int64, f float64) Vector2 {
		var v Vector2
		binary.LittleEndian.PutUint32(v[0:4], math.Float32bits(float32(n)))
		binary.LittleEndian.PutUint32(v[4:8], math.Float32bits(float32(f)))
		return v
	})
	fn(stub, args, GDExtensionUninitializedTypePtr(&ret[0]))

	want := [8]byte{}
	binary.LittleEndian.PutUint32(want[0:4], math.Float32bits(42))
	binary.LittleEndian.PutUint32(want[4:8], math.Float32bits(2.5))
	if string(ret[:]) != string(want[:]) {
		t.Fatalf("ret cell %v, want %v (Vector2 of 42, 2.5)", ret, want)
	}
}

func TestTypedWrapperInstanceVoidRet(t *testing.T) {
	stub := &spotStubClass{}
	var arg1 [8]byte
	binary.LittleEndian.PutUint64(arg1[:], 7)
	var arg2 [8]byte
	binary.LittleEndian.PutUint64(arg2[:], math.Float64bits(0.75))
	args := []GDExtensionConstTypePtr{
		GDExtensionConstTypePtr(&arg1[0]),
		GDExtensionConstTypePtr(&arg2[0]),
	}
	ret := [8]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}

	gotSelf := (*spotStubClass)(nil)
	var gotN int64
	var gotF float64
	fn := wrapPtrcallNone2[spotStubClass, *spotStubClass](func(self *spotStubClass, n int64, f float64) {
		gotSelf, gotN, gotF = self, n, f
	})
	fn(stub, args, GDExtensionUninitializedTypePtr(&ret[0]))

	if gotSelf != stub {
		t.Fatalf("handler received different receiver: got %p, want %p", gotSelf, stub)
	}
	if gotN != 7 {
		t.Fatalf("arg1 decoded %d, want 7", gotN)
	}
	if gotF != 0.75 {
		t.Fatalf("arg2 decoded %v, want 0.75", gotF)
	}
	for i, b := range ret {
		if b != 0xFF {
			t.Fatalf("void wrapper wrote ret cell at byte %d (%v)", i, ret)
		}
	}
}

func TestTypedWrapperStaticValueRet(t *testing.T) {
	var arg1 [8]byte
	binary.LittleEndian.PutUint64(arg1[:], 11)
	var arg2 [8]byte
	binary.LittleEndian.PutUint64(arg2[:], math.Float64bits(3.25))
	args := []GDExtensionConstTypePtr{
		GDExtensionConstTypePtr(&arg1[0]),
		GDExtensionConstTypePtr(&arg2[0]),
	}
	var ret [8]byte

	fn := wrapStaticPtrcallValue2[int64, float64, int64](func(n int64, f float64) int64 {
		if f == 3.25 {
			return n
		}
		return n * 2
	})
	// Statics receive no instance; nil must be tolerated.
	fn(nil, args, GDExtensionUninitializedTypePtr(&ret[0]))

	if binary.LittleEndian.Uint64(ret[:]) != 11 {
		t.Fatalf("ret %v, want 11 (static handler echoed its first arg)", ret)
	}
}

func TestTypedWrapperStaticVoidRetArity0(t *testing.T) {
	called := false
	fn := wrapStaticPtrcallNone0(func() { called = true })
	fn(nil, nil, GDExtensionUninitializedTypePtr(nil))
	if !called {
		t.Fatal("arity-0 static void wrapper did not invoke its handler")
	}
}

func TestTypedWrapperInstanceValueRetArity0(t *testing.T) {
	stub := &spotStubClass{}
	var ret [8]byte

	fn := wrapPtrcallValue0[spotStubClass, *spotStubClass](func(self *spotStubClass) uint64 {
		if self != stub {
			t.Fatalf("arity-0 wrapper resolved receiver %p, want %p", self, stub)
		}
		return 9
	})
	fn(stub, nil, GDExtensionUninitializedTypePtr(&ret[0]))

	if binary.LittleEndian.Uint64(ret[:]) != 9 {
		t.Fatalf("ret %v, want 9", ret)
	}
}

// TestTypedWrapperDispatchAllocsPerRun records the task 2.2 allocation
// budgets: the compiled-closure path itself must not allocate per call.
// The args slice and ret cell are engine-medium provided by the caller, so
// zero allocations is the whole budget the wiring tasks must match (the
// reflective baseline allocates reflect frames, value slices and boxed
// results per call).
func TestTypedWrapperDispatchAllocsPerRun(t *testing.T) {
	stub := &spotStubClass{}
	var arg [8]byte
	binary.LittleEndian.PutUint64(arg[:], 1)
	cell := GDExtensionConstTypePtr(&arg[0])
	args1 := []GDExtensionConstTypePtr{cell}
	args2 := []GDExtensionConstTypePtr{cell, cell}
	retPtr := GDExtensionUninitializedTypePtr(&arg[0])

	if n := testing.AllocsPerRun(200, func() {
		wrapPtrcallValue2[spotStubClass, *spotStubClass](func(self *spotStubClass, n int64, f float64) int64 {
			return n
		})(stub, args2, retPtr)
	}); n > 0 {
		t.Fatalf("instance value-ret arity-2 wrapper allocates %v per call, want 0", n)
	}

	if n := testing.AllocsPerRun(200, func() {
		wrapPtrcallNone1[spotStubClass, *spotStubClass](func(self *spotStubClass, n int64) {})(stub, args1, nil)
	}); n > 0 {
		t.Fatalf("instance void-ret arity-1 wrapper allocates %v per call, want 0", n)
	}

	if n := testing.AllocsPerRun(200, func() {
		wrapStaticPtrcallValue2[int64, float64, int64](func(n int64, f float64) int64 { return n })(nil, args2, retPtr)
	}); n > 0 {
		t.Fatalf("static value-ret arity-2 wrapper allocates %v per call, want 0", n)
	}

	if n := testing.AllocsPerRun(200, func() {
		wrapStaticPtrcallNone0(func() {})(nil, nil, nil)
	}); n > 0 {
		t.Fatalf("static void arity-0 wrapper allocates %v per call, want 0", n)
	}
}

// spotRetStub is a stub copy-encodable return standing in for the generated
// copy-ret methods: it hands its carry bytes to the ret cell exactly like a
// one-line RetEncodeInto does around its engine constructor (which these
// engine-free tests cannot call — the global bindings tables are nil), and
// counts RetDestroy calls so the destroy-source gate is observable.
type spotRetStub [8]byte

// spotRetDestroyCount records stub source releases; the copy-ret tests reset
// it before driving a wrapper.
var spotRetDestroyCount int

func (r spotRetStub) RetEncodeInto(retPtr GDExtensionUninitializedTypePtr) {
	copy((*[8]byte)(unsafe.Pointer(retPtr))[:], r[:])
}

func (r spotRetStub) RetDestroy() {
	spotRetDestroyCount++
}

// Compile-time witnesses: every emitted copy-ret builtin satisfies the
// family's constraint through its generated one-line method (task 1.4's
// emission is complete and shape-correct without touching the engine).
var (
	_ RetCopyEncodable = StringName{}
	_ RetCopyEncodable = NodePath{}
	_ RetCopyEncodable = Callable{}
	_ RetCopyEncodable = Signal{}
	_ RetCopyEncodable = Array{}
	_ RetCopyEncodable = Dictionary{}
	_ RetCopyEncodable = PackedByteArray{}
	_ RetCopyEncodable = PackedInt32Array{}
	_ RetCopyEncodable = PackedInt64Array{}
	_ RetCopyEncodable = PackedFloat32Array{}
	_ RetCopyEncodable = PackedFloat64Array{}
	_ RetCopyEncodable = PackedStringArray{}
	_ RetCopyEncodable = PackedVector2Array{}
	_ RetCopyEncodable = PackedVector3Array{}
	_ RetCopyEncodable = PackedVector4Array{}
	_ RetCopyEncodable = PackedColorArray{}
)

func TestTypedWrapperInstanceCopyRet(t *testing.T) {
	stub := &spotStubClass{}
	var arg [8]byte
	binary.LittleEndian.PutUint64(arg[:], 5)
	args := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(&arg[0])}
	var ret [8]byte

	fn := wrapPtrcallCopyRet1[spotStubClass, *spotStubClass, int64, spotRetStub](func(self *spotStubClass, n int64) spotRetStub {
		if self != stub {
			t.Fatalf("handler received different receiver: got %p, want %p", self, stub)
		}
		var out spotRetStub
		binary.LittleEndian.PutUint64(out[:], uint64(n)*100)
		return out
	})
	spotRetDestroyCount = 0
	fn(stub, args, GDExtensionUninitializedTypePtr(&ret[0]))

	if binary.LittleEndian.Uint64(ret[:]) != 500 {
		t.Fatalf("ret %v, want 500 (the stub copy-encoder's bytes)", ret)
	}
	// The int64 slot is not the return's type, so no echo is possible and the
	// source release fires exactly like the reflective matrix's
	// `if destroySource { inst.Destroy() }` gate for a classified-owned value.
	if spotRetDestroyCount != 1 {
		t.Fatalf("source destroyed %d times, want 1", spotRetDestroyCount)
	}
}

func TestTypedWrapperStaticCopyRetArity0(t *testing.T) {
	var ret [8]byte

	fn := wrapStaticPtrcallCopyRet0[spotRetStub](func() spotRetStub {
		var out spotRetStub
		binary.LittleEndian.PutUint64(out[:], 0xABCD)
		return out
	})
	spotRetDestroyCount = 0
	fn(nil, nil, GDExtensionUninitializedTypePtr(&ret[0]))

	if binary.LittleEndian.Uint64(ret[:]) != 0xABCD {
		t.Fatalf("ret %v, want 0xABCD", ret)
	}
	// No slots means no echo candidate: the freshly built return is owned and
	// destroyed after encoding.
	if spotRetDestroyCount != 1 {
		t.Fatalf("source destroyed %d times, want 1", spotRetDestroyCount)
	}
}

// TestPtrcallSelectsTypedDispatch pins the task 4.1 fallback seam in
// GoMethodMetadata.Ptrcall (the single entry both the ptrcall callback and the
// virtual callback reach). A metadata carrying a compiled typedDispatch closure
// dispatches through it with zero reflection; the reflective Func is left as
// the zero reflect.Value, so a closure that returns the right value also proves
// the reflective body (which would panic on that zero Value) was never entered.
func TestPtrcallSelectsTypedDispatch(t *testing.T) {
	stub := &spotStubClass{}
	var arg [8]byte
	binary.LittleEndian.PutUint64(arg[:], 7)
	args := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(&arg[0])}
	var ret [8]byte

	md := &GoMethodMetadata{
		typedDispatch: wrapPtrcallValue1[spotStubClass, *spotStubClass](
			func(self *spotStubClass, n int64) int64 { return n * 3 }),
	}
	md.Ptrcall(stub, args, GDExtensionUninitializedTypePtr(&ret[0]))

	if got := binary.LittleEndian.Uint64(ret[:]); got != 21 {
		t.Fatalf("typed dispatch produced %d, want 21 (7*3 through the closure)", got)
	}
}

// TestPtrcallFallsBackWhenDispatchNil pins the other half of the seam: with no
// typedDispatch closure (the name-string path, or an untypeable signature) the
// call runs the reflective body. The unconfigured metadata panics inside that
// body, which is exactly the path the typed closure would have bypassed — so a
// panic proves the fallback fired rather than a nil closure being invoked.
func TestPtrcallFallsBackWhenDispatchNil(t *testing.T) {
	stub := &spotStubClass{}
	var arg [8]byte
	binary.LittleEndian.PutUint64(arg[:], 7)
	args := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(&arg[0])}
	var ret [8]byte

	md := &GoMethodMetadata{} // typedDispatch nil
	_, fired := recoverFatal(func() {
		md.Ptrcall(stub, args, GDExtensionUninitializedTypePtr(&ret[0]))
	})
	if !fired {
		t.Fatal("nil typedDispatch must fall through to the reflective path")
	}
}

// TestPtrcallTypedReflectiveByteParity is task 4.2: the same Go method bound
// through the typed closure and through the reflective (name-string) body must
// write byte-identical bytes to the engine return cell. int64 is the cleanest
// witness — both paths raw-store a native cell with no engine call and no
// reference effect — so the comparison is a pure Go byte diff, engine-free.
func TestPtrcallTypedReflectiveByteParity(t *testing.T) {
	// The reflective body encodes returns through the primitive encoder
	// globals, which VariantInitBindings normally primes at engine startup.
	// The int64 reflect encode closure is a raw native-cell store (no engine
	// pointer dereferenced), so priming through the unit-test seam keeps this
	// parity check engine-free.
	InitPrimativeTypeEncodersForUnitTests()

	stub := &spotStubClass{}
	var arg [8]byte
	binary.LittleEndian.PutUint64(arg[:], 7)
	gdArgs := []GDExtensionConstTypePtr{GDExtensionConstTypePtr(&arg[0])}

	var typedRet, reflectRet [8]byte

	typedMD := &GoMethodMetadata{
		typedDispatch: wrapPtrcallValue1[spotStubClass, *spotStubClass]((*spotStubClass).Triple),
	}
	typedMD.Ptrcall(stub, gdArgs, GDExtensionUninitializedTypePtr(&typedRet[0]))

	reflectMD := &GoMethodMetadata{
		ClassName:        "SpotStub",
		GdMethodName:     "triple",
		GoMethodName:     "Triple",
		Func:             reflect.ValueOf((*spotStubClass).Triple),
		GoReturnType:     reflect.TypeFor[int64](),
		GoReturnStyle:    ValueReturnStyle,
		GoArgumentTypes:  []reflect.Type{reflect.TypeFor[int64]()},
		gdeArgumentTypes: []GDExtensionVariantType{GDEXTENSION_VARIANT_TYPE_INT},
	}
	reflectMD.Ptrcall(stub, gdArgs, GDExtensionUninitializedTypePtr(&reflectRet[0]))

	if typedRet != reflectRet {
		t.Fatalf("typed dispatch %x != reflective dispatch %x", typedRet, reflectRet)
	}
	if binary.LittleEndian.Uint64(typedRet[:]) != 21 {
		t.Fatalf("both paths produced %d, want 21", binary.LittleEndian.Uint64(typedRet[:]))
	}
}
