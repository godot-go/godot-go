# Spec Delta

## ADDED Requirements

### Requirement: Generated object returns resolve to the runtime-class wrapper

A generated method that returns a non-reference-counted engine object SHALL
return the wrapper bound for the returned object's runtime class when one
resolves and satisfies the method's declared return type, not a new wrapper
typed to that declared class. A reference-counted return keeps its declared-class
owned wrapper, because that wrapper owns the reference the return slot transfers.

#### Scenario: Narrowing a returned object succeeds

- **WHEN** a generated method declares a base-class return such as `Node` and the engine returns an object whose runtime class is `Label`
- **THEN** the returned wrapper stands for the `Label` runtime class
- **AND** a Go type assertion of the returned value to `Label` succeeds
- **AND** the caller obtains it with that assertion alone, without calling a
  library helper or re-resolving the binding
- **AND** the method's declared return type is unchanged, so existing callers
  compile as written

#### Scenario: Repeated returns yield the same wrapper

- **WHEN** two generated calls return the same engine object
- **THEN** both calls return the identical Go wrapper value

#### Scenario: Widening still works on a returned object

- **WHEN** a generated method returns an object whose runtime class is a subclass of the declared return class
- **THEN** the returned wrapper satisfies the declared return type
- **AND** it also satisfies every ancestor class interface

#### Scenario: Null engine return stays a null wrapper

- **WHEN** the generated method's return slot holds a null object pointer
- **THEN** the returned wrapper carries no engine object and no engine binding
  lookup is issued

### Requirement: Unresolvable return binding falls back to the declared class

When no wrapper can be resolved through the instance binding for a returned
engine object — the class has no registered binding callbacks, or the engine
declines to create a binding — or the resolved wrapper does not satisfy the
declared return type, the generated method SHALL return a wrapper built over its
declared return class instead of failing the call.

#### Scenario: Class without binding callbacks still returns a usable wrapper

- **WHEN** a generated object return resolves to an engine object whose class has no registered binding callbacks
- **THEN** the method returns a wrapper over the declared return class
- **AND** the call does not panic or return a nil wrapper in place of a non-null object

#### Scenario: Wrapper that fails the declared type is not returned

- **WHEN** the instance binding resolves to a wrapper that is not an instance of the method's declared return class
- **THEN** the method returns a wrapper over the declared return class
- **AND** the caller never receives a wrapper that fails the declared return type

#### Scenario: Returned fallback wrapper still reaches the engine object

- **WHEN** a caller invokes an engine method through a fallback wrapper
- **THEN** the call reaches the object the return slot carried
