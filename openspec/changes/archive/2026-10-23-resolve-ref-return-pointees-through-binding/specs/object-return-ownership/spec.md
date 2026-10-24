# Spec Delta

## ADDED Requirements

### Requirement: Pointee resolution does not move the transferred reference count
Resolving the runtime-class wrapper for a reference-counted return SHALL NOT acquire or release an additional reference on the returned object. The binding lookup and wrapper construction are refcount-neutral: the `Ref` owns exactly the one reference the engine transferred through the return slot, regardless of which wrapper the `Ref` holds internally.

#### Scenario: Binding lookup on a cold cache does not bump the count
- **WHEN** a generated refcounted return resolves through a binding that does not yet exist and the create callback constructs a new wrapper for the object
- **THEN** the object's reference count after the return is exactly one more than before the call (the engine's transfer), with no additional increment from the binding creation

#### Scenario: Refcount stability across resolved returns
- **WHEN** a caller invokes a generated method returning the same refcounted object N times, dropping each `Ref` between iterations
- **THEN** the object's reference count at iteration N equals the count at iteration 1, with no per-call drift whether or not the pointee resolves through the binding
