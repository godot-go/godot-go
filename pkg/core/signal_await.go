package core

import (
	"reflect"
	"strconv"
	"sync"
	"unsafe"

	. "github.com/godot-go/godot-go/pkg/builtin"
	"github.com/godot-go/godot-go/pkg/constant"
	. "github.com/godot-go/godot-go/pkg/ffi"
	"github.com/godot-go/godot-go/pkg/log"
	"go.uber.org/zap"
)

// AwaitResult is the payload delivered by a signal Await channel: the emitted
// arguments, marshalled to Go-native values (see marshalAwaitPayload).
//
// A receive on the channel reports ok == true for a delivered emission —
// possibly with an empty Args list for a zero-argument signal — and
// ok == false only when the channel was closed without a send because the
// awaited object was freed and the signal can never fire.
type AwaitResult struct {
	Args []any
}

// AwaitOpaque wraps a payload value that has no Go-native representation
// (exotic built-ins such as Callable and Signal, and containers past the
// recursion budget). The type name and engine string form are preserved for
// diagnostics; the value itself is dropped.
type AwaitOpaque struct {
	TypeName string
	Text     string
}

// awaitState is the engine-free core of one Await subscription: a buffered
// cap-1 result channel plus the terminal-state latch.
//
// Exactly one terminal path wins: deliver() (signal emitted) or dissolve()
// (connection died with the emission never happening, e.g. the object was
// freed). The latch makes the loser a no-op, so a delivery followed by the
// engine's free callback — which routinely run back to back — can never
// double-close, and the buffered channel guarantees neither path blocks the
// emitting thread even when the awaiting goroutine has not reached the
// receive yet.
type awaitState struct {
	results   chan AwaitResult
	closeOnce sync.Once
}

func newAwaitState() *awaitState {
	return &awaitState{results: make(chan AwaitResult, 1)}
}

// deliver publishes the emitted arguments and closes the channel. It never
// blocks and is safe to call from any thread.
func (s *awaitState) deliver(args []any) {
	s.closeOnce.Do(func() {
		s.results <- AwaitResult{Args: args}
		close(s.results)
	})
}

// dissolve closes the channel without sending: the awaited signal can never
// fire. It never blocks and is safe to call from any thread.
func (s *awaitState) dissolve() {
	s.closeOnce.Do(func() {
		close(s.results)
	})
}

// Await subscribes a one-shot connection to the named signal on obj and
// returns a channel delivering the first emission as an AwaitResult. It is
// the Go equivalent of GDScript's `await obj.signal_name`: the awaiting
// goroutine parks on the channel receive and costs nothing until the signal
// fires. Await itself must be called from a thread that may call into the
// engine (the main thread); receiving from the returned channel is legal
// from any goroutine.
//
// # Terminal semantics
//
// Exactly one of two outcomes closes the channel:
//
//   - The signal emits: one AwaitResult arrives (Args holds the emission
//     arguments in declaration order, possibly empty), then the channel is
//     closed. Use `res := <-ch` when only delivery matters.
//
//   - The subscription dies before any emission — the object is freed, or
//     engine teardown drops the connection: the channel closes without a
//     send. The receive then reports ok == false, so a pending await can
//     never hang past the lifetime of what it waits for:
//
//     res, ok := <-ch
//     if !ok {
//     // the object was freed; the signal can never fire
//     }
//
// # Composition idioms
//
// Because Await returns an ordinary channel it composes with the whole
// standard toolkit — no engine knowledge at the call site:
//
//	// timeout
//	select {
//	case res := <-core.Await(enemy, "defeated"):
//	    handle(res)
//	case <-time.After(2 * time.Second):
//	    timeout()
//	}
//
//	// cancellation with context
//	ch := core.Await(boss, "defeated")
//	select {
//	case res := <-ch:
//	    handle(res)
//	case <-ctx.Done():
//	}
//
//	// racing several signals
//	select {
//	case res := <-core.Await(a, "hit"):  ...
//	case res := <-core.Await(b, "hit"):  ...
//	}
//
// Each Await arms one emission; to keep observing a signal, call Await
// again after each delivery (or keep a long-lived connection via a Go
// closure Callable from NewGoCallable).
//
// # Payload types
//
// Args values are Go-owned and safe to carry to any goroutine: nil, bool,
// int64, float64, string (from String/StringName/NodePath), math POD structs
// (Vector2, Color, Transform3D, ...), Object handles for engine objects, and
// recursive []any / map[any]any for Array/Dictionary payloads (bounded
// nesting; deeper values degrade to AwaitOpaque). Values with no Go form
// (Callable, Signal, ...) arrive as AwaitOpaque carrying the type name and
// engine text for diagnostics.
//
// # Threading hazards
//
// The channel hands over pure Go data — but anything reachable from the
// engine stays engine-adjacent. Two rules keep the boundary safe:
//
//   - Do not call engine methods from the woken goroutine. The await result
//     is a snapshot; hop back to the main thread (or re-enter through a
//     mechanism that does) before touching Godot objects again. Object
//     handles inside Args are references to live engine objects, not thread
//     safe views on them.
//   - Do not store raw Variants across the await boundary — the emission's
//     argument views die when the callback returns, and Variant refcounts
//     are not thread safe. Await never hands any out for exactly this
//     reason.
//
// # Misuse
//
// Calling Await with a signal name the object does not declare, or against
// a Connect the engine rejects, raises an immediate panic at the call site
// naming class and signal (no dangling subscription is created) instead of
// returning a channel that can never resolve.
func Await(obj Object, signalName string) <-chan AwaitResult {
	if obj == nil {
		log.Panic("Await requires a non-nil object")
	}
	state := newAwaitState()
	callable := newGoCallable(
		func(args []Variant) (Variant, error) {
			state.deliver(marshalAwaitArgs(args))
			return Variant{}, nil
		},
		state.dissolve,
	)
	sn := NewStringNameWithLatin1Chars(signalName)
	defer sn.Destroy()
	if err := obj.Connect(sn, callable, uint32(constant.OBJECT_CONNECT_FLAGS_CONNECT_ONE_SHOT)); err != constant.OK {
		callable.Destroy()
		cls := obj.GetClass()
		className := cls.ToUtf8()
		cls.Destroy()
		log.Panic("await connect failed",
			zap.String("class", className),
			zap.String("signal", signalName),
			zap.Int("error", int(err)),
		)
	}
	// Load-bearing release (design D2): after this the one-shot connection
	// holds the only reference, making the free callback the deterministic
	// end-of-subscription oracle.
	callable.Destroy()
	return state.results
}

// marshalAwaitArgs converts borrowed emission arguments into Go-owned
// values. It runs on the emitting thread while the views are alive; nothing
// it returns references engine memory — refcounted payloads are converted
// (strings, recursive containers) or reduced to descriptors (AwaitOpaque),
// and plain-data POD payloads are copied out of the Variant (dereferencing
// the engine's heap slot for the oversized types).
func marshalAwaitArgs(args []Variant) []any {
	out := make([]any, len(args))
	for i := range args {
		out[i] = marshalAwaitValue(args[i], 0)
	}
	return out
}

// awaitMarshalDepthLimit bounds container recursion so self-referential
// Array/Dictionary payloads terminate.
const awaitMarshalDepthLimit = 64

// awaitPayloadOffset is where a Variant's payload union begins after the
// type tag and alignment padding of the engine's 24-byte layout.
const awaitPayloadOffset = 8

func marshalAwaitValue(v Variant, depth int) any {
	switch v.GetType() {
	case GDEXTENSION_VARIANT_TYPE_NIL:
		return nil
	case GDEXTENSION_VARIANT_TYPE_BOOL:
		return v.ToBool()
	case GDEXTENSION_VARIANT_TYPE_INT:
		return v.ToInt64()
	case GDEXTENSION_VARIANT_TYPE_FLOAT:
		return v.ToFloat64()
	case GDEXTENSION_VARIANT_TYPE_STRING:
		return v.ToGoString()
	case GDEXTENSION_VARIANT_TYPE_STRING_NAME, GDEXTENSION_VARIANT_TYPE_NODE_PATH:
		return v.Stringify()
	case GDEXTENSION_VARIANT_TYPE_OBJECT:
		if awaitSlotPointer(v) == nil {
			return nil
		}
		return v.ToObject()
	case GDEXTENSION_VARIANT_TYPE_VECTOR2:
		return awaitInlinePOD[Vector2](v)
	case GDEXTENSION_VARIANT_TYPE_VECTOR2I:
		return awaitInlinePOD[Vector2i](v)
	case GDEXTENSION_VARIANT_TYPE_RECT2:
		return awaitInlinePOD[Rect2](v)
	case GDEXTENSION_VARIANT_TYPE_RECT2I:
		return awaitInlinePOD[Rect2i](v)
	case GDEXTENSION_VARIANT_TYPE_VECTOR3:
		return awaitInlinePOD[Vector3](v)
	case GDEXTENSION_VARIANT_TYPE_VECTOR3I:
		return awaitInlinePOD[Vector3i](v)
	case GDEXTENSION_VARIANT_TYPE_VECTOR4:
		return awaitInlinePOD[Vector4](v)
	case GDEXTENSION_VARIANT_TYPE_VECTOR4I:
		return awaitInlinePOD[Vector4i](v)
	case GDEXTENSION_VARIANT_TYPE_PLANE:
		return awaitInlinePOD[Plane](v)
	case GDEXTENSION_VARIANT_TYPE_QUATERNION:
		return awaitInlinePOD[Quaternion](v)
	case GDEXTENSION_VARIANT_TYPE_COLOR:
		return awaitInlinePOD[Color](v)
	case GDEXTENSION_VARIANT_TYPE_RID:
		return awaitInlinePOD[RID](v)
	case GDEXTENSION_VARIANT_TYPE_TRANSFORM2D:
		return awaitHeapPOD[Transform2D](v)
	case GDEXTENSION_VARIANT_TYPE_AABB:
		return awaitHeapPOD[AABB](v)
	case GDEXTENSION_VARIANT_TYPE_BASIS:
		return awaitHeapPOD[Basis](v)
	case GDEXTENSION_VARIANT_TYPE_TRANSFORM3D:
		return awaitHeapPOD[Transform3D](v)
	case GDEXTENSION_VARIANT_TYPE_PROJECTION:
		return awaitHeapPOD[Projection](v)
	case GDEXTENSION_VARIANT_TYPE_ARRAY:
		return awaitMarshalArray(v, depth)
	case GDEXTENSION_VARIANT_TYPE_DICTIONARY:
		return awaitMarshalDictionary(v, depth)
	default:
		return AwaitOpaque{TypeName: awaitTypeName(v.GetType()), Text: v.Stringify()}
	}
}

// awaitInlinePOD bit-copies a Variant's inline payload into the typed Go
// value (all inline POD types share the engine struct layout).
func awaitInlinePOD[T any](v Variant) T {
	var out T
	copy(awaitRawBytes(unsafe.Pointer(&out), unsafe.Sizeof(out)), v[awaitPayloadOffset:])
	return out
}

// awaitHeapPOD copies an oversized POD payload: the engine stores
// Transform2D/AABB/Basis/Transform3D/Projection as a heap pointer in the
// Variant slot, so a bit-copy of the slot would hand out a dangling pointer.
// The pointee is read while the emission keeps it alive.
func awaitHeapPOD[T any](v Variant) T {
	var out T
	if slot := awaitSlotPointer(v); slot != nil {
		copy(awaitRawBytes(unsafe.Pointer(&out), unsafe.Sizeof(out)), awaitRawBytes(slot, unsafe.Sizeof(out)))
	}
	return out
}

// awaitSlotPointer reads the 8-byte union slot as a raw pointer by byte
// copy, avoiding any uintptr-to-Pointer conversion vet flags as misuse.
func awaitSlotPointer(v Variant) unsafe.Pointer {
	var slot unsafe.Pointer
	copy(awaitRawBytes(unsafe.Pointer(&slot), 8), v[awaitPayloadOffset:16])
	return slot
}

func awaitRawBytes(p unsafe.Pointer, n uintptr) []byte {
	return unsafe.Slice((*byte)(p), n)
}

func awaitMarshalArray(v Variant, depth int) any {
	if depth >= awaitMarshalDepthLimit {
		return AwaitOpaque{TypeName: "Array", Text: "depth limit exceeded"}
	}
	sizeV, err := v.Call("size", nil)
	if err != nil {
		return AwaitOpaque{TypeName: "Array", Text: err.Error()}
	}
	size := int(sizeV.ToInt64())
	sizeV.Destroy()
	out := make([]any, 0, min(size, 64))
	for i := 0; i < size; i++ {
		el, err := v.GetIndexed(i)
		if err != nil {
			break
		}
		out = append(out, marshalAwaitValue(el, depth+1))
		el.Destroy()
	}
	return out
}

func awaitMarshalDictionary(v Variant, depth int) any {
	if depth >= awaitMarshalDepthLimit {
		return AwaitOpaque{TypeName: "Dictionary", Text: "depth limit exceeded"}
	}
	keysV, err := v.Call("keys", nil)
	if err != nil {
		return AwaitOpaque{TypeName: "Dictionary", Text: err.Error()}
	}
	defer keysV.Destroy()
	sizeV, err := keysV.Call("size", nil)
	if err != nil {
		return AwaitOpaque{TypeName: "Dictionary", Text: err.Error()}
	}
	size := int(sizeV.ToInt64())
	sizeV.Destroy()
	out := make(map[any]any, min(size, 64))
	for i := 0; i < size; i++ {
		keyTmp, err := keysV.GetIndexed(i)
		if err != nil {
			break
		}
		valTmp, err := v.GetKeyed(keyTmp)
		key := awaitMarshalKey(keyTmp, depth)
		keyTmp.Destroy()
		if err != nil {
			continue
		}
		out[key] = marshalAwaitValue(valTmp, depth+1)
		valTmp.Destroy()
	}
	return out
}

// awaitMarshalKey keeps map keys comparable: GDScript permits container
// keys, whose marshalled forms (slices/maps) cannot key a Go map, so they
// degrade to their string form.
func awaitMarshalKey(kv Variant, depth int) any {
	m := marshalAwaitValue(kv, depth)
	if m == nil {
		return nil
	}
	if t := reflect.TypeOf(m); t != nil && !t.Comparable() {
		return kv.Stringify()
	}
	return m
}

// awaitTypeName names opaque payloads for diagnostics.
func awaitTypeName(t GDExtensionVariantType) string {
	switch t {
	case GDEXTENSION_VARIANT_TYPE_CALLABLE:
		return "Callable"
	case GDEXTENSION_VARIANT_TYPE_SIGNAL:
		return "Signal"
	default:
		return "variant_type_" + strconv.Itoa(int(t))
	}
}
