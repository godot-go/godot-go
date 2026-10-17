# Proposal

## Why

GDScript code uses `await obj.signal` to write linear wait-for-event logic; godot-go code today can only attach callbacks, and hand-rolling a channel over the C1 `NewGoCallable` primitive has no lifetime rules and risks engine-owned memory crossing goroutines. `docs/overview.md` promises "goroutines to wrap signal and callable" — C1 delivered the callable half, and roadmap C2 is the await half. gdext proves the pattern is tractable; Go's channel idiom makes it simpler.

## What Changes

- New `core.Await(obj Object, signalName string) <-chan AwaitResult`: a one-shot await bridge. A goroutine blocks on `<-core.Await(...)` with the same reading as GDScript's `await`.
- Emission payload is marshalled to Go-native values on the emitting thread (POD bit-copies, `String`/`StringName` → Go `string`, `Object` → interface handle, `nil` → nil). No engine-owned refcounted Variant ever crosses the channel.
- Liveness via C1's free callback: if the signal's object is freed before emission, the engine drops the one-shot connection, the Go callable's free callback fires deterministically, and the channel is closed without a send — the receiver never hangs (`v, ok := <-ch` yields `ok == false`).
- Standard Go `select` composes timeouts (`time.After`), cancellation (`context`), and racing multiple signals for free.
- Document the L1 pattern (Go closure → `Connect` on a signal, already possible since C1) alongside the new `Await` API.

## Non-goals

- No main-thread hop-back (`PostToMain`, `Callable.call_deferred` helper): engine access after the await stays the caller's responsibility (documented hazard). A future round-trip change can add it.
- No multi-shot `Subscribe`/`Cancel` subscription API; loop-and-reawait covers repetition in v1.
- No typed generic payloads; v1 delivers `[]any`.
- No benchmarks: performance work is deferred until feature-complete (project decision).

## Capabilities

### New Capabilities
- `signal-await-bridge`: one-shot engine-signal → Go-channel bridge: connection lifecycle, emission marshalling to Go-native values, deterministic close semantics.

### Modified Capabilities
- (none — `custom-callable-binding` requirements are unchanged; this capability consumes them as-is)

## Impact

- New `pkg/core/signal_await.go` + tests (`pkg/core/signal_await_*_test.go`).
- Consumes: `custom-callable-binding` (C1 `NewGoCallable` + free callback), generated `Object.Connect`/`Disconnect` (`CONNECT_ONE_SHOT`), `Variant.GetType` and existing builtin decode helpers.
- Test demo: GDScript-driven scenarios in `test/demo/main.gd` + assertions in `test/pkg/signal_await_tests.go`.
- `docs/roadmap.md`: C2 status update at completion.
