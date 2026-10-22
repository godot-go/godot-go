package core

import (
	"reflect"
	"strings"
	"testing"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/gdclassimpl"
	"github.com/godot-go/godot-go/pkg/gdclassinit"
)

// unregisteredTestInterface is an interface with no class or reference backing,
// used to pin bind-time rejection of unmappable interface spellings.
type unregisteredTestInterface interface {
	Marker()
}

// TestVariantTypeMappingTable pins the closed mapping from Go types to engine
// variant types: scalars, every Godot value type (which is a byte array in Go but
// must advertise its declared type), object carriers, and the untyped sentinel.
func TestVariantTypeMappingTable(t *testing.T) {
	cases := []struct {
		name string
		typ  reflect.Type
		want GDExtensionVariantType
	}{
		{"nil type", nil, GDEXTENSION_VARIANT_TYPE_NIL},
		{"bool", reflect.TypeOf(false), GDEXTENSION_VARIANT_TYPE_BOOL},
		{"int", reflect.TypeOf(int(0)), GDEXTENSION_VARIANT_TYPE_INT},
		{"int8", reflect.TypeOf(int8(0)), GDEXTENSION_VARIANT_TYPE_INT},
		{"int16", reflect.TypeOf(int16(0)), GDEXTENSION_VARIANT_TYPE_INT},
		{"int32", reflect.TypeOf(int32(0)), GDEXTENSION_VARIANT_TYPE_INT},
		{"int64", reflect.TypeOf(int64(0)), GDEXTENSION_VARIANT_TYPE_INT},
		{"uint", reflect.TypeOf(uint(0)), GDEXTENSION_VARIANT_TYPE_INT},
		{"uint8", reflect.TypeOf(uint8(0)), GDEXTENSION_VARIANT_TYPE_INT},
		{"uint16", reflect.TypeOf(uint16(0)), GDEXTENSION_VARIANT_TYPE_INT},
		{"uint32", reflect.TypeOf(uint32(0)), GDEXTENSION_VARIANT_TYPE_INT},
		{"uint64", reflect.TypeOf(uint64(0)), GDEXTENSION_VARIANT_TYPE_INT},
		{"float32", reflect.TypeOf(float32(0)), GDEXTENSION_VARIANT_TYPE_FLOAT},
		{"float64", reflect.TypeOf(float64(0)), GDEXTENSION_VARIANT_TYPE_FLOAT},
		{"string", reflect.TypeOf(""), GDEXTENSION_VARIANT_TYPE_STRING},
		{"String", reflect.TypeOf(String{}), GDEXTENSION_VARIANT_TYPE_STRING},
		{"StringName", reflect.TypeOf(StringName{}), GDEXTENSION_VARIANT_TYPE_STRING_NAME},
		{"NodePath", reflect.TypeOf(NodePath{}), GDEXTENSION_VARIANT_TYPE_NODE_PATH},
		{"Vector2", reflect.TypeOf(Vector2{}), GDEXTENSION_VARIANT_TYPE_VECTOR2},
		{"Vector2i", reflect.TypeOf(Vector2i{}), GDEXTENSION_VARIANT_TYPE_VECTOR2I},
		{"Rect2", reflect.TypeOf(Rect2{}), GDEXTENSION_VARIANT_TYPE_RECT2},
		{"Rect2i", reflect.TypeOf(Rect2i{}), GDEXTENSION_VARIANT_TYPE_RECT2I},
		{"Vector3", reflect.TypeOf(Vector3{}), GDEXTENSION_VARIANT_TYPE_VECTOR3},
		{"Vector3i", reflect.TypeOf(Vector3i{}), GDEXTENSION_VARIANT_TYPE_VECTOR3I},
		{"Vector4", reflect.TypeOf(Vector4{}), GDEXTENSION_VARIANT_TYPE_VECTOR4},
		{"Vector4i", reflect.TypeOf(Vector4i{}), GDEXTENSION_VARIANT_TYPE_VECTOR4I},
		{"Transform2D", reflect.TypeOf(Transform2D{}), GDEXTENSION_VARIANT_TYPE_TRANSFORM2D},
		{"Plane", reflect.TypeOf(Plane{}), GDEXTENSION_VARIANT_TYPE_PLANE},
		{"Quaternion", reflect.TypeOf(Quaternion{}), GDEXTENSION_VARIANT_TYPE_QUATERNION},
		{"AABB", reflect.TypeOf(AABB{}), GDEXTENSION_VARIANT_TYPE_AABB},
		{"Basis", reflect.TypeOf(Basis{}), GDEXTENSION_VARIANT_TYPE_BASIS},
		{"Transform3D", reflect.TypeOf(Transform3D{}), GDEXTENSION_VARIANT_TYPE_TRANSFORM3D},
		{"Color", reflect.TypeOf(Color{}), GDEXTENSION_VARIANT_TYPE_COLOR},
		{"RID", reflect.TypeOf(RID{}), GDEXTENSION_VARIANT_TYPE_RID},
		{"Callable", reflect.TypeOf(Callable{}), GDEXTENSION_VARIANT_TYPE_CALLABLE},
		{"Signal", reflect.TypeOf(Signal{}), GDEXTENSION_VARIANT_TYPE_SIGNAL},
		{"Dictionary", reflect.TypeOf(Dictionary{}), GDEXTENSION_VARIANT_TYPE_DICTIONARY},
		{"Array", reflect.TypeOf(Array{}), GDEXTENSION_VARIANT_TYPE_ARRAY},
		{"PackedByteArray", reflect.TypeOf(PackedByteArray{}), GDEXTENSION_VARIANT_TYPE_PACKED_BYTE_ARRAY},
		{"PackedInt32Array", reflect.TypeOf(PackedInt32Array{}), GDEXTENSION_VARIANT_TYPE_PACKED_INT32_ARRAY},
		{"PackedInt64Array", reflect.TypeOf(PackedInt64Array{}), GDEXTENSION_VARIANT_TYPE_PACKED_INT64_ARRAY},
		{"PackedFloat32Array", reflect.TypeOf(PackedFloat32Array{}), GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT32_ARRAY},
		{"PackedFloat64Array", reflect.TypeOf(PackedFloat64Array{}), GDEXTENSION_VARIANT_TYPE_PACKED_FLOAT64_ARRAY},
		{"PackedStringArray", reflect.TypeOf(PackedStringArray{}), GDEXTENSION_VARIANT_TYPE_PACKED_STRING_ARRAY},
		{"PackedVector2Array", reflect.TypeOf(PackedVector2Array{}), GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR2_ARRAY},
		{"PackedVector3Array", reflect.TypeOf(PackedVector3Array{}), GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR3_ARRAY},
		{"PackedColorArray", reflect.TypeOf(PackedColorArray{}), GDEXTENSION_VARIANT_TYPE_PACKED_COLOR_ARRAY},
		{"PackedVector4Array", reflect.TypeOf(PackedVector4Array{}), GDEXTENSION_VARIANT_TYPE_PACKED_VECTOR4_ARRAY},
		{"Projection", reflect.TypeOf(Projection{}), GDEXTENSION_VARIANT_TYPE_PROJECTION},
		{"Variant", reflect.TypeOf(Variant{}), GDEXTENSION_VARIANT_TYPE_VARIANT_MAX},
		{"RefImage", reflect.TypeFor[RefImage](), GDEXTENSION_VARIANT_TYPE_OBJECT},
		{"bare Ref", reflect.TypeFor[Ref](), GDEXTENSION_VARIANT_TYPE_OBJECT},
		{"*gdclassimpl.NodeImpl", reflect.TypeOf(&gdclassimpl.NodeImpl{}), GDEXTENSION_VARIANT_TYPE_OBJECT},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ReflectTypeToGDExtensionVariantType(c.typ); got != c.want {
				t.Fatalf("variant type for %s: got %v, want %v", c.name, got, c.want)
			}
		})
	}
}

// TestVariantTypeUnsupportedKindRejectionPanics pins the remaining bind-time
// rejections: kinds with no engine representation, unregistered structs and
// interfaces, and fixed arrays whose element is not a byte. The class registry is
// seeded empty so the interface probe fails with its own diagnostic instead of a
// nil-map runtime error.
func TestVariantTypeUnsupportedKindRejectionPanics(t *testing.T) {
	cases := []struct {
		name string
		typ  reflect.Type
		want string
	}{
		{"map", reflect.TypeOf(map[string]int{}), "unhandled reflected go kind"},
		{"chan", reflect.TypeOf((chan int)(nil)), "unhandled reflected go kind"},
		{"func", reflect.TypeOf(func() {}), "unhandled reflected go kind"},
		{"uintptr", reflect.TypeOf(uintptr(0)), "unhandled reflected go kind"},
		{"complex64", reflect.TypeOf(complex64(0)), "unhandled reflected go kind"},
		{"complex128", reflect.TypeOf(complex128(0)), "unhandled reflected go kind"},
		{"unsafe.Pointer", reflect.TypeOf(unsafe.Pointer(nil)), "unhandled reflected go kind"},
		{"pointer to Variant", reflect.TypeOf((*Variant)(nil)), "unhandled go pointer"},
		{"unknown struct", reflect.TypeOf(struct{ X int }{}), "unhandled go struct"},
		{"unknown interface", reflect.TypeFor[unregisteredTestInterface](), "unhandled go interface"},
		{"non-byte array", reflect.TypeOf([4]float32{}), "unhandled go array element type"},
	}

	withRegisteredClasses(t, map[string]*ClassInfo{}, func() {
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				msg, fired := recoverFatal(func() {
					ReflectTypeToGDExtensionVariantType(c.typ)
				})
				if !fired {
					t.Fatalf("expected bind-time panic for %s, got none", c.name)
				}
				if !strings.Contains(msg, c.want) {
					t.Fatalf("expected %q in the diagnostic for %s, got: %s", c.want, c.name, msg)
				}
			})
		}
	})
}

// TestVariantTypeEngineClassInterfaceMapping pins the object spelling a generated
// engine-class interface takes: the mapping recognises it through the engine-class
// constructor table, so the probe needs an entry for the name.
func TestVariantTypeEngineClassInterfaceMapping(t *testing.T) {
	for _, name := range []string{"Node", "Object"} {
		// Restore an existing entry rather than deleting it: the seeded names
		// may be populated by the time this runs in a full-process suite.
		if previous, existed := gdclassinit.GDNativeConstructors.Get(name); existed {
			defer gdclassinit.GDNativeConstructors.Set(name, previous)
		} else {
			defer gdclassinit.GDNativeConstructors.Delete(name)
		}
		gdclassinit.GDNativeConstructors.Set(name, nil)
	}

	withRegisteredClasses(t, map[string]*ClassInfo{}, func() {
		cases := []struct {
			name string
			typ  reflect.Type
		}{
			{"Node interface", reflect.TypeFor[Node]()},
			{"Object interface", reflect.TypeFor[Object]()},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				if got := ReflectTypeToGDExtensionVariantType(c.typ); got != GDEXTENSION_VARIANT_TYPE_OBJECT {
					t.Fatalf("variant type for %s: got %v, want OBJECT", c.name, got)
				}
			})
		}
	})
}

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
