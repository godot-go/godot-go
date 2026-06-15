extends "res://test_base.gd"

var custom_signal_emitted = null

class TestClass:
	func test(p_msg: String) -> String:
		return p_msg + " world"

func _ready():
	var example: Example = $Example
	test_suite(1, example)
	print("DEBUG: test_suite returned, about to exit")
	var tree = get_tree()
	tree.quit(0)

func test_suite(i: int, example: Example):
	print("test suite run %d" % [i])
	
	# Signal
	print("DEBUG: signal test")
	example.emit_custom_signal("Button", 42)
	assert_equal(custom_signal_emitted, ["Button", 42])
	
	# To string
	print("DEBUG: to_string test")
	assert_equal(example.to_string(),'[ GDExtension::Example <--> Instance ID:%s ]' % example.get_instance_id())
	
	# Simple methods
	print("DEBUG: simple func tests")
	example.simple_func()
	assert_equal(custom_signal_emitted, ['simple_func', 3])
	example.simple_const_func(123)
	assert_equal(custom_signal_emitted, ['simple_const_func', 4])
	
	# Pass core reference
	print("DEBUG: image ref test")
	assert_equal(example.image_ref_func(null), "invalid")
	var image = Image.new()
	assert_equal(example.image_ref_func(image), "valid")
	
	# Return values
	print("DEBUG: return something test")
	assert_equal(example.return_something("some string", 7.0/6, 7.0/6 * 1000, 2147483647, -127, -32768, 2147483647, 9223372036854775807), "1. some string42, 2. %.6f, 3. %f, 4. 2147483647, 5. -127, 6. -32768, 7. 2147483647, 8. 9223372036854775807" % [7.0/6, 7.0/6 * 1000])
	assert_equal(example.return_something_const(), get_viewport())
	var null_ref = example.return_empty_ref()
	assert_equal(null_ref, null)
	assert_equal(example.get_v4(), Vector4(1.2, 3.4, 5.6, 7.8))
	assert_equal(example.test_node_argument(example), example)
	
	# VarArg
	print("DEBUG: varargs tests")
	assert_equal(example.varargs_func("some", "arguments", "to", "test"), 4)
	assert_equal(example.varargs_func("some"), 1)
	assert_equal(example.varargs_func_nv("some", "arguments", "to", "test"), 46)
	example.varargs_func_void("some", "arguments", "to", "test")
	assert_equal(custom_signal_emitted, ["varargs_func_void", 5])
	
	# Default args
	print("DEBUG: def_args tests")
	assert_equal(example.def_args(), 300)
	assert_equal(example.def_args(50), 250)
	assert_equal(example.def_args(50, 100), 150)
	
	# Array
	print("DEBUG: array tests")
	assert_equal(example.test_array(), [1, 2])
	assert_equal(example.test_tarray(), [ Vector2(1, 2), Vector2(2, 3) ])
	assert_equal(example.test_dictionary(), {"hello": "world", "foo": "bar"})
	var array: Array[int] = [1, 2, 3]
	assert_equal(example.test_tarray_arg(array), 6)
	
	# Callable and bind
	print("DEBUG: callable_bind test")
	example.callable_bind()
	assert_equal(custom_signal_emitted, ["bound", 11])
	
	# String ops
	print("DEBUG: string_ops test")
	assert_equal(example.test_string_ops(), "ABCĎE")
	
	# Utility functions
	print("DEBUG: str utility test")
	assert_equal(example.test_str_utility(), "Hello, World! The answer is 42")
	print("DEBUG: instance_from_id test")
	assert_equal(example.test_instance_from_id_utility(), example)
	
	# Vector ops
	print("DEBUG: vector_ops test")
	assert_equal(example.test_vector_ops(), 105)
	
	# Properties
	print("DEBUG: properties test")
	assert_equal(example.group_subgroup_custom_position, Vector2(0, 0))
	example.group_subgroup_custom_position = Vector2(50, 50)
	assert_equal(example.group_subgroup_custom_position, Vector2(50, 50))
	
	# Constants
	print("DEBUG: constants test")
	assert_equal(Example.FIRST, 0)
	assert_equal(Example.ANSWER_TO_EVERYTHING, 42)
	assert_equal(Example.CONSTANT_WITHOUT_ENUM, 314)
	
	# BitFields
	print("DEBUG: bitfield test")
	assert_equal(Example.FLAG_ONE, 1)
	assert_equal(Example.FLAG_TWO, 2)
	assert_equal(example.test_bitfield(0), 0)
	assert_equal(example.test_bitfield(Example.FLAG_ONE | Example.FLAG_TWO), 3)
	
	# Virtual method
	print("DEBUG: virtual method test")
	var event = InputEventKey.new()
	event.key_label = KEY_H
	event.unicode = 72
	get_viewport().push_input(event)
	assert_equal(custom_signal_emitted, ["_input: H", 72])
	
	# GD extension class calls
	print("DEBUG: gd extension class calls")
	assert_equal(example.test_get_child_node("Label"), example.get_node("Label"))
	example.test_set_position_and_size(Vector2(320, 240), Vector2(100, 200))
	assert_equal(example.get_position(), Vector2(320, 240))
	assert_equal(example.get_size(), Vector2(100, 200))
	
	print("DEBUG: test_parent_is_nil test")
	assert_equal(example.test_parent_is_nil(), null)
	
	print("DEBUG: all tests passed, about to return from test_suite")
func _on_Example_custom_signal(signal_name, value):
	custom_signal_emitted = [signal_name, value]
