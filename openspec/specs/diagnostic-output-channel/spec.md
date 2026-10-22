## Purpose

Defines the binding's single diagnostic surface: how the entry threshold is chosen,
where entries are written, which entries carry a call stack, the fatal path binding
faults take, the one boundary that converts a panic into a reported failure, and the
availability of diagnostics while the engine tears down.

## Requirements

### Requirement: The entry threshold is chosen from the environment at process start

The library SHALL select its entry threshold from a single environment variable read at
process start, SHALL apply the default threshold when that variable is absent or cannot
be understood, and SHALL suppress entries below the threshold in all packages.

#### Scenario: Threshold not configured

- **WHEN** the extension is loaded with no threshold configured
- **THEN** warning-level entries and above are emitted
- **AND** debug and info entries are not emitted

#### Scenario: Verbose threshold configured

- **WHEN** the environment requests the information threshold before the process starts
- **THEN** information-level entries are emitted

#### Scenario: Unparseable threshold configured

- **WHEN** the environment names a threshold the library cannot understand
- **THEN** the default threshold is applied and start-up continues

### Requirement: Entries are written to standard output without engine routing

Diagnostic entries SHALL be written to the process's standard output, serialized so that
concurrent writers never interleave parts of different entries, and SHALL NOT be sent
through the engine's message-printing interface.

#### Scenario: Entries reach the run log

- **WHEN** the extension logs during a run
- **THEN** the entries appear in the process output the harness captures

#### Scenario: Concurrent writers

- **WHEN** two goroutines log at the same moment
- **THEN** every entry appears whole on its own line

#### Scenario: Engine print interface stays unused

- **WHEN** the library reports a diagnostic
- **THEN** the engine's error, warning, and script-error printing is not used for it

### Requirement: Entries at the warning threshold and above carry a call stack

The library SHALL attach a call stack to entries at the warning threshold and above and
SHALL NOT attach one to entries below it.

#### Scenario: Warning entry

- **WHEN** a warning-level entry is emitted
- **THEN** the entry includes a call stack

#### Scenario: Debug entry

- **WHEN** a debug-level entry is emitted at a threshold that admits it
- **THEN** the entry carries no call stack

### Requirement: Binding faults are recorded before they unwind

A binding fault the caller cannot induce SHALL be recorded as an entry and then unwind
through the panic path, so the fault is never hidden by the configured threshold and
never reported as a silent process exit.

#### Scenario: Fault while binding an unsupported construct

- **WHEN** a bind request reaches a construct the binding cannot represent
- **THEN** an entry identifying the fault is recorded
- **AND** the request unwinds instead of returning a partially bound result

#### Scenario: Fault is visible at the default threshold

- **WHEN** a binding fault occurs with the default threshold applied
- **THEN** the fault entry is emitted even though quieter entries are suppressed

#### Scenario: Fault names what it rejected

- **WHEN** a request is refused for a specific type, class, or name
- **THEN** the recorded entry identifies the rejected construct

### Requirement: Only the user closure boundary converts a panic into a reported failure

A panic raised inside a user closure invoked from a Godot-side callable SHALL be
recovered at that boundary, reported as a failed call the engine can surface, and
recorded, leaving the process running. No other engine callback boundary SHALL recover.

#### Scenario: User closure panics

- **WHEN** a callable's user closure panics while the engine invokes it
- **THEN** the panic is recovered, the call is reported as failed, and an entry is recorded
- **AND** the engine continues running

#### Scenario: Ordinary method dispatch is not recovered

- **WHEN** a bound Go method panics during an engine call
- **THEN** no recovery occurs at that boundary
- **AND** the fault reaches the process rather than becoming a reported call failure

### Requirement: Diagnostics remain available while the engine tears down

Diagnostic writing SHALL keep working during and after engine deinitialisation, including
from main-loop shutdown hooks, without depending on any engine-provided output facility.

#### Scenario: Shutdown hook reports

- **WHEN** a main-loop shutdown hook logs while the engine is tearing down
- **THEN** the entry is written to the captured output

#### Scenario: Teardown diagnostics survive without the engine

- **WHEN** an entry is emitted after the engine has released its own facilities
- **THEN** the entry is still written and the process is not faulted by the log call
