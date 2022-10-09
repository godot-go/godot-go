package main

//go:generate go run cmd/main.go --clangApi --extensionApi

/*
#include <stdio.h>

typedef void (*CallVirtual)(void *p_instance, int a, int b);

static void * testfunc(void *user_data) {
	printf("C.testfunc: %p %04x\n", user_data, *(int*)(user_data));
	return user_data;
}
*/
import "C"
import (
	"log"
	"runtime/cgo"
	"unsafe"
)

type TestOpaque struct {
	Opaque [24]uint8
	X int64
	Y int64
}

func (c *TestOpaque) ptr() unsafe.Pointer {
	return unsafe.Pointer(&c.Opaque)
}

func (c *TestOpaque) TestMethod(name string) {
}

func (c *TestOpaque) TestCallback(a C.int, b C.int) {
	log.Printf("TestCallback:  %d, %d, %d, %d", c.X, c.Y, a, b)
}

//export TestCallbackFunction
func TestCallbackFunction(userData unsafe.Pointer, a C.int, b C.int) {
	log.Printf("TestCallbackFunction: %p %d, %d", userData, a, b)
}

func main() {
	obj := TestOpaque{
		X: 100,
		Y: 200,
	}

	h := cgo.NewHandle(TestCallbackFunction)
	userData := unsafe.Pointer(&h)

	ret := C.testfunc(userData)

	log.Printf("input: %p %p %p\n", &obj, userData, &h)

	retHandle := (*cgo.Handle)(ret)

	log.Printf("handle value: %p %+v %p\n", retHandle, *retHandle, retHandle.Value())

	retCB := retHandle.Value().(func(obj unsafe.Pointer, a C.int, b C.int))

	retCB(unsafe.Pointer(&obj), 1, 2)

	retHandle.Delete()
}
