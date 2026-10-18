# Tasks

## 1. Static metadata core

- [x] 1.1 Add `IsStatic()` accessor on `GoMethodMetadata` (derived from
      `MethodFlags & METHOD_FLAG_STATIC`) and a receiver-free metadata
      constructor taking a package `func` `reflect.Value` with argument
      index base 0, validating `NumIn()` against `len(argNames)`; verify
      with `go build ./pkg/core` plus unit tests constructing metadata
      from a sample function and asserting the arity validation fires.
- [x] 1.2 Skip receiver injection in `CallWithError`, `Ptrcall`, and
      variadic `CallSlice` when `IsStatic()`; verify with unit tests
      dispatching static metadata without a receiver while existing
      instance-dispatch tests stay green.

## 2. Bind API

- [x] 2.1 Implement `ClassDBBindStaticMethod[T GDClass](inst T,
      gdMethodName string, fn any, argNames []string,
      defaultValues []Variant)` and its internal path: class name from
      `inst`, metadata from the static constructor, registration carrying
      `METHOD_FLAG_STATIC`; verify with `go build ./pkg/core` and unit
      tests that a non-callable `fn` fails fatally naming class and
      method, and that the shared internal path rejects static+virtual.

## 3. Callback null-instance branches

- [x] 3.1 In `GoCallback_MethodBindMethodCall` and
      `GoCallback_MethodBindMethodPtrcall`, branch on `bind.IsStatic()`:
      static skips instance resolution (and moves `pinner.Pin(instPtr)`
      into the non-static branch) and logs `bind.ClassName` instead of
      `inst.GetClass()`; non-static keeps the existing nil-fatal; verify
      with `go build ./pkg/core` and `go vet ./pkg/core` clean.

## 4. In-engine harness scenarios

- [x] 4.1 Register static fixtures on the harness `Example` class (typed
      static for ptrcall, dynamic-invoked static, static with default
      arguments, variadic static, static taking an Object, static
      returning an Object) and add GDScript class-path drivers asserting
      delivered values and object identity; verify the scenario verdicts
      come back green in a `make test` run.
- [x] 4.2 Add misuse and introspection scenarios: argument-count mismatch
      on a static reports a call error without aborting, duplicate-name
      static bind fails fatally with the prior registration intact
      (recovered), and `ClassDB.class_has_method` reports the static on
      the class with no instance; verify the scenario verdicts come back
      green in a `make test` run.
- [x] 4.3 Record the engine's actual legality behavior for a static call
      made on an instance as an observed scenario verdict (no assertion
      beyond what the engine does); verify the verdict prints in the
      suite output.

## 5. Documentation

- [x] 5.1 Rewrite the "Static Methods" section of `docs/overview.md` to
      document `ClassDBBindStaticMethod` with a usage example, and
      correct the "Static Variables" section to state it is blocked on a
      GDExtension ABI that does not exist; verify the documented
      signature matches `pkg/core/classdb.go` and the example compiles as
      written against the harness fixtures.
- [x] 5.2 Update `docs/roadmap.md`: mark B2 static methods DONE and B2
      static variables BLOCKED with the ABI reason; verify the roadmap
      section matches the shipped capability and the overview wording.

## 6. Integration checks

- [x] 6.1 Run the full gate `GODOT=/home/pcting/bin/godot make test` and
      confirm `test/check_test_output.sh` passes (extension loaded,
      TESTS FINISHED, failures 0, no ObjectDB leaks, exit 0).
- [x] 6.2 Mutation-prove the null-instance branches bite: temporarily
      revert the static branch in one callback to the unconditional
      nil-panic and confirm the suite fails, restore, repeat for the
      other callback, and confirm the suite is green again; record both
      mutation results in the change notes.
- [x] 6.3 Run revive with `.revive.toml` over the touched packages and
      confirm no new violation categories versus baseline.

## Verification notes

- 6.2 mutation proofs (2026-10-16): reverting the varcall callback's static
  branch to the unconditional nil-panic fails the harness with
  `panic: GDExtensionClassInstancePtr canoot be null` (method_bind_callback.go:49)
  on the first varcall-routed static call; godot exits 134 and the gate fails
  as `FAIL[no-driver-summary]`. The same revert on the ptrcall callback fails
  identically (line 238) on the first typed static call through the ptrcall
  fast path. Restoring both branches returns the suite to green: 1381
  assertions, 0 failures, shutdown marker present, no ObjectDB leaks, no
  orphan StringNames, exit 0.
- Task 4.3 in-engine observation: the engine dispatches a static invoked on an
  instance through the generic `Object.call` channel -- the static functor
  ignores the instance and the call returned 3; the suite output records
  `static-via-instance observation: 3`.
- 6.3 revive baseline comparison: remaining findings in touched files are all
  pre-existing categories (dot imports, `V_`/`GoCallback_` underscore naming,
  missing package/exported comments on legacy symbols). The two new
  unused-parameter instances introduced during implementation (dead
  `argumentNames` in the shared metadata builder after the validation
  extraction; an unused test fixture) were removed; zero new categories.
