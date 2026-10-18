# Spec Delta

## Purpose

Bridges Godot signal emissions to Go channels so a goroutine can block on `core.Await` like GDScript blocks on `await`, with Go-native payloads and deterministic channel-close semantics.

## ADDED Requirements

### Requirement: Await delivers one emission through a channel
`core.Await(object, signalName)` SHALL return a receive-only channel that delivers exactly one result: the first emission of the named signal on the given object, with the emission arguments in declaration order. The awaiting goroutine blocks on the channel and resumes when the result arrives. After a successful delivery the channel SHALL be closed. Awaiting a timer's `timeout` signal SHALL work identically to awaiting any other object signal.

#### Scenario: Emit after await
- **WHEN** a goroutine calls `core.Await(node, "hit")` and the node later emits `hit` with arguments `4` and `"critical"`
- **THEN** the channel yields one result carrying the arguments `4`, `"critical"` in order, and the channel is then closed

#### Scenario: Zero-argument signal
- **WHEN** a goroutine awaits a signal that is emitted with no arguments
- **THEN** the channel yields one result with an empty argument list and closes

#### Scenario: Timer await
- **WHEN** a goroutine awaits `timeout` on a SceneTree timer and the timer elapses
- **THEN** the channel yields the result instead of blocking forever

### Requirement: One-shot registration
Each `core.Await` call SHALL subscribe with a one-shot connection: only the first subsequent emission reaches that channel. Later emissions SHALL NOT reach it, and issuing a fresh `core.Await` SHALL arm a new independent result channel.

#### Scenario: Repeated emissions
- **WHEN** a goroutine awaits `"damage"` and the signal emits twice
- **THEN** only the first emission is delivered on that channel, and a second `core.Await` call is required to observe the next emission

### Requirement: Payload values are Go-native
Delivered argument values SHALL be converted on the emitting thread into Go-owned representations usable after the emission has returned: nil → nil, boolean → `bool`, integers → `int64`, floats → `float64`, `String`/`StringName` → Go `string`, plain-data math built-ins → value-copied equivalents, and engine objects → object interface handles. No engine-owned refcounted Variant SHALL be handed to the receiver without conversion, and reading a delivered payload SHALL NOT require the main thread or the emission to still be on the stack.

#### Scenario: String payload outlives the emission
- **WHEN** a signal is emitted with a `String` argument and the awaiting goroutine reads the payload string after the emitting function has returned
- **THEN** the string is a valid Go string with the emitted content

#### Scenario: Object payload arrives as a handle
- **WHEN** a signal is emitted with an Object argument
- **THEN** the payload carries that object's Go interface handle identifying the same engine object

### Requirement: Emission never blocks on the awaiting goroutine
Delivering a result SHALL NOT block the emitting code, whether or not a receiver has reached the channel: delivery completes even when the awaiting goroutine is descheduled or never scheduled, and the emission call returns normally.

#### Scenario: Late receiver
- **WHEN** a signal is emitted while the awaiting goroutine has not yet been scheduled to receive
- **THEN** the emission completes normally and the result is still delivered when the goroutine later receives

### Requirement: Freed object deterministically closes the await
If the awaited object is destroyed by the engine before its signal emits, the connection SHALL be released, the underlying Go callable SHALL be freed, and the channel SHALL be closed without a value being sent, so the receiver unblocks with a zero result and `ok == false` instead of hanging.

#### Scenario: Awaiter outlives the object
- **WHEN** a goroutine awaits a signal on a node and the node is freed before the signal emits
- **THEN** the channel closes without a send and the receive returns a zero result with `ok == false`

#### Scenario: Callable registry drains after free
- **WHEN** an awaited object is freed and the await resolves as closed
- **THEN** the Go callable registry count no longer includes the await's callable

### Requirement: Misuse is reported immediately, never as a silent hang
Calling `Await` with an invalid signal name on a valid object SHALL raise an immediate diagnostic at the call site instead of returning a channel that can never resolve, and SHALL NOT crash the process through the engine boundary.

#### Scenario: Unknown signal name
- **WHEN** code calls `core.Await` with a signal name the object does not declare
- **THEN** the call reports the failure to the caller at call time and no dangling subscription is created

### Requirement: Shutdown with pending awaits stays clean
When the engine shuts down while one or more awaits are pending, teardown SHALL release the subscriptions, pending channels SHALL unblock (closed or process exit — never a shutdown deadlock), and the extension unload SHALL leave no leaked objects and no leaked Go callable registrations.

#### Scenario: Exit while awaiting
- **WHEN** the engine exits while a goroutine is blocked in `core.Await`
- **THEN** shutdown completes normally and the shutdown leak checks report zero leaked objects and zero live Go callable registrations
