package gdnativetest

/*
#include <cgo_example.h>
#include <gdnative.wrapper.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"github.com/godot-go/godot-go/pkg/gdnative"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GodotObject", func() {
	When("calling GetClass()", func() {
		It("should return 'Object'", func() {
			obj := gdnative.NewObject()
			defer obj.Free()

			c := obj.GetClass()
			className := c.AsGoString()

			Ω(className).Should(Equal("Object"))
		})
	})

	When("calling GetMethodList() on an Object", func() {
		It("should return an array of methods containing a 'get_class' method", func() {
			obj := gdnative.NewObject()
			defer obj.Free()

			Ω(obj).Should(Not(BeNil()))

			arr := obj.GetMethodList()
			defer arr.Destroy()

			Ω(arr.Size()).Should(BeNumerically(">=", int32(40)))

			found := false

			for i := int32(0); i < arr.Size(); i++ {
				v := arr.Get(i)
				Ω(v.GetType()).Should(Equal(gdnative.GODOT_VARIANT_TYPE_DICTIONARY))
				dict := v.AsDictionary()

				nameStr := gdnative.NewStringFromGoString("name")
				defer nameStr.Destroy()

				vName := gdnative.NewVariantString(nameStr)
				getClassV := dict.Get(vName)
				defer getClassV.Destroy()

				Ω(getClassV.GetType()).Should(Equal(gdnative.GODOT_VARIANT_TYPE_STRING))

				getClassStr := getClassV.AsString()

				if getClassStr.AsGoString() == "get_class" {
					found = true
					break
				}
			}

			Ω(found).Should(BeTrue())
		})
	})
})
