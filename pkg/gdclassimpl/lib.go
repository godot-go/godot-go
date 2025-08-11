package gdclassimpl

import (
	"runtime"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

var (
	nullptr = unsafe.Pointer(nil)
	pnr runtime.Pinner
)

func (cx *ObjectImpl) ToGoString() string {
	if cx == nil || cx.Owner == nil {
		return ""
	}
	gdstr := cx.ToString()
	defer gdstr.Destroy()
	return gdstr.ToUtf8()
}

func GetInputSingleton() Input {
	owner := (*GodotObject)(unsafe.Pointer(GetSingleton("Input")))
	return NewInputWithGodotOwnerObject(owner)
}
