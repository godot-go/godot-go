# native-call-pinning Delta

## MODIFIED Requirements

### Requirement: Values That Must Outlive The Call Remain Valid
A Go value that C retains a reference to beyond the call that created it SHALL be pinned on a pinner whose lifetime matches that retention, and SHALL NOT be unpinned while any C-side reference remains. A value whose C-side reachability is provided entirely by a releasable handle — an `cgo.Handle` the engine returns to this extension on callback — SHALL NOT be pinned at all; the handle's release is the sole lifetime anchor.

#### Scenario: Method metadata survives unrelated per-call unpins
- **WHEN** a method is registered, pinning its `GoMethodMetadata` argument and return `PropertyInfo` `StringName` and hint `String` objects, and many unrelated generated calls subsequently run and unpin their own scratch
- **THEN** any later call or Godot reflection that reads those names observes valid, uncorrupted data

#### Scenario: Returned object stays usable after the call's pinner is released
- **WHEN** a generated method returns a reference-counted object to its Go caller and the call's per-call pinner has already run `Unpin()`
- **THEN** the caller can still use the returned object — read its properties and pass it into further Godot calls — with no corruption

#### Scenario: Handle-backed instance payload is not pinned
- **WHEN** a user-defined extension class instance is bound to its Godot object via a `cgo.Handle` stored as the extension instance
- **THEN** the wrapped Go values SHALL stay reachable through the handle alone, with no entry on any pinner
- **AND** releasing the handle at instance-free time makes the wrapped graph collectable
