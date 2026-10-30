package core

// typed dispatch machinery for engine->Go calls (typed-method-dispatch).
//
// The reflective paths (method_bind_reflect.go) rebuild the method signature
// per call through reflect.Value machinery. The typed paths instead store a
// compiled closure on the bind metadata, built once at bind time from
// type information the compiler verified. A nil closure means "keep using the
// reflective path"; nothing else is consulted on the hot path.
//
// Closure shapes follow design D1:
//   - ptrcallFn serves engine ptrcalls (GDScript -> Go with static typed
//     slots) and virtual callbacks; args are borrowed engine cells, retPtr is
//     engine-owned memory the closure constructs its return into.
//   - varcallFn serves engine varcalls; args are engine-visible Variants and
//     the returned error is the caller-induced decode-failure seam the varcall
//     callback reports through the engine's call-error slot (mirrors
//     GoMethodMetadata.CallWithError's contract: only decode failures are
//     errors; dispatch faults stay fatal).
//
// Every closure receives only engine-owned cells and writes into
// engine-owned return memory; the pinning rules of native-call-pinning
// (pin Go memory that crosses into the engine, never the other direction)
// therefore have no new obligations on the hot path — the reflective paths
// already hold them, and the typed decode below only reads engine cells.

import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

type ptrcallFn func(inst GDClass, args []GDExtensionConstTypePtr, retPtr GDExtensionUninitializedTypePtr)

// selfTyped is the instance-resolution contract (design D2) the generated
// instance wrappers carry. Engine→Go delivery presents the instance as a
// GDClass interface value whose dynamic type is the bound class's own *T; Go
// generics forbid asserting an interface value to *T (pointer to a type
// parameter) directly, so the wrapper asserts to the constraint-carried S
// instead — a compile-time-provable interface whose Self() returns exactly
// the *T the handler expects. The assertion is a single itab check and Self()
// a one-line method, so this adds no allocation and no reflection anywhere
// on the dispatch path.
//
// Every bound user class carries the same one-liner it already carries for
// GetClassName/GetParentClassName (pure-library route: no codegen step):
//
//	func (c *Example) Self() *Example { return c }
//
// Self() is exported because a bound class lives in the user's package, and a
// Go method with an unexported name does not satisfy an interface declared in
// another package. A class with an incompatible pre-existing Self() fails the
// constraint loudly at its first typed bind, mirroring how shape mismatches
// already surface at registration today.
type selfTyped[T any] interface {
	GDClass
	Self() *T
}

type varcallFn func(inst GDClass, args []Variant) (Variant, error)

// nativeSlot is the constraint family of engine ptrcall argument slots that
// decode as a plain raw load. It is a compile-time aggregate over the slot
// types the engine passes as native cells; the wrapper family is generic over
// it, and the decoder below compiles to the byte-identical load per concrete
// instantiation. It is never used as a runtime interface — no method call, no
// boxing of the loaded value, no allocation per argument.
//
// Membership is pinned to the reflective decoder's own cell loads
// (method_bind_reflect.go), making value parity a code-equality invariant:
//   - scalar slots: bool, ~int, and friends load with identical direct
//     dereferences as the reflective cases (lines 602-707). The small-width
//     signed/unsigned types byte-truncate the int64 cell exactly like the
//     reflective loads do on little-endian; int/uint/int64/uint64/float64
//     load full-width. This is code-equality: identical Go loads of identical
//     cells.
//   - vector/geometry builtins (Vector2..Vector4i, Rect2, Transform2D, Basis,
//     AABB, Plane, Quaternion, Transform3D, Projection) are native-layout
//     `[N]uint8` byte arrays; the reflective decode routes them through engine
//     copy constructors whose observable result for POD values is the copied
//     cell bytes, so the raw load is value-identical while dropping the
//     engine round-trip and the call-scoped pinner records.
//   - refcounted builtins (String, StringName, NodePath, RID, Array,
//     Dictionary, Callable, Signal) load raw with borrow semantics exactly
//     like the reflective cases ("nothing was acquired, released nowhere"):
//     the byte copy duplicates the engine-side handle without touching the
//     engine reference count, so no ownership is transferred and none of
//     AGENTS.md's copy-constructor requirements are bypassed (that rule
//     governs ownership transfer, not borrow reads).
//   - Packed*Array cells load raw, matching the reflective cases (the
//     reflective route goes through a copy constructor whose net reference
//     effect — acquire on decode, release after call — cancels to the same
//     borrowed value the raw load leaves).
//
// Excluded on purpose; binds carrying these parameter types store no typed
// closure and take the reflective fallback:
//   - float32: the reflective decoder double-loads then truncates; a raw
//     `*(*float32)` read would break that parity.
//   - Variant: the reflective decoder routes through VariantViewFromConstPtr,
//     whose view semantics are not a plain raw load.
//   - Go string, []T slices, Object/Ref types: UTF-8 conversion, slice deep
//     copy, and instance-binding resolution are per-call work that a uniform
//     raw load cannot express (the object fallback boundary is documented at
//     resolveObjectArgument, which typed binds don't replace).
type nativeSlot interface {
	bool |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float64 |
		Vector2 | Vector2i | Vector3 | Vector3i | Vector4 | Vector4i |
		Rect2 | Rect2i |
		Transform2D | Transform3D | Basis | AABB |
		Plane | Quaternion | Projection | Color |
		RID | Callable | Signal |
		String | StringName | NodePath |
		Array | Dictionary |
		PackedByteArray | PackedInt32Array | PackedInt64Array |
		PackedFloat32Array | PackedFloat64Array | PackedStringArray |
		PackedVector2Array | PackedVector3Array | PackedVector4Array |
		PackedColorArray
}

// loadTypedSlot[A nativeSlot] loads one engine ptrcall argument cell as the
// declared Go type. Per instantiation the body compiles down to the same
// dereference the reflective decoder performs for that concrete type, with no
// interface conversion and no allocation.
func loadTypedSlot[A nativeSlot](ptr GDExtensionConstTypePtr) A {
	return *(*A)(unsafe.Pointer(ptr))
}

// retValueSlot is the constraint family of ptrcall return values that encode
// as a plain raw store. Membership is pinned to the reflective return
// encoder's audited matrix (variant_refect_value.go, backed by the
// synchronous matrix audit): scalars, the vector/geometry builtins, RID, and
// String return through their XEncoder.EncodeTypePtrArg, which is exactly
// `*(*R)(pOut) = value` — the store below — with no engine calls and no
// reference effects.
//
// Deliberately absent (they must route through the engine copy constructor,
// transferring ownership via an engine reference — never a raw byte store;
// see variant_copy_ownership, borrow-echo-detection, and the AGENTS.md
// copy-constructor rule): StringName, NodePath, Callable, Signal, Array,
// Dictionary, every Packed*Array, and Variant. Those travel through the
// copy-return family (per-type generated RetEncodeInto methods); binds
// carrying them are served there or by the reflective fallback.
type retValueSlot interface {
	bool |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float64 |
		Vector2 | Vector2i | Vector3 | Vector3i | Vector4 | Vector4i |
		Rect2 | Rect2i |
		Transform2D | Transform3D | Basis | AABB |
		Plane | Quaternion | Projection | Color |
		RID | String
}

// storeTypedRet[R retValueSlot] stores the handler's return into the
// engine-owned return cell as a raw native value. Per instantiation this is
// the same store the reflective encoder performs for value types, with no
// interface conversion and no allocation.
func storeTypedRet[R retValueSlot](retPtr GDExtensionTypePtr, value R) {
	*(*R)(unsafe.Pointer(retPtr)) = value
}

// Varcall-side constraints (varArgSlot/varRetSlot) and the bind-time slot
// factories (newVarArgDecoder, newVarArgReleaser, newVarRetEncoder) are
// generated into var_args.gen.go from cmd/generate/typedwrapper.

// notificationFn is the fixed-shape closure for the notification virtual
// (design D6 / notification-virtual-dispatch): shape taken from
// GoCallback_ClassCreationInfoNotification — int32 what plus the
// GDExtensionBool reversed flag — return void.
type notificationFn func(inst GDClass, what int32, reversed bool)

// propGetFn is the fixed-shape closure for the `_get` property virtual
// (property-access-dispatch), shape from GoCallback_ClassCreationInfoGet:
// the property name arrives as UTF-8 text and the handler reports
// (value, exists) with the same false-means-unhandled contract.
type propGetFn func(inst GDClass, name string) (Variant, bool)

// propSetFn is the fixed-shape closure for the `_set` property virtual
// (property-access-dispatch), shape from GoCallback_ClassCreationInfoSet:
// UTF-8 name plus the incoming Variant; bool reports whether the property
// was consumed.
type propSetFn func(inst GDClass, name string, value Variant) bool

// propCanRevertFn is the fixed-shape closure for `_property_can_revert`
// (property-revert-virtuals), shape from its dispatch site: the property
// name arrives as a StringName value.
type propCanRevertFn func(inst GDClass, name StringName) bool

// propGetRevertFn is the fixed-shape closure for `_property_get_revert`
// (property-revert-virtuals), shape from its dispatch site: StringName name
// in, (revert value, has-revert) out with the reflective dispatch's bool
// contract.
type propGetRevertFn func(inst GDClass, name StringName) (Variant, bool)

// toStringFn is the fixed-shape closure for the `to_string` virtual
// (to-string-virtual-dispatch), shape from resolveToStringFor: receiver
// only, a plain Go string out — the shape the reflective dispatch guards on
// at runtime the bind makes a compile-time guarantee.
type toStringFn func(inst GDClass) string
