# Design

## Context

`printStacktrace()` runs at every `cgo_callfn_*` callback entry guarded only by the `enablePrintStacktrace` global. On Windows its branch is commented-out MSVC-style scaffolding. CI builds Windows with mingw-w64 (gcc), not MSVC, and links nothing beyond the default mingw libs. The test harness evidence chain is stdout -> `test/check_test_output.sh` -> both CI jobs; the suite already carries a fixture table for the gate itself (`make test_gate`).

## Goals / Non-Goals

**Goals:** Windows prints a real native callstack when enabled; the capture is verified in-suite on both CI platforms with ~2 banners of log volume; unresolved frames stay addressable offline.

**Non-Goals:** No PDB generation or `-gcodeview` build-flag surgery; no symbolization of DWARF on Windows at runtime; no change to the glibc/musl capture behavior beyond the additive predicate; no global default-on capture (log spam).

## Decisions

### D1: DbgHelp walk with SRWLOCK guard, replacing the MSVC scaffolding

`RtlCaptureContext` -> `SymInitialize(process, NULL, TRUE)` -> `StackWalk64` loop (cap 25) -> per-frame `SymFromAddr`, then `SymCleanup`. The whole sequence is wrapped in a process-wide `SRWLOCK` (statically initializable via `INIT_SRWLOCK`, no init call): the DbgHelp API set is not thread-safe and callbacks fire on whatever engine thread runs them; a racy crash during a debug window would poison CI nondeterministically. The scaffolding's IA64 branch is dropped (dead platform, was never compiled); amd64 and x86 branches are implemented, keyed on both `_M_X64`/`_M_IX86` and mingw's `__x86_64__`/`__i386__` macro families.

- Toolchain gate found by the first Windows CI compile: mingw hides `INIT_SRWLOCK`/`AcquireSRWLockExclusive` behind `_WIN32_WINNT >= _WIN32_WINNT_VISTA` and its default can be lower, so the TU pins `_WIN32_WINNT`/`WINVER` to `0x0601` before including any system header.

- Alternative: `RtlCaptureStackBackTrace` (ntdll) — simpler but skips the symbol layer entirely; the existing scaffolding and the glibc path's shape (names when possible) argue for DbgHelp.

### D2: Unresolved frames print `module.dll+0xoffset (ip 0x…)`

The system `dbghelp.dll` cannot read mingw's DWARF, so `SymFromAddr` will usually fail and the scaffolding's bare `???` is useless offline. `SymGetModuleBase64` + `GetModuleFileNameA` give the containing module; printing `basename+offset` plus the raw instruction pointer makes every frame resolvable with addr2line/gdb. When a PDB does exist (MSVC-built users), names print as `name+0xdisp`.

### D3: Link via `#cgo windows LDFLAGS: -ldbghelp`

Placed in the preamble of the Go file that includes `stacktrace.h` (`pkg/core/godot.go`), so the dependency travels with the code that needs it on every build path, not just the Makefile. GOOS-conditioned by cgo itself; no Makefile change.

### D4: Availability predicate instead of platform guessing in tests

`bool stacktraceCaptureAvailable(void)` in `stacktrace.c` mirrors the compile-time cascade; exported through `pkg/core` Go helpers. The probe line reports the truth from the binary, so the gate's conditional rule ("claim requires banner") works on every libc without the shell knowing platforms.

### D5: Three-call probe window with phase strings

`Example.test_stacktrace_probe(phase)`: `"start"` prints `stacktrace-probe capture=<0|1>` (direct `fmt.Printf`, not zap, so the banner text is exactly greppable), then sets the C flag; `"hit"` returns `"hit"` — its callback entry emits the banner when available; `"stop"` clears the flag and returns `"off"`. GDScript asserts the three returns (platform-neutral: passes identically on stub platforms) and the gate carries the platform-conditional expectation. Measured host volume: 26 banners for the window (the hit/stop callbacks' own engine FFI crossings fire the wrapped callfn sites too), produced inside three synchronous calls — still bounded, no spam.

- Alternative: enable for the full run at LOG_LEVEL=debug — rejected: log spam and it couples a C debug aid to the logger level again (the coupling the upstream commented line had).

### D6: Gate rule placement and strictness

Rules appended after the main-loop-shutdown check and before the catch-all `godot-exited-non-zero` check, so a real stacktrace failure reports as itself instead of being masked. Probe-absent is a hard failure (the driver is the suite; a silently vanishing section is drift); `capture=1`-without-banner is a hard failure (the only in-repo proof Windows/Linux capture actually prints). `capture=0` without banner passes (musl stub). Fixtures: three new logs + updates to the two pass-path existing fixtures; the six earlier failure fixtures exit before these rules and are untouched.

## Risks / Trade-offs

- [SymInitialize/Cleanup per call is heavyweight] → acceptable: debug aid, narrow probe window; the per-call form matches the existing scaffolding.
- [`--quit-after 12` teardown races the probe window] → window is three synchronous calls inside `_ready`; teardown happens at scene-tree exit.
- [mingw `printf` format portability for `%llx`] → cast all `DWORD64` values to `unsigned long long`; UCRT supports it.
- [No local mingw on dev hosts] → Windows compile verification runs on CI; glibc path and gate fixtures verify locally first.
