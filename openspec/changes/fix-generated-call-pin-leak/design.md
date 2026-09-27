# Design

## Context

Verified state of the repo (re-read 2026-09-27):

- Package-global pinners: `pkg/builtin/lib.go:33`, `pkg/ffi/lib.go:19`, `pkg/gdclassimpl/lib.go:12`, `pkg/core/lib.go:38`, `pkg/gdutilfunc/lib.go:15`, `pkg/gdclassinit/lib.go:21` — all `pnr runtime.Pinner{}`, none ever unpinned.
- `pkg/gdclassimpl/classes.gen.go` contains 63,086 `pnr.Pin(...)` calls; `grep -rn Unpin pkg/` returns zero hits.
- Generated method bodies pin, on the global pinner: the return slot (`retPtr`), the owner pointer cell (`cOwner`), the argument slice backing array (`cArgs`), and every argument slot — see `cmd/generate/gdclassimpl/classes.go.tmpl:140-191`, `cmd/generate/gdutilfunc/utilityfunctions.go.tmpl:59-87`, `cmd/generate/builtin/builtinclasses.go.tmpl:92-192`, `cmd/generate/builtin/variant.go.tmpl:22-66`.
- `cmd/generate/ffi/ffi_wrapper.go.tmpl` (lines ~86, ~138, ~193) already contains the correct scoped pattern — `pinner := runtime.Pinner{}` + `defer pinner.Unpin()` — but commented out. The live return-value pin in `pkg/ffi/ffi_wrapper.gen.go` comes from `cgoPinReturnType` (`cmd/generate/ffi/templatefunctions.go:268-281`), which emits `pnr.Pin(ret)`.
- Hand-written per-call pinners with the same leak: `pkg/ffi/builtin_ptrcall.go` (every `CallBuiltin*` helper) and `pkg/core/method_bind_callback.go` (every Godot→Go virtual call pins `inst` and argument slots).
- Empirical evidence (sibling change `fix-object-arg-ptrcall-encoding`, task 4.3): `GetShape()` pins its `Shape2DImpl` return wrapper per call; the pinned wrapper keeps the refcounted resource reachable at shutdown, tripping `Leaked instance: CircleShape2D ... RID allocations of type 'GodotShape2D' were leaked at exit`. `SetShape` alone leaked nothing.

Relevant `runtime.Pinner` semantics (Go 1.25 runtime, `runtime/pinner.go`):

- "A Pinner arranges for its objects to be automatically unpinned some time after it becomes unreachable" — a package-level pinner is reachable forever, so its referents are retained forever. This is the leak mechanism.
- "It's safe to call Pin on non-Go pointers, in which case Pin will do nothing" — pins on C/engine-memory pointers (e.g. `pnr.Pin(ret)` on FFI return values, `pnr.Pin(instPtr)` on engine instance pointers) are no-ops today.
- Multiple pins of the same object are reference-counted (`incPinCounter`), so a per-call pin layered over a registration-time pin unpins independently and safely.
- `Unpin()` returns the pinner to a per-P `pinnerCache`, so per-call pinner construction is cheap and reuse is explicitly encouraged by the runtime docs.

## Goals / Non-Goals

**Goals:**
- Pin lifetime for call scratch is bounded to the call: every pin made by a generated (or hand-written per-call) body is released when that body returns.
- No unbounded retention across repeated calls; the `GetShape`-style exit leak disappears.
- Values the engine retains beyond the call (method metadata strings, binding callbacks, version struct) remain pinned and valid for their full retention lifetime.
- One uniform pattern in generated code, produced by templates and verifiable by scan.
- `go build ./...`, `make generate`, and `GODOT=/home/pcting/bin/godot make test` stay green.

**Non-Goals:**
- Not eliminating pinning, nor redesigning it (no arenas, pools, or a custom pin abstraction). Pins remain; only their lifetime shrinks.
- Not changing registration-time/one-time pinning (`GoMethodMetadata` PropertyInfo strings, instance-binding callbacks, `FFI.GodotVersion`, `ClassInfo`); those stay on the package pinner by design.
- Not auditing away no-op pins on C-memory pointers where scoping them is sufficient (e.g. the callback's `rReturn` pin is scoped, not deleted, to preserve current safety posture).
- Performance tuning beyond leak removal; per-call Pinner cost is accepted (amortized by the runtime's per-P cache).
- No changes to varcall/ptrcall semantics, argument encoding, or refcounting.

## Decisions

### D1 — Per-call `runtime.Pinner` + `defer pinner.Unpin()` in every generated call body
Each generated per-call body opens with:

```go
var pinner runtime.Pinner
defer pinner.Unpin()
```

and every `Pin` in that body targets `pinner`. This mirrors the pattern already written (commented) in `ffi_wrapper.go.tmpl`. The pinner never escapes the body, so the deferred `Unpin` always runs, including on panic paths.

Alternatives considered:
- *Shared pooled pinner keyed by goroutine*: extra machinery for identical semantics; harder to audit; rejected.
- *Passing a pinner parameter into the `CallFunc_*` FFI wrappers*: signature churn across thousands of wrappers; the body-level pinner already covers everything the body hands to C; rejected.

### D2 — Classify every pin by post-call retention (the "must outlive" rule)
The governing question for each pin: **does any C-side reference to the pinned Go object exist after the call returns?**

| Category | Examples | C retains after call? | Pinner |
|---|---|---|---|
| Call-scoped scratch | `retPtr`, `cOwner`, `cArgs`, `argPtrSlice[j]` cells, encoded variant temporaries (`classes.go.tmpl`, `builtinclasses.go.tmpl`, `variant.go.tmpl`, `utilityfunctions.go.tmpl`, `builtin_ptrcall.go`) | No — ptrcall/varcall read the pointer values synchronously and copy them; the engine stores no pointer to these Go cells | Per-call, `defer Unpin()` |
| Returned objects | `NewRefShape2DGDExtensionIternalConstructor(&ret)` in `GetShape` | No — the engine placement-constructs a native handle *into* the Go slot; afterward only Go code references the wrapper, and the C object holds its own refcount | Per-call; the wrapper stays valid as an ordinary Go value after unpin |
| Registration-time retention | `GoMethodMetadata` `gdeReturnPropNameStringName` / `gdeArgPropNameStringNames` / hint strings (`pkg/core/method_bind.go:511-560`), instance-binding callback pointers (`pkg/builtin/lib.go:51-53`), `FFI.GodotVersion` (`pkg/ffi/lib.go:54`), `ClassInfo` (`pkg/core/types.go:121`) | Yes — the engine's method-bind/callback structures hold these pointers and read them later | Package-level `pnr`, never unpinned; bounded by registration count |

The templates distinguish these **statically**: per-call method bodies never store pins, and registration/init templates are the only ones that emit package-pinner pins. The review rule that falls out: `pnr.Pin` may appear only in registration/initialization code, never inside a generated call body (enforced by the guard test in tasks 4.2).

Why unpinning the return slot is safe (the subtle case): a pinned object is retained *and* its finalizer deferred; that is exactly why the `GetShape` wrapper leaked its refcounted resource. After `Unpin`, the wrapper is still a perfectly valid Go value owned by the caller through ordinary pointers; it simply becomes movable and collectable again, which is what lets refcount release happen at shutdown.

### D3 — Defer ordering and nil-slot behavior
`defer pinner.Unpin()` is registered first, so it runs last (LIFO) — after any `defer vN.Destroy()` variant temporaries. `Unpin` only clears pin bits on recorded Go objects, so it cannot interact with C-side destroy ordering. Typed-nil slots (`retPtr := (GDExtensionTypePtr)(nullptr)` on void-returning methods) need no special casing: `setPinned` silently ignores pointers outside the Go heap.

### D4 — Drop the global return-value pin in the FFI wrappers
`cgoPinReturnType` (`cmd/generate/ffi/templatefunctions.go:274`) emits `pnr.Pin(ret)` for pointer returns from C functions. All current such returns are engine-memory pointers, so these pins are documented no-ops today — but they encode the wrong lifetime story and would become real leaks if any wrapper ever returned a Go pointer. Decision: remove the global return-value pin. Callers that need to retain a returned pointer across further cgo calls are covered by their own body-level scoped pinner (D1); a task audits that no `CallFunc_*` return value flows into C as a Go pointer without such a pin.

### D5 — Hand-written per-call helpers adopt the same scoped pattern
`pkg/ffi/builtin_ptrcall.go` (six `CallBuiltin*` helpers) and `pkg/core/method_bind_callback.go` (per-virtual-call pins: `instPtr`, `inst`, argument slots, `rReturn`) leak per call exactly like generated code. They get body-scoped pinners rather than per-pin removal, preserving the current safety posture while bounding lifetime. The callback's pins on engine-memory pointers (`instPtr`, `rReturn`) are no-ops but are scoped rather than deleted so the pattern stays uniform.

### D6 — Keep the package-level pinner for registration-time retention, and guard it
The package `pnr` variables remain, documented as "program-lifetime retention only". A guard test scans generated call-path sources and fails if `pnr.Pin` appears inside a generated method/constructor/utility/variant body, so the regression cannot silently return.

## Risks / Trade-offs

- **R1 — Unpinning while C still holds a reference** would cause use-after-free or GC-move corruption. Mitigated by D2's retention rule (ptrcall/varcall are synchronous and copy pointer values), by keeping all genuinely-retained pins on the package pinner, and by regression tests: the `GetShape` exit-leak test, the metadata-after-unpin reflection test, and the existing object-identity round-trip tests from the sibling change.
- **R2 — Per-call Pinner overhead.** Each call constructs a `runtime.Pinner`; the runtime's per-P `pinnerCache` amortizes allocation and the finalizer, and the Go docs encourage reuse. Cost is one small object per call — acceptable against unbounded retention. No benchmark gate in this change.
- **R3 — Huge generated diff** (~63k pin sites). Mitigated by changing only templates and reviewing the regenerated output by pattern (declaration + `pinner.Pin` swap + `defer`), not line-by-line.
- **R4 — Runtime "leaking pinned pointer" panic** fires if a pinner is finalized with pins still held. Prevented by body-scoped `defer Unpin()`; the pinner never escapes the body.
- **R5 — Double-unpin throws** (`object already unpinned`). Avoided because `Unpin` is called only once via `defer` and per-call pinners are never reused across calls.
- **Trade-off — permanent registration pins remain.** They are intentional, bounded by registered class/method count, and required by the engine's retention of metadata pointers. This change bounds the *per-call* leak; it does not make the process pin-free.

## Corrections found during implementation

Two premises above did not survive contact with the running engine. Both are recorded here rather than quietly patched around.

**D2's leak mechanism was wrong.** The `GetShape()` exit leak was attributed to the pin retaining the return wrapper so its finalizer could not run. With every pin scoped, a 200-round `get_shape()` loop still leaked `Reference count: 200` — one per call. The real cause is `NewRef` in `pkg/builtin/ref_generic.go`, which deliberately sets no `Unref` finalizer because it models transfer semantics ("the other side already holds a reference"). The engine adds +1 on each returned `Ref<Shape2D>` and nothing ever releases it. Scoping pins makes the wrapper collectable but does not unref it. That is the object-return-ownership defect, a separate change with its own premature-free risk surface; task 6.2 is left open against it.

**D1's blanket rule cannot cover the `As*Ptr` accessors.** Removing the pin from `AsGDExtensionConstStringNamePtr` / `AsGDExtensionConstStringPtr` — which looked like ordinary call scratch — panicked at startup under `cgocheck=1` with `argument of cgo function has Go pointer to unpinned Go pointer`. Registration code stores these pointers *inside* Go structs (`GDExtensionPropertyInfo`, argument-info arrays) that are then passed to C, so the pointee of the cgo argument contains Go pointers that must themselves be pinned. The accessors keep their package pin; the generated per-call path uses the new `AsGDExtensionConstStringNamePtrPinned(*runtime.Pinner)` / `AsGDExtensionConstStringPtrPinned`, where the pointer is a direct argument to a synchronous call that copies it, so the caller's pinner owns the lifetime.

The generalisation worth keeping: **the pinning question is not "is this value short-lived" but "what does C do with the pointer, and is it reachable through another Go pointer at the moment of the crossing."** Nested-in-struct crossings have strictly stricter requirements than direct-argument crossings.
