package gdclassimpl

import (
	"runtime"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
)

var (
	nullptr = unsafe.Pointer(nil)
	// pnr is for program-lifetime retention only: values the engine keeps a
	// pointer to after the call that registers them returns. Per-call scratch
	// must use a body-scoped runtime.Pinner with defer Unpin() instead.
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
