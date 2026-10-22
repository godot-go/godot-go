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
class, or a non-inheriting object), and narrow the wrapper with a Go type
assertion, `wrapper.(InputEventKey)`. Direction decides what else is needed.
Generated wrappers embed their ancestor chain, so a wrapper standing for a
subclass already satisfies its parents' interfaces: widening to an ancestor, or
rejecting an unrelated class, is the assertion alone. Narrowing a wrapper in hand
does not work that way, because a wrapper returned by a generated method is typed
to the method's declared class — a `Node`-typed wrapper for a `Sprite2D` engine
object is not a `Sprite2D` and the assertion reports false. Re-resolve from the
engine object first with `ObjectFromInstanceBinding(wrapper.GetGodotObjectOwner())`,
which falls back to the callbacks registered for the object's runtime class, and
assert on that result. Callers that hold a `Variant` reach the same point through
`ObjectFromVariant`. Downstream code calling `builtin.ObjectCastTo` must migrate
the same way; the function is removed, not deprecated.
