# Spec Delta

## ADDED Requirements

### Requirement: Windows native capture via DbgHelp
On Windows builds the library SHALL print a native call stack from `printStacktrace()` when the global enable flag is set, walking frames with the process's stack-walking facilities. Each frame line SHALL carry a symbol name with displacement when the resolver can attribute it, and otherwise the containing module name with the instruction offset into the module (plus the raw instruction pointer), so unresolved frames remain actionable offline.

#### Scenario: Enabled on Windows
- **WHEN** the global enable flag is true and a bound GDExtension callback fires on Windows
- **THEN** the captured native callstack frames are printed between the standard start/end banner lines

#### Scenario: Unresolved frames stay actionable
- **WHEN** a captured frame cannot be attributed to a symbol name (e.g. no PDB is available for a mingw build carrying DWARF)
- **THEN** that frame line still prints the module name plus the offset into the module and the raw instruction address

### Requirement: CI-visible capture probe
The test suite SHALL, on every platform run, emit exactly one probe line reporting whether native capture is compiled in, and SHALL perform at least one enabled capture window around a bound callback. The runtime output gate SHALL treat a claimed-available capture with no produced banner as a test failure, and SHALL treat the probe line's absence as a test failure.

#### Scenario: Windows CI proves real capture
- **WHEN** `make test` runs on Windows CI and the probe reports capture available
- **THEN** the run's output contains at least one backtrace banner and the gate passes

#### Scenario: Stub platform passes without banners
- **WHEN** a platform whose build compiles capture out runs the suite and the probe reports capture unavailable
- **THEN** the gate passes and requires no banner

#### Scenario: Claimed capture produces nothing
- **WHEN** a run reports capture available but its output contains no backtrace banner
- **THEN** the runtime gate fails naming the stacktrace capture as the cause

#### Scenario: Capture state does not leak
- **WHEN** the probe's capture window closes
- **THEN** subsequent callbacks produce no further banner output for the rest of the run

#### Scenario: Companion Go stack prints
- **WHEN** the probe's hit phase runs on any platform, including one whose native capture is stubbed out
- **THEN** the output contains a companion marker line followed by Go frames carrying source file and line numbers, and the runtime gate fails the run if that companion stack is absent while the probe ran
