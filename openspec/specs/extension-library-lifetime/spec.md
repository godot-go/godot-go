# extension-library-lifetime Specification

## Purpose
Define how long the extension's native module remains mapped, so that teardown never executes code from an unmapped Go runtime.

## Requirements

### Requirement: Windows Module Stays Mapped Until Process Exit

On Windows, the extension SHALL keep its own native module mapped until the process exits. The engine's extension-unload path, which calls the platform unload function during shutdown, MUST NOT unmap the module.

#### Scenario: Normal process exit after a test run

- **WHEN** a Windows process that loaded the extension finishes a run and begins shutdown
- **THEN** the process exits with status zero instead of faulting with an access violation

#### Scenario: Engine requests extension unload

- **WHEN** the engine unloads the extension during shutdown
- **THEN** the module remains mapped, so no runtime thread can execute unmapped code

### Requirement: Deinitialization Is Independent Of Unmapping

Preventing the unmap SHALL NOT change the initialization-level contract: every level that was initialized is still deinitialized in reverse order before the process exits.

#### Scenario: Deinit levels still run

- **WHEN** the engine tears down the extension on Windows
- **THEN** the extension's deinitialization callbacks run for every initialized level
- **AND** the module remains mapped afterwards

### Requirement: Other Platforms Keep Their Lifetime

On platforms other than Windows, the extension SHALL retain its existing module lifetime behavior; this change SHALL NOT alter the engine's ability to unload it there.

#### Scenario: Linux and macOS teardown

- **WHEN** a Linux or macOS process tears down the extension
- **THEN** the existing unload behavior is unchanged
