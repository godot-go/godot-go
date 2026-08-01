## Why

`make test` fails on `test_set_position_and_size()`: after `SetSize(100, 200)` the control's `get_size()` still returns `(0, 0)` — setting the size has no effect. This surfaced after upgrading the bindings to Godot 4.7.2.

The previous investigation misattributed this to a "GDExtension method hash collision" and worked around it by (1) documenting a fake engine limitation, (2) skipping the failing assertions, and (3) rerouting generated `Control` methods through `Object.Call()`. That theory is wrong. Godot's `classdb_get_method_bind` and `Object::callp` both resolve methods **by name**; the method hash is only a verification value, and collisions among method hashes are handled by name-keyed maps (verified against Godot source). CI on the 4.6.3 baseline passed with the full assertion active.

The real cause is a godot-go binding bug in GDVIRTUAL dispatch. `GetVirtualCallWithData` always returns non-nil call data, so Godot believes every virtual is overridden. For a Go class that does not implement a virtual with a return value (e.g. `_get_maximum_size`), the callback returns without writing the return buffer and Godot reads uninitialized memory: `Control::get_maximum_size()` returns `(0, 0)` instead of the engine default `(-1, -1)`. Godot 4.7 added a maximum-size clamp to `Control::set_size`, so every GDExtension `Control` subclass now clamps its size to `(0, 0)`.

## What Changes

- **Fix GDVIRTUAL dispatch** so unimplemented virtuals return `nil` call data, letting Godot fall back to engine defaults.
- **Remove the `hasHashCollision` generator hack**; `Control` methods return to the standard method-bind codegen path.
- **Restore the skipped test coverage** (`SetSize` call and the `get_size() == (100, 200)` assertion).
- **Remove the incorrect hash-collision documentation** and README claims.

## Capabilities

### New Capabilities

- `gdvirtual-unimplemented-defaults`: Unimplemented GDVIRTUAL methods on GDExtension classes fall back to engine defaults instead of uninitialized memory.

### Modified Capabilities

## Impact

- `pkg/core/classdb_callback.go` — return `nil` from `GetVirtualCallWithData` for unimplemented virtuals
- `cmd/generate/gdclassimpl/classes.go.tmpl`, `templatefunctions.go`, `generate.go` — remove the `hasHashCollision` branch
- `pkg/gdclassimpl/classes.gen.go` — regenerated; `Control` methods use the method-bind path
- `test/pkg/example.go`, `test/demo/main.gd` — restore assertions
- `README.md`, `docs/hash-collisions.md` — remove incorrect claims

## Non-goals

- Fixing Godot's GDVIRTUAL machinery (it behaves as designed).
- Reverting the Godot 4.7.2 header upgrade (this change makes the binding correct on 4.7).
- "Fixing" method hash values (they are correct and harmless).
- Auditing every GDVIRTUAL for other uninitialized-return call sites.
