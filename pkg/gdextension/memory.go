package gdextension

/*
#include <godot/gdnative_interface.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"math"
	"reflect"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdnative"
)

// AllocCopy returns a duplicated data allocated in C memory.
//
// NOTE: Memory allocated in C is NOT managed by Go GC; therefore, gdnative#Free
// must be called on the pointer to release the memory back to the OS.
func AllocCopy(src unsafe.Pointer, bytes uint64) unsafe.Pointer {
	m := GDNativeInterface_mem_alloc(internal.gdnInterface, bytes)

	C.memcpy(m, src, C.size_t(bytes))

	return m
}

// AllocZeros returns zeroed out bytes allocated in C memory.
//
// NOTE: Memory allocated in C is NOT managed by Go GC; therefore, gdnative#Free
// must be called on the pointer to release the memory back to the OS.
func AllocZeros(bytes uint64) unsafe.Pointer {
	m := GDNativeInterface_mem_alloc(internal.gdnInterface, bytes)

	C.memset(m, 0, C.size_t(bytes))

	return m
}

func AllocArray[T any](len uint64) *[math.MaxInt32]T {
	var t T

	bytes := (uint64)(unsafe.Sizeof(t)) * len

	return (*[math.MaxInt32]T)(AllocZeros(bytes))
}

func Alloc(bytes uint64) unsafe.Pointer {
	return GDNativeInterface_mem_alloc(internal.gdnInterface, bytes)
}

func Realloc(p_memory unsafe.Pointer, bytes uint64) unsafe.Pointer {
	return GDNativeInterface_mem_realloc(internal.gdnInterface, p_memory, bytes)
}

func Free(p_ptr unsafe.Pointer) {
	GDNativeInterface_mem_free(internal.gdnInterface, p_ptr)
}

var sizeOfPtr = unsafe.Sizeof(GDNativeTypePtr(uintptr(0)))

func AllocGDNativeTypePtrArray(size int) *[MAX_ARG_COUNT]GDNativeTypePtr {
	m := AllocZeros(uint64(uintptr(size) * sizeOfPtr))

	return (*[MAX_ARG_COUNT]GDNativeTypePtr)(m)
}

// AllocNewArrayAsUnsafePointer returns a C array of *Variant copy allocated in
// C memory.
//
// NOTE: Memory allocated in C is NOT managed by Go GC; therefore, gdnative#Free
// must be called on the pointer to release the memory back to the OS.
// TODO: investigate whether this should be an array of []*Variant.opaque or not.
func AllocNewArrayAsUnsafePointer(p_args []*Variant) unsafe.Pointer {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p_args))

	size := uint64(sizeOfPtr * uintptr(len(p_args)))

	return AllocCopy(unsafe.Pointer(header.Data), size)
}

func VariantPtrSliceToGDNativeVariantPtr(p_args []*Variant) *GDNativeVariantPtr {
	gdnVariantPtrArgs := AllocNewArrayAsUnsafePointer(p_args)

	for i := range p_args {
		ai := (*unsafe.Pointer)(unsafe.Add(gdnVariantPtrArgs, uintptr(i)*sizeOfPtr))
		*ai = unsafe.Pointer(p_args[i].ptr())
	}

	return (*GDNativeVariantPtr)(gdnVariantPtrArgs)
}

func VariantPtrSliceToGDNativeVariantPtrArray(p_args []*Variant) *[MAX_ARG_COUNT]GDNativeVariantPtr {
	gdnVariantPtrArgs := (*[MAX_ARG_COUNT]GDNativeVariantPtr)(AllocNewArrayAsUnsafePointer(p_args))

	for i := range p_args {
		gdnVariantPtrArgs[i] = (GDNativeVariantPtr)(unsafe.Pointer(p_args[i].ptr()))
	}

	return gdnVariantPtrArgs
}
