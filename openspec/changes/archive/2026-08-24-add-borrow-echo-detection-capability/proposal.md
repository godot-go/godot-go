## Why

The borrow-echo detection introduced for containers and extended to `StringName`/`NodePath` is content-based (`reflect.DeepEqual` byte comparison against the call's borrow-decoded arguments). That mechanism has a documented, inherent limitation: an *unmutated defensive clone* of an argument — owned, but byte-identical to the borrow because Godot copy constructors share storage (CoW) or intern data (StringName) — is classified as an echo and its reference never released: a silent leak of one reference per call. The limitation currently lives only in type-specific spec scenarios and PR review commentary; there is no single normative home for how echo classification works and what it guarantees across types and call styles.

## What Changes

- Add a new `borrow-echo-detection` capability defining the general contract: which types are covered, how classification works on both return paths, what happens to echoes vs non-echoes, and the same-content clone misclassification as accepted, bounded behavior.
- No production code changes: this change captures existing, verified behavior as a specification.

## Capabilities

### New Capabilities

- `borrow-echo-detection`: Content-based classification of returned refcounted built-ins against a call's borrow-decoded arguments, deciding whether the Go-side source value is destroyed after encoding — including the documented same-content clone leak limitation.

### Modified Capabilities

None. The container and stringname capabilities keep their type-specific scenarios; this capability holds the cross-type contract they share.
