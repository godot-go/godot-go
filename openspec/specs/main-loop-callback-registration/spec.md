## Purpose

Defines how a Go-defined GDExtension registers and receives Godot main-loop
phase callbacks (startup, frame, shutdown), including registration timing,
single-slot overwrite semantics, tolerance of engines without the interface
function, and panic behavior at the callback boundary.

## Requirements

### Requirement: Registered phase callbacks fire at their main-loop phase
The system SHALL invoke a registered startup hook once after the engine is
fully initialized and before scene processing begins, a registered frame hook
once per process frame after node `_process` handling, and a registered
shutdown hook during main-loop shutdown while the engine is still fully
initialized.

#### Scenario: Startup hook observable before scene ready
- **WHEN** an extension registers a startup hook and a scene's `_ready` runs
- **THEN** evidence recorded by the startup hook is already observable

#### Scenario: Frame hook counts awaited frames
- **WHEN** an extension registers a frame hook and the test driver awaits
  process frames
- **THEN** the hook's invocation count increases by at least one per awaited
  frame

#### Scenario: Shutdown hook fires on exit
- **WHEN** an extension registers a shutdown hook and the engine quits
- **THEN** the hook runs and records evidence in the run output before the
  process exits

### Requirement: Phase hooks take effect from before Init
Hooks assigned through the initialization object's registration calls before
`Init()` is invoked SHALL all be effective for the process lifetime; the
engine MUST learn of them no later than core-level initialization so the
startup phase can never be missed.

#### Scenario: Hooks registered before Init all fire
- **WHEN** startup, frame, and shutdown hooks are registered prior to `Init()`
  and the engine runs and exits
- **THEN** each registered phase fires exactly per its phase semantics

#### Scenario: No hooks means no registration
- **WHEN** an extension registers no main-loop hooks at all
- **THEN** no main-loop registration is sent to the engine and extension
  behavior is unchanged

### Requirement: One hook per phase with last-write-wins
Assigning a hook to a phase SHALL replace any hook previously assigned to
that same phase; the replaced hook MUST NOT be invoked again. Phases are
independent of each other.

#### Scenario: Reassignment replaces the hook
- **WHEN** a second startup hook is assigned before `Init()` after a first one
- **THEN** only the second hook runs at startup

### Requirement: Engines without the interface function keep loading
When the host engine does not provide the main-loop registration interface
function (pre-4.5), the system SHALL skip main-loop registration with a log
record, and extension initialization MUST complete without a crash.

#### Scenario: Older engine loads unharmed
- **WHEN** an extension with registered hooks loads into an engine lacking
  the interface function
- **THEN** the extension initializes normally, a skip is logged, and no hook
  is invoked

### Requirement: Hook panics are not intercepted
The system SHALL NOT recover panics raised inside hook callbacks; a panicking
hook propagates across the engine boundary with the same fatal outcome as
every other binding callback, and this policy SHALL be documented.

#### Scenario: Panicking hook is fatal, not swallowed
- **WHEN** a registered hook panics during its phase
- **THEN** the process fails loudly rather than the phase being skipped
  silently
