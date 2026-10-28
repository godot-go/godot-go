# Proposal

## Why

Every engine→Go call — ptrcall from GDScript, every virtual callback (`_process`, `_draw`, notifications), every varcall — routes through `reflect.Value.Call`: arguments are re-built as `[]reflect.Value` (boxing each value), call frames are reconstructed per call, `GoMethodMetadata.String()` and `ReflectValueSliceToString` format eager log fields for messages typically discarded, and `reflect.DeepEqual` boxes arguments/return on every return to classify borrow echoes. This is a fixed allocation tax on the engine→Go half of every frame in a Go-backed game, paid most heavily on per-frame virtual dispatch. Removing it also yields compile-time-checked registration (method expressions instead of method-name strings resolved via `t.MethodByName`).

## What Changes

- Add a compiled invoke-closure field to `GoMethodMetadata` so engine→Go ptrcall and virtual dispatch invoke typed Go functions directly, with zero per-call reflection (no `[]reflect.Value` building, no `reflect.Value.Call` frame, no boxing).
- Generate the wrapper family (per argument arity × return style × static/receiver, plus variadic and trailing-default-fill variants for varcall) using the existing `.tmpl` code machinery; per-type argument decode moves to typed conversion code emitted for built-in types, preserving `varcall-argument-decode` and `ptrcall-object-argument-decode` semantics.
- Re-implement borrow-echo classification on typed values (byte-compare without `reflect.DeepEqual`), preserving every `borrow-echo-detection` scenario.
- Add typed registration API overloads accepting method expressions (compile-time-checked). The existing name-string `ClassDBBindMethod*` API remains supported and keeps reflection on its own path; reflection is otherwise confined to registration time.
- Notification and property-callback dispatch (`classdb_callback.go`) switch to the same compiled closures.

## Capabilities

### New Capabilities

- `typed-method-dispatch`: engine→Go method dispatch (ptrcall, virtual callbacks, varcall) routes through compiled typed closures with a typed, compile-time-checked registration API; per-call reflection is prohibited on hot paths; decode, lifecycle, borrow-echo, default-argument, and error-reporting behaviors are preserved and cross-referenced to the capabilities that pin them.

### Modified Capabilities

None — existing capabilities pin behaviors this change preserves, not the reflection mechanism that currently implements them.

## Impact

- `pkg/core/method_bind.go` (metadata fields, `Ptrcall`, `Call`/`CallVariadic`, echo check), `pkg/core/classdb_callback.go` (callback dispatch), `pkg/core/classdb.go` (typed bind overloads; existing name-string API unchanged).
- `cmd/generate/*.tmpl` templates plus regenerated `pkg/builtin/*.gen.go` / `pkg/gdclassimpl/*.gen.go` for typed decoder methods and the wrapper family.
- Additive Go API; no GDExtension ABI change; no engine behavior change. Existing fixture suites (varcall decode, borrow-echo, argument metadata, string marshalling) provide the parity net.

## Non-goals

- MethodBind/StringName caching on the Go→engine generated path — that path has no reflection and owns most of the measured per-frame garbage (~57 MB/s in dungeon_crawler); separate change.
- Fixing eager debug-log field evaluation in the call paths.
- Removing the name-string registration API or departing from route-B (pure library, no user-project codegen step).
- Reflect-free handling for signatures that cannot be typed at bind time; they keep the reflect fallback.
