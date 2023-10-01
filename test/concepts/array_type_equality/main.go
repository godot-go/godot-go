package main

import (
	"fmt"
	"unsafe"
)

type TestOpaque [24]uint8

func (c *TestOpaque) ptr() unsafe.Pointer {
	return unsafe.Pointer(c)
}

func NewTestOpaque() *TestOpaque {
	ret := TestOpaque{}

	ret[0] = 1
	ret[1] = 2
	ret[2] = 4
	ret[3] = 8
	ret[4] = 16

	return &ret
}

func CreateTestOpaque() TestOpaque {
	ret := TestOpaque{}

	ret[0] = 1
	ret[1] = 2
	ret[2] = 4
	ret[3] = 8
	ret[4] = 16

	return ret
}

func TestOpaquePtr() {
	a := NewTestOpaque()
	iface := (interface{})(a)
	b := (*[24]uint8)(a)

	fmt.Printf("TestOpaquePtr\n")
	fmt.Printf("a: %p\n", a)
	fmt.Printf("a.ptr(): %p\n", a.ptr())
	fmt.Printf("iface: %p\n", iface)
	fmt.Printf("cast: %p\n", (*[24]uint8)(a))
	fmt.Printf("b: %p\n", b)
	fmt.Printf("\n")

	c := *b
	d := &c

	fmt.Printf("Copy\n")
	fmt.Printf("c: %p\n", &c)
	fmt.Printf("d: %p\n", d)
	fmt.Printf("\n")

	fmt.Printf("TestOpaquePtr\n")
	fmt.Printf("a: %p\n", a)
	fmt.Printf("cast: %p\n", b)
	fmt.Printf("\n")

	fmt.Printf("Equal\n")
	fmt.Printf("%b\n", (*[24]uint8)(a) == b)
	fmt.Printf("%b\n", *a == c)
	fmt.Printf("\n")

	fmt.Printf("Not Equal\n")
	fmt.Printf("%b\n", a == (*TestOpaque)(d))
	fmt.Printf("%b\n", b == d)
	fmt.Printf("\n")
}

func main() {
	TestOpaquePtr()
}
