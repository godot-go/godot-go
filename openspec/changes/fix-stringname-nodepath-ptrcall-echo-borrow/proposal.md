## Why

StringName/NodePath values received as ptrcall arguments are byte-copy borrows holding no refcount, but their ptrcall return encoding (`GDExtensionTypePtrFromReflectValue`) unconditionally calls `inst.Destroy()` after copy-constructing into the output buffer. A Go method that echoes a received StringName/NodePath argument therefore releases a refcount Godot's caller still holds — the exact destroy-the-borrow bug class the container work fixed via `isPtrcallBorrowEcho` + the `destroySource` flag, which was never extended to these two types (flagged in PR #140 review).

## What Changes

- Gate the `StringName`/`NodePath` `Destroy()` calls in `GDExtensionTypePtrFromReflectValue` on the existing `destroySource` flag.
- Extend `isPtrcallBorrowEcho` (`pkg/core/method_bind.go`) to recognize echoed `StringName`/`NodePath` returns.
- Add echo (return the received argument) and fresh-return (return a differently-valued StringName/NodePath) example methods with GDScript round-trip assertions in both call styles.
- Sensitivity-verify: disabling the detection must reproduce orphan/double-free symptoms; re-enable and confirm clean.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `stringname-lifecycle-management`: ptrcall return cleanup gains a borrow-echo exception — an echoed borrowed argument is copied into the output buffer without destroying the source; a new requirement pins the echo-detection contract for these types.
