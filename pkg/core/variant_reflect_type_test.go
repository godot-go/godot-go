package core

import (
	"reflect"
	"testing"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
)

// TestVariantTypeSliceMappingToPackedArrays pins the closed D2 table of the
// support-native-slice-marshalling design: every supported slice element maps
// to its engine packed array variant type, never to the element's scalar type.
func TestVariantTypeSliceMappingToPackedArrays(t *testing.T) {
	cases := []struct {
		name string
		typ  reflect.Type
		want GDExtensionVariantType
	}{
		{"byte", reflect.TypeOf([]byte(nil)), GDEXTENSION_VARIANT_TYPE_PACKED_BYTE_ARRAY},
		{"int32", reflect.TypeOf([]int32(nil)), GDEXTENSION_VARIANT_TYPE_PACKED_INT32_ARRAY},
		{"int", reflect.TypeOf([]int(nil)), GDEXTENSION_VARIANT_TYPE_PACKED_INT64_ARRAY},
		{"int64", reflect.TypeOf([]int64(nil)), GDEXTENSION_VARIANT_TYPE_PACKED_INT64_ARRAY},
		{"float32", reflect.TypeOf([]float32(nil)), GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY},
		{"float64", reflect.TypeOf([]float64(nil)), GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT64_ARRAY},
		{"string", reflect.TypeOf([]string(nil)), GDEXTENSION_VARIANT_TYPE_PACKED_STRING_ARRAY},
		{"Vector2", reflect.TypeOf([]Vector2(nil)), GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR2_ARRAY},
		{"Vector3", reflect.TypeOf([]Vector3(nil)), GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR3_ARRAY},
		{"Vector4", reflect.TypeOf([]Vector4(nil)), GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR4_ARRAY},
		{"Color", reflect.TypeOf([]Color(nil)), GDEXTENSION_VARIANT_TYPE_PACKED_COLOR_ARRAY},
		// The variadic binding convention predates packed-slice marshalling:
		// a trailing []Variant slot registers as the scalar VARIANT type.
		{"Variant (variadic slot)", reflect.TypeOf([]Variant(nil)), GDEXTENSION_VARIANT_TYPE_VARIANT_MAX},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ReflectTypeToGDExtensionVariantType(c.typ); got != c.want {
				t.Fatalf("variant type for %s: got %v, want %v", c.name, got, c.want)
			}
		})
	}
}

// TestVariantTypeSliceRejectionPanics pins bind-time rejection for element
// types without a packed array counterpart (design D2): binding fails loudly
// naming the element type instead of silently registering a wrong variant.
func TestVariantTypeSliceRejectionPanics(t *testing.T) {
	rejected := []any{
		[]bool{},
		[]uint32{},
		[]uint64{},
		[]Vector2i{},
		[]Vector3i{},
		[]String{},
		[][]int{},
	}
	for _, sample := range rejected {
		sliceType := reflect.TypeOf(sample)
		t.Run(sliceType.String(), func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("expected bind-time panic for %s, got none", sliceType)
				}
			}()
			ReflectTypeToGDExtensionVariantType(sliceType)
		})
	}
}
