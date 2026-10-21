# Design

## Context

See `proposal.md` for motivation. Constraints that shape the approach:

- `.github/workflows/ci_linux.yaml` runs the full chain on `ubuntu-latest` plus a musl build-only job; `ci_windows.yaml` adds capture evidence and artifact upload. macOS has neither.
- `Makefile:20-28` derives the output name from `GOOS`/`GOARCH`, giving `test/demo/lib/libgodotgo-test-macos-<arch>.framework`; `test/demo/example.gdextension:8-11` maps `macos.{debug,release}.{arm64,x86_64}` to those names. Nothing needs to change for a native-arch leg.
- `pkg/log/stacktrace.c` compiles the execinfo branch on Apple (`__has_include(<execinfo.h>)`, `__APPLE__` fallback), so the probe reports `capture=1` and prints the banner, but `backtrace_symbols` yields module+offset strings, not DWARF source lines.
- `test/check_test_output.sh:40` and `test/check_editor_output.sh:40` strip ANSI with GNU-only `\x1b` escapes.
- Runner facts (Oct 2026): `macos-latest` is macOS 26 **arm64**, 3-core M1 / 7 GB; Intel is `macos-26-intel`, 4-core / 14 GB. `macos-14` retires 2026-11-02.
- Godot asset: `Godot_v4.7-stable_macos.universal.zip` → `Godot.app/Contents/MacOS/Godot`.
- Godot's `OS_MacOS::get_framework_executable` (`../godot/platform/macos/os_macos.mm:366`) returns the path unchanged when it is not a bundle directory, and `open_dynamic_library` `dlopen`s it — so a plain Mach-O file named `*.framework` loads.
- `pkg/util` supplies `SetThreadName` only for linux (`prctl`) and windows (a no-op). `test/pkg/lib.go:38` calls it, so `GOOS=darwin` fails to compile; the first run of this change's own workflow surfaced it as `undefined: util.SetThreadName`.
- Measured sibling runtimes: Linux `make build` 111 s, `make test` 13 s; whole Windows job ~5 min.

## Goals / Non-Goals

**Goals:**

- Exercise build, extension load, the headless suite, the editor run, and the capture probe on both macOS architectures.
- Keep the new workflow a sibling of the existing two: same triggers, same step names, same gate scripts.
- Keep the shared gate scripts valid on both userlands.

**Non-Goals:**

- Real framework bundles, signing, notarization, packaging.
- Source-line stacktrace evidence on macOS.
- Any retry/masking logic that would hide a platform fault.

## Decisions

**Matrix pairs each architecture with a native runner.** `include` entries bind `macos-latest`+`arm64` and `macos-26-intel`+`amd64`, and the job passes `GOOS=darwin GOARCH=<arch>`. The alternative — one runner, two `GOARCH` values — makes the amd64 leg a cgo cross-compile that needs `-arch x86_64` plumbing the Makefile does not do, and the built library could not be exercised by the arm64 engine anyway, since the `.gdextension` file selects entries per architecture.

**`fail-fast: false`, unlike the Linux/Windows workflows.** While the platform is new, one architecture failing must not cancel the other; both results are the evidence.

**Explicit runner labels rather than `macos-13`/`macos-14`.** `macos-14` is being retired; `macos-latest` is documented as arm64 today and the Intel leg needs `macos-26-intel` regardless.

**Capture evidence stops at `capture=1` plus the banner.** The gate already enforces exactly that pair, platform-agnostically. Porting the Windows `file:line` assertion would fail against execinfo by construction; linking libbacktrace on darwin to satisfy it is out of scope.

**Keep the plain-Mach-O `.framework`.** Godot resolves it through the non-bundle fallback. Emitting a real bundle is a Makefile/build change with no CI-only benefit.

**Portable ANSI strip via a literal ESC byte.** `esc=$(printf '\033')` then `sed "s/${esc}\\[[0-9;?]*[a-zA-Z]//g"`. Considered: `perl -pe` (extra runtime dependency), Homebrew `gsed` (not guaranteed present), bash ANSI-C quoting inside the pattern (works, but embeds a raw control character in the source line).

**Artifact names carry the architecture** (`macos-arm64-test-lib`, `macos-amd64-test-lib`); `upload-artifact` rejects duplicate names across matrix legs.

**The darwin build gap is closed with a no-op stub, matching Windows.** macOS does have `pthread_setname_np`, but its one-argument form renames only the calling thread, and using it would add cgo to `pkg/util` for a test-only diagnostic; `threadname_windows.go` is already a no-op, so parity beats a new dependency. A real implementation is deferred until something consumes the thread name on darwin.

## Risks / Trade-offs

- **macOS shutdown `dlclose` faults like the Windows teardown crash** → out of scope per the proposal; the first run is the probe. Mitigation is a follow-up change, not masking; the job stays off branch protection meanwhile.
- **Editor doc-cache race on a cold checkout** → reuse `make ci_gen_test_project_files` unchanged, which already carries the frame budget for this.
- **7 GB arm64 runner vs a heavy `-gcflags=all="-N -l"` build** → 30 min job budget with 3/7 min step timeouts; if it OOMs, the follow-up is a CI-specific build profile, not a quieter gate.
- **BSD userland differences beyond `sed`** → the gate already uses `grep -aE`, bash here-strings and portable `awk`; the evidence step must avoid GNU-only flags when it is written.
- **`make generate` output formatting depends on installed `clang-format`/`goimports`** → the `git diff` step stays informational, and the build uses the regenerated sources either way.
- **Godot.app quarantine on the runner** → `curl` does not set the attribute; `xattr -dr com.apple.quarantine` is applied defensively and ignored if absent.

## Migration Plan

Merge adds one workflow file and one behaviour-preserving change to each gate script. Rollback is deleting `.github/workflows/ci_macos.yaml`; the gate-script change is inert on Linux and Windows because the ESC-byte form is accepted by GNU `sed` too.

## Open Questions

- Whether macOS teardown survives `dlclose` of a Go c-shared library — answerable only by the first run, and it does not change the specs, approach, or tasks.
- Whether execinfo frames are informative enough to assert on beyond the banner — deferrable; the gate's current pair is sufficient proof that capture ran.
