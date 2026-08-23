## Purpose

Defines the refcount ownership rules for container built-in types (`Array`, `Dictionary`, and the `Packed*Array` types) as they cross the Godot↔Go boundary, so that argument decoding does not leak references and ptrcall returns hand Godot an owned value.

## ADDED Requirements

### Requirement: Container Argument Decoding Does Not Retain References
Container values received as arguments from Godot are valid only for the duration of the call; retaining one beyond the call requires an explicit copy. The Go side SHALL NOT hold a refcount that outlives the call: a ptrcall container argument is decoded as a borrow (no Go-side refcount), and a varcall container argument is decoded as an owned copy that is released after the call.

#### Scenario: Ptrcall decode of a container argument
- **WHEN** a Godot ptrcall passes a container argument (`Array` or a `Packed*Array`) to a Go method
- **THEN** the container is decoded by byte-copy without invoking a GDExtension copy constructor, leaving the Godot-side refcount unchanged

#### Scenario: Varcall decode of a container argument
- **WHEN** a Godot varcall passes a container argument (`Array`, `Dictionary`, or a `Packed*Array`) to a Go method
- **THEN** the container is produced via the type-from-variant constructor as an owned copy that stays alive for the duration of the call (correct whether the supplied Variant is the same container type or a convertible one, e.g. `Array[int]` → `PackedInt64Array`), and the copy is released only after the return value has been encoded, so no reference is retained by the Go side after the call

### Requirement: Borrowed Ptrcall Container Arguments Share Storage With The Caller
A ptrcall container borrow does not increment the refcount, so Godot's copy-on-write cannot isolate subsequent writes: a Go method that mutates a borrowed container argument writes through to the caller's underlying storage (previously the owned-copy decode triggered copy-on-write on first mutation). Callers and callees SHALL treat ptrcall-received containers as aliased with the caller's data, not as private copies. Note that this aliasing applies to varcall as well for `Array` and `Dictionary`: those types are shared-reference containers without copy-on-write, so mutating even an owned decoded copy modifies the caller's data; only the copy-on-write `Packed*Array` types are isolated by varcall's owned decode.

#### Scenario: Mutation of a borrowed argument propagates to the caller
- **WHEN** a Go method mutates a container argument it received via ptrcall
- **THEN** the mutation is applied to storage shared with the GDScript caller and the caller observes the mutated contents in its original container

#### Scenario: Varcall mutation isolation holds only for copy-on-write types
- **WHEN** a GDScript caller passes a `Packed*Array` to a Go method via varcall and the method appends to it
- **THEN** the owned decode's refcount triggers copy-on-write on first write, the caller's original packed array is unchanged; when the same experiment is performed with an `Array`, the caller's array does change, because `Array` shares one underlying container across all handles

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
