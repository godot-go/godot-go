package extensionapi

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
		return fmt.Sprintf("NewVariantWrapped(%s)", innerText)
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
	case "uint","int","uint8","int8","uint16","int16","uint32","int32","uint64":
		return "int64"
	case "float32", "float64":
		return "float64"
	default:
		// log.Panic("unhandled number type", zap.String("type", t))
		return t
	}

	panic("unhandled number type")
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
	case "void":
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
	case "int":
		t = "int32"
	case "uint64_t":
		t = "uint64"
	case "bool":
		t = "bool"
	case "String":
		t = "String"
	case "Nil":
		t = "Variant"
	case "":
		t = ""
	default:
		t = strcase.ToCamel(t)
	}

	if isTypedArray {
		return "[]" + strings.Repeat("*", indirection) + t
	} else {
		return strings.Repeat("*", indirection) + t
	}
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
	ret = strings.Replace(ret, "_4_", "4_", 1)
	ret = strings.Replace(ret, "_3_", "3_", 1)
	ret = strings.Replace(ret, "_2_", "2_", 1)

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

func needsCopyInsteadOfMove(typeName string) bool {
	_, ok := needsCopySet[typeName]

	return ok
}

func isCopyConstructor(typeName string, c extensionapiparser.ClassConstructor) bool {
	return len(c.Arguments) == 1 && c.Arguments[0].Type == typeName
}

func goEncoder(goType string) string {
	if goType == "Object" {
		return ""
	}

	return upperFirstChar(goType) + "Encoder"
}

func goEncodeArg(goType string, argName string) string {
	switch goType {
	case "Variant", "Vector3", "Vector3i", "Vector4", "Vector4i",
		"Plane", "AABB", "Basis", "Transform3", "Projection", "Color",
		"Transform3D":
		return fmt.Sprintf("&%s", argName)
	default:
		return argName
	}
}

var referenceEncoderTypes = []string{
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
