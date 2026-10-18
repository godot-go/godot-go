#!/usr/bin/env bash
#
# Gate for the headless test run.
#
# Parses Godot's captured output and fails loudly, with one distinct message per
# failure mode, whenever the harness did not genuinely run the suite. Exit 0 means
# all of: the extension loaded, the driver parsed and completed, a positive number
# of assertions ran, the driver reported zero failures, the main-loop shutdown
# marker fired, and no engine objects leaked.
#
# Before this gate existed, a run with no built library logged
# "Can't open dynamic library" and a GDScript parse error, ran zero assertions,
# and still exited 0. Green meant nothing.
#
# Usage: check_test_output.sh <log-path> <godot-exit-status>
#
# The patterns below are coupled to the driver's summary banner in
# test/demo/test_base.gd (exit_with_status). If that banner is reworded, this
# gate must be updated with it -- a mismatch fails loudly as "no driver summary",
# never as a false pass.

set -uo pipefail

log="${1:-}"
godot_status="${2:-0}"

if [[ -z "$log" ]]; then
	echo "FAIL[harness]: usage: check_test_output.sh <log-path> <godot-exit-status>"
	exit 1
fi

if [[ ! -f "$log" ]]; then
	echo "FAIL[harness]: test output log not found: ${log}"
	echo "FAIL[harness]: the test run produced no log at all"
	exit 1
fi

# print_rich wraps the counts in ANSI colour codes ("PASSES: \e[92m1079\e[39m"),
# so strip escapes before parsing anything or the numbers never match.
plain="$(sed -e 's/\x1b\[[0-9;?]*[a-zA-Z]//g' "$log")"

# Membership checks test $plain with `grep -q ... <<<"$plain"`, never
# `printf '%s\n' "$plain" | grep -q`. Under `set -o pipefail` the piped form
# is racy: grep -q exits at the first match, and if printf has not finished
# writing (the log is far larger than the pipe buffer once gctrace lines pile
# up) printf dies of SIGPIPE with status 141, so the pipeline reports failure
# even though grep matched -- a green run can falsely fail as
# "missing-main-loop-shutdown-marker". A here-string reads from a temp file:
# no pipe, no SIGPIPE, no inversion.

# evidence <regex> -- print matching lines so the failure carries its own proof.
evidence() {
	local matches
	matches="$(printf '%s\n' "$plain" | grep -aE "$1" | tail -5)"
	if [[ -n "$matches" ]]; then
		printf '%s\n' "$matches" | sed 's/^/    | /'
	fi
}

# fail <mode> <message> <evidence-regex>
fail() {
	echo "FAIL[$1]: $2"
	evidence "$3"
	echo "FAIL[$1]: full output: ${log}"
	exit 1
}

# 1. Extension failed to load. Checked first: when the library is missing the
#    driver also fails to parse, and "extension not loaded" is the real cause.
if grep -qaE "Can't open dynamic library|GDExtension dynamic library not found" <<<"$plain"; then
	fail extension-not-loaded \
		"the GDExtension could not be loaded, so no Go code ran and no test result is meaningful. Was 'make build' run?" \
		"Can't open dynamic library|GDExtension dynamic library not found|Cannot get class"
fi

# 2. Driver script failed to parse.
if grep -qaE 'Failed to load script.*Parse error' <<<"$plain"; then
	fail test-script-parse-error \
		"the GDScript test driver failed to parse, so no assertions ran" \
		'Failed to load script|Parse error'
fi

# 3. No summary banner. The catch-all that makes "silently did nothing" impossible.
if ! grep -qa 'TESTS FINISHED' <<<"$plain"; then
	fail no-driver-summary \
		"the test driver never printed its 'TESTS FINISHED' banner, so the suite did not complete (godot exit status ${godot_status})" \
		'ERROR|SCRIPT ERROR|handle_cr'
fi

# 4. Counts must be readable. Banner present but unparseable counts means the
#    banner format drifted away from this gate.
passes="$(printf '%s\n' "$plain" | sed -n 's/^[[:space:]]*PASSES:[[:space:]]*\([0-9][0-9]*\).*/\1/p' | tail -1)"
failures="$(printf '%s\n' "$plain" | sed -n 's/^[[:space:]]*FAILURES:[[:space:]]*\([0-9][0-9]*\).*/\1/p' | tail -1)"

if [[ -z "$passes" || -z "$failures" ]]; then
	fail unparseable-summary \
		"the 'TESTS FINISHED' banner is present but its PASSES/FAILURES counts could not be read; the driver's banner format and this gate have drifted apart" \
		'PASSES:|FAILURES:'
fi

total=$((passes + failures))

# 5. Zero assertions. A run that reports a summary but verified nothing.
if [[ "$total" -eq 0 ]]; then
	fail zero-tests \
		"the driver reported zero tests executed (PASSES: 0, FAILURES: 0) -- nothing was verified" \
		'PASSES:|FAILURES:'
fi

# 6. Assertion failures, reported with the driver's own count.
if [[ "$failures" -gt 0 ]]; then
	fail driver-reported-failures \
		"the driver reported ${failures} failure(s) out of ${total} assertions" \
		'== FAILURE|Expected .* but got'
fi

# 7. Leak detection, preserved from the original inline gate.
if grep -qaE 'ObjectDB instances were leaked|Leaked instance:' <<<"$plain"; then
	fail leaked-engine-objects \
		"leaked engine objects detected in test output" \
		'ObjectDB instances were leaked|Leaked instance:'
fi

# 8. Main-loop shutdown marker. The test extension's shutdown hook logs
#    GODOTGO_MAIN_LOOP_SHUTDOWN during engine teardown; a missing marker means
#    the shutdown callback never fired (or was never registered), so the run
#    cannot be trusted to have exercised the main-loop phases.
if ! grep -qa 'GODOTGO_MAIN_LOOP_SHUTDOWN' <<<"$plain"; then
	fail missing-main-loop-shutdown-marker \
		"the test extension's main-loop shutdown marker is missing: the registered shutdown callback never ran during teardown" \
		'MAIN_LOOP|GDExtensionManager'
fi

# 9. The gate found nothing wrong, yet godot still exited non-zero. Report it
#    rather than passing: something outside the known modes went wrong.
if [[ "$godot_status" -ne 0 ]]; then
	fail godot-exited-non-zero \
		"godot exited with status ${godot_status} although the driver reported a clean run (${passes} passed, 0 failed)" \
		'ERROR|SCRIPT ERROR|handle_cr|Segmentation fault'
fi

echo "PASS: extension loaded, ${total} assertions ran, 0 failures, main-loop shutdown marker present, no leaked engine objects"
exit 0
