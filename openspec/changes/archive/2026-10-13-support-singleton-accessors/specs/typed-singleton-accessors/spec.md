# Spec Delta

## Purpose

Defines how engine singletons declared in the extension API are exposed to Go
code: generated typed accessor coverage, identity-stable caching, and
nil-for-missing behavior.

## ADDED Requirements

### Requirement: Every declared singleton has a generated typed accessor
The system SHALL expose, for each engine singleton entry in the extension API
metadata, a generated package-level accessor returning the singleton as its
corresponding generated class interface, with no per-entry handwritten code
and no filtering of editor or platform entries.

#### Scenario: Coverage mirrors the metadata
- **WHEN** bindings are generated against an API declaring N singleton entries
- **THEN** the generated surface contains exactly one typed accessor per
  entry, including editor-only and platform-specific ones

#### Scenario: Accessor reaches the live engine singleton
- **WHEN** the extension calls the accessor for a singleton the engine
  provides, such as `OS`
- **THEN** the returned value answers to that class's methods

#### Scenario: Metadata change flows on regeneration
- **WHEN** the metadata's singleton list gains or loses an entry and bindings
  are regenerated
- **THEN** the accessor set gains or loses exactly the matching accessor with
  no manual edits

### Requirement: Repeated calls return an identical wrapper
Each accessor SHALL resolve its singleton at most once per process and SHALL
return the same wrapper value on every call, so callers may rely on value
identity.

#### Scenario: Two sequential calls compare equal
- **WHEN** an accessor is called twice for the same singleton
- **THEN** both calls return equal values referring to the same wrapper

#### Scenario: Concurrent first callers share one wrapper
- **WHEN** several goroutines call an uninitialized accessor concurrently
- **THEN** every call returns the same wrapper value and no partially
  initialized value is observable

### Requirement: Missing singletons yield nil without caching absence
When the engine does not currently provide a singleton, the accessor SHALL
return a nil interface value that callers can check, and SHALL NOT cache the
absence: a later call re-attempts resolution and returns non-nil if the
singleton has since been registered.

#### Scenario: Editor-only singleton in a runtime build
- **WHEN** an accessor's singleton is absent from the running engine
- **THEN** the call returns nil and the caller's nil check succeeds without a
  panic

#### Scenario: Singleton registered later becomes visible
- **WHEN** a name resolved nil and the engine subsequently provides that
  singleton
- **THEN** a subsequent accessor call returns a non-nil wrapper
