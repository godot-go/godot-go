## 1. Registration resolution

- [x] 1.1 In `pkg/core/classdb.go`, add the qualified-name resolver per design Decision 2: given receiver type T and GDScript method name, compute `V_<TypeName>_<PascalMethodName>` for T and each embedded Impl field (depth-first, most-derived first); return the first match
- [x] 1.2 Wire the resolver into `classDBBindMethodVirtual`'s path so virtual registration uses it while non-virtual binding keeps exact-name lookup unchanged
- [x] 1.3 Panic with a clear diagnostic when no qualified implementation resolves anywhere in the chain, or when a caller passes a flat `V_<Method>` Go name directly: the message names the requested GDScript virtual, the searched qualified names, and the expected `V_<ClassName>_<Method>` shape (design Decision 3 — clean break, no fallback)
- [x] 1.4 Verify `go build ./pkg/...` compiles

## 2. Demo hierarchy

- [x] 2.1 In `test/pkg`, add a two-level demo: a wrapper struct embedding an engine `Impl` leaf that declares a base-level qualified virtual, plus a user class embedding that wrapper with its own derived-level qualified override of the same virtual; register both classes via `ClassDBRegisterClass` + `ClassDBBindMethodVirtual`
- [x] 2.2 Migrate every `Example` flat virtual (`V_ToString`, `V_Ready`, `V_Input`, `V_Set`, `V_Get`, `V_PropertyCanRevert`) to qualified form (`V_Example_Ready`, ...) — required, since flat registration now panics and the suite must start clean
- [x] 2.3 Verify `go build ./test/...` compiles

## 3. GDScript assertions

- [x] 3.1 In `test/demo/main.gd`, add assertions: the derived class's virtual dispatches to the derived implementation; a second instance exercising only the base level dispatches to the base implementation; delegation from derived to base produces the composed expected result; the unimplemented virtual on some other registered class still falls back cleanly (no error output)
- [x] 3.2 Sensitivity check: temporarily reintroduce one flat registration (e.g. bind `V_Ready` for `_ready`) and confirm startup panics immediately with the migration diagnostic from task 1.3; restore qualified form and confirm green

## 4. Docs and validate

- [x] 4.1 Update `docs/overview.md#virtual-methods`: document the required qualified convention, most-derived-wins rule, explicit delegation pattern, and that flat `V_<Method>` registration panics by design
- [x] 4.2 Run `go vet ./pkg/core/... ./test/pkg/...`
- [x] 4.3 Run `GODOT=/path/to/godot make build` and `make test`; suite green with no new errors or orphan warnings in output
