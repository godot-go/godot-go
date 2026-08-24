## Context

The container work (`fix-container-arg-refcount-leak`) introduced `destroySource` gating in `GDExtensionTypePtrFromReflectValue` and content-based borrow-echo detection in `isPtrcallBorrowEcho` (`pkg/core/method_bind.go`), but only for the ten container types. The PR #140 review identified that StringName/NodePath remain exposed to the same destroy-the-borrow bug class. Validating this change against current code surfaced **three** instances of that class:

1. **Ptrcall return** — `GDExtensionTypePtrFromReflectValue` (`pkg/core/variant_refect_value.go`) copy-constructs into the output buffer then unconditionally calls `inst.Destroy()` for StringName/NodePath. Echoing a ptrcall borrow corrupts a caller-held refcount. (Original scope.)
2. **Varcall return** — `GDExtensionVariantPtrFromReflectValue` (same file) encodes into the output Variant then unconditionally destroys. Same corruption via `.call(...)`. Discovered during validation; both call sites of this function live in `GoMethodMetadata.Call`.
3. **Varcall post-call arg loop** — `GoMethodMetadata.Call` (`pkg/core/method_bind.go:347-353`) destroys StringName/NodePath args after every standard varcall. But varcall decode (`convertVariantToGoTypeReflectValue`) is construct → byte-copy → destroy at decode time, i.e. a pure borrow holding nothing — so the loop steals one refcount decrement per call *today* (existing tests included). Added in `89d0772 (#135)` under owned-decode assumptions and made stale when decode switched to borrows.

Other ground truth verified against current code:

- Ptrcall decode of StringName/NodePath args is byte-copy borrow; varcall decode is construct→copy→destroy-at-decode (also a borrow) — so neither call style ever hands Go an owned StringName/NodePath argument.
- `GDExtensionTypePtrFromReflectValue` already receives `destroySource bool`; the flag simply is not consulted by these two cases. `GDExtensionVariantPtrFromReflectValue` has no such parameter yet and exactly two call sites, both in `GoMethodMetadata.Call`.
- `isPtrcallBorrowEcho` compares dynamic values with `reflect.DeepEqual`, gated by a type switch over the ten container types.
- `StringName`/`NodePath` are `[8]uint8` (a single data pointer), so DeepEqual echo detection is trivially sound for true echoes.

## Goals / Non-Goals

**Goals:**
- Echoing a borrowed StringName/NodePath via either return path neither leaks nor double-frees.
- Fresh StringName/NodePath returns keep balanced encode+destroy accounting.
- No post-call machinery step releases borrowed arguments (only owned decodes get released, as with containers).
- Same detection mechanism as containers — one heuristic, one set of semantics.

**Non-Goals:**
- `String` lifecycle (deferred separately since the original proposal).
- `PackedVector4Array` decode cases (tracked by `add-packed-vector4-array-decode`).
- Refcount-aware (non-content-based) echo detection; the documented bounded-leak limitation applies here as it does for containers.

## Decisions

### Decision 1: Reuse `destroySource` + extend the existing type switch

Add `StringName`/`NodePath` to `isPtrcallBorrowEcho`'s switch. Gate the ptrcall cases' `inst.Destroy()` behind the existing flag. For the varcall return path, give `GDExtensionVariantPtrFromReflectValue` a matching `destroySource bool` parameter and compute it in `GoMethodMetadata.Call` with `isPtrcallBorrowEcho(ret[0], args)` before encoding. The variadic branch always passes true: variadic methods receive raw Variants, never hold borrows, so any returned StringName/NodePath is Go-created and its destroy is correct. Minimal diff, identical semantics to containers across both signatures.

**Alternative considered:** pointer-identity comparison instead of bytes. More precise in principle, but changes the detection contract mid-flight and risks false negatives if Godot hands back equivalent-but-distinct handles; rejected for consistency.

### Decision 2: Test fresh returns with different contents, not clones

Godot interns StringNames: two independently constructed StringNames with the same name can share storage, making a same-content clone byte-identical to the borrow and thus subject to the documented echo-misclassification leak (as with container defensive clones). NodePath values are distinct buffers. To keep tests deterministic:
- Echo methods return the received argument itself (trivially an echo → not destroyed).
- "Fresh" methods construct a value with *different* contents than the argument (guaranteed non-echo → destroyed after encoding).

The interning nuance is recorded as a known limitation mirroring the container capability's byte-identical-return scenario.

### Decision 3: Delete the stale post-call arg loop outright

The `case StringName / case NodePath` loop after `md.Func.Call` releases values the Go side never owns under current decode semantics — there is no state in which it is correct, so removal (not gating) is right. This also fixes a live latent bug: every existing varcall taking these types (e.g. `test_string_name_arg_echo`, `test_node_path_arg_echo`) currently steals one refcount decrement per invocation; the green suite masks it because caller-held and interned references absorb the stolen decrement. Removal matches the container convention codified in `destroyOwnedContainerArgs`: release only owned decodes, only after the return value is encoded — borrowed decodes are never released by us.

**Risk noted:** if #135 introduced the loop to silence real orphan warnings that had another cause, removal could resurface them; `make test` output is checked for orphan warnings explicitly (task 3.2) so any such regression is caught immediately.

### Decision 4: Sensitivity verification mirrors the container change

Two temporary mutations, each rebuilt, run, and reverted:
- Disable the new `StringName`/`NodePath` cases in `isPtrcallBorrowEcho` → echo round-trips through either path must surface orphan/double-free symptoms (proves the tests exercise the guard).
- Re-add a destroy of borrowed varcall args (the deleted loop) → repeated varcall echo calls must surface drift/symptoms (proves task coverage of hazard 3).

## Risks / Trade-offs

- **Content-based misclassification leak (documented)** → an unmutated same-content clone returned through either path may be treated as an echo and leak one reference per call. Mitigation: already the accepted, documented limitation for containers; spec scenario records it for these types.
- **Interning ambiguity in future tests** → someone adding a "clone with same content" test would unknowingly test the leak path. Mitigation: Decision 2 rationale documents why test fresh-returns use different contents.
- **Signature change ripples** → `GDExtensionVariantPtrFromReflectValue` gains a parameter; exactly two call sites exist, both in `GoMethodMetadata.Call`. Low blast radius.

## Migration Plan

Single commit behind no feature flag. Behavior changes: echoed borrows are no longer destroyed (fixes caller-side corruption), and borrowed args are no longer spuriously decremented post-call (fixes a live latent bug). Rollback is a revert.

## Open Questions

- None.
