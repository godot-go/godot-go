# Spec Delta

## Purpose

Governs how the binding layer hands one Variant to another place, and whether the
recipient owns a reference or merely borrows the source's storage. Ownership that
a caller has to infer by reading a function body is ownership that will eventually
be gotten wrong, and getting it wrong here means freeing memory still in use.

## ADDED Requirements

### Requirement: A Variant copy helper's ownership is readable from its name

Every exported helper that produces a Variant from another Variant, or from a
pointer to one, SHALL make the ownership of the result readable without reading
the helper's body. A helper returning an owned value SHALL read as a constructor;
a helper returning a non-owning alias SHALL name that it is a view.

#### Scenario: Caller chooses between an owned copy and a borrowed view

- **WHEN** a developer needs a Variant to pass along and finds two candidate
  helpers
- **THEN** the name alone distinguishes the one that returns a value the caller
  owns and must release from the one that returns a borrowed alias the caller must
  never release

#### Scenario: Owned and non-owning helpers are not confusable by prefix

- **WHEN** the two helpers are listed together in an editor's completion popup
- **THEN** neither is a near-variant of the other's name such that picking the
  wrong one is a plausible typo

### Requirement: Non-owning Variant views are never destroyed by the library

A non-owning Variant view SHALL NOT be destroyed on any path, including reject and
error paths. Destroying one frees storage the engine still owns and the source
Variant still points at.

#### Scenario: Varcall rejection allocates nothing to unwind

- **WHEN** a varcall is rejected after its arguments have been marshalled as
  borrowed views
- **THEN** the reject path releases no argument view, because none was ever owned

#### Scenario: Heap-backed Variant contents survive a borrowed view's lifetime

- **WHEN** a borrowed view of a heap-backed Variant (`String`, `Array`,
  `Dictionary`, or an object) goes out of scope
- **THEN** the source Variant remains valid and its contents intact

### Requirement: The ownership difference is pinned by an executable check

An automated check SHALL distinguish the owned copy from the non-owning view, so
that replacing one with the other fails at test time rather than corrupting memory
at runtime.

#### Scenario: Swapping the owned copy for the view is caught

- **WHEN** a call site that destroys its Variant is switched to the non-owning
  helper
- **THEN** the check fails before the change can ship

#### Scenario: Value Variants are exempt from the hazard

- **WHEN** a borrowed view is taken of a value Variant such as `bool`, `int`, or
  `Vector2`
- **THEN** the view is indistinguishable from a copy in behaviour, and the check
  records this as expected rather than as a pass that proves safety
