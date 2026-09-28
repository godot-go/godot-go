extends Node

var test_passes := 0
var test_failures := 0

# Whether the most recent expect_rejected_call() was actually rejected. Read it
# immediately after the call -- it is a live result, not a record, and nested or
# interleaved uses overwrite it.
var last_call_was_rejected := false

func __get_stack_frame():
	var me = get_script()
	for s in get_stack():
		if s.source == me.resource_path:
			return s
	return null

func __assert_pass():
	test_passes += 1

func __assert_fail():
	test_failures += 1
	var s = __get_stack_frame()
	if s != null:
		print_rich ("[color=red] == FAILURE: In function %s() from '%s' on line %s[/color]" % [s.function, s.source, s.line])
	else:
		print_rich ("[color=red] == FAILURE (run with --debug to get more information!) ==[/color]")

func assert_equal(actual, expected):
	if actual == expected:
		__assert_pass()
	else:
		__assert_fail()
		print ("    |-> Expected '%s' but got '%s'" % [expected, actual])

func assert_true(v):
	assert_equal(v, true)

func assert_false(v):
	assert_equal(v, false)

func assert_not_equal(actual, expected):
	if actual != expected:
		__assert_pass()
	else:
		__assert_fail()
		print ("    |-> Expected '%s' NOT to equal '%s'" % [expected, actual])

# Makes a call the suite expects the engine to reject, without stranding the
# caller, and records whether the rejection actually happened.
#
# Why this exists: a rejected call leaves the caller in an awkward spot. Written
# with the dynamic call() form, the engine's rejection makes GDScript abandon the
# frame, so everything written after it -- assertions, free() calls -- silently
# stops executing while the suite still reports green. That is the worst kind of
# pass: coverage disappears without a failure to point at.
#
# callv() handles the same rejection differently: it reports it as an engine
# error and returns normally, so the caller keeps running. This primitive
# standardises on callv() so the caller is never stranded.
#
# Detection: a rejected call yields null; an accepted one yields the method's
# return value. last_call_was_rejected is set from that, after the call.
#
# Constraint: the target must return a value. A void method also yields null, so
# for a void target the flag cannot tell rejection apart from normal completion.
# Assert an observable side effect instead -- a counter the body increments, say.
#
# Read last_call_was_rejected immediately after the call:
#
#     expect_rejected_call(example, "takes_sprite", [a_label])
#     assert_true(last_call_was_rejected)
func expect_rejected_call(obj: Object, method: String, args: Array) -> void:
	# Marker first. If the run ever dies inside this call, the last line in the
	# log names the primitive and its target instead of the run just stopping.
	print ("expect_rejected_call: %s (rejection expected)" % method)

	# A mistyped method name would otherwise read as a rejection that worked.
	# Fail as the typo it actually is.
	if not obj.has_method(method):
		last_call_was_rejected = false
		__assert_fail()
		print ("    |-> expect_rejected_call: target has no method '%s' -- that is a typo, not a rejection" % method)
		return

	last_call_was_rejected = obj.callv(method, args) == null

# The banner below is machine-consumed: test/check_test_output.sh parses
# "TESTS FINISHED", "PASSES: <n>" and "FAILURES: <n>" to decide whether the
# suite genuinely ran. Do not reword these without updating that gate -- a
# mismatch fails loudly as "no driver summary", never as a false pass.
func exit_with_status() -> void:
	var success: bool = (test_failures == 0)
	print ("")
	print_rich ("[color=%s] ==== TESTS FINISHED ==== [/color]" % ("green" if success else "red"))
	print ("")
	print_rich ("   PASSES: [color=green]%s[/color]" % test_passes)
	print_rich ("   FAILURES: [color=red]%s[/color]" % test_failures)
	print ("")

	if success:
		print_rich("[color=green] ******** PASSED ******** [/color]")
	else:
		print_rich("[color=red] ******** FAILED ********[/color]")
	print("")

	get_tree().quit(0 if success else 1)
