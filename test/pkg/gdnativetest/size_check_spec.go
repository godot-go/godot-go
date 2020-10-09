package gdnativetest

/*
#include <cgo_example.h>
#include <gdnative.wrapper.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"unsafe"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Data width", func() {
	It("should have expectred data type width", func() {
		var (
			f64  float64 = -1.0
			f32  float32 = -1.0
			i    int32   = 127
			b    bool    = true
		)

		var (
			cd   C.double = (C.double)(f64)
			cf   C.float  = (C.float)(f32)
			ci   C.int    = (C.int)(i)
			cb   C.bool   = (C.bool)(b)
		)

		Ω(unsafe.Sizeof(cd)).Should(Equal(unsafe.Sizeof(f64)))
		Ω(unsafe.Sizeof(cf)).Should(Equal(unsafe.Sizeof(f32)))
		Ω(unsafe.Sizeof(ci)).Should(Equal(unsafe.Sizeof(i)))
		Ω(unsafe.Sizeof(cb)).Should(Equal(unsafe.Sizeof(b)))

		Ω(cf).Should(BeEquivalentTo(f32))
		Ω(cd).Should(BeEquivalentTo(f64))
		Ω(ci).Should(BeEquivalentTo(i))
		Ω(cb).Should(BeEquivalentTo(b))
	})
})
