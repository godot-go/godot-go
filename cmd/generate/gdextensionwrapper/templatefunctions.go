package gdextensionwrapper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/godot-go/godot-go/cmd/gdextensionparser/clang"
)

func add(a int, b int) int {
	return a + b
}

func goArgumentType(t clang.PrimativeType, name string) string {
	n := strings.TrimSpace(t.Name)

	hasReturnPrefix := strings.HasPrefix(name, "r_")

	switch n {
	case "void":
		if t.IsPointer {
			return "unsafe.Pointer"
		}

		return ""
	case "float", "real_t":
		if t.IsPointer {
			panic(fmt.Sprintf("unhandled type: %s", t.String()))
		}

		return "float32"
	case "size_t":
		if t.IsPointer {
			panic(fmt.Sprintf("unhandled type: %s", t.String()))
		}

		return "uint64"
	case "char":
		if t.IsPointer {
			if hasReturnPrefix {
				return "*Char"
			} else {
				return "string"
			}
		}

		panic(fmt.Sprintf("unhandled type: %s", t.String()))
	case "int32_t":
		if t.IsPointer {
			panic(fmt.Sprintf("unhandled type: %s", t.String()))
		}

		return "int32"
	case "char16_t":
		if t.IsPointer {
			return "*Char16T"
		}

		panic(fmt.Sprintf("unhandled type: %s", t.String()))
	case "char32_t":
		if t.IsPointer {
			return "*Char32T"
		}

		return "Char32T"
	case "wchar_t":
		if t.IsPointer {
			return "*WcharT"
		} else {
			panic(fmt.Sprintf("unhandled type: %s", t.String()))
		}
	case "uint8_t":
		if t.IsPointer {
			return "*Uint8T"
		}

		return "Uint8T"
	case "uint64_t":
		if t.IsPointer {
			return "*Uint64T"
		}

		return "Uint64T"
	default:
		if t.IsPointer {
			return fmt.Sprintf("*%s", n)
		} else {
			return n
		}
	}
}

func goReturnType(t clang.PrimativeType) string {
	n := strings.TrimSpace(t.Name)

	switch n {
	case "float", "real_t":
		if t.IsPointer {
			return "*float32"
		} else {
			return "float32"
		}
	case "double":
		if t.IsPointer {
			return "*float64"
		} else {
			return "float64"
		}
	case "int32_t":
		if t.IsPointer {
			return "*int32"
		} else {
			return "int32"
		}
	case "int64_t":
		if t.IsPointer {
			return "*int64"
		} else {
			return "int64"
		}
	case "uint64_t":
		if t.IsPointer {
			return "*uint64"
		} else {
			return "uint64"
		}
	case "uint8_t":
		if t.IsPointer {
			return "*uint8"
		} else {
			return "uint8"
		}
	case "char16_t":
		if t.IsPointer {
			return "*Char16T"
		} else {
			return "Char16T"
		}
	case "char32_t":
		if t.IsPointer {
			return "*Char32T"
		} else {
			return "Char32T"
		}
	case "void":
		if t.IsPointer {
			return "unsafe.Pointer"
		} else {
			return ""
		}
	default:
		if t.IsPointer {
			return fmt.Sprintf("*%s", n)
		} else {
			return n
		}
	}
}

func goEnumValue(v clang.EnumValue, index int) string {
	if v.IntValue != nil {
		return strconv.Itoa(*v.IntValue)
	} else if v.ConstRefValue != nil {
		return *v.ConstRefValue
	} else if index == 0 {
		return "iota"
	} else {
		return ""
	}
}

func cgoCastArgument(a clang.Argument) string {
	if a.Type.Primative != nil {
		t := a.Type.Primative

		n := strings.TrimSpace(t.Name)

		hasReturnPrefix := strings.HasPrefix(a.Name, "r_")

		switch n {
		case "void":
			if t.IsPointer {
				return fmt.Sprintf("unsafe.Pointer(%s)", a.Name)
			} else {
				panic(fmt.Sprintf("unhandled type: %s", t.String()))
			}
		case "char":
			if t.IsPointer {
				if hasReturnPrefix {
					return fmt.Sprintf("(*C.char)(%s)", a.Name)
				} else {
					return fmt.Sprintf("C.CString(%s)", a.Name)
				}
			} else {
				panic(fmt.Sprintf("unhandled type: %s", t.String()))
			}
		default:
			if t.IsPointer {
				return fmt.Sprintf("(*C.%s)(%s)", n, a.Name)
			} else {
				return fmt.Sprintf("(C.%s)(%s)", n, a.Name)
			}
		}
	} else if a.Type.Function != nil {
		return fmt.Sprintf("(*[0]byte)(%s)", a.Type.Function.Name)
	}

	panic("unhandled type")
}

func cgoCleanUpArgument(a clang.Argument, index int) string {
	if a.Type.Primative != nil {
		t := a.Type.Primative
		n := strings.TrimSpace(t.Name)

		hasReturnPrefix := strings.HasPrefix(a.Name, "r_")

		switch n {
		case "char":
			if t.IsPointer {
				if !hasReturnPrefix {
					return fmt.Sprintf("C.free(unsafe.Pointer(arg%d))", index)
				}
				return ""

			} else {
				panic(fmt.Sprintf("unhandled type: %s", t.String()))
			}
		default:
			return ""
		}
	} else if a.Type.Function != nil {
		return ""
	}

	panic("unhandled type")
}

func cgoCastReturnType(t clang.PrimativeType, argName string) string {
	n := strings.TrimSpace(t.Name)

	switch n {
	case "int32_t":
		if t.IsPointer {
			return fmt.Sprintf("(*int32)(%s)", argName)
		} else {
			return fmt.Sprintf("int32(%s)", argName)
		}
	case "int64_t":
		if t.IsPointer {
			return fmt.Sprintf("(*int64)(%s)", argName)
		} else {
			return fmt.Sprintf("int64(%s)", argName)
		}
	case "uint64_t":
		if t.IsPointer {
			return fmt.Sprintf("(*uint64)(%s)", argName)
		} else {
			return fmt.Sprintf("uint64(%s)", argName)
		}
	case "uint8_t":
		if t.IsPointer {
			return fmt.Sprintf("(*uint8)(%s)", argName)
		} else {
			return fmt.Sprintf("uint8(%s)", argName)
		}
	case "char16_t":
		if t.IsPointer {
			return fmt.Sprintf("(*Char16T)(%s)", argName)
		} else {
			panic(fmt.Sprintf("unhandled type: %s, %v", t.String(), t))
		}
	case "char32_t":
		if t.IsPointer {
			return fmt.Sprintf("(*Char32T)(%s)", argName)
		} else {
			panic(fmt.Sprintf("unhandled type: %s, %v", t.String(), t))
		}
	case "void":
		if t.IsPointer {
			return fmt.Sprintf("unsafe.Pointer(%s)", argName)
		} else {
			panic(fmt.Sprintf("unhandled type: %s", t.String()))
		}
	case "float", "real_t":
		if t.IsPointer {
			return fmt.Sprintf("(*float32)(%s)", argName)
		} else {
			return fmt.Sprintf("float32(%s)", argName)
		}
	case "double":
		if t.IsPointer {
			return fmt.Sprintf("(*float64)(%s)", argName)
		} else {
			return fmt.Sprintf("float64(%s)", argName)
		}
	default:
		if t.IsPointer {
			return fmt.Sprintf("(*%s)(%s)", n, argName)
		} else {
			return fmt.Sprintf("(%s)(%s)", n, argName)
		}
	}
}
