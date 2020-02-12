package main

//go:generate go env
//go:generate go run cmd/main.go --clangApi --extensionApi

// #include <stdio.h>
// #include <stdlib.h>
// #include <string.h>
import "C"
import (
	"fmt"
	"math"
	"unsafe"

	"github.com/davecgh/go-spew/spew"
)

type TestOpaque struct {
	opaque [24]uint8
}

func (c *TestOpaque) ptr() unsafe.Pointer {
	return unsafe.Pointer(&c.opaque)
}

func (c *TestOpaque) TestMethod(name string) {
}

func NewTestOpaque() *TestOpaque {
	ret := TestOpaque{}

	ret.opaque[5] = 1

	return &ret
}

func NewTestOpaquePtr() *TestOpaque {
	ret := TestOpaque{}

	ret.opaque[5] = 123

	return &ret
}

func AllocZeros(bytes uint64) unsafe.Pointer {
	m := C.malloc((C.ulong)(bytes))

	C.memset(m, 0, C.size_t(bytes))

	return m
}

func Free(p_ptr unsafe.Pointer) {
	C.free(p_ptr)
}

func AllocArray[T any](len uint64) (*[math.MaxUint32]T, *T) {
	var t T

	bytes := (uint64)(unsafe.Sizeof(t)) * len

	ptr := ((*T)(AllocZeros(bytes)))

	ret := (*[math.MaxUint32]T)(unsafe.Pointer(ptr))

	return ret, ptr
}

func TestAllocArray() {
	a := NewTestOpaque()

	spew.Dump((uint64)(unsafe.Sizeof(a)))

	str := "hello"

	size := len(str)

	cstr, ptr := AllocArray[C.char]((uint64)(size + 1))

	spew.Dump(cstr[:size+1])

	cstr[0] = 'H'
	cstr[1] = 'E'
	cstr[2] = 'L'
	cstr[3] = 'L'
	cstr[4] = 'O'
	cstr[5] = '\000'

	shrinked := cstr[:size+1]

	spew.Dump(shrinked)

	spew.Dump(&cstr[0], ptr)

	Free(unsafe.Pointer(ptr))
}

func TestOpaquePtr() {
	a := NewTestOpaque()
	iface := (interface{})(a)

	fmt.Printf("%p\n", a)
	fmt.Printf("%p\n", a.ptr())
	fmt.Printf("%p\n", iface)
}

type Example struct {
	TestOpaque
	InternalValue int8
}

func TestExamplePtr() {
	a := &Example{}
	iface := (interface{})(a)

	fmt.Printf("%p\n", a)
	fmt.Printf("%p\n", a.ptr())
	fmt.Printf("%p\n", iface)
}

func main() {
	TestExamplePtr()
}
