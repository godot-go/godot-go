# Design

## Context

godot-cpp master is the reference and this change mirrors it deliberately.
Verified against upstream (`src/godot.cpp:218-220`, `include/godot_cpp/godot.hpp`
@ master):

- `InitObject` exposes `register_startup_callback` / `register_frame_callback`
  / `register_shutdown_callback`, each a single assignment into an
  `InitData.main_loop_callbacks` struct field (no user_data argument anywhere
  — all three C callbacks are `void(void)`).
- Registration with the engine happens inside `initialize_level` when
  `p_level == CORE` and at least one of the three fields is non-null
  (`has_main_loop_callbacks()`), passing the library pointer and the struct.
- No recovery semantics: a throwing hook crosses the engine boundary.

godot-go current state (`pkg/core/godot.go`, `pkg/core/lib.go`):

- `_GDExtensionBindingInit` holds `pLibrary`; `FFI.LoadProcAddresses` retains
  it as `FFI.Library`. Init/terminate dispatch uses package-level arrays
  (`GDExtensionBindingInitCallbacks`, `GDExtensionBindingTerminateCallbacks`)
  from behind exported level callbacks.
- The FFI layer is already complete for this feature:
  `GDExtensionMainLoopCallbacks` (`pkg/ffi/ffi_wrapper.gen.go:4972`),
  `CallFunc_GDExtensionInterfaceRegisterMainLoopCallbacks` (:4938), and the
  `register_main_loop_callbacks` loader entry (`ffi.gen.go:205,392`). On
  pre-4.5 engines the loader logs the known missing-function warning and
  leaves `FFI.RegisterMainLoopCallbacks` nil.
- `GDExtensionBindingCallback` (`func()`) already exists
  (`pkg/builtin/lib.go:20`) and types all eight existing
  `InitObject.RegisterXxxInitializer/Terminator` methods.

See proposal.md for motivation; the delta specs for requirements.

## Goals / Non-Goals

**Goals:**
- API and timing that a godot-cpp reader recognizes 1:1 (three `InitObject`
  methods, CORE-level registration, single slot per phase).
- Zero codegen: consume the generated FFI surface as-is.
- `make test` proves all three phases end-to-end, including teardown.

**Non-Goals:**
- Fan-out/multiple subscribers, deregistration, per-hook panic isolation
  (proposal Non-goals).
- Scheduling infrastructure for C2 — the frame hook stays a plain pass-through.
- Wiring anything at non-CORE levels; the engine's registration is per-library
  and phase-less.

## Decisions

### D1: Three `InitObject` methods over package-level functions
`RegisterStartupCallback(fn GDExtensionBindingCallback)`,
`RegisterFrameCallback(...)`, `RegisterShutdownCallback(...)` on `InitObject`,
symmetric with the eight existing register methods and with godot-cpp's
`InitObject::register_*_callback`. Alternative: package-level `core.Register…`
callable anytime — rejected: it invites post-`Init()` calls that can no longer
reach the engine (see D3) and breaks the established registration-before-Init
contract.

### D2: Package-level slot variables
The engine's callbacks carry no user_data, so dispatch must resolve to
package state. Three nullable `GDExtensionBindingCallback` slots in `pkg/core`
(next to `GDExtensionBindingInitCallbacks` in `lib.go`), assigned by D1's
methods. Alternative: stash them on an `InitData`-equivalent reachable via
`r_initialization` user data — godot-go's level dispatch doesn't use user data
(`SetUserData(nil)`), so introducing it just for this would be a foreign
pattern for zero gain.

### D3: Register at CORE level inside `GDExtensionBindingInitializeLevel`
Mirror `godot.cpp:218`: when `pLevel == CORE` and any slot is non-nil, build
the `GDExtensionMainLoopCallbacks` struct and call
`CallFunc_GDExtensionInterfaceRegisterMainLoopCallbacks(FFI.Library, &cbs)`.
Struct fields receive an exported trampoline pointer only for phases whose
slot is non-nil; unfilled phases stay nil, exactly like godot-cpp's struct.
A local struct is safe: the C signature takes `const *` and the engine reads
the three raw pointers synchronously into its own storage.
Consequence (same as godot-cpp): assignments after the CORE level have
already run do not reach the engine; hooks are register-before-`Init()`.
Alternative: register during `_GDExtensionBindingInit` — works with the engine
but diverges from reference timing and would run before user code has
necessarily finished hook setup ordering assumptions.

### D4: Three `//export` trampolines in a new `pkg/core/main_loop_callback.go`
`GoCallback_MainLoopStartup/Frame/Shutdown` follow the `GoCallback_*` pattern
in `classdb_callback.go`: nil-check the slot, invoke if set, otherwise return.
Always-compiled and referenced only from the struct built in D3. The exported
name must be referenced from C via the cgo-generated prototypes (same as the
other `GoCallback_*` fns used in class registration).

### D5: No panic recovery
No `recover()` anywhere in the trampolines — consistent with every existing
`GoCallback_*` trampoline (some even `log.Panic` across the boundary) and with
godot-cpp's unrecovered calls. Policy documented in `docs/overview.md`.

### D6: Runtime version guard replaces godot-cpp's `#if >= 5`
Skip registration when `FFI.RegisterMainLoopCallbacks == nil`, logging at
debug level (the loader already warned about the missing function). Keeps the
repo's 4.3+ support intact; no compile-time branching.

### D7: Demo fixtures and shutdown marker
`test/pkg/example.go` gains a fixture class exposing to GDScript: startup
flag, frame count, and a reset — hooks record into plain Go atomics (no engine
objects created inside hooks, avoiding teardown races with the leak check).
The shutdown hook logs a unique literal marker (e.g.
`GODOTGO_MAIN_LOOP_SHUTDOWN`) at a level that survives the harness
`LOG_LEVEL=WARN` filter. `test/check_test_output.sh` gains a gate requiring
that literal, with its own distinct failure message (delta spec).

### D8: Multi-iteration harness run (`--quit-after 5`) and primed frame awaits
Engine facts (`../godot main/main.cpp`): `--quit` sets `quit_after = 1`, so
`Main::iteration` runs exactly once; the extension frame callback fires near
the end of an iteration (`GDExtensionManager::frame()`, line 5135) and the
quit check follows it (line 5188). `SceneTree::process` fires `process_frame`
before that hook, and an awaiting coroutine resumes inside the dispatch.
Consequences: under `--quit` the frame hook is invisible to GDScript and a
second `await process_frame` never resumes (banner would be lost). The test
target therefore runs `--quit-after 5`, and the GDScript test primes with one
await (absorbing the resume-before-hook boundary) and then measures per-await
deltas of at least one, which satisfies the spec scenario literally. Five
iterations leave margin: the coroutine completes by iteration 3.

## Risks / Trade-offs

- [Frame hook adds a cgo hop per frame] → Only wired when a hook is
  registered; an empty slot check is negligible per frame; documented cost is
  one no-arg cgo callback (~tens of ns).
- [`--quit-after 5` leaves a fixed iteration budget] → The coroutine
  completes by iteration 3 of 5; if the driver ever needs more frames the
  run fails loudly as "no driver summary", never as a false pass (D8).
- [Marker check makes every future unrelated `make test` run depend on the
  fixture] → Intended per delta spec; the failure message names the missing
  marker so diagnosis is one step.
- [Hook panics abort the process] → Accepted for parity (D5); a panicking
  hook is a bug the current binding surfaces identically everywhere else.
- [Engine behavior on double `register_main_loop_callbacks` per library is
  unspecified] → Never triggered: registration runs once per process at CORE
  level.

## Migration Plan

Purely additive; no caller migration. Rollback = revert the commit(s).

## Open Questions

None blocking. (Exact marker string and hook fixture method names are
implementation details resolvable during tasks.)
