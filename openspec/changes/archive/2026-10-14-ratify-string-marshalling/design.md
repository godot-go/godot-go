# Design

## Context

See proposal.md for motivation. The current boundary state, as verified in
exploration:

- Bound Go methods already accept all four string-shaped signature types.
  Native `string` decodes via `Variant.ToGoString()` (varcall) and
  `*(*String)(arg).ToUtf8()` (ptrcall); `StringName`/`NodePath` decode as
  call-scoped borrows; `String` decodes as an owned copy the machinery
  releases post-call. These regimes are pinned by
  `stringname-lifecycle-management`, `borrow-echo-detection`, and
  `variant-copy-ownership`; the native-`string` row is unpinned.
- Copy-constructor primitives exist for `StringName` and `NodePath`
  (`StringNameCopyConstructor`/`NodePathCopyConstructor` in
  `pkg/builtin/char_string.go`, using builtin `constructor_1` method
  bindings). `globalStringMethodBindings.constructor_1` is already resolved
  in generated bindings, so `String` needs only the wrapper.
- AGENTS.md house rule: ownership of refcounted types transfers only via
  the GDExtension copy constructor, never raw byte copies.

## Goals / Non-Goals

**Goals:**

- One authoritative table: for each signature type x variant type, accepted,
  converted, or rejected - and the ownership regime per position.
- Retention of a borrowed value is a one-call, allocation-free operation.
- Every ratified row has a demo fixture and passes the existing leak and
  orphan gates.

**Non-Goals:**

- Designing the Go-calls-engine-side call ergonomics (generated
  `StringName` parameters) - untouched.
- Changing any decode/encode behavior, including rejection semantics for
  non-convertible variants (owned by `varcall-argument-decode`).

## Decisions

**D1: Method name `Copy()`, not `Own()`.** `Copy()` is the Go-idiomatic
clone verb and matches the copy-constructor it wraps; `Own()` advertises a
lifetime contract the method cannot enforce (the caller must still
`Destroy()`). Parity: godot-cpp expresses the same operation as the copy
constructor.

**D2: Implement all three `Copy()` methods on the builtin `constructor_1`
ptrcall (copy constructor).** Alternative considered: the manual idiom
`New<Kind>WithUtf8Chars(x.ToUtf8())`, which costs 5-6 interface crossings,
two engine allocations, a UTF-32/UTF-8/UTF-32 round trip and an interning
hash (~microseconds) versus one ptrcall and a refcount increment
(~sub-100ns). The copy constructor is also the only AGENTS.md-sanctioned
ownership transfer and has zero fidelity concerns for content Godot strings
can hold beyond UTF-8 round-trip validity. For `NodePath` (not interned)
the copy shares the object with a refcount bump - semantically "clone",
matching godot-cpp.

**D3: Conversion breadth is the engine's conversion table, not stringify.**
A `string` parameter accepts exactly what the engine's strict conversion
rules convert to `STRING`: `StringName` and `NodePath` arrive as their text
(pinned by fixtures in both call styles); a variant with no `STRING`
conversion (e.g. `int`) is rejected as an invalid call argument at the
engine's call layer before the Go method executes - verified empirically,
`call("test_string_param_text_conversion", "42", 42)` errors with "Cannot
convert argument 3 from int to String" and never enters the binder. Rejected
alternatives: binder-side stringify leniency for arbitrary variants (the
binder never sees them, and adding it would mean intercepting engine
rejections) and strict rejection of `StringName`/`NodePath` (behavior
change, diverges from the engine).

**D4: Mixed-type safety relies on engine-side call routing; fixtures verify
it rather than new checks.** GDScript/GDExtension route mismatched
direct-call arguments through the converting path before the Go ptrcall
decode assumes the slot type, so the binder adds no runtime variant-type
validation. New fixtures (`StringName`/`NodePath`/int into `string`
parameters, direct + generic call) pin the routing assumption empirically;
if a mismatch ever reaches the Go slot raw, the fixture fails loudly at the
leak or equality gate instead of silently corrupting.

**D5: Fixtures follow the failure-count convention.** New
`TestStringBoundary*` methods in `test/pkg/example.go` return `int64`
failure counts asserted `== 0` from GDScript where multiple crossings are
batched; single-value echoes use `assert_equal` as existing string tests do.
Retention fixture stores a `Copy()` on receiver state and re-reads it in a
later frame, proving post-call validity across the shutdown leak gate.

**D5a: Releasing a struct-field-retained cell must not pass the field
address to cgo.** `make test` runs with `GODEBUG=cgocheck=1`; when a cgo
argument points anywhere inside a Go heap object, the runtime
(`cgoCheckUnknownPointer`) scans the entire object and panics if any tracked
pointer slot holds an unpinned Go pointer. `Example` holds slice and
interface fields, so `e.field.Destroy()` panics on the siblings, not on the
retained cell. The fixture moves ownership into a local cell, zeroes the
field, and destroys the local. Value-receiver reads (`ToUtf8`) copy the cell
to the stack and are unaffected. This is a repo-wide cgocheck interaction,
not a `Copy()` defect; the fixture documents the release idiom.

**D6: TODO:199 resolution.** Replace the TODO comment with a pointer to the
`string-boundary-marshalling` spec summarizing the table; no code change at
the site.

## Risks / Trade-offs

- [Ratified leniency hides caller type bugs] → fixtures document exact
  outputs; the spec text makes stringify an intentional contract, so any
  future tightening is a visible, deliberate change.
- [`String.Copy()` wrapper may surface a refcount accounting difference
  from `StringName`'s (String is not interned)] → the leak gate and an
  echo/destroy fixture for `String.Copy()` cover both directions.
- [Engine routing assumption (D4) may differ between Godot versions] →
  fixtures run against the pinned Godot binary in `make test`; a routing
  change fails there, not in user code.
