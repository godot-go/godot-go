# Spec Delta: native-stacktrace-capture

## MODIFIED Requirements

### Requirement: Windows native capture via DbgHelp

On Windows builds where the DbgHelp capture path is compiled in via the documented build-time opt-in, the library SHALL print a native call stack from `printStacktrace()` when the global enable flag is set, walking frames with the process's stack-walking facilities. Each frame line SHALL carry a symbol name with displacement when the resolver can attribute it, and otherwise the containing module name with the instruction offset into the module (plus the raw instruction pointer), so unresolved frames remain actionable offline. A default Windows build SHALL compile the capture path out entirely, SHALL NOT load the DbgHelp component, and its availability report SHALL state that native capture is unavailable while leaving the enable flag settable.

#### Scenario: Enabled on Windows
- **WHEN** the build was compiled with the documented DbgHelp opt-in, the global enable flag is true, and a bound GDExtension callback fires on Windows
- **THEN** the captured native callstack frames are printed between the standard start/end banner lines

#### Scenario: Default build compiles capture out
- **WHEN** a default Windows build runs and the capture availability is queried (including with the global enable flag set)
- **THEN** availability reports capture unavailable, no DbgHelp module is loaded into the process, and no native stacktrace output is produced

#### Scenario: Unresolved frames stay actionable
- **WHEN** a captured frame cannot be attributed to a symbol name (e.g. no PDB is available for a mingw build carrying DWARF)
- **THEN** that frame line still prints the module name plus the offset into the module and the raw instruction address

### Requirement: Windows frames carry source line info when a PDB is linked

When the Windows capture path is compiled in, captured frames whose addresses resolve to source through the linked PDB SHALL print the source file and line alongside the symbol, and the Windows CI capture-evidence step SHALL fail a run in which capture is compiled in but no captured frame carries file:line information. When the capture path is compiled out, the evidence step SHALL accept the explicit unavailable report as a healthy state and SHALL NOT require banner or line evidence.

#### Scenario: Resolved frame prints file and line

- **WHEN** the capture window prints a frame that resolves through the PDB
- **THEN** the frame line includes the source file name and line number in addition to the symbol and displacement

#### Scenario: Frames without line info keep the offline form

- **WHEN** a captured frame's module carries no line information (for example Go runtime or engine DLLs)
- **THEN** the frame prints unchanged from the symbol+offset or `module+0xoffset (ip 0x…)` forms

#### Scenario: Missing line evidence fails the Windows job

- **WHEN** a Windows CI run reports capture available but its captured banners contain no frame with file:line
- **THEN** the stacktrace evidence verification step fails the job naming the missing line-info evidence

#### Scenario: Compiled-out capture passes the evidence step

- **WHEN** a Windows CI run's probe reports capture unavailable because the capture path was compiled out
- **THEN** the stacktrace evidence verification step passes without requiring banner or line evidence
