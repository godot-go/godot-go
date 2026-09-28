extends "res://test_base.gd"

var custom_signal_emitted = null

class TestClass:
	func test(p_msg: String) -> String:
		return p_msg + " world"

func _ready():
	var example: Example = $Example
	test_suite(1, example)
	test_object_args(example)
	test_pin_scratch_stays_flat(example)
	test_return_ownership(example)
	test_ptrcall_object_decode(example)
	test_user_defined_object_args(example)
	test_varcall_arg_rejection(example)
	# example.group_subgroup_custom_position = Vector2(0, 0)
	# custom_signal_emitted = null
	# var t = get_tree()
	# if t != null:
	# 	await t.create_timer(3.0).timeout
	# test_suite(2, example)
	exit_with_status()

func test_suite(i: int, example: Example):
	print("test suite run %d" % [i])
	# Signal.
	example.emit_custom_signal("Button", 42)
	assert_equal(custom_signal_emitted, ["Button", 42])

	# To string. The expected value carries a non-ASCII marker that only the Go
	# V_Example_ToString virtual produces -- the binding's fallback format is
	# ASCII-only, so this fails if dispatch or the UTF-8 encoding breaks.
	assert_equal(example.to_string(),'[ GDExtension::Example <--> Instance ID:%s | é中 ]' % example.get_instance_id())
	# It appears there's a bug with instance ids :-(
	#assert_equal($Example/ExampleMin.to_string(), 'ExampleMin:[Wrapped:%s]' % $Example/ExampleMin.get_instance_id())

	# godot-go will probably not support static methods since they don't exist in go
	# Call static methods.
	# assert_equal(Example.test_static(9, 100), 109);
	# It's void and static, so all we know is that it didn't crash.
	# Example.test_static2()

	# Property list.
	example.property_from_list = Vector3(100, 200, 300)
	assert_equal(example.property_from_list, Vector3(100, 200, 300))
	var prop_list = example.get_property_list()
	for prop_info in prop_list:
		if prop_info['name'] == 'mouse_filter':
			assert_equal(prop_info['usage'], PROPERTY_USAGE_NO_EDITOR)

	# Call simple methods.
	example.simple_func()
	assert_equal(custom_signal_emitted, ['simple_func', 3])
	example.simple_const_func(123)
	assert_equal(custom_signal_emitted, ['simple_const_func', 4])

	# Pass custom reference.
	# assert_equal(example.custom_ref_func(null), -1)
	# var ref1 = ExampleRef.new()
	# ref1.id = 27
	# assert_equal(example.custom_ref_func(ref1), 27)
	# ref1.id += 1;
	# assert_equal(example.custom_const_ref_func(ref1), 28)

	# Pass core reference.
	assert_equal(example.image_ref_func(null), "invalid")
	# assert_equal(example.image_const_ref_func(null), "invalid")
	var image = Image.new()
	assert_equal(example.image_ref_func(image), "valid")
	# assert_equal(example.image_const_ref_func(image), "valid")

	# Ref round-trip: Go returns a Ref<Image> to Godot (encode) and receives one (decode).
	var ret_ref = example.pass_through_ref_image(image)
	assert_not_equal(ret_ref, null)
	assert_equal(ret_ref, image)
	assert_equal(example.return_nil_ref_image(), null)

	# Ref count lifecycle on a RefCounted object passed from Godot.
	assert_equal(example.test_ref_lifecycle(image), 1)
	assert_equal(example.test_finalizer_release(image), 1)

	# Basic built-in types: String / StringName / NodePath args (ptrcall).
	assert_equal(example.test_string_arg_echo("hello"), "hello")
	assert_equal(example.test_string_name_arg_echo(&"my_name"), "my_name")
	assert_equal(example.test_node_path_arg_echo(NodePath("root/child")), "root/child")

	# Basic built-in types via Variant call (varcall).
	assert_equal(example.call("test_string_arg_echo", "hello"), "hello")
	assert_equal(example.call("test_string_name_arg_echo", &"my_name"), "my_name")
	assert_equal(example.call("test_node_path_arg_echo", NodePath("root/child")), "root/child")

	# StringName / NodePath returns (ptrcall and varcall).
	assert_equal(example.test_return_string_name(), &"returned_name")
	assert_equal(example.test_return_node_path(), NodePath("root/child"))
	assert_equal(example.call("test_return_string_name"), &"returned_name")
	assert_equal(example.call("test_return_node_path"), NodePath("root/child"))

	# Borrowed StringName/NodePath echoes must not be destroyed by the return
	# path in either call style; fresh differently-valued returns exercise the
	# destroy-after-encode accounting. Repeated iterations surface refcount
	# drift as orphan warnings or crashes.
	for _i in range(50):
		assert_equal(example.test_echo_string_name_arg(&"echo_name"), &"echo_name")
		assert_equal(example.call("test_echo_string_name_arg", &"echo_name"), &"echo_name")
		assert_equal(example.test_fresh_string_name(&"echo_name"), &"returned_fresh_name")
		assert_equal(example.call("test_fresh_string_name", &"echo_name"), &"returned_fresh_name")
		assert_equal(example.test_echo_node_path_arg(NodePath("echo/path")), NodePath("echo/path"))
		assert_equal(example.call("test_echo_node_path_arg", NodePath("echo/path")), NodePath("echo/path"))
		assert_equal(example.test_fresh_node_path(NodePath("echo/path")), NodePath("fresh/path"))
		assert_equal(example.call("test_fresh_node_path", NodePath("echo/path")), NodePath("fresh/path"))

	# Scalar echo (bool / int64 / float64 / Go string).
	assert_equal(example.test_scalar_echo(true, 42, 1.5, "scalar"), "true:42:1.5:scalar")
	assert_equal(example.call("test_scalar_echo", true, 42, 1.5, "scalar"), "true:42:1.5:scalar")

	# Large uint64 preserves its bit pattern.
	assert_equal(example.test_uint64_echo(1 << 63), 1 << 63)
	assert_equal(example.call("test_uint64_echo", 1 << 63), 1 << 63)

	# Narrow numeric returns (varcall widening) + ptrcall.
	assert_equal(example.test_return_int8(), 127)
	assert_equal(example.call("test_return_int8"), 127)
	assert_equal(example.test_return_uint16(), 65535)
	assert_equal(example.call("test_return_uint16"), 65535)
	assert_equal(example.test_return_float32(), 1.5)
	assert_equal(example.call("test_return_float32"), 1.5)

	# Repeated Go string returns via varcall (exercise lifecycle).
	var repeated_count: int = 0
	for n in range(500):
		assert_equal(example.call("test_go_string_repeat"), "repeat_me_string_for_leak_check")
		repeated_count += 1
	assert_equal(repeated_count, 500)

	# Vector built-in types: echo round-trips (ptrcall and varcall).
	var _v2i := Vector2i(1, 2)
	assert_equal(example.test_echo_vector2i(_v2i), _v2i)
	assert_equal(example.call("test_echo_vector2i", _v2i), _v2i)

	var _v3 := Vector3(1.0, 2.0, 3.0)
	assert_equal(example.test_echo_vector3(_v3), _v3)
	assert_equal(example.call("test_echo_vector3", _v3), _v3)

	var _v3i := Vector3i(1, 2, 3)
	assert_equal(example.test_echo_vector3i(_v3i), _v3i)
	assert_equal(example.call("test_echo_vector3i", _v3i), _v3i)

	var _rect2 := Rect2(1.0, 2.0, 3.0, 4.0)
	assert_equal(example.test_echo_rect2(_rect2), _rect2)
	assert_equal(example.call("test_echo_rect2", _rect2), _rect2)

	var _rect2i := Rect2i(1, 2, 3, 4)
	assert_equal(example.test_echo_rect2i(_rect2i), _rect2i)
	assert_equal(example.call("test_echo_rect2i", _rect2i), _rect2i)

	var _t2d := Transform2D.IDENTITY
	assert_equal(example.test_echo_transform2d(_t2d), _t2d)
	assert_equal(example.call("test_echo_transform2d", _t2d), _t2d)

	var _v4i := Vector4i(1, 2, 3, 4)
	assert_equal(example.test_echo_vector4i(_v4i), _v4i)
	assert_equal(example.call("test_echo_vector4i", _v4i), _v4i)

	var _plane := Plane(1.0, 0.0, 0.0, 5.0)
	assert_equal(example.test_echo_plane(_plane), _plane)
	assert_equal(example.call("test_echo_plane", _plane), _plane)

	var _quat := Quaternion(0.0, 0.0, 0.0, 1.0)
	assert_equal(example.test_echo_quaternion(_quat), _quat)
	assert_equal(example.call("test_echo_quaternion", _quat), _quat)

	var _aabb := AABB(Vector3(0, 0, 0), Vector3(10, 10, 10))
	assert_equal(example.test_echo_aabb(_aabb), _aabb)
	assert_equal(example.call("test_echo_aabb", _aabb), _aabb)

	var _basis := Basis.IDENTITY
	assert_equal(example.test_echo_basis(_basis), _basis)
	assert_equal(example.call("test_echo_basis", _basis), _basis)

	var _t3d := Transform3D.IDENTITY
	assert_equal(example.test_echo_transform3d(_t3d), _t3d)
	assert_equal(example.call("test_echo_transform3d", _t3d), _t3d)

	var _proj := Projection(Vector4(1, 0, 0, 0), Vector4(0, 1, 0, 0), Vector4(0, 0, 1, 0), Vector4(0, 0, 0, 1))
	assert_equal(example.test_echo_projection(_proj), _proj)
	assert_equal(example.call("test_echo_projection", _proj), _proj)

	# Container built-in types
	var _arr := Array([1, 2, 3])
	assert_equal(example.test_echo_array(_arr), _arr)
	assert_equal(example.call("test_echo_array", _arr), _arr)

	var _pba := PackedByteArray([10, 20, 30])
	assert_equal(example.test_echo_packed_byte_array(_pba), _pba)
	assert_equal(example.call("test_echo_packed_byte_array", _pba), _pba)

	var _pi32 := PackedInt32Array([1, 2, 3])
	assert_equal(example.test_echo_packed_int32_array(_pi32), _pi32)
	assert_equal(example.call("test_echo_packed_int32_array", _pi32), _pi32)

	var _pi64 := PackedInt64Array([1, 2, 3])
	assert_equal(example.test_echo_packed_int64_array(_pi64), _pi64)
	assert_equal(example.call("test_echo_packed_int64_array", _pi64), _pi64)

	var _pf32 := PackedFloat32Array([1.5, 2.5, 3.5])
	assert_equal(example.test_echo_packed_float32_array(_pf32), _pf32)
	assert_equal(example.call("test_echo_packed_float32_array", _pf32), _pf32)

	var _pf64 := PackedFloat64Array([1.5, 2.5, 3.5])
	assert_equal(example.test_echo_packed_float64_array(_pf64), _pf64)
	assert_equal(example.call("test_echo_packed_float64_array", _pf64), _pf64)

	var _psa := PackedStringArray(["hello", "world"])
	assert_equal(example.test_echo_packed_string_array(_psa), _psa)
	assert_equal(example.call("test_echo_packed_string_array", _psa), _psa)

	var _pv2a := PackedVector2Array([Vector2(1, 2), Vector2(3, 4)])
	assert_equal(example.test_echo_packed_vector2_array(_pv2a), _pv2a)
	assert_equal(example.call("test_echo_packed_vector2_array", _pv2a), _pv2a)

	var _pv3a := PackedVector3Array([Vector3(1, 2, 3), Vector3(4, 5, 6)])
	assert_equal(example.test_echo_packed_vector3_array(_pv3a), _pv3a)
	assert_equal(example.call("test_echo_packed_vector3_array", _pv3a), _pv3a)

	var _pv4a := PackedVector4Array([Vector4(1, 2, 3, 4), Vector4(5, 6, 7, 8)])
	assert_equal(example.test_echo_packed_vector4_array(_pv4a), _pv4a)
	assert_equal(example.call("test_echo_packed_vector4_array", _pv4a), _pv4a)

	var _pca := PackedColorArray([Color(1, 0, 0), Color(0, 1, 0)])
	assert_equal(example.test_echo_packed_color_array(_pca), _pca)
	assert_equal(example.call("test_echo_packed_color_array", _pca), _pca)

	# Container arg consume: correct contents observed, no retained reference.
	assert_equal(example.test_consume_array(_arr), _arr.size())
	assert_equal(example.call("test_consume_array", _arr), _arr.size())
	assert_equal(example.test_consume_packed_byte_array(_pba), _pba.size())
	assert_equal(example.call("test_consume_packed_byte_array", _pba), _pba.size())
	assert_equal(example.test_consume_packed_int32_array(_pi32), _pi32.size())
	assert_equal(example.call("test_consume_packed_int32_array", _pi32), _pi32.size())
	assert_equal(example.test_consume_packed_int64_array(_pi64), _pi64.size())
	assert_equal(example.call("test_consume_packed_int64_array", _pi64), _pi64.size())
	assert_equal(example.test_consume_packed_float32_array(_pf32), _pf32.size())
	assert_equal(example.call("test_consume_packed_float32_array", _pf32), _pf32.size())
	assert_equal(example.test_consume_packed_float64_array(_pf64), _pf64.size())
	assert_equal(example.call("test_consume_packed_float64_array", _pf64), _pf64.size())
	assert_equal(example.test_consume_packed_string_array(_psa), _psa.size())
	assert_equal(example.call("test_consume_packed_string_array", _psa), _psa.size())
	assert_equal(example.test_consume_packed_vector2_array(_pv2a), _pv2a.size())
	assert_equal(example.call("test_consume_packed_vector2_array", _pv2a), _pv2a.size())
	assert_equal(example.test_consume_packed_vector3_array(_pv3a), _pv3a.size())
	assert_equal(example.call("test_consume_packed_vector3_array", _pv3a), _pv3a.size())
	assert_equal(example.test_consume_packed_color_array(_pca), _pca.size())
	assert_equal(example.call("test_consume_packed_color_array", _pca), _pca.size())

	# Borrowed ptrcall container arguments share storage with the caller.
	# Godot Array/Dictionary are shared-reference types (no copy-on-write),
	# so a Go-side mutation is visible to the caller through either call
	# style. Packed*Array borrows are READ-ONLY: the borrow holds no
	# refcount, so a mutating call may free or reallocate the caller's
	# buffer (undefined behavior) and is deliberately not exercised here.
	var _mut_arr := Array([1, 2])
	example.test_mutate_array(_mut_arr)
	assert_equal(_mut_arr, [1, 2, 42])
	example.call("test_mutate_array", _mut_arr)
	assert_equal(_mut_arr, [1, 2, 42, 42])

	# Byte-identical defensive clones are classified as borrow echoes
	# (contents round-trip in both call styles); a rebuilt array differs from
	# every borrowed argument and its owned reference is destroyed by the
	# return path, leaving the caller's array untouched.
	var _cl_arr := Array(["a", "b"])
	assert_equal(example.test_clone_echo_array(_cl_arr), _cl_arr)
	assert_equal(example.call("test_clone_echo_array", _cl_arr), _cl_arr)
	assert_equal(example.test_rebuild_array_plus(_cl_arr), Array(["a", "b", 7]))
	assert_equal(example.call("test_rebuild_array_plus", _cl_arr), Array(["a", "b", 7]))
	assert_equal(_cl_arr, Array(["a", "b"]))

	# Dictionary arguments: echo round-trips contents and consume observes
	# the key count without retaining a Go-side reference.
	var _dict := {"a": 1, "b": 2}
	assert_equal(example.call("test_echo_dictionary", _dict), _dict)
	assert_equal(example.call("test_consume_dictionary", _dict), 2)

	# Shared-reference mutation: Dictionary has no copy-on-write, so a key
	# added by the Go method must be visible in the caller's original
	# dictionary after the call.
	example.call("test_mutate_dictionary", _dict)
	assert_true(_dict.has("go_added"))
	assert_equal(_dict["go_added"], 42)

	# Hierarchical virtual methods (qualified V_ClassName_Method names).
	# The base level serves its own instance; the derived instance dispatches
	# to the most-derived implementation, whose composed value proves explicit
	# delegation to the base-level implementation. Example never implements
	# this virtual, so it still receives the engine default (-1, -1),
	# pinning unimplemented fallback across the migration.
	var _hbase := TestHierarchicalBase.new()
	assert_equal(_hbase.get_maximum_size(), Vector2(11, 22))
	var hderived := TestHierarchicalDerived.new()
	assert_equal(hderived.get_maximum_size(), Vector2(111, 22))
	assert_equal(example.get_maximum_size(), Vector2(-1, -1))
	_hbase.free()
	hderived.free()

	# Direct ptrcall dispatch with a Dictionary argument is unsupported
	# today: the ptrcall argument decode has no Dictionary case
	# (reflectFuncCallArgsFromGDExtensionConstTypePtrSliceArgs), so a call
	# like `example.test_echo_dictionary(_dict)` panics deterministically
	# with "MethodBind.Ptrcall reflected as array does not support type:
	# Dictionary" and aborts the process. That failure is intentionally NOT
	# exercised here; varcall (.call) is the only supported style for
	# Dictionary arguments. If a ptrcall decode case lands later, add echo/
	# consume assertions for that call style alongside this comment.

	# Return values.
	assert_equal(example.return_something("some string", 7.0/6, 7.0/6 * 1000, 2147483647, -127, -32768, 2147483647, 9223372036854775807), "1. some string42, 2. %.6f, 3. %f, 4. 2147483647, 5. -127, 6. -32768, 7. 2147483647, 8. 9223372036854775807" % [7.0/6, 7.0/6 * 1000])
	assert_equal(example.return_something_const(), get_viewport())
	var null_ref = example.return_empty_ref()
	assert_equal(null_ref, null)
	# var ret_ref = example.return_extended_ref()
	# assert_not_equal(ret_ref.get_instance_id(), 0)
	# assert_equal(ret_ref.get_id(), 0)
	assert_equal(example.get_v4(), Vector4(1.2, 3.4, 5.6, 7.8))
	assert_equal(example.test_node_argument(example), example)

	# VarArg method calls.
	# var var_ref = ExampleRef.new()
	# assert_not_equal(example.extended_ref_checks(var_ref).get_instance_id(), var_ref.get_instance_id())
	assert_equal(example.varargs_func("some", "arguments", "to", "test"), 4)
	assert_equal(example.varargs_func("some"), 1)
	assert_equal(example.varargs_func(), 0)
	assert_equal(example.varargs_func_nv("some", "arguments", "to", "test"), 46)
	example.varargs_func_void("some", "arguments", "to", "test")
	assert_equal(custom_signal_emitted, ["varargs_func_void", 5])

	# Method calls with default values.
	assert_equal(example.def_args(), 300)
	assert_equal(example.def_args(50), 250)
	assert_equal(example.def_args(50, 100), 150)

	# Partial trailing default: a supplied, b defaults to 200.
	assert_equal(example.partial_def_args(5), 205)
	assert_equal(example.partial_def_args(5, 7), 12)
	# Zero args is unsatisfiable (missing 2 > 1 default) and must be rejected.
	assert_equal(example.test_partial_def_args_rejects_short(), 0)

	# Varcall argument validation: too-few/too-many/non-convertible calls must be
	# reported to the engine (not abort the process) and must never run the bound
	# body; a correct call must dispatch. Returning 0 means all held; reaching the
	# next assertion proves the process survived every rejected call.
	assert_equal(example.test_call_error_reporting(), 0)

	# Array and Dictionary
	assert_equal(example.test_array(), [1, 2])
	assert_equal(example.test_tarray(), [ Vector2(1, 2), Vector2(2, 3) ])
	assert_equal(example.test_dictionary(), {"hello": "world", "foo": "bar"})
	var array: Array[int] = [1, 2, 3]
	print("array: ", array)
	assert_equal(example.test_tarray_arg(array), 6)

	example.callable_bind()
	assert_equal(custom_signal_emitted, ["bound", 11])

	# String += operator
	assert_equal(example.test_string_ops(), "ABCĎE")

	# UtilityFunctions::str()
	assert_equal(example.test_str_utility(), "Hello, World! The answer is 42")

	# UtilityFunctions::instance_from_id()
	assert_equal(example.test_instance_from_id_utility(), example)

	# # Test converting string to char* and doing comparison.
	# assert_equal(example.test_string_is_fourty_two("blah"), false)
	# assert_equal(example.test_string_is_fourty_two("fourty two"), true)

	# # String::resize().
	# assert_equal(example.test_string_resize("What"), "What!?")

	# mp_callable() with void method.
	# var mp_callable: Callable = example.test_callable_mp()
	# assert_equal(mp_callable.is_valid(), true)
	# mp_callable.call(example, "void", 36)
	# assert_equal(custom_signal_emitted, ["unbound_method1: Example - void", 36])

	# # Check that it works with is_connected().
	# assert_equal(example.renamed.is_connected(mp_callable), false)
	# example.renamed.connect(mp_callable)
	# assert_equal(example.renamed.is_connected(mp_callable), true)
	# # Make sure a new object is still treated as equivalent.
	# assert_equal(example.renamed.is_connected(example.test_callable_mp()), true)
	# assert_equal(mp_callable.hash(), example.test_callable_mp().hash())
	# example.renamed.disconnect(mp_callable)
	# assert_equal(example.renamed.is_connected(mp_callable), false)

	# # mp_callable() with return value.
	# var mp_callable_ret: Callable = example.test_callable_mp_ret()
	# assert_equal(mp_callable_ret.call(example, "test", 77), "unbound_method2: Example - test - 77")

	# # mp_callable() with const method and return value.
	# var mp_callable_retc: Callable = example.test_callable_mp_retc()
	# assert_equal(mp_callable_retc.call(example, "const", 101), "unbound_method3: Example - const - 101")

	# # mp_callable_static() with void method.
	# var mp_callable_static: Callable = example.test_callable_mp_static()
	# mp_callable_static.call(example, "static", 83)
	# assert_equal(custom_signal_emitted, ["unbound_static_method1: Example - static", 83])

	# # Check that it works with is_connected().
	# assert_equal(example.renamed.is_connected(mp_callable_static), false)
	# example.renamed.connect(mp_callable_static)
	# assert_equal(example.renamed.is_connected(mp_callable_static), true)
	# # Make sure a new object is still treated as equivalent.
	# assert_equal(example.renamed.is_connected(example.test_callable_mp_static()), true)
	# assert_equal(mp_callable_static.hash(), example.test_callable_mp_static().hash())
	# example.renamed.disconnect(mp_callable_static)
	# assert_equal(example.renamed.is_connected(mp_callable_static), false)

	# # mp_callable_static() with return value.
	# var mp_callable_static_ret: Callable = example.test_callable_mp_static_ret()
	# assert_equal(mp_callable_static_ret.call(example, "static-ret", 84), "unbound_static_method2: Example - static-ret - 84")

	# # CallableCustom.
	# var custom_callable: Callable = example.test_custom_callable();
	# assert_equal(custom_callable.is_custom(), true);
	# assert_equal(custom_callable.is_valid(), true);
	# assert_equal(custom_callable.call(), "Hi")
	# assert_equal(custom_callable.hash(), 27);
	# assert_equal(custom_callable.get_object(), null);
	# assert_equal(custom_callable.get_method(), "");
	# assert_equal(str(custom_callable), "<MyCallableCustom>");

	# PackedArray iterators
	assert_equal(example.test_vector_ops(), 105)
	# assert_equal(example.test_vector_init_list(), 105)

	# Properties.
	assert_equal(example.group_subgroup_custom_position, Vector2(0, 0))
	example.group_subgroup_custom_position = Vector2(50, 50)
	assert_equal(example.group_subgroup_custom_position, Vector2(50, 50))

	# # Test Object::cast_to<>() and that correct wrappers are being used.
	# var control = Control.new()
	# var sprite = Sprite2D.new()
	# var example_ref = ExampleRef.new()

	# assert_equal(example.test_object_cast_to_node(control), true)
	# assert_equal(example.test_object_cast_to_control(control), true)
	# assert_equal(example.test_object_cast_to_example(control), false)

	# assert_equal(example.test_object_cast_to_node(example), true)
	# assert_equal(example.test_object_cast_to_control(example), true)
	# assert_equal(example.test_object_cast_to_example(example), true)

	# assert_equal(example.test_object_cast_to_node(sprite), true)
	# assert_equal(example.test_object_cast_to_control(sprite), false)
	# assert_equal(example.test_object_cast_to_example(sprite), false)

	# assert_equal(example.test_object_cast_to_node(example_ref), false)
	# assert_equal(example.test_object_cast_to_control(example_ref), false)
	# assert_equal(example.test_object_cast_to_example(example_ref), false)

	# control.queue_free()
	# sprite.queue_free()

	# Test conversions to and from Variant.
	# assert_equal(example.test_variant_vector2i_conversion(Vector2i(1, 1)), Vector2i(1, 1))
	# assert_equal(example.test_variant_vector2i_conversion(Vector2(1.0, 1.0)), Vector2i(1, 1))
	# assert_equal(example.test_variant_int_conversion(10), 10)
	# assert_equal(example.test_variant_int_conversion(10.0), 10)
	# assert_equal(example.test_variant_float_conversion(10.0), 10.0)
	# assert_equal(example.test_variant_float_conversion(10), 10.0)

	# # Test that ptrcalls from GDExtension to the engine are correctly encoding Object and RefCounted.
	# var new_node = Node.new()
	# example.test_add_child(new_node)
	# assert_equal(new_node.get_parent(), example)

	# var new_tileset = TileSet.new()
	# var new_tilemap = TileMap.new()
	# example.test_set_tileset(new_tilemap, new_tileset)
	# assert_equal(new_tilemap.tile_set, new_tileset)
	# new_tilemap.queue_free()

	# # Test variant call.
	# var test_obj = TestClass.new()
	# assert_equal(example.test_variant_call(test_obj), "hello world")

	# Constants.
	assert_equal(Example.FIRST, 0)
	assert_equal(Example.ANSWER_TO_EVERYTHING, 42)
	assert_equal(Example.CONSTANT_WITHOUT_ENUM, 314)

	# BitFields.
	assert_equal(Example.FLAG_ONE, 1)
	assert_equal(Example.FLAG_TWO, 2)
	assert_equal(example.test_bitfield(0), 0)
	assert_equal(example.test_bitfield(Example.FLAG_ONE | Example.FLAG_TWO), 3)

	# Test variant iterator.
	# assert_equal(example.test_variant_iterator([10, 20, 30]), [15, 25, 35])
	# assert_equal(example.test_variant_iterator(null), "iter_init: not valid")

	# RPCs.
	# assert_equal(example.return_last_rpc_arg(), 0)
	# example.test_rpc(42)
	# assert_equal(example.return_last_rpc_arg(), 42)
	# example.test_send_rpc(100)
	# assert_equal(example.return_last_rpc_arg(), 100)

	# Virtual method.
	var event = InputEventKey.new()
	event.key_label = KEY_H
	event.unicode = 72
	get_viewport().push_input(event)
	assert_equal(custom_signal_emitted, ["_input: H", 72])

	# gd extension class calls
	assert_equal(example.test_get_child_node("Label"), example.get_node("Label"))
	example.test_set_position_and_size(Vector2(320, 240), Vector2(100, 200))
	assert_equal(example.get_position(), Vector2(320, 240))
	assert_equal(example.get_size(), Vector2(100, 200))
	# example.test_cast_to()

	# var body = CharacterBody2D.new()
	# var motion = Vector2(1.0, 2.0)
	# body.move_and_collide(motion, true, 0.5, true)
	# example.test_character_body_2d(body)
	# body.queue_free()

	assert_equal(example.test_parent_is_nil(), null)

	# Property revert virtuals (creation-info callbacks).
	# property_from_list was set to a non-default value above, so only it is
	# revertable; dproperty_0 never claims revert support.
	assert_equal(example.property_can_revert("property_from_list"), true)
	assert_equal(example.property_can_revert("dproperty_0"), false)
	assert_equal(example.property_get_revert("property_from_list"), Vector3(42, 42, 42))
	# A class that never bound the revert virtuals reports no revert support.
	var _hnoRevert := TestHierarchicalBase.new()
	assert_equal(_hnoRevert.property_can_revert("property_from_list"), false)
	assert_equal(_hnoRevert.property_get_revert("property_from_list"), null)

	# Notification dispatch (creation-info callback).
	var codes_before: int = example.get_notification_codes().size()
	var probe_code: int = 424242
	example.notification(probe_code, false)
	var codes: PackedInt32Array = example.get_notification_codes()
	assert_equal(codes.size(), codes_before + 1)
	assert_equal(codes[codes.size() - 1], probe_code)
	_hnoRevert.free()

func _on_Example_custom_signal(signal_name, value):
	custom_signal_emitted = [signal_name, value]

# Object-argument encoding tests (openspec: fix-object-arg-ptrcall-encoding).
# Each Go method returns 1 on success, a negative code on a failed assertion, or
# -900 if the Go side panicked. Nodes are freed here so the ObjectDB leak check
# in `make test` stays clean; freeing a parent frees its children.
func test_object_args(example: Example):
	print("test object args")

	# Object parameters and returns are declared with static types throughout this
	# function. That is itself part of the assertion: GDScript validates every call
	# site against the class the binder advertises, so a parameter still
	# advertising the owning class (Example) would fail here at parse time rather
	# than at runtime.

	# AddChild: the call shape that used to segfault.
	var child: Node = Node.new()
	assert_equal(example.test_object_arg_add_child(child), 1)
	assert_equal(child.get_parent(), example)

	# Identity carried through two object arguments.
	var parent: Node = Node.new()
	var kid: Node = Node.new()
	assert_equal(example.test_object_arg_identity(parent, kid), 1)
	assert_equal(kid.get_parent(), parent)
	assert_equal(parent.get_child_count(), 1)

	# Ref<Shape2D> argument: the engine must hold the very object we passed.
	var cs: CollisionShape2D = CollisionShape2D.new()
	var circle: CircleShape2D = CircleShape2D.new()
	assert_equal(example.test_object_arg_set_shape(cs, circle), 1)
	assert_equal(cs.shape, circle)

	# A typed-nil Ref clears it.
	assert_equal(example.test_object_arg_set_shape_typed_nil(cs), 1)
	assert_equal(cs.shape, null)

	# A live Ref holding no object clears it too.
	assert_equal(example.test_object_arg_set_shape_invalid_ref(cs), 1)
	assert_equal(cs.shape, null)

	# Nil non-refcounted (plain engine class) argument.
	var owned: Node = Node.new()
	assert_equal(example.test_object_arg_nil_plain_object(owned), 1)
	assert_equal(owned.get_owner(), null)

	# A base-class parameter accepts a typed subclass instance, which only parses
	# if the parameter advertises Shape2D rather than the owning class.
	var cs2: CollisionShape2D = CollisionShape2D.new()
	var circle2: CircleShape2D = CircleShape2D.new()
	assert_equal(example.test_object_arg_base_class_param(circle2), 1)

	# A typed return assignment: the return PropertyInfo must advertise Node.
	var returned: Node = example.test_object_arg_return_node(child)
	assert_equal(returned, child)

	# Repeated object arguments must not drift the reference count.
	var cs3: CollisionShape2D = CollisionShape2D.new()
	var shape3: CircleShape2D = CircleShape2D.new()
	assert_equal(example.test_object_arg_refcount_stability(cs3, shape3), 1)

	child.free()
	parent.free()
	cs.free()
	cs2.free()
	cs3.free()
	owned.free()
	circle = null
	circle2 = null
	shape3 = null

	# Let Go run the finalizers on Ref values decoded from the calls above, so no
	# Go-held reference survives into the engine's leak check.
	assert_equal(example.test_object_arg_release(), 1)

	# A Go-side get_shape() loop is exercised by test_return_ownership below.
	# It was withheld here while the object-return path leaked one engine reference
	# per call; that is now fixed, so the loop runs and the exit leak check gates it.


func test_return_ownership(example: Example):
	print("test return ownership")

	# Every generated method returning a refcounted object receives a reference
	# the engine already counted and transferred into the Go return slot. The
	# wrapper must own it and release it exactly once.
	var cs = CollisionShape2D.new()
	var circle = CircleShape2D.new()
	cs.shape = circle

	assert_equal(example.test_return_refcount_stability(cs, 200), 1)
	assert_equal(example.test_return_dropped_releases(cs), 1)
	assert_equal(example.test_return_unref_idempotence(cs), 1)
	assert_equal(example.test_borrowed_ref_never_releases(cs), 1)
	assert_equal(example.test_null_return_schedules_no_release(cs), 1)

	cs.free()
	circle = null

	# Collect while the engine is still running so no Go-held reference survives
	# into the exit leak check and no finalizer races engine teardown.
	assert_equal(example.test_return_ownership_release(), 1)


func test_pin_scratch_stays_flat(example: Example):
	print("test pin scratch stays flat")

	# Every generated call pins a return slot, an owner cell, an argument slice
	# and its argument cells. Those pins used to be taken on a package-global
	# pinner that was never released, so live Go heap climbed with the number of
	# calls and never came back down. Driving thousands of real generated calls
	# through the engine and sampling heap after a forced GC must land in a
	# constant band regardless of how many calls ran.
	#
	# Value-returning calls only, so nothing here touches refcount ownership.
	assert_equal(example.test_pin_scratch_stays_flat(2000, 10000), 1)


# Ptrcall object-argument decoding (openspec: fix-ptrcall-object-arg-decode).
# The objects reach Go through the varcall path, which works today, and each Go
# test then rebuilds the ptrcall argument cell by hand. That keeps this
# independent of fix-object-argument-type-metadata, without which no typed
# GDScript call could reach ptrcall at all.
func test_ptrcall_object_decode(example: Example):
	print("test ptrcall object decode")

	# Untyped on purpose: a typed call site would be rejected at parse by
	# the metadata defect that fix-object-argument-type-metadata owns. These
	# tests exercise the Go decoder, not GDScript static typing.
	var node = Node.new()
	var cs = CollisionShape2D.new()
	var circle = CircleShape2D.new()
	cs.shape = circle

	assert_equal(example.test_ptrcall_decode_plain_object(node), 1)
	assert_equal(example.test_ptrcall_decode_subclass(circle), 1)
	assert_equal(example.test_ptrcall_decode_null_object(), 1)
	assert_equal(example.test_ptrcall_decode_ref_regression(circle), 1)
	assert_equal(example.test_ptrcall_decode_undecodable_interface(node), 1)

	node.free()
	cs.free()
	circle = null

	assert_equal(example.test_ptrcall_decode_release(), 1)


# User-defined extension class arguments (openspec:
# fix-user-defined-class-object-arg-decode).
# Passing a TestHierarchicalDerived into a Go method declaring a plain Node
# used to segfault: the binding slot held a cgo.Handle value for user-defined
# classes while the reader assumed the address of a Go interface.
#
# Untyped on purpose. These must travel the varcall path, which is the path
# under test; a typed declaration would route the call to ptrcall instead.
func test_user_defined_object_args(example: Example):
	print("test user-defined object args")

	var derived = TestHierarchicalDerived.new()
	var plain = Node.new()

	assert_equal(example.test_user_defined_node_arg(derived, derived.get_instance_id()), 1)
	assert_equal(example.test_user_defined_arg_stable(derived), 1)
	assert_equal(example.test_engine_class_arg_still_resolves(plain, plain.get_instance_id()), 1)
	assert_equal(example.test_unresolvable_binding_is_typed_error(), 1)

	derived.free()
	plain.free()


# Varcall argument rejection (openspec: fix-varcall-arg-error-reporting).
#
# A wrong-class object used to abort the whole process from inside the Go
# decoder, off the cgo boundary. These assert the call is rejected instead: the
# method body must not run, and the run must survive far enough to print its
# own summary.
#
# Dynamic example.call() on purpose. A statically typed call site would be
# settled by GDScript's analyzer and never reach the varcall path under test.
func test_varcall_arg_rejection(example: Example):
	print("test varcall arg rejection")

	var label = Label.new()
	var sprite = Sprite2D.new()
	var arr = [1, 2, 3]

	# The decoder-level contract, driven through the Go seam.
	assert_equal(example.call("test_varcall_decode_wrong_class", label), 1)
	assert_equal(example.call("test_varcall_decode_subclass_still_accepted", label), 1)
	assert_equal(example.call("test_varcall_decode_owned_prefix_released", arr, label), 1)
	assert_equal(example.call("test_varcall_decode_success_releases_owned_container", arr), 1)

	# Control: the declared class is accepted and the probe body really runs.
	assert_equal(example.call("test_varcall_reject_probe", sprite), 1)
	var before = example.test_varcall_reject_probe_count()

	# This call MUST go through expect_rejected_call. Written inline, the engine's
	# rejection makes GDScript abandon this function, so everything below it --
	# including the frees -- silently stops executing while the suite still
	# reports green. The primitive contains the abandonment here so the
	# assertions and cleanup below actually run.
	expect_rejected_call(example, "test_varcall_reject_probe", [label])
	assert_true(last_call_was_rejected)
	assert_equal(example.test_varcall_reject_probe_count(), before)

	label.free()
	sprite.free()
