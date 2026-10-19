## ADDED Requirements

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
