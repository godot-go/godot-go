## 1. Callback implementations

- [x] 1.0 Verify no flat-name virtual registrations remain in `test/pkg/example.go`: `implement-hierarchical-virtual-methods` already migrated the demo to `V_Example_*` qualified names (bind lines at example.go:839-844); confirm and note any stragglers
- [x] 1.1 In `pkg/core/classdb_callback.go`, implement `GoCallback_ClassCreationInfoPropertyCanRevert`: resolve the instance handle and registered class, look up `_property_can_revert` in `VirtualMethodMap` (absent → return 0), wrap `p_name` as a borrowed `StringName` (never destroy), call the Go method with receiver-first reflect args, and return its bool as `GDExtensionBool`
- [x] 1.2 Implement `GoCallback_ClassCreationInfoPropertyGetRevert` the same way for `_property_get_revert`: call a `(Variant, bool)`-returning method, write the Variant into `r_ret` only when handled (mirroring the `_get` write-back), return 1/0 accordingly
- [x] 1.3 Implement `GoCallback_ClassCreationInfoNotification`: resolve instance handle and registered class, look up `_notification` in `VirtualMethodMap` (absent → no-op), call `(what int32, reversed bool)` void method with receiver-first reflect args; never panics on missing registration
- [x] 1.4 Verify `go build ./pkg/core/...` compiles

## 2. Demo class coverage

- [x] 2.1 Add `V_Example_PropertyGetRevert(p_name StringName) (Variant, bool)` to `test/pkg/example.go` per design Decision 4 (`Vector3(42,42,42)` for `"property_from_list"`, unhandled otherwise) and bind it via `ClassDBBindMethodVirtual(t, "V_Example_PropertyGetRevert", "_property_get_revert", []string{"name"}, nil)`
- [x] 2.2 Add `V_Example_Notification(what int32, reversed bool)` to `test/pkg/example.go` that appends received notification codes to an exported slice, and bind it via `ClassDBBindMethodVirtual(t, "V_Example_Notification", "_notification", []string{"what", "reversed"}, nil)`
- [x] 2.3 Verify `go build ./test/...` compiles

## 3. GDScript assertions

- [x] 3.1 In `test/demo/main.gd`, add a property-revert block: `example.property_can_revert("property_from_list")` is true (value was set to non-default earlier in the suite) and false for `"dproperty_0"`; `example.property_get_revert("property_from_list")` equals `Vector3(42, 42, 42)`; also assert `property_can_revert` on an instance of a class that never bound the virtuals returns false
- [x] 3.2 Add a notification block: trigger a notification on `example` (e.g. `NOTIFICATION_ENTER_TREE` fires during scene add, or call `example.notify(...)` directly) and assert the Go-side recorded codes contain it; assert a class that never bound `_notification` accepts `notify(...)` without error
- [x] 3.3 Run the suite once with the new `_property_can_revert` dispatch temporarily returning 0 to confirm the new assertions fail loudly (sensitivity check); restore and confirm green

## 4. Validate

- [x] 4.1 Run `go vet ./pkg/core/... ./test/pkg/...`
- [x] 4.2 Run `GODOT=/path/to/godot make build` and `make test`; suite green with no orphan StringName warnings or new errors in output
- [x] 4.3 Make `make test` fail when Godot reports leaked engine objects at exit: tee run output and exit non-zero on `ObjectDB instances were leaked` / `Leaked instance:` lines (currently warnings-only, which let the notify()-era leak slip through a green run); verify the full suite still passes under the stricter gate
