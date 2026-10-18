## Purpose

Governs the return buffer that generated bindings for vararg Godot methods pass across the GDExtension boundary, so that Godot never writes past the Go storage it is given and every vararg method returns its declared type intact.

## Requirements

### Requirement: Vararg Return Buffer Is A Variant
A generated binding for a vararg method SHALL pass a Go `Variant`-sized return buffer to the engine, regardless of the method's declared return type, because the engine writes a full `Variant` for every vararg call.

#### Scenario: Vararg method with a non-Variant declared return
- **WHEN** a generated binding is produced for a vararg method whose declared return type is `Error`
- **THEN** the return buffer passed to the engine is sized and typed as `Variant`, not as `Error`

#### Scenario: Return value is converted to the declared type
- **WHEN** the engine returns from a vararg call with a `Variant` result
- **THEN** the binding converts that result to the declared return type and releases the `Variant`

### Requirement: Vararg Return Conversion Preserves Declared Type
A generated vararg binding SHALL return the declared Go return type, and the conversion from `Variant` SHALL yield the same value the engine produced for that call.

#### Scenario: Error-returning vararg method
- **WHEN** `Object.emit_signal`, `Node.rpc`, or `Node.rpc_id` is called and the engine returns an error code
- **THEN** the Go method returns that error code as `Error`

#### Scenario: Variant-returning vararg method is unchanged
- **WHEN** a vararg method whose declared return type is `Variant` is called
- **THEN** the binding returns the engine's `Variant` result without an extra conversion

### Requirement: Repeated Vararg Calls Do Not Corrupt Adjacent Memory
Repeated vararg calls SHALL NOT overwrite memory outside their return buffer, so Go data returned by other calls stays intact across garbage collection.

#### Scenario: Signal emit interleaved with a long string return
- **WHEN** a loop emits a signal through `Object.emit_signal` and then calls a method that returns a long Go string of about 1 KB, for several hundred iterations, with garbage collection running
- **THEN** every returned string matches the value it was built with, and no string is truncated or contains unrelated data

### Requirement: Generator Rejects Unsupported Vararg Return Types
The code generator SHALL reject a vararg method whose declared return type cannot be carried by a `Variant` buffer and conversion, rather than generating a binding with an undersized buffer.

#### Scenario: New vararg method with a supported return type
- **WHEN** the generator processes a vararg method that returns `Variant`, `Error`, or no value
- **THEN** it generates a binding with a `Variant` return buffer

#### Scenario: New vararg method with an unsupported return type
- **WHEN** the generator processes a vararg method whose declared return type has no supported conversion from `Variant`
- **THEN** generation fails with an error naming the method and its return type

### Requirement: Built-In And Utility Vararg Returns Are Audited
Every vararg method in the class, built-in class, and utility function bindings SHALL have a return buffer whose size matches its declared return type, or a `Variant` for `Variant`-returning methods.

#### Scenario: Audit finds a mismatch
- **WHEN** the audit finds a vararg method whose return buffer is smaller than the value Godot writes
- **THEN** that method is fixed under the same requirements as `emit_signal`, `rpc`, and `rpc_id`

#### Scenario: Audit finds no mismatch
- **WHEN** the audit finds that a vararg method's return buffer already matches its declared type
- **THEN** that method is left unchanged and recorded as audited
