# Proposal

## Why

The generated FFI loader requests GDExtension interface functions by names
reverse-engineered from the Go typedef identifiers (snake-case plus a hand-patched
fixup table). For four functions the reconstructed string never matches the symbol
the engine registers, so their slots load as nil and the engine logs
"Attempt to get non-existent interface function" at startup:

| loader asks | engine registers |
|---|---|
| `callable_custom_get_user_data` | `callable_custom_get_userdata` |
| `classdb_register_extension_class_5` | `classdb_register_extension_class5` |
| `classdb_register_extension_class_6` | `classdb_register_extension_class6` |
| `get_variant_get_internal_ptr_func` | `variant_get_ptr_internal_getter` |

These false alarms are baked into the AGENTS.md "Known Issues" block and into the
test gate's tolerated-warning list, and the `callable_custom` one is the roadmap's
cited risk against C1 (Go func → Callable bridge): correct here removes a phantom
blocker and makes `callable_custom_get_userdata` usable when C2/C4 need to unwrap
a custom Callable.

## What Changes

- The code generator derives each `LoadProcAddress` string from the authoritative
  `@name` doc-comment metadata of the corresponding `GDExtensionInterface*` typedef
  in `godot_headers/godot/gdextension_interface.h`, instead of snake-casing the
  Go identifier.
- The existing snake-case + fixup transform is kept only as a fallback when a
  typedef carries no `@name`; a generator test enforces that every interface-function
  typedef in the bundled header has `@name` coverage, so future
  `make update_godot_headers_from_binary` refreshes cannot silently reintroduce this
  bug class.
- `pkg/ffi/ffi.gen.go` is regenerated; exactly four `LoadProcAddress` strings
  change; all interface slots resolve non-nil and the four startup warnings
  disappear from test runs.
- The AGENTS.md "Known Issues" block drops the three resolved warning entries.

## Capabilities

### New Capabilities

- `interface-function-loading`: the extension resolves every GDExtension interface
  function it generates wrappers for, using the registration names the engine
  publishes in the interface header.

### Modified Capabilities

- (none — the `test-harness-integrity` requirement that tolerates the four startup
  warnings is a permission, not an assertion; it stays satisfied and acts as a
  regression net. No requirement changes.)

## Impact

- Codegen: `cmd/generate/ffi/` (header `@name` scan, template function registration,
  `ffi.go.tmpl` load line).
- Regenerated: `pkg/ffi/ffi.gen.go` (four load strings; generated C wrapper files
  unaffected).
- Tests: new `cmd/generate/ffi` unit tests (map builder, coverage, fallback,
  regression); full `make test` gate goes from four tolerated warnings to zero.
- Docs: AGENTS.md known-issues list.
- No runtime behaviour change for existing code: none of the four affected
  `CallFunc_*` wrappers has any caller today.

## Non-goals

- Implementing the C1 Go-func-to-Callable bridge itself (next stacked change).
- Changing `test/check_test_output.sh` gates or the `test-harness-integrity` spec.
- Removing the legacy snake-case fixup table (it remains the fallback path).
- Teaching the clang AST parser to capture doc comments (a separate header scan is
  used instead).
