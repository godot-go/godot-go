## Context

The existing `stringname-lifecycle-management` spec requires zero orphan
StringName warnings at `make test` exit. Two commits broke compliance:

- `b4a0880` reverted return-encoding Destroy calls in
  `variant_refect_value.go` from `CopyConstructor + Destroy` to
  byte-copy-only, leaking refcounts on Go→Godot returns.
- `6450f9c` changed argument decoding in `method_bind_reflect.go` from
  borrow semantics (byte-copy+Destroy for varcall, byte-copy for ptrcall)
  to ownership semantics (ToStringName for varcall, CopyConstructor for
  ptrcall) without adding post-call cleanup, leaking refcounts on
  Godot→Go arguments.

The copy constructor helpers (`StringNameCopyConstructor`,
`NodePathCopyConstructor`) already exist in `pkg/builtin/char_string.go`.

## Goals / Non-Goals

**Goals:**
- Zero orphan StringName warnings from `make test`
- Correct refcount semantics at all StringName/NodePath GDExtension boundaries
- No double-free hazards

**Non-Goals:**
- No new ownership-tracking infrastructure
- No changes to the encoder subsystem (`variant_builtinclass_encoder.go`)
- No changes to `char_string.go`

## Decisions

### Decision 1: Return encoding uses CopyConstructor + Destroy

For ptrcall return, use `StringNameCopyConstructor(rOut, ...)` followed by
`inst.Destroy()` instead of byte-copy (`*(*T)(rOut) = inst`). The copy
constructor properly increments the Godot-side refcount so the output
buffer holds an owned reference. Destroying the Go-side copy releases
the original refcount, leaving exactly one reference in the output buffer.

For varcall return, the variant constructor (`VariantFromTypeConstructorFunc`)
already increments the refcount internally. We only need to add
`inst.Destroy()` after encoding to release the Go-side refcount.

**Alternative considered:** Byte-copy without Destroy. This would transfer
the pointer but neither increment nor decrement the refcount. If the Go
side created the StringName with +1, Godot's output buffer holds +1. When
Godot destructs the buffer, refcount goes to 0. This works for ptrcall
returns where Godot directly destructs the buffer, but fails for varcall
returns where the variant constructor adds an extra +1. Using
CopyConstructor + Destroy is consistently correct for both paths.

### Decision 2: Argument decoding uses borrow semantics

Arguments received from Godot live in Godot-managed memory (varcall
variants or ptrcall const-type pointers). The Go side should not acquire
a refcount that outlives the method call. Borrow semantics achieve this:

**Varcall decode:** `arg.ToStringName()` → byte-copy → `Destroy()`.
The copy constructor (+1) is immediately released (-1) after borrowing
the pointer via byte-copy. The borrowed value shares Godot's refcount.

**Ptrcall decode:** Direct byte-copy of the const-type pointer. No copy
constructor call, no refcount change. The Go byte array holds the same
pointer Godot passed.

**Alternative considered:** CopyConstructor without Destroy (ownership).
This would leak a +1 refcount for every argument decoded, causing orphans
for every method call with StringName/NodePath arguments. Rejected.

### Decision 3: No change to encoder subsystem

The `createBuiltinClassEncoder` in `variant_builtinclass_encoder.go`
generates decode/encode functions using Go generics. The decode functions
use byte-copy for all types, including StringName/NodePath. Rather than
special-casing refcounted types in the generic encoder, we handle them
in the calling code (`method_bind_reflect.go`, `variant_refect_value.go`)
where the type is known.

## Risks / Trade-offs

- **Borrow safety**: The borrowed byte-copy in argument decoding holds a
  raw pointer to Godot-managed memory. This is safe because the method call
  runs synchronously within Godot's call stack — Godot's owning reference
  outlives the borrow by construction.

- **Future refactor**: If the encoder subsystem is refactored to handle
  refcounted types natively, the special cases in `method_bind_reflect.go`
  and `variant_refect_value.go` could be removed. This is deferred.
