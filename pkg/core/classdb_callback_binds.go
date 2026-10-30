package core

// Typed binds for the engine class-callback virtuals (typed-method-dispatch
// design D6). The classdb_callback.go dispatch sites carry fixed shapes
// decided by the GDExtension ABI: notification, property get/set, the two
// revert queries, and the to_string conversion. These binds attach a named
// fixed-shape closure to the method metadata at registration, so those
// dispatch sites invoke the handler through a compiled closure — receiver
// resolution is the same selfTyped assertion the ptrcall wrappers use, and
// no reflect.Value frames appear per call. Handlers registered through the
// name-string API keep the reflective Func.Call fallback at those sites,
// unchanged.
//
// Every bind runs the shared typed-bind validation core (classDBBuildTyped
// Metadata): the qualified V_ resolution against the embedding chain,
// duplicate-name rejection, and the argument-name gates fire exactly as
// they do for the generic arity overloads. The engine method registration
// (classDBRegisterMethod) is likewise shared, so GDScript can still call the
// virtual explicitly; where the handler signature fits the slot constraint
// unions (_notification, _property_can_revert) the explicit-call paths ride
// the compiled ptrcall/varcall closures too, and where it does not (_get,
// _set, _property_get_revert, to_string) those calls fall back reflectively
// while the engine callback slot itself still uses the compiled closure.

import (
	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
)

// The engine-reserved virtual names these binds own. The dispatch sites in
// classdb_callback.go / to_string.go look the virtuals up under exactly
// these strings, so the binds fix them instead of accepting a parameter a
// caller could mistype into an unreachable registration.
const (
	gdVirtualNotification      = "_notification"
	gdVirtualGet               = "_get"
	gdVirtualSet               = "_set"
	gdVirtualPropertyCanRevert = "_property_can_revert"
	gdVirtualPropertyGetRevert = "_property_get_revert"
)

// newNotificationShape closes a typed notification handler over the
// delivery shape of GoCallback_ClassCreationInfoNotification.
func newNotificationShape[TImpl any, S selfTyped[TImpl]](fn func(*TImpl, int32, bool)) notificationFn {
	return func(inst GDClass, what int32, reversed bool) {
		fn(inst.(S).Self(), what, reversed)
	}
}

// newPropGetShape closes a typed `_get` handler over the delivery shape of
// GoCallback_ClassCreationInfoGet (UTF-8 property name in).
func newPropGetShape[TImpl any, S selfTyped[TImpl]](fn func(*TImpl, string) (Variant, bool)) propGetFn {
	return func(inst GDClass, name string) (Variant, bool) {
		return fn(inst.(S).Self(), name)
	}
}

// newPropSetShape closes a typed `_set` handler over the delivery shape of
// GoCallback_ClassCreationInfoSet (UTF-8 name plus engine-owned Variant
// view in, consumed-flag out).
func newPropSetShape[TImpl any, S selfTyped[TImpl]](fn func(*TImpl, string, Variant) bool) propSetFn {
	return func(inst GDClass, name string, value Variant) bool {
		return fn(inst.(S).Self(), name, value)
	}
}

// newPropCanRevertShape closes a typed `_property_can_revert` handler over
// its dispatch shape (Godot-owned StringName value in).
func newPropCanRevertShape[TImpl any, S selfTyped[TImpl]](fn func(*TImpl, StringName) bool) propCanRevertFn {
	return func(inst GDClass, name StringName) bool {
		return fn(inst.(S).Self(), name)
	}
}

// newPropGetRevertShape closes a typed `_property_get_revert` handler over
// its dispatch shape (StringName in, revert value plus has-revert flag out).
func newPropGetRevertShape[TImpl any, S selfTyped[TImpl]](fn func(*TImpl, StringName) (Variant, bool)) propGetRevertFn {
	return func(inst GDClass, name StringName) (Variant, bool) {
		return fn(inst.(S).Self(), name)
	}
}

// newToStringShape closes a typed to_string handler over the delivery shape
// resolveToStringFor uses (receiver in, plain Go string out).
func newToStringShape[TImpl any, S selfTyped[TImpl]](fn func(*TImpl) string) toStringFn {
	return func(inst GDClass) string {
		return fn(inst.(S).Self())
	}
}

// ClassDBBindVirtualNotification binds the `_notification` virtual with a
// compiled callback closure. The signature (int32 what, bool reversed) is
// the engine slot's own, so explicit calls to the registered method ride
// the compiled ptrcall and varcall closures as well.
func ClassDBBindVirtualNotification[TImpl any, S selfTyped[TImpl]](t S, fn func(*TImpl, int32, bool)) {
	md := classDBBuildTypedMetadata(t, fn, gdVirtualNotification, METHOD_FLAG_VIRTUAL,
		[]string{"what", "reversed"}, nil,
		wrapPtrcallNone2[TImpl, S, int32, bool](fn),
		wrapVarcallNone2[TImpl, S, int32, bool](fn, nil, gdVirtualNotification))
	md.cbNotification = newNotificationShape[TImpl, S](fn)
	classDBRegisterMethod(md, md.GoMethodName)
}

// ClassDBBindVirtualGet binds the `_get` property virtual. The callback
// closure serves the engine's property-fetch slot; explicit method calls
// fall back reflectively (Go string and Variant returns stay outside the
// slot constraint unions).
func ClassDBBindVirtualGet[TImpl any, S selfTyped[TImpl]](t S, fn func(*TImpl, string) (Variant, bool)) {
	md := classDBBuildTypedMetadata(t, fn, gdVirtualGet, METHOD_FLAG_VIRTUAL,
		[]string{"name"}, nil, nil, nil)
	md.cbPropGet = newPropGetShape[TImpl, S](fn)
	classDBRegisterMethod(md, md.GoMethodName)
}

// ClassDBBindVirtualSet binds the `_set` property virtual.
func ClassDBBindVirtualSet[TImpl any, S selfTyped[TImpl]](t S, fn func(*TImpl, string, Variant) bool) {
	md := classDBBuildTypedMetadata(t, fn, gdVirtualSet, METHOD_FLAG_VIRTUAL,
		[]string{"name", "value"}, nil, nil, nil)
	md.cbPropSet = newPropSetShape[TImpl, S](fn)
	classDBRegisterMethod(md, md.GoMethodName)
}

// ClassDBBindVirtualPropertyCanRevert binds `_property_can_revert`. The
// StringName-in/bool-out signature fits the slot unions, so explicit calls
// ride the compiled closures too.
func ClassDBBindVirtualPropertyCanRevert[TImpl any, S selfTyped[TImpl]](t S, fn func(*TImpl, StringName) bool) {
	md := classDBBuildTypedMetadata(t, fn, gdVirtualPropertyCanRevert, METHOD_FLAG_VIRTUAL,
		[]string{"name"}, nil,
		wrapPtrcallValue1[TImpl, S, StringName, bool](fn),
		wrapVarcallValue1[TImpl, S, StringName, bool](fn, nil, gdVirtualPropertyCanRevert))
	md.cbPropCanRevert = newPropCanRevertShape[TImpl, S](fn)
	classDBRegisterMethod(md, md.GoMethodName)
}

// ClassDBBindVirtualPropertyGetRevert binds `_property_get_revert`.
func ClassDBBindVirtualPropertyGetRevert[TImpl any, S selfTyped[TImpl]](t S, fn func(*TImpl, StringName) (Variant, bool)) {
	md := classDBBuildTypedMetadata(t, fn, gdVirtualPropertyGetRevert, METHOD_FLAG_VIRTUAL,
		[]string{"name"}, nil, nil, nil)
	md.cbPropGetRevert = newPropGetRevertShape[TImpl, S](fn)
	classDBRegisterMethod(md, md.GoMethodName)
}

// ClassDBBindVirtualToString binds the `to_string` conversion virtual. The
// compiled closure answers resolveToStringFor directly; the reflective
// dispatch's runtime guard that the virtual "returns a single Go string"
// becomes a compile-time property of this bind.
func ClassDBBindVirtualToString[TImpl any, S selfTyped[TImpl]](t S, fn func(*TImpl) string) {
	md := classDBBuildTypedMetadata(t, fn, toStringGdName, METHOD_FLAG_VIRTUAL,
		nil, nil, nil, nil)
	md.cbToString = newToStringShape[TImpl, S](fn)
	classDBRegisterMethod(md, md.GoMethodName)
}
