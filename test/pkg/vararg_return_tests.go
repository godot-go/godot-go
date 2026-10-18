package pkg

import (
	"fmt"
	"runtime"
	"strings"

	. "github.com/godot-go/godot-go/pkg/builtin"
	. "github.com/godot-go/godot-go/pkg/constant"
)

// In-engine regression for issue #151 (openspec change
// fix-vararg-return-buffer).
//
// The generated vararg bindings passed the declared return type (Error,
// 8 bytes) as the r_return buffer of object_method_bind_call, while the
// engine always writes a full 24-byte Variant there. The write clobbered
// adjacent Go memory, which garbage collection turned into truncated
// strings or corrupted data in roughly 1 headless run of 15. The bindings
// now buffer a Variant and convert; these tests drive the fixed varcall
// path through the live engine.

// varargGuardPayload builds a deterministic ~1 KiB Go string. The issue's
// symptom was a returned string truncated at a zero byte landed by the
// Variant write, so the guard rejects any zero byte as well as any
// difference from the payload.
func varargGuardPayload() string {
	unit := "godot-go-issue-151-0123456789-"
	var b strings.Builder
	for b.Len() < 1024 {
		b.WriteString(unit)
	}
	return b.String()[:1024]
}

// TestVarargReturnHeapIntegrity loops generated EmitSignal varcalls
// interleaved with freshly cloned 1 KiB Go strings, forcing GC periodically:
// the reproduction shape from the issue (emit, then check a long string,
// several hundred iterations, several headless runs). Every returned string
// must stay byte-for-byte intact.
//
// Returns 1 on success, a negative code per failure mode, and -900 if the
// Go side panicked.
func (e *Example) TestVarargReturnHeapIntegrity(iterations int64) int32 {
	return objectArgGuard("vararg_return_heap", func() int32 {
		if iterations <= 0 {
			return -1
		}
		payload := varargGuardPayload()
		signal := NewStringNameWithLatin1Chars("custom_signal")
		defer signal.Destroy()

		for i := int64(0); i < iterations; i++ {
			// A generated varcall whose declared return type is Error: the
			// exact bindings that used to clobber (Object.emit_signal).
			// The arguments match the registered custom_signal signature,
			// (String, int), so the connected handler runs without a
			// call-error.
			name := NewStringWithLatin1Chars(fmt.Sprintf("issue151-%d", i))
			v0 := NewVariantString(name)
			v1 := NewVariantInt64(i)
			err := e.EmitSignal(signal, v0, v1)
			v0.Destroy()
			v1.Destroy()
			name.Destroy()
			if err != OK {
				fmt.Printf("  vararg_return_heap: emit_signal returned %d on iteration %d\n", err, i)
				return -2
			}

			// A fresh 1 KiB Go string used after the call returns.
			got := strings.Clone(payload)
			if strings.IndexByte(got, 0) >= 0 {
				fmt.Printf("  vararg_return_heap: guard string carries a zero byte on iteration %d\n", i)
				return -4
			}
			if got != payload {
				fmt.Printf("  vararg_return_heap: guard string differs from payload on iteration %d\n", i)
				return -3
			}

			if i%16 == 0 {
				runtime.GC()
			}
		}
		return 1
	})
}

// TestVarargErrorReturns asserts the converted vararg bindings hand back
// the Error value the engine wrote into the Variant return buffer.
//
// n MUST be a Node that is NOT inside the tree (the driver passes a bare
// Node.new()): Node::rpcp fails with ERR_UNCONFIGURED before it ever reads
// multiplayer state, so rpc and rpc_id return values are deterministic
// without a multiplayer context.
//
// Returns 1 on success, a negative code per failure mode, and -900 if the
// Go side panicked.
func (e *Example) TestVarargErrorReturns(n Node) int32 {
	return objectArgGuard("vararg_error_returns", func() int32 {
		if n == nil {
			return -1
		}

		// Emitting the registered custom_signal returns OK.
		signal := NewStringNameWithLatin1Chars("custom_signal")
		defer signal.Destroy()
		name := NewStringWithLatin1Chars("probe")
		defer name.Destroy()
		v0 := NewVariantString(name)
		defer v0.Destroy()
		v1 := NewVariantInt64(7)
		defer v1.Destroy()
		if err := e.EmitSignal(signal, v0, v1); err != OK {
			fmt.Printf("  vararg_error_returns: custom_signal emit returned %d, want OK\n", err)
			return -2
		}

		// A nonexistent signal returns ERR_UNAVAILABLE: Object::emit_signalp
		// rejects it with "Can't emit non-existing signal" (the engine also
		// prints that ERROR diagnostic; the run gate does not count engine
		// diagnostics as failures).
		bad := NewStringNameWithLatin1Chars("no_such_signal_issue151")
		defer bad.Destroy()
		if err := e.EmitSignal(bad); err != ERR_UNAVAILABLE {
			fmt.Printf("  vararg_error_returns: bad signal emit returned %d, want ERR_UNAVAILABLE\n", err)
			return -3
		}

		// Rpc and RpcId from an out-of-tree node return ERR_UNCONFIGURED.
		method := NewStringNameWithLatin1Chars("does_not_need_to_exist")
		defer method.Destroy()
		if err := n.Rpc(method); err != ERR_UNCONFIGURED {
			fmt.Printf("  vararg_error_returns: out-of-tree Rpc returned %d, want ERR_UNCONFIGURED\n", err)
			return -4
		}
		if err := n.RpcId(1, method); err != ERR_UNCONFIGURED {
			fmt.Printf("  vararg_error_returns: out-of-tree RpcId returned %d, want ERR_UNCONFIGURED\n", err)
			return -5
		}
		return 1
	})
}
