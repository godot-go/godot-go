# Tasks

## 1. Return-resolution helper

- [x] 1.1 Add the unexported `resolveGoStructObjectReturn[T Object](declared T) T` to a hand-written `pkg/gdclassimpl/object_return.go` (sibling of the generated files): zero for a typed-nil declared wrapper, bound wrapper when it satisfies `T`, declared wrapper otherwise; split the engine-free half (`resolveObjectReturnFrom`) so the four branches unit-test with fake wrappers (`go test ./pkg/gdclassimpl/`); confirm `make generate` leaves the file in place.

## 2. Generator return branch

- [x] 2.1 Change the non-refcounted object-return branch of `cmd/generate/gdclassimpl/classes.go.tmpl` from `return &ret` to `return resolveGoStructObjectReturn[{{ $fnReturnType }}](&ret)`; the type argument must be explicit because inference binds `T` to the declared struct pointer and every call would fall back. Leave the `NewRef*GDExtensionReturnOwner` branch untouched and verify by reading the rendered branch in a regenerated class (e.g. `NodeImpl.GetNodeOrNull`).
- [x] 2.2 Regenerate with `GODOT=... make generate` and verify the only `classes.gen.go` changes are object-return lines (`git diff --stat`, plus spot-check `GetNodeOrNull` and a Ref-returning method such as `GetShape`), and that `pkg/gdclassimpl/object_return.go` survives the regeneration.
- [x] 2.3 Verify the library still builds and vets (`make build`, `go vet ./pkg/...`).

## 3. Witness and documentation for the flipped contract

- [x] 3.1 Update `Example.TestObjectCasting` in `test/pkg/example.go`: the plain `node.(Label)` assertion must now succeed with no library helper call in between, keep the explicit `ObjectFromInstanceBinding` clause as coverage for non-generated wrappers, and add a repeated-return identity check; verify with `make build`.
- [x] 3.2 Update the `test_object_casting` comment in `test/demo/main.gd` so it describes narrowing without re-resolution, and verify `example.test_object_casting()` still returns 0 in the runtime suite.
- [x] 3.3 Rewrite the "Narrowing to a subclass" section of `docs/overview.md` and record the generated-return resolution in the capability-map row (describe the behaviour, not the unexported helper name); verify the snippets match the new witness and that no section still claims a generated return needs re-resolution.

## 4. Integration verification

- [x] 4.1 Run the full gate with `GODOT=... make test` and verify both the runtime suite and the editor-plugin gate exit green, with no new leak or binding-accounting failures in `test-output.log`.
- [x] 4.2 Validate the change artifacts with `openspec validate resolve-object-returns-through-binding --strict` and verify it reports the change as valid.

## Workflow follow-up

- Archive the change once review requirements are satisfied.
- Verify the archived spec landed in `openspec/specs/instance-binding-lookup/spec.md`.
