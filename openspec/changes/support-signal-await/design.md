# Design

## Context

C1 shipped `core.NewGoCallable(fn) Callable` with a registry whose free callback fires exactly when the engine drops the last reference (`GoCallableRegistryCount` is the observable). The generated `Object` interface exposes `Connect(signal StringName, callable Callable, flags uint32) Error` / `Disconnect`, and `constant.OBJECT_CONNECT_FLAGS_CONNECT_ONE_SHOT = 4` exists. Refcounted `Variant`s are unsafe to hand across goroutines (engine refcounts are not thread-safe; callback args are borrowed views that die when the callback returns). The app-facing frame/shutdown callbacks in `pkg/core` are single-slot last-write-wins — this design deliberately avoids claiming them.

Reference designs: gdext `task/` (`SignalFutureResolver` + main-thread pinning; resolver `Drop` = freed-object signal) and the engine's own GDScript await (`CONNECT_ONE_SHOT` connect of the coroutine state). Both were reviewed during exploration; see the change discussion.

## Goals / Non-Goals

**Goals:**
- `Await` is safe to compose with `select`, `context`, and `for range` with zero engine knowledge at the call site.
- Every awaiter path — emitted, object-freed, engine-shutdown — ends in a closed channel with no leaked Callable registration.
- Zero new engine-loop hooks; zero observable behavior change to C1 (the registry gains an internal, non-exported `onFree` notice invoked by the free callback — discovered as necessary during implementation, since the liveness path must reach the await-state when the connection dies; `NewGoCallable` and all C1 spec behavior are untouched).

**Non-Goals:**
- No hop-back-to-main (`PostToMain`); no `Callable.call_deferred` helper (future round-trip change).
- No multi-shot subscriptions with user-visible handles; no cancellation API beyond channel semantics.
- No enforcement of main-thread calling (godot-go generates no thread checks anywhere; see Risks).

## Decisions

### D1: API shape — `Await(obj Object, signalName string) <-chan AwaitResult`
`type AwaitResult struct { Args []any }`. Delivery is send-then-close; the freed-object path is close-without-send. This lets `v, ok := <-ch` distinguish "fired with zero args" (`ok == true`, empty `Args`) from "never fires" (`ok == false`), which a bare `chan []any` cannot. Alternatives rejected: `chan Variant` (violates the Go-native payload spec requirement), generic `Await1[T]` (deferred; v1 keeps one untyped shape).

The parameter is the generated `Object` interface — it carries `Connect`/`Disconnect`, every class Impl satisfies it, and `Ref*` wrappers reach it via `.Ptr()`. A structural mini-interface was rejected in favor of the project's canonical type.

### D2: The engine owns the only Callable reference
After `obj.Connect(name, callable, CONNECT_ONE_SHOT)` succeeds, the Go side immediately calls `callable.Destroy()`. The one-shot connection then holds the only reference, which is what makes liveness work:
- **Emitted:** engine disconnects, invokes the callback, releases its copies → free callback fires → registry drains. The callback delivers via the channel; it owns no Callable reference to release.
- **Freed object:** engine drops the connection → last reference gone → free callback fires → the callback's shared await-state closes the channel (no send).

Any surviving Go reference would keep the free callback silent on object-free, breaking the liveness oracle — hence "connect and immediately destroy" is load-bearing, not cleanup. (C1's ownership model guarantees this: the engine received its own copy at connect; destroying the Go handle is assignment-semantics safe.)

### D3: Shared await-state guarded by `sync.Once`
Emission-delivery and free-driven-close can both fire (free may run right after delivery once the engine's copies die). Both paths close through one `sync.Once`; delivery does `ch <- res` (buffered, cap 1) before `close(ch)`, so the send never blocks and the Once makes double-close impossible. No mutex is needed for the payload: exactly one writer path can send.

### D4: Marshalling runs on the emitting thread, into Go-owned values
Inside the C1 callback the borrowed arg views are converted immediately (gdext does the same in its resolver): `NIL`→nil, `BOOL`→`bool`, `INT`→`int64`, `FLOAT`→`float64`, `STRING`/`STRING_NAME`/`NODE_PATH`→Go `string`, plain-data math built-ins (vectors, color, rect, basis, plane, quaternion, aabb, RID) → value bit-copied into their Go struct equivalents (no refcount involved), objects → `Object` interface handles (handle may cross threads; *calling* stays main-side by existing project convention), `ARRAY`→`[]any` and `DICTIONARY`→`map[any]any` recursively through the same table (GDScript forbids container keys, so keys stay comparable), and everything exotic (`Callable`, `Signal`, string-ish char types) → `AwaitOpaque{ TypeName string; Text string }` built from `Typeof` + `Stringify()`. Alternatives rejected: passing raw `Variant` copies for refcounted types (re-arms the landmine), and `Stringify()`-ing everything (lossy).

The `Object` interface path uses the existing binding lookup (`ObjectFromVariant`); POD structs use the type's Go layout directly. Implementation discovered Variant stores POD math types in two classes — engine-heap pointer slots (all oversized types) and inline payloads (all types that fit inline). Bit-copying the slot of a heap-stored type would hand out a dangling pointer (the landmine itself), so `awaitHeapPOD` dereferences and copies the pointee during emission; `awaitInlinePOD` bit-copies the inline payload at offset 8 (engine layout: 4-byte type tag + pad + 16-byte union, heap-stored types larger than 16 bytes). Layouts are validated by the harness scenario, not assumed.

### D5: Call-site misuse fails loud
`Connect` returning non-OK (unknown signal, invalid object) → `log.Panic` naming object class + signal name, matching `NewGoCallable`'s nil-arg precedent and the generated code's methodbind-miss panic style. A silently unresolvable channel is the worse failure mode (spec: Misuse).

### D6: Shutdown needs no new machinery
Engine teardown frees objects, which drops connections, which fires free callbacks — C1's suite already demonstrates the registry draining to zero at unload. Pending awaiter goroutines unblock via close, or simply die with the process (Go exit never waits on goroutines), so shutdown cannot deadlock. A shutdown-callback fan-out was rejected: it would require changing the single-slot app-callback semantics (the exploration landmine) to solve a problem the reference-drop path already solves.

## Risks / Trade-offs

- **Threaded emissions:** conversion reads borrowed arg memory on the emitting thread. Safe for the duration of the call (engine keeps args alive during emit; same assumption as gdext), but a signal emitted from a `Thread` delivers payloads on a non-main thread and post-await engine calls then follow the project's general main-thread hazard. → Documented in the API doc comment and `docs`; v1 tests stick to main-thread emission.
- **`Await` itself must be called on a usable thread:** `Connect` mutates engine object state, so `Await` off the main thread is the same hazard as any generated method called off-thread. No runtime check exists project-wide; we stay consistent rather than inventing one here. → Documented; a binding-wide main-thread assert is a separate future capability.
- **ONE_SHOT ordering:** whether the engine disconnects before or after invoking the callback changes only *when* the free callback fires, not the outcome — both paths funnel through the `sync.Once`. → Verified by registry-count scenarios in tests.
- **Recursive `Array`/`Dictionary` payloads:** deep structures copy deep (cost O(size), and self-referential containers would recurse forever). → Depth/size cap with an `AwaitOpaque` marker past the cap (cap chosen at implementation, ~64).
- **Opaque fallback is lossy for Callable/Signal payloads:** accepted for v1; typed payload conversion is the natural follow-up change.

## Open Questions

- Whether v2 adds a `Subscribe` (multi-shot) primitive on the same await-state machinery — intentionally left unspecified; nothing in v1 blocks it.
