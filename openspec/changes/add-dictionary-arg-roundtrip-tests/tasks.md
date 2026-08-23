## 1. Example methods

- [ ] 1.1 In `test/pkg/example.go`, add `TestEchoDictionary(d Dictionary) Dictionary` (returns the received dictionary), `TestConsumeDictionary(d Dictionary) int64` (returns a scalar derived from contents, e.g. key count), and `TestMutateDictionary(d Dictionary)` (adds one known key), following the existing container test-method patterns
- [ ] 1.2 Register the three methods via `ClassDBBindMethod` in `RegisterClassExample`
- [ ] 1.3 Verify `go build ./test/...` compiles

## 2. GDScript assertions

- [ ] 2.1 In `test/demo/main.gd`, add varcall assertions: `example.call("test_echo_dictionary", d)` equals the input contents; `example.call("test_consume_dictionary", d)` returns the expected scalar; `example.call("test_mutate_dictionary", d)` makes the new key visible in the caller's original dictionary (shared-reference semantics)
- [ ] 2.2 Probe direct ptrcall dispatch (`example.test_echo_dictionary(d)`) for a `Dictionary` argument: if unsupported, assert the failure is loud and deterministic (document the observed behavior in a comment); if it works, add echo/consume assertions for that call style too
- [ ] 2.3 Verify no orphan/double-free symptoms appear in test output after the new block

## 3. Validate

- [ ] 3.1 Run `go vet ./test/...` on changed files
- [ ] 3.2 Run `GODOT=/path/to/godot make build` and `make test`; suite green with the new assertions included
