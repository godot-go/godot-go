## Why

The container-lifecycle-management work gave `Dictionary` arguments varcall owned-copy decode plus post-call release (`destroyOwnedContainerArgs`), but the demo test suite never passes a `Dictionary` argument into a Go method — the decode/release machinery and shared-reference mutation semantics are unexercised for this type (flagged in PR #140 review).

## What Changes

- Add echo and arg-consume example methods taking a `Dictionary` argument in `test/pkg/example.go`, registered via `ClassDBBindMethod`.
- Add GDScript assertions in `test/demo/main.gd`: contents round-trip correctly through both call styles where supported, consuming observes correct contents without retaining a reference, and mutating a received `Dictionary` propagates to the caller (shared-reference semantics, mirroring `test_mutate_array`).
- Determine whether ptrcall supports `Dictionary` arguments at all: there is currently no `Dictionary` case in the ptrcall decode (`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs`), so direct calls may be unsupported; document actual support and cover whichever call styles are valid.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

None. This change only adds test coverage for requirements that already exist in the `container-lifecycle-management` capability (argument decode does not retain references; round-trips are correct and leak-free); no normative behavior is introduced or altered.
