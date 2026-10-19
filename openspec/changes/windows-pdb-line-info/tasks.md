# Tasks

## 1. Evidence gate red first

- [ ] 1.1 Extend the `Verify Stacktrace Capture Evidence` step in `ci_windows.yaml`: extract the first banner block (awk between the start/end markers) and require a frame carrying `file.ext:line`. Push before any implementation — the Windows job must fail with the new `::error::` (RED); this behavior only exists on Windows, so CI is the test surface (host stub headers model compilation only).

## 2. Implementation

- [ ] 2.1 `Makefile`: windows-scoped debug info for the `build` target — `-gcodeview` replaces `-ggdb` when `GOOS=windows`, and `-Wl,--pdb=<bare basename>.pdb` joins the link flags so the PDB lands next to the DLL with a directory-free name in the CodeView record.
- [ ] 2.2 `pkg/log/stacktrace.c`: after `SymFromAddr` succeeds, try `SymGetLineFromAddr64` and append ` file:line` to the frame line on success. Syntax-check the `_WIN32` branch against the extended stub headers (`IMAGEHLP_LINE64`, `SymGetLineFromAddr64`); Windows CI is the real compile.
- [ ] 2.3 Confirm the artifact upload step (already on this branch from the stacked base) publishes the `.pdb` alongside the DLL.

## 3. Proof

- [ ] 3.1 Windows CI GREEN: banner frames carry `file:line`, `windows stacktrace capture verified` still prints, existing probe/banner assertions unchanged.
- [ ] 3.2 Host `make build` + `make test` green (flag change is `GOOS=windows`-scoped; Linux/alpine paths byte-identical), `make test_gate` green.

## 4. Docs

- [ ] 4.1 README Windows offline note: with `-gcodeview` the C lines live in the PDB (`llvm-symbolizer -e=<dll> -p=<pdb>`), Go frames stay on `go tool addr2line`.
