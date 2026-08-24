## Purpose

Defines how the Go bindings classify values returned from Go methods against the borrow-decoded arguments of the same call, deciding whether the Go-side source value is destroyed after its contents are encoded into Godot's output. Covers all refcounted built-in types that use borrow argument decoding (`StringName`, `NodePath`, `Array`, and the `Packed*Array` types) on both the ptrcall and varcall return paths, and documents the inherent limitation of content-based classification.

## ADDED Requirements

### Requirement: Returned Values Are Classified Against Borrow-Decoded Arguments By Content
Before destroying a returned refcounted built-in, the return path SHALL compare it byte-for-byte with every argument decoded as a borrow for this call. A returned value byte-equal to such an argument SHALL be classified as a borrow echo and its source left undestroyed (the encoded output already received its own reference); any other returned value SHALL be classified as owned and its source destroyed after encoding.

#### Scenario: Echo leaves the source undestroyed
- **WHEN** a Go method returns the same borrowed argument it received, via either call style
- **THEN** the output receives a fresh reference and the borrowed source is not released

#### Scenario: Distinct value is destroyed after encoding
- **WHEN** a Go method returns a value whose bytes differ from every borrow-decoded argument of the call
- **THEN** the source is classified as owned and destroyed after the output is encoded

### Requirement: Same-Content Clones Are Misclassified As Echoes
Content-based classification cannot distinguish an owned copy from the borrow it was cloned from when their bytes are identical — Godot copy constructors share storage (copy-on-write) and StringName interns data. An unmutated defensive clone returned through either path SHALL be treated as a borrow echo, silently retaining exactly one reference per call; this bounded leak is accepted behavior and SHALL NOT crash or corrupt state. Methods that clone an argument and return it unmutated accept this leak; mutating the clone before returning, or returning a freshly built value, avoids it.

#### Scenario: Unmutated defensive clone leaks one reference per call
- **WHEN** a Go method clones a borrowed container or interned-type argument without modifying it and returns the clone through either return path
- **THEN** the clone is classified as a borrow echo and its own reference is retained, one leaked reference per call, with no crash and no double-free

#### Scenario: Mutated clone avoids misclassification
- **WHEN** the clone's bytes differ from the borrow before being returned (because the method mutated it, or built the result independently)
- **THEN** it is classified as owned and destroyed normally after encoding

### Requirement: Detection Spans The Refcounted Built-In Types On Both Return Paths
The same classification mechanism SHALL apply to `StringName`, `NodePath`, `Array`, and the nine `Packed*Array` types on both the ptrcall and varcall return paths. Return paths that cannot receive borrows (calls whose arguments are never decoded, such as variadic dispatch) SHALL NOT apply echo suppression and always destroy Go-created sources.

#### Scenario: Variadic dispatch always destroys
- **WHEN** a variadic Go method returns a refcounted built-in it constructed from raw Variants
- **THEN** no echo classification applies and the source is destroyed after encoding
