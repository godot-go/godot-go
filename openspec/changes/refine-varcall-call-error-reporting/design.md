# Design: refine-varcall-call-error-reporting

## Context

PR #147 added varcall call-error reporting (capability `method-call-error-reporting`): `GoCallback_MethodBindMethodCall` validates arity via `classifyVarcallArity` and per-argument convertibility via `varcallArgumentConvertible` before invoking the bound method, writing `GDExtensionCallError` on rejects. Review found five follow-up issues: the too-few `expected` value deviates from godot-cpp, the arity check hardcodes `Call`'s leading-index default-fill model without a lockstep note, the "unreachable" comment on `Call`'s unfilled-slot case only holds for the varcall path, rejects emit no debug diagnostic, and two docs statements misstate which conversions `variant_can_convert_strict` allows. See proposal.md for motivation.

Reference facts verified against sources:

- godot-cpp `call_with_variant_args_dv` (`include/godot_cpp/core/binder_common.hpp`) sets `expected = sizeof...(P)` (the declared count) for **both** too-many and too-few, and fills defaults trailing: `args[i] = default_values[i - p_argcount + (dvs - missing)]`, rejecting only when `missing > dvs`.
- The engine's `Variant::can_convert_strict` (`core/variant/variant.cpp`) lists BOOL/INT/FLOAT as mutually strict-convertible, STRING as a valid target from many types, but **not** STRING as a strict source for INT/FLOAT/BOOL (commented out), and NIL converts strictly only to OBJECT.
- The engine passes the caller's raw argument count to the extension varcall callback and performs no pre-validation (`GDExtensionMethodBind::call`, `core/extension/gdextension.cpp`), so the extension owns arity/type reporting.

## Goals / Non-Goals

**Goals:**
- Make the reported `expected` value match the godot-cpp reference for arity errors.
- Make the coupling between `classifyVarcallArity` and `Call`'s fill model explicit and drift-proof at the comment level.
- Make every rejection observable in the extension's own debug logs.
- Correct the two inaccurate documentation statements.

**Non-Goals:**
- Changing which calls are accepted vs rejected (the predicate and satisfiability model are unchanged).
- Fixing the leading-index default-fill model to the trailing convention (separate change; see D2).
- Adding validation for direct Go callers of `GoMethodMetadata.Call` (see D3).

## Decisions

### D1 — Too-few `expected` becomes the declared count

`classifyVarcallArity` currently returns `required = declared - defArgs` (clamped at 0) for `varcallArityTooFew`. Change it to return `declared`, matching godot-cpp's `call_with_variant_args_dv`, which reports `sizeof...(P)` for both arity errors. The engine renders `expected` in user-visible call-error text, so parity means a godot-go extension reports the same number as a godot-cpp extension would. AGENTS.md establishes godot-cpp as the reference implementation whose patterns must be followed.

Alternative considered: keep `required` and merely fix the PR description's parity claim — rejected because the reference implementation is the standard, and the required count remains derivable by callers from the registered default-argument count.

### D2 — Lockstep comments, not a fill-model fix

The too-few condition `max(supplied, defaults) < declared` is deliberately the mirror of `Call`'s leading-index fill (`slot i fills iff i < supplied || i < defaults`). Under the engine's trailing-default convention, a binding with `0 < defaults < declared` and `supplied == declared - defaults` is valid, but the leading-index model cannot fill the trailing slot — so the check rejects it. This is not a regression (pre-PR the same call hit `log.Panic` in `Call`), and no binding in the repo has partial defaults.

Decision: keep the condition and add cross-reference comments at both sites — in `classifyVarcallArity` and in `Call`'s fill loop — stating: (a) the condition must equal `Call`'s fill satisfiability, (b) the trailing-default limitation for partial-default bindings, and (c) that moving the fill model to the trailing convention requires changing the check to `declared - supplied > defaults` and updating `TestClassifyVarcallArityMatchesCallFill` in the same commit.

Alternative considered: fix the fill model to the trailing convention now — rejected as a separate behavior change to default application that needs its own capability delta; bundling it here would obscure the review fixes.

### D3 — Comment caveat on `Call`, no re-added panic

`GoMethodMetadata.Call` is exported; the varcall callback is its only in-repo caller and is arity-guarded, but a direct Go caller passing too few arguments now silently fills remaining slots with zero-value `Variant`s (all-zero bytes ≈ NIL variant, coercing to 0/nil downstream) instead of the removed `log.Panic`. Decision: extend the existing comment to document this exported-surface caveat rather than re-adding a panic, keeping panics reserved for internal faults per the capability's requirement.

Alternative considered: a defensive `log.Panic` on unfilled slots — rejected because it reintroduces the fatal path the capability deliberately removed, and direct-call validation is a distinct concern that can be scoped separately if demand appears.

### D4 — Debug diagnostic inside the reject helpers

Emit `log.Debug` inside `rejectVarcallArity` and `rejectVarcallInvalidArgument` (passing the `*GoMethodMetadata` for the method name), logging the bound method, error kind, and the `argument`/`expected` values written. Placing it in the helpers guarantees every reject path logs uniformly, instead of duplicating log calls at each callback return site. Cost is confined to reject paths; debug level is off by default.

### D5 — Documentation corrections

- `docs/overview.md`: replace the ambiguous "numeric/string inter-conversion ... still pass" phrasing with the precise set: numeric inter-conversion (int/float/bool) passes, any→STRING passes, `nil`→object passes; STRING→number and `nil`→scalar are rejected as `INVALID_ARGUMENT`.
- Archived `openspec/changes/archive/2026-09-11-surface-method-call-errors/design.md`: correct the sentence "`defaultArguments` is `nil` for every existing binding" with a dated correction note — `DefArgs` is bound with two defaults (all-arguments-defaulted), which is exercised by the defaults-satisfy-short-call path. The surrounding reasoning remains valid for that case.

Alternative considered: leave the archive untouched and record the correction only in this design — rejected because the misleading sentence sits in the most-quotable place; a minimal dated correction preserves the historical record while stopping the error's spread.

## Risks / Trade-offs

- **[`expected` value change is externally visible]** Any consumer parsing the too-few `expected` would see `declared` instead of `required`. → Mitigation: matches godot-cpp, which is the compatibility target; the repo's engine harness does not assert on `expected`, and unit tests are updated in lockstep.
- **[Debug log noise on repeated rejects]** A caller loop that repeatedly triggers rejects emits one debug line per reject. → Mitigation: debug level is off by default; rejects are exceptional and each line is small.
- **[Archive edit convention]** Editing an archived change may conflict with strict archive immutability expectations. → Mitigation: single factual correction with a dated pointer to this change; no rewriting of the archived reasoning.
- **[Lockstep drift persists]** Comments only prevent drift if read. → Mitigation: `TestClassifyVarcallArityMatchesCallFill` already mechanically ties the check to the fill condition, so a fill-model change without a check change fails the test.

## Open Questions

- Whether direct Go callers of `GoMethodMetadata.Call` eventually need a validated entry point (D3 alternative). Deferrable: it does not change the specs, this approach, or the task breakdown.
