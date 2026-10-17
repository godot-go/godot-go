# string-boundary-marshalling Specification

## Purpose

Defines how string-shaped values - native Go `string`, and the Godot string
types `String`, `StringName`, and `NodePath` - cross the boundary of Go
methods bound to GDScript: which signature types are supported, which
runtime variant types they accept and how mismatches convert, who owns each
decoded value, and how a borrowed value is retained past the call.

## Requirements

### Requirement: Native Go Strings Are First-Class At Bound Methods
A bound Go method SHALL be able to declare `string` for parameters and
return values. Registration SHALL declare the Godot `STRING` variant type
for those positions, and a round trip SHALL preserve the exact text in both
direct and generic (varargs) call styles.

#### Scenario: String parameter round trip in both call styles
- **WHEN** GDScript calls a bound Go method with a `string` parameter
  directly and via the generic call interface with the text `"scalar"`
- **THEN** the Go method observes `"scalar"` in both calls

#### Scenario: Go string return is caller-owned
- **WHEN** a bound Go method returns a native `string`
- **THEN** the GDScript caller receives the exact text
- **AND** the Go side releases its temporary after encoding, so repeated
  calls produce no orphaned-reference or leak-gate failures

### Requirement: String Parameters Accept Convertible Variants
When a bound method declares a `string` parameter and GDScript supplies a
variant that is not `STRING`, the call layer SHALL convert it where the
engine's strict conversion rules allow: text-shaped variants
(`STRING_NAME`, `NODE_PATH`) yield their content as text. A variant with no
such conversion SHALL be rejected as an invalid call argument before the Go
method runs; the method SHALL observe nothing from a rejected call.

#### Scenario: StringName argument reaches a string parameter
- **WHEN** GDScript passes a `StringName` value to a bound Go method
  parameter declared as native `string`, directly and via generic call
- **THEN** the Go method observes the name's text

#### Scenario: NodePath argument reaches a string parameter
- **WHEN** GDScript passes a `NodePath` value to a bound Go method parameter
  declared as native `string`
- **THEN** the Go method observes the path's text

#### Scenario: Non-convertible variant is rejected before the method runs
- **WHEN** GDScript passes an `int` value to a bound Go method parameter
  declared as native `string` via generic call
- **THEN** the call is rejected as an invalid argument
- **AND** the Go method body does not execute, without crashing the process
  or silently abandoning the calling script's remaining assertions

### Requirement: String Argument Retention Is Regime-Defined
The ownership regime of each decoded string-shaped argument SHALL be stable
and observable: values received as native `string` are Go-owned copies safe
to retain; `StringName` and `NodePath` arguments are call-scoped borrows
that SHALL NOT be retained past the call; `String` arguments are released by
the call machinery after the call and SHALL NOT be destroyed or retained by
the method.

#### Scenario: Retaining a native string is safe
- **WHEN** a bound Go method stores its `string` parameter on receiver state
  and reads it after the call completes
- **THEN** the stored text remains valid

#### Scenario: Borrowed arguments are never released by the machinery
- **WHEN** a bound Go method receives `StringName` or `NodePath` arguments
  in either call style
- **THEN** no post-call step of this binding destroys those argument values
  and repeated calls neither leak references to the Godot caller nor
  release references the Godot caller still holds

### Requirement: Owned-Copy Helpers Retain Borrowed String Values
`String`, `StringName`, and `NodePath` SHALL each provide a `Copy()` method
returning a new value of the same type holding an independent reference to
identical content, which the caller owns and SHALL release with `Destroy()`.
Copying SHALL be valid for both owned and borrowed receivers, and the copy
SHALL remain usable after the call that supplied a borrowed receiver ends.

#### Scenario: Copy of a borrowed StringName survives the call
- **WHEN** a bound Go method receives a borrowed `StringName` argument,
  calls `Copy()` on it, stores the result, and later the call ends
- **THEN** reading the stored copy yields the original text
- **AND** `Destroy()`-ing the stored copy leaves no orphaned-reference and
  no double-free symptoms in test output

#### Scenario: Copy preserves content for all three types
- **WHEN** `Copy()` is called on each of a `String`, a `StringName`, and a
  `NodePath` value and the copy's text is compared to the source's
- **THEN** the texts are identical for all three types
