## Context

The container decode paths were added by `implement-container-built-in-types`
using owned copy-constructors for arguments, which leak one refcount per
container argument per call (see proposal.md — Why). The project already solved
the identical problem for `StringName`/`NodePath` (see
`openspec/specs/stringname-lifecycle-management/spec.md` and archived change
`2026-08-02-fix-orphan-stringname-warnings`): borrow semantics for argument
decoding plus copy-constructor ownership transfer for ptrcall returns. This
change applies that same, already-tested pattern to the container types.

The copy-constructor helper pattern (`StringNameCopyConstructor`,
`NodePathCopyConstructor`) lives in `pkg/builtin/char_string.go` and calls
Godot's `constructor_1` via `CallBuiltinConstructor`. Each container type's
method bindings expose an equivalent `constructor_1` (e.g.
`globalArrayMethodBindings.constructor_1`).

## Goals / Non-Goals

**Goals:**
- No leaked refcounts for container arguments in either call style.
- Correct ptrcall echo semantics (returning a received container argument).
- No double-free / use-after-free hazards.
- Lifecycle semantics consistent with `StringName`/`NodePath`.

**Non-Goals:**
- No changes to the generic encoder subsystem (`variant_builtinclass_encoder.go`).
- No GC/finalizer-based release mechanism.
- `String` and Go-created varcall container returns (pre-existing, separate concerns).

## Decisions

### Decision 1: Argument decoding uses borrow semantics

Arguments received from Godot live in Godot-managed memory. The Go side must
not hold a refcount that outlives the call.

- **Ptrcall decode** (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`):
  replace `v := NewXWithX(*pV)` with a direct byte-copy `v := *(*X)(arg)`,
  matching the existing `String`/`StringName`/`NodePath` cases. No constructor
  call, no refcount change.
- **Varcall decode** (`convertVariantToGoTypeReflectValue`): decode as an
  owned copy (`v := arg.ToX(); return reflect.ValueOf(v)`) and release it in
  `GoMethodMetadata.Call` via `destroyOwnedContainerArgs` after the return value
  is encoded. The `StringName`/`NodePath` byte-copy borrow (`arg.ToX()` →
  byte-copy → `Destroy()`) is NOT usable here: when the supplied Variant is a
  *different* container type than the Go parameter (e.g. `Array[int]` →
  `PackedInt64Array`), the type-from-variant constructor allocates a fresh
  buffer that the Variant does not own, so destroying the constructed copy at
  decode time leaves the caller reading freed memory. Owned-copy + post-call
  release is correct in both the same-type and cross-type cases and is
  leak-free (one acquire at decode, one release after the call).

**Alternative considered:** the byte-copy borrow for varcall (matching
`StringName`/`NodePath`). It was implemented first, but it regressed
cross-type conversions — `test_tarray_arg` (an `Array[int]` supplied to a
`PackedInt64Array` parameter) read garbage because the converted buffer was
freed at decode time. Owned decode + post-call release supersedes it. The
borrow is retained for ptrcall, where Godot hands over a pointer it keeps valid
for the call and no conversion occurs (validated calls require an exact type
match, so a mismatched arg dispatches as a varcall instead).

### Decision 2: Ptrcall return encoding uses CopyConstructor + Destroy

In `GDExtensionTypePtrFromReflectValue`, replace
`XEncoder.EncodeTypePtrArg(inst, rOut)` (shallow byte-copy) with
`XCopyConstructor(rOut, inst.NativeConstPtr())` followed by `inst.Destroy()`,
for `Array` and the nine `Packed*Array` types — exactly mirroring the existing
`StringName`/`NodePath` cases. The copy constructor increments the refcount so
the output buffer holds an owned reference; destroying the Go-side value
releases the original, leaving exactly one reference in the output buffer. This
is what makes echoing a borrowed argument safe: the output buffer gets a fresh
owned copy rather than aliasing the caller's data.

New helpers `ArrayCopyConstructor`, `PackedByteArrayCopyConstructor`, … are
added beside `StringNameCopyConstructor`, each calling the type's
`constructor_1`.

**Alternative considered:** shallow byte-copy without destroy (current
behavior). It only balances for echo-style methods and leaks otherwise; it also
leaves the output buffer aliasing the caller's memory. Rejected.

### Decision 3: Varcall return path unchanged

The varcall return path wraps the value in a Variant via the
variant-from-type constructor, which already acquires its own reference, so
echoing a borrowed container argument through varcall is already correct. No
change is required there for this leak.

## Risks / Trade-offs

- **Borrow lifetime** → a Go method that stores a container argument past the
  call will hold a dangling reference. Mitigation: this matches the documented
  StringName/NodePath contract; document it and cover with an echo/consume test
  suite. Callers needing persistence must copy explicitly.
- **Echo refcount accounting** → the `CopyConstructor + Destroy` return on a
  borrowed value is subtle. Mitigation: it is the exact, tested StringName
  pattern; add ptrcall+varcall echo round-trip tests for every container type
  and run `make test` to confirm no orphan/double-free symptoms.
- **Helper sprawl** → ten new copy-constructor helpers. Mitigation: they are
  one-line wrappers over `CallBuiltinConstructor`, identical in shape to the
  existing two.

## Review Findings (PR #140)

Captured from the post-implementation code review so they survive archiving.
The core fix was verified sound: refcount accounting traced through both call
styles, confirmed against Godot's CoW internals (`Array::_ref` shares `_p`),
`make test` green locally (637 pass) and CI green on linux/windows.

- **Content-based borrow-echo detection (medium)** → `isPtrcallBorrowEcho`
  (`pkg/core/method_bind.go:428`) classifies echoes by byte equality
  (`reflect.DeepEqual`), but Godot copy constructors produce byte-identical
  values differing only in refcount. A Go method returning an *unmutated
  defensive clone* of a received argument is therefore misclassified as a
  borrow echo and the clone's refcount is never released: a silent,
  bounded leak of one reference per such call, no crash. Inherent to
  content-based heuristics for CoW types; recorded as the
  "Byte-identical return classified as a borrow echo" scenario in the
  capability spec. A precise fix would need ownership tracking rather than
  content comparison.
- **Mutation propagation through ptrcall borrows (low)** → because the
  byte-copy borrow does not increment the refcount, first-write copy-on-write
  does not trigger, so a Go method mutating a received container mutates the
  GDScript caller's storage (under the previous owned-copy decode the refcount
  bump isolated the caller). This behavior change was not in the original risk
  list; it is now explicit contract semantics in the capability spec
  ("Borrowed Ptrcall Container Arguments Share Storage With The Caller")
  rather than accidental. Empirical refinement from implementing the tests:
  `Array`/`Dictionary` are shared-reference containers with no copy-on-write,
  so mutations propagate through *varcall's owned decode* as well — only the
  `Packed*Array` types are isolated by varcall ownership. A copy-constructor
  clone of an `Array` argument aliases the original, so "clone then mutate"
  also mutates the caller's argument and keeps the return byte-equal to it
  (echo-classified); tests that need a non-echo return must build a fresh
  container instead.
- **Dictionary arg-level test coverage gap (low)** → the varcall owned-copy
  decode plus `destroyOwnedContainerArgs` release machinery covers
  `Dictionary`, but the demo test project has no Dictionary argument
  round-trip assertions exercising them. Follow-up test work.
- **Same bug class remains for basic refcounted types (observation)** →
  `StringName`/`NodePath` ptrcall echoes go through CopyConstructor+Destroy
  without borrow-echo detection, so the destroy-the-borrow bug class this
  change fixes for containers may still exist there. Candidate sibling
  change; `String` and `PackedVector4Array` lifecycle remain deferred per
  proposal Non-goals.

## Migration Plan

Single commit behind no feature flag; the container decode paths are new and
unreleased, so there is no existing behavior to migrate. Rollback is a revert.

## Open Questions

- Whether `Dictionary` should also move its ptrcall return encoding to
  copy-constructor for consistency (it has no ptrcall arg decode, so it is not
  required to fix this leak). Deferred; can be folded in if trivial.
