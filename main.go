package main

//go:generate go run cmd/main.go --clangApi --extensionApi

/*
typedef void (*CallVirtual)(void *p_instance, int a, int b);


static int testfunc(void *user_data) {
	CallVirtual cb = (CallVirtual)(user_data);
	cb(NULL, 1, 2);
	return 123;
}
*/
import "C"
import (
	"fmt"
	"log"
	"runtime/cgo"
	"unsafe"
)

type TestOpaque struct {
	opaque [24]uint8
}

func (c *TestOpaque) ptr() unsafe.Pointer {
	return unsafe.Pointer(&c.opaque)
}

func (c *TestOpaque) TestMethod(name string) {
}

func (c *TestOpaque) TestCallback(a C.int, b C.int) {
	log.Printf("TestCallback: %d, %d", a, b)
}

//export TestCallbackFunction
func TestCallbackFunction(c unsafe.Pointer, a C.int, b C.int) {
	log.Printf("TestCallback: %d, %d", a, b)
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

	fmt.Printf("a: %p\n", a)
	fmt.Printf("a.ptr(): %p\n", a.ptr())
	fmt.Printf("iface: %p\n", iface)
}

func main() {
	userData := unsafe.Pointer(cgo.NewHandle(TestCallbackFunction))

	a := C.testfunc(userData)

	log.Printf("%p %d\n", &userData, a)
}
