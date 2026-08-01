## Purpose

Provides a type-safe, reference-counted smart pointer (`Ref[T]`) for Godot `RefCounted` objects in Go, with correct refcount management across the Go-to-Godot boundary, including finalizer-backed release and Variant/ptr call encoding.

## Requirements

### Requirement: Type-safe Ref pointer
The system SHALL provide a `Ref[T]` smart pointer that stores the concrete type `T` and returns it from `Ptr()` without a runtime type assertion. A `Ref` interface SHALL expose the refcount protocol (`ToObject`, `Ref`, `Unref`, `IsValid`) without requiring `Ptr`.

#### Scenario: Concrete access without assertion
- **WHEN** a `RefImage` is created from an `Image` object and `Ptr()` is called
- **THEN** it returns the `Image` directly and does not panic on a type mismatch.

#### Scenario: Nil detection
- **WHEN** `IsValid()` is called on an empty or released `Ref`
- **THEN** it returns `false`.

### Requirement: RefCounted reference lifecycle
The system SHALL manage the Godot reference count across create, copy, and release: creating a `Ref` for a user-owned object calls `InitRef()`, copying a `Ref` calls `Reference()`, and releasing calls `Unreference()` exactly once per acquired reference. A released `Ref` SHALL be a safe no-op if released again.

#### Scenario: Create and destroy
- **WHEN** a new object is wrapped with the init constructor and the ref is released
- **THEN** `GetReferenceCount()` returns `0` and the object is freed.

#### Scenario: Copy balances with release
- **WHEN** a `Ref` is copied and both copies are released
- **THEN** the reference count returns to zero and the object is freed exactly once.

#### Scenario: Double release is safe
- **WHEN** `Unref()` is called twice on the same `Ref`
- **THEN** the second call is a no-op and does not corrupt the reference count.

### Requirement: Release of Go-held references
The system SHALL release a Go-held reference when a `Ref` becomes unreachable (finalizer-backed), so a dropped `Ref` does not leak the underlying Godot object.

#### Scenario: Dropped Ref is collected
- **WHEN** a `Ref` is created, dropped without an explicit `Unref()`, and the Go GC runs
- **THEN** the underlying reference count is decremented and the object is eventually freed.

#### Scenario: Explicit Unref wins over finalizer
- **WHEN** a `Ref` is explicitly released and later collected by the GC
- **THEN** the finalizer does not release again.

### Requirement: Ref return encoding to Godot
The system SHALL encode a Go method's `Ref` return value into Godot for both the ptr call and Variant call paths without panicking, passing the wrapped object to Godot so its own `Ref` wrapping works.

#### Scenario: Ptr call return
- **WHEN** a Go method returning `RefImage` is invoked via ptr call
- **THEN** Godot receives an OBJECT-typed return referencing the same object and the method does not panic.

#### Scenario: Variant call return
- **WHEN** a Go method returning `RefImage` is invoked via Variant call
- **THEN** Godot receives a `Variant` of type `OBJECT` referencing the same object.

#### Scenario: Nil Ref return
- **WHEN** a Go method returns a nil/invalid `Ref`
- **THEN** Godot receives a valid null return value and the method does not panic.

### Requirement: Ref argument decoding
The system SHALL decode Godot `Ref` arguments into Go without adding a reference (transfer semantics), so repeated calls do not leak.

#### Scenario: Repeated ref arguments
- **WHEN** a Go method receives a `Ref` argument many times and never retains it
- **THEN** the underlying object's reference count does not grow across calls.

### Requirement: User-defined RefCounted classes
The system SHALL support user-defined GDExtension classes that extend `RefCounted` (e.g. a custom `Resource`), with correct refcount behavior during Godot-side `Ref` copy and destroy, without relying on extension reference callbacks.

#### Scenario: Godot-side Ref copy of a user class
- **WHEN** Godot copies a `Ref` to a user-defined Go `Resource`
- **THEN** the object's reference count increments, and decrements when destroyed, with no double-counting.
