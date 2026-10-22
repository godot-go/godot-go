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
# The escape byte is built with printf rather than written as a sed hex escape:
# GNU sed understands one, BSD sed (macOS CI) does not, and would leave the
# colour codes in place, so every clean run would fail as "no driver summary".
esc="$(printf '\033')"
plain="$(sed -e "s/${esc}\\[[0-9;?]*[a-zA-Z]//g" "$log")"

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

# 9. Stacktrace capture probe. The Go suite opens one narrow capture window
#    (Example.test_stacktrace_probe) and prints "stacktrace-probe capture=<0|1>"
#    straight from the binary's compile-time truth. Two failure modes:
#    the section vanished entirely (driver drift), or the build CLAIMS capture
#    is available but no banner was printed (the Windows libbacktrace path or
#    the execinfo path silently failing -- this is the only in-repo proof they
#    run).
if ! grep -qa 'stacktrace-probe capture=' <<<"$plain"; then
	fail stacktrace-probe-missing \
		"the stacktrace probe section did not run: expected a 'stacktrace-probe capture=<0|1>' line from Example.test_stacktrace_probe" \
		'stacktrace-probe|TESTS FINISHED'
fi
if grep -qa 'stacktrace-probe capture=1' <<<"$plain" &&
	! grep -qa '___ C start backtrace ___' <<<"$plain"; then
	fail stacktrace-capture-claimed-but-silent \
		"the build reported native stacktrace capture available, but the enabled capture window printed no banner (expected '___ C start backtrace ___')" \
		'stacktrace-probe|___ C '
fi
#    Companion Go stack: the probe's hit phase prints the current goroutine's
#    stack via runtime.Stack -- file:line for the Go side on every platform,
#    independent of the native capture. Pure Go, so its absence means the
#    probe body regressed even where capture itself is stubbed out.
if grep -qa 'stacktrace-probe capture=' <<<"$plain" &&
	! grep -A4 'stacktrace-probe go-stack' <<<"$plain" | grep -q '\.go:[0-9]\+'; then
	fail stacktrace-go-stack-missing \
		"the stacktrace probe did not print its companion Go stack: expected a 'stacktrace-probe go-stack' line followed by frames carrying file.go:line" \
		'stacktrace-probe|\.go:'
fi

# 10. User-callable panic barrier. The driver invokes callables whose closures
#     panic (Example.create_panic_callable, openspec:
#     diagnostic-output-channel) and prints 'callable-panic-probe attempted=<n>'
#     afterwards. Conditioned on that marker like check 9 conditions on its
#     probe line, so a suite variant that skips the probe cannot fail here for
#     the wrong reason; when the probe did run, the boundary must have
#     recovered, reported the failed call, and recorded an entry -- the
#     GDScript assertions prove recovery and a live run, this proves the entry.
if grep -qa 'callable-panic-probe attempted=' <<<"$plain" &&
	! grep -qa 'go callable panicked' <<<"$plain"; then
	fail callable-panic-entry-missing \
		"a user callable closure panicked in the suite but the recoverable-panic boundary recorded no 'go callable panicked' entry" \
		'callable-panic-probe|go callable panicked|TESTS FINISHED'
fi

# 11. Signal duplicate gate degraded. ClassDBAddSignal consults the engine's
#     inheritance-inclusive member set (openspec: class-metadata-registration);
#     this warning means the consult failed and only Go-recorded names were
#     checked. The suite registers signals on every run, so the warning's
#     presence here always means drift: most likely the hand-written
#     class_has_signal method-bind hash in pkg/core/classdb.go no longer
#     matches the generated ClassDBImpl.ClassHasSignal body after an
#     extension_api refresh -- the silent engine-side drop this gate exists to
#     catch would otherwise return unnoticed.
if grep -qa 'inheritance was not checked' <<<"$plain"; then
	fail signal-gate-degraded \
		"the signal duplicate gate degraded to Go-recorded names only; re-check the class_has_signal method-bind hash in pkg/core/classdb.go against the generated ClassDBImpl.ClassHasSignal body" \
		'inheritance was not checked'
fi

# 12. The gate found nothing wrong, yet godot still exited non-zero. Report it
#    rather than passing: something outside the known modes went wrong.
if [[ "$godot_status" -ne 0 ]]; then
	fail godot-exited-non-zero \
		"godot exited with status ${godot_status} although the driver reported a clean run (${passes} passed, 0 failed)" \
		'ERROR|SCRIPT ERROR|handle_cr|Segmentation fault'
fi

echo "PASS: extension loaded, ${total} assertions ran, 0 failures, main-loop shutdown marker present, no leaked engine objects"
exit 0
