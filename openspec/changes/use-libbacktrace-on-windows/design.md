# Design

## Context

See `proposal.md` for motivation. The current C capture seam is `pkg/log/stacktrace.c`:

- `printStacktrace()` is called at every GDExtension callback boundary, guarded by `enablePrintStacktrace` (default false).
- Linux/macOS use `execinfo`; Windows uses DbgHelp behind `GODOT_GO_ENABLE_DBGHELP`.
- Windows additionally carries a `-gcodeview` directive in `pkg/log/lib.go` and lld `/DEBUG --pdb` flags in the `Makefile`, solely so DbgHelp can print `file:line`.

Godot 4.7-stable vendors a PE/COFF-only libbacktrace at `thirdparty/libbacktrace` and uses it in `platform/windows/crash_handler_windows_signal.cpp` for MinGW builds. Its `read.c` is patched to `_lseeki64`/`_read`, so the vendored copy is Windows-only by construction. Godot's source is already patched for large files (`patches/0001-big-files-support.patch`).

Constraints:

- The extension is built with mingw (`GOOS=windows`, cgo), not MSVC.
- The module is now pinned in the process (`pkg/core/module_windows.go`), so a leak-on-purpose state is safe for the process lifetime.
- The existing capture evidence chain is `test-output.log` → `test/check_test_output.sh` + the Windows CI evidence step.
- `backtrace_simple` in libbacktrace uses `_Unwind_Backtrace`; the project already observed that unwinding across Go/C frames is the risky part, while `CaptureStackBackTrace` is proven in CI.

## Goals / Non-Goals

**Goals:**

- Windows capture uses the same libbacktrace implementation as Godot's MinGW build.
- C frames in the extension DLL print `function file:line` live; foreign frames stay actionable offline.
- Remove DbgHelp, the Windows opt-in, CodeView/PDB generation, and the lld link dependency.
- Keep Linux/macOS/musl behavior unchanged.

**Non-Goals:**

- libbacktrace on Unix (proposal scope A only).
- Symbolizing engine/GDScript frames or adding a crash handler.
- Resolving Go frames through libbacktrace.
- Upgrading or patching libbacktrace beyond Godot's 4.7-stable copy.

## Decisions

### D1: Vendor Godot 4.7-stable's PE/COFF libbacktrace, not full upstream

Copy the libbacktrace sources, headers, `config.h`, and `LICENSE` from Godot's `thirdparty/libbacktrace` into the project. Set `HAVE_ZLIB`/`HAVE_LIBLZMA` to 0: the build produces uncompressed C DWARF, and `pecoff.c` does not map `.zdebug_*` anyway, so the compression code would only add link dependencies. Keep the big-files patch already present in Godot's sources.

Omit the unwind entry points (`backtrace.c`, `simple.c`, and `print.c`). Only `backtrace_pcinfo` is used; keeping the other entry points links `_Unwind_Backtrace` and, under LLVM-MinGW, adds a runtime `libunwind.dll` import the extension does not need.

Alternatives:

- Full upstream libbacktrace with ELF/Mach-O (`elf.c`, `macho.c`, …) — needed only for scope B; much larger and needs per-OS `config.h`.
- Link a system libbacktrace — none is installed or linkable on mingw.
- Keep DbgHelp — rejected by the proposal.

### D2: Hybrid capture: keep `CaptureStackBackTrace`, use libbacktrace for resolution

Capture frames with the current `CaptureStackBackTrace(0, 25, ...)` walk, then resolve each PC with `backtrace_pcinfo`. Do not use `backtrace_simple`/`backtrace_full` for the walk.

Rationale: `backtrace_simple` delegates to `_Unwind_Backtrace`, which is exactly the Go/C boundary risk the project has avoided. `CaptureStackBackTrace` is already proven in Windows CI and walks through Go frames. libbacktrace is used for what it is good at here: DWARF line/function lookup.

Consequence: the `backtrace_simple`/`backtrace_full` entry points and their unwind implementation are not vendored at all; the library is present only for DWARF symbolization.

### D3: One lazily-created state for the extension DLL

On the first enabled capture:

1. `GetModuleHandleExA(GET_MODULE_HANDLE_EX_FLAG_FROM_ADDRESS | GET_MODULE_HANDLE_EX_FLAG_UNCHANGED_REFCOUNT, (LPCSTR)(const void *)&enablePrintStacktrace, &module)`.
2. `GetModuleFileNameA` into a static buffer (the buffer must outlive the state; `backtrace_create_state` requires a permanent filename).
3. Read `SizeOfImage` from the mapped PE header, and the preferred `ImageBase` from the DLL file on disk. Windows rewrites the mapped `OptionalHeader.ImageBase` to the actual load address, while the DWARF addresses in the file use the preferred base; using the mapped value would make the delta zero and every lookup miss.
4. `backtrace_create_state(path, 1, error_callback, NULL)` once, under a lock.
5. Store `runtime_base = (uintptr_t)module` and `delta = runtime_base - preferred_image_base`.

For each captured PC:

- If it is inside `[runtime_base, runtime_base + SizeOfImage)`, call `backtrace_pcinfo(state, pc - delta, ...)`.
- Otherwise, skip libbacktrace and print `module+0xoffset (ip 0x...)` via `GetModuleHandleExA`/`GetModuleFileNameA`.

The state is intentionally never freed: libbacktrace has no free API, and the module is pinned on Windows. Using `threaded=1` makes the state safe across engine threads; keep a lock around the one-time create.

Alternatives:

- One state per loaded module — engine modules have no debug info in release builds and add complexity; rejected.
- A state for the Godot executable — resolves the wrong module's frames; rejected.
- `threaded=0` — callbacks can fire on multiple engine threads; rejected.

### D4: cgo packaging: dedicated package first, sources beside `stacktrace.c` as fallback

cgo compiles only `.c` files in the package directory. Place the vendored sources in `pkg/log/libbacktrace/` with a Windows-only Go wrapper, and add a Windows-only blank import in `pkg/log` so the package archive (and its C objects) is linked. The wrapper's `init()` calls a tiny C anchor that references `backtrace_create_state`, which forces the object to be pulled even if no Go code calls it. `stacktrace.c` keeps calling the C `backtrace_*` symbols directly.

The first implementation task validates this link. If cgo cannot resolve C symbols across packages, move the sources next to `stacktrace.c` with `_windows.c` suffixes and delete the wrapper package. This is a build-mechanics fallback, not a behavior change.

### D5: Remove the DbgHelp/CodeView/PDB path entirely

- `pkg/log/stacktrace.c`: delete the DbgHelp branch, `-ldbghelp`-dependent includes, and the `GODOT_GO_ENABLE_DBGHELP` gate.
- `pkg/log/lib.go`: remove `#cgo windows LDFLAGS: -ldbghelp` and `#cgo windows CFLAGS: -gcodeview`; add the libbacktrace include path.
- `pkg/log/stacktrace.h` and `pkg/core/stacktrace.go`: update comments; `stacktraceCaptureAvailable()` returns true on Windows.
- `Makefile`: remove `DEBUG_INFO_EXTLDFLAGS` and `GO_LINK_LDFLAGS` (they exist only for the PDB). Keep `-ggdb` so C TUs carry DWARF.
- `.github/workflows/ci_windows.yaml`: remove the lld PATH check and PDB upload; expect `capture=1`, a banner, and a `file:line` frame; assert the DLL import table has no `dbghelp.dll`.
- `README.md`: replace the DbgHelp opt-in and PDB/`llvm-symbolizer` instructions with the libbacktrace behavior and the existing `go tool addr2line` offline recipe.

Alternative: keep DbgHelp as a fallback when libbacktrace resolution fails — rejected: it keeps the dependency and the PDB machinery for a path the new resolver is expected to cover.

### D6: Accept compressed Go DWARF; do not force `-compressdwarf=0`

On PE, Go's linker compresses its own DWARF into `.zdebug_*`, which `pecoff.c` does not map. C DWARF from gcc remains `.debug_*` and is resolvable. Go frames already print through `runtime.Stack` in the probe, so this is a non-loss.

`build-full` keeps its existing `-compressdwarf=0`; the normal `build` target does not need to change. If future work wants libbacktrace to resolve Go frames too, the options are to map `.zdebug_*` in `pecoff.c` or add `-compressdwarf=0` to `build`.

### D7: Compile Windows capture in by default

`stacktraceCaptureAvailable()` reports true on Windows, and `printStacktrace()` still produces nothing until `enablePrintStacktrace` is set. The old opt-in existed to avoid loading `dbghelp.dll` during the teardown investigation; issue #172 is fixed by pinning the module, and libbacktrace does not load a debugger DLL.

## Risks / Trade-offs

- **libbacktrace cannot resolve the DLL's mingw DWARF** → This is the make-or-break assumption. The first task is a minimal Windows build that prints one resolved C frame; if it fails, stop and revisit (DbgHelp fallback or a newer libbacktrace) rather than landing a half-working resolver.
- **Cross-package C linking fails** → D4's fallback moves the sources beside `stacktrace.c`; no spec or API change.
- **Go frames do not resolve** → Accepted; `runtime.Stack` is the Go-frame source. The CI evidence only requires one C `file:line`.
- **State creation runs on a callback thread** → One-time lock; `threaded=1`; the filename buffer is static.
- **Non-ASCII DLL paths** → `GetModuleFileNameA` can truncate. CI paths are ASCII; if this matters later, switch to `GetModuleFileNameW` plus a UTF-8 conversion. Deferred.
- **Binary size grows** → ~240 KB of vendored source plus the DWARF parser; acceptable for a debug-aid dependency.
- **Vendored code maintenance** → Pin the exact Godot 4.7-stable snapshot, retain `LICENSE`, and record the upstream version in the vendored directory.
- **CI artifact consumers expect a PDB** → Only the Windows test workflow consumed it; the step is updated in the same change.

## Migration Plan

- Not a deployed service. Rebuild the extension; Windows consumers no longer pass `GODOT_GO_ENABLE_DBGHELP`.
- No public Go API change: `SetPrintStacktrace` and `StacktraceCaptureAvailable` keep their signatures.
- Rollback: revert the change; the old DbgHelp opt-in path returns, including the PDB build flags.
- Order: land the build spike first, then the resolver, then the build/CI cleanup, then docs and the spec sync.

## Open Questions

None blocking. The cgo placement and DWARF readability are resolved by the first implementation task, not by deferring a decision.
