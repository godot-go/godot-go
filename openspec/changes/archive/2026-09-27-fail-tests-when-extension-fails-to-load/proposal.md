# Proposal

## Why

`make test` can report success when the extension never loaded and zero tests ran. Two verified defects in the `test` target (Makefile:100-110):

1. **No build dependency.** `test:` does not depend on `build`, while `generate:` depends on `clean`, and `clean:` deletes `test/demo/lib/libgodotgo-*`. After `make generate`, `make test` runs against a missing binary.
2. **The gate only looks for leaks.** The sole failure condition is `grep -qE "ObjectDB instances were leaked|Leaked instance:"` on `test-output.log`. Nothing asserts that the extension loaded, the driver parsed, or any assertion executed.

Observed this session: with the `.so` missing, the run logged `Can't open dynamic library…` and a `Parse error` for `res://main.gd`, yet `make test` **exited 0**.

Correction to the initial read: the driver calls `get_tree().quit(1)` on failures (test/demo/test_base.gd:59) and the target uses `bash -o pipefail`, so `FAILURES > 0` already exits non-zero — but silently, indistinguishable from a crash. The gate never parses the `TESTS FINISHED / PASSES / FAILURES` banner, so every failure mode except a leak presents as green or an unexplained exit code.

## What Changes

- `test` gains a `build` dependency so the library is always current.
- The gate moves into a small check script that parses `test-output.log` (ANSI-stripped) and **positively asserts** the extension loaded, the driver summary banner is present, the total test count is greater than zero, and the driver-reported `FAILURES` count is zero — preserving the existing leak check.
- Each failure mode (extension not loaded, script parse error, missing summary / zero tests, driver-reported failures, leaks) emits its own distinct, actionable message.
- Known startup warnings (`get_variant_get_internal_ptr_func`, `callable_custom_get_user_data`, `classdb_register_extension_class_5/6`) never fail the gate.

**Known-failure decision (explicit):** the pre-existing `to_string` failure at test/demo/main.gd:28 (tracked by `fix-to-string-virtual-dispatch`) is **not allowlisted** — it keeps the target red until that change lands. An allowlist would recreate the "green means nothing" ambiguity this change removes; the new gate reports it unmistakably as "driver reported 1 failure(s)", never a harness-integrity failure.

## Capabilities

### New Capabilities
- `test-harness-integrity`: The headless test target guarantees a current extension library and fails loudly with a distinct, actionable message when the extension fails to load, the driver script fails to parse, the driver reports no summary or zero tests, or the driver reports any failures — while preserving leak detection and passing genuinely clean runs.

### Modified Capabilities
(none)

## Non-goals

- No change to the test driver, suite content, or the GDExtension itself.
- No allowlist mechanism for known assertion failures (see decision above).
- No CI workflow changes; this targets the local `make test` gate only.
- No distinct exit codes per failure mode — distinct human-readable messages are the contract.

## Acceptance

- Removing the built library makes `make test` exit non-zero with the extension-load-failure message instead of 0.
- `make build && GODOT=/home/pcting/bin/godot make test` reports exactly the one known tracked failure (red, "driver reported 1 failure(s)", no harness-integrity failure); the same command exits 0 once `fix-to-string-virtual-dispatch` lands.
- The known startup warnings never trip the gate.
