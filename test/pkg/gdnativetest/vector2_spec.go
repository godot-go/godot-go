package gdnativetest

import (
	"math"

	"github.com/godot-go/godot-go/pkg/gdnative"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vector2", func() {
	Context("SetX()", func() {
		It("should update the X value", func() {
			v := gdnative.NewVector2(1.0, 0.0)
			v.SetX(2.0)
			Ω(v.GetX()).Should(BeNumerically("~", 2.0, 0.001))
		})
	})

	Context("SetY()", func() {
		It("should update the Y value", func() {
			v := gdnative.NewVector2(1.0, 0.0)
			v.SetY(100.0)
			Ω(v.GetY()).Should(BeNumerically("~", 100.0, 0.001))
		})
	})

	Context("calling Angle()", func() {
		It("should be valid", func() {
			v1 := gdnative.NewVector2(1.0, 0.0)
			v2 := gdnative.NewVector2(0.0, 1.0)
			v3 := gdnative.NewVector2(5.0, 100.0)
			Ω(v1.Angle()).Should(BeNumerically("~", 0.0, 0.001))
			Ω(v2.Angle()).Should(BeNumerically("~", 90*(math.Pi/180.0), 0.001))
			Ω(v3.IsNormalized()).Should(BeFalse())
		})
	})

	Context("calling Normalized()", func() {
		It("should be valid", func() {
			v := gdnative.NewVector2(0.5, 0.6)
			norm := v.Normalized()
			h := float32(math.Sqrt(0.5*0.5 + 0.6*0.6))
			Ω(norm.GetX()).Should(Equal(0.5 / h))
			Ω(norm.GetY()).Should(Equal(0.6 / h))
			Ω(norm.IsNormalized()).Should(BeTrue())
		})
	})

	Context("calling Tangent()", func() {
		It("should be valid", func() {
			v := gdnative.NewVector2(1.0, 0.0)
			tangent := v.Tangent()
			Ω(tangent.GetX()).Should(BeNumerically("~", 0.0, 0.001))
			Ω(tangent.GetY()).Should(BeNumerically("~", -1.0, 0.001))
		})
	})

	Context("calling OperatorMultiplyScalar()", func() {
		It("should be valid", func() {
			v := gdnative.NewVector2(1.0, 3.0)
			v2 := v.OperatorMultiplyScalar(5)
			gdstr := v2.AsString()
			str := gdstr.AsGoString()
			Ω(v2.GetX()).Should(BeNumerically("~", 5.0, 0.001), str)
			Ω(v2.GetY()).Should(BeNumerically("~", 15.0, 0.001), str)
		})
	})

	Context("calling Length()", func() {
		It("should be valid", func() {
			v := gdnative.NewVector2(2.0, 10.0)
			length := v.Length()
			Ω(length).Should(BeNumerically("~", 10.198, 0.001))
		})
	})
})
