package builtin

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
	"github.com/iancoleman/strcase"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var (
	goReturnType = goArgumentType
)

func goFormatFieldName(n string) string {
	n = goArgumentName(n)

	return fmt.Sprintf("%s%s", strings.ToUpper(n[:1]), n[1:])
}

var (
	underscoreDigitRe = regexp.MustCompile(`_(\d)`)
	digitDRe          = regexp.MustCompile(`(\d)_([iIdD])`)
)

func screamingSnake(v string) string {
	v = strcase.ToScreamingSnake(v)

	return digitDRe.ReplaceAllString(underscoreDigitRe.ReplaceAllString(v, `$1`), `${1}${2}`)
}

func goArgumentName(t string) string {
	switch t {
	case "string":
		return "strValue"
	case "internal":
		return "internalMode"
	case "type":
		return "typeName"
	case "range":
		return "valueRange"
	case "default":
		return "defaultName"
	case "interface":
		return "interfaceName"
	case "map":
		return "resourceMap"
	case "var":
		return "varName"
	case "func":
		return "callbackFunc"
	default:
		return t
	}
}

func typeHasPtr(t string) bool {
	switch t {
	case "float", "int", "Object":
		return false
	default:
		return true
	}
}

func goDecodeNumberType(t string) string {
	switch t {
	case "uint","int","uint8","int8","uint16","int16","uint32","int32","uint64":
		return "int64"
	case "float32", "float64":
		return "float64"
	default:
		// log.Panic("unhandled number type", zap.String("type", t))
		return t
	}
}

func goVariantFunc(t string, arg string, classes []extensionapiparser.Class) string {
	if strings.HasPrefix(t, "enum::") {
		return fmt.Sprintf("NewVariantInt64(int64(%s))", arg)
	}

	if strings.HasPrefix(t, "const ") {
		t = t[6:]
	}

	if strings.HasPrefix(t, "bitfield") {
		return fmt.Sprintf("NewVariantInt64(int64(%s))", arg)
	}

	if strings.HasPrefix(t, "typedarray::") {
		t = t[12:]
	}

	if strings.HasSuffix(t, "**") {
		t = strings.TrimSpace(t[:len(t)-2])
	}

	if strings.HasSuffix(t, "*") {
		t = strings.TrimSpace(t[:len(t)-1])
	}

	switch t {
	case "Vector2i", "Vector3i", "Vector4i", "Rect2i":
	case "float", "real_t":
		t = "Float64"
		arg = fmt.Sprintf("float64(%s)", arg)
	case "double":
		t = "Float64"
	case "int8", "int16", "int", "int32":
		t = "Int64"
		arg = fmt.Sprintf("int64(%s)", arg)
	case "int64":
		t = "Int64"
	case "uint8", "uint8_t", "uint16", "uint16_t", "uint32", "uint32_t", "uint64", "uint64_t":
		t = "Int64"
		arg = fmt.Sprintf("int64(%s)", arg)
	case "bool":
		t = "Bool"
	case "String":
		t = "String"
	case "Nil":
		t = "Variant"
	case "Variant":
		return arg
	default:
		found := false
		for _, c := range classes {
			if c.Name == t {
				t = "Object"
				found = true
				break
			}
		}
		if !found {
			t = strcase.ToCamel(t)
		}
	}
	return fmt.Sprintf("NewVariant%s(%s)", t, arg)
}

func goArgumentType(t string) string {
	if strings.HasPrefix(t, "enum::") {
		t = t[6:]
	}

	if strings.HasPrefix(t, "const ") {
		t = t[6:]
	}

	if strings.HasPrefix(t, "bitfield") {
		t = t[8:]
	}

	var (
		indirection int
		isTypedArray bool
	)

	if strings.HasPrefix(t, "typedarray::") {
		t = t[12:]
		isTypedArray = true
	}

	if strings.HasSuffix(t, "**") {
		indirection = 2
		t = strings.TrimSpace(t[:len(t)-2])
	}

	if strings.HasSuffix(t, "*") {
		indirection = 1
		t = strings.TrimSpace(t[:len(t)-1])
	}

	switch t {
	case "void", "":
		if isTypedArray {
			log.Panic("unexpected type array")
		}

		switch indirection {
		case 0:
			return ""
		case 1:
			return "unsafe.Pointer"
		case 2:
			return "*unsafe.Pointer"
		default:
			panic("unexepected pointer indirection")
		}
	case "Vector2i", "Vector3i", "Vector4i", "Rect2i":
	case "float", "real_t":
		t = "float32"
	case "double":
		t = "float64"
	case "int8":
		t = "int8"
	case "int16":
		t = "int16"
	case "int32":
		t = "int32"
	case "int", "int64":
		t = "int64"
	case "uint8", "uint8_t":
		t = "uint8"
	case "uint16", "uint16_t":
		t = "uint16"
	case "uint32", "uint32_t":
		t = "uint32"
	case "uint64", "uint64_t":
		t = "uint64"
	case "bool":
		t = "bool"
	case "String":
		t = "String"
	case "Nil":
		t = "Variant"
	default:
		t = strcase.ToCamel(t)
	}

	// if isTypedArray {
	// 	return "[]" + strings.Repeat("*", indirection) + t
	// } else {
	// 	return strings.Repeat("*", indirection) + t
	// }
	return strings.Repeat("*", indirection) + t
}

func coalesce(params ...string) string {
	for _, p := range params {
		if p != "" {
			return p
		}
	}

	return ""
}

func goHasArgumentTypeEncoder(t string) bool {
	if strings.HasPrefix(t, "enum::") {
		t = t[6:]
	}

	if strings.HasPrefix(t, "const ") {
		t = t[6:]
	}

	if strings.HasPrefix(t, "bitfield") {
		t = t[8:]
	}

	var (
		indirection int
	)

	if strings.HasSuffix(t, "**") {
		indirection = 2
		t = strings.TrimSpace(t[:len(t)-2])
	}

	if strings.HasSuffix(t, "*") {
		indirection = 1
		t = strings.TrimSpace(t[:len(t)-1])
	}

	switch t {
	case "void":
		switch indirection {
		case 0:
			return false
		case 1:
			return false
		case 2:
			return false
		default:
			panic("unexepected pointer indirection")
		}
	case "Vector2i", "Vector3i", "Vector4i", "Rect2i":
		return true
	case "float", "real_t":
		return true
	case "double":
		return true
	case "int":
		return true
	case "uint64_t":
		return true
	case "bool":
		return true
	case "String":
		return true
	case "Nil":
		return true
	case "":
		return false
	}

	return true
}

func goClassInterfaceName(c string) string {
	return c
}

func snakeCase(v string) string {
	ret := strcase.ToSnake(v)

	ret = strings.Replace(ret, "_64_", "64_", 1)
	ret = strings.Replace(ret, "_32_", "32_", 1)
	ret = strings.Replace(ret, "_16_", "16_", 1)
	ret = strings.Replace(ret, "_8_", "8_", 1)
	ret = strings.Replace(ret, "_4_i", "4i", 1)
	ret = strings.Replace(ret, "_3_i", "3i", 1)
	ret = strings.Replace(ret, "_2_i", "2i", 1)
	ret = strings.Replace(ret, "_4_d", "4d", 1)
	ret = strings.Replace(ret, "_3_d", "3d", 1)
	ret = strings.Replace(ret, "_2_d", "2d", 1)
	ret = strings.Replace(ret, "_4", "4", 1)
	ret = strings.Replace(ret, "_3", "3", 1)
	ret = strings.Replace(ret, "_2", "2", 1)


	return ret
}

func goClassConstantName(c string, n string) string {
	return fmt.Sprintf("%s_%s",
		strings.ToUpper(strcase.ToSnake(c)),
		strings.ToUpper(strcase.ToSnake(n)),
	)
}

func goMethodName(n string) string {
	if strings.HasPrefix(n, "_") {
		return fmt.Sprintf("Internal_%s", strcase.ToCamel(n))
	}

	return strcase.ToCamel(n)
}

func nativeStructureFormatToFields(f string) string {
	sb := strings.Builder{}
	fields := strings.Split(f, ";")

	for i := range fields {
		fields[i] = strings.TrimSpace(fields[i])
		pair := strings.SplitN(fields[i], " ", 2)

		t := strings.TrimSpace(pair[0])
		n := strings.TrimSpace(pair[1])

		if strings.Contains(n, "=") {
			nPairs := strings.SplitN(n, "=", 2)
			n = nPairs[0]
		}

		if strings.HasPrefix(n, "(*") {
			sb.WriteString("/* ")

			hasPointer := strings.HasPrefix(n, "*")

			if hasPointer {
				n = n[1:]
			}

			sb.WriteString(goFormatFieldName(n))
			sb.WriteString(" ")

			if strings.HasPrefix(n, "*") {
				sb.WriteString("*")
			}

			sb.WriteString(goArgumentType(t))
			sb.WriteString("*/\n")
		} else {

			hasPointer := strings.HasPrefix(n, "*")

			if hasPointer {
				n = n[1:]
			}

			sb.WriteString(goFormatFieldName(n))
			sb.WriteString(" ")

			if t == "void" && hasPointer {
				sb.WriteString("unsafe.Pointer")
			} else {
				if strings.HasPrefix(n, "*") {
					sb.WriteString("*")
				}

				sb.WriteString(goArgumentType(t))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

var (
	operatorIdName = map[string]string{
		"==":     "equal",
		"!=":     "not_equal",
		"<":      "less",
		"<=":     "less_equal",
		">":      "greater",
		">=":     "greater_equal",
		"+":      "add",
		"-":      "subtract",
		"*":      "multiply",
		"/":      "divide",
		"unary-": "negate",
		"unary+": "positive",
		"%":      "module", // this seems like a mispelling, but it stems from gdextension_interface.h constant GDEXTENSION_VARIANT_OP_MODULE
		"<<":     "shift_left",
		">>":     "shift_right",
		"&":      "bit_and",
		"|":      "bit_or",
		"^":      "bit_xor",
		"~":      "bit_negate",
		"and":    "and",
		"or":     "or",
		"xor":    "xor",
		"not":    "not",
		"in":     "in",
	}
)

func getOperatorIdName(op string) string {
	return operatorIdName[op]
}

func lowerFirstChar(n string) string {
	return fmt.Sprintf("%s%s", strings.ToLower(n[:1]), n[1:])
}

func upperFirstChar(n string) string {
	return fmt.Sprintf("%s%s", strings.ToUpper(n[:1]), n[1:])
}

var (
	needsCopySet = map[string]struct{}{
		"Dictionary": {},
	}
)

func goEncoder(goType string) string {
	return upperFirstChar(goType) + "Encoder"
}

type encoderTypeMetadata struct {
	IsReference bool
	Encodings []encoding
}

type encoding struct {
	Name string
	NativeType string
	EncodeType string
}

var encoderTypeMap = map[string]encoderTypeMetadata{
	"GDEXTENSION_VARIANT_TYPE_NIL": {
		IsReference: false,
		Encodings: []encoding{
			{ "Nil", "Nil", "Nil" },
		},
	},

	/*  atomic types */
	"GDEXTENSION_VARIANT_TYPE_BOOL": {
		IsReference: false,
		Encodings: []encoding{
			{ "Bool", "bool", "uint8" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_INT": {
		IsReference: false,
		Encodings: []encoding{
			{ "Uint", "uint", "int64" },
			{ "Int", "int", "int64" },
			{ "Uint8", "uint8", "int64" },
			{ "Int8", "int8", "int64" },
			{ "Uint16", "uint16", "int64" },
			{ "Int16", "int16", "int64" },
			{ "Uint32", "uint32", "int64" },
			{ "Int32", "int32", "int64" },
			{ "Uint64", "uint64", "int64" },
			{ "Int64", "int64", "int64" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_FLOAT": {
		IsReference: false,
		Encodings: []encoding{
			{ "Float32", "float32", "float64" },
			{ "Float64", "float64", "float64" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_STRING": {
		IsReference: true,
		Encodings: []encoding{
			{ "String", "String", "String", },
		},
	},

	/* math types */
	"GDEXTENSION_VARIANT_TYPE_VECTOR2": {
		IsReference: true,
		Encodings: []encoding{
			{ "Vector2", "Vector2", "Vector2" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_VECTOR2I": {
		IsReference: true,
		Encodings: []encoding{
			{ "Vector2i", "Vector2i", "Vector2i" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_VECTOR3": {
		IsReference: true,
		Encodings: []encoding{
			{ "Vector3", "Vector3", "Vector3" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_VECTOR3I": {
		IsReference: true,
		Encodings: []encoding{
			{ "Vector3i", "Vector3i", "Vector3i" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_TRANSFORM2D": {
		IsReference: true,
		Encodings: []encoding{
			{ "Transform2D", "Transform2D", "Transform2D" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_VECTOR4": {
		IsReference: true,
		Encodings: []encoding{
			{ "Vector4", "Vector4", "Vector4" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_VECTOR4I": {
		IsReference: true,
		Encodings: []encoding{
			{ "Vector4i", "Vector4i", "Vector4i" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_PLANE": {
		IsReference: true,
		Encodings: []encoding{
			{ "Plane", "Plane", "Plane" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_QUATERNION": {
		IsReference: true,
		Encodings: []encoding{
			{ "Quaternion", "Quaternion", "Quaternion" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_AABB": {
		IsReference: true,
		Encodings: []encoding{
			{ "AABB", "AABB", "AABB" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_BASIS": {
		IsReference: true,
		Encodings: []encoding{
			{ "Basis", "Basis", "Basis" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_TRANSFORM3D": {
		IsReference: true,
		Encodings: []encoding{
			{ "Transform3D", "Transform3D", "Transform3D" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_PROJECTION": {
		IsReference: true,
		Encodings: []encoding{
			{ "Projection", "Projection", "Projection" },
		},
	},

	/* misc types */
	"GDEXTENSION_VARIANT_TYPE_COLOR": {
		IsReference: true,
		Encodings: []encoding{
			{ "Color", "Color", "Color" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_STRING_NAME": {
		IsReference: true,
		Encodings: []encoding{
			{ "StringName", "StringName", "StringName" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_NODE_PATH": {
		IsReference: true,
		Encodings: []encoding{
			{ "NodePath", "NodePath", "NodePath" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_RID": {
		IsReference: true,
		Encodings: []encoding{
			{ "RID", "RID", "RID" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_OBJECT": {
		IsReference: true,
		Encodings: []encoding{
			{ "GodotObject", "GodotObject", "GodotObject" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_CALLABLE": {
		IsReference: true,
		Encodings: []encoding{
			{ "Callable", "Callable", "Callable" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_SIGNAL": {
		IsReference: true,
		Encodings: []encoding{
			{ "Signal", "Signal", "Signal" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_DICTIONARY": {
		IsReference: true,
		Encodings: []encoding{
			{ "Dictionary", "Dictionary", "Dictionary" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_ARRAY": {
		IsReference: true,
		Encodings: []encoding{
			{ "Array", "Array", "Array" },
		},
	},

	/* typed arrays */
	"GDEXTENSION_VARIANT_TYPE_PACKED_BYTE_ARRAY": {
		IsReference: true,
		Encodings: []encoding{
			{ "PackedByteArray", "PackedByteArray", "PackedByteArray" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_PACKED_INT32_ARRAY": {
		IsReference: true,
		Encodings: []encoding{
			{ "PackedInt32Array", "PackedInt32Array", "PackedInt32Array" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_PACKED_INT64_ARRAY": {
		IsReference: true,
		Encodings: []encoding{
			{ "PackedInt64Array", "PackedInt64Array", "PackedInt64Array" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY": {
		IsReference: true,
		Encodings: []encoding{
			{ "PackedFloat32Array", "PackedFloat32Array", "PackedFloat32Array" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT64_ARRAY": {
		IsReference: true,
		Encodings: []encoding{
			{ "PackedFloat64Array", "PackedFloat64Array", "PackedFloat64Array" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_PACKED_STRING_ARRAY": {
		IsReference: true,
		Encodings: []encoding{
			{ "PackedStringArray", "PackedStringArray", "PackedStringArray" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR2_ARRAY": {
		IsReference: true,
		Encodings: []encoding{
			{ "PackedVector2Array", "PackedVector2Array", "PackedVector2Array" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR3_ARRAY": {
		IsReference: true,
		Encodings: []encoding{
			{ "PackedVector3Array", "PackedVector3Array", "PackedVector3Array" },
		},
	},
	"GDEXTENSION_VARIANT_TYPE_PACKED_COLOR_ARRAY": {
		IsReference: true,
		Encodings: []encoding{
			{ "PackedColorArray", "PackedColorArray", "PackedColorArray" },
		},
	},

	// variant
	"GDEXTENSION_VARIANT_TYPE_VARIANT_MAX": {
		IsReference: true,
		Encodings: []encoding{
			{ "Variant", "Variant", "Variant" },
		},
	},
}

func filterReferences(mds []encoderTypeMetadata) []encoderTypeMetadata {
	filtered := make([]encoderTypeMetadata, 0, len(mds))
	for i := range mds {
		if mds[i].IsReference {
			filtered = append(filtered, mds[i])
		}
	}
	return filtered
}

func mapEncodingNames(mds []encoderTypeMetadata) []string {
	values := make([]string, 0, len(mds) * 2)
	for i := range mds {
		for j := range mds[i].Encodings {
			values = append(values, mds[i].Encodings[j].Name)
		}
	}
	return values
}

var values = maps.Values(encoderTypeMap)
var referenceEncoderTypeNames = mapEncodingNames(filterReferences(values))

func goEncodeIsReference(goType string) bool {
	return slices.Contains(referenceEncoderTypeNames, goType)
}

func astVariantMetadata(cEnumVariantType string) encoderTypeMetadata {
	return encoderTypeMap[cEnumVariantType]
}
