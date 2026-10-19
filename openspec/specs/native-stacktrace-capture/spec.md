# native-stacktrace-capture Specification

## Purpose

Governs the optional C-side callstack capture at GDExtension callback boundaries: how it is enabled, which platforms provide real capture, and how it degrades where the platform lacks a native backtrace interface.

## Requirements

### Requirement: Build succeeds without execinfo
The library SHALL compile and link on systems whose C standard library provides no `execinfo` backtrace interface (for example musl libc), with no source modifications and no extra packages required.

#### Scenario: Alpine musl build
- **WHEN** the build pipeline (`make generate` then `make build`) is run on Alpine Linux with only the base C toolchain installed
- **THEN** the build completes successfully and produces the GDExtension shared library

#### Scenario: Header carries no platform dependency
- **WHEN** any package in the project includes the native stacktrace header
- **THEN** that include compiles whether or not the platform provides `execinfo`

### Requirement: Real capture where execinfo exists
On platforms where the `execinfo` backtrace interface is available (glibc Linux, Apple platforms), the system SHALL print the native call stack when a callback is invoked and capture is enabled via the global enable flag.

#### Scenario: Enabled on glibc
- **WHEN** the global enable flag is true and a bound GDExtension callback fires on glibc Linux
- **THEN** the captured native callstack frames are printed

#### Scenario: Capture stays off by default
- **WHEN** the global enable flag has not been set on any platform
- **THEN** no native stacktrace output is produced

### Requirement: Graceful degradation without execinfo
On platforms without `execinfo`, the capture facility SHALL compile to a no-op: the enable flag remains settable, callback dispatch behavior is unchanged, and no output, crash, or link failure results.

#### Scenario: musl with flag enabled
- **WHEN** the global enable flag is set to true on a musl-based system and a bound callback fires
- **THEN** the callback completes normally and no native stacktrace output is produced

### Requirement: Opt-in restore via external execinfo implementation
The build SHALL allow a user on a platform without native `execinfo` to restore real capture without source modifications, by passing a documented preprocessor define that selects the execinfo code path and linking an external implementation.

#### Scenario: Alpine with an external execinfo implementation
- **WHEN** the documented preprocessor define and linker flags pointing at an external library that provides the glibc-compatible `execinfo` interface are used on Alpine (Alpine ships no installable implementation, so the library is user-supplied)
- **THEN** the execinfo capture path is compiled in and linking succeeds against that library

### Requirement: Windows native capture via DbgHelp
On Windows builds the library SHALL print a native call stack from `printStacktrace()` when the global enable flag is set, walking frames with the process's stack-walking facilities. Each frame line SHALL carry a symbol name with displacement when the resolver can attribute it, and otherwise the containing module name with the instruction offset into the module (plus the raw instruction pointer), so unresolved frames remain actionable offline.

#### Scenario: Enabled on Windows
- **WHEN** the global enable flag is true and a bound GDExtension callback fires on Windows
- **THEN** the captured native callstack frames are printed between the standard start/end banner lines

#### Scenario: Unresolved frames stay actionable
- **WHEN** a captured frame cannot be attributed to a symbol name (e.g. no PDB is available for a mingw build carrying DWARF)
- **THEN** that frame line still prints the module name plus the offset into the module and the raw instruction address

### Requirement: Windows frames carry source line info when a PDB is linked

Captured Windows frames whose addresses resolve to source through the linked PDB SHALL print the source file and line alongside the symbol, and the Windows CI capture-evidence step SHALL fail a run where no captured frame carries file:line information.

#### Scenario: Resolved frame prints file and line

- **WHEN** the capture window prints a frame that resolves through the PDB
- **THEN** the frame line includes the source file name and line number in addition to the symbol and displacement

#### Scenario: Frames without line info keep the offline form

- **WHEN** a captured frame's module carries no line information (for example Go runtime or engine DLLs)
- **THEN** the frame prints unchanged from the symbol+offset or `module+0xoffset (ip 0x…)` forms

#### Scenario: Missing line evidence fails the Windows job

- **WHEN** a Windows CI run's captured banners contain no frame with file:line
- **THEN** the stacktrace evidence verification step fails the job naming the missing line-info evidence

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
