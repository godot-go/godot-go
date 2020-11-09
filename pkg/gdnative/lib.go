package gdnative

/*
#cgo CFLAGS: -DX86=1 -g -fPIC -std=c99 -I${SRCDIR}/../../godot_headers -I${SRCDIR}/../../pkg/gdnative
#include <cgo_gateway_register_class.h>
#include <cgo_gateway_class.h>
#include <nativescript.wrapper.gen.h>
#include <gdnative_api_struct.gen.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	"reflect"
	"sort"
	"unsafe"

	"github.com/godot-go/godot-go/pkg/log"
)

type RegisterStateStruct struct {
	NativescriptHandle unsafe.Pointer
	LanguageIndex      C.int
	TagDB              tagDB
	Stats              RuntimeStats
}

type RuntimeStats struct {
	InitCount          int
	NativeScriptAllocs map[string]int
	NativeScriptFrees  map[string]int
	GodotTypeAllocs    map[string]int
	GodotTypeFrees     map[string]int
}

// LogMemoeryLeak will log memory leak messages if there's an imbalance
// of allocs and frees per entity.
func (r RuntimeStats) LogObjectLeak() {
	checkAllocsAndFrees := func(category string, allocMap map[string]int, freeMap map[string]int) {
		if len(allocMap) != len(freeMap) {
			log.Error(
				"alloc and free maps not aligned",
				StringField("category", category),
			)
		}

		// keys is a sorted set of keys in the GodotTypeAllocs map
		keys := make([]string, 0, len(allocMap))
		for k := range allocMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			allocs := allocMap[k]
			frees := freeMap[k]
			if allocs > frees {
				log.Warn(
					"memory leak detected",
					StringField("category", category),
					StringField("type", k),
					AnyField("allocs", allocs),
					AnyField("frees", frees),
				)
			}
		}
	}

	checkAllocsAndFrees("godot_type", r.GodotTypeAllocs, r.GodotTypeFrees)
	checkAllocsAndFrees("nativescript", r.NativeScriptAllocs, r.NativeScriptFrees)
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

// AllocZeros returns zeroed out bytes allocated in C memory.
//
// NOTE: Memory allocated in C is NOT managed by Go GC; therefore, gdnative#Free
// must be called on the pointer to release the memory back to the OS.
func AllocZeros(p_bytes int32) unsafe.Pointer {
	m := Alloc(p_bytes)

	C.memset(m, 0, C.size_t(p_bytes))

	return m
}

// AllocCopy returns a duplicated data allocated in C memory.
//
// NOTE: Memory allocated in C is NOT managed by Go GC; therefore, gdnative#Free
// must be called on the pointer to release the memory back to the OS.
func AllocCopy(src unsafe.Pointer, p_bytes int32) unsafe.Pointer {
	m := Alloc(p_bytes)

	C.memcpy(m, src, C.size_t(p_bytes))

	return m
}

// AllocNewSlice returns a new slice allocated in C memory at the specified
// size and a pointer to the C memory. Please do not attempt to resize the
// slice.
//
// NOTE: Memory allocated in C is NOT managed by Go GC; therefore, gdnative#Free
// must be called on the pointer to release the memory back to the OS.
func AllocNewSlice(size int) ([]unsafe.Pointer, unsafe.Pointer) {
	ptrCArr := AllocZeros(int32(unsafe.Sizeof(uintptr(0))) * int32(size))

	var arr []unsafe.Pointer
	header := (*reflect.SliceHeader)(unsafe.Pointer(&arr))
	header.Data = uintptr(ptrCArr)
	header.Len = size
	header.Cap = size

	for i := range arr {
		if uintptr(arr[i]) != uintptr(0) {
			log.Panic("value should be a zero uintptr", StringField("method", "AllocNewSlice"), AnyField("i", i))
		}
	}

	return arr, ptrCArr
}

var sizeOfVariantPtr = unsafe.Sizeof(&Variant{})

// AllocNewArrayAsUnsafePointer returns a C array of *Variant copy allocated in
// C memory.
//
// NOTE: Memory allocated in C is NOT managed by Go GC; therefore, gdnative#Free
// must be called on the pointer to release the memory back to the OS.
func AllocNewArrayAsUnsafePointer(p_args []*Variant) unsafe.Pointer {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p_args))

	size := int32(sizeOfVariantPtr * uintptr(len(p_args)))

	return AllocCopy(unsafe.Pointer(header.Data), size)
}

// ArrayRefFromPtrSlice returns an unsafe.Pointer to specified slice's SliceHeader.Data.
//
// NOTE: Memory allocated in C is NOT managed by Go GC; therefore, gdnative#Free
// must be called on the pointer to release the memory back to the OS.
func ArrayRefFromPtrSlice(p_args []unsafe.Pointer) unsafe.Pointer {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p_args))

	return unsafe.Pointer(header.Data)
}

// AllocNewVariantPtrSlice returns a C array of *Variant allocated in
// C memory. Please do not attempt to resize the slice.
func AllocNewVariantPtrSlice(size int) ([]*Variant, unsafe.Pointer) {
	ptrCArr := Alloc(int32(unsafe.Sizeof(uintptr(0))) * int32(size))

	var arr []*Variant
	header := (*reflect.SliceHeader)(unsafe.Pointer(&arr))
	header.Data = uintptr(ptrCArr)
	header.Len = size
	header.Cap = size

	return arr, ptrCArr
}

// WrapUnsafePointerAsSlice returns a slice at the specified size wrapping ref.
// Please do not attempt to resize the slice.
//
// NOTE: Memory allocated in C is NOT managed by Go GC; therefore, gdnative#Free
// must be called on the pointer to release the memory back to the OS.
func WrapUnsafePointerAsSlice(size int, ref unsafe.Pointer) []unsafe.Pointer {
	var arr []unsafe.Pointer
	header := (*reflect.SliceHeader)(unsafe.Pointer(&arr))
	header.Data = uintptr(ref)
	header.Len = size
	header.Cap = size

	return arr
}

// CastVariantPtrSliceToUnsafePointerSlice casts []*Variant into []unsafe.Pointer.
//
// NOTE: Memory allocated in C is NOT managed by Go GC; therefore, gdnative#Free
// must be called on the pointer to release the memory back to the OS.
func CastVariantPtrSliceToUnsafePointerSlice(variants []*Variant) []unsafe.Pointer {
	return *(*[]unsafe.Pointer)(unsafe.Pointer(&variants))
}

// AddrAsString return the memory addres of the GodotObject
func (o GodotObject) AddrAsString() string {
	return fmt.Sprintf("%p", &o)
}
