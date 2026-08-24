## Why

StringName/NodePath values received as arguments are borrows holding no refcount in **both** call styles, but all three release points in the call machinery destroy them unconditionally:

1. **Ptrcall return** (`GDExtensionTypePtrFromReflectValue`): `inst.Destroy()` runs unconditionally after copy-constructing into the output buffer.
2. **Varcall return** (`GDExtensionVariantPtrFromReflectValue`): same unconditional `inst.Destroy()` after encoding into the output Variant.
3. **Varcall post-call arg loop** (`GoMethodMetadata.Call`, `pkg/core/method_bind.go:347-353`): destroys the StringName/NodePath args again even though decode already released its own constructor temp — a spurious refcount decrement on *every* varcall taking these types today (e.g. `test_string_name_arg_echo`, `test_node_path_arg_echo`), added in `89d0772 (#135)` under owned-decode assumptions and made stale when decode switched to destroy-at-decode-time borrows.

A Go method that echoes a received StringName/NodePath argument through either return path therefore releases a refcount Godot's caller still holds — the exact destroy-the-borrow bug class the container work fixed via `isPtrcallBorrowEcho` + the `destroySource` flag, which was never extended to these two types (flagged in PR #140 review; hazards 2 and 3 surfaced while validating this change against current code).

## What Changes

- Extend `isPtrcallBorrowEcho` (`pkg/core/method_bind.go`) to recognize echoed `StringName`/`NodePath` returns.
- Gate the `StringName`/`NodePath` `Destroy()` calls in `GDExtensionTypePtrFromReflectValue` on the existing `destroySource` flag (ptrcall).
- Thread an echo-aware `destroySource` decision through the varcall return path: `GDExtensionVariantPtrFromReflectValue` gains a `destroySource` parameter like its type-ptr counterpart; both call sites in `GoMethodMetadata.Call` updated (the variadic branch always destroys — it never holds borrows).
- Remove the stale post-call StringName/NodePath arg-destroy loop from `GoMethodMetadata.Call` (**fixes a live latent double-release affecting existing tests**).
- Add echo (return the received argument) and fresh-return (return a differently-valued StringName/NodePath) example methods with GDScript round-trip assertions in both call styles.
- Sensitivity-verify per path: disabling detection or re-adding the stale loop must reproduce orphan/double-free symptoms; re-enabled state confirms clean.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `stringname-lifecycle-management`: ptrcall AND varcall return cleanup gain a borrow-echo exception — an echoed borrowed argument is encoded into Godot's output without destroying the source; a new requirement pins echo-detection for these types across call styles, and another pins that the call machinery never releases borrowed arguments after the call.
