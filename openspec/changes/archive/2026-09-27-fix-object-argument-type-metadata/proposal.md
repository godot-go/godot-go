# Proposal

## Why

`NewGoMethodMetadata` (`pkg/core/method_bind.go`) builds every argument's `PropertyInfo` with the **owning class** as `class_name`:

```go
argPropClassNameStringNames[i] = NewStringNameWithLatin1Chars(className) // the registering class
```

For arguments whose variant type is OBJECT, Godot uses that `class_name` to statically type the parameter. A method bound on `Example` that takes a `Node` therefore advertises the parameter as class `Example`. GDScript static typing rejects every correctly-typed caller:

```
SCRIPT ERROR: Parse Error: Invalid argument for "test_object_arg_ref_round_trip()" function:
  argument 1 should be "Example" but is "CollisionShape2D".
```

Today this is masked only because the demo declares object variables untyped (`var cs = CollisionShape2D.new()`); the workaround comment in `test/demo/main.gd` (~line 558) records that the metadata bug was deferred while landing the object-argument encoding fix (commit e727e39). The object-typed **return** `PropertyInfo` has the identical defect (`returnPropClassNameStringName = NewStringNameWithLatin1Chars(className)`), so a method on `Example` returning `Node` advertises a return class of `Example`.

## What Changes

- In `NewGoMethodMetadata`, object-typed (OBJECT variant) arguments advertise the argument's **own** Godot class: plain engine interface `Node` → `"Node"`; smart-pointer interface `RefShape2D` → `"Shape2D"`.
- The class name is resolved from the Go argument type via the authoritative registries (`GDNativeConstructors`, `GDClassRefConstructors`, `Internal.GDRegisteredGDClasses`) — the same lookup style `ReflectTypeToGDExtensionVariantType` already uses — not by unchecked string-sniffing. An OBJECT-typed argument that resolves to no registered class fails loudly at bind time.
- Object-typed returns are fixed symmetrically via the same resolution.
- Value-type (non-object) arguments and returns keep their current metadata unchanged.
- Remove the untyped-variable workaround in `test/demo/main.gd`; declare object test variables with static types and add typed round-trip coverage.

No public bind-API signature changes (`ClassDBBindMethod`/`ClassDBBindMethodVirtual` keep their shape), and no codegen changes: the fix lives entirely in the runtime binder.

## Capabilities

### New Capabilities

- `method-argument-type-metadata`: How a Go-bound method's object-typed parameters and return advertise their Godot class in `PropertyInfo.class_name`, so GDScript static typing accepts correct classes and subclasses and rejects wrong ones, while value-type metadata stays as-is.

### Modified Capabilities

(none)

## Non-goals

- **Argument/return encoding across the boundary** — governed by `object-argument-encoding` (commit e727e39); untouched here.
- **The `name` field of argument `PropertyInfo`** (currently the Go type name rather than the bound argument name) — a separate cosmetic concern.
- **Editing `*.gen.*` files** — none are affected; `make generate` is run only to confirm no drift.
- **Vararg bindings** — registered with zero named arguments; unaffected.
