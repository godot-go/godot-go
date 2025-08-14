//go:build linux

package util

import (
	"fmt"
	"runtime"
	"syscall"
	"unsafe"
)

func SetThreadName(name string) {
	if runtime.GOOS == "linux" {
		_, _, errno := syscall.Syscall6(syscall.SYS_PRCTL, syscall.PR_SET_NAME, uintptr(unsafe.Pointer(syscall.StringBytePtr(name))), 0, 0, 0, 0)
		if errno != 0 {
			fmt.Printf("Error setting thread name: %v\n", errno)
		}
	}
}
