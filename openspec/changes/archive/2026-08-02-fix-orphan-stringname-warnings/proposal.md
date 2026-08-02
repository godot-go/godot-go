## Why

`make test` exits with 5 orphan StringName warnings despite a prior archived
change (`fix-make-test-orphan-stringnames`) that claimed to eliminate them.
Two subsequent commits reintroduced the warnings:

1. `b4a0880` reverted the return-encoding Destroy calls for StringName/NodePath
   in `variant_refect_value.go`, dropping back to byte-copy without cleanup.
2. `6450f9c` changed varcall/ptrcall argument decoding from borrow semantics
   (byte-copy+Destroy) to ownership semantics without adding cleanup, leaking
   refcounts on every incoming StringName/NodePath argument.

## What Changes

- Restore `inst.Destroy()` after varcall return encoding in
  `GDExtensionVariantPtrFromReflectValue`
- Restore copy-constructor + Destroy in ptrcall return encoding in
  `GDExtensionTypePtrFromReflectValue`
- Revert varcall argument decode to borrow semantics: `ToStringName()` +
  byte-copy + `Destroy()`
- Revert ptrcall argument decode to borrow semantics: byte-copy only (no
  copy-constructor, no ownership transfer)

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `stringname-lifecycle-management`: Restore compliance with the requirement
  "No Orphaned StringName Warnings Remain" — spec already mandates zero orphans
  at `make test` exit but the implementation drifted out of compliance.

## Impact

- `pkg/core/variant_refect_value.go` — return encoding for StringName/NodePath
- `pkg/core/method_bind_reflect.go` — varcall/ptrcall argument decoding
- No API changes, no Godot version impact, no GDScript-side impact

## Non-goals

- Not introducing new ownership-tracking infrastructure for Go-side
  StringName/NodePath lifecycle
- Not changing the encoder subsystem in `variant_builtinclass_encoder.go`
- Not modifying `char_string.go` (copy constructors already exist there)
