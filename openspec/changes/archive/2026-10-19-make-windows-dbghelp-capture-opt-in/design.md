# Design

## Context

See `proposal.md` — Why. The C capture lives in `pkg/log/stacktrace.c` and is compiled for every package build; `pkg/core/stacktrace.go` exposes `SetPrintStacktrace` and `StacktraceCaptureAvailable` over it. The Windows implementation includes `<windows.h>`/`<dbghelp.h>` and calls DbgHelp stack-walking APIs, which places `dbghelp.dll` in the import table of the extension: Windows loads it at DLL load time, independent of the runtime enable flag.

On Windows the suite's capture probe reports compile-time truth through `stacktraceCaptureAvailable()`, and `.github/workflows/ci_windows.yaml` asserts that evidence directly because the Test step is `continue-on-error`. The musl opt-in precedent (`GODOT_GO_ENABLE_EXECINFO` + `EXTRA_CGO_CFLAGS`) already establishes how build-time toggles are documented and passed.

## Goals / Non-Goals

**Goals:**

- A default Windows build contains no DbgHelp dependency: no import, no load, no detach path.
- Capture remains fully available (banners, symbols, PDB file:line, offline resolution) when opted in.
- The availability report and CI evidence step agree in both build modes.
- Non-Windows behavior (glibc/Apple automatic execinfo; musl with or without an external implementation) is unchanged.

**Non-Goals:**

- Fixing the teardown crash itself; this is a mitigation and a discriminating experiment.
- Changing the Go API shape, the probe protocol, or the offline symbolization recipe.
- Adding a dedicated Windows CI job that builds with the opt-in enabled (follow-up once #172 is resolved).

## Decisions

### D1: One macro gates the includes, the body, and the availability report

`GODOT_GO_ENABLE_DBGHELP` gates the `<windows.h>`/`<dbghelp.h>` includes together with the capture body and the `stacktraceCaptureAvailable()` result. Gating the include is what removes the import; a runtime-only toggle (keeping the include) would leave `dbghelp.dll` loaded and its detach path runnable, so it cannot serve either the mitigation or the experiment. A link-time shim that stubs DbgHelp symbols was rejected as more machinery than the interim need justifies.

### D2: Opt-in rather than default-on with an opt-out

The mitigation must apply to stock builds and CI without extra configuration, so the default is compiled-out. This mirrors the existing `GODOT_GO_ENABLE_EXECINFO` opt-in pattern, keeping the two platform toggles conceptually parallel. Default-on would require every builder to know the flag to get the mitigation, defeating the purpose.

### D3: `stacktraceCaptureAvailable()` remains the single source of truth

The probe line, the Go-side gate decisions, and the CI evidence step all derive from the availability report. No new API, no probe-format change. When the macro is absent on Windows, `capture=0` flows through the existing probe unchanged.

### D4: The CI evidence step distinguishes “compiled out” from “broken”

The step treats the explicit `capture=0` probe line as a healthy compiled-out state (printing the mitigation reason), while keeping the existing failures: `capture=1` with no banner, and `capture=1` with no captured frame carrying file:line. Whenever capture is compiled in, the windows-pdb-line-info guarantees are still enforced. Removing the assertion entirely was rejected: it would let a silently broken capture go green once the macro is re-enabled.

### D5: Execinfo paths are not touched

The `#elif` branches for glibc/Apple and the musl external implementation keep their current behavior and documentation; only the Windows branch gains the conditional layer.

## Risks / Trade-offs

- [Suspect is wrong: the crash persists without dbghelp] → The build doubles as the experiment; the diagnosis continues in `openspec/changes/diagnose-windows-teardown-crash/`, and the toggle is trivially reversible.
- [Default CI no longer exercises the Windows DbgHelp path, so PDB line-info could regress unnoticed] → The evidence gate still enforces everything whenever capture is compiled in; add an opt-in capture build to Windows CI as a follow-up if the toggle outlives the investigation.
- [Windows consumers relying on native capture lose it by default] → Documented in README with the exact `EXTRA_CGO_CFLAGS` invocation, and the availability report points at the compiled-out state.
- [Macro drift between docs, code, and CI] → The macro name appears in one C file, one README section, and one CI comment; the probe line remains the observable evidence in both modes.

## Migration Plan

- Deploy: default Windows builds report capture unavailable; consumers who want capture rebuild with `EXTRA_CGO_CFLAGS='-DGODOT_GO_ENABLE_DBGHELP'`.
- Rollback: enable the macro in build configuration (single flag), or remove the gating once #172 is resolved; no data or API migration is involved.
