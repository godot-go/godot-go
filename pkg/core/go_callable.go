package core

// #include <godot/gdextension_interface.h>
// #include "go_callable.h"
import "C"
import (
	"runtime"
	"runtime/cgo"
	"runtime/debug"
	"sync/atomic"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// goCallableEntry is the registry value behind a Go-backed custom Callable.
type goCallableEntry struct {
	fn func(args []Variant) (Variant, error)
	// onFree optionally runs when the engine releases the last reference —
	// the internal seam the signal-await bridge uses as its object-death
	// oracle (see signal_await.go). It must never block.
	onFree func()
}

// goCallableLive counts registered Go Callables. Incremented on creation and
// decremented in the engine's free callback, so it observes exactly the
// entries whose handles are still valid.
var goCallableLive int64

// NewGoCallable wraps a Go function as an engine-native Callable that
// GDScript can store, invoke, connect to signals, and defer with.
//
// The args slice and every Variant inside it are borrowed from the engine for
// the duration of the invocation: to keep one, copy it during the call with
// NewVariantNativeCopy (see VariantViewFromConstPtr for the aliasing rules).
//
// The returned Variant is copy-constructed into the engine's result slot,
// mirroring godot-cpp's assignment-based call semantics: the closure keeps
// ownership of whatever it returns and must release it like any other
// Go-owned Variant. Returning a zero-value Variant yields a nil result.
//
// Returning an error or panicking fails the script-side call: the failure is
// logged and reported through the engine's call-error channel; it never
// unwinds through the C boundary.
//
// The returned Callable holds one reference. Release it with Destroy when the
// Go side is done; the engine releases its own copies independently.
func NewGoCallable(fn func(args []Variant) (Variant, error)) Callable {
	return newGoCallable(fn, nil)
}

// newGoCallable is the shared constructor; onFree extends the free callback
// with a non-blocking notice that the last engine reference just died.
func newGoCallable(fn func(args []Variant) (Variant, error), onFree func()) Callable {
	if fn == nil {
		log.Panic("NewGoCallable requires a non-nil function")
	}
	h := cgo.NewHandle(&goCallableEntry{fn: fn, onFree: onFree})
	var pinner runtime.Pinner
	defer pinner.Unpin()
	info := NewGDExtensionCallableCustomInfo2(
		C.cgo_handle_to_ptr(C.uintptr_t(h)),
		unsafe.Pointer(FFI.Library),
		GDObjectInstanceID(0),
		(GDExtensionCallableCustomCall)(C.cgo_callcustom_call),
		(GDExtensionCallableCustomIsValid)(nil),
		(GDExtensionCallableCustomFree)(C.cgo_callcustom_free),
		(GDExtensionCallableCustomHash)(nil),
		(GDExtensionCallableCustomEqual)(nil),
		(GDExtensionCallableCustomLessThan)(nil),
		(GDExtensionCallableCustomToString)(C.cgo_callcustom_to_string),
	)
	pinner.Pin(&info)
	var ret Callable
	pinner.Pin(&ret)
	CallFunc_GDExtensionInterfaceCallableCustomCreate2(
		(GDExtensionUninitializedTypePtr)(unsafe.Pointer(&ret)),
		&info,
	)
	atomic.AddInt64(&goCallableLive, 1)
	return ret
}

// GoCallback_GoCallableCall is invoked by the engine when script calls a Go
// Callable.
//
//export GoCallback_GoCallableCall
func GoCallback_GoCallableCall(
	callableUserData unsafe.Pointer,
	argPtrs *C.GDExtensionConstVariantPtr,
	argumentCount C.GDExtensionInt,
	rReturn C.GDExtensionVariantPtr,
	rError *C.GDExtensionCallError,
) {
	// Panic barrier around the user closure only: a closure panic is
	// caller-induced and must never unwind through cgo (process abort), so it
	// is recovered and reported as a failed call. Panics from our own
	// bookkeeping keep their fatality, matching the varcall callback's
	// internal-fault policy.
	invoking := false
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if !invoking {
			panic(r)
		}
		writeNilCallReturn(rReturn)
		rError.error = C.GDEXTENSION_CALL_ERROR_INVALID_METHOD
		rError.argument = -1
		rError.expected = -1
		log.Error("go callable panicked",
			zap.Any("panic", r),
			zap.ByteString("stack", debug.Stack()),
		)
	}()
	ud := (cgo.Handle)(callableUserData)
	entry, ok := ud.Value().(*goCallableEntry)
	// A dead or wrong-typed handle means our own free/registration bookkeeping
	// is broken; like the varcall callback's null-handle fault this is not a
	// caller-inducible condition and stays fatal.
	if !ok || entry == nil || entry.fn == nil {
		log.Panic("unable to retrieve go callable userdata")
	}

	// These are non-owning bit-copy views of the engine's borrowed arguments.
	// Nothing is allocated here, so no path below may call Destroy on them
	// (that would free engine-owned memory).
	argPtrSlice := unsafe.Slice((*GDExtensionConstVariantPtr)(argPtrs), int(argumentCount))
	args := make([]Variant, argumentCount)
	for i := range argPtrSlice {
		args[i] = VariantViewFromConstPtr(argPtrSlice[i])
	}

	// invoking is cleared only on normal return: if the closure panics it
	// stays set, so the barrier above attributes the panic to the user.
	invoking = true
	ret, err := entry.fn(args)
	invoking = false
	if err != nil {
		writeNilCallReturn(rReturn)
		rError.error = C.GDEXTENSION_CALL_ERROR_INVALID_METHOD
		rError.argument = -1
		rError.expected = -1
		log.Debug("go callable failed", zap.Error(err))
		return
	}
	if ret == (Variant{}) {
		// A zero-value Variant is unconstructed storage; leave a well-formed
		// nil instead of copy-constructing from it.
		writeNilCallReturn(rReturn)
	} else {
		// Assignment semantics (godot-cpp parity): the engine receives its own
		// reference, so passing a borrowed argument straight back is safe. The
		// closure keeps ownership of ret.
		var pinner runtime.Pinner
		defer pinner.Unpin()
		dest := (GDExtensionUninitializedVariantPtr)(rReturn)
		pinner.Pin(dest)
		CallFunc_GDExtensionInterfaceVariantNewCopy(dest, ret.NativeConstPtr())
	}
	setCallErrorOK(rError)
}

// GoCallableRegistryCount reports the number of live Go-backed custom
// Callable registry entries. It drops exactly when the engine invokes the
// free callback, so tests can observe deterministic release.
func GoCallableRegistryCount() int64 {
	return atomic.LoadInt64(&goCallableLive)
}

// GoCallback_GoCallableFree is invoked by the engine when the last reference
// to a Go Callable is destroyed. The handle dies here, which is the sole
// legitimate end of the registry entry's lifetime.
//
//export GoCallback_GoCallableFree
func GoCallback_GoCallableFree(callableUserData unsafe.Pointer) {
	ud := (cgo.Handle)(callableUserData)
	// Run the optional death notice while the handle still resolves; the
	// notice contract is non-blocking (await-state close only).
	if entry, ok := ud.Value().(*goCallableEntry); ok && entry != nil && entry.onFree != nil {
		entry.onFree()
	}
	ud.Delete()
	atomic.AddInt64(&goCallableLive, -1)
}

// GoCallback_GoCallableToString is invoked by the engine to render the
// callable, e.g. from GDScript str(cb).
//
//export GoCallback_GoCallableToString
func GoCallback_GoCallableToString(
	callableUserData unsafe.Pointer,
	rIsValid *C.GDExtensionBool,
	rOut C.GDExtensionStringPtr,
) {
	ud := (cgo.Handle)(callableUserData)
	if _, ok := ud.Value().(*goCallableEntry); !ok {
		// Nothing to render; leave the engine fallback in place.
		*rIsValid = 0
		return
	}
	GDExtensionStringPtrWithUtf8Chars((GDExtensionStringPtr)(rOut), "GoCallable")
	// Write the flag *through* the out pointer, exactly as the class
	// to_string callback learned: rebinding this function's local copy leaves
	// the engine's bool false and discards the string we constructed.
	*rIsValid = C.GDExtensionBool(1)
}
