# Design: support-trailing-default-arguments

## Context

See proposal.md for motivation. Current state: `GoMethodMetadata.Call` (method_bind.go) fills arguments leading-index — `slot i ← gdArgs[i]` if `i < supplied`, else `DefaultArguments[i]` if `i < defaults`. `DefaultArguments` is stored positionally from index 0, so the model only produces correct results when `defaults == declared` (every parameter defaulted). `classifyVarcallArity` gates too-few as `max(supplied, defaults) < declared`, the mirror of that leading fill. The engine registers the defaults as trailing values (Godot convention) and passes the caller's raw argument count to the varcall callback with no pre-validation, so the extension owns both the gate and the fill.

## Goals / Non-Goals

**Goals:**
- Make the fill express Godot's trailing-default convention for any `defaults <= declared`.
- Keep the arity gate mechanically equal to the fill's satisfiability (the lockstep invariant from the prior change).
- Make misuse of the exported `Call` by direct Go callers loud rather than silent.
- Preserve the all-defaulted `DefArgs` behavior exactly.

**Non-Goals:**
- No change to type-convertibility or the ptrcall path (see proposal Non-goals).
- No new public call API; the direct-Go check lives inside `Call`.

## Decisions

### D1 — Trailing-convention fill

Introduce `defaultStart = declared - defaults`. Fill each slot:

```
slot i:  i < supplied          → gdArgs[i]                       (caller, leading)
         i >= defaultStart     → DefaultArguments[i-defaultStart] (trailing default)
         otherwise             → unfilled
```

This maps `DefaultArguments[k]` to parameter `declared - defaults + k`, matching godot-cpp's `call_with_variant_args_dv` (`args[i] = default_values[i + dvs - declared]`). When `defaults == declared`, `defaultStart == 0` and the mapping is identical to today, so `DefArgs` is unchanged.

Alternative considered: keep leading fill and require bindings to pre-pad `DefaultArguments` to the front — rejected because it contradicts how the engine registers trailing defaults and pushes the burden onto every binding author.

### D2 — Arity gate becomes `declared - supplied > defaults`

Too-few iff the number of missing arguments exceeds the available defaults. This is the exact negation of D1's satisfiability: a slot is unfilled iff `supplied < defaultStart`, i.e. `declared - supplied > defaults`. The `TestClassifyVarcallArityMatchesCallFill` sweep is rewritten against the trailing fill condition so the two stay mechanically tied (the lockstep invariant). The `expected` field keeps the declared-count value established by the prior refine change; only the satisfiability boundary moves.

### D3 — Direct-Go callers panic on an unfilled slot

Because the varcall path pre-validates via `classifyVarcallArity`, by the time it calls `Call` every slot is fillable; the unfilled branch is reachable only by a direct Go caller bypassing the gate. Make that branch `log.Panic` naming the bound method and the unfilled index, replacing the current silent zero-value `Variant`. A compiled Go caller under-supplying a required argument is a programming error, so a loud panic is appropriate and never fires on the engine path.

Alternative considered: a separate `CallChecked(inst, args) (Variant, error)` entry point — rejected as YAGNI; it duplicates `Call` and requires callers to opt in, leaving the silent path reachable.

### D4 — Bind-time `defaults > declared` guard

The trailing formula needs `defaultStart >= 0`. Reject `len(defaultArguments) > argumentCount` at method-registration time with a clear panic naming the method. Under the old leading fill this case was silently ignored (extra defaults unreachable); making it fatal prevents a negative-index crash in the new fill.

## Risks / Trade-offs

- **[Fill-model change touches the hot dispatch path]** → Mitigation: the change is a bounded index remap in one loop; the all-defaulted path is provably identical and covered by the existing `DefArgs` engine trio.
- **[Panic on direct `Call` could surprise existing Go callers]** → Mitigation: no in-repo caller under-supplies `Call` (the varcall path pre-validates); the panic only fires on genuine misuse, and the message names the method and slot.
- **[Two pending changes touch `classifyVarcallArity`]** → Mitigation: this change alters only the too-few boundary expression, not the `expected` value the refine change owns; sequence the refine change first to avoid a sync collision.

## Open Questions

- Whether the direct-Go panic should later be softened to a returned error if a non-fatal validated entry point gains demand. Deferrable: it does not change the specs, the fill model, or the task breakdown.
