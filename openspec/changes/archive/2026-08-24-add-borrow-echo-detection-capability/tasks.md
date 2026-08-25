## 1. Verify spec claims against code

- [x] 1.1 Confirm the covered-type list in the capability matches `isPtrcallBorrowEcho`'s type switch exactly (`StringName`, `NodePath`, `Array`, nine `Packed*Array`) and that both return-path call sites consult it while the variadic branch always destroys
- [x] 1.2 Confirm no production code change is required for any scenario in this capability (all describe current behavior)

## 2. Validate

- [x] 2.1 Run `openspec validate add-borrow-echo-detection-capability`
- [x] 2.2 Run `go build ./...` as a sanity check that the repo is untouched
