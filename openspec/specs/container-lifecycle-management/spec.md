## Purpose

Defines the refcount ownership rules for container built-in types (`Array`, `Dictionary`, and the `Packed*Array` types) as they cross the Godot↔Go boundary, so that argument decoding does not leak references and ptrcall returns hand Godot an owned value.

## Requirements

### Requirement: Container Argument Decoding Does Not Retain References
Container values received as arguments from Godot are valid only for the duration of the call; retaining one beyond the call requires an explicit copy. The Go side SHALL NOT hold a refcount that outlives the call: a ptrcall container argument is decoded as a borrow (no Go-side refcount), and a varcall container argument is decoded as an owned copy that is released after the call.

#### Scenario: Ptrcall decode of a container argument
- **WHEN** a Godot ptrcall passes a container argument (`Array` or a `Packed*Array`) to a Go method
- **THEN** the container is decoded by byte-copy without invoking a GDExtension copy constructor, leaving the Godot-side refcount unchanged

#### Scenario: Varcall decode of a container argument
- **WHEN** a Godot varcall passes a container argument (`Array`, `Dictionary`, or a `Packed*Array`) to a Go method
- **THEN** the container is produced via the type-from-variant constructor as an owned copy that stays alive for the duration of the call (correct whether the supplied Variant is the same container type or a convertible one, e.g. `Array[int]` → `PackedInt64Array`), and the copy is released only after the return value has been encoded, so no reference is retained by the Go side after the call

### Requirement: Shared-Reference Container Arguments Propagate Mutations To The Caller
`Array` and `Dictionary` are shared-reference containers without copy-on-write: every handle (borrowed or owned) wraps the same underlying container. A Go method that mutates such an argument therefore mutates the GDScript caller's container regardless of call style. Callers and callees SHALL treat received `Array`/`Dictionary` arguments as aliased with the caller's data, not as private copies.

#### Scenario: Mutation of a borrowed argument propagates to the caller
- **WHEN** a Go method mutates an `Array` or `Dictionary` argument it received via ptrcall
- **THEN** the caller observes the mutated contents in its original container

#### Scenario: Varcall mutations propagate for shared-reference types only
- **WHEN** a Go method appends to an `Array` received via varcall
- **THEN** the caller's array changes too, because `Array` shares one underlying container across all handles

### Requirement: Borrowed Packed Array Arguments Are Read-Only
A ptrcall packed-array borrow (`Packed*Array`) holds no `CowData` refcount, so Godot considers the caller's buffer exclusively owned: any mutating operation that grows or reallocates the buffer frees storage the caller still references. Go methods SHALL NOT mutate a packed-array argument received via borrow; retaining one beyond the call also requires an explicit copy.

#### Scenario: Mutating a borrowed packed array is unsupported
- **WHEN** a Go method appends to a packed-array argument it received via ptrcall
- **THEN** the behavior is undefined and may corrupt or free the caller's buffer; conforming methods must copy before mutating

#### Scenario: Varcall owned decode isolates packed array mutations
- **WHEN** a GDScript caller passes a `Packed*Array` to a Go method via varcall and the method appends to it
- **THEN** the owned decode's refcount triggers copy-on-write on first write and the caller's original packed array is unchanged

### Requirement: Container Ptrcall Return Transfers Ownership
Container values returned from Go to Godot through the ptrcall path SHALL be written into Godot's output buffer via the GDExtension copy constructor, so Godot receives an owned value. The Go-side source value SHALL be destroyed only when it owns a refcount (i.e. it is not a borrowed argument decoded for this call); destroying a borrow would release a reference Godot's caller still holds.

#### Scenario: Ptrcall return of a Go container
- **WHEN** a Go method returns a container through the ptrcall return path
- **THEN** the value is copied into Godot's output buffer via the copy constructor and the Go-side copy destroyed, leaving Godot the sole owner

#### Scenario: Echoing a borrowed container argument via ptrcall
- **WHEN** a Go method returns the same container argument it received via ptrcall
- **THEN** Godot's output buffer receives a fresh owned copy and the round-trip neither leaks nor double-frees a reference

#### Scenario: Byte-identical return classified as a borrow echo
- **WHEN** a Go method returns a container that compares equal by content to a borrowed ptrcall argument — including an unmutated defensive clone of that argument, which differs from the borrow only in refcount, not bytes — because borrow-echo detection is content-based (`reflect.DeepEqual`) and cannot distinguish the two cases
- **THEN** the return is classified as a borrow echo and its source value is not destroyed, so an owned clone misclassified this way silently retains one reference per call (bounded leak, no crash)

### Requirement: Container Round-Trips Are Correct And Leak-Free
Container arguments and return values SHALL round-trip with correct contents in both call styles, and consuming a container argument SHALL NOT leave a Go-side reference retained after the call.

#### Scenario: Echo round-trip preserves contents
- **WHEN** a GDScript caller passes a container to an echo method via ptrcall and via varcall
- **THEN** the returned container equals the input container's contents

#### Scenario: Arg-consume method leaves no retained reference
- **WHEN** a GDScript caller passes a container to a method that reads but does not return it, via ptrcall and via varcall
- **THEN** the method observes the correct contents and no container reference is retained by the Go side after the call

### Requirement: Dictionary Arguments Round-Trip Without Retention
A `Dictionary` received as an argument SHALL round-trip with correct contents through every supported call style, and consuming one SHALL NOT leave a Go-side reference retained after the call. Because `Dictionary` is a shared-reference container without copy-on-write, a mutation made by the Go method SHALL be visible to the caller regardless of call style.

#### Scenario: Varcall echo preserves contents
- **WHEN** a GDScript caller passes a `Dictionary` to an echo method via varcall
- **THEN** the returned dictionary equals the input dictionary's contents and no reference is retained by the Go side after the call

#### Scenario: Arg-consume leaves no retained reference
- **WHEN** a GDScript caller passes a `Dictionary` to a method that reads it and returns a scalar derived from its contents, via varcall
- **THEN** the method observes the correct contents and no reference is retained after the call

#### Scenario: Mutation propagates to the caller
- **WHEN** a Go method adds a key to a `Dictionary` argument it received
- **THEN** the caller observes the new key in its original dictionary, because `Dictionary` shares one underlying container across all handles
