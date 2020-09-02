package gdnativetest

/*
#include <cgo_example.h>
#include <gdnative_wrappergen.h>
#include <stdlib.h>
*/
import "C"
import (
	"github.com/godot-go/godot-go/pkg/gdnative"
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Variant", func() {
	When("calling GoTypeToVariant", func() {
		It("should return a nil type", func() {
			v := gdnative.GoTypeToVariant(reflect.ValueOf(nil))
			defer v.Destroy()

			Ω(v.GetType()).Should(Equal(gdnative.GODOT_VARIANT_TYPE_NIL))
		})

		It("should return an int type from 0 value", func() {
			v := gdnative.GoTypeToVariant(reflect.ValueOf(0))
			defer v.Destroy()

			Ω(v.GetType()).Should(Equal(gdnative.GODOT_VARIANT_TYPE_INT))
		})

		It("should return a string type from empty string", func() {
			str := gdnative.NewStringFromGoString("")
			defer str.Destroy()
			v := gdnative.GoTypeToVariant(reflect.ValueOf(str))
			defer v.Destroy()

			Ω(v.GetType()).Should(Equal(gdnative.GODOT_VARIANT_TYPE_STRING))
		})

		It("should return a string value", func() {
			str := gdnative.NewStringFromGoString("my name")
			defer str.Destroy()
			v := gdnative.GoTypeToVariant(reflect.ValueOf(str))
			defer v.Destroy()
			gs := v.AsString()
			gon := gs.AsGoString()

			Ω(v.GetType()).Should(Equal(gdnative.GODOT_VARIANT_TYPE_STRING))
			Ω(gon).Should(Equal("my name"))
		})

		It("should return a real type with float32", func() {
			var f float32  = 3.6
			v := gdnative.GoTypeToVariant(reflect.ValueOf(f))
			defer v.Destroy()

			Ω(v.GetType()).Should(Equal(gdnative.GODOT_VARIANT_TYPE_REAL))
			Ω(v.AsReal()).Should(BeNumerically("~", 3.6, 0.0001))
		})

		It("should return a real type with float64", func() {
			var f float64  = 3.6
			v := gdnative.GoTypeToVariant(reflect.ValueOf(f))
			defer v.Destroy()

			Ω(v.GetType()).Should(Equal(gdnative.GODOT_VARIANT_TYPE_REAL))
			Ω(v.AsReal()).Should(BeNumerically("~", 3.6, 0.0001))
		})

		It("should return an Array type", func() {
			f := 1.0
			vf := gdnative.GoTypeToVariant(reflect.ValueOf(f))
			defer vf.Destroy()
			str := gdnative.NewStringFromGoString("my name")
			defer str.Destroy()
			vs := gdnative.GoTypeToVariant(reflect.ValueOf(str))
			defer vs.Destroy()
			arr := gdnative.NewArray()
			defer arr.Destroy()
			arr.Append(vf)
			arr.Append(vs)
			v := gdnative.GoTypeToVariant(reflect.ValueOf(arr))
			defer v.Destroy()

			Ω(v.GetType()).Should(Equal(gdnative.GODOT_VARIANT_TYPE_ARRAY))
			Ω(arr.Size()).Should(BeEquivalentTo(2))

			rf := gdnative.VariantToGoType(arr.Get(0))
			Ω(rf.Interface()).Should(Equal(f))

			rs := gdnative.VariantToGoType(arr.Get(1))
			Ω(rs.Interface()).Should(Equal(str))
		})

		It("should return an Object type", func() {
			n := gdnative.NewNode2D()
			defer n.Free()
			o := n.GetOwnerObject()
			strClassName := n.GetClass()
			gs := strClassName.AsGoString()

			Ω(gs).Should(Equal("Node2D"))

			vec := gdnative.NewVector2(1.0, 5.0)

			Ω(vec.GetX()).Should(BeEquivalentTo(1.0))
			Ω(vec.GetY()).Should(BeEquivalentTo(5.0))

			n.SetGlobalPosition(vec)
			p := n.GetPosition()

			Ω(p.GetX()).Should(BeEquivalentTo(1.0))
			Ω(p.GetY()).Should(BeEquivalentTo(5.0))

			n2 := gdnative.NewNode2D()

			n.AddChild(n2, true)

			Ω(n.GetChildCount()).Should(Equal(int32(1)))

			v := gdnative.GoTypeToVariant(reflect.ValueOf(o))
			defer v.Destroy()

			Ω(v.GetType()).Should(Equal(gdnative.GODOT_VARIANT_TYPE_OBJECT))

			newNode2d := gdnative.NewNode2DWithOwner(v.AsObject())
			Ω(n).Should(Equal(newNode2d))
		})
	})
})
