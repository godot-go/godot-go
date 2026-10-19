# Design: Windows PDB line info

## Context

`windows-stacktrace-capture` shipped the DbgHelp walk with module+offset frames because the only debug info the mingw link produced (DWARF) is invisible to `dbghelp.dll`. This change adds the missing half: CodeView/PDB, so the same walk can print live source lines.

## Decisions

### D1: CodeView is a per-package opt-in via `#cgo windows CFLAGS`, not a global flag

GCC's debug-info formats are an exclusive selection (`debug_type_masks[]` in `opts.cc`); the last `-g*` format flag on the command line wins. The original design switched `GOOS=windows` build-target C flags from `-ggdb` to `-gcodeview` globally, and that died when the runner's compiler ICEd on generated TUs (D6). The build's `CGO_CFLAGS` now keep `-ggdb` everywhere, and only `pkg/log` — the capture code, whose `printStacktrace` frame is always frame[0] of a captured backtrace — appends `-gcodeview` via a `#cgo windows CFLAGS:` directive; cgo places package-directive flags after the environment flags, so the directive wins for that package. `pkg/core` (the `GoCallback_*` glue) tried the same directive and was reverted: every one of its cgo preamble TUs ICEd (D6). Its frames keep the symbol-only form, and their lines stay resolvable offline. Consequence, accepted: all remaining C TUs keep DWARF line tables inside the DLL — they print in the symbol-only/module+offset forms live and stay resolvable offline with standard DWARF tools against the DLL itself. Go portions keep their own DWARF/gopclntab because `gc` does not read `CGO_CFLAGS`, so the existing `go tool addr2line` recipe still resolves Go frames offline.

### D2: The PDB sits next to the library and shares its basename

ld embeds the `--pdb=` string verbatim into the CodeView RSDS debug record, and writes the file relative to the linker's CWD — which during `go build` is the repo root, not the output directory. So the flag passes the explicit `$(CURDIR)/$(OUTPUT_PATH)/<basename>.pdb`: on the build machine (CI runs there too) the embedded absolute path hits directly, and dbghelp's documented fallback — search the image's own directory by PDB basename — covers relocated use such as CI artifacts, where the upload keeps the PDB beside the DLL. An absolute build-machine path in the record is therefore deliberate, not sloppiness; a directory-free name would land the file in the wrong place at link time. A fresh CI checkout surfaced one more wrinkle: the PDB write happens inside the external link, before `go build` creates the `-o` destination's parent directory, so `ld` failed with `warning: cannot create PDB file: No such file or directory` and — being only a warning — linked on with no PDB and no debug directory entry. The `build` target therefore pre-creates the output directory (`mkdir -p`); the evidence gate is what caught this silent degrade.

### D3: Frame format gains a trailing ` file:line` only on line lookup success

```
[0] printStacktrace+0x42 stacktrace.c:86
[3] GoCallback_MethodBindMethodPtrcall+0xd48
[5] libgodotgo-test-windows-amd64.dll+0x987ad (ip 0x7fff900f87ad)
```

Frames whose line is unknown — Go runtime, engine DLLs, and the C TUs that stay on DWARF (`pkg/core` glue, `pkg/ffi`) — print exactly as before, and existing gate/verify anchors (banner lines, probe markers) are untouched. `IMAGEHLP_LINE64` needs only `SizeOfStruct`, `FileName`, `LineNumber`; `SymSetOptions(… | SYMOPT_LOAD_LINES)` is already in place and must stay before `SymInitialize`.

### D4: Toolchain floors are probe-verified, and the probe caught a real ICE

`--pdb=` requires binutils >= 2.40 (patch series merged October 2022); `-gcodeview` requires GCC >= 14 (not 13 — GCC 13 rejects it as an unrecognized debug output level). The `windows-latest` runner ships MinGW-Builds GCC 15.2.0 (`x86_64-posix-seh-rev1`) with binutils far above both floors — established by the runner diagnostics, and no version sniffing is added to the Makefile: an ld too old for `--pdb` fails the link loudly rather than degrading silently, and the verify step (D5) catches debug-info regressions that link clean. The same diagnostics surfaced a compiler crash that shaped the capture code — D6.

### D6: The Win32 walk uses `CaptureStackBackTrace` because dwarf2codeview ICEs on `STACKFRAME64`

The first `-gcodeview` build segfaulted the compiler (`stacktrace.c:178:1: internal compiler error: Segmentation fault`) while every other cgo TU compiled fine with the same flags. A minimal-TU bisect on the runner (GCC 15.2.0) pinned the trigger: an empty function, a function-`static` `SRWLOCK`, and a local `CONTEXT` (even with `RtlCaptureContext`) all compile, but a function declaring a local `STACKFRAME64` — which mingw declares with nested anonymous `DUMMYUNIONNAME` unions — reliably segfaults the CodeView writer, independent of optimization level (`-O1` crashes too; `-g1` instead emits assembly gas cannot resolve: `can't resolve .debug$T`). This is the same GCC `dwarf2codeview` ICE family as the empty-struct and artificial-variable fixes. The response is structural, not flag-surgical: `printStacktrace()` calls `CaptureStackBackTrace(0, 25, …)` for a plain array of return addresses and resolves each with the unchanged `SymFromAddr` / `SymGetLineFromAddr64` / module+offset chain — no `STACKFRAME64`, no `CONTEXT`, and the x64/x86 register-selection `#if`s disappear. The replacement shape was pre-validated as a minimal TU on the very runner that crashed. Accepted behavior deltas: frame[0] is `printStacktrace` itself (the seeded `StackWalk64` emitted the same self-frame), and return addresses walk fine through optimized code where an RBP chain would stall.

The next run proved the writer is fragile beyond our hand-written TU: with `stacktrace.c` fixed, `-gcodeview` then ICEd on `pkg/ffi/ffi_wrapper.gen.c:1842` and on `pkg/ffi`'s cgo export prologue (`cgo-gcc-prolog:4890`) — generated code that cannot be restructured, while every other hand-written cgo TU compiled clean (2 ICEs out of 34 compiles). No flag shape avoids this class of bug, so the response generalizes into the per-package opt-in of D1 rather than further TU surgery. The subsequent opt-in attempt narrowed the safe set further: with only `pkg/log` and `pkg/core` carrying the directive, `pkg/log`'s four TUs compiled fine but all five `pkg/core` cgo preamble TUs ICEd at EOF (`_cgo_export.c:519:1`, `gdextension_binding_init.c:13:1`, `classdb_callback.c:85:1`, `go_callable.c:23:1`, `method_bind.c:17:1` — the same `dwarf2codeview` crash signature), so the opt-in set stops at `pkg/log`.


### D5: Evidence gate extension is the TDD surface

The Windows verify step gains one check after the existing probe/banner greps: extract a banner block (`awk` between the start/end markers) and require `/[A-Za-z0-9_.-]+\.(c|h|s):[0-9]+/`. The rule was committed before the implementation, but CI never saw that intermediate commit (the branch push predates the PR), so RED was established on the host instead: the committed pipeline run verbatim against the real pre-implementation artifact log exits 1 on its `::error::` branch, and passes synthetic line-bearing banners (including Windows `C:\…\file.c:12` paths — the character class excludes backslashes and colons, so the basename still matches). CI's GREEN run remains the product-side proof; the host stub headers model compilation, not debug-info loading.

## Risks / Trade-offs

- `ld` PDB writing is younger than its DWARF path; exotic CodeView records from cgo-generated C could expose bugs. Mitigation: the link is a hard failure, the suite runs the same binaries, and only the test target changes.
- Losing DWARF in the C portion means `go tool addr2line` no longer resolves *C* frame lines in the Windows DLL; the PDB (uploaded as an artifact alongside) plus `llvm-symbolizer` covers that case.
- PDB size on disk (debug build, -O0) is large; it lives only in `test/demo/lib` and CI artifacts, never in distributions.
