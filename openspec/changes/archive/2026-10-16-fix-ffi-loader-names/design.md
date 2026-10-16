# Design

## Context

See proposal.md - Why. Implementation-level context:

- `pkg/ffi/ffi.gen.go` is produced from `cmd/generate/ffi/ffi.go.tmpl`, which emits
  `LoadProcAddress("{{ loadProcAddressName $f.Name }}")` where `$f.Name` is the Go
  typedef identifier collected from the clang AST (`GDExtensionInterface*`).
- `loadProcAddressName` (`cmd/generate/ffi/templatefunctions.go`) snake-cases the
  identifier and hand-patches collisions (`_utf_8_`, `_2`, …). The patch table is
  how the four wrong names arose: nobody added entries for `_5`/`_6` when the
  engine grew `classdb_register_extension_class5/6`, `GetUserData` snake-cases to
  `user_data`, and the `GDExtensionInterfaceGetVariantGetInternalPtrFunc` typedef
  kept its legacy identifier while the header's `@name` moved to
  `variant_get_ptr_internal_getter`.
- The clang AST parser (`cmd/gdextensionparser`) does not carry doc comments, but
  the bundled header `godot_headers/godot/gdextension_interface.h` is engine-generated
  with a consistent `@name` doc block above every interface-function typedef
  (179 `@name` entries; coverage confirmed for the four drift cases).
- `loadProcAddressName` has a second consumer (`gdiFunctionName` → the
  `CallFunc_*` wrapper generator); the Go-side API names (`FFI.CallableCustomGetUserData`,
  `CallFunc_GDExtensionInterfaceCallableCustomGetUserData`) derive from the typedef
  identifier, not the load string.

## Goals / Non-Goals

**Goals:**

- Generated load strings equal the engine-published registration names.
- The mapping is self-healing across future `make update_godot_headers_from_binary`
  refreshes, enforced by a test.
- Zero observable change to the Go-side FFI API and to every currently-correct
  generated file except the four fixed strings.

**Non-Goals:**

- Renaming any Go field or `CallFunc_*` wrapper.
- Capturing comments inside the clang AST parser.
- Making any currently-unused wrapper's behaviour observable (no callers added).

## Decisions

**D1 - Source load names from a `@name` text scan of the bundled header.**
`ffi.Generate` reads `godot_headers/godot/gdextension_interface.h` (same path
convention the parser uses) and scans it line-wise: remember the most recent
`* @name <value>` seen, and when an interface-function typedef declaration
`… (*GDExtensionInterfaceX>(` closes, record `X -> value`. The template then calls
a new function that consults this map first.
*Alternative - extend the fixup table*: smallest diff, but it is the exact drift
class that broke; every engine header refresh would need a manual re-audit
(`_5`/`_6` are evidence it was not done). Rejected.
*Alternative - teach the clang parser to emit comments*: larger change, alters the
AST format artifact, and the doc-comment convention makes a text scan sufficient.
Rejected.

**D2 - Map hit wins; legacy transform remains the fallback; builder is lenient.**
`interfaceFunctionName(id)` returns the `@name` value when the map contains `id`,
otherwise falls back to `loadProcAddressName(id)`. The map builder treats a missing
or unreadable header as an empty map, so the existing `TestGenerate` (which runs
against a temp tree with no header) keeps passing unchanged. Coverage of the real
bundled header is enforced by a dedicated test, not by leniency.

**D3 - Regeneration diff budget: exactly four strings.**
`make generate` must change only the four `LoadProcAddress` strings in
`pkg/ffi/ffi.gen.go` (`get_user_data→get_userdata`, `class_5→class5`,
`class_6→class6`, `get_variant_get_internal_ptr_func→variant_get_ptr_internal_getter`).
The audit of all 177 requested names against the engine's `REGISTER_INTERFACE_FUNC`
/ `register_interface_function` inventory already shows the corrected set resolves
completely; any generated diff line beyond these four means either engine
`@name`/registration inconsistency or header drift and must be investigated before
committing.

**D4 - Leave the test gate and its tolerated-warnings spec untouched.**
The `test-harness-integrity` requirement is a permission ("SHALL NOT treat … as a
failure"); it stays satisfied when the warnings vanish and continues to act as a
regression net should one ever return. Only the AGENTS.md known-issues block, which
documents current run output, is updated.

## Risks / Trade-offs

- [Line-wise `@name` adjacency scan mispairs a name with a typedef] → The
  full-load scenario (zero missing-function errors against the pinned engine) plus
  the four-line regeneration-diff audit catch any mispair immediately; the header
  is machine-generated with a stable convention.
- [A future header drops `@name` from some typedef] → Fallback preserves today's
  behaviour for that one entry while the coverage-guard test fails loudly, per the
  spec.
- [Engine `@name` doc drifts from the actually registered symbol] → Possible in
  theory; the regeneration-diff audit cross-checks changed strings against the
  engine registration inventory, and the pinned-engine full-load test verifies
  end-to-end.
- [`loadProcAddressName` shared with `gdiFunctionName`] → Its behaviour is not
  changed; only the load-string site in `ffi.go.tmpl` switches to the map-aware
  function, leaving `ffi_wrapper.gen.*` byte-identical.
