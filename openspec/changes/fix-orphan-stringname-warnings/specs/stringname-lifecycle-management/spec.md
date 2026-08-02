## ADDED Requirements

### Requirement: Return Path Refcount Cleanup
StringName and NodePath values created on the Go side and returned to
Godot SHALL have their Go-side refcount contribution released after
encoding into Godot's output buffer, preventing refcount leaks that
manifest as orphan StringName warnings at process exit.

#### Scenario: Varcall return of Go-created StringName
- **WHEN** a Go method returns a StringName (created via
  `NewStringNameWithUtf8Chars` or equivalent) through the varcall return
  path
- **AND** the StringName is encoded into a Godot Variant via
  `GDExtensionVariantPtrFromReflectValue`
- **THEN** the Go-side StringName SHALL be destroyed after encoding,
  ensuring the Godot Variant holds the sole remaining refcount

#### Scenario: Ptrcall return of Go-created StringName
- **WHEN** a Go method returns a StringName through the ptrcall return
  path via `GDExtensionTypePtrFromReflectValue`
- **THEN** the StringName SHALL be copied into Godot's output buffer
  using the GDExtension copy constructor and the Go-side copy SHALL be
  destroyed, ensuring proper ownership transfer

### Requirement: Argument Decoding Borrow Semantics
StringName and NodePath values received as arguments from Godot SHALL use
borrow semantics: the Go side SHALL NOT acquire an owned refcount that
survives the method call.

#### Scenario: Varcall decode of a StringName argument
- **WHEN** a Godot varcall passes a StringName argument to a Go method
- **AND** the argument is decoded via `convertVariantToGoTypeReflectValue`
- **THEN** the decoded StringName SHALL be created via copy constructor
  then immediately destroyed after byte-copying, leaving a borrowed value
  with no Go-side refcount ownership

#### Scenario: Ptrcall decode of a StringName argument
- **WHEN** a Godot ptrcall passes a StringName argument to a Go method
  via `reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`
- **THEN** the StringName SHALL be decoded via byte-copy only, without
  calling the GDExtension copy constructor, preserving the Godot-side
  refcount unchanged
