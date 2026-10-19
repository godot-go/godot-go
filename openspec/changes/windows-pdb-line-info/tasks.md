# Tasks

## 1. Evidence gate red first

- [x] 1.1 Extended the `Verify Stacktrace Capture Evidence` step in `ci_windows.yaml`: extract the first banner block (awk between the start/end markers) and require a frame carrying `file.ext:line`. CI never observed the gate-only commit (the branch push predates PR #173, so no `pull_request` run existed); RED was instead demonstrated by running the committed pipeline verbatim against the pre-implementation artifact log from run 39193759198 (banner block extracted, no line-bearing frame -> exit 1) and GREEN-simulated against synthetic line-bearing banners including Windows backslash paths (exit 0). This behavior only exists on Windows, so CI remains the product-side proof.

## 2. Implementation

- [x] 2.1 `Makefile`: the `build` target's windows link gains `-Wl,--pdb=$(CURDIR)/$(OUTPUT_PATH)/<basename>.pdb`, pinning the PDB next to the DLL (a bare name would write into the linker's CWD, the repo root — see design D2). Global `CGO_CFLAGS` keep `-ggdb` after the D1 opt-in pivot; verified via `make -n` expansion for both GOOS values.
- [x] 2.2 `pkg/log/stacktrace.c`: after `SymFromAddr` succeeds, try `SymGetLineFromAddr64` and append ` file:line` to the frame line on success. Syntax-checked the `_WIN32` branch against the extended stub headers (`IMAGEHLP_LINE64`, `SymGetLineFromAddr64`) with `-fsyntax-only -Wall -std=gnu90`; Windows CI is the real compile.
- [ ] 2.3 Confirm the artifact upload step (already on this branch from the stacked base) publishes the `.pdb` alongside the DLL.
- [x] 2.4 Restructured the Win32 walk around the runner's dwarf2codeview ICE (design D6): replaced the `RtlCaptureContext`+`StackWalk64`/`STACKFRAME64` block with a `CaptureStackBackTrace` return-address walk — the `STACKFRAME64` local is the verified crash trigger on MinGW-Builds GCC 15.2.0, and the replacement shape was pre-validated as a minimal TU on the same runner. Host stub syntax checks clean on both `_WIN32` and host paths; host build, full suite, and fixtures green.
- [x] 2.5 After the generated `pkg/ffi` TUs also ICEd (design D6, second paragraph), scoped `-gcodeview` down to a `#cgo windows CFLAGS:` directive (design D1 opt-in): global flags reverted to DWARF. A `pkg/core` directive was tried next and failed as documented — all five of its cgo preamble TUs ICEd (design D6) — so the opt-in ships on `pkg/log/lib.go` alone; `GoCallback_*` frames keep the symbol-only form with offline DWARF resolution.

## 3. Proof

- [ ] 3.1 Windows CI GREEN: banner frames carry `file:line`, `windows stacktrace capture verified` still prints, existing probe/banner assertions unchanged.
- [ ] 3.2 Host `make build` + `make test` green (flag change is `GOOS=windows`-scoped; Linux/alpine paths byte-identical), `make test_gate` green.

## 4. Docs

- [ ] 4.1 README Windows offline note: with `-gcodeview` the C lines live in the PDB (`llvm-symbolizer -e=<dll> -p=<pdb>`), Go frames stay on `go tool addr2line`.
