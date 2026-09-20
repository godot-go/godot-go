# Proposal: support-trailing-default-arguments

## Why

The varcall default-fill uses a leading-index model (`slot i ← DefaultArguments[i]`), which can only express defaults when *every* parameter is defaulted. A binding with partial defaults (`0 < defaults < declared`) is unreachable: the arity gate forces the caller to supply all arguments, and the fill would drop a trailing default into the wrong (leading) slot. This diverges from Godot's trailing-default convention and godot-cpp's `call_with_variant_args_dv`, and leaves direct Go callers of the exported `Call` silently receiving zero-value `Variant`s in unfilled slots.

## What Changes

- Switch the default fill in `GoMethodMetadata.Call` to the **trailing convention**: the bound defaults array maps to the last N parameters; an omitted slot `i` fills from `DefaultArguments[i - (declared - defaults)]`.
- Change the too-few arity gate to `declared - supplied > defaults` (missing exceeds available defaults), matching the trailing fill's satisfiability.
- Add a **bind-time guard**: binding more defaults than declared parameters is rejected with a clear panic.
- Add **direct-Go validation**: a direct call to the exported `Call` that leaves a slot unfilled (bypassing varcall pre-validation) now panics with the method name and unfilled index instead of silently zero-filling.
- Backward compatible: the all-defaulted `DefArgs` binding keeps identical behavior.

## Capabilities

### New Capabilities
- `method-default-arguments`: how bound default values map to a method's parameters (trailing convention), how omitted slots are filled on the varcall path, the bind-time default-count guard, and validation of direct Go callers of the exported `Call`.

### Modified Capabilities
<!-- None. The existing method-call-error-reporting requirement already phrases the too-few condition as "cannot be satisfied even after applying trailing default arguments"; this change makes that satisfiability precise via the new capability without altering the reporting requirement or its expected-field value. -->

## Impact

- `pkg/core/method_bind.go` — trailing fill loop, unfilled-slot panic, bind-time `defaults > declared` guard.
- `pkg/core/method_bind_call_error.go` — too-few gate becomes `declared - supplied > defaults`.
- `pkg/core/method_bind_call_error_test.go` — updated boundary cases, trailing fill-model sweep, guard test.
- `test/pkg/example.go`, `test/demo/main.gd` — new `PartialDefArgs` binding and end-to-end assertions.
- `docs/overview.md` — trailing-default semantics.

## Non-goals

- No change to which argument types are convertible.
- No ptrcall-path changes (the typedef carries no error slot).
- No new public "validated call" API; validation is enforced inside the existing `Call`.
- No migration for existing bindings — none use partial defaults today.
