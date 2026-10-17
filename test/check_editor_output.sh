#!/usr/bin/env bash
#
# Gate for the headless editor run (editor plugin harness, openspec change
# support-editor-plugins).
#
# The runtime gate (check_test_output.sh) keys on the GDScript driver banner,
# which does not exist in editor mode: the editor never runs the main scene.
# This gate instead keys on the Go-side fixture marker that fires when the
# engine instantiates the Go plugin class (GODOTGO_EDITOR_PLUGIN_ENTERED from
# the bound _enter_tree virtual), the main-loop shutdown marker, leak reports,
# and exit-time orphan StringNames. Exit 0 means all of: the extension
# loaded, the plugin fixture was instantiated by the editor, the shutdown
# hook fired, nothing leaked, no heap-interned StringName outlived the run,
# and Godot exited zero.
#
# Orphan StringName lines are only emitted with --verbose. Lines with
# "static: 1" are engine SNAME static interns by design (they appear the
# first time the plugin add/virtual paths run) and are tolerated; any
# non-static ("static: 0") orphan is an extension-owned allocation that never
# released its name, and fails the gate.
#
# Usage: check_editor_output.sh <log-path> <godot-exit-status>

set -uo pipefail

log="${1:-}"
godot_status="${2:-0}"

if [[ -z "$log" ]]; then
	echo "FAIL[harness]: usage: check_editor_output.sh <log-path> <godot-exit-status>"
	exit 1
fi

if [[ ! -f "$log" ]]; then
	echo "FAIL[harness]: editor test output log not found: ${log}"
	echo "FAIL[harness]: the editor test run produced no log at all"
	exit 1
fi

plain="$(sed -e 's/\x1b\[[0-9;?]*[a-zA-Z]//g' "$log")"

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

# 1. Extension failed to load: nothing Go-side could have run.
if printf '%s\n' "$plain" | grep -qaE "Can't open dynamic library|GDExtension dynamic library not found"; then
	fail extension-not-loaded \
		"the GDExtension could not be loaded in the editor, so no Go code ran and no plugin result is meaningful. Was 'make build' run?" \
		"Can't open dynamic library|GDExtension dynamic library not found"
fi

# 2. Plugin fixture marker. The engine instantiates the fixture plugin class
#    when the EditorNode drains the queued add; the bound _enter_tree virtual
#    prints this marker from Go. Missing means the registration never took
#    effect (or the editor never reached that startup phase).
if ! printf '%s\n' "$plain" | grep -qa 'GODOTGO_EDITOR_PLUGIN_ENTERED'; then
	fail no-fixture-marker \
		"the editor plugin fixture was never instantiated by the editor (marker GODOTGO_EDITOR_PLUGIN_ENTERED missing); registration did not take effect (godot exit status ${godot_status})" \
		'GODOTGO_EDITOR_PLUGIN|GDExtensionManager|ERROR'
fi

# 3. Main-loop shutdown marker: the editor teardown must have run the
#    registered main-loop shutdown hook.
if ! printf '%s\n' "$plain" | grep -qa 'GODOTGO_MAIN_LOOP_SHUTDOWN'; then
	fail missing-main-loop-shutdown-marker \
		"the main-loop shutdown marker is missing: the editor teardown never reached the shutdown callback, so the run cannot be trusted to have exercised editor shutdown" \
		'MAIN_LOOP|GDExtensionManager'
fi

# 4. ObjectDB leaks.
if printf '%s\n' "$plain" | grep -qaE 'ObjectDB instances were leaked|Leaked instance:'; then
	fail leaked-engine-objects \
		"leaked engine objects detected in editor test output" \
		'ObjectDB instances were leaked|Leaked instance:'
fi

# 5. Heap-interned StringNames outliving the run (static: 0). Static SNAME
#    interns (static: 1) are engine-internal and tolerated; see header.
if printf '%s\n' "$plain" | grep -qaE 'Orphan StringName: .*\(static: 0,'; then
	fail orphaned-string-names \
		"non-static orphan StringNames outlived the editor run: the extension allocated names it never released" \
		'Orphan StringName'
fi

# 6. Non-zero godot exit despite everything above looking clean.
if [[ "$godot_status" -ne 0 ]]; then
	fail godot-exited-non-zero \
		"godot exited with status ${godot_status} although the editor fixture markers were clean" \
		'ERROR|SCRIPT ERROR|handle_cr|Segmentation fault'
fi

echo "PASS: editor plugin fixture instantiated, main-loop shutdown marker present, no leaked engine objects, no orphaned StringNames"
exit 0
