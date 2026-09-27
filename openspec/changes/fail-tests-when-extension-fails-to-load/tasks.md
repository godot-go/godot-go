# Tasks

## 1. Check script

- [ ] 1.1 Create `test/check_test_output.sh` implementing the gate from design D3: ANSI-strip the log (D4), then check in order — extension-load error markers (`Can't open dynamic library`, `GDExtension dynamic library not found`), script parse markers (`Failed to load script` + `Parse error`), missing `TESTS FINISHED` banner, `PASSES + FAILURES == 0`, `FAILURES > 0` (report the count), and the preserved leak markers (`ObjectDB instances were leaked`, `Leaked instance:`). Each mode prints its own distinct, actionable message plus a tail of the relevant log lines; exit 0 only when all checks pass (D7). Takes the log path as `$1` and Godot's exit status as `$2`.
- [ ] 1.2 Add fixture logs under `test/fixtures/` (clean-pass with the known startup warnings present, missing-library, parse-error, no-banner, zero-tests, failures=1, leaked) and a small runner that asserts the script's exit status and message for each fixture; keep it runnable in one command.

## 2. Makefile wiring

- [ ] 2.1 Change the `test` target to depend on `build` (`test: build`) so the extension library is always current (D1).
- [ ] 2.2 Replace the inline leak-only gate in `test` with: run the Godot pipeline capturing its status via `|| status=$?` (so `-e` no longer short-circuits the gate), then invoke `test/check_test_output.sh` with `$(CURDIR)/test-output.log` and the captured status; propagate the script's failure, and if the script passes but Godot exited non-zero, fail with a "godot exited non-zero (status N)" message plus a log tail (D5).
- [ ] 2.3 Add a comment in `test/demo/test_base.gd` above `exit_with_status()` noting the `TESTS FINISHED` / `PASSES:` / `FAILURES:` banner format is machine-consumed by `test/check_test_output.sh` and must not be reworded without updating the gate.

## 3. Verification

- [ ] 3.1 Negative acceptance: `rm -f test/demo/lib/libgodotgo-* && GODOT=/home/pcting/bin/godot make test` exits non-zero with the extension-load-failure message (previously exited 0).
- [ ] 3.2 Positive acceptance: `GODOT=/home/pcting/bin/godot make build && GODOT=/home/pcting/bin/godot make test` shows the gate parsing the banner and reporting exactly the one known tracked failure — red with an explicit "driver reported 1 failure(s)" message and no harness-integrity failure — confirming the no-allowlist decision reads clearly in the output.
- [ ] 3.3 Confirm the known startup warnings (`get_variant_get_internal_ptr_func`, `callable_custom_get_user_data`, `classdb_register_extension_class_5/6`) never trip the gate, using the clean-pass fixture and the real baseline log.
- [ ] 3.4 Confirm the build dependency closes the stale-library hole: `GODOT=/home/pcting/bin/godot make generate && GODOT=/home/pcting/bin/godot make test` rebuilds the library and reaches the driver summary instead of running against a missing binary.
- [ ] 3.5 Confirm the preserved leak check still fires against the leak fixture and that the fixture runner from 1.2 is green.
