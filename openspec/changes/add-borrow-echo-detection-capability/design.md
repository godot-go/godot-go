## Context

Borrow argument decoding (no Go-side refcount) requires return paths to decide whether a returned value owns its refcount before destroying the source. Pointer-identity checks are unreliable here: Godot copy constructors share underlying storage and StringName interns equal names, so "the same logical value" can appear as either identical or distinct pointers depending on construction history. The bindings therefore classify by content (`reflect.DeepEqual` over the dynamic values, skipping the receiver), implemented once in `isPtrcallBorrowEcho` (`pkg/core/method_bind.go`) and consulted by both `GDExtensionTypePtrFromReflectValue` (ptrcall) and `GDExtensionVariantPtrFromReflectValue` (varcall) call sites in `GoMethodMetadata`.

Covered types today: `StringName`, `NodePath`, `Array`, and nine `Packed*Array` types — exactly those whose **ptrcall** arguments decode as borrows, with one exception: `PackedVector4Array`'s ptrcall argument decodes as an owned copy (`NewPackedVector4ArrayWithPackedVector4Array`, scoped to `add-packed-vector4-array-decode`) and its ptrcall return is a raw byte-copy encoder with no source release, so echo suppression is inert for it on that path; it remains in the detection switch harmlessly. Their varcall argument decodes differ per type (`StringName`/`NodePath` decode as borrows; containers decode as owned copies released post-call by `destroyOwnedContainerArgs`), and source-release behavior on return varies by type × path: ptrcall releases all covered types except `PackedVector4Array`; the varcall encoder releases only `StringName`/`NodePath` — container returns via varcall get no return-path source release, a pre-existing behavior deliberately deferred in the container change's Non-Goals and now recorded normatively here rather than described as correct.

## Goals / Non-Goals

**Goals:**
- One normative home for the echo-classification contract and its limitation, spanning types and call styles.
- Make the accepted same-content clone leak explicit, bounded, and actionable for method authors.
- State the type × path release matrix accurately instead of overclaiming uniform destroy-after-encode semantics.

**Non-Goals:**
- Changing detection to ownership tracking or refcount inspection (the precise fix; deferred until the leak matters in practice).
- Fixing or altering the varcall container return-path non-release (belongs to the container capability's deferred concerns; this spec records it as-is).
- Altering any type-specific scenarios already recorded in `container-lifecycle-management` and `stringname-lifecycle-management`; this capability is the general contract they instantiate.
- Covering `Dictionary` (no ptrcall arg decode/return destroy path uses echo suppression today).

## Decisions

### Decision 1: Specify the leak as accepted behavior rather than a bug to fix

The misclassification is inherent to content-based detection: any mechanism that cannot observe refcounts will confuse an unmutated clone with its source. Recording it normatively (bounded to one reference per call, silent, crash-free) tells method authors exactly when it applies and how to avoid it, without committing to ownership-tracking machinery now. If that machinery lands later, this capability's limitation requirement is the one that changes.

## Risks / Trade-offs

- **Spec overlap** → type-specific capabilities repeat parts of this contract in their scenarios. Mitigation: this capability holds the general rules; type specs hold per-type scenarios. Divergence would surface in review of future lifecycle changes.

## Migration Plan

Specification-only addition behind no code change; nothing to migrate. Rollback is a revert.

## Open Questions

- None.
