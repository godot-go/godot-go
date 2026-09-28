extends Node

# Negative fixture for the object-argument type metadata (openspec:
# fix-object-argument-type-metadata).
#
# This script is deliberately NOT loaded by main.gd. It must FAIL to parse: it
# passes a typed `Image` where the bound method advertises a `CollisionShape2D`
# parameter. That rejection is the point -- it proves the advertised class is
# actually enforced by GDScript static typing rather than being decorative, and
# that fixing the metadata did not simply turn everything into `Object` and stop
# checking.
#
# Verify (expected: a parse error naming the argument and the class mismatch):
#   GODOT=/path/to/godot make check_type_rejection
#
# If this file ever parses cleanly, the metadata is wrong again.

func reject_wrong_class(example: Example) -> void:
	var img: Image = Image.new()
	# test_object_arg_set_shape advertises (CollisionShape2D, RefShape2D).
	example.test_object_arg_set_shape(img, img)
