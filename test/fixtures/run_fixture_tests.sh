#!/usr/bin/env bash
#
# Fixture tests for test/check_test_output.sh (runtime gate) and
# test/check_editor_output.sh (editor plugin gate).
#
# Each fixture is a captured-shaped Godot run log. The table below pins, for each
# one, the godot exit status to feed the gate, the exit status the gate must
# produce, and the message fragment that identifies the failure mode. This is
# what keeps the six failure modes distinguishable from each other: a gate that
# failed every case with the same message would pass a naive "exits non-zero"
# test and still be useless.
#
# Run: test/fixtures/run_fixture_tests.sh

set -uo pipefail

here="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
gate="${here}/../check_test_output.sh"
editor_gate="${here}/../check_editor_output.sh"

if [[ ! -f "$gate" ]]; then
	echo "FAIL: gate script not found at ${gate}"
	exit 1
fi

if [[ ! -f "$editor_gate" ]]; then
	echo "FAIL: editor gate script not found at ${editor_gate}"
	exit 1
fi

# fixture | godot_status | want_gate_exit | want_message_substring
cases=(
	"clean_pass.log|0|0|PASS: extension loaded"
	"missing_library.log|0|1|FAIL[extension-not-loaded]"
	"parse_error.log|0|1|FAIL[test-script-parse-error]"
	"no_banner.log|1|1|FAIL[no-driver-summary]"
	"zero_tests.log|0|1|FAIL[zero-tests]"
	"failures_one.log|1|1|FAIL[driver-reported-failures]"
	"leaked.log|0|1|FAIL[leaked-engine-objects]"
	"godot_nonzero.log|1|1|FAIL[godot-exited-non-zero]"
	"stacktrace_silent.log|0|1|FAIL[stacktrace-capture-claimed-but-silent]"
	"stacktrace_stub.log|0|0|PASS: extension loaded"
	"stacktrace_probe_missing.log|0|1|FAIL[stacktrace-probe-missing]"
	"stacktrace_no_go_stack.log|0|1|FAIL[stacktrace-go-stack-missing]"
)

# editor gate fixture | godot_status | want_gate_exit | want_message_substring
# editor_static_intern_ok.log pins the tolerance: engine SNAME static interns
# (static: 1) must pass while a non-static orphan must not.
editor_cases=(
	"editor_clean_pass.log|0|0|PASS: editor plugin fixture instantiated"
	"editor_missing_library.log|0|1|FAIL[extension-not-loaded]"
	"editor_no_fixture.log|0|1|FAIL[no-fixture-marker]"
	"editor_no_shutdown.log|0|1|FAIL[missing-main-loop-shutdown-marker]"
	"editor_leaked.log|0|1|FAIL[leaked-engine-objects]"
	"editor_orphan.log|0|1|FAIL[orphaned-string-names]"
	"editor_static_intern_ok.log|0|0|PASS: editor plugin fixture instantiated"
	"editor_nonzero.log|1|1|FAIL[godot-exited-non-zero]"
)

pass=0
fail=0

# run_cases <gate-script> <case>...  -- shared checker for both gate tables.
run_cases() {
	local run_gate="$1"
	shift
	local entry file status want_exit want_msg log out got_exit ok
	for entry in "$@"; do
		IFS='|' read -r file status want_exit want_msg <<<"$entry"
		log="${here}/${file}"

		if [[ ! -f "$log" ]]; then
			echo "MISSING fixture: ${file}"
			fail=$((fail + 1))
			continue
		fi

		out="$(bash "$run_gate" "$log" "$status" 2>&1)"
		got_exit=$?

		ok=1
		if [[ "$got_exit" -ne "$want_exit" ]]; then
			ok=0
		fi
		if [[ "$want_exit" -ne 0 ]] && [[ "$out" != *"$want_msg"* ]]; then
			ok=0
		fi

		if [[ "$ok" -eq 1 ]]; then
			echo "  ok   ${file}  (exit ${got_exit}, ${want_msg})"
			pass=$((pass + 1))
		else
			echo "  FAIL ${file}"
			echo "         wanted exit ${want_exit} containing '${want_msg}'"
			echo "         got    exit ${got_exit}"
			printf '%s\n' "$out" | sed 's/^/           | /'
			fail=$((fail + 1))
		fi
	done
}

run_cases "$gate" "${cases[@]}"
echo ""
echo "editor gate fixtures:"
run_cases "$editor_gate" "${editor_cases[@]}"

# A run that produced no log at all must not pass, for either gate.
for check_gate in "$gate" "$editor_gate"; do
	out="$(bash "$check_gate" "${here}/does_not_exist.log" 0 2>&1)"
	if [[ $? -ne 0 && "$out" == *"FAIL[harness]"* ]]; then
		echo "  ok   <absent log vs $(basename "$check_gate")>  (exit 1, FAIL[harness])"
		pass=$((pass + 1))
	else
		echo "  FAIL <absent log vs $(basename "$check_gate")> did not report a harness failure"
		fail=$((fail + 1))
	fi
done

echo ""
echo "fixture tests: ${pass} passed, ${fail} failed"
[[ "$fail" -eq 0 ]] || exit 1
