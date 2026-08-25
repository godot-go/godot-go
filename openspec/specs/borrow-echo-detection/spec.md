## Purpose

Defines how the Go bindings classify values returned from Go methods against the borrow-decoded arguments of the same call, deciding whether the Go-side source value is destroyed after its contents are encoded into Godot's output. Covers all refcounted built-in types that use borrow argument decoding (`StringName`, `NodePath`, `Array`, and the `Packed*Array` types) on both the ptrcall and varcall return paths, and documents the inherent limitation of content-based classification.

## Requirements

### Requirement: Returned Values Are Classified Against Call Arguments By Content
Before any source release on a return path, the returned refcounted built-in SHALL be compared byte-for-byte with every argument of the call. A returned value byte-equal to an argument SHALL be classified as a borrow echo and its source left unreleased (the encoded output already received its own reference); echo suppression is what prevents the return path from releasing a reference Godot's caller still holds.

#### Scenario: Echo leaves the source undestroyed
- **WHEN** a Go method returns the same borrowed argument it received, via either call style
- **THEN** the output receives a fresh reference and the borrowed source is not released

### Requirement: Non-Echo Source Release Depends On Type And Return Path
Where a return path releases Go-created sources after encoding — ptrcall for all covered types except `PackedVector4Array` (whose ptrcall return is a raw byte-copy with no source release), and varcall for `StringName`/`NodePath` — a value classified as owned (not an echo) SHALL be destroyed after encoding. Container values returned through varcall are not released by the return path (their varcall lifecycle is governed by owned-copy argument decoding and post-call release per `container-lifecycle-management`; that Go-created container sources are not released by the return encoder itself is recorded deferred behavior there).

#### Scenario: Distinct value on a releasing path is destroyed after encoding
- **WHEN** a Go method returns a StringName, NodePath (via either call style), or container other than `PackedVector4Array` (via ptrcall) whose bytes differ from every argument of the call
- **THEN** the source is classified as owned and destroyed after the output is encoded

#### Scenario: Container returned via varcall has no return-path release
- **WHEN** a Go method returns an `Array` or `Packed*Array` through the varcall return path
- **THEN** the output Variant holds its own reference and the return path performs no source release; lifecycle accounting for the source follows the container capability's owned-copy rules

### Requirement: Same-Content Clones Are Misclassified As Echoes
Content-based classification cannot distinguish an owned copy from the borrow it was cloned from when their bytes are identical — Godot copy constructors share storage (copy-on-write) and StringName interns data. An unmutated defensive clone returned through a path that would release it (ptrcall for all covered types except `PackedVector4Array`, whose ptrcall return performs no source release; varcall for `StringName`/`NodePath`) SHALL be treated as a borrow echo, silently retaining exactly one reference per call; this bounded leak is accepted behavior and SHALL NOT crash or corrupt state. Methods that clone an argument and return it unmutated accept this leak; mutating the clone before returning, or returning a freshly built value, avoids it.

#### Scenario: Unmutated defensive clone leaks one reference per call
- **WHEN** a Go method clones a borrowed container or interned-type argument without modifying it and returns the clone through a releasing return path
- **THEN** the clone is classified as a borrow echo and its own reference is retained, one leaked reference per call, with no crash and no double-free

#### Scenario: Mutated clone avoids misclassification
- **WHEN** the clone's bytes differ from the borrow before being returned (because the method mutated it, or built the result independently)
- **THEN** it is classified as owned and destroyed normally after encoding

### Requirement: Detection Spans The Refcounted Built-In Types
The same classification mechanism SHALL apply to `StringName`, `NodePath`, `Array`, and the nine `Packed*Array` types wherever a return path performs source release. Dispatch styles whose methods receive raw Variants and never hold borrows (variadic calls) SHALL NOT apply echo suppression.

#### Scenario: Variadic dispatch applies no echo suppression
- **WHEN** a variadic Go method returns a refcounted built-in it constructed from raw Variants
- **THEN** no echo classification applies and the source is released per the type-and-path matrix above
