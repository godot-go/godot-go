# Design: Windows PDB line info

## Context

`windows-stacktrace-capture` shipped the DbgHelp walk with module+offset frames because the only debug info the mingw link produced (DWARF) is invisible to `dbghelp.dll`. This change adds the missing half: CodeView/PDB, so the same walk can print live source lines.

## Decisions

### D1: Replace `-ggdb` with `-gcodeview` for windows rather than combine

GCC's debug-info formats are an exclusive selection (`debug_type_masks[]` in `opts.cc`; combining formats beyond the legacy VMS+DWARF case is explicitly discouraged). For `GOOS=windows` the `build` target's debug C flags become `-g -gcodeview -O0 …`; every other platform keeps `-g -ggdb`. Consequence, accepted: the C portion of the DLL loses DWARF line tables (they move to the PDB); Go portions keep their own DWARF/gopclntab because `gc` does not read `CGO_CFLAGS`, so the existing `go tool addr2line` recipe still resolves Go frames offline.

### D2: PDB name is the bare basename of the library with `.pdb`

ld embeds the `--pdb=` string verbatim into the CodeView RSDS debug record. Passing a name without a directory makes dbghelp's default lookup (module directory first, then symbol path) find `libgodotgo-test-windows-amd64.pdb` sitting next to the DLL — which is where the build places it and where the CI artifact upload already globs `*.pdb`. Absolute build-machine paths in the record would fail on any other machine.

### D3: Frame format gains a trailing ` file:line` only on line lookup success

```
[0] GoCallback_MethodBindMethodPtrcall+0xd48 _cgo_.c:123
[3] libgodotgo-test-windows-amd64.dll+0x987ad (ip 0x7fff900f87ad)
```

Frames whose line is unknown (Go runtime, engine DLLs) print exactly as before, and existing gate/verify anchors (banner lines, probe markers) are untouched. `IMAGEHLP_LINE64` needs only `SizeOfStruct`, `FileName`, `LineNumber`; `SymSetOptions(… | SYMOPT_LOAD_LINES)` is already in place and must stay before `SymInitialize`.

### D4: Toolchain floors are proven by CI, not probed

`--pdb=` requires binutils >= 2.40 (patch series merged October 2022); `-gcodeview` requires GCC >= 13. `windows-latest` MSYS2 ships far above both. No version sniffing is added to the Makefile: an ld too old for `--pdb` fails the link loudly rather than degrading silently, and the verify step (D5) catches debug-info regressions that link clean.

### D5: Evidence gate extension is the TDD surface

The Windows verify step gains one check after the existing probe/banner greps: extract a banner block (`awk` between the start/end markers) and require `/[A-Za-z0-9_.-]+\.(c|h|s):[0-9]+/`. Committed before the implementation, this is the RED state on the Windows CI job — the only place this behavior can be tested (no local mingw; host stub headers model compilation, not debug-info loading).

## Risks / Trade-offs

- `ld` PDB writing is younger than its DWARF path; exotic CodeView records from cgo-generated C could expose bugs. Mitigation: the link is a hard failure, the suite runs the same binaries, and only the test target changes.
- Losing DWARF in the C portion means `go tool addr2line` no longer resolves *C* frame lines in the Windows DLL; the PDB (uploaded as an artifact alongside) plus `llvm-symbolizer` covers that case.
- PDB size on disk (debug build, -O0) is large; it lives only in `test/demo/lib` and CI artifacts, never in distributions.
