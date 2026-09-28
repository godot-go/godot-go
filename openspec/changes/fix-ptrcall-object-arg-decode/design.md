# Design

## Context

`reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs` (`pkg/core/method_bind_reflect.go:440`) decodes the arguments of a Godot→Go call that arrived through ptrcall. Its `reflect.Interface` case (~line 572) reads:

```go
v := reflect.Zero(t)
inst := v.Interface()
switch inst.(type) {
case Object:
    // read runtime class name, look up GDNativeConstructors, build wrapper
default:
    if t.Implements(refType) { /* Ref branch */ }
    log.Panic("unsupported interface type", ...)
}
```

For an interface `t`, `reflect.Zero(t)` is the zero *interface* value; `.Interface()` returns a nil `interface{}`. A Go type switch matches a concrete case only against a non-nil dynamic type, so `case Object:` is unreachable for every interface parameter and control always lands in `default`. Only `Ref`-implementing parameters survive; everything else panics.

Measured directly, in isolation, on the types in question:

| Expression | Result |
|---|---|
| `reflect.Zero[Node]().Interface() == nil` | `true` |
| type switch of that value against `Object` | no match |
| `reflect.TypeOf[Node]().Implements(Object)` | `true` |

The logic inside the dead `case Object:` is sound and is what the fix needs: call `object_get_class_name` on the pointer, look the name up in `GDNativeConstructors`, build the wrapper from `(*GodotObject)(gdObjPtr)`. It has simply never run. The same file already uses the working dispatch form at ~1076 — `switch { case t.Implements(refType): … case t.Implements(gdClassType): … }` — so the broken shape is the anomaly rather than the convention.

Interface relationships were checked rather than assumed: `Object`-implementing and `Ref`implementing interfaces are disjoint (`Node`, `Shape2D`, `RefCounted` implement `Object` only; `Ref`, `RefShape2D` implement `Ref` only), so the branch order cannot misroute a type.

## Goals / Non-Goals

**Goals:**
- An object-typed interface parameter decodes correctly when the call arrives through ptrcall.
- Subclass instances satisfy base-class parameters, because resolution uses the argument's *runtime* class name.
- A null object argument decodes to the zero interface value without touching the null pointer.
- The `Ref` branch is untouched.

**Non-Goals:**
- The instance-binding shape and user-defined extension classes arriving as object arguments (`fix-user-defined-class-object-arg-decode`).
- `PropertyInfo.class_name` advertisement (`fix-object-argument-type-metadata`).
- Outbound Go→Godot encoding.

## Decisions

**D1 — Dispatch with `switch { case t.Implements(X): }`.** Replace the zero-value type switch with the form already used elsewhere in this file. `gdObjectType` is checked before `refType`; the disjointness above makes the order immaterial today, and checking the broader interface first matches the structure the dead code intended.

**D2 — Null is checked before any engine call.** The dead `case Object:` never had to consider null because it never ran. `object_get_class_name` through a null pointer is not a recoverable failure, so the branch tests the pointer first and assigns `reflect.Zero(t)` when null. This also makes `test_object_arg_nil_plain_object`-style calls safe once they route through ptrcall.

**D3 — Resolution stays keyed on the runtime class name, not the declared parameter name.** This is what makes subclass polymorphism work for free: a `CircleShape2D` passed to a `Shape2D` parameter resolves the `CircleShape2D` constructor, and the resulting concrete value satisfies `Shape2D`. Deriving from the declared name instead would break every subclass call.

**D4 — Do not attempt to make the type switch work.** Producing a non-nil value of an arbitrary interface type requires an instance, which is exactly what the decoder is trying to obtain. `Implements` on the `reflect.Type` is the idiomatic answer and needs no instance.

**D5 — The plain-object branch must dereference the argument cell; the dead code does not.** This is a second defect hiding behind the first. The ptrcall convention is that each `args[i]` points at the value, and for an object the value *is* the object pointer. The working outbound encoding proves the shape:

```go
var argObj_j GDExtensionObjectPtr
argObj_j = ObjectArgPtr(arg)                       // the object pointer
argPtrSlice[j] = (GDExtensionConstTypePtr)(unsafe.Pointer(&argObj_j))  // pointer to the cell
```

So the inbound decoder must read `*(*GDExtensionObjectPtr)(arg)`. The dead branch instead writes `gdObjPtr := (GDExtensionConstObjectPtr)(arg)` — the address of the cell, not the object. Making the branch reachable without fixing this would build a wrapper around a stack address and crash or corrupt, which is worse than the current loud panic. Both defects must be fixed together.

The `Ref` branch is correct as written and needs no dereference: the engine's `PtrToArg<Ref<T>>::encode` stores an actual `Ref<T>` in the cell, and `ref_get_object` takes a pointer to that container.

## Risks / Trade-offs

- **Assignability mismatch.** `args[i+1] = reflect.ValueOf(concrete)` requires the concrete class to implement the declared parameter interface. GDScript's static check should prevent a wrong-class call, but if one arrives, `Value.Call` panics with a reflect-level message. Mitigation: the failure is loud and names the types; a defensive `Implements` check with a clearer message is a candidate follow-up but is not added here, since inventing a second error path for a case the engine already rejects would be unverifiable.
- **A wrapper is allocated per call** by `constructor(owner)`. This matches the pre-existing intent of the dead branch and acquires no engine reference, so it neither leaks nor releases.
- **User-defined classes still fail.** `GDNativeConstructors` holds only engine classes, so a user-defined extension class arriving as a ptrcall object argument still hits "does not support gdextension class type". That is the separate binding-shape change and is deliberately out of scope here.

## Verification Note

This change is only observable end-to-end once typed GDScript object arguments parse, which requires `fix-object-argument-type-metadata`. To keep this change falsifiable on its own, the tests drive a typed object argument through ptrcall directly at the Go level rather than depending on the metadata fix.
