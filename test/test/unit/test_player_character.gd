extends "res://addons/gut/test.gd"

const PlayerCharacter = preload("res://libs/player_character.gdns")


func test_player_characater_creation():
	assert_string_starts_with(str(PlayerCharacter), '[NativeScript:')
	assert_eq(typeof(PlayerCharacter), TYPE_OBJECT)
	var inst = PlayerCharacter.new()
	assert_string_starts_with(str(inst), '[KinematicBody2D:')
	inst.free()


func test_player_characater_teleport():
	var inst = PlayerCharacter.new()
	assert_eq(inst.position, Vector2(0, 0))
	inst.random_teleport(0.0)
	assert_eq(inst.position, Vector2(0, 0))
	inst.random_teleport(5.0)
	assert_almost_eq(inst.position, Vector2(-5.0, -5.0), Vector2(5.0, 5.0))
	inst.free()


func test_player_characater_name():
	var inst = PlayerCharacter.new()
	assert_eq(inst.name, '')
	inst.random_name()
	assert_true(inst.name != '')
	inst.free()


func test_run_go_ginkgo_testsuite():
	var inst = PlayerCharacter.new()
	inst.run_ginkgo_testsuite()
	inst.free()
