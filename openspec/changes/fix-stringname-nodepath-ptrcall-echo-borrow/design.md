## Context

The container work (`fix-container-arg-refcount-leak`) introduced `destroySource` gating in `GDExtensionTypePtrFromReflectValue` and content-based borrow-echo detection in `isPtrcallBorrowEcho` (`pkg/core/method_bind.go`), but only for the ten container types. The `StringName`/`NodePath` cases still destroy unconditionally (`pkg/core/variant_refect_value.go`, cases at the top of the switch), so echoing a borrowed argument through ptrcall releases a caller-held refcount. The PR #140 review identified this as the same bug class remaining in the basic refcounted types.

Ground truth verified against current code:

- Ptrcall decode of StringName/NodePath args is a byte-copy borrow (no refcount) per the existing capability spec.
- Ptrcall return: `StringNameCopyConstructor(rOut, ...)` / `NodePathCopyConstructor(rOut, ...)` followed by unconditional `inst.Destroy()`.
- `GDExtensionTypePtrFromReflectValue` already receives `destroySource bool`; the flag simply is not consulted by these two cases.
- `isPtrcallBorrowEcho` compares dynamic values with `reflect.DeepEqual` across call arguments (skipping the receiver), gated by a type switch over the ten container types.

## Goals / Non-Goals

**Goals:**
- Echoing a borrowed StringName/NodePath via ptrcall neither leaks nor double-frees.
- Fresh StringName/NodePath returns keep balanced copy-construct + destroy accounting.
- Same detection mechanism as containers — one heuristic, one set of semantics.

**Non-Goals:**
- `String` lifecycle (deferred separately since the original proposal).
- `PackedVector4Array` decode cases (tracked by `add-packed-vector4-array-decode`).
- Refcount-aware (non-content-based) echo detection; the documented bounded-leak limitation applies here as it does for containers.

## Decisions

### Decision 1: Reuse `destroySource` + extend the existing type switch

Gate the two `inst.Destroy()` calls behind the existing flag and add `StringName`/`NodePath` to `isPtrcallBorrowEcho`'s switch. Minimal diff, identical semantics to containers.

**Alternative considered:** pointer-identity comparison (compare data pointers instead of bytes). More precise in principle, but changes the detection contract mid-flight and risks false negatives if Godot hands back equivalent-but-distinct handles; rejected for consistency.

### Decision 2: Test fresh returns with different contents, not clones

Godot interns StringNames: two independently constructed StringNames with the same name can share storage, making a same-content clone byte-identical to the borrow and thus subject to the documented echo-misclassification leak (as with container defensive clones). NodePath values are distinct buffers. To keep tests deterministic:
- Echo methods return the received argument itself (trivially an echo → not destroyed).
- "Fresh" methods construct a value with *different* contents than the argument (guaranteed non-echo → destroyed after encoding).

The interning nuance is recorded as a known limitation mirroring the container capability's byte-identical-return scenario.

### Decision 3: Sensitivity verification mirrors the container change

Temporarily disabling `isPtrcallBorrowEcho` must surface orphan/double-free symptoms on echo round-trips (proving the tests exercise the guard), then be reverted. This is the same red-check pattern used when the container tests were added.

## Risks / Trade-offs

- **Content-based misclassification leak (documented)** → an unmutated same-content clone returned by ptrcall may be treated as an echo and leak one reference per call. Mitigation: already the accepted, documented limitation for containers; spec scenario records it for these types.
- **Interning ambiguity in future tests** → someone adding a "clone with same content" test would unknowingly test the leak path. Mitigation: Decision 2 rationale documents why test fresh-returns use different contents.

## Migration Plan

Single commit behind no feature flag; behavior changes only for Go methods that currently echo borrowed StringName/NodePath args through ptrcall (today they corrupt caller-held refcounts). Rollback is a revert.

## Open Questions

- None.
