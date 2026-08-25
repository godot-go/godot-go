## Purpose

Extends `container-lifecycle-management` with explicit requirements for `Dictionary` arguments, whose decode/release and shared-reference semantics were specified generically but never exercised end-to-end.

## ADDED Requirements

### Requirement: Dictionary Arguments Round-Trip Without Retention
A `Dictionary` received as an argument SHALL round-trip with correct contents through every supported call style, and consuming one SHALL NOT leave a Go-side reference retained after the call. Because `Dictionary` is a shared-reference container without copy-on-write, a mutation made by the Go method SHALL be visible to the caller regardless of call style.

#### Scenario: Varcall echo preserves contents
- **WHEN** a GDScript caller passes a `Dictionary` to an echo method via varcall
- **THEN** the returned dictionary equals the input dictionary's contents and no reference is retained by the Go side after the call

#### Scenario: Arg-consume leaves no retained reference
- **WHEN** a GDScript caller passes a `Dictionary` to a method that reads it and returns a scalar derived from its contents, via varcall
- **THEN** the method observes the correct contents and no reference is retained after the call

#### Scenario: Mutation propagates to the caller
- **WHEN** a Go method adds a key to a `Dictionary` argument it received
- **THEN** the caller observes the new key in its original dictionary, because `Dictionary` shares one underlying container across all handles
