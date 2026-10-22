extends "res://test_base.gd"

var custom_signal_emitted = null
var _pending_await_target = null

class TestClass:
	func test(p_msg: String) -> String:
		return p_msg + " world"

class AwaitTarget:
	signal hit(penalty: int, kind: String)

class AwaitMarshalTarget:
	signal marshalled(a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q)

class AwaitOneShotTarget:
	signal ping(v: int)

class AwaitFreedTarget:
	signal never()

func _ready():
	var example: Example = $Example
	test_suite(1, example)
	test_object_args(example)
	test_pin_scratch_stays_flat(example)
	test_return_ownership(example)
	test_ptrcall_object_decode(example)
	test_user_defined_object_args(example)
	test_varcall_arg_rejection(example)
	test_property_hints(example)
	# example.group_subgroup_custom_position = Vector2(0, 0)
	# custom_signal_emitted = null
	# var t = get_tree()
	# if t != null:
	# 	await t.create_timer(3.0).timeout
	# test_suite(2, example)
	test_property_hint_metadata_late(example)
	test_property_list_persistence(example)
	await test_main_loop_callbacks(example)
	assert_equal(example.test_singleton_helpers(), 0)
	await test_string_boundary_marshalling(example)
	await test_native_slice_marshalling(example)
	test_variant_matrix_audit(example)
	await test_ref_method_boundary(example)
	await test_go_callable_bridge(example)
	await test_signal_await_delivery(example)
	await test_signal_await_marshal(example)
	await test_signal_await_oneshot(example)
	await test_signal_await_misuse(example)
	await test_signal_await_freed(example)
	await test_signal_await_timer(example)
	await test_signal_await_shutdown_pending(example)
	test_static_methods(example)
	test_editor_plugin_runtime_guard(example)
	test_instance_churn(example)
	test_binding_release_accounting(example)
	test_property_list_churn(example)
	test_vararg_return_buffer(example)
	test_binding_substrate(example)
	test_stacktrace_probe(example)
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

# Property hint metadata tests (openspec: support-property-hints). Registers
# hinted properties on Example in Go and reads them back through the engine.
func test_property_hints(example: Example):
	var props := {}
	for prop_info in example.get_property_list():
		props[prop_info['name']] = prop_info

	# Range hint reports unchanged.
	assert_true(props.has("hinted_range"))
	assert_equal(props["hinted_range"]['type'], TYPE_FLOAT)
	assert_equal(props["hinted_range"]['hint'], PROPERTY_HINT_RANGE)
	assert_equal(props["hinted_range"]['hint_string'], "0:100:1")

	# Enum hint reports unchanged.
	assert_true(props.has("hinted_enum"))
	assert_equal(props["hinted_enum"]['hint'], PROPERTY_HINT_ENUM)
	assert_equal(props["hinted_enum"]['hint_string'], "idle,walk,run")

	# Resource-type hint binds class_name from the hint string.
	assert_true(props.has("hinted_texture"))
	assert_equal(props["hinted_texture"]['hint'], PROPERTY_HINT_RESOURCE_TYPE)
	assert_equal(props["hinted_texture"]['hint_string'], "Texture2D")
	assert_equal(props["hinted_texture"]['class_name'], &"Texture2D")

	# Explicit qualifier applies when no resource-type hint.
	assert_true(props.has("hinted_node"))
	assert_equal(props["hinted_node"]['hint'], PROPERTY_HINT_NONE)
	assert_equal(props["hinted_node"]['class_name'], &"Node")

	# Storage-only usage reports unchanged and is not editor-visible.
	assert_true(props.has("storage_only"))
	assert_equal(props["storage_only"]['usage'], PROPERTY_USAGE_STORAGE)
	assert_equal(props["storage_only"]['usage'] & PROPERTY_USAGE_EDITOR, 0)

	# The legacy hint-free fixture still reports defaults.
	assert_true(props.has("group_subgroup_custom_position"))
	assert_equal(props["group_subgroup_custom_position"]['hint'], PROPERTY_HINT_NONE)
	assert_equal(props["group_subgroup_custom_position"]['hint_string'], "")
	assert_equal(props["group_subgroup_custom_position"]['usage'], PROPERTY_USAGE_DEFAULT)

	# Read-only property (no setter name) registers successfully.
	assert_true(props.has("readonly_value"))

	# Read/write through the hinted properties works.
	example.hinted_range = 42.5
	assert_equal(example.hinted_range, 42.5)
	example.hinted_enum = 2
	assert_equal(example.hinted_enum, 2)

# Late re-read of hint metadata after all other suites have run: the
# registration-time StringName/String temporaries were released right after
# ClassDBAddPropertyInfo, so corrupted or dangling metadata text here would
# surface in this final read.
func test_property_hint_metadata_late(example: Example):
	var seen_range := false
	var seen_texture := false
	for prop_info in example.get_property_list():
		if prop_info['name'] == "hinted_range":
			seen_range = true
			assert_equal(prop_info['hint_string'], "0:100:1")
		elif prop_info['name'] == "hinted_texture":
			seen_texture = true
			assert_equal(prop_info['hint_string'], "Texture2D")
			assert_equal(prop_info['class_name'], &"Texture2D")
	assert_true(seen_range)
	assert_true(seen_texture)

# Double-read gate for the list-declared custom properties (openspec
# fix-class-property-list-lifecycle): the get_property_list callback used to
# publish one Go-side array whose element strings were destroyed by the first
# free_property_list2 call, wiping the list-declared properties out of every
# later query. property_from_list and dproperty_* are reachable only through
# the class callback, so the vanishing-list bug passed while every
# add_property fixture above stayed green.
func test_property_list_persistence(example: Example):
	var first := {}
	for prop_info in example.get_property_list():
		first[prop_info['name']] = prop_info
	var second := {}
	for prop_info in example.get_property_list():
		second[prop_info['name']] = prop_info
	for prop_name in ["property_from_list", "dproperty_0", "dproperty_1", "dproperty_2", "property_hint_string"]:
		assert_true(first.has(prop_name))
		assert_true(second.has(prop_name))
		if first.has(prop_name) and second.has(prop_name):
			assert_equal(second[prop_name]['type'], first[prop_name]['type'])
			assert_equal(second[prop_name]['hint'], first[prop_name]['hint'])
			assert_equal(second[prop_name]['hint_string'], first[prop_name]['hint_string'])
			assert_equal(second[prop_name]['usage'], first[prop_name]['usage'])
	if first.has("property_from_list"):
		assert_equal(first["property_from_list"]['type'], TYPE_VECTOR3)
		assert_equal(first["property_from_list"]['usage'], PROPERTY_USAGE_DEFAULT)
		assert_equal(first["property_from_list"]['class_name'], &"Example")
	if first.has("property_hint_string"):
		# Positive content check: a non-empty hint string must round-trip
		# through the class callback in both reads.
		assert_equal(first["property_hint_string"]['hint_string'], "review-carrier")
		assert_equal(second["property_hint_string"]['hint_string'], "review-carrier")

# Main-loop callback tests (openspec: support-main-loop-callbacks).
# The startup hook fires during engine startup (Main::start calls
# GDExtensionManager::startup() before the first iteration), so its evidence
# is already set while this scene's _ready runs.
func test_main_loop_callbacks(example: Example):
	assert_true(example.main_loop_startup_ran())

	# The frame hook fires at the end of Main::iteration, after the
	# process_frame dispatch that resumes this coroutine. The priming await
	# absorbs that ordering boundary; every measurement await afterwards must
	# observe the previous frame's hook (per-await delta >= 1). The run uses
	# --quit-after 12 so several iterations complete before exit even with the
	# frame waits in the string- and slice-retention fixtures.
	await get_tree().process_frame
	var last := example.main_loop_frame_count()
	for i in range(2):
		await get_tree().process_frame
		var now := example.main_loop_frame_count()
		print("main-loop frame count: %d -> %d" % [last, now])
		assert_true(now >= last + 1)
		last = now

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


# string-boundary-marshalling fixtures: mixed variant types arriving at a
# native `string` parameter, Copy() content preservation, and retention of a
# borrowed StringName past its call.
func test_string_boundary_marshalling(example: Example) -> void:
	print("test string boundary marshalling")

	# A StringName's text arrives at a string parameter, direct and generic.
	assert_equal(example.test_string_param_text_conversion("boundary_name", &"boundary_name"), 0)
	assert_equal(example.call("test_string_param_text_conversion", "boundary_name", &"boundary_name"), 0)

	# A NodePath's text arrives at a string parameter. Non-implicit argument
	# conversions only exist on the generic call path; the static analyzer
	# rejects them at direct call sites.
	assert_equal(example.call("test_string_param_text_conversion", "boundary/path", NodePath("boundary/path")), 0)

	# A variant without a String conversion is rejected at the call layer
	# before the Go method runs. Contained via expect_rejected_call so the
	# engine rejection cannot abandon the rest of this function.
	expect_rejected_call(example, "test_string_param_text_conversion", ["42", 42])
	assert_true(last_call_was_rejected)

	# Copy() preserves content for all three engine string types.
	assert_equal(example.test_copy_content_preservation(), 0)

	# A borrowed StringName retained via Copy() survives into later frames.
	assert_equal(example.test_string_copy_store(&"retained_name"), 0)
	await get_tree().process_frame
	await get_tree().process_frame
	assert_equal(example.test_string_copy_verify_release("retained_name"), 0)


# support-native-slice-marshalling fixtures: Go native slice parameters decode
# from every Packed*K kind. These probes assert the decode side; the echo pass
# that also asserts the return side accompanies the return encoder.
func test_native_slice_marshalling(example: Example) -> void:
	print("test native slice marshalling")

	var bytes = PackedByteArray()
	bytes.resize(1000)
	var bytes_sum = 0
	for i in 1000:
		bytes[i] = i % 256
		bytes_sum += i % 256

	var i32 = PackedInt32Array([1, -2, 300000, -2147483648, 2147483647])
	var i64 = PackedInt64Array([1, -2, 1099511627776, -1099511627769])
	var f32 = PackedFloat32Array([1.5, -0.25, 448.0])
	var f64 = PackedFloat64Array([1.5, -0.25, 448.0])
	var strs = PackedStringArray(["a", "héllo", "日本", ""])
	var strs_utf8_len = 0
	for s in strs:
		strs_utf8_len += s.to_utf8_buffer().size()
	var v2 = PackedVector2Array([Vector2(1, 2), Vector2(-3.5, 4)])
	var v3 = PackedVector3Array([Vector3(1, 2, 3), Vector3(-1, -2, -3)])
	var v4 = PackedVector4Array([Vector4(1, 2, 3, 4), Vector4(-1, -2, -3, -4)])
	var colors = PackedColorArray([Color(0.25, 0.5, 0.75, 1), Color(0, 0, 0, 0)])

	# Varcall decode: lengths for all ten packed kinds.
	assert_equal(example.call("test_slice_len_bytes", bytes), 1000)
	assert_equal(example.call("test_slice_len_int32", i32), 5)
	assert_equal(example.call("test_slice_len_int64", i64), 4)
	assert_equal(example.call("test_slice_len_float32", f32), 3)
	assert_equal(example.call("test_slice_len_float64", f64), 3)
	assert_equal(example.call("test_slice_len_strings", strs), 4)
	assert_equal(example.call("test_slice_len_vector2", v2), 2)
	assert_equal(example.call("test_slice_len_vector3", v3), 2)
	assert_equal(example.call("test_slice_len_vector4", v4), 2)
	assert_equal(example.call("test_slice_len_color", colors), 2)

	# Empty arrays decode to zero-length slices (D6).
	assert_equal(example.call("test_slice_len_bytes", PackedByteArray()), 0)
	assert_equal(example.call("test_slice_len_strings", PackedStringArray()), 0)

	# Content: checksums catch corruption the length alone would hide.
	assert_equal(example.call("test_slice_checksum_bytes", bytes), bytes_sum)
	assert_equal(example.call("test_slice_checksum_int64", i64), 6)
	assert_equal(example.call("test_slice_checksum_strings", strs), strs_utf8_len)
	assert_equal(example.call("test_slice_checksum_float64", f64), 449.25)

	# A variant with no packed conversion is rejected at the call layer
	# before the Go method runs. Contained via expect_rejected_call so the
	# engine rejection cannot abandon the rest of this function.
	expect_rejected_call(example, "test_slice_len_bytes", [42])
	assert_true(last_call_was_rejected)

	# The strict conversion gate treats Nil as non-convertible to a packed
	# parameter too: null is rejected at the call layer, it does not reach
	# the Go decoder (engine evidence: "Cannot convert argument 2 from Nil
	# to PackedByteArray"). Contained, and the surviving calls below prove
	# the run continued past both rejections.
	expect_rejected_call(example, "test_slice_len_bytes", [null])
	assert_true(last_call_was_rejected)
	assert_equal(example.test_slice_checksum_bytes(bytes), bytes_sum)

	# Direct typed calls take the ptrcall path with the engine's borrowed
	# cell: same fixtures must produce the same results. A direct call with
	# a bad type is settled by the GDScript analyzer at parse time, so only
	# the varcall path carries the contained-rejection pins above.
	assert_equal(example.test_slice_len_bytes(bytes), 1000)
	assert_equal(example.test_slice_len_int32(i32), 5)
	assert_equal(example.test_slice_len_int64(i64), 4)
	assert_equal(example.test_slice_len_float32(f32), 3)
	assert_equal(example.test_slice_len_float64(f64), 3)
	assert_equal(example.test_slice_len_strings(strs), 4)
	assert_equal(example.test_slice_len_vector2(v2), 2)
	assert_equal(example.test_slice_len_vector3(v3), 2)
	assert_equal(example.test_slice_len_vector4(v4), 2)
	assert_equal(example.test_slice_len_color(colors), 2)
	assert_equal(example.test_slice_len_bytes(PackedByteArray()), 0)
	assert_equal(example.test_slice_checksum_int64(i64), 6)
	assert_equal(example.test_slice_checksum_strings(strs), strs_utf8_len)

	# Return encoder: echoes must come back element-equal on both paths (GDScript
	# == compares packed arrays element-for-element), proving the built cells
	# carry the data across the boundary.
	assert_equal(example.call("test_slice_echo_bytes", bytes), bytes)
	assert_equal(example.call("test_slice_echo_int32", i32), i32)
	assert_equal(example.call("test_slice_echo_int64", i64), i64)
	assert_equal(example.call("test_slice_echo_float32", f32), f32)
	assert_equal(example.call("test_slice_echo_float64", f64), f64)
	assert_equal(example.call("test_slice_echo_strings", strs), strs)
	assert_equal(example.call("test_slice_echo_vector2", v2), v2)
	assert_equal(example.call("test_slice_echo_vector3", v3), v3)
	assert_equal(example.call("test_slice_echo_vector4", v4), v4)
	assert_equal(example.call("test_slice_echo_color", colors), colors)
	assert_equal(example.test_slice_echo_bytes(bytes), bytes)
	assert_equal(example.test_slice_echo_int32(i32), i32)
	assert_equal(example.test_slice_echo_int64(i64), i64)
	assert_equal(example.test_slice_echo_float32(f32), f32)
	assert_equal(example.test_slice_echo_float64(f64), f64)
	assert_equal(example.test_slice_echo_strings(strs), strs)
	assert_equal(example.test_slice_echo_vector2(v2), v2)
	assert_equal(example.test_slice_echo_vector3(v3), v3)
	assert_equal(example.test_slice_echo_vector4(v4), v4)
	assert_equal(example.test_slice_echo_color(colors), colors)

	# A nil Go slice returns as an empty container, never null (design D6).
	var nil_echo = example.test_slice_echo_nil_bytes()
	assert_true(nil_echo != null)
	assert_true(nil_echo.is_empty())

	# Fidelity (spec requirement 4): exact text, 10,000-element payloads,
	# int64 extremes, float edges, and per-component equality on both call
	# paths. Component comparisons use values exactly representable in the
	# engine's float32 storage (Vector2/Vector3/Vector4/Color components).
	var utf8 = PackedStringArray(["héllo", "日本", "😀"])
	var utf8_direct = example.test_slice_echo_strings(utf8)
	var utf8_var = example.call("test_slice_echo_strings", utf8)
	assert_true(utf8_direct[0] == "héllo")
	assert_true(utf8_direct[1] == "日本")
	assert_true(utf8_direct[2] == "😀")
	assert_true(utf8_var[0] == "héllo")
	assert_true(utf8_var[2] == "😀")

	var big = PackedByteArray()
	big.resize(10000)
	for i in 10000:
		big[i] = (i * 37 + 11) % 256
	assert_equal(example.test_slice_echo_bytes(big), big)
	assert_equal(example.call("test_slice_echo_bytes", big), big)

	var extremes = PackedInt64Array([
		-9223372036854775807 - 1, -4611686018427387903, -1, 0,
		4611686018427387903, 9223372036854775807
	])
	assert_equal(example.test_slice_echo_int64(extremes), extremes)
	assert_equal(example.call("test_slice_echo_int64", extremes), extremes)

	var f32_edge = PackedFloat32Array([0.1, -0.0, 1e-30, 3.4028234663852886e38])
	assert_equal(example.test_slice_echo_float32(f32_edge), f32_edge)
	assert_equal(example.call("test_slice_echo_float32", f32_edge), f32_edge)
	var f64_edge = PackedFloat64Array([0.1, -0.0, 1e-300, 1.7976931348623157e308])
	assert_equal(example.test_slice_echo_float64(f64_edge), f64_edge)
	assert_equal(example.call("test_slice_echo_float64", f64_edge), f64_edge)

	var v2e = PackedVector2Array([Vector2(1.5, -2.25), Vector2(-0.5, 1024.5)])
	var e2d = example.test_slice_echo_vector2(v2e)
	var e2v = example.call("test_slice_echo_vector2", v2e)
	assert_true(e2d[0].x == 1.5 and e2d[0].y == -2.25)
	assert_true(e2v[1].x == -0.5 and e2v[1].y == 1024.5)

	var v3e = PackedVector3Array([Vector3(-1.5, 2.25, -3.125)])
	var e3d = example.test_slice_echo_vector3(v3e)
	assert_true(e3d[0].x == -1.5 and e3d[0].y == 2.25 and e3d[0].z == -3.125)

	var v4e = PackedVector4Array([Vector4(-1.5, 2.25, 3.125, 1024.5)])
	var e4d = example.test_slice_echo_vector4(v4e)
	var e4v = example.call("test_slice_echo_vector4", v4e)
	assert_true(e4d[0].x == -1.5 and e4d[0].w == 1024.5)
	assert_true(e4v[0].y == 2.25 and e4v[0].z == 3.125)

	var cole = PackedColorArray([Color(0.25, 0.5, 0.75, 1), Color(0, 0, 0, 0.125)])
	var ecd = example.test_slice_echo_color(cole)
	var ecv = example.call("test_slice_echo_color", cole)
	assert_true(ecd[0].r == 0.25 and ecd[0].g == 0.5 and ecd[0].b == 0.75 and ecd[0].a == 1)
	assert_true(ecv[1].r == 0 and ecv[1].a == 0.125)

	# Independence (spec requirement 5): Go mutates its decoded copy and the
	# GDScript original must be untouched on both paths (the ptrcall path is
	# the dangerous one: it decodes from the engine's borrowed cell).
	var indep = PackedByteArray([10, 20, 30, 250])
	assert_equal(example.test_slice_mutate_probe(indep), 310)
	assert_true(indep[0] == 10 and indep[1] == 20 and indep[2] == 30 and indep[3] == 250)
	assert_equal(example.call("test_slice_mutate_probe", indep), 310)
	assert_true(indep[0] == 10 and indep[1] == 20 and indep[2] == 30 and indep[3] == 250)

	# Retention: a decoded slice stored on the Go side stays intact across
	# frames although the engine array it came from is long gone, with no
	# cgocheck violation and no leaked engine objects (D1/D5a).
	assert_equal(example.test_slice_retain_store(bytes), 0)
	await get_tree().process_frame
	await get_tree().process_frame
	assert_equal(example.test_slice_retain_verify(1000, bytes_sum), 0)

# Ref argument/return crossing (openspec: ref-method-boundary).
# Direct calls take the ptrcall path and example.call() pins varcall; both
# decode through one shared path. Mismatches go through expect_rejected_call
# because only varcall contains them: a mismatched direct call is a ptrcall
# panic by design (no CallError slot in the ABI), so the cells proving that
# behavior are driven inside Go, not from here.
func test_ref_method_boundary(example: Example) -> void:
	print("test ref method boundary")

	var shape: Shape2D = RectangleShape2D.new()
	var res: Resource = shape

	# Valid Ref decodes on both paths.
	assert_equal(example.test_ref_boundary_valid(res), 1)
	assert_equal(example.call("test_ref_boundary_valid", res), 1)

	# The echoed Ref identifies the same object through the return encoder.
	assert_equal(example.test_ref_boundary_echo(res).get_instance_id(), res.get_instance_id())
	assert_equal(example.call("test_ref_boundary_echo", res).get_instance_id(), res.get_instance_id())

	# A null argument is the zero Ref, not a rejected call: the body runs and
	# reports 0.
	var nilres: Resource = null
	assert_equal(example.test_ref_boundary_valid(nilres), 0)
	assert_equal(example.call("test_ref_boundary_valid", null), 0)
	assert_true(example.test_ref_boundary_echo(nilres) == null)

	# An invalid Ref return crosses back as null.
	assert_true(example.test_ref_boundary_null_return() == null)

	# Bare builtin.Ref accepts any reference-counted object on both paths.
	assert_equal(example.test_bare_ref_valid(res), 1)
	assert_equal(example.call("test_bare_ref_valid", res), 1)

	# A user-defined extension RefCounted decodes to its live Go wrapper on
	# both paths: the Id was set through the binding and lives only there.
	var user_ref = ExampleRef.new()
	user_ref.set_id(777)
	assert_equal(example.test_ref_boundary_user_ref_id(user_ref), 777)
	assert_equal(example.call("test_ref_boundary_user_ref_id", user_ref), 777)

	# The bare Ref return form crosses back with the same identity.
	assert_equal(example.test_bare_ref_echo(user_ref).get_instance_id(), user_ref.get_instance_id())
	assert_equal(example.call("test_bare_ref_echo", user_ref).get_instance_id(), user_ref.get_instance_id())

	# A subclass argument slices to the declared pointee while the object
	# keeps its runtime identity.
	assert_equal(example.test_ref_boundary_subclass_sliced(shape), 1)

	# The decoded Ref holds a reference of its own (acquire, godot-cpp parity).
	assert_equal(example.test_ref_boundary_refcount(res), 1)

	# Wrong-class arguments are caller errors, contained through varcall, and
	# the body must not run: a non-reference-counted object and a
	# reference-counted object outside the declared pointee's branch.
	var label = Label.new()
	var probe_before = example.test_ref_boundary_reject_probe_count()
	expect_rejected_call(example, "test_ref_boundary_reject_probe", [label])
	assert_true(last_call_was_rejected)
	var not_a_resource = RefCounted.new()
	expect_rejected_call(example, "test_ref_boundary_reject_probe", [not_a_resource])
	assert_true(last_call_was_rejected)
	expect_rejected_call(example, "test_bare_ref_valid", [label])
	assert_true(last_call_was_rejected)
	assert_equal(example.test_ref_boundary_reject_probe_count(), probe_before)

	# ptrcall cells driven in Go: the fast path refuses wrong classes loudly,
	# and a refusal there must never reach a live call site.
	assert_equal(example.test_ref_boundary_ptrcall_cell(res), 1)
	assert_equal(example.test_ref_boundary_ptrcall_bare_ref_cell(shape), 1)
	assert_equal(example.test_ref_boundary_ptrcall_nil_cell(), 1)
	assert_equal(example.test_ref_boundary_ptrcall_mismatch_cell(label), 1)
	label.free()

	# Retention across a frame boundary: the script drops its references and
	# only the acquired one can keep the stashed object alive and readable.
	assert_equal(example.test_ref_boundary_stash(res), 1)
	res = null
	shape = null
	await get_tree().process_frame
	assert_equal(example.test_ref_boundary_stash_alive(), 1)

	# Release every reference this suite acquired before the engine takes its
	# exit leak snapshot.
	assert_equal(example.test_ref_boundary_release(), 1)

# 36-type variant matrix audit (openspec: support-variant-reflect-completeness).
# Every TestAudit* echo is driven twice: a direct call (ptrcall fast path) and
# example.call() (varcall). Together with the metadata the binder derives from
# each signature, that pins mapper acceptance, both decoders, and both return
# encoders per variant type. The ten packed rows reuse the slice echoes bound
# by support-native-slice-marshalling; NIL rides the Variant echo as null.
func test_variant_matrix_audit(example: Example) -> void:
	print("test variant matrix audit")

	assert_true(example.test_audit_bool(true))
	assert_true(example.call("test_audit_bool", false) == false)
	assert_equal(example.test_audit_int(-42), -42)
	assert_equal(example.call("test_audit_int", 4611686018427387903), 4611686018427387903)
	assert_equal(example.test_audit_float(3.25), 3.25)
	assert_equal(example.call("test_audit_float", -0.5), -0.5)
	assert_equal(example.test_audit_string("audit"), "audit")
	assert_equal(example.call("test_audit_string", "héllo 世界 😀"), "héllo 世界 😀")
	assert_equal(example.test_audit_vector2(Vector2(1.5, -2.25)), Vector2(1.5, -2.25))
	assert_equal(example.call("test_audit_vector2", Vector2(-3, 4)), Vector2(-3, 4))
	assert_equal(example.test_audit_vector2i(Vector2i(3, -4)), Vector2i(3, -4))
	assert_equal(example.call("test_audit_vector2i", Vector2i.MIN), Vector2i.MIN)
	assert_equal(example.test_audit_rect2(Rect2(1.5, 2.5, 3.0, -4.0)), Rect2(1.5, 2.5, 3.0, -4.0))
	assert_equal(example.call("test_audit_rect2", Rect2()), Rect2())
	assert_equal(example.test_audit_rect2i(Rect2i(1, -2, 3, 4)), Rect2i(1, -2, 3, 4))
	assert_equal(example.call("test_audit_rect2i", Rect2i()), Rect2i())
	assert_equal(example.test_audit_vector3(Vector3(1, -2, 3.5)), Vector3(1, -2, 3.5))
	assert_equal(example.call("test_audit_vector3", Vector3.ZERO), Vector3.ZERO)
	assert_equal(example.test_audit_vector3i(Vector3i(1, -2, 3)), Vector3i(1, -2, 3))
	assert_equal(example.call("test_audit_vector3i", Vector3i.MAX), Vector3i.MAX)
	assert_equal(example.test_audit_transform2d(Transform2D(0.5, Vector2(1, 2))),
		Transform2D(0.5, Vector2(1, 2)))
	assert_equal(example.call("test_audit_transform2d", Transform2D()), Transform2D())
	assert_equal(example.test_audit_vector4(Vector4(1, -2, 3.5, 0)), Vector4(1, -2, 3.5, 0))
	assert_equal(example.call("test_audit_vector4", Vector4.ZERO), Vector4.ZERO)
	assert_equal(example.test_audit_vector4i(Vector4i(1, -2, 3, -4)), Vector4i(1, -2, 3, -4))
	assert_equal(example.call("test_audit_vector4i", Vector4i.ZERO), Vector4i.ZERO)
	assert_equal(example.test_audit_plane(Plane(Vector3(0, 1, 0), 2.5)), Plane(Vector3(0, 1, 0), 2.5))
	assert_equal(example.call("test_audit_plane", Plane()), Plane())
	assert_equal(example.test_audit_quaternion(Quaternion(0.5, 0.5, 0.5, 0.5)),
		Quaternion(0.5, 0.5, 0.5, 0.5))
	assert_equal(example.call("test_audit_quaternion", Quaternion.IDENTITY), Quaternion.IDENTITY)
	assert_equal(example.test_audit_aabb(AABB(Vector3(1, 2, 3), Vector3(4, 5, 6))),
		AABB(Vector3(1, 2, 3), Vector3(4, 5, 6)))
	assert_equal(example.call("test_audit_aabb", AABB()), AABB())
	assert_equal(example.test_audit_basis(Basis(Vector3(2, 0, 0), Vector3(0, 3, 0), Vector3(0, 0, 4))),
		Basis(Vector3(2, 0, 0), Vector3(0, 3, 0), Vector3(0, 0, 4)))
	assert_equal(example.call("test_audit_basis", Basis.IDENTITY), Basis.IDENTITY)
	assert_equal(example.test_audit_transform3d(
		Transform3D(Basis(Vector3(2, 0, 0), Vector3(0, 3, 0), Vector3(0, 0, 4)), Vector3(1, 2, 3))),
		Transform3D(Basis(Vector3(2, 0, 0), Vector3(0, 3, 0), Vector3(0, 0, 4)), Vector3(1, 2, 3)))
	assert_equal(example.call("test_audit_transform3d", Transform3D()), Transform3D())
	assert_equal(example.test_audit_projection(Projection()), Projection())
	assert_equal(example.call("test_audit_projection", Projection()), Projection())
	assert_equal(example.test_audit_color(Color(0.25, 0.5, 0.75, 1.0)), Color(0.25, 0.5, 0.75, 1.0))
	assert_equal(example.call("test_audit_color", Color.TRANSPARENT), Color.TRANSPARENT)
	assert_equal(example.test_audit_string_name(&"audit_name"), &"audit_name")
	assert_equal(example.call("test_audit_string_name", &""), &"")
	assert_equal(example.test_audit_node_path(^"Some/Path/Node"), ^"Some/Path/Node")
	assert_equal(example.call("test_audit_node_path", ^"."), ^".")
	assert_equal(example.test_audit_rid(RID()), RID())
	assert_equal(example.call("test_audit_rid", RID()), RID())

	# OBJECT: identity survives both paths; the Node lives until the free below.
	var probe_node = Label.new()
	assert_equal(example.test_audit_object(probe_node).get_instance_id(), probe_node.get_instance_id())
	assert_equal(example.call("test_audit_object", probe_node).get_instance_id(), probe_node.get_instance_id())
	probe_node.free()

	var probe_callable := Callable(example, "to_string")
	assert_true(example.test_audit_callable(probe_callable) == probe_callable)
	assert_true(example.call("test_audit_callable", probe_callable) == probe_callable)

	var probe_signal = example.tree_entered
	assert_true(example.test_audit_signal(probe_signal) == probe_signal)
	assert_true(example.call("test_audit_signal", probe_signal) == probe_signal)

	var probe_dictionary := {"a": 1, "b": [1, 2], 3: &"three"}
	assert_true(example.test_audit_dictionary(probe_dictionary) == probe_dictionary)
	assert_true(example.call("test_audit_dictionary", probe_dictionary) == probe_dictionary)

	var probe_array := [1, "two", Vector2(3, 4), null]
	assert_true(example.test_audit_array(probe_array) == probe_array)
	assert_true(example.call("test_audit_array", probe_array) == probe_array)

	# NIL row: the zero Variant crosses back as null on both paths.
	assert_true(example.test_audit_variant(null) == null)
	assert_true(example.call("test_audit_variant", null) == null)

	# VARIANT_MAX row: a typed value keeps type and content through the
	# Variant echo on both paths.
	var typed_variant = example.test_audit_variant(Vector2(1.5, 2.5))
	assert_true(typed_variant is Vector2)
	assert_equal(typed_variant, Vector2(1.5, 2.5))
	assert_equal(example.call("test_audit_variant", &"variant_name"), &"variant_name")

	# The ten packed rows reuse the native-slice echoes; content equality for
	# packed arrays compares elementwise, proving the deep copy.
	assert_equal(example.test_slice_echo_bytes(PackedByteArray([1, 2, 3])), PackedByteArray([1, 2, 3]))
	assert_equal(example.call("test_slice_echo_bytes", PackedByteArray([4, 5])), PackedByteArray([4, 5]))
	assert_equal(example.test_slice_echo_int32(PackedInt32Array([1, -2])), PackedInt32Array([1, -2]))
	assert_equal(example.call("test_slice_echo_int32", PackedInt32Array([3])), PackedInt32Array([3]))
	assert_equal(example.test_slice_echo_int64(PackedInt64Array([-9223372036854775807 - 1, 7])),
		PackedInt64Array([-9223372036854775807 - 1, 7]))
	assert_equal(example.call("test_slice_echo_int64", PackedInt64Array([8])), PackedInt64Array([8]))
	assert_equal(example.test_slice_echo_float32(PackedFloat32Array([1.5, -2.5])),
		PackedFloat32Array([1.5, -2.5]))
	assert_equal(example.call("test_slice_echo_float32", PackedFloat32Array([0.0])), PackedFloat32Array([0.0]))
	assert_equal(example.test_slice_echo_float64(PackedFloat64Array([1.5, -0.25])),
		PackedFloat64Array([1.5, -0.25]))
	assert_equal(example.call("test_slice_echo_float64", PackedFloat64Array([2.0])), PackedFloat64Array([2.0]))
	assert_equal(example.test_slice_echo_strings(PackedStringArray(["a", "héllo"])),
		PackedStringArray(["a", "héllo"]))
	assert_equal(example.call("test_slice_echo_strings", PackedStringArray(["b"])), PackedStringArray(["b"]))
	assert_equal(example.test_slice_echo_vector2(PackedVector2Array([Vector2(1, 2)])),
		PackedVector2Array([Vector2(1, 2)]))
	assert_equal(example.call("test_slice_echo_vector2", PackedVector2Array([Vector2(3, 4)])),
		PackedVector2Array([Vector2(3, 4)]))
	assert_equal(example.test_slice_echo_vector3(PackedVector3Array([Vector3(1, 2, 3)])),
		PackedVector3Array([Vector3(1, 2, 3)]))
	assert_equal(example.call("test_slice_echo_vector3", PackedVector3Array([Vector3(4, 5, 6)])),
		PackedVector3Array([Vector3(4, 5, 6)]))
	assert_equal(example.test_slice_echo_vector4(PackedVector4Array([Vector4(1, 2, 3, 4)])),
		PackedVector4Array([Vector4(1, 2, 3, 4)]))
	assert_equal(example.call("test_slice_echo_vector4", PackedVector4Array([Vector4(5, 6, 7, 8)])),
		PackedVector4Array([Vector4(5, 6, 7, 8)]))
	assert_equal(example.test_slice_echo_color(PackedColorArray([Color(0.5, 0.25, 0.125, 1.0)])),
		PackedColorArray([Color(0.5, 0.25, 0.125, 1.0)]))
	assert_equal(example.call("test_slice_echo_color", PackedColorArray([Color(1, 0, 0, 0.5)])),
		PackedColorArray([Color(1, 0, 0, 0.5)]))

# support-go-callable-bridge: Go closures as engine-native Callables.
func test_go_callable_bridge(example: Example) -> void:
	print("test go callable bridge")
	await _go_callable_scenarios(example)

func _go_callable_scenarios(example: Example) -> void:

	# Creation, validity, and repeated invocations of a captured Go closure.
	var counter_cb: Callable = example.create_counter_callable()
	assert_true(counter_cb.is_valid())
	assert_equal(counter_cb.call(), 1)
	assert_equal(counter_cb.call(), 2)
	assert_equal(counter_cb.call(), 3)
	# The varcall path (dynamic call) reaches the same machinery.
	assert_equal(example.call("create_counter_callable").call(), 1)

	# to_string: the custom label replaces the generic fallback rendering.
	assert_equal(str(counter_cb), "GoCallable")

	# Copy-if-store: the stored copy outlives the borrowed argument it came
	# from and reads back equal on a later invocation.
	var copy_cb: Callable = example.create_copy_stored_callable()
	assert_equal(copy_cb.call("stored string"), 1)
	assert_equal(copy_cb.call(), "stored string")
	assert_equal(copy_cb.call("replacement"), 1)
	assert_equal(copy_cb.call(), "replacement")

	# Identity and signal dedup: connecting the same Callable twice yields one
	# connection; distinct creations yield distinct connections that both run.
	var conns_before: int = example.custom_signal.get_connections().size()
	example.reset_go_callable_invocation_count()
	var probe_a: Callable = example.create_probe_callable()
	var probe_b: Callable = example.create_probe_callable()
	example.custom_signal.connect(probe_a)
	example.custom_signal.connect(probe_a)
	assert_equal(example.custom_signal.get_connections().size(), conns_before + 1)
	example.emit_custom_signal("dedup", 1)
	assert_equal(example.go_callable_invocation_count(), 1)
	example.custom_signal.connect(probe_b)
	assert_equal(example.custom_signal.get_connections().size(), conns_before + 2)
	example.emit_custom_signal("dedup", 2)
	assert_equal(example.go_callable_invocation_count(), 3)

	# Copies of a Callable compare equal to their source; separate creations
	# are distinct.
	var counter_copy: Callable = counter_cb
	assert_true(counter_copy == counter_cb)
	assert_false(probe_a == probe_b)

	# call_deferred runs the same body on a later frame. The two-frame wait
	# absorbs the flush boundary before measuring.
	var deferred_cb: Callable = example.create_probe_callable()
	deferred_cb.call_deferred()
	await get_tree().process_frame
	await get_tree().process_frame
	assert_equal(example.go_callable_invocation_count(), 4)
	example.custom_signal.disconnect(probe_a)
	example.custom_signal.disconnect(probe_b)

	# Closure errors and panics fail the call without taking the process down;
	# repeated calls on the same callable keep failing gracefully, and the
	# bindings stay healthy afterwards.
	assert_true(_try_go_callable_call(example.create_error_callable()) == null)
	assert_true(_try_go_callable_call(example.create_error_callable()) == null)
	assert_true(_try_go_callable_call(example.create_panic_callable()) == null)
	assert_true(_try_go_callable_call(example.create_panic_callable()) == null)
	var health_cb: Callable = example.create_counter_callable()
	assert_equal(health_cb.call(), 1)

	# Reentrancy: the closure calls Object.call inside its own invocation,
	# which varcalls back into a Go-bound method (engine -> Go -> engine -> Go).
	var reentrant_cb: Callable = example.create_reentrant_callable()
	assert_equal(reentrant_cb.call(), 42)

	# Deterministic release: entries appear with creation and disappear the
	# moment the last script reference drops, with no Go-side action.
	# Two GDScript call-site scratches pin one Callable each for the life of
	# the script: the dynamic-call probe above and the last loop iteration of
	# live.append. Those are engine-side references, not stale registry
	# entries. The loop therefore creates one extra sacrificial entry so the
	# post-clear count lands on a provable baseline + 1 instead of racing
	# interpreter scratch lifetimes.
	var baseline: int = example.go_callable_registry_count()
	var live: Array = []
	for i in range(11):
		live.append(example.create_counter_callable())
	assert_equal(example.go_callable_registry_count(), baseline + 11)
	live.clear()
	assert_equal(example.go_callable_registry_count(), baseline + 1)

	# Release word: the stored String copy dies here with the fixture. The
	# entries still registered are the named callables above, which drop with
	# this frame; engine-side teardown then leaves nothing for the leak gate.
	example.release_go_callable_stored_copy()

# A failing Callable invocation must degrade to null, not abort; isolating the
# call in a frame keeps a hypothetical interpreter-level abort contained.
func _try_go_callable_call(cb: Callable) -> Variant:
	return cb.call()

func test_signal_await_delivery(example: Example) -> void:
	print("test signal await delivery")
	var target := AwaitTarget.new()
	example.await_hit_start(target)
	target.emit_signal("hit", 4, "critical")
	# The probe joins the Go consumer directly, so no frames are burned.
	assert_equal(example.await_hit_code(), 1)

func test_signal_await_marshal(example: Example) -> void:
	print("test signal await marshal table")
	var target := AwaitMarshalTarget.new()
	example.await_marshal_start(target, example)
	target.emit_signal("marshalled", true, 7, 2.5, "text", &"sn_name",
		NodePath("Node/sub"), Vector2(1.5, -2.25), example,
		[1, [2, "deep"]], {"k": 1, "nest": {"m": true}},
		target.marshalled.emit, null,
		Vector3(1, 2, 3), Color(0.25, 0.5, 0.75, 1), Vector2i(7, -9),
		Transform2D(Vector2(1, 0), Vector2(0, 1), Vector2(10, 20)), Basis())
	assert_equal(example.await_marshal_code(), 1)

func test_signal_await_oneshot(example: Example) -> void:
	print("test signal await one-shot")
	var target := AwaitOneShotTarget.new()
	example.await_oneshot_start(target)
	target.emit_signal("ping", 1)
	# Delivered, and the engine freed the one-shot callable with the connection.
	assert_equal(example.await_oneshot_code(), 1)
	# A second emission finds no connection: the registry stays drained.
	target.emit_signal("ping", 2)
	assert_equal(example.await_oneshot_silence_code(), 1)
	# A fresh Await re-subscribes and receives the next emission.
	example.await_oneshot_rearm(target)
	target.emit_signal("ping", 3)
	assert_equal(example.await_oneshot_rearm_code(), 1)

func test_signal_await_misuse(example: Example) -> void:
	print("test signal await misuse")
	var target := AwaitOneShotTarget.new()
	var baseline: int = example.go_callable_registry_count()
	# Unknown signal must panic at the call site...
	assert_equal(example.await_misuse_recovered(target), 1)
	# ...without leaving a dangling subscription behind.
	assert_equal(example.go_callable_registry_count(), baseline)

func test_signal_await_freed(example: Example) -> void:
	print("test signal await freed object")
	var target := AwaitFreedTarget.new()
	example.await_freed_start(target)
	# Drop the only reference: the object dies, the connection goes with it,
	# and the free callback must close the channel without a send.
	target = null
	assert_equal(example.await_freed_code(), 1)

func test_signal_await_timer(example: Example) -> void:
	print("test signal await scene tree timer")
	var timer := get_tree().create_timer(1.0)
	example.await_timer_start(timer)
	# The headless run throttles nothing: --quit-after counts iterations and
	# each real delta is tens of microseconds, so wall-clock timers need a
	# massive scale factor to cross 1.0 s within the frame budget (1e6 makes
	# even a 10 us iteration contribute 10 ms of scaled time).
	Engine.time_scale = 1000000.0
	await get_tree().process_frame
	await get_tree().process_frame
	Engine.time_scale = 1.0
	assert_equal(example.await_timer_code(), 1)

func test_signal_await_shutdown_pending(example: Example) -> void:
	print("test signal await pending at shutdown")
	# Keep the target alive until engine teardown by parking it on self.
	var target := AwaitFreedTarget.new()
	_pending_await_target = target
	example.await_shutdown_arm(target)
	# No probe: teardown must drain the subscription; the suite gate
	# (banner, zero failures, no leaks) judges the run.

func test_static_methods(example: Example) -> void:
	# Bind-side static methods (openspec: support-static-methods).
	# Typed GDScript call through the class path: ptrcall fast path, engine
	# passes a null instance pointer and the bind dispatches without receiver.
	assert_equal(Example.static_add(2, 3), 5)
	# Default-filled calls travel the varcall channel: the trailing bonus
	# parameter is filled from the registered default.
	assert_equal(Example.static_with_default(5), 15)
	assert_equal(Example.static_with_default(5, 100), 105)
	# Callable-routed and Variant-typed invocations take the varcall path.
	var static_callable: Callable = Example.static_add
	assert_equal(static_callable.call(40, 2), 42)
	var boxed: Variant = 20
	assert_equal(Example.static_add(boxed, 22), 42)
	# Variadic static collects the raw Variants with no receiver prepended.
	assert_equal(Example.static_sum_varargs(1, 2, 3, 4), 10)
	assert_equal(Example.static_sum_varargs(), 0)
	# Object argument and object return with no instance anywhere: the
	# standard decoding and ownership paths serve static binds unchanged.
	assert_equal(Example.static_object_class(example), "Example")
	assert_equal(Example.static_echo_node(example), example)
	# ClassDB introspection sees the static on the class, no instance needed.
	assert_true(ClassDB.class_has_method(&"Example", &"static_add", true))
	# Wrong arity on the static is rejected through the engine's call-error
	# channel without dispatching the Go body, and a correctly-shaped call on
	# the same bind works immediately after. Driven from Go because a
	# GDScript-level bad call aborts this function instead of letting the
	# rejection be observed (the engine still prints its own error line).
	assert_equal(example.static_bad_arity_rejected(), 1)
	# Bind-time misuse fails fatally and is recoverable; the rejected
	# duplicate leaves the original registration answering.
	assert_equal(example.static_duplicate_bind_recovered(), 1)
	assert_equal(Example.static_add(1, 2), 3)
	assert_equal(example.static_non_callable_recovered(), 1)
	# Observation only (no legality asserted): the static invoked on the
	# instance's generic Object.call channel, driven from Go so no script
	# stack can abort on it. Whatever code prints, the assertion after this
	# line is the liveness proof.
	print("static-via-instance observation: %d" % example.static_instance_call_observation())
	assert_equal(Example.static_add(10, 10), 20)

func test_editor_plugin_runtime_guard(example: Example) -> void:
	# Editor plugin operations outside an editor build must reject with a
	# recoverable panic (openspec: support-editor-plugins): the engine-side
	# registration is a silent no-op in runtime builds, so silently
	# pretending to succeed is the failure mode guarded against. The
	# recoverable PANIC logs this provokes are announced by the Go probe.
	assert_equal(example.editor_plugin_rejected_outside_editor(), 1)

func test_instance_churn(example: Example) -> void:
	# Instance-scoped Go retention gates (openspec fix-instance-handle-lifecycle,
	# issue #152): user-defined instances freed by the engine must stop
	# retaining their Go wrappers, and engine-class binding churn must release
	# every created binding handle exactly once.
	for i in 250:
		var warm := ExampleRef.new()
		warm.set_id(i)
	example.test_instance_churn_begin()
	for i in 5000:
		var r := ExampleRef.new()
		r.set_id(i)
		example.test_instance_churn_capture(r)
	assert_equal(example.test_instance_churn_end(1, 5000), 1)
	for i in 25000:
		var r := ExampleRef.new()
		r.set_id(i)
		example.test_instance_churn_capture(r)
	assert_equal(example.test_instance_churn_end(2, 25000), 1)
	assert_equal(example.test_instance_churn_verify(), 1)

# Query/release retention gate for the per-query property arrays (openspec
# fix-class-property-list-lifecycle): thousands of get_property_list cycles
# must leave no query array outstanding (each released by the paired
# free_property_list2) and must retain no Go heap. The arrays are C memory,
# so the release accounting itself is asserted on the Go side. The teardown
# check in the same fixture drives cache teardown after zero, one, and many
# query cycles.
func test_property_list_churn(example: Example) -> void:
	assert_equal(example.test_prop_list_teardown_check(), 1)
	for i in 200:
		example.get_property_list()
	assert_equal(example.test_prop_list_churn_begin(), 1)
	var last_size := 0
	for i in 4000:
		last_size = example.get_property_list().size()
	assert_true(last_size > 0)
	assert_equal(example.test_prop_list_churn_end(4000), 1)

func test_binding_release_accounting(example: Example) -> void:
	example.test_binding_churn_begin()
	for i in 5000:
		var b := RefCounted.new()
		example.test_binding_churn_capture(b)
	assert_equal(example.test_binding_churn_verify(), 1)

# Issue #151 regression: the generated vararg bindings (EmitSignal, Rpc,
# RpcId) used to write the engine's 24-byte Variant return into an 8-byte
# Error buffer, corrupting adjacent Go memory ~1 run in 15. See
# test/pkg/vararg_return_tests.go for the Go side.
func test_vararg_return_buffer(example: Example) -> void:
	assert_equal(example.test_vararg_return_heap_integrity(400), 1)
	# An out-of-tree Node makes rpc/rpc_id deterministically ERR_UNCONFIGURED
	# (Node::rpcp fails before reading multiplayer state), so the Error
	# conversion is checked against a value the engine guarantees.
	var loose_node := Node.new()
	assert_equal(example.test_vararg_error_returns(loose_node), 1)
	loose_node.free()

# Witnesses for openspec document-binding-substrate. The demo already exercises
# most of the substrate; these assertions pin the parts that had no direct
# caller: cast identity, the registration bookkeeping, the handled get/set path
# as a plain property round-trip, a variadic bind with a second return, and the
# advertised shape of Variant slots.
func test_binding_substrate(example: Example) -> void:
	# Cast contract: identity for the extension class and its parent, nil for an
	# incompatible class, and inheritance accepts ancestors only.
	assert_equal(example.test_object_cast(), 0)

	# Registration leaves the class in the registry at the scene level, with a
	# callable constructor and an encoder that accepts an instance.
	assert_equal(example.test_registration_bookkeeping(), 0)

	# _get/_set handled path: dproperty_* is reachable only through the class
	# callbacks, so a round-trip proves the value went through both.
	example.dproperty_0 = Vector2(1, 2)
	example.dproperty_1 = Vector2(3, 4)
	example.dproperty_2 = Vector2(5, 6)
	assert_equal(example.dproperty_0, Vector2(1, 2))
	assert_equal(example.dproperty_1, Vector2(3, 4))
	assert_equal(example.dproperty_2, Vector2(5, 6))

	# A variadic bind with two declared returns delivers only the first.
	assert_equal(example.varargs_func_nv_bool(), 70)
	assert_equal(example.varargs_func_nv_bool("a", "b"), 72)

	# Static variadic dispatch passes the caller's arguments with no receiver.
	assert_equal(Example.static_sum_varargs(), 0)
	assert_equal(Example.static_sum_varargs(1, 2, 3, 4), 10)

	# A Variant return advertises nil plus the NIL_IS_VARIANT usage flag; a
	# Variant parameter advertises nil without that flag. The calls below prove
	# the parameter still dispatches, so a metadata-only assertion cannot pass
	# while decoding is dead.
	var vararg_return = null
	var variant_arg = null
	for entry in example.get_method_list():
		if entry["name"] == "varargs_func":
			vararg_return = entry["return"]
		elif entry["name"] == "test_variant_vector2i_conversion":
			variant_arg = entry["args"][0]
	assert_true(vararg_return != null)
	assert_equal(vararg_return["type"], TYPE_NIL)
	assert_true((vararg_return["usage"] & PROPERTY_USAGE_NIL_IS_VARIANT) != 0)
	assert_true(variant_arg != null)
	assert_equal(variant_arg["type"], TYPE_NIL)
	assert_true((variant_arg["usage"] & PROPERTY_USAGE_NIL_IS_VARIANT) == 0)
	assert_equal(example.test_variant_vector2i_conversion(Vector2i(1, 1)), Vector2i(1, 1))
	assert_equal(example.test_variant_vector2i_conversion(Vector2(1.0, 1.0)), Vector2i(1, 1))

# Drives the C-side capture at GDExtension callback boundaries (openspec:
# windows-stacktrace-capture). The Go side prints "stacktrace-probe
# capture=<0|1>" from the binary's compile-time truth and opens a narrow
# capture window around the calls below. The runtime gate requires at least
# one "___ C start backtrace ___" banner whenever capture=1 is claimed, and
# these value assertions pass identically on stub platforms.
func test_stacktrace_probe(example: Example) -> void:
	assert_equal(example.test_stacktrace_probe("start"), "on")
	assert_equal(example.test_stacktrace_probe("hit"), "hit")
	assert_equal(example.test_stacktrace_probe("stop"), "off")
