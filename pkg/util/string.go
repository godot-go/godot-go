package util

import (
	"reflect"
	"strings"
)

func ReflectValueSliceToString(values []reflect.Value) string {
	var sb strings.Builder
	sb.WriteString("(")
	for i := range values {
		if i != 0 {
			sb.WriteString(",")
		}
		sb.WriteString(values[i].Type().Name())
		sb.WriteString("/")
		sb.WriteString(values[i].Kind().String())
		sb.WriteString("(")
		sb.WriteString(values[i].String())
		sb.WriteString(")")
	}
	sb.WriteString(")")
	return sb.String()
}
