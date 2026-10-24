# Spec Delta

## MODIFIED Requirements

### Requirement: Generated object returns resolve to the runtime-class wrapper
A generated method that returns an engine object SHALL resolve the returned owner through the instance binding and present the wrapper for the object's runtime class when one resolves and satisfies the method's declared return type. A non-reference-counted return hands that wrapper to the caller. A reference-counted return stores it as the pointee inside the declared `Ref`; the `Ref` interface type is unchanged.

#### Scenario: Narrowing a returned object succeeds
- **WHEN** a generated method declares a base-class return such as `Node` and the engine returns an object whose runtime class is `Label`
- **THEN** the returned wrapper stands for the `Label` runtime class
- **AND** a Go type assertion of the returned value to `Label` succeeds
- **AND** the caller obtains it with that assertion alone, without calling a library helper or re-resolving the binding
- **AND** the method's declared return type is unchanged, so existing callers compile as written

#### Scenario: Repeated returns yield the same wrapper
- **WHEN** two generated non-refcounted calls return the same engine object
- **THEN** both calls return the identical Go wrapper value

#### Scenario: Widening still works on a returned object
- **WHEN** a generated method returns an object whose runtime class is a subclass of the declared return class
- **THEN** the returned wrapper satisfies the declared return type
- **AND** it also satisfies every ancestor class interface

#### Scenario: Null engine return stays a null wrapper
- **WHEN** the generated method's return slot holds a null object pointer
- **THEN** the returned wrapper carries no engine object and no engine binding lookup is issued

#### Scenario: Narrowing a refcounted return's pointee succeeds
- **WHEN** a generated method declares `RefShape2D` and the engine returns an object whose runtime class is `CircleShape2D`
- **THEN** `Ptr()` on the returned `Ref` yields a value that satisfies the `CircleShape2D` interface
- **AND** the `Ref` type itself is `RefShape2D`, unchanged from the method signature

#### Scenario: Repeated refcounted returns agree at the pointee level
- **WHEN** two calls to the same generated method returning a refcounted object retrieve the same engine object
- **THEN** `Ptr()` on both returned `Ref` values yields the identical Go wrapper value
- **AND** each `Ref` remains a distinct owning value
