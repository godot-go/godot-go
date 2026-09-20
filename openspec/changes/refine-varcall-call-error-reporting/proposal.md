# Proposal: refine-varcall-call-error-reporting

## Why

Code review of the varcall call-error reporting (PR #147, capability `method-call-error-reporting`) surfaced five follow-up issues: the too-few `expected` count deviates from the godot-cpp reference, the arity check silently hardcodes the leading-index default-fill model (a drift hazard if trailing-default semantics are ever fixed), the "unreachable" claim on `GoMethodMetadata.Call` only holds for the varcall path, rejected calls leave no debug-trace, and two documentation statements are inaccurate about which conversions `variant_can_convert_strict` actually allows.

## What Changes

- Align the `TOO_FEW_ARGUMENTS` `expected` field with godot-cpp's `call_with_variant_args_dv`, which reports the **declared** argument count (not `declared - defaults`). The engine surfaces `expected` in user-visible error text; parity with godot-cpp extensions is the project's reference standard.
- Add lockstep cross-reference comments between `classifyVarcallArity` and `GoMethodMetadata.Call`'s leading-index default fill: document that a partial trailing-default binding (`0 < defaults < declared`) is rejected as `TOO_FEW_ARGUMENTS` under the current model, and that fixing the fill model to the trailing convention requires changing the check to `declared - supplied > defaults` in the same commit.
- Extend the `Call` unfilled-slot comment to note the exported-surface caveat: direct Go callers bypass varcall validation and would receive zero-value `Variant`s (≈ NIL) in unfilled slots instead of the removed panic.
- Emit a debug-level log on every varcall rejection (method name, error code, `argument`, `expected`) so rejects are visible in debug logs; today only the engine sees them.
- Documentation corrections:
  - `docs/overview.md`: clarify that string→number and `nil`→scalar arguments are **rejected** (STRING is not in the engine's strict-convert source list for INT/FLOAT/BOOL; NIL converts only to OBJECT), while numeric inter-conversion, any→STRING, and `nil`→object still pass.
  - Correct the archived `surface-method-call-errors` design.md statement that "`defaultArguments` is nil for every existing binding" — `DefArgs` is bound with two defaults.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

- `method-call-error-reporting`: too-few `expected` becomes the declared argument count (godot-cpp parity); rejected calls additionally emit a debug-level diagnostic.

## Impact

- `pkg/core/method_bind_call_error.go` — `classifyVarcallArity` returns the declared count for too-few; lockstep comments.
- `pkg/core/method_bind_call_error_test.go` — update expected counts for too-few cases.
- `pkg/core/method_bind_callback.go` — debug log on reject paths.
- `pkg/core/method_bind.go` — comment-only caveat on `Call`.
- `docs/overview.md`, archived `openspec/changes/archive/2026-09-11-surface-method-call-errors/design.md` — documentation corrections.
- No engine-harness behavior change beyond the `expected` field value; `test/demo` assertions do not inspect `expected` and stay green.

## Non-goals

- No fix to the leading-index default-fill model itself (trailing-default support for partial defaults remains a separate change).
- No change to which calls are accepted vs rejected (the `variant_can_convert_strict` predicate set is unchanged).
- No ptrcall-path error reporting (the typedef carries no error slot).
- No new user-facing API for returning errors from bound Go methods.
