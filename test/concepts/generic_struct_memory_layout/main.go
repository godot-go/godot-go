package main

import (
	"fmt"
	"unsafe"
)

type BaseStruct[T any] struct {
	opaque T
}

type GenericStruct BaseStruct[[8]uint8]

type RegularStruct struct {
	opaque [8]uint8
}

type TestEncoder[I any, T BaseStruct[I]] struct {
}

func TestGenericStruct() {
	a := GenericStruct{}
	b := RegularStruct{}

	fmt.Printf("GenericStruct\n")
	fmt.Printf("a: %p\n", a)
	fmt.Printf("&a: %p\n", &a)
	fmt.Printf("&a.opaque: %p\n", &a.opaque)
	fmt.Printf("sizeof a: %p\n", unsafe.Sizeof(a))
	fmt.Printf("sizeof b: %p\n", unsafe.Sizeof(b))
	fmt.Printf("sizeof e: %p\n", unsafe.Sizeof(TestEncoder[[8]uint8, BaseStruct[[8]uint8]]{}))
	fmt.Printf("\n")
}

func main() {
	TestGenericStruct()
}
