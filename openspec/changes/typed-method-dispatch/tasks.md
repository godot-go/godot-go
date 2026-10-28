# Tasks

## 1. Typed decoder interfaces and generated per-type decoders

- [x] 1.1 Define `argDecode`/`returnEncode` (ptrcall) and `varDecode` (varcall) interfaces in `pkg/core`, mirroring the parameterization of the existing decode functions (arg-vs-return direction, refcounted borrow/copy-constructor chemistry, object-argument decode, pinning behavior); verify with `go build ./pkg/core/...` and a doc-comment cross-check against `varcall-argument-decode`, `ptrcall-object-argument-decode`, `native-call-pinning` requirements
- [x] 1.2 Define the constraint unions `nativeSlot`/`retValueSlot` and the generic raw-load/store (`loadTypedSlot[A nativeSlot]`, `storeTypedRet[R retValueSlot]`) in `pkg/core/method_bind_typed.go` over the generated builtin byte-layout types (byte-parity with the reflective decoder's own raw loads; float32/Variant excluded for parity); add the per-type round-trip unit tests (engine-free byte-parity tier in `pkg/core`, `ptrslot_roundtrip_test.go`); verify membership against the reflective decode/encode matrix (Vector2-family raw load parity, String raw-store returns, Callable/Signal copy-ctor returns)
- [ ] 1.3 Document and enforce the fallback boundary for object-argument signatures: `GDClass`/`Ref`-typed parameters are outside the `nativeSlot` union by design (binding resolution / ref-acquire is per-call work), so binds whose signatures contain them keep the reflective path via the name-string API; add a unit test asserting such binds store no typed closure and instance resolution still matches `ptrcall-object-argument-decode` scenarios through the fallback
- [x] 1.4 Emit per-type `RetEncodeInto` copy-constructor methods for the 16 copy-ret builtins (StringName, NodePath, Callable, Signal, Array, Dictionary, 10 Packed*) via `cmd/generate/typedwrapper`, each reproducing its existing copy constructor call with no new pinning or destroy policy (the copy ctors pin their source call-scoped; the `destroySource` gate is the caller's per borrow-echo classification), co-emitting the exported `RetCopyEncodable` constraint next to the methods it constrains (exported because a dot-import only brings exported identifiers into scope) and the copy-ret wrapper family delegating to them through the constraint; unit tests confirm stored values shape-check engine-free (16 compile witnesses + instance/static stub spot tests + alloc budgets) and the engine tier's reflective-parity fixtures assert byte equality

## 2. Wrapper family generation

- [x] 2.1 Add the ptrcall arity-family template (arity 0..16 × {value-ret, void} × {instance, static}) to `cmd/generate/typedwrapper` and emit `pkg/core/wrappers.gen.go`; instance wrappers resolve receivers through the `selfTyped` contract (probe-verified: generic assertions to `*T` are compile-impossible, the constraint-carried interface target legal); verify generation output compiles (`go build ./pkg/core/...`) and spot-check representative arities (0 and 2, instance+static+value+void) engine-free through synthesized cells with a stub class (`typed_wrapper_test.go`); varcall wrappers are scope of task 5.1
- [ ] 2.2 Unit-test each generated wrapper kind: typed function invoked through the closure produces identical arguments/return as the reflective path; record `testing.AllocsPerRun` budgets observed (expected small constant, no reflect frame) as the baseline the wiring tasks must match

## 3. Typed registration overloads

- [ ] 3.1 Add typed overloads to `pkg/core/classdb.go` accepting method expressions for instance/static/virtual binds; they build the compiled closure at bind time and reuse ALL existing validation (duplicate-name rejection, `V_` prefix rules, embedding-chain virtual resolution); verify unit tests cover validation parity with the name-string path, and document the typed API + fallback contract in the `classdb.go` package doc
- [ ] 3.2 Add a compile-rejection fixture (testdata package + `go test` harness asserting build failure) proving a mismatched-handler bind via typed overload does not compile; verify the fixture fails without the fix and passes with it

## 4. Wire ptrcall and virtual dispatch to compiled closures

- [ ] 4.1 Add the `dispatch` closure fields to `GoMethodMetadata` with fallback selection (`dispatch == nil` → reflective path) in `Ptrcall` and the virtual-callback path; verify typed-bound methods dispatch through closures in unit tests and the existing suites stay green
- [ ] 4.2 Confirm the reflective path byte-for-byte unchanged for name-string binds; verify by diffing dispatch flow in a unit test that forces fallback via an untypeable signature and compares results against a typed bind's results

## 5. Varcall typed closures with defaults/variadic parity

- [ ] 5.1 Implement typed varcall wrappers: direct variadic closures (`...Variant`, no `CallSlice`) and non-variadic wrappers reproducing `fillCallArgs` trailing-default convention including the loud under-supply panic; verify existing varcall suites extended to typed-bound examples stay green and the panic test parity holds
- [ ] 5.2 Re-implement borrow-echo classification with typed compare helpers per D5 and add differential fixtures (echo, non-echo, COW same-content clone, StringName intern equality, variadic no-suppression) asserting identical classification on both paths; verify every `borrow-echo-detection` scenario passes through the typed path

## 6. Fixed-shape callback closures

- [ ] 6.1 Replace the generic `Func.Call` sites in `pkg/core/classdb_callback.go` (notification, property get/set, property list, revert, string-conversion) with named typed closures attached at registration, name-string handlers falling back; verify `notification-virtual-dispatch`, `property-access-dispatch`, `property-revert-virtuals`, `to-string-virtual-dispatch` suites stay green

## 7. Integration and allocation gates

- [ ] 7.1 Convert representative `test/demo` classes to the typed overloads while keeping at least one name-string-bound class exercising the fallback; verify `GODOT=/path/to/godot make test` passes both gates (runtime suite + editor run) with typed dispatch live
- [ ] 7.2 Add the `AllocsPerRun` budget test covering a representative arity matrix (0/2/5 args, each return style, virtual and static) asserting per-call allocation caps; verify the test fails if any reflect machinery is reintroduced on the dispatch path
- [ ] 7.3 Run the full suite plus `make test_gate` and confirm zero leaked engine objects / shutdown markers unchanged; verify logs show no new orphan StringNames or leak counters drifting

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
- Verify the archived result.
