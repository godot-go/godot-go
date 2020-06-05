package shared

// translation between C to Go
var CToGoValueTypeMap = map[string]string{
	"bool":               "bool",
	"uint8_t":            "uint8",
	"uint32_t":           "uint32",
	"uint64_t":           "uint64",
	"int64_t":            "int64",
	"double":             "float64",
	"wchar_t":            "int32",
	"char":               "C.char",
	"int":                "int32",
	"size_t":             "int64",
	"void":               "void",
	"string":             "string",
	"float":              "float32",
	"godot_real":         "float32",
	"godot_bool":         "bool",
	"godot_int":          "int32",
	"godot_object":       "GodotObject",
	"godot_vector3_axis": "Vector3Axis",
	"godot_variant_type": "VariantType",
	"godot_error":        "Error",
	"godot_string":       "String",
}
