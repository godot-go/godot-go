package main

/*
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"fmt"
	"runtime"
	"strconv"
	"unsafe"

	"github.com/CannibalVox/cgoalloc"
)

type TestOpaque struct {
	A int
	B int
	C int
}

func TestPointerAdd(mem cgoalloc.Allocator) {
	fmt.Printf("\nTestPointerAdd\n")

	/* output:
	sizeof TestOpaque: 24
	offset TestOpaque.A: 0
	offset TestOpaque.B: 8
	offset TestOpaque.C: 16
	*/
	fmt.Printf("sizeof TestOpaque: %d\n", unsafe.Sizeof(TestOpaque{}))
	fmt.Printf("offset TestOpaque.A: %d\n", unsafe.Offsetof(TestOpaque{}.A))
	fmt.Printf("offset TestOpaque.B: %d\n", unsafe.Offsetof(TestOpaque{}.B))
	fmt.Printf("offset TestOpaque.C: %d\n", unsafe.Offsetof(TestOpaque{}.C))

	len := 5
	ptr := mem.Malloc(int(unsafe.Sizeof(TestOpaque{}) * uintptr(len)))
	arr := (*[5]TestOpaque)(ptr)

	for i := 0; i < 5; i++ {
		arr[i].A = i
		arr[i].B = i * 10
		arr[i].C = i * 100
	}

	/* output:
	0: &{0 0 0}
	1: &{1 10 100}
	2: &{2 20 200}
	3: &{3 30 300}
	4: &{4 40 400}
	*/
	for i := 0; i < 5; i++ {
		obj := (*TestOpaque)(unsafe.Add(ptr, int(unsafe.Sizeof(TestOpaque{}))*i))
		fmt.Printf("%d: %v\n", i, obj)
	}

	mem.Free(ptr)
}

func TestStackAndHeap(mem cgoalloc.Allocator) {
	fmt.Printf("\nTestMemoryAllocation\n")

	cStackVar1 := int(88)
	cStackVar2 := int(88)
	cStackVar3 := int(88)
	cHeapVar1 := (*int)(mem.Malloc(10))
	cHeapVar2 := (*int)(mem.Malloc(10))
	cHeapVar3 := (*int)(mem.Malloc(10))
	goStackVar1 := strconv.Itoa(88)
	goStackVar2 := strconv.Itoa(88)
	goStackVar3 := strconv.Itoa(88)

	fmt.Printf("cStackVar1: %p\n", &cStackVar1)
	fmt.Printf("cStackVar2: %p\n", &cStackVar2)
	fmt.Printf("cStackVar3: %p\n", &cStackVar3)
	fmt.Printf("cHeapVar1: %p\n", cHeapVar1)
	fmt.Printf("cHeapVar2: %p\n", cHeapVar2)
	fmt.Printf("cHeapVar3: %p\n", cHeapVar3)
	fmt.Printf("goStackVar1: %p\n", &goStackVar1)
	fmt.Printf("goStackVar2: %p\n", &goStackVar2)
	fmt.Printf("goStackVar3: %p\n", &goStackVar3)

	mem.Free(unsafe.Pointer(cHeapVar1))
	mem.Free(unsafe.Pointer(cHeapVar2))
	mem.Free(unsafe.Pointer(cHeapVar3))
}

func TestAllocCopy(mem cgoalloc.Allocator) {
	fmt.Printf("\nTestAllocCopy\n")

	fmt.Printf("sizeof pointer: %d\n", unsafe.Sizeof(uintptr(0)))
	fmt.Printf("sizeof C.char: %d\n", C.sizeof_char)

	slice := []*TestOpaque{
		{1, 2, 3},
		{40, 50, 60},
		{700, 800, 900},
	}

	// this is required to prevent go panic because of AllocCopy through cgo:
	// panic: runtime error: cgo argument has Go pointer to unpinned Go pointer
	pinner := runtime.Pinner{}
	pinner.Pin(slice[0])
	pinner.Pin(slice[1])
	pinner.Pin(slice[2])
	defer pinner.Unpin()

	fmt.Printf("slice: %v\n", slice)
	fmt.Printf("slice.len: %v\n", len(slice))
	fmt.Printf("slice.cap: %v\n", cap(slice))
	fmt.Printf("slice[0]: %p %+v\n", &slice[0], slice[0])
	fmt.Printf("slice[1]: %p %+v\n", &slice[1], slice[1])
	fmt.Printf("slice[2]: %p %+v\n", &slice[2], slice[2])

	fmt.Printf("header: %p\n", &slice)
	fmt.Printf("header.Data: %p\n", unsafe.SliceData(slice))
	fmt.Printf("header.Len: %d\n", len(slice))
	fmt.Printf("header.Cap: %d\n", cap(slice))

	copiedData := AllocCopy(mem, unsafe.Pointer(unsafe.SliceData(slice)), int(unsafe.Sizeof(uintptr(0)))*len(slice))

	fmt.Printf("copiedData: %+v\n", unsafe.Pointer(copiedData))
	fmt.Printf("copiedData[0]: %+v\n", *(**TestOpaque)(unsafe.Add(copiedData, int(unsafe.Sizeof(uintptr(0)))*0)))
	fmt.Printf("copiedData[1]: %+v\n", *(**TestOpaque)(unsafe.Add(copiedData, int(unsafe.Sizeof(uintptr(0)))*1)))
	fmt.Printf("copiedData[2]: %+v \n", *(**TestOpaque)(unsafe.Add(copiedData, int(unsafe.Sizeof(uintptr(0)))*2)))
}

func AllocCopy(mem cgoalloc.Allocator, src unsafe.Pointer, bytes int) unsafe.Pointer {
	m := mem.Malloc(bytes)

	C.memcpy(m, src, C.size_t(bytes))

	return m
}

func main() {
	mem := &cgoalloc.DefaultAllocator{}

	TestPointerAdd(mem)

	TestStackAndHeap(mem)

	TestAllocCopy(mem)
}
