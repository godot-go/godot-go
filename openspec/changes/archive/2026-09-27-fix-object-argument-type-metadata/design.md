# Design

## Context

`NewGoMethodMetadata` (`pkg/core/method_bind.go`) builds the `GDExtensionPropertyInfo` for every bound method's arguments (loop at ~lines 243–257) and return (~lines 219–229). Both set `class_name` to the **owning** class (`className`), because that was the only class name in scope. Godot treats `class_name` as authoritative for OBJECT-variant parameters and returns: GDScript static typing validates call sites against it, so an `Example`-bound method taking `Node` rejects every typed caller. The object-argument encoding fix (e727e39) landed with a workaround — untyped `var` declarations in `test/demo/main.gd` — and explicitly deferred this metadata bug.

The correct class name is recoverable from the Go type. The generator already proves the mapping is structural: `cmd/generate/gdclassimpl/classes.go.tmpl` picks `RefXxx` vs `Xxx` via `goArgumentType` + `$view.IsRefcountedClassName` / `$view.ContainsClassName` (~lines 102–108), backed by `cmd/extensionapiparser/model.go`. At runtime the same correspondence is registered authoritatively: `GDNativeConstructors` is keyed by exact Godot class names (`"Node"`, `"Object"`), `GDClassRefConstructors` by the pointee class name (`"Shape2D"` → `NewRefShape2DAsRef`), and `Internal.GDRegisteredGDClasses` by registered extension class names. `ReflectTypeToGDExtensionVariantType` (`pkg/core/variant_reflect_type.go`, ~lines 223–249) already consults these registries to decide OBJECT-ness, including the `Ref`-prefix check `GDClassRefConstructors.Get(tn[3:])`.

## Goals / Non-Goals

**Goals:**
- OBJECT-typed parameters advertise their own Godot class (`Node` → `"Node"`, `RefShape2D` → `"Shape2D"`).
- OBJECT-typed returns advertise their own Godot class, symmetrically.
- GDScript typed variables of the correct class or a subclass pass static checking; wrong classes are engine-rejected.
- Value-type (non-OBJECT) parameter and return metadata is bit-identical to today.
- Loud bind-time failure when an OBJECT-typed Go type resolves to no registered class.
- No change to the public bind API (`ClassDBBindMethod`, `ClassDBBindMethodVirtual`) and no codegen changes.

**Non-Goals:**
- Changing object argument/return *encoding* (owned by `object-argument-encoding`).
- Renaming the argument `PropertyInfo.name` field (currently the Go type name; cosmetic, separate concern).
- Vararg bindings (registered with zero named arguments).
- Engine-side polymorphism semantics — we only advertise the class; the engine decides acceptance.

## Decisions

**D1: Resolve the class name in the runtime binder via registry lookup, not by carrying it through the public bind API.**
Threading a per-argument Godot class name through `NewGoMethodMetadata` would require new parameters on `ClassDBBindMethod`/`ClassDBBindMethodVirtual` — a breaking change at every existing call site (test fixtures, user classes) for information that is already authoritative in the registries. Instead, a new internal helper in `pkg/core` (e.g. `godotClassNameForObjectType(t reflect.Type) (string, bool)`) resolves the name once inside `NewGoMethodMetadata`, and the result is stored in the existing lifecycle-managed `gdeArgPropClassNameStringNames` / `gdeReturnPropClassNameStringName` fields — no lifecycle changes.

**D2: Resolution order is plain-name first, then `Ref`-prefix, with generic fallbacks.**
For an OBJECT-variant Go type `t` with name `tn`:
1. `GDNativeConstructors.Get(tn)` → `tn` (engine classes: `Node`, `Object`, `Image`).
2. `Internal.GDRegisteredGDClasses.Get(tn)` → `tn` (user extension classes).
3. `strings.HasPrefix(tn, "Ref")` and `GDClassRefConstructors.Get(tn[3:])` succeeds → `tn[3:]` (smart pointers: `RefShape2D` → `Shape2D`).
4. Generic object interfaces (`Object`, `Ref`, `GDClass`, `GDExtensionClass`) → `"Object"`.
5. Otherwise → not resolved → `log.Panic` identifying method and type.
Plain-name-first ordering avoids a mis-strip if a user registers an extension class literally named `RefFoo`. Every candidate is validated against a registry before acceptance, so this is registry-backed resolution, not unchecked string-sniffing. The `Ref`-prefix convention is already load-bearing in `ReflectTypeToGDExtensionVariantType`, so no new coupling is introduced.

**D3: Gate on the already-computed variant type.**
The helper is consulted only where `ReflectTypeToGDExtensionVariantType` returned OBJECT. Value types keep the existing `className` metadata untouched, guaranteeing zero behavior change for them and keeping the diff surgical.

**D4: Fix the return path in the same change.**
The return `PropertyInfo` has the identical defect and the same resolution applies. Fixing only arguments would leave typed GDScript return assignments (`var n: Node = example.foo()`) mis-advertised and invite a second, larger workaround.

## Risks / Trade-offs

- **Tighter return class may surface engine return-validation errors.** With the correct class advertised, a Go method that returns an object of a different class than its declared Go return type now hits engine validation instead of silently passing. Accepted: the Go return type is the contract, and silent mis-advertising was the bug.
- **Registry population ordering.** Resolution requires `GDNativeConstructors`/`GDClassRefConstructors` to be populated before method binds. They are populated during extension init before any class registers methods; a unit test that exercises resolution must seed the registries directly (they are plain Go sync maps).
- **Generic `Ref` edge case.** `strings.HasPrefix("Ref", "Ref")` is true but `tn[3:]` is empty; the generic-interface fallback (D1 step 4) must be checked so a bare `Ref` advertises `"Object"` rather than `""`.
- **Wrong-class rejection is a GDScript parse error**, so it cannot live in the same compiled demo script. Covered by a negative fixture script validated with `godot --check-only` (or editor verification) rather than a runtime assert.
- **Trade-off accepted:** resolution duplication between the generator (`IsRefcountedClassName`) and the runtime registries is tolerable because both derive from the same `extension_api.json` naming, and the runtime cannot call generator code.

## Corrections found during implementation

**The resolution order missed pointer types, and the guard caught it.** `reflect.Type.Name()` is empty for a pointer, and concrete Go class implementations are pointers — `ReturnEmptyRef` returns `*pkg.ExampleRef`. Running the resolver as designed made every such return look unresolvable. The helper now walks to the element type first.

**The "fail loudly at bind time" requirement was wrong and has been revised.** It fired on `*pkg.ExampleRef`, a legitimate object type whose `RegisterClassExampleRef()` is commented out at `test/pkg/lib.go:20`. A Go type can be a valid object without being a registered Godot class, so unresolvable is not a bind error. Panicking broke a working bind; advertising the owning class would be a lie. The requirement now specifies `"Object"` plus a warning — truthful, since every engine object is an `Object`, so no correctly typed call is rejected.

**This change could not land before the ptrcall object decode, and the reason is worth keeping straight.** Landing the metadata fix alone made typed GDScript object arguments parse, which made the engine dispatch them through ptrcall rather than varcall, which hit `panic: unsupported interface type` in a decoder that had never received an object interface before. The work was stashed and reapplied after `fix-ptrcall-object-arg-decode` landed. Two defects in two different paths, each invisible while the other masked it: the metadata bug kept typed calls from parsing, and the decode bug kept the ptrcall object branch from ever being reached. Fixing either one alone exposed the other.

**Task 2.3 as written is not achievable in a plain `go test`.** `NewGoMethodMetadata` builds `StringName`s and needs a live engine. The resolution logic is unit-tested where it can be, and the wiring is verified in-engine by typed call sites plus a falsification showing those call sites are gated by the advertised class.
