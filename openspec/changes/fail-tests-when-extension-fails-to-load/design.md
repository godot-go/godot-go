# Design

## Context

The `test` target (Makefile:100-110) runs Godot headless piped to `tee test-output.log` under `bash -o pipefail -ec`, then applies a single gate: `grep -qE "ObjectDB instances were leaked|Leaked instance:"`. Verified against the current tree:

- `test:` has no `build` dependency; `generate: clean` and `clean:` deletes `test/demo/lib/libgodotgo-*`, so a post-`generate` `make test` runs against a missing binary.
- With the `.so` missing, the run logs `Can't open dynamic library…`, `GDExtension dynamic library not found…`, `Failed to load script "res://main.gd" with error "Parse error"`, `Cannot get class 'Example'` — and the target exits 0.
- The driver (`test/demo/test_base.gd`) prints `==== TESTS FINISHED ====`, `   PASSES: <n>`, `   FAILURES: <n>` via `print_rich` (ANSI-colored) and calls `get_tree().quit(0 if success else 1)`.
- Correction to the initial finding: because the target uses `bash -o pipefail`, the driver's `quit(1)` *does* propagate — the current baseline (1077 passes / 1 known failure) already exits non-zero (observed: `make: *** [Makefile:101: test] Error 1`). The defect is that this failure is silent: no gate message, indistinguishable from a Godot crash, and `-e` short-circuits before the leak check even runs. The true holes are the absence of any positive assertion (no-banner runs exit 0) and the missing build dependency.

## Goals / Non-Goals

**Goals:**
- A green `make test` means the extension loaded, the driver ran, and the driver reported a positive test count with zero failures and no leaks.
- Every failure mode emits one distinct, actionable message; no failure mode presents as green or as a bare exit code.
- Preserve the existing leak check verbatim in behavior.

**Non-Goals:**
- No changes to the test driver, suite content, or the extension.
- No allowlist mechanism for assertion failures (see Decisions).
- No CI workflow changes; no per-mode distinct exit codes (messages are the contract).

## Decisions

### D1: `test` depends on `build`

`test: build` guarantees a current library. Alternative considered: an existence check that errors when the `.so` is missing — rejected because it still lets a *stale* library through, and `go build` is incremental so the cost is near zero when nothing changed.

### D2: Gate moves into a dedicated script (`test/check_test_output.sh`)

A single script owns all log patterns and messages instead of inline Makefile shell. Rationale: the gate grows from one grep to six checks; a script is readable, independently testable against fixture logs, and reusable by CI later. Alternative: inline greps — rejected as unmaintainable and untestable.

### D3: Positive assertion anchored on the driver banner; error markers checked first for precision

The `TESTS FINISHED` banner is the authoritative proof the extension loaded and the driver parsed (main.gd references the `Example` class; if the extension is missing the script cannot parse and the banner never prints). The script checks in order:

1. Extension-load error markers (`Can't open dynamic library`, `GDExtension dynamic library not found`) → "extension failed to load" message.
2. Script load/parse markers (`Failed to load script` + `Parse error`) → "test script parse error" message.
3. Banner absent → "driver did not report a summary" message.
4. `PASSES + FAILURES == 0` → "zero tests executed" message.
5. `FAILURES > 0` → "driver reported N failure(s)" message.
6. Leak markers → existing "leaked engine objects" message (preserved).
7. Otherwise pass.

Checking specific markers before the generic banner-absent case yields precise messages for the common breakages; the banner check is the catch-all that makes "silently did nothing" impossible.

### D4: Strip ANSI escapes before parsing

`print_rich` wraps counts in color codes (e.g. `PASSES: \e[92m1077\e[39m`). The script pipes the log through `sed 's/\x1b\[[0-9;]*m//g'` first so count extraction (`PASSES: ([0-9]+)`) is robust.

### D5: Capture Godot's exit status so the gate always runs

The recipe runs the pipeline with `|| status=$?` instead of relying on `-e`, then always runs the check script. Precedence: a gate message wins; if the gate passes but Godot exited non-zero for a reason outside the known modes, the target still fails with a "godot exited non-zero (status N)" message plus a log tail. This removes the current silent-short-circuit where a driver failure skips the gate entirely.

### D6: The known `to_string` failure is not allowlisted

The pre-existing failure at test/demo/main.gd:28 (tracked by `fix-to-string-virtual-dispatch`) keeps the target red until that change lands. Rationale: an allowlist re-introduces hidden failures behind a green light — the exact anti-pattern this change removes — and exact-match allowlists can mask regressions at the same line. With D3's messages, the known failure is unambiguous ("driver reported 1 failure(s)"), never a harness-integrity complaint, and the same command turns green the moment the tracked fix merges.

### D7: All gate failures exit 1 with distinct messages

Distinct exit codes per mode were considered and rejected: callers (make, CI) only branch on zero/non-zero, and the actionable information belongs in the message. The script also tails the relevant log lines on failure so the developer sees evidence without opening the file.

## Risks / Trade-offs

- **Coupling to the driver's banner format.** If `test_base.gd` changes its banner text, the gate misreports. Mitigation: all patterns live in one script; add a comment in `test_base.gd` noting the banner is machine-consumed by `test/check_test_output.sh`; a broken banner fails loudly (missing-summary), never silently green.
- **`test: build` adds build time to every test run.** Incremental Go builds make this negligible in the common case; correctness outweighs the seconds.
- **Fixture logs can drift from real Godot output.** Mitigation: tasks verify the script against real runs (missing-library negative case, current baseline positive-parse case), not only fixtures.
- **Marker-based detection is heuristic** (a new Godot version could reword errors). Mitigation: the banner/zero-count checks are the structural backstop — even if every marker string changed, a no-banner run still fails loudly.
