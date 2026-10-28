# Spec Delta

## Purpose

Defines how engine→Go method dispatch works once it is no longer reflection-driven: bound methods are invoked through dispatch closures compiled from type information captured at bind time, with a registration API that catches signature mismatches at compile time, while the reflective name-string API remains available as a fallback with identical observable semantics.

## ADDED Requirements

### Requirement: Ptrcall dispatch performs no per-call reflection

When the engine invokes a bound Go method through ptrcall, dispatch SHALL execute a closure compiled at bind time from the method's captured signature. The per-call path SHALL NOT re-derive the method signature, construct per-call abstract value slices, or invoke the method through abstract value calls.

#### Scenario: Per-frame virtual dispatch allocates no abstract value machinery

- **WHEN** a Go class implements a per-frame virtual and the engine invokes it every frame
- **THEN** each dispatch runs the compiled closure directly, with argument values arriving exactly as the corresponding reflective decode would produce them

#### Scenario: GDScript calls a bound method through ptrcall

- **WHEN** GDScript calls a bound Go method with argument types matching the bind
- **THEN** the Go method receives decoded argument values equal to the reflective decode's and its return is encoded normally

### Requirement: Virtual and callback dispatch use the same compiled closures

Virtual method callbacks, notification dispatch, and property/dispatcher callbacks SHALL route through the same compiled-closure mechanism as ptrcall; reflection SHALL survive only at bind time.

#### Scenario: Notification reaches a compiled closure

- **WHEN** the engine sends a notification callback to a registered Go class
- **THEN** the handler is invoked through a compiled closure with correctly decoded arguments

### Requirement: A compile-time-checked registration API exists alongside the name-string API

The binding SHALL offer registration overloads that accept a handler captured as a typed reference (a Go method expression), making a handler whose shape does not fit the bind unrepresentable at bind time. The existing name-string registration API SHALL continue to work with unchanged semantics, including virtual-name resolution along the embedding chain and duplicate-name rejection.

#### Scenario: Handler type mismatch is a compile error

- **WHEN** a handler reference whose Go signature does not match the bound shape is passed to a typed registration overload
- **THEN** the program does not compile

#### Scenario: Name-string API unchanged

- **WHEN** a class is registered through the existing name-string API
- **THEN** it binds and dispatches as before, including embedding-chain virtual resolution

### Requirement: Typed dispatch preserves decode and lifecycle semantics exactly

The compiled decode used by typed dispatch SHALL produce argument values, reference-count effects, container ownership, and pinning behavior identical to the reflective decode it replaces, per the capabilities that pin those behaviors.

#### Scenario: Refcounted and container arguments round-trip identically

- **WHEN** a bound Go method is called with String, StringName, NodePath, Array, Packed*Array, Vector, and object arguments (ptrcall and varcall respectively)
- **THEN** decoded values, returned encodings, and reference accounting match the reflective path's observable results

### Requirement: Typed dispatch preserves borrow-echo classification

Return-path echo classification under typed dispatch SHALL classify a returned refcounted built-in as a borrow echo exactly when the reflective path would (byte-equality against the call's decoded arguments), preserving the type-and-path release matrix and the variadic no-suppression rule.

#### Scenario: Echo and non-echo classifications agree with the legacy path

- **WHEN** Go methods return (a) an argument they received unchanged, (b) a distinct value through a releasing path, and (c) a refcounted built-in constructed inside a variadic method
- **THEN** classification, release, and retention outcomes match the borrow-echo capability's scenarios exactly

### Requirement: Varcall default and variadic semantics survive typed dispatch

Where a bound method declares trailing defaults or variadic arguments, typed dispatch SHALL fill omitted trailing arguments from the bind's stored defaults and route variadic calls as raw Variants with no echo suppression, matching the default-arguments and variadic-binding capabilities.

#### Scenario: Trailing defaults filled by the typed closure

- **WHEN** a GDScript call omits trailing arguments that the bind declares defaults for
- **THEN** the Go method observes the same resolved arguments the reflective path produces

#### Scenario: Variadic method through typed dispatch

- **WHEN** a variadic Go method is called through the typed mechanism
- **THEN** received Variants are owned raw copies and no echo classification applies

### Requirement: Untypeable signatures fall back without behavior change

A bind whose handler signature cannot be captured by a typed wrapper SHALL be registered through the reflective fallback and SHALL dispatch with the same decode, echo, default, and error-reporting semantics as typed binds; hot-path dispatch SHALL fail loudly rather than silently degrade into the fallback.

#### Scenario: Unexpected signature binds via fallback

- **WHEN** a handler with an exotic signature is bound through an API shape that cannot capture it as a typed wrapper
- **THEN** the bind succeeds and calls dispatched to it behave identically to typed binds
