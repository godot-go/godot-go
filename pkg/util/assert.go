package util

import (
	"unsafe"
)

var (
	deadptr32 = unsafe.Pointer(uintptr(0xdeaddead))
	deadptr64 = unsafe.Pointer(uintptr(0xdeaddeaddeaddead))
)

func IsDeadMemory(v unsafe.Pointer) bool {
	return v == deadptr32 || v == deadptr64
}