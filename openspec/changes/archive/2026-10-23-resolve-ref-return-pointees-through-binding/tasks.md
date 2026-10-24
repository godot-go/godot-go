# Tasks

## 1. Template and codegen

- [x] 1.1 Edit the refcounted-return branch in `cmd/generate/gdclassimpl/classes.go.tmpl` to wrap `&ret` with `resolveGoStructObjectReturn[{{ $fnReturnType }}](&ret)`, and verify the non-refcounted and null-return branches are untouched
- [x] 1.2 Run `GODOT=/home/pcting/bin/godot make generate` and verify `pkg/gdclassimpl/classes.gen.go` has exactly 574 refcounted-return sites using `resolveGoStructObjectReturn`
- [x] 1.3 Add a one-line note to the doc comment of `resolveGoStructObjectReturn` in `pkg/gdclassimpl/object_return.go` stating it is used by both non-refcounted and refcounted return paths, and verify `go build ./pkg/gdclassimpl/` succeeds

## 2. Witness tests

- [x] 2.1 Add `TestReturnRefPointeeNarrowing` to `test/pkg/return_ownership_tests.go` — call `CollisionShape2D.GetShape()`, assert `Ptr().(CircleShape2D)` succeeds, and verify it returns 1
- [x] 2.2 Add `TestReturnRefPointeeIdentity` to `test/pkg/return_ownership_tests.go` — two `GetShape()` calls return `Ref` values whose `Ptr()` are equal, and verify it returns 1
- [x] 2.3 Register both witness methods in `test/pkg/example.go` (method info + class info) and verify `make build` succeeds
- [x] 2.4 Add `assert_equal(example.test_return_ref_pointee_narrowing(cs), 1)` and `assert_equal(example.test_return_ref_pointee_identity(cs), 1)` to `test/demo/main.gd` inside `test_return_ownership`, and verify `GODOT=/home/pcting/bin/godot make test` passes both assertions

## 3. Documentation corrections

- [x] 3.1 Fix the stale comment at `test/pkg/ref_method_boundary_tests.go:76-78` — replace "Ptr() yields the declared class" with text noting the pointee is the runtime-class wrapper via the binding, and verify `go build ./test/...` compiles
- [x] 3.2 Update `docs/overview.md` narrowing section to show `Ptr().(CircleShape2D)` succeeding after `GetShape()`, and update the capability-map row for `instance-binding-lookup` to cover refcounted pointees

## 4. Integration verification

- [x] 4.1 Run `GODOT=/home/pcting/bin/godot make test` and verify 1485+ passes, 0 failures, exit-leak-check clean, and editor gate PASS
- [x] 4.2 Run `go vet ./pkg/gdclassimpl/ ./pkg/builtin/` and verify zero new warnings

## Workflow follow-up

- Sync the MODIFIED and ADDED requirements into the main specs at archive time.
- Revert any `test/demo/icon.png.import` drift from the test run before committing.
