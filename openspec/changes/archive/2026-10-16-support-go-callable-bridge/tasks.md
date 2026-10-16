# Tasks

## 1. Creation and invocation path

- [x] 1.1 Add `pkg/core/go_callable.c/.h`: extern declarations for
  `GoCallback_GoCallableCall/Free/ToString` plus `cgo_callcustom_*` shim
  functions, mirroring the `classdb_callback.c` technique. Verify: `make build`
  compiles the extension.
- [x] 1.2 Add `pkg/core/go_callable.go`: `NewGoCallable` API, `cgo.Handle`
  registry, `GDExtensionCallableCustomInfo2` population (`call_func`,
  `free_func`, `to_string_func`, `token = FFI.Library`, `object_id = 0`),
  `callable_custom_create2` into the returned `builtin.Callable`; implement
  `GoCallback_GoCallableFree` as `h.Delete()` + closure reference drop. Verify:
  `make build` and `go vet ./pkg/core/...` clean.
- [x] 1.3 Implement `GoCallback_GoCallableCall`: handle lookup, borrowed
  `[]builtin.Variant` views over `p_args`, closure invocation, bit-copy return
  into `r_return` (zero-value Variant falls back to `writeNilCallReturn`),
  `setCallErrorOK`. Verify: `make build` compiles.
- [x] 1.4 Add `test/pkg` fixture `CreateCounterCallable() builtin.Callable`
  (captured counter closure) and `test/demo` GDScript asserting `is_valid()`
  and `cb.call()` returning 1, 2, 3. Verify: `GODOT=/home/pcting/bin/godot
  make test` shows those scenario PASSES.
- [x] 1.5 Add a copy-storage fixture (closure copies a String argument with
  `NewVariantNativeCopy`, a later invocation returns the stored copy) and
  GDScript asserting the stored copy reads back equal after the originating
  call ended — pins the borrowed-argument copy-if-store contract. Verify:
  `make test` PASS. (Added during apply: the spec's "Copy stored beyond the
  call stays valid" scenario had no task.)

## 2. Failure policy

- [x] 2.1 Implement the closure-error path: zap-log the error,
  `writeNilCallReturn`, `r_error = GDEXTENSION_CALL_ERROR_INVALID_METHOD` with
  `argument/expected/passed = -1`. Verify: `make build` compiles.
- [x] 2.2 Wrap the callback body in panic-recover: log stack via the existing
  log package, report the same call failure; no panic unwinds into cgo. Verify:
  `make build` compiles.
- [x] 2.3 Add fixtures `CreateErrorCallable()` / `CreatePanicCallable()` and
  GDScript scenarios: `cb.call()` raises a script-side error, engine keeps
  running, and a subsequent `cb.call()` on the same callable still works.
  Verify: `make test` green with those PASSES and no crash.

## 3. Identity, dedup, string, deferred calls

- [x] 3.1 Implement `GoCallback_GoCallableToString` using
  `GDExtensionStringPtrWithUtf8Chars` and writing `*r_is_valid = 1` through the
  pointer (class `to_string` pattern). Verify: GDScript `str(cb)` assertion on
  the Go label PASSES in `make test`.
- [x] 3.2 Add GDScript dedup scenarios: connecting the same Callable twice to
  one signal yields one connection and one closure run per emission; two
  separate `NewGoCallable` values yield two connections and both bodies run.
  Verify: `make test` PASSES for both.
- [x] 3.3 Add a `cb.call_deferred(...)` GDScript scenario asserting the same
  closure body runs (main-thread deferred) and state updates. Verify:
  `make test` PASS after the deferred frame.

## 4. Lifecycle and leak gate

- [x] 4.1 Add a registry-count accessor in `pkg/core` and fixture method
  `GoCallableRegistryCount() int64`. Verify: `make build` compiles.
- [x] 4.2 Add GDScript scenario: create N callables, record count, null out
  references, assert the count decreases by N without Go-side action. Verify:
  `make test` PASS.
- [x] 4.3 Add the final registry assertion last in the GDScript suite
  (`TestRefBoundaryRelease`-last pattern): only intentionally retained
  callables remain registered before the shutdown marker. Verify: `make test`
  green including the ObjectDB leak gate.

## 5. Reentrancy and documentation

- [x] 5.1 Add fixture: a Go Callable whose closure invokes a bound Go class
  method on a captured object during its call, and GDScript asserting the
  nested result. Verify: `make test` PASS (no deadlock, correct value).
- [x] 5.2 Update `docs/roadmap.md`: replace the stale C1 "Risk to spike
  first" paragraph with the spike outcome (the `callable_custom_get_user_data`
  gap was an interface-loader name bug fixed by `fix-ffi-loader-names`) and
  mark C1 done per the DONE-note convention. Verify: diff reviewed, no other
  roadmap sections touched.
- [x] 5.3 Integration: `GODOT=/home/pcting/bin/godot make test` green per
  `test/check_test_output.sh` (TESTS FINISHED, PASSES > 0, FAILURES = 0,
  shutdown marker, no leaks, exit 0) and `openspec validate
  support-go-callable-bridge --strict` passes.
