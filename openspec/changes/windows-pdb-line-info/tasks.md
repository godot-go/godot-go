# Tasks

## 1. Evidence gate red first

- [x] 1.1 Extended the `Verify Stacktrace Capture Evidence` step in `ci_windows.yaml`: extract the first banner block (awk between the start/end markers) and require a frame carrying `file.ext:line`. CI never observed the gate-only commit (the branch push predates PR #173, so no `pull_request` run existed); RED was instead demonstrated by running the committed pipeline verbatim against the pre-implementation artifact log from run 39193759198 (banner block extracted, no line-bearing frame -> exit 1) and GREEN-simulated against synthetic line-bearing banners including Windows backslash paths (exit 0). This behavior only exists on Windows, so CI remains the product-side proof.

## 2. Implementation

- [x] 2.1 `Makefile`: windows-scoped debug info for the `build` target — `-gcodeview` replaces `-ggdb` when `GOOS=windows`, and `-Wl,--pdb=$(CURDIR)/$(OUTPUT_PATH)/<basename>.pdb` pins the PDB next to the DLL (a bare name would write into the linker's CWD, the repo root — see design D2). Verified via `make -n` expansion for both GOOS values; the Linux expansion is semantically unchanged.
- [x] 2.2 `pkg/log/stacktrace.c`: after `SymFromAddr` succeeds, try `SymGetLineFromAddr64` and append ` file:line` to the frame line on success. Syntax-checked the `_WIN32` branch against the extended stub headers (`IMAGEHLP_LINE64`, `SymGetLineFromAddr64`) with `-fsyntax-only -Wall -std=gnu90`; Windows CI is the real compile.
- [ ] 2.3 Confirm the artifact upload step (already on this branch from the stacked base) publishes the `.pdb` alongside the DLL.

## 3. Proof

- [ ] 3.1 Windows CI GREEN: banner frames carry `file:line`, `windows stacktrace capture verified` still prints, existing probe/banner assertions unchanged.
- [ ] 3.2 Host `make build` + `make test` green (flag change is `GOOS=windows`-scoped; Linux/alpine paths byte-identical), `make test_gate` green.

## 4. Docs

- [ ] 4.1 README Windows offline note: with `-gcodeview` the C lines live in the PDB (`llvm-symbolizer -e=<dll> -p=<pdb>`), Go frames stay on `go tool addr2line`.
