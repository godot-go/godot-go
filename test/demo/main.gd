extends Node

var exa = preload("example.gdextension")

func _ready():
	prints("is_library_open =", exa.is_library_open())

	# Bind signals
	prints("Signal bind")
	$Button.button_up.connect($Example.emit_custom_signal.bind("Button", 42))

	prints("")

	# Call static methods.
	# prints("Static method calls")
	# prints("  static (109)", Example.test_static(9, 100));
	# Example.test_static2();

	# # Call methods.
	# prints("Instance method calls")
	# $Example.simple_func()
	# ($Example as Example).simple_const_func(123) # Force use of ptrcall
	# prints("  returned", $Example.return_something("some string", 7.0/6, 7.0/6 * 1000, 2147483647, -127, -32768, 2147483647, 9223372036854775807))
	# prints("  returned const", $Example.return_something_const())
	# # prints("  returned ref", $Example.return_extended_ref())
	# prints("  returned ", $Example.get_v4())

	# prints("VarArg method calls")
	# var ref = ExampleRef.new()
	# prints("  sending ref: ", ref.get_instance_id(), "returned ref: ", $Example.extended_ref_checks(ref).get_instance_id())
	# prints("  vararg args", $Example.varargs_func("some", "arguments", "to", "test"))
	# prints("  vararg_nv ret", $Example.varargs_func_nv("some", "arguments", "to", "test"))
	# $Example.varargs_func_void("some", "arguments", "to", "test")

	prints("Method calls with default values")
	prints("  defval (300)", $Example.def_args())
	prints("  defval (250)", $Example.def_args(50))
	prints("  defval (150)", $Example.def_args(50, 100))

	prints("Array and Dictionary")
	prints("  test array", $Example.test_array())
	prints("  test dictionary", $Example.test_dictionary())

	prints("Properties")
	prints("  custom position is", $Example.group_subgroup_custom_position)
	$Example.group_subgroup_custom_position = Vector2(50, 50)
	prints("  custom position now is", $Example.group_subgroup_custom_position)

	prints("Constnts")
	prints("  FIRST", $Example.FIRST)
	prints("  ANSWER_TO_EVERYTHING", $Example.ANSWER_TO_EVERYTHING)
	prints("  CONSTANT_WITHOUT_ENUM", $Example.CONSTANT_WITHOUT_ENUM)

	prints("Others")
	prints("  CastTo", $Example.test_cast_to())

	prints("app is ready CI=", OS.get_environment("CI"))

	if OS.get_environment("CI") == "1":
		prints("CI env var detected: automating interactions")
		# force a button click event
		$Button.emit_signal("button_up")

		prints("force quit")
		get_tree().quit()

func _on_Example_custom_signal(signal_name, value):
	prints("Example emitted:", signal_name, value)
