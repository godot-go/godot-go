# Tasks

## 1. Cost baseline before any behavior change

- [x] 1.1 Add engine-free benchmarks in `pkg/core` over the dispatch shapes the previous change enumerated (instance/static value-ret and void-ret arity 0/2/5, copy-ret arity 0/1/2, the six callback closure shapes, and the `GoMethodMetadata.Ptrcall`/varcall seam with `Func` zero); verify `go test ./pkg/core/ -bench 'Dispatch' -benchmem` prints ns/op and allocs for every shape
- [x] 1.2 Add the callback-prologue measurement path engine-free (resolve-and-dispatch a notification and a property-get against a registered stub class), documenting in the task evidence which parts of the prologue the engine-free tier cannot reach (the engine `GetClass()` call itself)
- [x] 1.3 Turn the benchmark set into gating allocation budgets with floor-plus-margin assertions and a per-shape vacuity guard, in the style of `pkg/core/typed_dispatch_budget_test.go`; verify by sabotaging one compiled shape to allocate, confirming that shape's budget test fails naming shape, measured figure and budget, then reverting and confirming green
- [x] 1.4 Add the in-engine timing fixture in `test/pkg` printing one `dispatch_time: shape=… ns/call=…` line per covered shape (bind it in `test/pkg/example.go`, drive it from `test/demo/main.gd`); verify `GODOT=… make test` shows one line per shape in `test-output.log`
- [x] 1.5 Teach `test/check_test_output.sh` to require the timing records (presence per shape, never values) with matching cases in the gate's own fixture suite (`test/fixtures/run_fixture_tests.sh`); verify `make test_gate` covers the present, the missing-shape, and the absent-log cases
- [x] 1.6 Record the pre-change baseline figures (per-shape allocations and in-engine ns/call) in this change's task evidence, and document the new benchmark and budget invocations in the AGENTS.md testing section and `make help`; verify the figures are committed and the help text lists the targets

## 2. Diagnostics that cost nothing when discarded

- [x] 2.1 Implement lazy marshaling for the dispatch payloads: an `ObjectMarshaler` for `GoMethodMetadata` (class, method, argument types, variant types) and for the Variant collection payloads, passed as `zap.Object(...)` at the three dispatch sites in `pkg/core/method_bind.go`; verify a unit test with a capturing write syncer asserts nothing is marshaled while the threshold suppresses the entry, and that raising the threshold emits the same field names and content as before
- [x] 2.2 Remove the duplicated argument-collection description at the variadic dispatch site (`gdArgs` stringified twice) and route the remaining payload through the lazy form; verify a fixture at the permitting threshold shows each argument described exactly once for a two-argument call
- [x] 2.3 Prove the engine is not asked to stringify a discarded call's arguments: measure `variant_stringify` crossings for a fallback varcall at the suppressing threshold before and after (an allocation-and-crossing counter, not a log grep); verify the after figure stops growing with argument count
- [x] 2.4 Re-baseline the fallback allocation budgets downward now that the payloads are lazy, keeping the compiled-shape budgets at zero and the reflective-twin floor-plus-margin checks intact; verify `go test ./pkg/core/ -run 'Budget'` is green and the new figures land in the task evidence beside the 2.x baseline from group 1
- [x] 2.5 Confirm no gate depends on a dispatch line it suppresses: run `GODOT=… make test` and `make test_gate` at the default threshold and once with the threshold raised; verify both runs pass and record which checks read dispatch diagnostics in the task evidence

## 3. Nil detection and dead encoder surface

- [x] 3.1 Emit a nil check on each generated wrapper type from the `cmd/generate` template (declared on the outer pointer type, never promoted through the embedded value), regenerate with `make generate`, and review the generated diff for shape; verify `go build ./...` and `gofmt -l` are clean and the generated files differ only by the new method
- [x] 3.2 Add the nil check to the wrapper contract and switch `pkg/builtin/object_arg.go:30` and `pkg/gdclassimpl/object_return.go:78` from `reflect.ValueOf` to it; update every hand-written implementer found by the compile-time assertions in the same commit; verify the existing typed-nil encoding fixtures still encode a null pointer and the object-argument suites are green
- [x] 3.3 Re-measure the two switched call sites against the group 1 baseline and record the delta in the task evidence (the previous figures were ~6.7ns against ~0.47ns for the bare accessor); verify the timing fixture or an engine-free benchmark shows the improvement
- [x] 3.4 Remove the unused `DecodeReflectTypePtr` / `DecodeReflectVariantPtr` methods from the encoder interface and every per-type encoder; verify `grep -rn` finds no remaining callers, `go build ./...` succeeds, and the suite is green

## 4. Instance-held class identity and the teardown basis

- [x] 4.1 Before changing resolution, pin today's teardown behavior with fixtures: reproduce the pre-delete notification delivery and the destructing-instance revert query, asserting silent absorption and a negative answer respectively, and register them in the runtime suite; verify both pass against the current tree
- [x] 4.2 Add an explicit "wrapper is being freed" state to the instance wrapper, set by `FreeInstance` at the point the existing re-entrancy guard fires, and expose a read accessor; verify a unit test observes the transition and that a second free still no-ops
- [x] 4.3 Store the registered class name on `WrappedClassInstance` at construction and populate it at every construction site; verify compile-time coverage of the sites, that the instance-construction retention requirement still holds (no program-lifetime retention), and that the instance allocation budget from group 1 is unchanged
- [x] 4.4 Switch the seven class callback prologues that still cross into the engine to learn a class name — `GetPropertyList`, `PropertyCanRevert`, `PropertyGetRevert`, `ValidateProperty`, `Notification`, `Get`, `Set` in `pkg/core/classdb_callback.go` — to the instance-held identity plus the freeing-state decline, removing the `GetClass()`/`ToUtf8()`/`Destroy()` ceremony and the string-keyed lookup from each (`resolveToString` already answers from the Go-side class name, and is the pattern to follow); verify the 4.1 fixtures pass on the new basis and the notification, property-access, and property-revert suites are green
- [x] 4.5 Keep the loud-failure half reachable: verify with a fixture that a property get, set, or validation request whose class has no registration still fails through the fatal channel with the existing diagnostic, and that the timing fixture shows the prologue cost drop on the live path

## 5. Integration gate

- [x] 5.1 Run `go test ./pkg/...` (excluding the pre-existing engine-free `pkg/builtin` fault already documented by the previous change), then `GOTMPDIR=… GODOT=… make test` and `make test_gate`; verify zero failures, both gates green, zero `Orphan StringName` lines in `test-output.log` and `editor-output.log`, and the `instance_churn` counters unchanged
- [x] 5.2 Record the before/after cost table (allocations per shape from the budgets, in-engine ns/call from the timing fixture) in the task evidence, and correct any cost claim in `docs/overview.md` that the new numbers contradict; verify the table is committed and each row names its measurement source

## Workflow follow-up

- Get the change reviewed under the project's normal requirements (no push or PR is implied by these tasks).
- Archive with `openspec-archive-change`: the new `dispatch-cost-baseline` capability and the four modified capabilities (`diagnostic-output-channel`, `notification-virtual-dispatch`, `property-revert-virtuals`, `property-access-dispatch`) sync to main specs as part of the archive.
- The follow-up change that compiles the name-string bind path at bind time starts from this change's recorded baseline, and revises `typed-method-dispatch`.
