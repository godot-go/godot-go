package core

import (
	"reflect"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// typed-method-dispatch task 5.1 unit tier (engine-free).
//
// The typed varcall closures are built from the generated factories in
// var_args.gen.go (bind-time reflection only). What is observable without a
// live engine:
//   - the trailing-default fill is literally shared (fillCallArgs delegates to
//     fillVarcallCallArgs, and the wrappers call the shared function),
//   - the under-supply panic carries the identical payload on both paths,
//   - CallWithError selects the typed closure exactly like Ptrcall selects the
//     ptrcall dispatch,
//   - the bind-time factories construct for every representative union member.
//
// Decode/encode execution crosses into the engine (To*/XEncoder calls), so
// typed-vs-reflective value parity for decode, return encoding, and the
// variadic raw-spread convention lives in the GODOT tier.

// The union memberships are proven by instantiating these constraint-carried
// factories: a non-member type argument fails to compile, so the instantiation
// lines are the witnesses (a constraint interface cannot itself appear as a
// variable's type).
func witnessVarRet[R varRetSlot]() {}
func witnessVarArg[A varArgSlot]() {}

var (
	// Every value-return-family return type satisfies the varcall return
	// constraint: the retValueSlot family plus float32 and Go string.
	_ = witnessVarRet[bool]
	_ = witnessVarRet[int]
	_ = witnessVarRet[int64]
	_ = witnessVarRet[uint64]
	_ = witnessVarRet[float32]
	_ = witnessVarRet[float64]
	_ = witnessVarRet[string]
	_ = witnessVarRet[Vector2]
	_ = witnessVarRet[Color]
	_ = witnessVarRet[RID]
	_ = witnessVarRet[String]

	// Representative varcall argument slots: scalar, Go string, a
	// convert-construct borrow, and owned containers.
	_ = witnessVarArg[bool]
	_ = witnessVarArg[int]
	_ = witnessVarArg[float32]
	_ = witnessVarArg[string]
	_ = witnessVarArg[StringName]
	_ = witnessVarArg[Array]
	_ = witnessVarArg[PackedColorArray]
)

// The full nativeSlot ⊆ varArgSlot proof is the package itself: the shared
// bind overloads in typed_binds.gen.go constrain A nativeSlot and then
// instantiate the varcall wrappers, which require varArgSlot. Any member
// missing from the varcall union would fail that compilation.

// TestFillVarcallCallArgsMatchesFillCallArgs pins the delegation: the typed
// wrappers' statically-known arity must produce exactly the same filled slot
// array as the reflective metadata-driven fill, and the variadic metadata must
// keep the nil short-circuit the typed variadic closures hard-code by omission.
func TestFillVarcallCallArgsMatchesFillCallArgs(t *testing.T) {
	// Variant values here are opaque 24-byte payloads; fill moves them by
	// value without touching the engine.
	marker := func(b byte) Variant {
		var v Variant
		v[0] = b
		return v
	}
	supplied := []Variant{marker(1), marker(2), marker(3)}
	defaults := []Variant{marker(20), marker(30)}

	md := &GoMethodMetadata{
		GdMethodName:     "filled",
		DefaultArguments: defaults,
		// fillCallArgs counts defaults through the ptr-parallel slice the bind
		// core maintains alongside DefaultArguments; both exist together in
		// real metadata, and only the length is consulted here.
		gdeDefaultArgumentPtrs: []GDExtensionVariantPtr{nil, nil},
		gdeArgumentTypes: []GDExtensionVariantType{
			GDEXTENSION_VARIANT_TYPE_NIL,
			GDEXTENSION_VARIANT_TYPE_NIL,
			GDEXTENSION_VARIANT_TYPE_NIL,
			GDEXTENSION_VARIANT_TYPE_NIL,
			GDEXTENSION_VARIANT_TYPE_NIL,
		},
	}

	cases := []struct {
		name     string
		gdArgs   []Variant
		declared int
	}{
		{"all supplied", supplied, 3},
		{"trailing defaults fill", supplied[:1], 3},
		{"exact boundary", supplied[:3], 5},
		{"all defaults", nil, 2},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			md.gdeArgumentTypes = md.gdeArgumentTypes[:tc.declared]
			defStart := tc.declared - len(defaults)
			if defStart < 0 {
				defStart = 0
			}
			got := fillVarcallCallArgs(tc.gdArgs, tc.declared, defStart, defaults, md.GdMethodName)
			want := md.fillCallArgs(tc.gdArgs)
			if len(got) != len(want) {
				t.Fatalf("fill produced %d slots, want %d", len(got), len(want))
			}
			for i := range want {
				if got[i] != want[i] {
					t.Fatalf("slot %d: typed fill %v != reflective fill %v", i, got[i], want[i])
				}
			}
		})
	}

	variadicMD := &GoMethodMetadata{IsVariadic: true}
	if got := variadicMD.fillCallArgs(supplied); got != nil {
		t.Fatalf("variadic reflective fill = %v, want nil (the typed variadic closures skip fill)", got)
	}
}

// TestTypedVarcallUnderSupplyPanicsLikeReflective pins the loud under-supply
// convention: the shared fill raises the identical payload on both paths, and
// the panic fires before any decode (so it is engine-free on the typed side).
func TestTypedVarcallUnderSupplyPanicsLikeReflective(t *testing.T) {
	stub := &spotStubClass{}

	typedMsg, typedFired := recoverFatal(func() {
		fn := wrapVarcallNone1[spotStubClass, *spotStubClass](
			func(self *spotStubClass, n int64) {}, nil, "under")
		fn(stub, nil)
	})
	if !typedFired {
		t.Fatal("typed varcall closure must panic on an unfilled slot")
	}

	reflectMsg, reflectFired := recoverFatal(func() {
		md := &GoMethodMetadata{
			GdMethodName:     "under",
			GoArgumentTypes:  []reflect.Type{reflect.TypeFor[int64]()},
			gdeArgumentTypes: []GDExtensionVariantType{GDEXTENSION_VARIANT_TYPE_INT},
		}
		md.CallWithError(stub)
	})
	if !reflectFired {
		t.Fatal("reflective CallWithError must panic on an unfilled slot")
	}
	if typedMsg != reflectMsg {
		t.Fatalf("typed panic payload %q != reflective panic payload %q", typedMsg, reflectMsg)
	}
}

// sentinel is the marker a handcrafted typed closure returns to prove the
// typed branch ran; the reflective body cannot produce it on these fixtures.
func TestCallWithErrorSelectsTypedVarcall(t *testing.T) {
	stub := &spotStubClass{}
	var sentinel Variant
	sentinel[0] = 0xA5

	called := false
	md := &GoMethodMetadata{
		typedVarcall: func(inst GDClass, args []Variant) (Variant, error) {
			called = true
			if inst != stub {
				t.Fatalf("typed varcall received receiver %p, want %p", inst, stub)
			}
			return sentinel, nil
		},
	}
	got, err := md.CallWithError(stub)
	if err != nil {
		t.Fatalf("typed varcall returned error: %v", err)
	}
	if !called {
		t.Fatal("CallWithError did not invoke the typed varcall closure")
	}
	if got != sentinel {
		t.Fatal("CallWithError did not return the typed closure's Variant")
	}
}

// TestCallWithErrorFallsBackWhenVarcallNil pins the other half of the seam:
// with no typedVarcall closure the call runs the reflective body — an
// unconfigured metadata panics inside it, proving the fallback fired rather
// than a nil closure being invoked.
func TestCallWithErrorFallsBackWhenVarcallNil(t *testing.T) {
	md := &GoMethodMetadata{} // typedVarcall nil
	_, fired := recoverFatal(func() {
		md.CallWithError(&spotStubClass{})
	})
	if !fired {
		t.Fatal("nil typedVarcall must fall through to the reflective path")
	}
}

// The variadic closures' raw-spread convention (the collected engine Variants
// reach the handler unmodified, with no fill step) is exercised in the GODOT
// tier: every typed closure ends in NewVariantNil()/the Variant return encode,
// both of which dereference the engine table, so no engine-free run of the
// generated variadic wrapper is possible. Task 7.x's demo conversion binds
// typed variadic methods there and the existing varcall suites compare the
// observed arguments against the reflective baseline.

// TestVarcallFactoriesBuildEngineFree pins that the bind-time factory trio
// constructs for representative members of both unions without touching the
// engine (the closures only dereference engine pointers when invoked).
func TestVarcallFactoriesBuildEngineFree(t *testing.T) {
	if newVarArgDecoder[int64]() == nil {
		t.Fatal("scalar-kind decoder factory returned nil")
	}
	if newVarArgDecoder[StringName]() == nil {
		t.Fatal("array-kind decoder factory returned nil")
	}
	if newVarRetEncoder[int64]() == nil {
		t.Fatal("scalar-kind return encoder factory returned nil")
	}
	if newVarRetEncoder[String]() == nil {
		t.Fatal("array-kind return encoder factory returned nil")
	}

	// The releaser folds to a no-op for non-owned slots and resolves a real
	// destroy closure for the owned containers (invoking either needs the
	// engine, construction does not).
	relInt := newVarArgReleaser[int64]()
	relInt(0)
	if newVarArgReleaser[Array]() == nil {
		t.Fatal("owned-container releaser factory returned nil")
	}
}
