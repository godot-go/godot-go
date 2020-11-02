extends CanvasLayer

var player: KinematicBody2D
var speed: Label
var position: Label
var playerName: Label

func _ready():
	var err
	player = $"/root/World/Objects/Player"

	err = player.connect("moved", self, "_on_player_moved")
	if err != OK:
		print("failure to connect to moved player signal")

	err = player.connect("name_changed", self, "_on_player_name_changed")
	if err != OK:
		print("failure to connect to name_changed player signal")

	var stats = $"TopPanel/Margin/Rows/Stats"
	speed = stats.get_node("Speed")
	position = stats.get_node("Position")
	playerName = stats.get_node("Name")


func _on_player_name_changed() -> void:
	playerName.text = player.name


func _on_player_moved(velocity: Vector2) -> void:
	speed.text = "(%.2f, %.2f)" % [velocity.x, velocity.y]
	var pos = player.position
	position.text = "(%.2f, %.2f)" % [pos.x, pos.y]
