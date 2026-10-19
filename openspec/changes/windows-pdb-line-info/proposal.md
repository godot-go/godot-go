# Proposal: Windows PDB line info in captured backtraces

## Why

The Windows capture prints frames as `symbol+0xdisplacement` or `module+0xoffset (ip 0x…)` because the system `dbghelp.dll` reads only PDB/COFF debug info while the mingw build emits DWARF. Anyone reading a captured trace (e.g. pasted from a CI log) still cannot see which source line a frame is in without an offline round trip against the published debug library. GCC >= 13 can emit CodeView debug info (`-gcodeview`) and GNU ld >= 2.40 writes the matching PDB (`--pdb=<name>`); with `SYMOPT_LOAD_LINES` (already set by the walk) dbghelp then answers `SymGetLineFromAddr64` per frame, giving live `file:line` for the C side on Windows.

## What Changes

- Windows debug builds compile C sources with `-gcodeview` (replacing `-ggdb` when `GOOS=windows`) and link the PDB next to the library (`test/demo/lib/libgodotgo-test-windows-amd64.pdb`): the embedded absolute path hits directly on the build machine, and dbghelp's same-directory basename fallback covers relocated artifacts.
- The DbgHelp walk appends ` file:line` to frames where `SymGetLineFromAddr64` succeeds; every other frame form is unchanged.
- The Windows CI `Verify Stacktrace Capture Evidence` step additionally requires at least one banner frame carrying `file.ext:line`, so a build that silently loses line info (missing PDB, GUID mismatch) cannot pass.
- README: Windows offline recipe updated — C frames resolve via the uploaded PDB (`llvm-symbolizer`), Go frames stay on `go tool addr2line`.

## Scope note

Non-goals: line info for Go-runtime frames from dbghelp (their debug info stays DWARF/gopclntab and is already covered by the companion Go stack); PDB generation for `build-full` or the godot-integration build targets (the `build` test target is the gated surface); engine or GDScript frames (unreachable from native code).
