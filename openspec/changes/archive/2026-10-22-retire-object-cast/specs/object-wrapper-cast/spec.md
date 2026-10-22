# Spec Delta

## REMOVED Requirements

### Requirement: Casting uses the engine's class tag and resolves through the instance binding

**Reason**: The cast entry point had no caller inside the binding substrate. The
casts the library actually performs run in the argument and Ref decode paths,
which compose the class-tag inheritance check with instance-binding resolution so
they can return a typed error naming both the declared and the actual class
(`pkg/core/method_bind_reflect.go`). A nil-returning, stringly-typed convenience
wrapper cannot say why a cast failed, and its only callers were the demo witness
written for it and one demo input handler.

**Migration**: Use the two behaviours the cast composed. Ask
`ObjectInheritsClass(objectPtr, "ClassName")` whether the object is an instance of
the class or one of its subclasses (false for a null object, empty name, unknown
class, or a non-inheriting object); resolve the wrapper with
`ObjectFromInstanceBinding` (or reuse the wrapper already in hand) and narrow it
with a Go type assertion, `wrapper.(InputEventKey)`. Generated wrappers embed
their ancestor chain, so a wrapper whose engine object is a subclass satisfies the
parent's interface while an unrelated class fails the assertion. Callers that hold
a `Variant` reach the same point through `ObjectFromVariant`. Downstream code
calling `builtin.ObjectCastTo` must migrate the same way; the function is removed,
not deprecated.
