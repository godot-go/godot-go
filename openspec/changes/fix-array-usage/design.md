## Context

See proposal.md. This change corrects a misdiagnosis and fixes the actual bug.

### How method resolution actually works (Godot source)

1. `classdb_get_method_bind` → `ClassDB::get_method_with_compatibility` (`core/object/class_db.cpp`) looks up `method_map` keyed by method **name**; the hash is only a `(*method)->get_hash() == p_hash` verification, not a lookup key.
2. `Object::callp` (`core/object/object.cpp`) also resolves via `ClassDB::get_method(get_class_name(), p_method)` — by name.
3. Method hashes routinely collide across the API (`3341600327` appears 183 times across all classes); the name-keyed map makes this irrelevant.

The "hash collision" claims in the previous change are contradicted by these paths.

### The real bug: unimplemented GDVIRTUAL return values

Godot's `GDVIRTUAL0RC` macro (`core/object/gdvirtual.gen.h`) for a virtual with a return value:

```cpp
PtrToArg<m_ret>::EncodeT ret;                       // uninitialized local
_get_extension()->call_virtual_with_data(..., &ret);
r_ret = (m_ret)ret;                                 // read back unconditionally
```

`Control::get_maximum_size()` initializes `ms = Vector2(-1, -1)` then calls `GDVIRTUAL_CALL(_get_maximum_size, ms)`. When the extension claims the virtual is overridden, Godot discards its own default and reads the callback's buffer.

godot-go registers `get_virtual_call_data2` → `GoCallback_ClassCreationInfoGetVirtualCallWithData`, which **always returns non-nil** (`pkg/core/classdb_callback.go`). So for a Go class that does not implement `_get_maximum_size`, Godot calls `GoCallback_ClassCreationInfoCallVirtualWithData`, which returns early without writing `r_ret` (`pkg/core/classdb_callback.go:117-123`). Godot then reads uninitialized memory → `(0, 0)`.

### Why Godot 4.7 exposed it

Commit `bdca8b66e7` ("Add `custom_maximum_size` to `Control`", landed in 4.7-stable, **not** in 4.6.x) added a maximum-size clamp to `Control::set_size`:

```cpp
Size2 max = get_combined_maximum_size();            // (0,0) on GDExtension subclasses
if (max.x >= 0 && new_size.x > max.x) new_size.x = max.x;   // clamps to 0
```

So any `set_size` call on a GDExtension `Control` subclass is clamped to `(0, 0)`. On 4.6.3 there was no clamp, so the latent binding bug was invisible — which is why CI was green at HEAD.

### Evidence (live, 4.7.2)

| Probe (GDScript) | `Example` (GDExtension) | Plain `Control` |
|---|---|---|
| `set_size(Vector2(100,200))` → `get_size()` | `(0,0)` | `(100,200)` |
| `get_maximum_size()` | `(0,0)` | `(-1,-1)` |
| `get_combined_maximum_size()` | `(0,0)` | `(-1,-1)` |
| `set_position` / `set_pivot_offset` / `set_rotation` | work | work |

`set_size` is the only broken method: it is the one that consults `get_combined_maximum_size()`.

## Goals / Non-Goals

**Goals:**
- Make unimplemented GDVIRTUALs fall back to engine defaults.
- Remove the `hasHashCollision` workaround and restore standard method-bind codegen.
- Restore full `TestSetPositionAndSize` coverage.

**Non-Goals:**
- Fixing Godot's GDVIRTUAL machinery.
- Reverting the 4.7.2 header upgrade.
- Changing method-hash values.

## Decisions

### Decision 1: Root cause is GDVIRTUAL return handling, not hash collision

Supported by: Godot's name-keyed method resolution; the ubiquitous hash collisions (`3341600327` appears 183 times) that are clearly harmless; and CI passing on 4.6.3 with identical generated code. The hash-collision story was a misdiagnosis.

### Decision 2: Return `nil` call data for unimplemented virtuals

`GoCallback_ClassCreationInfoGetVirtualCallWithData` SHALL return `nil` when `ci.VirtualMethodMap[methodName]` is absent. Godot then sets the virtual fn ptr to `_INVALID_GDVIRTUAL_FUNC_ADDR` (`core/object/object.cpp:902`) and skips the call, preserving engine defaults. Implemented virtuals continue to return `pUserdata`.

### Decision 3: Remove the `hasHashCollision` workaround

It reroutes `Control` methods to the `Object.Call()` Variant path based on a false theory, silently weakening codegen and contradicting its own rationale. Revert to the standard `classdb_get_method_bind` + `object_method_bind_ptrcall` path.

### Decision 4: Restore full test coverage

Re-enable `SetSize` in `TestSetPositionAndSize` and restore the `get_size() == (100, 200)` assertion in the GDScript suite.

## Risks / Trade-offs

| Risk | Mitigation |
|---|---|
| Other GDVIRTUALs with non-default returns were silently zeroed; user code may have relied on the zeros | The zeros were uninitialized memory, not a contract; note in proposal and monitor |
| `get_virtual_call_data2` is also used by tooling/editor paths | `make test` exercises class registration and virtual dispatch end-to-end |
| Extra map lookup in the dispatch path | Negligible; `CallVirtualWithData` already performs the same lookup |

## Migration Plan

None. Behavior change is a bug fix toward correct engine defaults; codegen must be regenerated.

## Open Questions

- Which other GDVIRTUALs with non-default return values were affected (e.g. `_get_contents_minimum_size`, bool-returning virtuals)? An audit after this lands would be valuable.
- Should `CallVirtualWithData` defensively zero the return buffer when a virtual is implemented but `Ptrcall` fails?
