package gdclassimpl

import (
	"fmt"
	"strings"

	"github.com/godot-go/godot-go/cmd/extensionapiparser"
)

// A vararg class call is a varcall: object_method_bind_call always writes a
// full Variant into r_return (godot-cpp binding_generator.py declares
// "Variant ret;" for every vararg binding and casts the result to the
// declared type). The Go return buffer therefore must be a Variant, and any
// non-Variant declared return type needs a conversion from it.
//
// varargReturnConversions maps the declared Go return type to the expression
// that converts the engine-written Variant buffer named retVariant to that
// type. An empty expression means no conversion: the empty type is void (a
// null return pointer, no buffer at all) and "Variant" hands the engine's
// Variant back untouched, matching the existing vararg bindings.
var varargReturnConversions = map[string]string{
	"":        "",
	"Variant": "",
	"Error":   "Error(retVariant.ToInt64())",
}

// varargReturnConversion returns the conversion expression for a vararg
// method's declared Go return type, and whether the type is supported.
func varargReturnConversion(goType string) (string, bool) {
	expr, ok := varargReturnConversions[goType]
	return expr, ok
}

// varargReturnExpr is the template function rendering the conversion of the
// engine-written Variant buffer to the declared return type. It panics for
// types the generator guard would have rejected, so a template never emits
// a conversion that varargReturnConversions does not declare.
func varargReturnExpr(goType string) string {
	expr, ok := varargReturnConversion(goType)
	if !ok {
		panic("unsupported vararg return type " + goType)
	}
	return expr
}

// checkVarargReturnTypes fails generation when a generated vararg class
// method declares a return type that has no conversion from the engine's
// Variant return buffer. Adding a type here means extending
// varargReturnConversions and the classes.go.tmpl conversion path together.
func checkVarargReturnTypes(eapi extensionapiparser.ExtensionApi) error {
	var problems []string
	for _, c := range eapi.Classes {
		for _, m := range c.Methods {
			if !m.IsVararg || m.IsVirtual {
				continue
			}
			rt := goReturnType(typeOrMeta(m.ReturnValue.Meta, m.ReturnValue.Type))
			if _, ok := varargReturnConversion(rt); !ok {
				problems = append(problems, fmt.Sprintf("%s.%s declares return type %q", c.Name, m.Name, rt))
			}
		}
	}
	if len(problems) > 0 {
		return fmt.Errorf("vararg methods need a Variant return conversion; unsupported declared return types: %s", strings.Join(problems, "; "))
	}
	return nil
}
