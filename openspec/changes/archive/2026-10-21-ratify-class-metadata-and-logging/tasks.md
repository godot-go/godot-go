# Tasks

## 1. Prove the engine answers the duplicate question

- [x] 1.1 From inside the class-registration path, query `ClassDB.class_has_signal` for a signal the class carries only by inheritance, a name the class owns, and an unused name, logging all three answers at warning level so the gated run shows them, and verify with `GODOT=/path/to/godot make test` that `test-output.log` shows the inherited name reporting true at registration time while the unused name reports false.
- [x] 1.2 Record the outcome of 1.1 against decision D1 in `design.md`, confirming the boolean query or switching the gate to `class_get_signal_list(class, false)`, and verify the recorded outcome matches the observed log lines before any gate code is written.

## 2. Inheritance-aware signal name gating

- [x] 2.1 Extract the duplicate decision into a testable helper used by `ClassDBAddSignal` that reports a collision from the engine answer plus the class's own recorded names, and verify with `go test ./pkg/core` covering own-name collision, inherited-name collision, unused name, and a nil-singleton degradation to the recorded names with a warning that inheritance was not checked.
- [x] 2.2 Route `ClassDBAddSignal` through that helper, failing through the library's fatal path with a diagnostic naming the class and the signal, retaining no name for a rejected registration, and verify with `GODOT=/path/to/godot make test` that the demo's existing custom signal still registers and no assertion regressions appear.
- [x] 2.3 Add a recovered demo probe that attempts to register a signal whose name the class inherits and a second duplicate of a name the class already registered, asserting both fail with a diagnostic naming class and signal and that neither name becomes one the class reports, and verify with `GODOT=/path/to/godot make test` reporting `FAILURES: 0` with the new assertions executed.
- [x] 2.4 Confirm the fatal diagnostic no longer describes a constant when a signal collides, by asserting the caught message text in the 2.3 probe, and verify with `GODOT=/path/to/godot make test`.

## 3. Enum grouping and bitfield-ness that reach the engine

- [x] 3.1 Add GDScript assertions for the current bitfield state — enum presence, enum reported as a bitfield, each flag's owning enum, and unqualified flag values — and verify they fail on the unmodified code for the metadata assertions while the value assertions pass, proving the assertions are wired to the right question.
- [x] 3.2 Forward the caller's enum name from `ClassDBBindBitfieldFlag` to the engine, and update its documentation to state that a `Class.EnumName`-qualified enum name is accepted and shortened by the engine to the portion after its first separator, and verify that the 3.1 assertions all pass under `GODOT=/path/to/godot make test`.
- [x] 3.3 Register a second, distinct flag enum and an enum-less constant on the demo class, and verify with `GODOT=/path/to/godot make test` that both enums are reported independently, each constant reports the expected owning enum or none, and the run reports no new engine errors.
- [x] 3.4 Add a recovered probe binding the same constant name under two different enums, asserting the second binding fails fatally naming the class and that the first binding keeps its value, and verify with `GODOT=/path/to/godot make test` reporting `FAILURES: 0`.

## 4. Property groups published and observable

- [x] 4.1 Add GDScript assertions reading the class's registered property list for a group entry and a subgroup entry with the demo's display names and prefixes, and an ungrouped property that matches no prefix, and verify with `GODOT=/path/to/godot make test` that all four metadata assertions pass.
- [x] 4.2 Add a recovered probe registering a property group for a class that was never registered, asserting the request fails loudly and the engine is not called, and verify with `GODOT=/path/to/godot make test` reporting `FAILURES: 0`.

## 5. Diagnostic surface witnesses

- [x] 5.1 Add package tests that re-exec the test binary in a child mode to observe threshold selection at start — unconfigured yields the default, a configured value admits quieter entries, an unparseable value yields the default with start-up continuing — and verify with `go test ./pkg/log`.
- [x] 5.2 Extend the child mode to emit entries from several goroutines and assert in the parent that every entry arrives whole on its own line on standard output and that suppressed levels are absent, and verify with `go test ./pkg/log`.
- [x] 5.3 Assert in the same harness that a warning-level entry carries a call stack and an admitted debug entry carries none, and verify with `go test ./pkg/log`.
- [x] 5.4 Add a package test that a binding fault records its entry before unwinding, that the entry survives the default threshold, and that it identifies the rejected construct, and verify with `go test ./pkg/log`.
- [x] 5.5 Locate or add an executable witness that a panic inside a user callable closure is recovered, reported as a failed call, recorded as an entry, and leaves the run alive, and verify with `GODOT=/path/to/godot make test` that the assertion executes and `FAILURES: 0`.
- [x] 5.6 Confirm the teardown requirement has a live witness by pointing the existing main-loop shutdown marker assertion at the new requirement in a comment beside the fixture, and verify with `GODOT=/path/to/godot make test` that the marker is still required and present.

## 6. Documentation and boundaries

- [x] 6.1 Add a "Class metadata and diagnostics" section to `docs/overview.md` mapping both new capabilities to their spec paths and the public entry points they govern (`ClassDBAddSignal`, `ClassDBBindConstant`, `ClassDBBindEnumConstant`, `ClassDBBindBitfieldFlag`, `ClassDBAddPropertyGroup`, `ClassDBAddPropertySubgroup`, and the `log` entry points), and verify the two spec sources exist with `ls openspec/changes/ratify-class-metadata-and-logging/specs/*/spec.md`.
- [x] 6.2 Re-read both delta specs against the code and the adjacent capabilities (`class-registration`, `method-call-error-reporting`, `native-stacktrace-capture`, `test-harness-integrity`) and correct any scenario that cannot be observed as written, and verify `openspec validate "ratify-class-metadata-and-logging" --strict` passes with no warnings.

## 7. Integration checks

- [x] 7.1 Run `go build ./...` and `go test ./pkg/core ./pkg/log ./pkg/builtin -skip 'TestRefAcquireUnrefReleasesExactlyOnce|TestRefAcquireFinalizerReleasesWhenUnreachable'` and confirm both are green with the pre-existing engine-dependent exclusions unchanged.
- [x] 7.2 Run `GODOT=/path/to/godot make test` and confirm the runtime suite reports `FAILURES: 0` and the editor-plugin gate reports PASS with no new engine errors in the captured output.

## 8. Review fixes

- [x] 8.1 Fail the gated run when the signal duplicate gate degrades: add a `signal-gate-degraded` check to `test/check_test_output.sh` for the `inheritance was not checked` warning, with a `signal_gate_degraded.log` fixture and table row, and verify the fixture suite reports zero failures.
- [x] 8.2 Condition the callable-panic gate check on the probe's own marker: have the driver print `callable-panic-probe attempted=<n>` after the panicking-callable assertions, gate on it like the stacktrace check, carry the marker into the four full-run fixtures, and verify the fixture suite and `GODOT=/path/to/godot make test`.
- [x] 8.3 Rename `signalNameCollision`'s completeness result to `decisive` so the flag says what it means on the recorded-name path, and verify with `go test ./pkg/core`.
- [x] 8.4 Correct the proposal's gate-query name to `class_has_signal` and record the bitfield enum-name consequence and the gate sentinel in the proposal and design, and verify `openspec validate "ratify-class-metadata-and-logging" --strict`.

## Workflow follow-up

- Review both specs against the adjacent capabilities before archiving; the metadata capability must not absorb method or virtual name collisions, which `class-registration` owns.
- Archive the change; confirm the main spec count moves from 47 to 49 and each new spec carries its Purpose.
- Deferred candidates for separate changes: aligning signal parameter `class_name` with the reference implementation, per-class tracking and de-duplication of property groups, and the codegen contract.
