package main

//go:generate go run cmd/main.go --clangApi --extensionApi

// #include <stdio.h>
// #include <stdlib.h>
// #include <string.h>
import "C"
import (
	"fmt"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdnative"
	. "github.com/godot-go/godot-go/pkg/gdextension"
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

	var (
		ptr GDNativeTypePtr
		obj ObjectImpl
	)

	println(ptr)
	println(obj.GetClassName())
}
