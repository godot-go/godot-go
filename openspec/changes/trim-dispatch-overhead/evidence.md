# Task evidence

Running log of measured figures and gate output backing each completed task.

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
