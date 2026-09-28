#!/usr/bin/env bash
#
# Fixture tests for test/check_test_output.sh.
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

if [[ ! -f "$gate" ]]; then
	echo "FAIL: gate script not found at ${gate}"
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
)

pass=0
fail=0

for entry in "${cases[@]}"; do
	IFS='|' read -r file status want_exit want_msg <<<"$entry"
	log="${here}/${file}"

	if [[ ! -f "$log" ]]; then
		echo "MISSING fixture: ${file}"
		fail=$((fail + 1))
		continue
	fi

	out="$(bash "$gate" "$log" "$status" 2>&1)"
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

# A run that produced no log at all must not pass.
out="$(bash "$gate" "${here}/does_not_exist.log" 0 2>&1)"
if [[ $? -ne 0 && "$out" == *"FAIL[harness]"* ]]; then
	echo "  ok   <absent log>  (exit 1, FAIL[harness])"
	pass=$((pass + 1))
else
	echo "  FAIL <absent log> did not report a harness failure"
	fail=$((fail + 1))
fi

echo ""
echo "fixture tests: ${pass} passed, ${fail} failed"
[[ "$fail" -eq 0 ]] || exit 1
