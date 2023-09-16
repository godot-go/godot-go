Need to find a place to place this logic:
```
	// behavior ported from godot-cpp
	switch hint {
	case PROPERTY_HINT_RESOURCE_TYPE:
		className = hintString
	default:
		className = pClassName
	}
```

* Look into switching over Object to have an opaque array to resolve the segfault issue with ToObject()