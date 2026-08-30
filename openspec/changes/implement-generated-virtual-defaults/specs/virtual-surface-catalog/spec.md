## Purpose

Makes Godot's virtual-method surface typed, discoverable, and compile-time verifiable in Go: generated per-class interfaces declare every overridable engine virtual under the qualified naming convention with exact signatures, a checked-in census records the full classification, and documentation states plainly — with pinned tests — that overriding replaces engine behavior wholesale because GDExtension cannot reach engine-default bodies.

## ADDED Requirements

### Requirement: Virtual Surface Census Is Complete And Classified
The repository SHALL contain a checked-in census derived from `extension_api.json` classifying every declared virtual into Category A (non-void return with a plain wrapper on the same hierarchy), B (remaining lifecycle/side-effect hooks), or C (creation-info routed names, which never appear in the API surface), with the classification criterion documented in the census file itself.

#### Scenario: Census counts match the API surface
- **WHEN** the census is regenerated from `godot_headers/extension_api.json`
- **THEN** the total equals the number of `is_virtual` entries (1437 today) and every entry appears in exactly one category

### Requirement: Generated Interfaces Declare The Virtual Surface With Exact Signatures
For every class in `extension_api.json` declaring Category-A virtuals, code generation SHALL emit an interface on that class's generated layer declaring each virtual under the qualified convention (`V_<ClassName>_<MethodName>`) with Godot's exact parameter and return types.

#### Scenario: Catalog surfaces exact override signatures
- **WHEN** a developer inspects the generated layer of a class declaring Category-A virtuals
- **THEN** each virtual appears under its qualified name with Godot's exact parameter and return types

### Requirement: Per-Method Override Signatures Are Compile-Time Verifiable
A user struct SHALL be able to verify one overridden virtual's exact signature at compile time through an anonymous-interface assertion — without satisfying any class-wide interface, without generated helper types, and without runtime behavior.

#### Scenario: Signature mismatch fails compilation
- **WHEN** a user struct asserts an anonymous interface containing a qualified virtual method whose declared types differ from the struct's actual method
- **THEN** compilation fails instead of deferring the mismatch to registration time

### Requirement: Generated Declarations Impose No Runtime Behavior
Generated virtual-surface declarations SHALL NOT register anything, alter dispatch, or add dispatch-time cost; binding a virtual remains an explicit `ClassDBBindMethodVirtual` call, and classes that never bind keep unregistered-virtual semantics unchanged.

#### Scenario: Unbound catalog leaves behavior untouched
- **WHEN** the generated interfaces are regenerated and the full suite runs
- **THEN** no additional virtuals are registered and all existing dispatch behavior is preserved

### Requirement: Delegation Impossibility Is Documented And Pinned
docs/overview.md SHALL document that a registered virtual override replaces engine behavior wholesale — engine-default bodies cannot be reached from GDExtension because Godot resolves virtual presence once per instance and never re-consults — and the test suite SHALL pin the recursion trap: a delegating override aborts at a bounded depth instead of recursing silently.

#### Scenario: Delegating override hits the bounded-depth guard
- **WHEN** a registered `_get_maximum_size` override delegates through the plain wrapper
- **THEN** the call terminates via the depth guard with a clear recursion diagnostic rather than unbounded recursion
