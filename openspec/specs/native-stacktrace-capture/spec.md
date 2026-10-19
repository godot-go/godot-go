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
