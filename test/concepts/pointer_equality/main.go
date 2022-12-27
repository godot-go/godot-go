package main

import (
	"fmt"
	"unsafe"
)

type TestOpaque struct {
	opaque [24]uint8
}

func (c *TestOpaque) ptr() unsafe.Pointer {
	return unsafe.Pointer(&c.opaque)
}

func NewTestOpaque() *TestOpaque {
	ret := TestOpaque{}

	ret.opaque[5] = 1

	return &ret
}

func TestOpaquePtr() {
	a := NewTestOpaque()
	iface := (interface{})(a)

	fmt.Printf("TestOpaquePtr\n")
	fmt.Printf("a: %p\n", a)
	fmt.Printf("a.ptr(): %p\n", a.ptr())
	fmt.Printf("iface: %p\n", iface)
	fmt.Printf("\n")
}

func main() {
	TestOpaquePtr()
}
