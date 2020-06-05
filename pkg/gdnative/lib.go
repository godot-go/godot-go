package gdnative

/*
#include <cgo_gateway_register_class.h>
#include <cgo_gateway_class.h>
#include <nativescript.gen.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	"github.com/pcting/godot-go/pkg/log"
	"reflect"
	"unsafe"
)

type RegisterStateStruct struct {
	NativescriptHandle unsafe.Pointer
	LanguageIndex      C.int
	TagDB              tagDB
	InitCount          int
}

var (
	CoreApi           *C.godot_gdnative_core_api_struct
	Core11Api         *C.godot_gdnative_core_1_1_api_struct
	Core12Api         *C.godot_gdnative_core_1_2_api_struct
	GDNativeLibObject *C.godot_object
	NativescriptApi   *C.godot_gdnative_ext_nativescript_api_struct
	Nativescript11Api *C.godot_gdnative_ext_nativescript_1_1_api_struct
	PluginscriptApi   *C.godot_gdnative_ext_pluginscript_api_struct
	AndroidApi        *C.godot_gdnative_ext_android_api_struct
	ARVRApi           *C.godot_gdnative_ext_arvr_api_struct
	VideodecoderApi   *C.godot_gdnative_ext_videodecoder_api_struct
	NetApi            *C.godot_gdnative_ext_net_api_struct
	Net32Api          *C.godot_gdnative_ext_net_3_2_api_struct
	RegisterState     RegisterStateStruct
)

var (
	emptyVariantPtr   = &Variant{}
	emptyUnsafePtr    = unsafe.Pointer(&emptyVariantPtr)
	emptyUnsafePtrPtr = unsafe.Pointer(&emptyUnsafePtr)
)

func AllocZeros(p_bytes int32) unsafe.Pointer {
	m := Alloc(p_bytes)

	C.memset(m, 0, C.ulong(p_bytes))

	return m
}

func AllocCopy(src unsafe.Pointer, p_bytes int32) unsafe.Pointer {
	m := Alloc(p_bytes)

	C.memcpy(m, src, C.ulong(p_bytes))

	return m
}

func NewSliceFromAlloc(size int) ([]unsafe.Pointer, unsafe.Pointer) {
	ptrCArr := AllocZeros(int32(unsafe.Sizeof(uintptr(0))) * int32(size))

	var arr []unsafe.Pointer
	header := (*reflect.SliceHeader)(unsafe.Pointer(&arr))
	header.Data = uintptr(ptrCArr)
	header.Len = size
	header.Cap = size

	for i := range arr {
		if uintptr(arr[i]) != uintptr(0) {
			log.WithField("method", "NewSliceFromAlloc").WithField("i", fmt.Sprintf("%d", i)).Panic("value should be a zero uintptr")
		}
	}

	return arr, ptrCArr
}

var sizeOfVariantPtr = unsafe.Sizeof(emptyVariantPtr)

func CArrayFromVariantPtrSlice(p_args []*Variant) unsafe.Pointer {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p_args))

	size := int32(sizeOfVariantPtr * uintptr(len(p_args)))

	return AllocCopy(unsafe.Pointer(header.Data), size)
}

func CArrayRefFromPtrSlice(p_args []unsafe.Pointer) unsafe.Pointer {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p_args))

	return unsafe.Pointer(header.Data)
}

func NewVariantPtrSliceFromAlloc(size int) ([]*Variant, unsafe.Pointer) {
	ptrCArr := Alloc(int32(unsafe.Sizeof(uintptr(0))) * int32(size))

	var arr []*Variant
	header := (*reflect.SliceHeader)(unsafe.Pointer(&arr))
	header.Data = uintptr(ptrCArr)
	header.Len = size
	header.Cap = size

	return arr, ptrCArr
}

func NewSliceFromCPtrPtrRef(size int, ref unsafe.Pointer) []unsafe.Pointer {
	var arr []unsafe.Pointer
	header := (*reflect.SliceHeader)(unsafe.Pointer(&arr))
	header.Data = uintptr(ref)
	header.Len = size
	header.Cap = size

	return arr
}

func NewUnsafePointerSliceFromVariantSlice(variants []*Variant) []unsafe.Pointer {
	return *(*[]unsafe.Pointer)(unsafe.Pointer(&variants))
}

func (o GodotObject) AddrAsString() string {
	return fmt.Sprintf("%p", &o)
}
