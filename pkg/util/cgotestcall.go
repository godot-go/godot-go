package util

// #include "cgotestcall.h"
import "C"
import "unsafe"

func CgoTestCall(data unsafe.Pointer) {
	C.cgo_testcall(data)
}
