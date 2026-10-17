# Proposal

## Why

Godot 4 classes expose static methods (callable on the class without an
instance), and godot-cpp supports authoring them via `bind_static_method`.
godot-go declares them unsupported — `pkg/core/classdb.go` carries a
commented-out `ClassDBBindStaticMethod` stub with "TODO: golang does not
have static methods" — yet the Go answer to a static method is a plain
package function, and the existing registration machinery is already
class-name-keyed (instance-free). Filling this gap closes roadmap item B2a
and removes a long-standing divergence from the reference implementation.

## What Changes

- New public API `core.ClassDBBindStaticMethod[T GDClass](inst T, gdMethodName string, fn any, argNames []string, defaultValues []Variant)`
  binds a Go package function as a static method of `T`'s registered class,
  registered with `METHOD_FLAG_STATIC`.
- `GoMethodMetadata` gains a static mode: a receiver-free metadata
  constructor from a package `func`, and dispatch paths (`CallWithError`,
  `Ptrcall`, variadic `CallSlice`) that skip receiver injection for static
  binds.
- The varcall and ptrcall callbacks (`GoCallback_MethodBindMethodCall`,
  `GoCallback_MethodBindMethodPtrcall`) currently hard-panic on a NULL
  instance pointer. They gain a static branch: when the bind is static, a
  NULL instance is valid and dispatch proceeds without instance resolution;
  for non-static binds NULL remains fatal.
- Bind-time validation: reject non-callable `fn`, reject the
  static+virtual flag combination, reject duplicate method names via the
  existing per-class method map.
- Harness scenarios call the new statics from GDScript through the class
  path (`Example.<static>(...)`) covering typed (ptrcall) and dynamic
  (varcall) invocations, default arguments, and error reporting.
- `docs/overview.md` "Static Methods" section documents the new support;
  the "Static Variables" section is corrected to state it is blocked on a
  GDExtension/LanguageExtension ABI that does not exist. `docs/roadmap.md`
  marks B2 methods done and B2 variables blocked.

## Capabilities

### New Capabilities

- `static-method-binding`: binding Go package functions as Godot class
  static methods — registration with the static flag, NULL-instance
  dispatch through both call paths, receiver-free reflection metadata, and
  bind-time validation of static binds.

### Modified Capabilities

(none — instance-method binding, argument decoding, defaults, error
reporting, and ownership requirements are unchanged; static binds reuse
them as-is)

## Impact

- `pkg/core/classdb.go` — new `ClassDBBindStaticMethod` API and static bind
  path (function reflection instead of method-name resolution).
- `pkg/core/method_bind.go` / `method_bind_reflect.go` — static metadata
  construction and receiver-skipping dispatch.
- `pkg/core/method_bind_callback.go` — NULL-instance branches in the two
  call callbacks.
- `test/pkg/` + `test/demo/main.gd` — static fixture methods and harness
  scenarios.
- `docs/overview.md`, `docs/roadmap.md` — static methods/variables wording.

## Non-goals

- Calling engine-class static methods ergonomically from Go (the
  `classes.go.tmpl` "refactor for static instantiation" TODO). Generated
  instance wrappers already work against static engine functors because the
  engine ignores the object pointer for statics; a call-side refactor is
  ergonomics polish, not a capability gap.
- Static variables (roadmap B2b): no mechanism exists in the vendored
  `gdextension_interface.h`; deferred pending an experimental
  LanguageExtension ABI.
- Engine-side enforcement of GDScript static-call syntax; we bind and
  dispatch, the engine decides call legality.
