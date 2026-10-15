# Proposal

## Why

All three string-shaped Go types (`string`, `String`, `StringName`, plus
`NodePath`) already cross the bound-method boundary in both call styles, but
the policy deciding *which conversions are legal* and *who owns what* is
split across code comments, a stale `TODO` at
`pkg/core/method_bind_reflect.go:199`, and two specs that cover only the
engine-type borrow cases. The native-`string` row of the table is unspecified,
retaining a borrowed value past a call has no first-class API, and Phase 1
items A1/A3/C1/B2 all inherit this table, so it must become explicit now
(roadmap item A2).

## What Changes

- Ratify the string marshalling table as a spec: for each Go signature type
  (`string`, `String`, `StringName`, `NodePath`), the accepted GDScript
  variant types, conversion behavior on mismatch, registration metadata, and
  the ownership regime (retainable copy / call-scoped borrow /
  machinery-released owned decode).
- Add owned-copy helpers `Copy()` on `String`, `StringName`, and `NodePath`
  implemented via the GDExtension copy constructor (the existing
  `StringNameCopyConstructor` / `NodePathCopyConstructor` pattern), giving
  users a cheap, safe way to retain a borrow past the call.
- Add demo fixtures pinning mixed-type crossings: `StringName`/`String`
  values arriving at a Go `string` parameter and vice versa, in both ptrcall
  and varcall styles, with the leak gate as acceptance evidence.
- Resolve the `:199` TODO comment by pointing it at the ratified spec.

No breaking changes: no existing signature, conversion, or ownership behavior
changes; all additions.

## Capabilities

### New Capabilities
- `string-boundary-marshalling`: the conversion matrix and ownership regimes
  for string-shaped values (`string`, `String`, `StringName`, `NodePath`) at
  the GDScript-to-Go bound-method boundary, including retention helpers for
  borrowed engine string values.

### Modified Capabilities

None. `stringname-lifecycle-management`, `borrow-echo-detection`, and
`variant-copy-ownership` already pin the engine-type borrow behaviors this
change ratifies; the native-`string` row is genuinely new specification
surface and lands in the new capability.

## Impact

- `pkg/builtin/char_string.go` (and/or a companion file): `String.Copy()`,
  `StringName.Copy()`, `NodePath.Copy()` wrappers over the copy constructors;
  `StringCopyConstructor` wiring if absent.
- `pkg/core/method_bind_reflect.go`: TODO comment resolution only; no
  behavior change.
- `test/pkg/example.go` + `test/demo/main.gd`: mixed-type string fixtures;
  leak-gate and orphan-warning gates must stay green.
- `docs/roadmap.md`: mark A2 DONE and check its Phase 1 checkbox.

## Non-goals

- No changes to decode/encode/return machinery behavior (borrows,
  destroy-after-encode, and borrow-echo detection stay as specced).
- No native-Go-string ergonomics on the Go-calls-engine side (generated
  `Set(Name StringName)`-style signatures are untouched).
- No `Array`/`PackedStringArray` element marshalling (roadmap A1, inherits
  this table).
- No automatic owned-copy or retention tracking of borrows; `Copy()` remains
  manual with explicit `Destroy()`.
