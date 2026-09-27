## Why

Generated ptrcall bindings cannot pass engine-class objects as arguments. `cmd/generate/gdclassimpl/classes.go.tmpl` encodes every argument as `unsafe.Pointer(&arg)` regardless of type, and its engine-class branch (`{{ if $view.ContainsClassName ... }}`) emits **byte-identical code to the value-type branch** — the commented-out `.Reference()` at the branch head shows the divergence was intended and never written.

For value types this is correct: the Go variable's memory holds the value. For an engine-class argument the Go variable is an *interface*, whose memory is the `{itab, dataptr}` header, so Godot reads the itab pointer as a `GDExtensionObjectPtr` and segfaults. Verified: a Go `_ready` that creates one `CharacterBody2D` and calls `AddChild` crashes with SIGSEGV before returning. **794 generated methods across 164 engine-class argument types** take at least one such argument today, so programmatic node spawning is unavailable from Go. Object *returns* work — `GetChild` decodes into a concrete `NodeImpl` whose first field is the owner pointer — which is why the gap stayed hidden.

## What Changes

- In `classes.go.tmpl`, make the engine-class argument branch encode the **object pointer** rather than the interface header: bind a local `GDExtensionObjectPtr` from `AsGDExtensionObjectPtr()` and pass *its* address, so the argument slot points at a pointer-valued cell as GDExtension requires.
- Cover both object shapes: plain engine classes (`Node`, reachable through `Wrapped`) and smart-pointer arguments (`RefShape2D` and friends, reachable through `Ref.ToObject()`).
- Define nil handling: a nil interface or invalid `Ref` argument SHALL encode as a null object pointer instead of panicking on a nil-method call.
- Regenerate bindings (`make generate`) and add round-trip tests asserting a Go method receives the same object it was given, plus nil and invalid-`Ref` cases.

No refcount change. Godot does not take ownership of call arguments — its C++ setter assigns and refcounts itself — so the commented-out `.Reference()` remains unnecessary and must not be reinstated.

## Capabilities

### New Capabilities

- `object-argument-encoding`: How Go-side engine-class objects (`Object` and `Ref[T]`) are encoded as ptrcall arguments crossing Go→Godot, including nil and invalid-reference handling.

### Modified Capabilities

None. `ref-smart-pointer` governs `Ref` *return* encoding and `Ref` *argument decoding* (Godot→Go); the Go→Godot argument direction is uncovered today. `basic-built-in-types` and `vector-built-in-types` cover value types, where `unsafe.Pointer(&arg)` is already correct.

## Impact

- `cmd/generate/gdclassimpl/classes.go.tmpl` — the argument-encoding branch (lines ~168–178).
- `pkg/gdclassimpl/classes.gen.go` — regenerated; all 794 affected method wrappers change shape.
- `test/pkg/example.go`, `test/demo/main.gd` — object-argument round-trip and nil tests.
- Unblocks downstream: `Node.AddChild`, `CollisionShape2D.SetShape`, and every other object-taking call becomes usable from Go.
- No signature changes; existing Go code that compiles today keeps compiling.

## Non-goals

- **Varcall argument encoding** — variadic calls marshal through `Variant`, which already encodes objects correctly.
- **Object return paths** — already correct; not touched.
- **Input event type-switching** (`event.(type)` not discriminating Godot event types) — a separate, independently-tracked limitation.
- **Value-type argument encoding** — unchanged.
- **Removing the duplicated branch** rather than fixing it — the branch must stay, because the two cases genuinely need different encodings.
