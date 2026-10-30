# Design

## Context

See `proposal.md` — Why for the motivation. The constraints that shape the how:

- **Resolution today.** Seven class callback prologues — `GetPropertyList`,
  `PropertyCanRevert`, `PropertyGetRevert`, `ValidateProperty`, `Notification`, `Get`,
  and `Set` in `pkg/core/classdb_callback.go` — each do `wci.Instance.GetClass()`
  (engine call returning a Godot `String`), `.ToUtf8()` (UTF-16 copy plus a Go
  allocation), `defer Destroy()` (engine call), then
  `Internal.GDRegisteredGDClasses.Get(className)` (hash the string, locked map).
  `WrappedClassInstance` carries only `Instance Object` (`pkg/builtin/wrapped.go:66`),
  so nothing cheaper is at hand at the boundary. The one callback that already gets it
  right is `GoCallback_ClassCreationInfoToString`, which answers from the Go-side
  `inst.GetClassName()` (`pkg/core/to_string.go:24`) — evidence that the cheaper source
  is available and correct for a registered instance.
- **Teardown protection is a side effect of that lookup.** Three capabilities state
  their teardown tolerance as "the reported class is not a registered GDClass":
  `notification-virtual-dispatch`, `property-revert-virtuals`, and the decline
  scenario inside `property-access-dispatch`'s "An unresolvable class fails a get,
  set, or validation request". Replacing the lookup removes the *trigger* those
  requirements are written against, while their guarantee must survive.
- **Diagnostics are eager.** `log` wraps zap with `Field` values the caller builds
  before the call, and `log.GetLevel()` (`pkg/log/logger.go:103`) is already
  exported. The heaviest payloads (`VariantSliceToString`, `md.String()`,
  `util.ReflectValueSliceToString`) therefore run at every dispatch even though the
  test and demo runs use `LOG_LEVEL=WARN`. `Variant.Stringify()` is
  `CallFunc_GDExtensionInterfaceVariantStringify` — an FFI crossing plus a Godot
  `String` plus a UTF-8 conversion, per element.
- **Allocation budgets gate well, time does not.** Change `typed-method-dispatch`
  found that per-call allocation counts are stable enough to assert with a floor and
  margin, while wall-clock figures are not. The repository has no `Benchmark*`
  functions at all, so there is no timing baseline to compare against.
- `pkg/builtin` cannot import `pkg/core` (the dependency runs the other way).

## Goals / Non-Goals

**Goals:**

- Each dispatch shape has a stated, measured per-call cost ceiling that a target
  enforces, and a recorded timing that nobody enforces.
- Discarded diagnostics cost nothing but a level comparison.
- Class resolution at a callback boundary performs no engine call and no string
  hash of an engine-allocated string.
- Teardown tolerance is expressed against binding state the binding controls, not
  against a coincidence of engine behavior.

**Non-Goals:**

- Making the fallback path compiled (follow-up change), and widening what the typed
  unions accept.
- Reducing the cost of the cgo crossing itself.
- Changing what the diagnostics *say* — only when they are constructed.

## Decisions

**D1 — Carry the class name on the instance, keep the map lookup.**
Add the registered class name to `WrappedClassInstance` at construction so prologues
read a Go string they already own, and keep `GDRegisteredGDClasses.Get(name)` as the
second step. Alternatives: calling `Instance.GetClassName()` per callback (a Go-side
constant, also engine-free, but recomputed and still a hash); storing
`ClassData any` holding `*core.ClassInfo` so the lookup disappears entirely (cheapest,
but puts an opaque cross-package pointer in the shared instance struct and needs an
interface assertion at every use). Chosen: the name now, `ClassData` held in reserve
as a follow-up if the benchmark still shows the hash. Rationale: smallest change to a
struct that every instance allocates, and it keeps `pkg/builtin` unaware of `core`.

**D2 — Re-base teardown tolerance on the freeing guard.**
`FreeInstance` already runs a re-entrancy guard (`class-instance-lifecycle`). Set an
explicit "instance is being freed" state there, before Go-side teardown begins, and
have prologues decline when it is set. Alternatives: keeping `GetClass()` as a second
confirmation (keeps exactly the cost we are removing); relying on the registered-class
lookup to keep failing (it will not — the instance now resolves to its own registered
class). The three affected requirements are MODIFIED so their condition reads against
wrapper state. The loud-failure half of `property-access-dispatch` stays: a get, set,
or validation request for a genuinely unregistered class still fails through the fatal
channel.

**D3 — Make dispatch payloads lazy with `zap.Object` marshaling.**
Give `GoMethodMetadata` and the Variant-slice payloads `ObjectMarshaler`
implementations and pass them as `zap.Object(...)`, so marshaling happens only when
the entry is actually written. Alternatives: a `log.GetLevel()` guard around each call
site (works, but repeats the level test at every site and loses the entry when someone
raises the level mid-session); deleting the entries (they are the only trace of
fallback dispatch, and the fallback is where the interesting bugs live). Extra
arguments at the same site that duplicate each other — the variadic entry stringifies
`gdArgs` twice — are removed outright.

**D4 — Codegen a nil check on the wrapper types instead of `reflect.ValueOf`.**
Emit `IsNil() bool` on each generated wrapper type (`func (n *NodeImpl) IsNil() bool {
return n == nil }`) and use it in place of the reflect-based typed-nil test. Calling a
method declared on the outer pointer type with a nil receiver is legal, which is what
the reflect check exists to do safely; a method promoted through the embedded value
would not be, so it must be declared on the generated type itself. Alternatives:
keeping `reflect.ValueOf` and caching per type (still `reflect`), and testing the raw
accessor (faults on a typed nil before any check can run, as the existing comment
records). The two call sites are documented at ~6.7ns against ~0.47ns for the bare
accessor, so this is a modest win taken mainly because it is free once the generator
emits it.

**D5 — Two-tier cost baseline: allocations gate, time is recorded.**
Engine-free tier: `pkg/core` benchmarks over the dispatch shapes the last change
enumerated (instance/static/virtual, arity 0/2/5, value/void/copy-ret, the callback
closure shapes, and the callback prologues), asserting allocation ceilings with floor
and margin in the style of `typed_dispatch_budget_test.go`, wired into `make test`.
In-engine tier: a demo fixture that records ns/call for the same shapes as
`dispatch_*:` log lines, checked by the output gate only for presence and for
allocation counters. Alternatives: gating CI on wall-clock deltas (rejected: the last
change established the flakiness), and benchmarks with no assertions at all (rejected:
an unread number changes nothing).

## Risks / Trade-offs

- **Teardown deliveries could start reaching handlers.** With the engine lookup gone,
  a pre-delete delivery resolves to a live registration and would run `_notification`,
  `_get`, `_set`, or a revert against a dying instance. Mitigation: land the freeing-guard
  fixtures *before* the hoist, reproducing the observed pre-delete `"Control"` delivery,
  and assert absorption on the new basis; keep the existing three scenarios observable
  under the new condition.
- **The output gate may depend on the very lines being made lazy.** Mitigation: run
  `make test_gate` and the full gate before and after the diagnostics change, and treat
  any check that reads dispatch diagnostics as evidence the diagnostics are load-bearing
  and must be handled deliberately.
- **`IsNil()` is an interface-surface change** for hand-written `Wrapped` implementers,
  including test stubs. Mitigation: enumerate implementers before generating, update
  them in the same commit, and lean on the compile-time `_ Wrapped` assertions to make
  an omission loud.
- **Allocation ceilings drift with Go versions.** Mitigation: floor-plus-margin
  assertions in the established style, regression-direction only.
- **Recorded timings may be read as a gate.** Mitigation: the fixture line states the
  budget status explicitly, as the existing budget fixture does.

## Migration Plan

1. Land the benchmark tier against the current tree — expected green with today's
   numbers recorded as the baseline.
2. Land the lazy diagnostics; re-baseline the fallback allocation figures downward.
3. Land the nil-check codegen and the dead-encoder removal.
4. Land the prologue hoist last, with its teardown fixtures in the same commit; the
   three spec deltas and this change archive together.
Rollback is per-commit; each step is independently revertable because no public API
moves.
