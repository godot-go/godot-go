package constant

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
	"github.com/iancoleman/strcase"
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

func goVariantConstructor(t, innerText string) string {
	switch t {
	case "float", "real_t", "double":
		return fmt.Sprintf("NewVariantFloat64(%s)", innerText)
	case "int", "uint64_t":
		return fmt.Sprintf("NewVariantInt64(int64(%s))", innerText)
	case "bool":
		return fmt.Sprintf("NewVariantBool(%s)", innerText)
	case "String":
		return fmt.Sprintf("NewVariantString(%s)", innerText)
	case "StringName":
		return fmt.Sprintf("NewVariantStringName(%s)", innerText)
	default:
		return fmt.Sprintf("NewVariantObject(%s)", innerText)
	}
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
	case "uint", "int", "uint8", "int8", "uint16", "int16", "uint32", "int32", "uint64":
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
		indirection  int
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

func goClassStructName(c string) string {
	return fmt.Sprintf("%sImpl", c)
}

func goClassEnumName(c, e, n string) string {
	return fmt.Sprintf("%s_%s_%s",
		strings.ToUpper(strcase.ToSnake(c)),
		strings.ToUpper(strcase.ToSnake(e)),
		strings.ToUpper(strcase.ToSnake(n)),
	)
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

var referenceEncoderTypes = []string{
	"Vector2",
	"Vector2i",
	"Vector3",
	"Vector3i",
	"Transform2D",
	"Vector4",
	"Vector4i",
	"Plane",
	"Quaternion",
	"AABB",
	"Basis",
	"Transform3D",
	"Projection",
	"Color",
	"String",
	"StringName",
	"NodePath",
	"RID",
	"Callable",
	"Signal",
	"Dictionary",
	"Array",
	"PackedByteArray",
	"PackedInt32Array",
	"PackedInt64Array",
	"PackedFloat32Array",
	"PackedFloat64Array",
	"PackedStringArray",
	"PackedVector2Array",
	"PackedVector3Array",
	"PackedColorArray",
	"Variant",
	"Object",
}

func goEncodeIsReference(goType string) bool {
	return slices.Contains(referenceEncoderTypes, goType)
}

func isRefcounted(goType string) bool {
	return slices.Contains(referenceEncoderTypes, goType)
}

func isSetterMethodName(methodName string) bool {
	return strings.HasPrefix(methodName, "Set")
}
