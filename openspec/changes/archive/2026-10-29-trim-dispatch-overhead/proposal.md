# Proposal

## Why

Engine→Go dispatch now runs typed binds through compiled closures, but the cost
around that closure is larger than the closure: every class callback crosses into
the engine twice and hashes a string to learn a class name that registration
already knew (`pkg/core/classdb_callback.go:376`, `:414`), and every name-string
dispatch builds diagnostic payloads that the default `LOG_LEVEL=WARN` throws away —
including one FFI `variant_stringify` crossing **per argument, twice**
(`pkg/core/method_bind.go:548-553`, `:586-592`, `:637-640`). Removing more
`reflect` without paying off this ceremony first would buy little, and with no
benchmark anywhere in the repository there is currently no number to prove
otherwise either way.

## What Changes

- Add a dispatch cost baseline: engine-free allocation budgets per dispatch shape
  (gating) plus wall-clock timings (recorded, not gating).
- Dispatch-path diagnostics construct their payloads only when the entry threshold
  will write the entry; the duplicated `gdArgs` stringification is removed.
- Class callback prologues (notification, property get and set, property
  validation, property list, revert pair) resolve the target class from
  instance-held registration identity instead of `GetClass()` + `ToUtf8()` + a
  string-keyed map lookup. The teardown tolerance those prologues currently derive
  from the failed lookup is re-based on the instance-freeing guard.
- Typed-nil guards on the outbound object argument and object return paths stop
  using `reflect.ValueOf` (`pkg/builtin/object_arg.go:30`,
  `pkg/gdclassimpl/object_return.go:78`).
- Remove the unused `DecodeReflectTypePtr` / `DecodeReflectVariantPtr` encoder
  methods, which have no callers.
- No public API changes, no bind-signature changes, no engine-visible behavior
  changes.

## Capabilities

### New Capabilities

- `dispatch-cost-baseline`: what each dispatch shape is allowed to cost per call,
  and how the harness detects drift in either direction.

### Modified Capabilities

- `diagnostic-output-channel`: the entry threshold governs payload construction,
  not only whether a written entry appears.
- `notification-virtual-dispatch`, `property-revert-virtuals`,
  `property-access-dispatch`: teardown tolerance is stated against the
  instance-freeing guard instead of against an unregistered class name reported by
  the engine.

## Impact

`pkg/core/classdb_callback.go`, `pkg/core/method_bind.go`,
`pkg/core/method_bind_callback.go`, `pkg/core/to_string.go`,
`pkg/builtin/wrapped.go` (instance identity), `pkg/builtin/object_arg.go`,
`pkg/builtin/variant_encoders.go` and its per-type encoders,
`pkg/gdclassimpl/object_return.go`, `cmd/generate` templates for wrapper nil
checks, and `test/pkg` fixtures plus the `make` targets carrying the budgets.

## Non-goals

- Compiling the name-string bind path at bind time. That is a follow-up change: it
  revises `typed-method-dispatch` and needs this change's baseline to be measured
  honestly.
- Extending typed closures to today's untypeable signatures (`float32`, `Variant`,
  `string`, slices, `Object`/`Ref` parameters).
- Any change to the generated typed closure families, the FFI boundary, or
  allocation behavior of the compiled path (already pinned at zero).
- Gating wall-clock time in CI.
