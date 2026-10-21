# Spec Delta

## REMOVED Requirements

### Requirement: Windows native capture via DbgHelp
**Reason**: Windows capture now uses the libbacktrace implementation bundled with Godot, is compiled in by default, and no longer imports DbgHelp.
**Migration**: Rebuild without `GODOT_GO_ENABLE_DBGHELP`; capture is available by default. Remove any workflow or documentation that treats `capture=0` as the normal Windows state.

### Requirement: Windows frames carry source line info when a PDB is linked
**Reason**: Windows frames are resolved from the extension module's embedded DWARF debug information, so no PDB is produced or required.
**Migration**: Remove PDB upload and `llvm-symbolizer --pdb` steps. Offline resolution uses the existing `go tool addr2line` recipe against the DLL.

## ADDED Requirements

### Requirement: Windows native capture
Windows builds SHALL compile native callstack capture in by default and SHALL print the captured stack from `printStacktrace()` between the standard start and end banners when the global enable flag is set. The library SHALL NOT import or load `dbghelp.dll`, and capture availability SHALL report true on Windows.

#### Scenario: Enabled on Windows
- **WHEN** the global enable flag is true and a bound GDExtension callback fires on Windows
- **THEN** the captured native callstack frames are printed between the standard start and end banner lines

#### Scenario: Default build has capture compiled in
- **WHEN** a default Windows build queries capture availability
- **THEN** availability reports true without any extra build flag

#### Scenario: No DbgHelp dependency
- **WHEN** a Windows build's module imports are inspected, or the process runs with capture enabled
- **THEN** `dbghelp.dll` is neither imported nor loaded

### Requirement: Windows frames resolve from embedded DWARF
On Windows, captured frames in the extension module SHALL be resolved from the module's embedded DWARF debug information. A resolved frame SHALL print a function name and source `file:line`. A frame in another module, or one with no resolvable debug information, SHALL print the containing module name, the offset into that module, and the raw instruction pointer. The Windows CI capture-evidence step SHALL fail a run in which capture is available but no captured frame carries `file:line`.

#### Scenario: Resolved frame prints function and source line
- **WHEN** the capture window prints a frame in the extension module whose address resolves through its DWARF
- **THEN** the frame line includes the function name and `file:line`

#### Scenario: Frames without line info keep the offline form
- **WHEN** a captured frame belongs to another module or has no resolvable debug information
- **THEN** the frame prints the module name, the offset into the module, and the raw instruction pointer

#### Scenario: Missing line evidence fails the Windows job
- **WHEN** a Windows CI run reports capture available but its captured banners contain no frame with `file:line`
- **THEN** the stacktrace evidence verification step fails the job naming the missing line-info evidence
