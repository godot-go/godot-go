# Proposal

## Why

Vararg method binds in the generated class wrappers pass the declared return type as the return buffer. For `Object.emit_signal`, `Node.rpc`, and `Node.rpc_id` the declared type is `Error` (8 bytes), but Godot writes a full `Variant` (24 bytes) into that buffer. The write overruns the Go stack or heap, and after a GC it corrupts unrelated Go data (issue #151). The same pattern could affect any vararg method whose declared return type is not `Variant`. Nothing audits this today, so the full set of affected methods is unknown.

## What Changes

- Audit every vararg method in the class, built-in class, and utility function bindings for return-buffer type mismatches.
- Change the generated vararg call path so the return buffer is always a `Variant`, then convert the result to the declared return type. This fixes `emit_signal`, `rpc`, and `rpc_id`, and covers any other affected method the audit finds.
- Make the generator reject or guard any future vararg method whose declared return type is not `Variant` or void, so the regression cannot return silently.
- Add a regression test that emits a signal in a loop and returns a long string, checking that the returned data is intact.

## Capabilities

### New Capabilities

- `vararg-call-return-marshalling`: how generated bindings for vararg methods size and convert the return buffer passed across the GDExtension boundary.

### Modified Capabilities

None.

## Impact

- `cmd/generate/gdclassimpl/classes.go.tmpl`: vararg return buffer and conversion.
- `pkg/gdclassimpl/classes.gen.go`: regenerated; `Object.emit_signal`, `Node.rpc`, `Node.rpc_id` change.
- Generator guard for vararg return types (`cmd/generate`).
- Tests under `test/demo` for the heap-corruption regression.
- No public API signature changes: the Go return types of `EmitSignal`, `Rpc`, and `RpcId` stay `Error`.

## Non-goals

- Changing the Go signatures or argument handling of vararg methods.
- Fixing the separate question of whether `Variant` values returned from vararg calls are destroyed by the caller (ownership is handled by the caller today).
- Changing the built-in class or utility function vararg paths unless the audit finds a mismatch there.
