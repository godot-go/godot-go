# Spec Delta

## Purpose

Defines the single mapping from Go types to Godot Variant types used when a bound
method's parameters and return value are advertised, decoded, and encoded, so
that binding, the ptrcall path, and the varcall path cannot disagree about what a
Go signature means.

## ADDED Requirements

### Requirement: Scalar and string Go types map one-to-one to engine variant types

A bound method signature SHALL advertise Go booleans as bool, every signed and
unsigned integer width as int, both float widths as float, and a native Go string
as a Godot string.

#### Scenario: Integer widths collapse to int

- **WHEN** a method takes an `int8`, an `int64`, a `uint16`, and a `uint`
- **THEN** all four parameters advertise the engine's int variant type

#### Scenario: Signedness does not change the variant type

- **WHEN** a method returns a `uint32`
- **THEN** its return advertises int and the value round-trips as an integer

#### Scenario: Native string advertises Godot string

- **WHEN** a method takes a Go `string`
- **THEN** the parameter advertises the string variant type

### Requirement: Godot value types map by their declared Go type, not their underlying bytes

Godot's fixed-size value types SHALL be recognised by their declared type —
including the string family, vectors, rectangles, transforms, plane, quaternion,
AABB, basis, colour, RID, callable, signal, dictionary, array, packed arrays, and
projection — and SHALL advertise the matching engine variant type even though
each is represented as a byte array.

#### Scenario: Vector2 is a vector, not a byte array

- **WHEN** a method takes a `Vector2`
- **THEN** the parameter advertises the vector2 variant type

#### Scenario: StringName is its own variant type

- **WHEN** a method returns a `StringName`
- **THEN** the return advertises the StringName variant type rather than string

#### Scenario: Packed array value keeps its packed type

- **WHEN** a method takes a `PackedVector3Array`
- **THEN** the parameter advertises the packed-vector3 variant type

### Requirement: Containers and objects map through their carrier type

A Go slice SHALL map to the packed array whose element type matches its element,
a bare Godot array or dictionary SHALL map to that container's variant type, and
object carriers — engine-class wrappers, generated class interfaces, and
reference interfaces — SHALL map to the object variant type. A nil type SHALL map
to nil.

#### Scenario: Native slice maps to its packed array

- **WHEN** a method takes a `[]float32`
- **THEN** the parameter advertises the packed-float32 variant type

#### Scenario: Slice of Variant is the vararg form

- **WHEN** a variadic method's trailing parameter is a `[]Variant`
- **THEN** the binding advertises the untyped-variant form instead of a packed array

#### Scenario: Reference interface is an object

- **WHEN** a method takes a generated `RefTexture2D` interface
- **THEN** the parameter advertises the object variant type

### Requirement: A Variant parameter or return is advertised as nil

A signature slot declared as Variant SHALL be advertised to the engine with the
nil type rather than the binding's internal untyped marker, which the engine
cannot name. A Variant return SHALL additionally carry the engine's
"nil is variant" usage flag; a Variant parameter advertises nil without that
flag.

#### Scenario: Variant argument appears as nil

- **WHEN** a method takes a Variant
- **THEN** the engine sees a nil-typed parameter
- **AND** the call still passes the caller's value through untyped

#### Scenario: Variant return appears as nil-with-variant

- **WHEN** a method returns a Variant
- **THEN** the engine sees a nil-typed return flagged as a variant
- **AND** reading the return yields the Variant the Go method produced

### Requirement: Kinds with no engine representation are rejected before registration

The binding SHALL fail a bind whose signature contains a Go kind with no engine
representation — a map, channel, function, `uintptr`, complex number, or unsafe
pointer — before engine registration.

#### Scenario: Map parameter

- **WHEN** a method takes a `map[string]int`
- **THEN** the bind fails before the engine registration call

#### Scenario: Function return

- **WHEN** a method returns a `func()`
- **THEN** the bind fails before the engine registration call

#### Scenario: Unsafe pointer

- **WHEN** a method takes an `unsafe.Pointer`
- **THEN** the bind fails before the engine registration call

### Requirement: Supported kinds with unmappable spellings are rejected

The binding SHALL fail a bind whose type cannot be mapped even though its kind is
otherwise supported: an unrecognised struct or interface, a concrete
reference-implementation struct or pointer, a slice whose element has no
packed-array counterpart, or a fixed array whose element is not a byte.

#### Scenario: Concrete reference implementation

- **WHEN** a method takes the generated concrete reference struct rather than its interface
- **THEN** the bind fails and names the supported reference spellings

#### Scenario: Slice with unmappable element

- **WHEN** a method takes a `[]bool`
- **THEN** the bind fails before the engine registration call

#### Scenario: Fixed array of non-byte elements

- **WHEN** a method takes a `[4]float32`
- **THEN** the bind fails before the engine registration call

#### Scenario: Struct that is neither a Godot value type nor a reference form

- **WHEN** a method takes an arbitrary user struct
- **THEN** the bind fails before the engine registration call

#### Scenario: Interface that is neither an object nor a reference

- **WHEN** a method takes an interface type with no class or reference backing
- **THEN** the bind fails before the engine registration call

#### Scenario: Pointer to Variant

- **WHEN** a method takes a pointer to Variant instead of a bare Variant
- **THEN** the bind fails before the engine registration call

### Requirement: Encoder coverage matches the mapping on both dispatch paths

For every Go type the mapping accepts, both the ptrcall encoder and the varcall
encoder SHALL produce a value of the advertised variant type, and a type the
mapping accepted but an encoder does not handle SHALL fail loudly instead of
writing an uninitialised output slot.

#### Scenario: Same type on both paths

- **WHEN** a method taking and returning a given mapped type is called once by ptrcall and once by varcall
- **THEN** both calls deliver and return the same value

#### Scenario: Unhandled value reaches an encoder

- **WHEN** an encoder receives a value its mapping accepted but has no case for
- **THEN** the call fails loudly rather than returning corrupted output
