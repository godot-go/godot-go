# Task evidence

Running log of measured figures and gate output backing each completed task.

## Task 5.1 — integration gate on the final tree

- `go test ./pkg/...` (excluding `pkg/builtin`'s pre-existing engine-free
  `TestRefAcquireUnrefReleasesExactlyOnce` SIGSEGV, documented by the
  previous change): every package ok.
- `GOTMPDIR=… GODOT=/home/pcting/bin/godot make test`: exit 0,
  `FAILURES: 0`, 22/22 `dispatch_time:` records, `make test_gate` 27/27.
- `Orphan StringName` lines: 0 in `test-output.log` and 0 in
  `editor-output.log`.
- `instance_churn` counters unchanged: `weak_alive=0/2048`,
  phase deltas +332 KiB / +1558 KiB inside the 976 KiB / 4883 KiB bands,
  `live=1 base=1`.
- `make test_dispatch_budget`: green (zero-allocation budgets hold on the
  final tree).

## Task 5.2 — change-level before/after cost table

Before = the group 1 consolidated baseline (`b463cdf`, cool machine).
After = the final group-4 tree (`f31ff8e`), same in-engine timing fixture,
hot machine (load-avg 6–10): every in-engine cell is inflated against the
baseline table, so rows carry the raw figures and the reading is floor- and
gate-relative. Allocation rows are gated and directly comparable.

Allocations per call (in-engine budget fixture, `dispatch_budget:` records;
after-run values identical to the group-2 post-change records — no group
moved an allocation budget):

| shape | before (group 1) | after | gate |
|---|---|---|---|
| compiled ptrcall (all cells) | 0 | 0 | budget 0 |
| compiled varcall (arity 0/2/5, virtual, container, copy-ret echo) | 3–14 (range as recorded) | 3 / 8 / 14 / 3 / 8 / 7 | caps 5 / 12 / 20 / 5 / 12 / 11 |
| reflective ptrcall (arity 0/2/5) | 27–56 (range as recorded) | 8 / 10 / 13 | floor 4, margin ×2 |
| reflective varcall (arity 0/2/5) | 38–249 (range as recorded) | 16 / 25 / 37 | floor 4, margin ×2 |
| engine-free budgets, 30 shapes | 0 | 0 | budget 0 (reflective control 32, floor 16) |

In-engine ns/call (timing fixture, recorded, presence-gated only):

| shape | before | after (hot machine) | reading |
|---|---|---|---|
| static_add_ptrcall | 35.1 | 107.0 | floor cell (~3.0x inflation) |
| value_arity0_ptrcall | 29.8 | 92.5 | floor cell (~3.1x) |
| value_arity2_ptrcall | 34.4 | 113.9 | floor cell |
| value_arity5_ptrcall | 47.7 | 153.2 | floor cell |
| virtual_arity0_ptrcall | 29.8 | 135.1 | floor cell |
| container_arg_ptrcall | 31.8 | 123.2 | floor cell |
| static_add_varcall | 3235.6 | 11484.5 | ~3.5x band |
| value_arity0_varcall | 1136.5 | 4735.2 | ~4.2x band |
| value_arity2_varcall | 3078.5 | 12876.2 | ~4.2x band |
| value_arity5_varcall | 5408.4 | 21774.0 | ~4.0x band |
| virtual_arity0_varcall | 1131.6 | 5033.1 | ~4.4x band |
| container_arg_varcall | 3647.0 | 15811.1 | ~4.3x band |
| copy_ret_ptrcall | 2593.5 | 11810.1 | ~4.6x band |
| copy_ret_varcall | 2831.2 | 11829.5 | ~4.2x band |
| reflective_value_arity0_ptrcall | 11464.7 | 7779.1 | below cool-machine baseline despite load |
| reflective_value_arity0_varcall | 15637.5 | 21222.4 | ~1.4x, under load inflation |
| reflective_value_arity2_ptrcall | 17451.4 | 18663.5 | flat |
| reflective_value_arity2_varcall | 91908.3 | 31873.7 | −65% (group 2 lazy payloads) |
| reflective_value_arity5_ptrcall | 24041.7 | 25896.8 | flat |
| reflective_value_arity5_varcall | 157166.8 | 54007.1 | −66% (group 2 lazy payloads) |
| prologue_notification | 29362.4 | 52518.1 | floor-normalized ~17k: ~42% under baseline |
| prologue_property_get | 48355.2 | 90663.7 | floor-normalized ~29k: ~40% under baseline |

Sources, per row family: allocations from the `dispatch_budget:` records in
`test-output.log` of the final `GODOT=… make test` run (gate:
`check_test_output.sh` + budgets in `test/pkg/dispatch_alloc_budget_tests.go`);
timing from the same run's `dispatch_time:` records
(`test/pkg/dispatch_time_tests.go`); the engine-free prologue resolve bench
(`BenchmarkDispatchPrologueNotification`/`PropGet`, 0 allocs/op, 243/340
ns/op hot vs 91/128 cool) and the two switched nil-check sites
(`object_return_bench_test.go`, `object_arg_bench_test.go`) carry their own
same-run A/B controls in the task 3.3 entry.

Honest summary: allocations moved exactly where the group 2/3 changes
intended (reflective dispatch −67..−85%, compiled paths untouched at zero,
new nil-check and identity paths allocation-free); the wall-clock rows are
presence-recorded and load-dominated — the two figures the change itself
targets (the reflective fallback varcalls, −65%/−66%, and the prologues,
~40% under baseline after floor normalization) are the ones that hold up
under the hot machine, and no row shows a regression the change owns.
`docs/overview.md` carries no numeric cost claim contradicted by these
figures; its statement that the name-string `ClassDBBindMethod` family
"stays reflective" remains true (change B will address it).

## Tasks 4.2–4.3 — freeing state and instance-held class name

`WrappedClassInstance` gained `MarkFreeing()`/`IsFreeing()` (a plain bool;
the instance-free callback and the class prologues are all engine callbacks
on the main thread) and a `ClassName string` field. `FreeInstance` calls
`MarkFreeing()` immediately after the re-entrancy guard deletes the
`GDClassInstances` entry — before the deferred `Destroy()` begins any
Go-side teardown — so the mark spans exactly the window a teardown delivery
can arrive in. The mark is idempotent; the re-entrant second free still
panics through the guard unchanged (class-instance-lifecycle contract).
`ClassName` is populated at all five construction sites: `SetConstructInfo`
and `WrappedPostInitialize` from the `extensionClassName` parameter (the
same registered name the handle is stored under), and the three probes from
the Go-side `GetClassName()`. Compile-time coverage is by construction — the
struct literal at every site names the field; a missing value would be a
visible zero, and the probe/bench witnesses exercise all shapes.

No program-lifetime retention: the string is a Go value inside the
per-instance wrapper, reachable only through the engine-held
`cgo.Handle`, released with it — no map or package variable keyed by
instance is touched. The instance-churn bands (`test_instance_churn` phase
gates) pass unchanged in-engine with the wrapper growth, and the
engine-free dispatch-budget gate (`make test_dispatch_budget`) stays at
zero allocations per call.

Unit witnesses: `TestWrappedClassInstanceMarkFreeingTransitions` (fresh
wrapper not freeing; mark observable; second mark a no-op) and
`TestWrappedClassInstanceIsFreeingNilReceiver` in
`pkg/builtin/wrapped_nil_ptr_test.go`.

## Tasks 4.4–4.5 — prologues read the instance identity; teardown declines on the freeing guard

All seven class callback prologues in `pkg/core/classdb_callback.go`
(`GetPropertyList`, `PropertyCanRevert`, `PropertyGetRevert`,
`ValidateProperty`, `Notification`, `Get`, `Set`) replaced the
`GetClass()`/`ToUtf8()`/`defer Destroy()` engine ceremony with
`className := wci.ClassName`, and each now opens with an `IsFreeing()`
decline (absorb for void deliveries, negative answer for queries) before
any registration lookup. `gdStrClass` no longer appears in the file.
The stale "reports its parent class name during teardown" rationale on the
lookup-miss branches was updated: that shape is now caught by the freeing
guard, and the miss branch remains as the unregistered-class defense — the
loud half (`Get`/`Set`/`ValidateProperty`/`GetPropertyList` panic
"invalid registered GDClass") is unchanged, the quiet half
(`Notification`/reverts) unchanged too. Two mislabeled diagnostics in the
`Set` prologue (`zap.String("class", name)` against the property name) were
fixed to log `className` while touching those branches.

Fixture: `test/pkg/teardown_tolerance_tests.go` extends the 4.1 pins with
the decline and loud-fail cells on the new basis, driving the real
callbacks through the extended probes (`*MarkedFreeing` variants build a
transient wrapper with `MarkFreeing()` set — the probe's private wrapper
only; the engine's binding for the same object keeps reporting not-freeing,
and `ValidateProperty` is exercised on the decline path only, since an
accepted validation would dereference probe-zero property info inside the
class hook). Cells and their meaning:

- Live control (registered `Example`): notification recorded, both revert
  queries answer, get reads `property_from_list`, set mutates it.
- Freeing-marked `Example`: notification not recorded, revert queries
  decline, get produces no value, set declined and the Go field untouched,
  validate and property-list decline/empty — all with a capturing writer
  at the permitting threshold showing zero log bytes.
- Loud half: get, set, validate and property-list for the unregistered
  presentation each raise through the fatal channel with the existing
  "invalid registered GDClass" diagnostic (observed via recover), so the
  fatal branch stays reachable on the new basis.

In-engine run: exit 0, FAILURES: 0, `make test_gate` 27/27. Printed
records:

```
teardown_tolerance: absorbed pre-delete delivery for *gdclassimpl.LabelImpl, no log
teardown_tolerance: revert queries declined for the unregistered presentation
teardown_tolerance: live path still dispatches get and set
teardown_tolerance: freeing-marked wrapper declines every delivery class, silently
teardown_tolerance: unregistered get, set, validate, and property-list requests still fail loudly
```

Prologue cost, live-path cells (recorded, not gated; the machine ran hot
at load-avg 6–10, so read them against same-run controls, not against the
group-1 absolutes):

| shape | group 1 | this run | same-run engine-crossing controls |
|---|---|---|---|
| prologue_notification | 29362.4 | 52518.1 | `static_add_ptrcall` 35.1→107.0, `value_arity0_ptrcall` 29.8→92.5 (~3.1x floor inflation) |
| prologue_property_get | 48355.2 | 90663.7 | `virtual_arity0_ptrcall` 29.8→135.1 |

Normalizing by the ~3.1x floor inflation the compiled ptrcall cells show,
the prologue cells land at ~17k/~29k ns — roughly 40–45% under their
group-1 figures. The engine-free prologue benches
(`BenchmarkDispatchPrologueNotification`/`PropGet`) were switched to read
the stored `ClassName` exactly as the prologues now do; they measure the
shared resolve floor (243/340 ns/op, 0 allocs, hot machine vs the group-1
91–128 ns band) and confirm the new basis itself is allocation-free.

## Task 4.1 — teardown-tolerance pins (pre-change basis)

Fixture `test/pkg/teardown_tolerance_tests.go`, driven from the demo
`main.gd` right after the diagnostics fixture. The engine presents the
teardown shape (instance reporting a parent class name) only during
shutdown, after the extension class leaves the class DB, so the fixture
cannot wait for it; the new `pkg/core/classdb_callback_probe.go` delivers
the identical callbacks through a transient `cgo.Handle` and the prologues
run unchanged. The unregistered presentation is a plain engine `Label`.

Three cells, all at the permitting (info) threshold with a capturing
writer, so silence is measured where entries would be written:

- Absorbed: pre-delete-shaped notification for the unregistered
  presentation returns with no panic and no log entry.
- Declined: `property_can_revert` answers false and `property_get_revert`
  produces no value, silently.
- Live control: the identical deliveries to the registered `Example` run
  the Go virtuals — the notification code is recorded (then restored, as
  the timing fixture does) and both revert queries answer for
  `property_from_list` — so cells 1–2 are decisions, not dead plumbing.

In-engine run: exit 0, FAILURES: 0. Printed record:

```
teardown_tolerance: absorbed pre-delete delivery for *gdclassimpl.LabelImpl, no log
teardown_tolerance: revert queries declined for the unregistered presentation
teardown_tolerance: live path still dispatches; absorption and decline are the unregistered branch alone
```

## Task 3.1 — codegen nil check

`cmd/generate/gdclassimpl/classes.go.tmpl` emits, per generated class,
`func (cx *XImpl) IsNil() bool { return cx == nil }` on the outer pointer
type. `make generate` produced exactly +8288 lines (1036 classes × 8), all in
`pkg/gdclassimpl/classes.gen.go`; the diff was machine-verified to contain
only the new method (zero added lines outside the doc comment + method
shape). `go build ./...` and `gofmt -l` clean (`cmd/main.go` is gofmt-dirty
on HEAD, untouched here).

Declaring on the outer type is load-bearing, verified experimentally: a
method declared on `*Base` satisfies an interface requirement for
`*Outer`/`*Middle` by promotion at compile time, but calling it on a typed
nil SIGSEGVs — Go addresses the embedded field while converting the receiver,
before any nil-receiver body runs. So every concrete outer type needs its own
declaration, which is exactly what the generator now emits (and the rule bit
the hand-written `fakeLabelImpl` in task 3.2).

Semantics: the generated form tests only the wrapper pointer (design D4),
shadowing `WrappedImpl.IsNil`'s Owner-aware form. Both call sites agree
either way: with a nil Owner `AsGDExtensionObjectPtr` already yields a null
pointer, and the object-return branch is about the wrapper's absence.

## Task 3.2 — contract switch off reflect.ValueOf

`Wrapped` gained `IsNil() bool` (documented: hand-written wrappers must
declare it themselves). `builtin.ObjectArgPtr` and
`gdclassimpl.isNilObjectWrapper` replaced `reflect.ValueOf(obj)` kind/nil
tests with the `obj.IsNil()` interface call; the `reflect` import is gone
from both files.

Hand-written implementers found by the compile-time assertions and updated in
the same commit: `typedProbe` and `fakeRefCounted` (pkg/builtin witnesses —
the latter embeds the `RefCounted` interface, where a promoted call would
deref the nil struct), `spotStubClass`, `TypedBindStub`, `OtherBindStub`
(pkg/core), the five `testdata/typedbind*` fixtures, `fakeObject` **and**
`fakeLabelImpl` (pkg/gdclassimpl — `fakeLabelImpl` compiled clean via
promotion from `fakeObject` and then SIGSEGV'd in the autogenerated wrapper,
the live confirmation of the 3.1 promotion rule), and the six `test/pkg`
user wrappers (`Example`, `ExampleRef`, `TestHierarchicalBase`,
`TestHierarchicalDerived`, `TestDelegationRepro`, `ExampleEditorPlugin`).

Verification: engine-free `go test ./pkg/gdclassimpl ./pkg/core ./pkg/log`
green; nil-encoding witnesses (`TestObjectArgPtrTypedNil`,
`TestObjectArgPtrLiveObject`, `TestRefArgPtrNilRef`, the `WrappedImpl`/
`WrappedClassInstance` nil-receiver cases) green. In-engine
`GODOT=… make test`: exit 0, FAILURES: 0, `make test_gate` 27/27, and the
object-argument fixture intact: `object_arg_boundary: 62 object/Ref-argument
binds, all reflective; fallback dispatch resolved instance …` — typed-nil
arguments still encode as null pointers.

## Task 3.3 — measured delta at the two switched sites

The pre-switch reflect form is kept as a same-run A/B control
(`pkg/builtin/object_arg_bench_test.go`,
`pkg/gdclassimpl/object_return_bench_test.go`); engine-free, allocation-free,
no benchmark iteration crosses into the engine. Figures are the mean of 5
reps at benchtime=300ms on a loaded machine (load-avg 6–10), which inflates
absolute values against the historical ~6.7ns/0.47ns pair (the bare-accessor
floor here is 5.68ns, ~12× the documented 0.47ns) — so the same-run deltas,
not the absolutes, are the evidence:

| site | shape | reflect (control) | IsNil | Δ |
|---|---|---|---|---|
| isNilObjectWrapper | typed nil | 15.26 ns | 5.76 ns | −62% |
| isNilObjectWrapper | live | 18.05 ns | 5.84 ns | −68% |
| ObjectArgPtr | typed nil | 16.20 ns | 14.98 ns | −1.2 ns |
| ObjectArgPtr | live | 23.24 ns | 26.00 ns | **+2.8 ns** |
| ObjectArgPtr | bare accessor (floor) | — | 5.68 ns | reference |

Honest reading: the object-return site drops to essentially accessor cost on
both shapes (~3× faster). At `ObjectArgPtr` the typed-nil path improves
slightly while the live path costs a stable +2.8 ns — the extra interface
call is not free where reflect used to run. Against a real ptrcall argument
encode, which crosses cgo (µs scale), both figures are noise-tier, so the
net is: cheaper where the check fires, negligible regression where it does
not, zero allocations everywhere, and one less `reflect` user on the hot
path. In-engine timing fixture unaffected (recorded, not gated): compiled
ptrcall shapes 80–155 ns/call unchanged; the reflective varcalls in this
engine run read 27.8/40.8 µs at arity 2/5 against group 2's 53.3/56.0 µs —
machine-load variance, in the same band, no object arguments involved.

## Task 3.4 — dead reflective decode surface removed

`ArgumentEncoder` lost `DecodeReflectTypePtr`/`DecodeReflectVariantPtr`;
`argumentEncoder` and `objectArgumentEncoder` lost the two closure fields and
the four method wrappers; the six creators (`variant_{bool,builtinclass,
number,object,string,variant}_encoder.go`) lost the closures that built them
and their struct-literal entries. Nothing ever called the decode half: the
interface's only consumer is the `GDRegisteredGDClassEncoders` registration
map in `pkg/gdclassinit`, and the dispatch paths use the `EncodeReflect*` and
plain `Decode*` surfaces. `grep -rn` for all four names across `pkg/ test/
cmd/` returns 0 matches.

Verification: `go build ./...`, `go vet`, `gofmt -l` clean; engine-free
suites green; in-engine `GODOT=… make test` FAILURES: 0 with 22/22 timing
records, `make test_gate` 27/27.

## Task 2.1 — lazy dispatch payloads

`pkg/core/dispatch_payload.go` adds `zapcore.ObjectMarshaler`/
`ArrayMarshaler`/`Stringer` payloads — `GoMethodMetadata.MarshalLogObject`
(class, method, paired `go_type/VARIANT_ENUM` argument types, return type),
`variantCollection` (per-element `Stringify()` at encode time),
`resolvedCollection` (cell-equality dedup → `resolution=same as gd_args`),
and `lazyText`. Eight eager sites rewired: four in `method_bind.go`
("Call Variadic", "Calling", "Call", "Ptrcall"), two "reflect method called"
in `classdb_callback.go` (the `_get` twin also moved its `v.ToString()`
engine crossing into the payload), two `spew.Sdump("argument converted")` in
`method_bind_reflect.go`. `pkg/log` gained a `swapWriter` indirection making
`SetWriteSyncer` actually live (the core captured stdout at init; it now
returns the previous writer).

Unit witnesses (`pkg/core/dispatch_payload_test.go`, engine-free):
`TestLazyPayloadsNotMarshaledWhileSuppressed` — capturing syncer, zero
marshaler/Stringer invocations while suppressed, exactly one each after
raising; `TestPtrcallSiteSuppressedWritesNothingRaisedRestoresContent` — the
real dispatch site writes nothing at warn and the same field names and
identification content at info; plus bind-structure (JSON), nil-metadata,
empty-collection, and dedup checks. Engine-free reach caveat: a non-empty
`variantCollection` marshal calls `Variant.Stringify`, which SIGSEGVs without
the engine table — so non-empty content assertions live in the in-engine
diag fixture, and that fault-is-the-proof split means a regression
reintroducing eager payloads faults the engine-free tier loudly.

## Task 2.2 — variadic duplicate removed

The variadic site's `resolved_args` field deleted outright (the handler
receives the collected Variants unchanged, so it was always a second
description of `gd_args`). Captured entry at the permitting threshold, from
the in-engine run (fixture `TestDispatchDiagnostics`, asserts exactly one
`"values"` array, both values present, no `resolved_args`):

```
Call Variadic {"bind": {"class": "Example", "method": "VarargsFunc",
  "argument_types": ["/GDEXTENSION_VARIANT_TYPE_VARIANT_MAX"],
  "return_type": "Variant"}, "gd_args": {"values": ["2", "3"]},
  "ret": "(Variant/array(<builtin.Variant Value>))"}
```

Before (red run): `"gd_args": "[]Variant(2, 3)", "resolved_args": "[]Variant(2, 3)"`.

## Task 2.3 — stringify crossings for discarded calls

Fixture `test/pkg/dispatch_diag_tests.go` swaps the loaded interface table's
`variant_stringify` for a `//export`ed Go counter (static C helper returns
its address; the swap lives only across the fixture's own dispatches and
defer-restores). Thresholds forced via `log.SetLevel`, never inherited.
Reflective fallback varcalls at arity 0/2/5:

| threshold | arity 0 | arity 2 | arity 5 |
|---|---|---|---|
| suppress — before | 0 | 8 | 20 |
| suppress — after | 0 | 0 | 0 |
| raised — before | 0 | 8 | 20 |
| raised — after | 0 | 2 | 5 |

Before-figures grew with arity (eager signature: `gd_args` + `resolved_args`
stringified per call, ×2 because the discarded Debug "Calling" entry also
paid eagerly). After: flat zero while suppressed; raised costs exactly one
stringify per argument (the dedup means `resolved_args` crosses only when
the fill actually changed cells). `GODOT=… make test` green with the whole
matrix and both captured entries printed.

## Task 2.4 — budgets re-baselined after laziness

- Engine-free gating (`go test ./pkg/core/ -run Budget`): 30 compiled-shape
  rows unchanged at zero; reflective control at arity 1 measured **9**
  allocations/call (was 32 — ~23 were the eager payload strings), so the
  vacuity floor moved 16 → 4 (half-margin philosophy kept, still ≫ the
  compiled zero). Green.
- In-engine budget fixture reflective twins now measure **8/10/13 ptrcall**
  and **16/25/37 varcall** allocs/call at arity 0/2/5 (was 27-56 / 38-249);
  floor 4 + margin ×2 checks intact, all pass. The checks are floor +
  relative, so nothing to tighten upward; the fixture header comment updated
  to the new figures with the before-figures noted.
- Side-effect free performance (timing fixture, recorded only): reflective
  varcall arity2 91908 → 53330 ns/call (−42%), arity5 157167 → 56000 ns/call
  (−64%) at the default warn threshold — the eager payloads were on the
  fallback hot path. arity0 unchanged (15638 → 16245, noise).

## Task 2.5 — no gate depends on a suppressed dispatch line

`GODOT=… make test` at the default (warn) threshold: exit 0, FAILURES: 0,
gate PASS line green. `LOG_LEVEL=info GODOT=… make test`: exit 0, FAILURES:
0 — the diag fixture's self-forced thresholds keep its matrix identical
under the raised env (suppress section still flat zero). `make test_gate`:
27/27 fixture cases. Which checks read dispatch diagnostics: **none** —
`check_test_output.sh` reads driver banners, probe records
(`stacktrace-probe`, `callable-panic-probe`, `dispatch_time:`,
`dispatch_budget:`) and engine error lines; the `Call`/`Call Variadic`/
`Ptrcall` entries were grep-verified as having no consumers before rewiring,
and flooding them at info breaks no check.

## Task 1.6 — baseline consolidated + discoverability

Consolidated pre-change baseline table added at the top of this file.
AGENTS.md gained a "Dispatch Cost Baseline" subsection under Testing (budget
gate, bench target, in-engine records, and the shape-list coupling). Makefile
gained `test_dispatch_bench` (`go test ./pkg/core/ -bench 'Dispatch'
-benchmem -run 'XXXnoTests'`); `make help` now lists both dispatch targets:

```
test_dispatch_bench   Record engine-free dispatch timings (ns/op; recorded, never gated)
test_dispatch_budget  Gate per-call allocation budgets for every dispatch shape (no Godot needed)
```

`make test_dispatch_bench` verified to print the matrix (prologue cells
92.3/115.1 ns/op, seam cells 16.7/20.0 ns/op on the verification run).

## Pre-change baseline (group 1 consolidated)

The figures every later group diffs against. Per-task detail follows in the
1.1–1.5 entries; recorded on `feat-trim-dispatch-overhead` at commit
`b463cdf`, linux/amd64, Intel Core i7-8650U @ 1.90GHz, Godot 4.8.beta.

| tier | figure | gate? |
|---|---|---|
| engine-free allocations, 30 compiled/budget shapes | 0 per call | gated (budget = 0) |
| engine-free allocations, reflective control (arity 1) | 32 per call | gated (floor 16) |
| engine-free timing | see task 1.1 / 1.2 tables (instance ~3ns, copy-ret ~14-17ns, seam ~18-22ns, varcall seam hand-off ~38ns, prologue resolve 91-128ns) | recorded |
| in-engine allocations (budget fixture) | compiled 0 ptrcall / 3-14 varcall; reflective 27-56 ptrcall / 38-249 varcall | gated (caps + margin) |
| in-engine timing, 22 shapes | see task 1.4 table | recorded (presence gated) |

Headline in-engine numbers for the before/after table (task 5.x):
compiled ptrcall ~30-48ns/call; compiled varcall ~1.1-5.4µs/call; reflective
ptrcall ~11-24µs/call; prologue_notification 29362.4ns/call;
prologue_property_get 48355.2ns/call.

## Task 1.5 — output gate requires the timing records

`test/check_test_output.sh` gained check 12 (`dispatch-timing-records-missing`):
the 22 shape names (list coupled to the `dispatch_time_tests.go` header) must
each appear as `dispatch_time: shape=<name> ns/call=`; values are never read.
The failure message names every missing shape. The old check 12
(godot-exited-non-zero) renumbered to 13; the PASS line gained
"dispatch timing records present".

`test/fixtures/run_fixture_tests.sh` coverage, per the task's three cases:

- present: `clean_pass.log` and `stacktrace_stub.log` extended with the 22
  records (still exit 0);
- missing-shape: new `dispatch_time_partial.log` (20 of 22 records; fails the
  mode) plus a targeted assertion that the message names both dropped shapes
  (`prologue_notification`, `copy_ret_varcall`);
- absent-log: already covered by the existing `does_not_exist.log` loop.
- bonus `dispatch_time_absent.log`: whole section vanished (driver drift) —
  same mode, all 22 named.

`godot_nonzero.log` (pre-existing, clean-driver-with-nonzero-exit) predates
timing and would have tripped check 12 before reaching its intended 13 — it
now carries the records too, matching what a current run looks like.

`make test_gate`: 27 passed, 0 failed. The gate re-run against the live
`test-output.log` passes with the extended PASS line (1490 assertions).

## Task 1.4 — in-engine timing fixture

Added `test/pkg/dispatch_time_tests.go` (`TestDispatchTimes`), bound via
`ClassDBBindMethod(t, "TestDispatchTimes", "test_dispatch_times", nil, nil)`
in `test/pkg/example.go`, driven from `test/demo/main.gd`
(`test_dispatch_times(example)`). Fixture verifies `Compiled()` on the seven
typed probes and `!Compiled()` on the three reflective twins before timing
(a silent fallback would mislabel records), reuses the budget fixture's
registered probe binds, and prints 22 machine-readable records.
`GODOT=… make test` run on this machine (i7-8650U, Godot 4.8.beta): all 22
`dispatch_time: shape=… ns/call=…` lines present, `FAILURES: 0`, gate green.

Pre-change in-engine baseline (ns/call, recorded — never gated):

| shape | ns/call |
|---|---|
| static_add_ptrcall | 35.1 |
| static_add_varcall | 3235.6 |
| value_arity0_ptrcall | 29.8 |
| value_arity0_varcall | 1136.5 |
| value_arity2_ptrcall | 34.4 |
| value_arity2_varcall | 3078.5 |
| value_arity5_ptrcall | 47.7 |
| value_arity5_varcall | 5408.4 |
| virtual_arity0_ptrcall | 29.8 |
| virtual_arity0_varcall | 1131.6 |
| container_arg_ptrcall | 31.8 |
| container_arg_varcall | 3647.0 |
| copy_ret_ptrcall | 2593.5 |
| copy_ret_varcall | 2831.2 |
| reflective_value_arity0_ptrcall | 11464.7 |
| reflective_value_arity0_varcall | 15637.5 |
| reflective_value_arity2_ptrcall | 17451.4 |
| reflective_value_arity2_varcall | 91908.3 |
| reflective_value_arity5_ptrcall | 24041.7 |
| reflective_value_arity5_varcall | 157166.8 |
| prologue_notification | 29362.4 |
| prologue_property_get | 48355.2 |

Reading for later tasks:

- Compiled ptrcall in-engine is ~30-48ns — the Go-side closure (~3ns) plus
  the test-side seam and measurement floor. Reflective ptrcall is 11-24µs
  (≈350-500x), consistent with the alloc-count gap (27-56 vs 0).
- Varcall figures include per-call Variant decode/encode crossings and the
  `ret.Destroy()` the probe owes the engine (same convention as the budget
  fixture's leak gates).
- The two prologue cells ride live engine paths (Object notification,
  property read through `_get`), so they include the engine round trip;
  that constant makes them the before/after pair for the task 4.4 hoist —
  notification ~29µs, property-get ~48µs pre-change.
- `e.notificationCodes` is restored after the notification cell; downstream
  notification assertions stayed green (FAILURES: 0).

## Task 1.3 — gating budget tests over the benchmark shape set

Added `pkg/core/dispatch_budget_gate_test.go`: 30 zero-budget rows covering
every benchmark shape (instance/static value-void-copy-ret, six callback
closures, prologue cells, four seam cells) plus the reflective control with a
floor of 16 (baseline 32). Each row reports `measured / budget / handler_calls`
every run and carries a witness-based vacuity guard (handler must advance at
least once per measured call). Wired into `make test` as the
`test_dispatch_budget` target (`go test ./pkg/core/ -run Budget -count=1`,
prerequisite of `test:`).

Falsification runs (each temporary, reverted immediately after capture):

1. Sabotage — handler allocates once per call
   (`gateSabotageSink = make([]byte, 8)` in the arity-0 value handler):

   ```
   dispatch budget instance value-ret arity-0  measured=1 budget=0 handler_calls=201
   instance value-ret arity-0 allocates 1 per call, budget 0
   --- FAIL: TestDispatchBudgetGateInstancePtrcall
   ```

   names the shape, the measured figure, and the budget, as required.

2. Sabotage — shape stops dispatching (handler witness increment removed):

   ```
   dispatch budget instance value-ret arity-0  measured=0 budget=0 handler_calls=0
   instance value-ret arity-0: vacuous budget — handler ran 0 times across
   200 measured calls; the shape stopped dispatching, so the budget cannot
   observe it
   --- FAIL: TestDispatchBudgetGateInstancePtrcall
   ```

3. After revert: `make test_dispatch_budget` green and
   `go test ./pkg/core/ -count=1` green; all 30 rows measured=0, control
   measured=32 against floor 16, every row handler_calls=201.

## Task 1.2 — prologue resolve-and-dispatch cells (engine-free reach)

Added `BenchmarkDispatchPrologueNotification` / `BenchmarkDispatchProloguePropGet`
to the same file (30 shapes total). Same machine/command family as above:

| shape | ns/op | allocs/op |
|---|---|---|
| prologue notification (resolve-and-dispatch) | 91.5 | 0 |
| prologue property-get (resolve-and-dispatch) | 127.8 | 0 |

Attribution probe (throwaway, run then deleted):

| component | ns/op |
|---|---|
| cgo handle Value + type assert + GetClassName | 30.5 |
| `GDRegisteredGDClasses.Get("SpotStub")` (SyncMap, string hash) | 36.2 |
| VirtualMethodMap lookup + metadata read + compiled cb dispatch | 31.2 |

Engine-free reach — what these cells cannot measure: the current prologues
read the class name via `wci.Instance.GetClass()` (engine call returning a
Godot `String`) + `ToUtf8()` (allocation + UTF-16 copy) + `Destroy()` (engine
call), none of which can execute without a live engine. The cells therefore
model name acquisition with the Go-side `GetClassName()` — the instance-held
identity source D1 provides — and measure the shared floor of before/after.
The engine ceremony those prologues pay today (the actual target of the task
4.4 hoist, plus the string-keyed hash of an engine-allocated name) is
recorded in-engine by the task 1.4 fixture.

Reading for later tasks: ~67ns of the reachable prologue is handle + registry
hash ceremony; the compiled callback underneath is ~3ns (closure cell). D1
keeps the registry lookup, so the engine-free tier expects a small drop
(virtual call → field read); the large saving is engine-side and shows in
task 1.4.

## Task 1.1 — engine-free benchmark set (pre-change baseline)

File: `pkg/core/dispatch_cost_bench_test.go` (28 shapes). Command:
`GOTMPDIR=… go test ./pkg/core/ -bench 'Dispatch' -benchmem -run 'XXXnoTests'`
on `linux/amd64`, Intel Core i7-8650U @ 1.90GHz. Every shape printed
ns/op and allocs; all allocation columns are zero, matching the compiled-path
contract. `go test ./pkg/core/ -count=1` stays green (benchmarks are not run
by the default target).

| shape | ns/op | allocs/op |
|---|---|---|
| instance value-ret arity 0 | 2.97 | 0 |
| instance value-ret arity 2 | 3.10 | 0 |
| instance value-ret arity 5 | 3.78 | 0 |
| instance void-ret arity 0 | 3.04 | 0 |
| instance void-ret arity 2 | 3.21 | 0 |
| instance void-ret arity 5 | 4.23 | 0 |
| instance copy-ret arity 0 | 14.45 | 0 |
| instance copy-ret arity 1 | 14.81 | 0 |
| instance copy-ret arity 2 | 16.95 | 0 |
| static value-ret arity 0 | 0.57 | 0 |
| static value-ret arity 2 | 1.25 | 0 |
| static value-ret arity 5 | 2.52 | 0 |
| static void-ret arity 0 | 2.34 | 0 |
| static void-ret arity 2 | 2.21 | 0 |
| static void-ret arity 5 | 3.22 | 0 |
| static copy-ret arity 0 | 12.91 | 0 |
| static copy-ret arity 1 | 10.64 | 0 |
| static copy-ret arity 2 | 12.69 | 0 |
| notification closure | 3.11 | 0 |
| prop-get closure | 3.25 | 0 |
| prop-set closure | 3.01 | 0 |
| prop-can-revert closure | 3.00 | 0 |
| prop-get-revert closure | 3.44 | 0 |
| to-string closure | 2.86 | 0 |
| seam value-ret arity 2 (Ptrcall) | 18.50 | 0 |
| seam void-ret arity 5 (Ptrcall) | 21.75 | 0 |
| seam copy-ret arity 2 (Ptrcall) | 18.21 | 0 |
| seam varcall value-ret arity 0 (CallWithError) | 38.25 | 0 |

Notes for later comparisons:

- The varcall seam cell times the `CallWithError` selection and hand-off with a
  shape-identical stand-in closure; generated varcall execution crosses the
  engine table on its return encode and is recorded in-engine (task 1.4).
- copy-ret cells carry the always-destroy path (`spotRetStub` echo compare is
  nil by contract); the echo-suppressed path needs the engine copy ctor.
- The seam cells land ~15ns above the bare closures — first candidate reading
  for where metadata-seam ceremony goes; task 1.2 adds the prologue cells where
  the engine-lookup ceremony lives.
