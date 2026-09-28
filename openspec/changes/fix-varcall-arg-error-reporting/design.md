# Design

## Context

See `proposal.md` — Why, for the defect and the verified reachable set.

The chain today, with the error channel missing in the middle:

```
GoCallback_MethodBindMethodCall(inst, args, rReturn, rError)   // owns rError
  └─ bind.Call(inst, args...) Variant                          // no error return
       └─ reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs(...)
            └─ convertVariantToGoTypeReflectValue(arg, type) (Value, error)
                 └─ if err != nil { log.Panic(...) }            // aborts the host
```

The per-argument converter already returns `(reflect.Value, error)`. The panic is in the loop that calls it, not in the converter. So the fix is plumbing, not new decoding logic.

Two constraints shape the plumbing:

- `GoMethodMetadata.Call` is **exported**. Only one production caller exists (`method_bind_callback.go:90`) plus one test, but it is public API of a library.
- Arguments are decoded left to right, and containers decode as **owned** copies. `Call` releases them at line 410, after the return value is encoded, because an echoed return must get its own reference first. A reject that happens at argument 3 therefore has arguments 1–2 live and owned.

## Goals / Non-Goals

**Goals:**

- A caller-induced decode failure becomes an engine-visible `INVALID_ARGUMENT`, never a process abort.
- No leak on the reject path, and no double-release against the success-path cleanup.
- `GoMethodMetadata.Call`'s exported signature unchanged.

**Non-Goals:**

- No `recover()` anywhere. The fix works by never panicking, not by catching panics.
- No ptrcall error reporting — its ABI has no call-error slot (see proposal Non-goals).
- No change to the strict-convert pre-pass or which calls it accepts.

## Decisions

### Add an error-returning entry point rather than changing `Call`

Add `func (md *GoMethodMetadata) CallWithError(inst GDClass, gdArgs ...Variant) (Variant, error)`. The varcall callback uses it. `Call` becomes a thin wrapper that calls `CallWithError` and panics on error — preserving today's behaviour for direct Go callers, who already bypass varcall validation and are documented as misusing the surface if they under-supply arguments.

**Rejected — change `Call` to return `(Variant, error)`:** a breaking change to a library's public API for the benefit of one internal caller. The wrapper costs four lines.

### The decoder unwinds its own partial work

`reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs` returns `([]reflect.Value, error)`. On failure it releases the owned container values decoded so far — reusing the existing `destroyOwnedContainerArgs` over the *successfully decoded prefix only* — and returns the error. It does not return the partial slice.

Ownership rule: **the decoder owns the prefix on failure; `Call` owns the full set on success.** Exactly one owner per path, so the two cleanup sites can never both fire.

**Rejected — return the partial slice and let `Call` unwind:** spreads knowledge of which values are owned across the boundary, which is precisely how double-frees get written.

Note `destroyOwnedContainerArgs` calls `arg.Interface()`, which panics on an unset `reflect.Value`, so the prefix must be the successfully-decoded entries, not a zero-filled slice.

### Wrap the failure in a typed error carrying index and parameter type

The loop knows the index and the declared type, so it emits:

```go
return nil, &VarcallArgDecodeError{Index: i, ParamType: t.String(), Err: err}
```

The callback reads `Index` directly and gets the expected variant type from `bind.gdeArgumentTypes[i]`, then calls the existing `rejectVarcallInvalidArgument`.

**Rejected — string-parse the error message for the index:** brittle and silly when the value is already in hand. Wrapping with `%w` keeps `errors.Is`/`errors.As` working, matching the style already used at `method_bind_reflect.go:231`.

### Classify failures by inducibility, not by call site

Only the object paths become reported errors:

| Site | Failure | Disposition |
|---|---|---|
| `:240` | `ObjectFromVariant` resolution error | **reported** — caller chose the object |
| `:261` | `unsupported object class %q for parameter type %s`, class name not registered | **reported** — caller chose the object |
| `:273` | class name registered but the wrapper does not implement the declared type | **reported** — caller chose the object |
| `:221` | missing `Ref` constructor for the declared type | **fatal** — author's binding is wrong |
| `:282` | unsupported interface kind | **fatal** — author's binding is wrong |
| `:459` | unsupported Go type | **fatal** — author's binding is wrong |

This follows the existing spec rule that only faults the caller cannot induce remain fatal.

**`:273` was not in the original plan.** This change's proposal and first draft of this
design both treated `:261` as *the* wrong-class trigger. It is not. `:261` only
fires when the class name is absent from `GDNativeConstructors`, and that map holds
every built-in Godot class -- `Label` included. So a `Label` passed to a parameter
declared `Sprite2D` cleared `:261`, the fallback built a `LabelImpl`, and returned
it without checking it satisfied `Sprite2D`. The mismatch then detonated one frame
later inside `reflect.Value.Call`:

```
panic: reflect: Call using *gdclassimpl.LabelImpl as type builtin.Sprite2D
    at method_bind.go (GoMethodMetadata.CallWithError)
```

That is outside the decoder and outside the returned-error contract, so the
error-return work alone left the headline scenario unfixed. The implements check at
`:273` closes it. The lesson is worth keeping: the decoder's job is not finished
when it has produced *a* value, only when it has produced one the callee can
accept.

### Rejection logs at debug level like its siblings

The decode reject emits the same shape of debug diagnostic that `rejectVarcallArity` and `rejectVarcallInvalidArgument` already produce — method name, error kind, `argument`, `expected` — so rejects stay observable in the extension's own logs. This matters more here than for arity: a wrong-class object used to be a loud crash, and a quiet rejection could read as the method silently not running.

## Risks / Trade-offs

- **[Double-release between the unwind and the success-path cleanup] → Mitigation:** the single-owner rule above, plus the test harness's existing leaked-engine-object detection, which fails the run on a leak, and a targeted test that a rejected call leaves container refcounts balanced.
- **[A wrong-class object now fails quietly where it once crashed loudly] → Mitigation:** debug diagnostic naming the method, argument index, and parameter type; the engine also surfaces the call error to the caller.
- **[`destroyOwnedContainerArgs` panics on an unset `reflect.Value`] → Mitigation:** unwind only the decoded prefix, never a zero-filled slice; covered by a test that fails on a later argument after an owned container argument.
- **[Two entry points (`Call` / `CallWithError`) invites drift] → Mitigation:** `Call` delegates to `CallWithError`, so there is one implementation; the wrapper's panic message points at the direct-caller contract.
- **[Trade-off: ptrcall keeps its abort behaviour] →** deliberate and recorded in the proposal; the ABI gives us nowhere to report, so pretending otherwise would be worse.

## Migration Plan

1. Add `VarcallArgDecodeError` and `CallWithError`; convert the decoder loop to return errors and unwind its prefix.
2. Point the varcall callback at `CallWithError` and map the typed error to `rejectVarcallInvalidArgument`.
3. Leave `Call` as a delegating wrapper; no external caller changes.
4. Run the full harness; confirm the previously-fatal scenarios now reject and the suite stays green.
5. Rollback is reverting the three files; no generated-code or data migration involved.

## Open Questions

- Whether `CallWithError` should be exported or kept unexported for in-package use only. Deferrable — it does not affect the spec or the task breakdown, and can be decided when the API surface is next reviewed.
