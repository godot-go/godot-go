package gdextension

import (
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/gdnative"

	"github.com/davecgh/go-spew/spew"
	"go.uber.org/zap"
)

func zapGDNativeVariantPtrp(key string, val *GDNativeVariantPtr, len int) zap.Field {
	if val == nil {
		return zap.Reflect(key, nil)
	}

	vArgs := (*[MAX_ARG_COUNT]*Variant)(unsafe.Pointer(val))

	return zap.String(key, spew.Sdump(vArgs[:len]))
}
