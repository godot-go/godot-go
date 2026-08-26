## Purpose

Lets Go classes implement Godot virtual methods at every level of a struct-embedding hierarchy simultaneously, using qualified names (`V_ClassName_MethodName`) so implementations at different depths coexist instead of shadowing each other. Dispatch selects the most-derived implementation for an instance; shallower implementations remain explicitly callable; virtuals no level implements fall back to engine defaults.

## Requirements

### Requirement: Qualified Names Identify Per-Level Implementations
A Go method named `V_<ClassName>_<MethodName>` SHALL be registrable as the implementation that type `<ClassName>`'s embedding level provides for the corresponding GDScript virtual (`_`-prefixed snake_case GDScript name maps to PascalCase in the Go name). Registration of a qualified name SHALL NOT conflict with implementations registered at other embedding levels.

#### Scenario: Base and derived levels both register the same virtual
- **WHEN** an embedded Impl type declares `V_<Base>_<Method>` and the embedding type declares `V_<Derived>_<Method>`, and the derived class is registered
- **THEN** registration succeeds without ambiguity and both methods remain callable from Go

### Requirement: Dispatch Selects The Most-Derived Implementation
When Godot invokes a virtual on an instance whose class registered multiple levels' implementations of that virtual, the binding SHALL invoke the implementation declared by the most-derived type in the instance's embedding chain.

#### Scenario: Derived override wins
- **WHEN** a virtual is invoked on an instance of a class whose own type declares a qualified implementation
- **THEN** that implementation runs, not one promoted from an embedded Impl type

#### Scenario: Promoted base implementation serves when derived does not declare one
- **WHEN** a virtual is invoked on an instance whose own type declares no implementation but an embedded Impl type does
- **THEN** the shallowest (nearest promoted) implementation in the chain runs

### Requirement: Shallower Implementations Remain Explicitly Callable
An implementation declared at a shallower embedding level SHALL remain directly invocable from Go through its promoted method selector, enabling delegation patterns where a derived implementation extends rather than replaces the base behavior.

#### Scenario: Delegation to the base-level implementation
- **WHEN** a derived `V_<Derived>_<Method>` calls the promoted `V_<Base>_<Method>` on its embedded Impl value
- **THEN** the base-level implementation executes with the embedded receiver

### Requirement: Unimplemented Virtuals Fall Back To Engine Defaults
When no type along an instance's embedding chain registers an implementation for a requested virtual, registration-time lookup SHALL report the virtual as unimplemented so Godot uses its engine default return value.

#### Scenario: No level implements the virtual
- **WHEN** a class registers without any qualified or flat implementation of a virtual Godot queries
- **THEN** call-data lookup reports it absent (nil), preserving the behavior pinned by gdvirtual-unimplemented-defaults

### Requirement: Flat Virtual Registrations Are Rejected At Registration Time
The flat `V_<MethodName>` form SHALL NOT register: attempting it panics at registration time with a diagnostic that names the requested GDScript virtual and the expected qualified form. Failure is startup-loud so no class ever ships with silently mis-resolved virtuals.

#### Scenario: Flat registration fails loudly and deterministically
- **WHEN** a virtual is registered with the flat name `V_Ready` for `_ready`
- **THEN** registration panics immediately with a message naming `_ready` and directing to `V_<ClassName>_Ready`

#### Scenario: Qualified registration of the same method succeeds
- **WHEN** after the failed attempt the same virtual is registered as `V_<ClassName>_Ready`
- **THEN** registration succeeds and dispatch reaches that implementation
