# Spec Delta

## Purpose

Governs ownership of Godot object references transferred to Go through generated object-return paths, so repeated object returns neither drift the engine reference count nor leak objects at shutdown, while non-owning (borrowing) wrappers remain safe to collect.

## ADDED Requirements

### Requirement: Engine-Transferred Return References Are Owned And Released Exactly Once
A generated method that returns a refcounted Godot object SHALL take ownership of the reference the engine transfers through the return slot, and SHALL release that reference exactly once — when the returned wrapper becomes unreachable (finalizer-backed) or when it is explicitly released.

#### Scenario: Repeated returns do not drift the reference count
- **WHEN** a caller invokes an object-returning method (e.g. `CollisionShape2D.GetShape()`) N times and drops each returned wrapper
- **THEN** the returned object's reference count is the same on iteration N as on iteration 1, with no per-call growth

#### Scenario: Dropped return wrapper releases the engine object
- **WHEN** a returned `Ref` wrapper is dropped while nothing else holds the object and the Go GC runs
- **THEN** the underlying Godot object's reference count drops to zero and the object is freed

### Requirement: Manual Unref Stays Idempotent With The Finalizer
Explicitly releasing an owned return wrapper SHALL clear its finalizer so a later garbage-collection pass performs no second release.

#### Scenario: Unref then GC does not double-release
- **WHEN** a caller invokes `Unref()` on an owned return wrapper and the GC later collects that wrapper
- **THEN** the underlying object is unreferenced only once and the reference count is not driven below its correct value

### Requirement: Borrowing Wrappers Never Release
Wrapping a refcounted object without taking ownership (the `NewRef` borrowing contract, as used by the generated `NewRefXAsRef` factories) SHALL NOT install a release finalizer, so collecting a borrower cannot decrement a reference held by another owner.

#### Scenario: Collected borrower does not over-release
- **WHEN** a borrowing wrapper around an object owned elsewhere becomes unreachable and the GC runs
- **THEN** the object's reference count is unchanged and the true owner's later `Unref()` still releases it correctly

### Requirement: Engine Exit Leak Check Stays Clean For Refcounted Returns
A workload that repeatedly calls object-returning methods and drops the results SHALL pass the engine's exit leak check.

#### Scenario: get_shape loop exits clean
- **WHEN** the demo runs a loop of 200 `CollisionShape2D.get_shape()` calls through Go and the engine shuts down
- **THEN** no `Leaked instance` and no leaked RID allocation errors are reported at exit
