---
name: update-godot-gdextension-interface
description: >-
  Update Godot GDExtension C header files and json metadata.
---
Run `godot --version` to determine the godot version:
```
% godot --version
4.5.1.stable.custom_build.f62fdbde1
```

## Update `godot_headers/` Files

Run `make update_godot_headers_from_binary` if the `extension_api.json` doesn't match the godot version:
```json
...
	"header": {
		"version_major": 4,
		"version_minor": 5,
		"version_patch": 1,
		"version_status": "stable",
		"version_build": "custom_build",
		"version_full_name": "Godot Engine v4.5.1.stable.custom_build",
		"precision": "single"
	},
...
```

This does 2 things 
1. GDExtension header file "gdextension_interface.h". This file is the base file required to implement a GDExtension.
2. Generate a JSON dump of the Godot API for GDExtension bindings named "extension_api.json" in the current folder.


## Generate Go Codegen wrapping the godot_headers files

Run `make generate` 
