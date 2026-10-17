# Design

## Context

See `proposal.md` - Why for motivation; `specs/property-hint-registration/spec.md`
for the behavior contract.

Current state:

- `ClassDBAddProperty` (`pkg/core/classdb.go:78-168`) validates getter/setter
  arity, then builds a `GDExtensionPropertyInfo` with hardcoded
  `PROPERTY_HINT_NONE`, empty `String` hint, `PROPERTY_USAGE_DEFAULT`, and
  `class_name` = the owning class's `StringName`.
- `ffi.NewGDExtensionPropertyInfo` is pointer-based: the caller must pre-create
  Godot `StringName`/`String` objects, pin them, and keep them alive across the
  engine call. Its `// TODO: create StringName locally here` is the D2 item.
- `ClassDBAddSignal` (`classdb.go:175-244`) demonstrates the lifecycle pattern
  this change generalizes: create temporaries, pin, call, destroy after
  (Godot copies `PropertyInfo` pointer data during registration).
- `pkg/constant` already generates `PropertyHint` and `PropertyUsageFlags`
  enums. `PROPERTY_HINT_RESOURCE_TYPE = 17` exists in this API version; there
  is no `RESOURCE_TYPE_BASETYPE`.
- Import graph: `pkg/ffi` is a leaf that cannot import `pkg/builtin`, so the
  ffi constructor can never grow Go-native string parameters.
- Only two callers of `ClassDBAddProperty` exist (`test/pkg/example.go:1113`,
  `test/pkg/exampleref.go:45`).

## Goals / Non-Goals

**Goals:**

- One Go-native property-description type that carries hint/hint string/usage
  without exposing FFI pointer types to users.
- class-name derivation ported from godot-cpp, placed exactly once.
- Registration-path string temporaries owned entirely by `pkg/core`.

**Non-Goals:**

- No change to signal parameter metadata, method argument/return
  `PropertyInfo` (`method_bind.go` paths), or `NewGDExtensionPropertyInfoFromNames`
  as used by method binding.
- No codegen changes; `pkg/ffi/property_info.go` keeps its pointer-based
  constructor signature.

## Decisions

### D1: `PropertySpec` struct in `pkg/core`, new entry point `ClassDBAddPropertyInfo`

```go
type PropertySpec struct {
    VariantType GDExtensionVariantType
    Name        string                 // property name as seen by GDScript
    Setter      string                 // bound method name; "" = read-only
    Getter      string                 // bound method name; required
    Hint        constant.PropertyHint  // PROPERTY_HINT_NONE default (zero)
    HintString  string
    Usage       constant.PropertyUsageFlags // zero = PROPERTY_USAGE_DEFAULT
    ClassName   string                 // explicit class qualifier; "" = derive
}

func ClassDBAddPropertyInfo(inst GDClass, spec PropertySpec)
```

`ClassDBAddProperty` is re-implemented as a thin wrapper that fills a
`PropertySpec` with hint NONE, empty hint string, and `USAGE_DEFAULT`, so all
existing callers keep compiling and observing identical behavior.

Alternatives considered:

- *Extend `ClassDBAddProperty`'s signature*: breaking for zero ergonomic gain;
  a struct also gives future metadata (see A3 in roadmap) a home without
  another signature churn.
- *Functional options (`WithHint(...)`)*: more ceremony per call site, no
  structural value; the repo's registration APIs are plain-argument style
  (`SignalParam` is already a small struct - this follows that precedent).
- *Put `PropertySpec` in `pkg/ffi`*: impossible - ffi cannot import `builtin`
  (import cycle), and users shouldn't touch pointer-typed helpers.

### D2: class-name derivation lives in `ClassDBAddPropertyInfo`, before conversion

```go
className := spec.ClassName
if spec.Hint == PROPERTY_HINT_RESOURCE_TYPE {
    className = spec.HintString   // ported from godot-cpp / TODO.md
} else if className == "" {
    className = owningClassName
}
```

This is the placement `TODO.md` asks for; `TODO.md`'s entry is deleted with the
change. Godot-cpp assigns the hint string verbatim (it may itself be
comma-separated) - we match that exactly rather than parsing. The
`class_name` `StringName` is always created from this resolved name,
replacing today's unconditional owning-class pointer.

### D3: lifecycle - core-owned temporaries, `pnr` + `defer Destroy`

All three refcounted temporaries (`StringName` property name, `String` hint
string, `StringName` class name) are created inside
`ClassDBAddPropertyInfo`, pinned with the package-lifetime `pnr`, and
destroyed via `defer` after the engine call returns - the same convention as
`ClassDBAddPropertyGroup` and `ClassDBAddSignal`. Registration runs once at
class-init time, so the program-lifetime pinner is acceptable and consistent;
Godot copies the data during the call, so post-call `Destroy` is safe
(precedent and rationale comment already in `ClassDBAddSignal`).

The empty-string hint case constructs an empty Godot `String` exactly as
today (never a raw NULL `hint_string`).

### D4: validation - keep existing panics, add one

Getter-presence/arity and setter-arity panics stay as-is. One addition:
registering with `PROPERTY_HINT_RESOURCE_TYPE` but an empty `HintString`
panics at registration - it is an authoring fault the engine would otherwise
receive as an empty `class_name`, and the repo convention is fatal authoring
errors (see `docs/overview.md` Call Error Reporting).

### D5: `ffi.NewGDExtensionPropertyInfo` TODO removed by documentation

The constructor stays pointer-based (D1 rationale, import cycle). The
`// TODO: create StringName locally here` comment is replaced with an
ownership contract comment: callers pass valid pointers for the duration of
the consuming call; `pkg/core` owns creation for property registration. D2 is
closed because the practical burden (caller-managed strings) is eliminated at
the API users actually touch.

## Risks / Trade-offs

- [Engine copies vs retains `PropertyInfo` fields] - if the engine retained
  rather than copied any field, post-call `Destroy` would dangle. Mitigated by
  direct precedent: `ClassDBAddSignal` destroys immediately after registration
  and signal metadata reads back correctly in harness assertions; the new GDScript
  property-list assertions re-verify this for properties specifically.
- [Usage flags zero-value ambiguity] - `PropertyUsageFlags(0)` could mean
  "unset" or an explicit no-flags value. Decision: zero means DEFAULT
  (sentinel); explicit no-storage/no-editor combos are expressed by OR-ing
  real flags. Documented on the struct field.
- [`RESOURCE_TYPE` hint string is passed verbatim] - a multi-class
  comma-separated hint string becomes one odd `class_name` StringName, same as
  godot-cpp. Accepted for parity; noted for future alignment.
- [Breaking nothing but touching shared code] - `ClassDBAddProperty` is
  behavior-frozen via its existing two call sites; the harness runs them
  unchanged.

## Migration Plan

Additive change; no user migration. Rollback = revert the commit; no data or
generated-code state involved. `TODO.md`'s property-hint entry and
`docs/roadmap.md` B1/D2 markers are updated at completion.

## Open Questions

None that affect specs, approach, or tasks.
