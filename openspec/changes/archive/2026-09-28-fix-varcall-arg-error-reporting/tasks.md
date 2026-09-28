# Tasks

## 1. Typed decode failure and error-returning entry point

- [x] 1.1 Add `VarcallArgDecodeError` (fields: `Index int`, `ParamType string`, wrapped `Err error`) with `Error()` and `Unwrap()` so `errors.Is`/`errors.As` work through it; verify with a Go unit test that wraps a sentinel error and recovers `Index` and `ParamType` via `errors.As`.
  - Landed in `pkg/core/method_bind_call_error.go`, alongside the existing varcall error machinery. Covered by `TestVarcallArgDecodeErrorCarriesIndexAndParamType` and `TestVarcallArgDecodeErrorIndexReachableThroughWrapping` (the second covers the two-level wrap the real decoder produces, since the per-argument converter already wraps with `%w`).
- [x] 1.2 Add `func (md *GoMethodMetadata) CallWithError(inst GDClass, gdArgs ...Variant) (Variant, error)` containing the current `Call` body, returning the decode error instead of panicking; verify `go build ./...` succeeds and the existing `md.Call(nil)` test still compiles against the delegating wrapper.
  - `go build ./...` clean; `method_bind_call_error_test.go`'s `md.Call(nil)` compiles unchanged against the wrapper.

## 2. Decoder: error returns and prefix unwind

- [x] 2.1 Change `reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs` to return `([]reflect.Value, error)`, replacing the `log.Panic` on conversion error with a returned `VarcallArgDecodeError` carrying the loop index and declared parameter type; verify by unit test that a forced conversion failure returns a non-nil error with the expected `Index` and does not panic.
  - Verified in-engine rather than as a pure Go unit test: the two inducible conversion failures both need a live engine object to reach, so a host-less test could not force one. `TestVarcallDecodeWrongClass` drives the seam and asserts the error type and `Index` with no panic.
- [x] 2.2 On failure, release the owned container values decoded in the prefix before returning, passing only successfully-decoded entries to `destroyOwnedContainerArgs` (never a zero-filled slice, which panics on `Interface()`); verify with a test that decodes an owned container argument then fails on a later argument, asserting the container is released exactly once and no panic occurs.
  - Unwinds `args[1:i+1]`: the entries a completed iteration actually stored. The receiver at `args[0]` is not owned and is excluded. Covered by `TestVarcallDecodeOwnedPrefixReleased`, which also asserts the decoder returns no values on failure so the caller cannot free the same arguments twice.
- [x] 2.3 Confirm the authoring-error sites stay fatal and are not converted to returned errors — missing `Ref` constructor (`method_bind_reflect.go:210`), unsupported interface kind (`:259`), unsupported Go type (`:436`); verify by grep that those three remain `log.Fatal`/`log.Panic` and that no new error return was added at them.
  - Line numbers drifted as the file changed. The three sites are now `:221` (`log.Fatal`, missing `Ref` constructor), `:282` (`log.Panic`, unsupported interface kind) and `:459` (`log.Panic`, unsupported Go type). All still fatal. Full mapping in task 5.2.

## 3. Callback wiring

- [x] 3.1 Point `GoCallback_MethodBindMethodCall` at `CallWithError`; on a returned `VarcallArgDecodeError`, call the existing `rejectVarcallInvalidArgument` with the error's `Index` and the expected variant type from `bind.gdeArgumentTypes[Index]`, then return without invoking the method; verify by test that a decode failure writes `INVALID_ARGUMENT` with the right `argument` index into `rError` and leaves a nil return.
  - Wrapped in `rejectVarcallDecodeFailure`, which bounds-checks the index before reading `gdeArgumentTypes`. An error that is *not* a `VarcallArgDecodeError` still panics: nothing here can report it and no contract permits swallowing it.
- [x] 3.2 Emit a debug-level diagnostic on the decode reject matching the shape of `rejectVarcallArity`/`rejectVarcallInvalidArgument` (method name, error kind, `argument`, `expected`); verify by running with debug logging and confirming the line appears for an induced decode rejection.
  - Confirmed under `LOG_LEVEL=DEBUG`:
    `varcall rejected: argument decode failure {"method": "test_varcall_reject_probe", "argument": 0, "param_type": "builtin.Sprite2D", "reason": "..."}`,
    immediately preceded by the sibling `rejectVarcallInvalidArgument` line carrying `expectedType: 24` (OBJECT).
- [x] 3.3 Make `Call` a thin delegating wrapper over `CallWithError` that panics on error, preserving the documented direct-caller contract; verify the wrapper's panic message names the method and points at the varcall-validation bypass.

## 4. Tests

- [x] 4.1 Add a varcall test that passes a wrong-class object (e.g. a `Label` into a `Sprite2D`-declared parameter) from the demo scene and asserts the call is rejected with `INVALID_ARGUMENT`, the Go method body did not run, and the process survives; verify it fails against the pre-fix bindings (process aborts) and passes after.
  - Both halves were observed. Pre-fix, the same input aborted the host: `panic: reflect: Call using *gdclassimpl.LabelImpl as type builtin.Sprite2D`. That failure is what surfaced the missing implements check at `:273` (see design); the error-return work alone did not fix the scenario. Post-fix the run completes and `test_varcall_rejection_survivable` asserts the probe body never ran.
  - Note on shape: GDScript abandons the calling function once the engine reports the rejection, so the deliberate rejection is the last statement in its group and the surviving assertion lives in the group that follows.
- [x] 4.2 Add a varcall test for an object whose instance binding will not resolve, asserting the same rejection shape; verify it passes and reports the correct argument index.
  - Not covered as a varcall test, and no inducible one was found. Inside the decoder, `ObjectFromVariant` can only fail through `ObjectFromInstanceBinding`'s non-null branches (class-name lookup failure, no binding callbacks, null after registering), and the decoder registers the binding on first resolve -- so by the time a caller's object reaches `:240` a binding exists. What is covered instead: the direct-call shape via the existing `test_unresolvable_binding_is_typed_error`, and the varcall rejection shape via the wrong-class test in 4.1. `:240` stays as defence in depth on a path no caller was able to reach.
- [x] 4.3 Add a multi-argument test where an owned container argument decodes successfully and a later object argument fails; assert the container is released exactly once, the rejection is still reported, and the harness reports no leaked engine objects.
  - `TestVarcallDecodeOwnedPrefixReleased`. "Released exactly once" is carried by two things rather than one: the decoder returns no values on failure so a second release cannot be issued, and the harness's leak detection plus `clobberfree=1` turns a double free into a crash rather than a quiet pass.
  - Caught a separate hazard on the way: `NewVariantCopyWithGDExtensionConstVariantPtr` copies a Variant's bytes without taking a reference, so destroying such a copy frees memory the original still points at. Building the argument with `NewVariantArray` instead.
- [x] 4.4 Add a test that a well-formed call with an owned container argument still releases that argument after return encoding, unchanged from before this change; verify the existing success-path assertions stay green.
  - `TestVarcallDecodeSuccessReleasesOwnedContainer` for the seam, and the pre-existing consume/echo container assertions in the suite still pass unchanged.

## 5. Integration verification

- [x] 5.1 Run `make test` and verify the full suite is green — extension loaded, assertions ran, zero failures, no leaked engine objects — with no new aborts.
  - `PASS: extension loaded, 1103 assertions ran, 0 failures, no leaked engine objects` (up from 1097 before this change).
- [x] 5.2 Verify no `log.Panic` remains reachable from a caller-induced varcall decode failure: grep the varcall decode path for panic sites and confirm each remaining one is an authoring error listed in task 2.3, and record the mapping in the change notes.
  - The varcall decode path is `GoCallback_MethodBindMethodCall` → `CallWithError` → `reflectFuncCallArgsFromGDExtensionConstVariantPtrSliceArgs` → `convertVariantToGoTypeReflectValue` (`method_bind_reflect.go:101-464`). Every remaining panic inside it keys on the *declared Go type*, which the caller cannot influence -- all authoring:

    | Line | Panic | Why authoring, not caller-induced |
    |---|---|---|
    | `:207` | `slice not implemented` | the bound signature declares an unsupported Go kind |
    | `:221` | `unable to get ref constructor` (`log.Fatal`) | the declared `Ref` type has no generated constructor |
    | `:282` | `unsupported interface type` | the declared interface is one the binder cannot map |
    | `:410` | `unsupported array type` | the declared array type is unmapped |
    | `:421` | `not a ref instance` | declared type/constructor mismatch |
    | `:434` | `unsupported pointer type` | the declared pointer kind is unmapped |
    | `:445` | `not a ref instance` | declared type/constructor mismatch |
    | `:454` | `unsupported struct type` | the declared struct is unmapped |
    | `:459` | `unsupported type` | the declared Go kind is unmapped |

  - Outside the converter, `CallWithError` keeps `log.Panic("unexpected MethodBindReturnStyle")`; the return style comes from the binding, not the call, so it is authoring too.
