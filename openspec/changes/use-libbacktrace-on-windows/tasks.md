# Tasks

## 1. Prove the libbacktrace build path (go/no-go)

- [x] 1.1 Vendor Godot 4.7-stable's PE/COFF libbacktrace sources, headers, `config.h`, and `LICENSE` into `pkg/log/libbacktrace/`; set `HAVE_ZLIB` and `HAVE_LIBLZMA` to 0; omit the unused unwind entry points (`backtrace.c`, `simple.c`, `print.c`); record the upstream commit and the big-files patch in `pkg/log/libbacktrace/README.md`. Verify: the file list matches Godot's `thirdparty/libbacktrace` minus the ELF/Mach-O sources and the three unwind entry points, `LICENSE` is present, and `README.md` names the 4.7-stable source.
- [x] 1.2 Add the Windows-only cgo wrapper package and the `pkg/log` Windows-only blank import plus a C symbol anchor so the vendored objects are linked; add the `#cgo windows CFLAGS: -I${SRCDIR}/libbacktrace` include path. Verify: a Windows `make build` links with no undefined `backtrace_*` symbols, and Linux `go build ./...` never compiles the Windows-only package.
- [x] 1.3 Build the Windows DLL and run a host-side harness that links the vendored libbacktrace against that PE and resolves `printStacktrace`'s preferred VA. Verify: the harness prints `stacktrace.c:<line>` for that address. If it does not, stop and revisit design D1/D4 before continuing.
- [x] 1.4 Keep the proven state/anchor scaffolding and leave no temporary harness in the repo. Verify: the Windows build still links and the Linux build is unchanged.

## 2. Implement the Windows resolver

- [x] 2.1 Replace the DbgHelp branch in `pkg/log/stacktrace.c` with libbacktrace state initialization: module path, `SizeOfImage` from the mapped PE header, preferred `ImageBase` from the DLL file (Windows rewrites the mapped value to the load address), runtime delta, `threaded=1`, and a one-time lock; remove the `<dbghelp.h>` include and the `GODOT_GO_ENABLE_DBGHELP` gate. Verify: a default Windows build compiles, `stacktraceCaptureAvailable()` returns true, and the DLL imports no `dbghelp.dll`.
- [x] 2.2 Resolve captured PCs inside the extension module with `backtrace_pcinfo` (return-address adjustment and inline-frame handling), printing `function file:line`; keep `module+0xoffset (ip)` for foreign or unresolved frames. Verify: the host-side harness resolves a `pkg/log/stacktrace.c:<line>` frame; the live Windows banner check is part of 5.2.
- [x] 2.3 Update `pkg/log/lib.go` to remove `-ldbghelp` and `-gcodeview` and add the vendored include path; update `pkg/log/stacktrace.h` and `pkg/core/stacktrace.go` comments and availability text. Verify: `go build ./...` on Linux, and a Windows `go build -x` log contains no `-gcodeview` or `-ldbghelp`.

## 3. Remove the PDB/lld build path and update CI

- [x] 3.1 Remove `DEBUG_INFO_EXTLDFLAGS` and `GO_LINK_LDFLAGS` and their comments from the `Makefile`; keep `-ggdb`. Verify: a Windows `make build` succeeds without `ld.lld` on PATH and emits no `.pdb`; a Linux `make build` is unchanged.
- [x] 3.2 Update `.github/workflows/ci_windows.yaml`: remove the LLVM PATH and `ld.lld` checks from Build and Test; require `capture=1`, a banner, and a `pkg/log/*.c:<line>` frame in the capture-evidence step; assert the DLL has no `dbghelp.dll` reference; drop `*.pdb` from the artifact upload. Verify: the YAML parses and the assertions are present; the end-to-end run is 5.2.
- [x] 3.3 Confirm the Linux and Alpine workflows need no changes for this path; verify `rg 'gcodeview|ld.lld|--pdb|dbghelp|DBGHELP' .github/workflows/` returns only the intended Windows assertions (no stale PDB/lld steps).

## 4. Documentation and provenance

- [x] 4.1 Rewrite the README Windows capture section: libbacktrace, compiled in by default, no `GODOT_GO_ENABLE_DBGHELP`, no PDB/`llvm-symbolizer`; keep the `go tool addr2line` offline recipe. Verify: `rg 'GODOT_GO_ENABLE_DBGHELP|llvm-symbolizer|--pdb' README.md` returns nothing and the documented `make build`/test commands still match the Makefile.
- [x] 4.2 Verify the vendored license and provenance are shipped with the source and referenced from the README where third-party code is discussed. Verify: the `LICENSE` file exists in `pkg/log/libbacktrace/` and the README names the vendored component.

## 5. Integration verification

- [x] 5.1 Run `openspec validate use-libbacktrace-on-windows --strict`, `make test_gate`, and `make test` on Linux. Verify: validation passes, the fixture table passes, and the Linux run still produces its execinfo capture and Go companion stack.
- [x] 5.2 Run the Windows CI job end to end. Verify: Build succeeds without LLVM/lld, the probe reports `capture=1`, at least one captured frame carries `pkg/log/*.c:file:line`, no `dbghelp.dll` import is found, and the artifact contains the DLL but no PDB.
- [x] 5.3 Re-run the Alpine `build_alpine` job. Verify: it still builds with capture compiled out and the `GODOT_GO_ENABLE_EXECINFO` opt-in path still links.

## Workflow follow-up

- Archive the change after merge and the spec sync; the `native-stacktrace-capture` delta replaces the DbgHelp/PDB requirements.
- Remove any remaining documentation that treats `capture=0` as the normal Windows state.
