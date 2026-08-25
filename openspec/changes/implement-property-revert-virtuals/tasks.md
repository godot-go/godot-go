## 1. Callback implementations

- [ ] 1.1 In `pkg/core/classdb_callback.go`, implement `GoCallback_ClassCreationInfoPropertyCanRevert`: resolve the instance handle and registered class, look up `_property_can_revert` in `VirtualMethodMap` (absent → return 0), wrap `p_name` as a borrowed `StringName` (never destroy), call the Go method with receiver-first reflect args, and return its bool as `GDExtensionBool`
- [ ] 1.2 Implement `GoCallback_ClassCreationInfoPropertyGetRevert` the same way for `_property_get_revert`: call a `(Variant, bool)`-returning method, write the Variant into `r_ret` only when handled (mirroring the `_get` write-back), return 1/0 accordingly
- [ ] 1.3 Verify `go build ./pkg/core/...` compiles

## 2. Demo class coverage

- [ ] 2.1 Add `V_PropertyGetRevert(p_name StringName) (Variant, bool)` to `test/pkg/example.go` per design Decision 4 (`Vector3(42,42,42)` for `"property_from_list"`, unhandled otherwise) and bind it via `ClassDBBindMethodVirtual(t, "V_PropertyGetRevert", "_property_get_revert", []string{"name"}, nil)`
- [ ] 2.2 Verify `go build ./test/...` compiles

## 3. GDScript assertions

- [ ] 3.1 In `test/demo/main.gd`, add a property-revert block: `example.property_can_revert("property_from_list")` is true (value was set to non-default earlier in the suite) and false for `"dproperty_0"`; `example.property_get_revert("property_from_list")` equals `Vector3(42, 42, 42)`; also assert `property_can_revert` on an instance of a class that never bound the virtuals returns false
- [ ] 3.2 Run the suite once with the new `_property_can_revert` dispatch temporarily returning 0 to confirm the new assertions fail loudly (sensitivity check); restore and confirm green

## 4. Validate

- [ ] 4.1 Run `go vet ./pkg/core/... ./test/pkg/...`
- [ ] 4.2 Run `GODOT=/path/to/godot make build` and `make test`; suite green with no orphan StringName warnings or new errors in output
