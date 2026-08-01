## Purpose

GDExtension classes that do not implement a GDVIRTUAL method must fall back to the engine's default return value instead of uninitialized memory. This fixes `Control::set_size` clamping to `(0,0)` on GDExtension `Control` subclasses (Godot 4.7+), which was previously misdiagnosed as a method-hash collision.

## Requirements

### Requirement: Unimplemented GDVIRTUAL methods fall back to engine defaults

When a GDExtension-registered Go class does not implement a GDVIRTUAL method that has a return value, the engine SHALL use its default return value for that virtual.

#### Scenario: Control.get_maximum_size returns the engine default
- **GIVEN** a Go class registered as a `Control` subclass that does not override `_get_maximum_size`
- **WHEN** `get_maximum_size()` is called
- **THEN** the result is `(-1, -1)` (the engine default), not `(0, 0)`

#### Scenario: Unimplemented virtuals do not claim call data
- **GIVEN** a Go class registered with the engine
- **WHEN** Godot queries `get_virtual_call_data2` for a virtual the class does not implement
- **THEN** the result is `nil`, so Godot treats the virtual as not overridden

### Requirement: Control.set_size takes effect on GDExtension Control subclasses

Calling `set_size` on a GDExtension `Control` subclass SHALL change the control's size, verifiable via `get_size()`.

#### Scenario: Size is set and retrieved
- **WHEN** `SetSize(Vector2(100, 200), true)` is called on an `Example` (a GDExtension `Control` subclass)
- **THEN** `GetSize()` returns `Vector2(100, 200)`

#### Scenario: TestSetPositionAndSize passes with full assertions
- **WHEN** the demo test suite runs `test_set_position_and_size`
- **THEN** both `get_position() == Vector2(320, 240)` and `get_size() == Vector2(100, 200)` pass

### Requirement: No generated method is rerouted for hash reasons

Generated method bindings SHALL use the standard method-bind path; no generated method SHALL be rerouted to `Object.Call()` based on method-hash values.

#### Scenario: SetSize uses the method-bind path
- **GIVEN** the generated `ControlImpl.SetSize`
- **WHEN** the class codegen is inspected
- **THEN** it uses `classdb_get_method_bind` + `object_method_bind_ptrcall`, not `Object.Call()`
