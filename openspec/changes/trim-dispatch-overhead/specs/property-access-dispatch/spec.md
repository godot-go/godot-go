# Spec Delta

## MODIFIED Requirements

### Requirement: An unresolvable class fails a get, set, or validation request

A property get, set, or validation request whose instance cannot be resolved to a registered class SHALL fail loudly through the library's fatal-error channel rather than dereferencing a missing registration. A revert or notification request for an instance that cannot be addressed — its Go wrapper having begun being freed, or its class having no registration — takes that situation as a deliberate decline, and deciding which case applies SHALL NOT depend on the class name the engine reports for the instance.

#### Scenario: Unknown class registration

- **WHEN** a get, set, or validation callback receives an instance whose class has no registration
- **THEN** the request fails loudly and the engine receives no fabricated value

#### Scenario: Revert on a torn-down instance

- **WHEN** a revert or notification callback receives an instance whose wrapper has begun being freed
- **THEN** the request declines without an error
