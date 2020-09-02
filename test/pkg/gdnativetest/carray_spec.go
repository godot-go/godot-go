package gdnativetest

/*
#include <cgo_example.h>
#include <gdnative.wrapper.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"github.com/godot-go/godot-go/pkg/gdnative"
	"unsafe"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func testcall(api unsafe.Pointer, pArgs []unsafe.Pointer, pRet unsafe.Pointer) {
	C.cgo_example_struct_from_p_args(
		(*C.godot_gdnative_core_api_struct)(api),
		*(**unsafe.Pointer)(unsafe.Pointer(&pArgs)),
		pRet,
	);
}

var _ = Describe("Array Helpers", func() {
	When("interoperating with C", func() {
		It("should round trip primatives", func() {
			var (
				name string  = "test_name"
				f    float64 = -1.0
				i    int32   = 127
				b    bool    = true
			)
		
			var (
				cn   *C.char = C.CString(name)
				cf   C.float = (C.float)(f)
				ci   C.int   = (C.int)(i)
				cb   C.bool  = (C.bool)(b)
			)
		
			defer C.free(unsafe.Pointer(cn))

			api := (*C.godot_gdnative_core_api_struct)(unsafe.Pointer(gdnative.CoreApi))
		
			v := C.cgo_example_struct(api, cn, cf, ci, cb)
		
			ret := (*C.example_struct)(unsafe.Pointer(&v))

			gs := (*gdnative.String)(unsafe.Pointer(&ret.name))
			defer gs.Destroy()

			Ω(gs.AsGoString()).Should(Equal(name))
			Ω(ret.f).Should(BeEquivalentTo(f))
			Ω(ret.i).Should(BeEquivalentTo(i))
			Ω(ret.b).Should(BeEquivalentTo(b))
		})
	})
	When("NewSliceFromAlloc", func() {
		It("should round-trip values in a []unsafe.Pointer", func() {
			ptrCArrSize := 5
			ptrArguments, ptrCArr := gdnative.NewSliceFromAlloc(ptrCArrSize)

			defer gdnative.Free(unsafe.Pointer(ptrCArr))

			var (
				name string  = "test_name"
				f    float64 = -1.0
				i    int32   = 127
				b    bool    = true
			)

			var (
				cn   *C.char = C.CString(name)
				cf   C.float = (C.float)(f)
				ci   C.int   = (C.int)(i)
				cb   C.bool  = (C.bool)(b)
			)

			defer C.free(unsafe.Pointer(cn))

			ptrArguments[0] = unsafe.Pointer(&cn)
			ptrArguments[1] = unsafe.Pointer(&cf)
			ptrArguments[2] = unsafe.Pointer(&ci)
			ptrArguments[3] = unsafe.Pointer(&cb)
			ptrArguments[4] = unsafe.Pointer(uintptr(0))

			var ret C.example_struct

			testcall(unsafe.Pointer(gdnative.CoreApi), ptrArguments, unsafe.Pointer(&ret))

			gs := (*gdnative.String)(unsafe.Pointer(&ret.name))
			defer gs.Destroy()

			Ω(gs.AsGoString()).Should(Equal(name))
			Ω(ret.f).Should(BeEquivalentTo(f))
			Ω(ret.i).Should(BeEquivalentTo(i))
			Ω(ret.b).Should(BeEquivalentTo(b))
		})
	})
	When("CArrayFromPtrSlice", func() {
		It("should return a copy of the data as an array", func() {
			var (
				pa = (*int32)(gdnative.AllocZeros(int32(unsafe.Sizeof(uintptr(0)))))
				pb = (*int32)(gdnative.AllocZeros(int32(unsafe.Sizeof(uintptr(0)))))
				pc = (*int32)(gdnative.AllocZeros(int32(unsafe.Sizeof(uintptr(0)))))
			)

			defer gdnative.Free(unsafe.Pointer(pa))
			defer gdnative.Free(unsafe.Pointer(pb))
			defer gdnative.Free(unsafe.Pointer(pc))

			*pa = 10
			*pb = 20
			*pc = 30

			slice := []unsafe.Pointer{
				unsafe.Pointer(pa),
				unsafe.Pointer(pb),
				unsafe.Pointer(pc),
			}

			data := (*[3]*int32)(gdnative.CArrayRefFromPtrSlice(slice))

			Ω(*data[0]).Should(BeEquivalentTo(10))
			Ω(*data[1]).Should(BeEquivalentTo(20))
			Ω(*data[2]).Should(BeEquivalentTo(30))


		})
		It("should copy pointer array", func() {
			
		})
	})
})
