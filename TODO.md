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

## cgo references
* GopherCon 2018 - Adventures in Cgo Performance: https://about.sourcegraph.com/blog/go/gophercon-2018-adventures-in-cgo-performance
* FFI Overhead: https://github.com/dyu/ffi-overhead