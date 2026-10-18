# Tasks

## 1. Audit vararg return buffers

- [x] 1.1 Script a check over `godot_headers/extension_api.json` that lists every vararg method in classes, built-in classes, and utility functions with its declared return type, and verify the output matches the audit table in design.md (15 class, 6 built-in, 12 utility).
- [x] 1.2 Confirm the built-in and utility vararg paths (`CallBuiltinMethodPtrRet`, utility ptrcall) size their buffers to the declared type, and record the result in design.md; verify with a grep showing no `Variant`-declared buffer is undersized.

## 2. Fix generated vararg class calls

- [x] 2.1 In `cmd/generate/gdclassimpl/classes.go.tmpl`, make vararg methods always use a `Variant` return buffer and convert to the declared type (`Error` via `ToInt64`, `Variant` returned as-is, void unchanged); verify with `make generate` and a diff showing only `emit_signal`, `rpc`, and `rpc_id` change in `pkg/gdclassimpl/classes.gen.go`.
- [x] 2.2 Add the generator guard that fails for a vararg method whose return type has no supported conversion, naming the method and type; verify by a generator unit test that an unsupported type returns an error and a supported type generates code.
- [x] 2.3 Destroy the `Variant` buffer after converting an `Error` return and verify with `make build` and a grep of the regenerated methods showing a `Destroy` call on the return `Variant`.

## 3. Regression tests

- [ ] 3.1 Add a `test/demo` class with a signal and a method returning a 1 KB Go string; add a test that loops several hundred `EmitSignal` calls interleaved with the string call under GC and checks each string; verify with `make test` passing.
- [ ] 3.2 Add a test that `EmitSignal`, `Rpc`, and `RpcId` return the expected `Error` value; verify with `make test`.
- [ ] 3.3 Run the regression test several times on the fixed build and record the pass count; verify all runs pass, and that the GOGC=off control is unaffected.

## 4. Verification

- [ ] 4.1 Run `make generate`, `make build`, and `make test` on a clean tree and verify all three succeed with no new warnings.
- [ ] 4.2 Run `openspec validate fix-vararg-return-buffer --strict` and verify it passes.

## Workflow follow-up

- Archive the change after the project's review requirements are satisfied.
