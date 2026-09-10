extends Node

# Pins the virtual delegation trap: TestDelegationRepro's _get_maximum_size
# override delegates through the plain wrapper, which re-enters the registered
# virtual. Run via `make test_delegation_trap` (a separate godot invocation
# expecting a non-zero exit with the bounded-depth recursion diagnostic) — not
# in main.tscn, because the normal suite must stay green.
func _ready() -> void:
	var repro := TestDelegationRepro.new()
	add_child(repro)
	repro.get_maximum_size()
	repro.free()
	get_tree().quit(0)
