# Overview

This will be a living doc which will provide an overview of key concepts in the godot-go bindings.

# Binding Substrate Contracts

Six capability specs own the substrate every bind path is built on. Each row
names the contract, the spec that pins it, and the public entry points it
governs. The specs live under `openspec/specs/`.

| Capability | Spec | Public entry points |
| --- | --- | --- |
| Extension bootstrap | [`extension-bootstrap`](../openspec/specs/extension-bootstrap/spec.md) | `core.NewInitObject`, `InitObject.Init`, `Register{Core,Server,Scene,Editor}Initializer`, `Register{Core,Server,Scene,Editor}Terminator`, `RegisterStartupCallback`, `RegisterFrameCallback`, `RegisterShutdownCallback` |
| Class registration | [`class-registration`](../openspec/specs/class-registration/spec.md) | `core.ClassDBRegisterClass`, `core.ClassDBUnregisterClass`, `core.Internal.GDRegisteredGDClasses`, `core.Internal.GDClassConstructors` |
| Variant type mapping | [`variant-type-mapping`](../openspec/specs/variant-type-mapping/spec.md) | `core.ReflectTypeToGDExtensionVariantType`, the `ArgumentEncoder` registered for each class |
| Property access dispatch | [`property-access-dispatch`](../openspec/specs/property-access-dispatch/spec.md) | The `_get`/`_set`/`_property_can_revert`/`_property_get_revert` virtuals bound with `core.ClassDBBindMethodVirtual`, and the validator passed to `core.ClassDBRegisterClass` |
| Vararg method binding | [`vararg-method-binding`](../openspec/specs/vararg-method-binding/spec.md) | `core.ClassDBBindMethodVarargs`, `core.ClassDBBindStaticMethod` (variadic statics) |
| Object wrapper and cast | [`object-wrapper-cast`](../openspec/specs/object-wrapper-cast/spec.md) | `builtin.WrappedImpl`, `builtin.NewWrappedFromGodotObject`, `builtin.ObjectCastTo`, `builtin.ObjectInheritsClass`, `builtin.ObjectArgPtr` |

# GDScript Language Feature Mapping

There are a lot of language features supported by GDScript that does not map cleanly to Go.

## Default Parameter Value

Default parameters are supported. Register methods with `ClassDBBindMethod` passing a `defaultValues []Variant` slice; GDScript callers may omit trailing arguments.

## Class Inheritance

Go does not support classical class inheritance. Instead, composition with struct embedding is used in its place. Lets take a look at the following example user-defined class:

```go
type PlayerCharacter struct {
	CharacterBody2DImpl
}

// interface test evidence
var _ CharacterBody2D = &PlayerCharacter{}
```

The user-defined `PlayerCharacter` class extends the `CharacterBody2D` interface by embedding the `CharacterBody2DImpl` struct. Following the definition of `CharacterBody2DImpl` we see the following definition:

```go
type CharacterBody2DImpl struct {
	PhysicsBody2DImpl
}
```

We see that `CharacterBody2DImpl` embeds the `PhysicsBody2DImpl` struct, which implements the `PhysicsBody2D` interface.

## Virtual Methods

Go does not natively support virtual functions or struct methods. Instead, a method naming convention is implemented: qualified methods named `V_<ClassName>_<MethodName>` are registered as Godot virtual methods.

```go
func (e *Example) V_Example_Ready() { ... }

...

// register the function with Godot
ClassDBBindMethodVirtual(t, "V_Example_Ready", "_ready", nil, nil)
```

* `V_` denotes that this is a virtual function.
* `Example` is the Go type whose embedding level implements the virtual.
* `Ready` matches the `_ready` gdscript method (the leading underscore is dropped and the rest is PascalCased).

### Qualified names are required

The deprecated flat `V_<MethodName>` form (`V_Ready`) panics at registration time with a diagnostic naming the expected qualified shape. Because Go struct embedding promotes methods across the whole chain, flat names allowed only one implementation slot per virtual: shadowing made base implementations unreachable, and diamond embedding made selectors ambiguous.

### Most-derived-wins resolution

`ClassDBBindMethodVirtual` resolves the implementation by walking the registering type's embedding chain, most-derived type first, looking for `V_<LevelTypeName>_<MethodName>`. The first match binds:

```go
type PlayerCharacter struct {
	ControlImpl
	CharacterBody2DBase // declares V_CharacterBody2DBase_GetMaximumSize
}

func (p *PlayerCharacter) V_PlayerCharacter_GetMaximumSize() Vector2 { ... }
```

An instance of `PlayerCharacter` dispatches to `V_PlayerCharacter_GetMaximumSize`; an instance of the base level dispatches to `V_CharacterBody2DBase_GetMaximumSize`.

### Explicit delegation replaces super-calls

Because each level's implementation has a distinct name, a derived implementation can extend rather than replace the base behavior by calling it through the promoted selector:

```go
func (p *PlayerCharacter) V_PlayerCharacter_GetMaximumSize() Vector2 {
	base := p.CharacterBody2DBase.V_CharacterBody2DBase_GetMaximumSize()
	return base.Add(Vector2{X: 100})
}
```

Virtuals are invoked through the GDExtension `get_virtual_call_data2` / `call_virtual_with_data` path. Virtuals that no level along the chain implements return `nil` call data so Godot falls back to its engine default; implemented virtuals are dispatched to the resolved Go method.

### Generated virtual surface catalog

`extension_api.json` declares 1437 virtual methods across 106 classes. Code generation emits a declaration-only interface for each class with Category-A virtuals (non-void return and a plain wrapper on the same class hierarchy — the subset whose result the engine consumes through that wrapper):

```go
// pkg/gdclassimpl/virtuals.gen.go
type ControlVirtuals interface {
	V_Control_GetMaximumSize() Vector2
	V_Control_GetMinimumSize() Vector2
	V_Control_GetTooltip(at_position Vector2) String
	V_Control_GetCursorShape(at_position Vector2) int32
}
```

These interfaces are a catalog: greppable, godoc-visible, zero runtime cost. Nothing is registered or dispatched from them, and satisfying one changes no behavior. The checked-in classification of the full virtual surface (`cmd/generate/gdclassimpl/virtual_census.json`) is a codegen test fixture whose criterion is documented in the file itself; a codegen test fails loudly if the generated surface and the census drift.

### Compile-time signature verification

Go interface satisfaction is all-or-nothing, and the class-wide `<Class>Virtuals` interfaces carry the Godot class-level names, so they do not serve per-override checking. Instead, verify one overridden virtual's exact signature with a per-method anonymous interface:

```go
// Compiles only if TestVirtualsConformance declares that exact method.
var _ interface{ V_TestVirtualsConformance_GetMinimumSize() Vector2 } = (*TestVirtualsConformance)(nil)
```

Take the parameter and return types from the generated catalog (for `_get_minimum_size`, `ControlVirtuals` says `Vector2`), name the method per the qualified convention, and a mismatch surfaces at build time instead of registration time. See `test/pkg/virtuals_conformance.go`.

### Delegation to engine defaults is impossible

Once a virtual is registered, an override of an engine GDVIRTUAL replaces its behavior wholesale: Godot resolves virtual presence once per object instance into a cached member pointer (`gdvirtual.gen.h` — absence caches as an INVALID sentinel) and never re-consults, so the engine-default body is unreachable from GDExtension for that instance. A plain-wrapper call from inside the override (e.g. a `_get_maximum_size` override calling `GetMaximumSize()`) re-enters the registered implementation forever:

| Reference binding | Override mechanism | Engine-default reach |
|---|---|---|
| godot-cpp | C++ inheritance; compile-time override detection via `if constexpr (!std::is_same_v<decltype(&B::_x), decltype(&T::_x)>)` | none |
| godot-rust | `I<Class>` capability traits; unimplemented trait methods are never registered | none ("No access to `super` methods", book) |

The constraint is pinned, not just documented: `TestDelegationRepro` (`test/pkg/delegation_repro.go`) registers a `_get_maximum_size` override that delegates through the plain wrapper, caps the re-entry depth, and panics with a `delegation recursion confirmed` diagnostic; `make test_delegation_trap` runs it in a separate godot invocation and fails unless the process aborts non-zero with that diagnostic. To extend engine or base behavior, delegate explicitly to a named level implementation (see Explicit delegation replaces super-calls above) — do not call the plain wrapper from inside an override.

## Property Registration

Properties are registered on Go-defined classes with `ClassDBAddProperty`,
which registers hint-free metadata (`PROPERTY_HINT_NONE`, empty hint string,
`PROPERTY_USAGE_DEFAULT`):

```go
ClassDBAddProperty(t, GDEXTENSION_VARIANT_TYPE_VECTOR2, "custom_position", "set_custom_position", "get_custom_position")
```

For full inspector metadata, use `ClassDBAddPropertyInfo` with a
`PropertySpec`:

```go
ClassDBAddPropertyInfo(t, PropertySpec{
	VariantType: GDEXTENSION_VARIANT_TYPE_OBJECT,
	Name:        "icon",
	Setter:      "set_icon",
	Getter:      "get_icon",
	Hint:        PROPERTY_HINT_RESOURCE_TYPE,
	HintString:  "Texture2D",
})
```

- `Hint`/`HintString` carry the `PROPERTY_HINT_*` inspector behavior (range
  strings, enum name lists, resource class filters, ...). The zero value of
  `Hint` is `PROPERTY_HINT_NONE`.
- `Usage` zero is a sentinel meaning `PROPERTY_USAGE_DEFAULT`; to reduce
  flags, OR real ones (e.g. `Usage: PROPERTY_USAGE_STORAGE` for a
  non-editor property).
- `class_name` derivation is ported from godot-cpp: with
  `PROPERTY_HINT_RESOURCE_TYPE`, the property's `class_name` is the resource
  class named by `HintString`; otherwise an explicit `ClassName` qualifier
  wins; otherwise the owning class name is used. `PROPERTY_HINT_RESOURCE_TYPE`
  with an empty `HintString` panics at registration (authoring fault).
- An empty `Setter` registers a read-only property.
- All `StringName`/`String` temporaries backing the registration are created
  and destroyed inside `ClassDBAddPropertyInfo`; callers pass plain Go
  strings and manage nothing. Godot copies the `PropertyInfo` pointer data
  during the registration call.

### Class-Level Property List Callback

`ClassDBRegisterClass` also accepts `[]PropertyDescriptor` entries whose
properties are exposed through the class `get_property_list` callback. This
input replaces the previous `[]GDExtensionPropertyInfo` parameter (and the
removed `ClassInfo.PropertyList` field): descriptors carry plain Go strings,
and the raw C structs plus the engine strings behind them are owned by the
library internals.

```go
props := []PropertyDescriptor{
	{ClassName: "Example", Type: GDEXTENSION_VARIANT_TYPE_VECTOR3,
		Name: "property_from_list", Hint: PROPERTY_HINT_NONE,
		Usage: PROPERTY_USAGE_DEFAULT},
}
ClassDBRegisterClass(NewExampleFromOwnerObject, props, ValidateExampleProperty, bind)
```

The lifetime of the backing data splits in two
(`openspec fix-class-property-list-lifecycle`):

- **Class-owned strings:** each descriptor's `ClassName`, `Name`, and
  `HintString` materialize once as engine `StringName`/`String` objects in
  stable C cells at registration. No Go memory backs them, so nothing is
  pinned and the GC can never move them. `ClassInfo.Destroy` (reached via
  `ClassDBUnregisterClass`) destroys each engine string exactly once.
- **Query-owned array:** every callback invocation `calloc`s a fresh
  `GDExtensionPropertyInfo` array from the class cells and returns it. The
  engine deep-copies the elements and releases the array through the paired
  `free_property_list2` callback in the same scope
  (`Object::get_property_list`), which frees the array only — never the
  class-owned strings. Repeated queries therefore always see the full list,
  and query timing can never double-free anything.

## Main-Loop Callbacks

Extensions can receive Godot's main-loop phase callbacks (Godot 4.5+) by
registering hooks on the `InitObject` **before** calling `Init()`:

```go
initObj.RegisterStartupCallback(func() { /* engine fully initialized, before scene processing */ })
initObj.RegisterFrameCallback(func() { /* every process frame, after Node._process */ })
initObj.RegisterShutdownCallback(func() { /* main loop shutting down, engine still initialized */ })
return initObj.Init()
```

This mirrors godot-cpp's `register_startup_callback` /
`register_frame_callback` / `register_shutdown_callback` exactly:

- One hook slot per phase; assigning again replaces the previous hook
  (last-write-wins). There is no deregistration.
- Registration with the engine happens once at the core initialization
  level, after the level's init callback has run — hooks assigned after the
  core level has already fired do not reach the engine.
- The frame hook runs on the engine main thread after all `_process()`
  methods on `Node` and before `ScriptServer::frame()`. Keep it non-blocking;
  it is the intended ledge for per-frame binding runtime.
- Engines older than 4.5 lack the interface function: registration is
  skipped with a debug log and the extension keeps loading.
- Hook panics are **not** recovered. A panicking hook aborts the process,
  the same policy as every other binding callback — a hook that can fail
  should contain its own `recover()` if that is unacceptable.

## Singleton Accessors

Every engine singleton declared in the extension API metadata gets a
generated typed accessor in `pkg/gdclassimpl`:

```go
engine := GetEngineSingleton()
if editor := GetEditorInterfaceSingleton(); editor != nil {
    // editor-only singleton: nil outside the editor
}
```

- Generation covers the API's `singletons` list verbatim — all entries,
  editor and platform ones included — via `singletons.go.tmpl` into
  `singletons.gen.go`. Adding or removing an API entry and regenerating adds
  or removes exactly one accessor; no handwritten per-singleton code.
- Identity-stable: the wrapper is resolved at most once per process and every
  call returns the same value. Publication is atomic and the cold path is
  mutex-guarded, so calling an accessor from any goroutine is safe.
- A singleton the engine does not currently provide returns a nil interface
  — check it like any other interface value. Absence is never cached: if the
  singleton appears later (`Engine.register_singleton`), the next call
  returns a live wrapper.

## Default Argument Values

Go does not support default parameter values in its syntax. Default argument values are instead passed through the `defaultValues` parameter of `ClassDBBindMethod` (and the `ClassDBBindMethodVirtual`/`ClassDBBindMethodVarargs` variants).

Defaults follow Godot's **trailing convention**: the bound defaults array maps to the **last** N parameters of the signature, so `defaultValues[k]` is the default for parameter `(declared - N + k)`. A caller may omit up to N trailing arguments; each omitted slot receives its own bound default by position. For example, a two-parameter method bound with a single default `[]Variant{200}` gives the second parameter a default of 200, so calling with just the first argument fills the second with 200. Binding more defaults than declared parameters is rejected at bind time.

## Call Error Reporting

When GDScript calls a bound method through the dynamic varcall path, the engine hands the call to Go together with a `GDExtensionCallError` slot. godot-go validates the call against the bound signature **before** invoking the method and reports a mismatch through that slot rather than aborting:

- **Too many arguments** → `GDEXTENSION_CALL_ERROR_TOO_MANY_ARGUMENTS` (`expected` = declared count). Variadic (`varargs`) methods accept any argument count and are never rejected here.
- **Too few arguments** → `GDEXTENSION_CALL_ERROR_TOO_FEW_ARGUMENTS`, only when the call cannot be satisfied even after applying bound default arguments. The required count is derived from the same rule used to fill defaults, so validation and default application never disagree.
- **Non-convertible argument type** → `GDEXTENSION_CALL_ERROR_INVALID_ARGUMENT` with the failing argument index and the expected variant type. Convertibility uses the engine's `variant_can_convert_strict` predicate (the same one godot-cpp uses). Under strict conversion these pass: numeric inter-conversion (int/float/bool), `STRING`↔`STRING_NAME`/`NODE_PATH`, and `nil` into an object parameter. These are rejected: `STRING`→number, number→`STRING`, and `nil` into a scalar parameter — `variant_can_convert_strict` is narrower than the loose `can_convert` table, so string/number inter-conversion is not permitted.

A rejected call leaves a nil return, never runs the bound Go method body, and keeps the process alive; the engine surfaces the call error to the caller as normal. Only faults a caller cannot induce (absent method user data, null instance) remain fatal. This applies to the varcall path only: the `ptrcall` fast path has no error slot in its GDExtension typedef.

## Static Methods

Go does not support static methods on structs, so godot-go binds a **package function** as a
class static method: `ClassDBBindStaticMethod[T GDClass](inst T, gdMethodName string, fn any, argNames []string, defaultValues []Variant)`.
The function carries no receiver — it is reflected with argument index base 0 — and `inst` is
used only to resolve the class name to register against. The bind registers with the engine's
static flag (mirroring godot-cpp's `bind_static_method`), so GDScript calls it through the class
path and the engine dispatches with a null instance pointer:

```go
// package-level function, no receiver
func ExampleStaticAdd(a int64, b int64) int64 { return a + b }

// inside the class registration callback (t is *Example)
core.ClassDBBindStaticMethod(t, "static_add", ExampleStaticAdd, []string{"a", "b"}, nil)
```

```gdscript
assert_equal(Example.static_add(2, 3), 5)
```

Static binds share every marshalling rule of instance binds: return styles, trailing default
arguments, a variadic `...Variant` tail (the engine's vararg flag is implied automatically),
and object arguments/returns through the standard ownership paths. Both dispatch channels
serve statics — the typed `ptrcall` fast path and the dynamic `varcall` path. A static bind
fails fatally at bind time when `fn` is not a Go function (naming class and method), when the
argument names or defaults outnumber the function's parameters, when the GDScript method name
is already taken on the class (the prior registration stays intact), or when the virtual flag
is combined with static (Godot has no static virtual dispatch). Calling a class static through
an instance is not asserted here: the engine dispatches it (the static functor ignores the
instance), which the harness records as an observation.

## Static Variables

**BLOCKED** — the GDExtension ABI offers no interface to register a class variable.
`godot_headers/godot/gdextension_interface.h` has no `classdb_register_extension_class_variable`
equivalent, so there is nothing a binding can call; the earlier note that "global variables can be
registered as gdscript static variables" has no engine channel. Go package-level variables remain
plain Go globals. This can only move once the experimental LanguageExtension interface (not
vendored here) reaches stable Godot.

## Packed Arrays

The generated `Packed*Array` types work and are tested in the tests. Go-native
slices cross the boundary directly: `[]byte`, `[]int32`, `[]int`/`[]int64`,
`[]float32`, `[]float64`, `[]string`, `[]Vector2`, `[]Vector3`, `[]Vector4`, and
`[]Color` map to the matching `Packed*K` container on arguments and returns.
Decoded slices are Go-owned deep copies (mutating one side never touches the
other, and a decoded slice stays valid after the call), and a nil or empty Go
slice returns as an empty container, never `null`. Element types with no packed
counterpart (`[]bool`, `[]uint16` and wider unsigned kinds, `[]Vector2i`/
`[]Vector3i`, slices of refcounted cells such as `[]String`, nested slices) are
rejected at bind time naming the element type. A trailing `...Variant` variadic
slot keeps its predating registration and is unaffected.

## Coroutines

Go does not support coroutines, so there is no `await` keyword that parks the
current thread inside engine execution. The equivalent affordance is built on
goroutines instead: `core.Await` bridges a Godot signal emission to a Go
channel, so any goroutine can block on an event the way GDScript blocks on
`await` — without stalling the engine's main thread.

```go
// GDScript:  var args = await enemy.defeated
res, ok := <-core.Await(enemy, "defeated")
if !ok {
	// enemy was freed before emitting; the signal can never fire
}
for _, arg := range res.Args {
	// Go-native payloads: int64, float64, string, Vector2, Object, ...
}
```

The subscription is one-shot. Because the result is an ordinary channel it
composes with the standard toolkit — `select` with `time.After` for timeouts,
`context.Context` for cancellation, or racing several signals. Emissions
deliver Go-owned values marshalled on the emitting thread; see the
`core.Await` doc comment for the payload table and threading hazards.

To keep observing a signal rather than awaiting one emission, connect a Go
closure directly: `core.NewGoCallable` wraps a `func([]Variant) (Variant,
error)` as an engine Callable, and any object's `Connect` accepts it like a
script-bound callable:

```go
cb := core.NewGoCallable(func(args []builtin.Variant) (builtin.Variant, error) {
	log.Info("hit", zap.Int64("penalty", args[0].ToInt64()))
	return builtin.Variant{}, nil // nil Variant result
})
defer cb.Destroy()
node.Connect(name, cb, uint32(constant.OBJECT_CONNECT_FLAGS_REFERENCE_COUNTED))
```

## Editor Plugins

A registered class descending from `EditorPlugin` becomes an active editor
plugin through `AddEditorPlugin` / `RemoveEditorPlugin`, mirroring godot-cpp's
`EditorPlugins::add_by_type<T>()` / `remove_by_type<T>()`:

```go
type MyPlugin struct {
	gdclassimpl.EditorPluginImpl
}

func (p *MyPlugin) GetClassName() string       { return "MyPlugin" }
func (p *MyPlugin) GetParentClassName() string { return "EditorPlugin" }

// from the extension's editor-level init callback:
ClassDBRegisterClass(newMyPlugin, nil, nil, func(t *MyPlugin) {})
AddEditorPlugin(&MyPlugin{})
```

- The instance argument carries only the registered class name. The engine
  instantiates the class itself when the add reaches `EditorNode` — an add
  issued during editor-level initialization is queued and drained when the
  `EditorNode` is constructed — so the editor init level is a fine call
  site, and the instantiated instance is a fresh Go object (the binding
  builds instances reflectively from the registered struct).
- Duplicate adds and removals of classes that were never added panic
  (`log.Panic`) before anything crosses to the engine; the panic is
  recoverable.
- Outside an actual editor session both operations panic instead of
  silently pretending: TOOLS-enabled engine builds dispatch the EDITOR
  init level even while running the project as a game, and there the
  `EditorInterface` singleton is absent. Gate plugin registration on the
  editor session — `GetEngineSingleton().IsEditorHint()` — and unregister
  the class from the symmetric editor terminator.
- Plugins left added are removed automatically when the editor-level
  deinitialization runs, after that level's terminator callback and in
  insertion order, mirroring godot-cpp
  `GDExtensionBinding::deinitialize_level`. Explicit `RemoveEditorPlugin`
  during the session is symmetric and stays fully supported.

## Built-in Types

### Basic Built-in Types

| GDScript Type | Go Type | Description |
| --- | --- | --- |
| `null` | `nil` | |
| `bool` | `bool` | |
| `int` | `int64` | All method parameters that use variations of `uint` and `int` will be converted to `int64` before passing over the value to Godot. |
| `float` | `float64` | `float32` will convert to `float64` before passing over the value to Godot. |
| `String` | `String` | There are helper functions to convert to go native `string`. |
| `StringName` | `StringName` | There are helper functions to convert to go native `string`. |
| `NodePath` | `NodePath` | |

### Vector Built-in Types

| GDScript Type | Go Type |
| --- | --- |
| `Vector2` | `Vector2` |
| `Vector2i` | `Vector2i` |
| `Rect2` | `Rect2` |
| `Vector3` | `Vector3` |
| `Vector3i` | `Vector3i` |
| `Transform2D` | `Transform2D` |
| `Plane` | `Plane` |
| `Quaternion` | `Quaternion` |
| `AABB` | `AABB` |
| `Basis` | `Basis` |
| `Transform3D` | `Transform3D` |

### Engine built-in Types

| GDScript Type | Go Type |
| --- | --- |
| `Color` | `Color` |
| `RID` | `RID` |
| `Object` | `Object` |

### Container Built-in Types

| GDScript Type | Go Type | Description |
| --- | --- | --- |
| `Array` | `Array` | `[]Variant`. |
| `PackedByteArray` | `PackedByteArray` | `[]byte`. |
| `PackedInt32Array` | `PackedInt32Array` | `[]int32`. |
| `PackedInt64Array` | `PackedInt64Array` | `[]int64`. |
| `PackedFloat32Array` | `PackedFloat32Array` | `[]float32`. |
| `PackedFloat64Array` | `PackedFloat64Array` | `[]float64`. |
| `PackedStringArray` | `PackedStringArray` | `[]string`. |
| `PackedVector2Array` | `PackedVector2Array` | `[]Vector2`. |
| `PackedVector3Array` | `PackedVector3Array` | `[]Vector3`. |
| `PackedVector4Array` | `PackedVector4Array` | `[]Vector4`. |
| `PackedColorArray` | `PackedColorArray` | `[]Color`. |
| `Dictionary` | `Dictionary` | No additional work needed. |
| `Signal` | `Signal` | No additional work needed. |
| `Callable` | `Callable` | No additional work needed. |

# Object Lifetime and Handle Ownership

A Go wrapper that stands for an engine object is anchored by a `cgo.Handle` — stored either as the extension instance of a user-defined class or in the object's instance-binding slot under this library's token. The Handle (and everything reachable through it) is kept alive exactly as long as the engine can hand the object back to Go, and not longer.

## One release owner per Handle

Each Handle is deleted by exactly one of the engine's release notifications, and no other path deletes it:

* **User-defined class instances** — `GoCallback_ClassCreationInfoFreeInstance` deletes the instance Handle after `Destroy()` returns (deferred LIFO, so re-entrant callbacks running during `Destroy` still see a live Handle). The binding free callback for this shape, `GoCallback_GDClassBindingFree`, is a deliberate no-op: its slot holds the same Handle and `FreeInstance` fires strictly earlier, from `Object::_predelete`.
* **Engine-class bindings** created by the engine's binding create callback — `GoCallback_GDExtensionBindingFree` deletes the Handle stored in the binding slot. The engine runs it exactly once per created binding.

Deleting a Handle makes later `Value()` lookups panic, so a stale slot read cannot silently return another object's wrapper; binding lookups recover those panics into typed errors (see `ObjectFromInstanceBinding`). This is safe because the engine nulls `_extension`/`_extension_instance` immediately after `free_instance` returns — no engine callback follows a freed instance with the stale pointer.

## Handles cross the C boundary as integers

A packed Handle is a small integer, and the Go runtime aborts (`invalid pointer found on stack`) when a GC stack scan finds a small integer in a GC pointer slot. The ffi generator therefore marks handle-carrying slots (`cmd/generate/ffi/handle_slots.go`) and renders them `uintptr_t` through the C shims and `cgo.Handle` on the Go side — binding `p_binding` parameters, `object_set_instance`'s `p_instance`, and the binding/userdata getters — and the engine→Go binding callbacks cross the value as `uintptr`. Struct userdata fields (`method_userdata`, `callable_userdata`) are written through the `cgo_set_handle_slot` shim so the cast lives in C.

## Pinning discipline

Per-instance values (class-name pointers, callback tables) are pinned only for the duration of the call that hands them to the engine (`runtime.Pinner` scoped to the construction path). The program-lifetime pinner retains registration-time data only — never per-instance state, which would outlive every release hook.
