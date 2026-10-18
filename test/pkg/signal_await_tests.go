package pkg

import (
	"encoding/binary"
	"math"
	"reflect"
	"time"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/core"
)

// In-engine fixtures for the signal await bridge (openspec change
// support-signal-await). Convention follows go_callable_tests.go: GDScript
// arms a scenario, lets frames pass, then reads a verdict code where 1 is
// success and each negative value names one failed assertion — a consumer
// goroutine records the verdict because GDScript can observe return values
// but not Go captures.

// Verdict codes for the emit-after-await scenario.
const (
	awaitHitPending    int64 = 0  // consumer has not resolved yet
	awaitHitNeverFires int64 = -1 // channel closed without a send
	awaitHitArgCount   int64 = -2 // payload arity wrong
	awaitHitArgPenalty int64 = -3 // first arg != int64(4)
	awaitHitArgKind    int64 = -4 // second arg != "critical"
	awaitHitNotClosed  int64 = -5 // channel still open after delivery
)

// awaitHitCode/awaitHitDone form a write-then-close handoff: the consumer
// goroutine writes the verdict before closing done, so a probe that observes
// the close reads a published value. The probe waits instead of the scenario
// burning --quit-after frames.
var (
	awaitHitCode int64
	awaitHitDone chan struct{}
)

// AwaitHitStart arms core.Await on the target's "hit" signal and consumes the
// result on its own goroutine, recording the verdict for the emission
// [4, "critical"]: both arguments in order with Go-native types and the
// channel closed after the delivery.
func (e *Example) AwaitHitStart(target Object) int64 {
	awaitHitDone = make(chan struct{})
	ch := Await(target, "hit")
	go func() {
		res, ok := <-ch
		switch {
		case !ok:
			awaitHitCode = awaitHitNeverFires
		case len(res.Args) != 2:
			awaitHitCode = awaitHitArgCount
		case res.Args[0] != int64(4):
			awaitHitCode = awaitHitArgPenalty
		case res.Args[1] != "critical":
			awaitHitCode = awaitHitArgKind
		default:
			if _, alive := <-ch; alive {
				awaitHitCode = awaitHitNotClosed
			} else {
				awaitHitCode = 1
			}
		}
		close(awaitHitDone)
	}()
	return 1
}

// AwaitHitCode waits (bounded) for AwaitHitStart's consumer verdict. The
// verdict is pure Go data; blocking here stalls only the headless test
// driver, never the correctness of the observation.
func (e *Example) AwaitHitCode() int64 {
	return awaitWaitCode(awaitHitDone, &awaitHitCode)
}

// awaitWaitCode joins a scenario consumer with a wall-clock guard: a verdict
// that never lands reports pending (0) rather than hanging the engine.
func awaitWaitCode(done chan struct{}, code *int64) int64 {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-done:
		return *code
	case <-timer.C:
		return awaitWaitTimedOut
	}
}

// awaitWaitTimedOut is the distinct verdict for a consumer that never
// resolved, so a hang cannot read as a scenario's own failure code.
const awaitWaitTimedOut int64 = -999

// Verdict codes for the marshal-table scenario: one negative code per payload
// assertion so a failure names the exact conversion that broke.
const (
	awaitMarshalPending    int64 = 0
	awaitMarshalNeverFires int64 = -1
	awaitMarshalArgCount   int64 = -2
	awaitMarshalBool       int64 = -20
	awaitMarshalInt        int64 = -21
	awaitMarshalFloat      int64 = -22
	awaitMarshalString     int64 = -23
	awaitMarshalStringName int64 = -24
	awaitMarshalNodePath   int64 = -25
	awaitMarshalVector2    int64 = -26
	awaitMarshalObject     int64 = -27
	awaitMarshalArray      int64 = -28
	awaitMarshalDictionary int64 = -29
	awaitMarshalOpaque     int64 = -30
	awaitMarshalNil        int64 = -31
	awaitMarshalVector3    int64 = -40
	awaitMarshalColor      int64 = -41
	awaitMarshalVector2I   int64 = -42
	awaitMarshalTransform2 int64 = -43
	awaitMarshalBasis      int64 = -44
)

var (
	awaitMarshalCode int64
	awaitMarshalDone chan struct{}
)

// AwaitMarshalStart arms core.Await on the target's "marshalled" signal and
// records the verdict comparing all twelve payload positions against
// independently-derived expectations. want is the object identity expected at
// position 7 (the same Go-decoded handle GDScript passes as argument).
func (e *Example) AwaitMarshalStart(target Object, want Object) int64 {
	awaitMarshalDone = make(chan struct{})
	ch := Await(target, "marshalled")
	go func() {
		awaitMarshalCode = awaitMarshalEvaluate(ch, want)
		close(awaitMarshalDone)
	}()
	return 1
}

func awaitMarshalEvaluate(ch <-chan AwaitResult, want Object) int64 {
	res, ok := <-ch
	if !ok {
		return awaitMarshalNeverFires
	}
	if len(res.Args) != 17 {
		return awaitMarshalArgCount
	}
	if b, isB := res.Args[0].(bool); !isB || !b {
		return awaitMarshalBool
	}
	if n, isN := res.Args[1].(int64); !isN || n != 7 {
		return awaitMarshalInt
	}
	if f, isF := res.Args[2].(float64); !isF || f != 2.5 {
		return awaitMarshalFloat
	}
	if s, isS := res.Args[3].(string); !isS || s != "text" {
		return awaitMarshalString
	}
	if s, isS := res.Args[4].(string); !isS || s != "sn_name" {
		return awaitMarshalStringName
	}
	if s, isS := res.Args[5].(string); !isS || s != "Node/sub" {
		return awaitMarshalNodePath
	}
	// Hand-built IEEE-754 little-endian layout for Vector2(1.5, -2.25),
	// independent of any production conversion code.
	var wantV2 Vector2
	binary.LittleEndian.PutUint32(wantV2[0:4], math.Float32bits(1.5))
	binary.LittleEndian.PutUint32(wantV2[4:8], math.Float32bits(-2.25))
	if v2, isV := res.Args[6].(Vector2); !isV || v2 != wantV2 {
		return awaitMarshalVector2
	}
	if res.Args[7] != want {
		return awaitMarshalObject
	}
	if !reflect.DeepEqual(res.Args[8], []any{int64(1), []any{int64(2), "deep"}}) {
		return awaitMarshalArray
	}
	if !reflect.DeepEqual(res.Args[9], map[any]any{"k": int64(1), "nest": map[any]any{"m": true}}) {
		return awaitMarshalDictionary
	}
	if op, isO := res.Args[10].(AwaitOpaque); !isO || op.TypeName == "" {
		return awaitMarshalOpaque
	}
	if res.Args[11] != nil {
		return awaitMarshalNil
	}
	// Inline POD beyond Vector2: Vector3(1,2,3), Color(.25,.5,.75,1),
	// Vector2i(7,-9) — hand-derived little-endian float32/int32 layouts.
	if v3, isV := res.Args[12].(Vector3); !isV || v3 != pod3(1, 2, 3) {
		return awaitMarshalVector3
	}
	if c, isC := res.Args[13].(Color); !isC || c != pod4(0.25, 0.5, 0.75, 1) {
		return awaitMarshalColor
	}
	if v2i, isV := res.Args[14].(Vector2i); !isV || v2i != pod2i(7, -9) {
		return awaitMarshalVector2I
	}
	// Heap-stored POD variants: Transform2D(0, Vector2(10,20)) and identity
	// Basis must arrive as copied values, proving the slot was dereferenced
	// rather than bit-cloned (a cloned dangling pointer would corrupt).
	if t2, isT := res.Args[15].(Transform2D); !isT || t2 != podTransform2(1, 0, 0, 1, 10, 20) {
		return awaitMarshalTransform2
	}
	if b, isB := res.Args[16].(Basis); !isB || b != podBasis(1, 0, 0, 0, 1, 0, 0, 0, 1) {
		return awaitMarshalBasis
	}
	return 1
}

// podN build expected POD payloads from hand-checked literals, independent of
// any production conversion code.
func pod3(a, b, c float32) Vector3 {
	var out Vector3
	putF32(out[:], a, b, c)
	return out
}

func pod4(a, b, c, d float32) Color {
	var out Color
	putF32(out[:], a, b, c, d)
	return out
}

func pod2i(a, b int32) Vector2i {
	var out Vector2i
	binary.LittleEndian.PutUint32(out[0:4], uint32(a))
	binary.LittleEndian.PutUint32(out[4:8], uint32(b))
	return out
}

func podTransform2(a, b, c, d, e, f float32) Transform2D {
	var out Transform2D
	putF32(out[:], a, b, c, d, e, f)
	return out
}

func podBasis(e00, e01, e02, e10, e11, e12, e20, e21, e22 float32) Basis {
	var out Basis
	putF32(out[:], e00, e01, e02, e10, e11, e12, e20, e21, e22)
	return out
}

func putF32(dst []byte, vals ...float32) {
	for i, v := range vals {
		binary.LittleEndian.PutUint32(dst[i*4:], math.Float32bits(v))
	}
}

// AwaitMarshalCode waits (bounded) for AwaitMarshalStart's consumer verdict.
func (e *Example) AwaitMarshalCode() int64 {
	return awaitWaitCode(awaitMarshalDone, &awaitMarshalCode)
}

// Verdict codes for the one-shot lifecycle scenario: the registry baseline
// proves the engine dropped the connection after the first emission, so the
// probes are "payload arrived", "registry drained", "re-arm delivers again".
const (
	awaitOneShotNeverFires  int64 = -1
	awaitOneShotBadArgs     int64 = -2
	awaitOneShotNotDrained  int64 = -3
	awaitOneShotReappear    int64 = -4
	awaitOneShotRearmMissed int64 = -5
)

var (
	awaitOneShotBaseline  int64
	awaitOneShotCode      int64
	awaitOneShotDone      chan struct{}
	awaitOneShotRearm     int64
	awaitOneShotRearmDone chan struct{}
)

// AwaitOneShotStart records the registry baseline, arms core.Await on the
// target's "ping" signal, and consumes the single expected emission
// [1] on a goroutine.
func (e *Example) AwaitOneShotStart(target Object) int64 {
	awaitOneShotBaseline = GoCallableRegistryCount()
	awaitOneShotDone = make(chan struct{})
	ch := Await(target, "ping")
	go func() {
		res, ok := <-ch
		switch {
		case !ok:
			awaitOneShotCode = awaitOneShotNeverFires
		case len(res.Args) != 1 || res.Args[0] != int64(1):
			awaitOneShotCode = awaitOneShotBadArgs
		default:
			awaitOneShotCode = 1
		}
		close(awaitOneShotDone)
	}()
	return 1
}

// AwaitOneShotCode joins the first consumer and asserts the connection died
// with the emission: the registry must be back at its baseline, which only
// holds if the engine disconnected (and freed) the one-shot callable.
func (e *Example) AwaitOneShotCode() int64 {
	code := awaitWaitCode(awaitOneShotDone, &awaitOneShotCode)
	if code != 1 {
		return code
	}
	if GoCallableRegistryCount() != awaitOneShotBaseline {
		return awaitOneShotNotDrained
	}
	return 1
}

// AwaitOneShotSilenceCode re-checks the registry after an emission nobody
// awaited: the freed connection must not have resurrected a callable.
func (e *Example) AwaitOneShotSilenceCode() int64 {
	if GoCallableRegistryCount() != awaitOneShotBaseline {
		return awaitOneShotReappear
	}
	return 1
}

// AwaitOneShotRearm arms a fresh Await for the next emission [3], proving a
// signal can be awaited again after its one-shot subscription completed.
func (e *Example) AwaitOneShotRearm(target Object) int64 {
	awaitOneShotRearmDone = make(chan struct{})
	ch := Await(target, "ping")
	go func() {
		res, ok := <-ch
		if ok && len(res.Args) == 1 && res.Args[0] == int64(3) {
			awaitOneShotRearm = 1
		} else {
			awaitOneShotRearm = awaitOneShotRearmMissed
		}
		close(awaitOneShotRearmDone)
	}()
	return 1
}

// AwaitOneShotRearmCode joins the re-armed consumer and asserts its emission
// arrived and drained the registry again.
func (e *Example) AwaitOneShotRearmCode() int64 {
	code := awaitWaitCode(awaitOneShotRearmDone, &awaitOneShotRearm)
	if code != 1 {
		return code
	}
	if GoCallableRegistryCount() != awaitOneShotBaseline {
		return awaitOneShotNotDrained
	}
	return 1
}

// AwaitMisuseRecovered calls Await on a signal the object does not define and
// reports whether the documented panic was raised: 1 only if recover() saw a
// panic at the call site.
func (e *Example) AwaitMisuseRecovered(target Object) int64 {
	recovered := false
	func() {
		defer func() { recovered = recover() != nil }()
		Await(target, "no_such_signal_at_all")
	}()
	if recovered {
		return 1
	}
	return -1
}

// Verdict codes for the freed-object liveness scenario.
const (
	awaitFreedDelivered  int64 = -1 // got a send although the object died first
	awaitFreedNotDrained int64 = -3 // callable outlived the freed connection
)

var (
	awaitFreedBaseline int64
	awaitFreedCode     int64
	awaitFreedDone     chan struct{}
)

// AwaitFreedStart arms core.Await on a signal that is never emitted and
// consumes the channel on a goroutine: the subscription must end via the
// callable free callback when GDScript drops the target.
func (e *Example) AwaitFreedStart(target Object) int64 {
	awaitFreedBaseline = GoCallableRegistryCount()
	awaitFreedDone = make(chan struct{})
	ch := Await(target, "never")
	go func() {
		_, ok := <-ch
		if ok {
			awaitFreedCode = awaitFreedDelivered
		} else {
			awaitFreedCode = 1
		}
		close(awaitFreedDone)
	}()
	return 1
}

// AwaitFreedCode joins the consumer and asserts the freed object closed the
// channel without a send and drained its callable.
func (e *Example) AwaitFreedCode() int64 {
	code := awaitWaitCode(awaitFreedDone, &awaitFreedCode)
	if code != 1 {
		return code
	}
	if GoCallableRegistryCount() != awaitFreedBaseline {
		return awaitFreedNotDrained
	}
	return 1
}

// AwaitShutdownArm leaves a subscription on a signal nothing will emit so
// teardown must drain it: the suite gate (TESTS FINISHED, zero failures, no
// ObjectDB leaks) judges the outcome, no probe reads this one back.
func (e *Example) AwaitShutdownArm(target Object) int64 {
	ch := Await(target, "never")
	go func() {
		<-ch // unblocks via teardown dissolve or process exit; never hangs shutdown
	}()
	return 1
}

// Verdict codes for the SceneTreeTimer scenario.
const (
	awaitTimerNeverFires int64 = -1
	awaitTimerArgs       int64 = -2 // timeout emitted arguments (it has none)
	awaitTimerNotClosed  int64 = -3
)

var (
	awaitTimerCode int64
	awaitTimerDone chan struct{}
)

// AwaitTimerStart arms core.Await on a SceneTreeTimer's "timeout" signal,
// the godot-cpp-world equivalently of GDScript's `await timer.timeout`:
// zero-argument emission must deliver an empty Args and close the channel.
func (e *Example) AwaitTimerStart(timer Object) int64 {
	awaitTimerDone = make(chan struct{})
	ch := Await(timer, "timeout")
	go func() {
		res, ok := <-ch
		switch {
		case !ok:
			awaitTimerCode = awaitTimerNeverFires
		case len(res.Args) != 0:
			awaitTimerCode = awaitTimerArgs
		default:
			if _, alive := <-ch; alive {
				awaitTimerCode = awaitTimerNotClosed
			} else {
				awaitTimerCode = 1
			}
		}
		close(awaitTimerDone)
	}()
	return 1
}

// AwaitTimerCode joins the timer consumer's verdict.
func (e *Example) AwaitTimerCode() int64 {
	return awaitWaitCode(awaitTimerDone, &awaitTimerCode)
}
