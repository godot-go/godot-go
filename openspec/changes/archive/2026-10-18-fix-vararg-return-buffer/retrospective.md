# Retrospective: Vararg Return Buffer Heap Corruption (issue #151)

## Before vs After

```
BEFORE: EmitSignal/Rpc/RpcId -> var ret Error (8B) written with a 24B Variant
AFTER:  var retVariant Variant (24B) + conversion, ~1-in-15 heap corruption gone
```

5/5 headless runs and the `GOGC=off` control pass the full suite gate
(1427 assertions). `make generate` is idempotent; strict OpenSpec validation
passes.

## What was tried, what worked

1. **Confirmed the issue against `origin/main` first.** The template and the
   three generated methods still carried the old shape; no open PR touched it.
2. **Full-surface audit instead of a point fix.** Enumerated every
   `is_vararg` method in `extension_api.json` across classes, built-in
   classes, and utility functions, then encoded the result as
   `cmd/generate/gdclassimpl/vararg_census.json` plus a drift test.
3. **Cross-referenced godot-cpp.** `binding_generator.py` declares
   `Variant ret;` for every vararg class binding and converts with
   `VariantCaster<Error>::cast(ret)` — confirming the Variant-buffer +
   int64-cast shape before writing the template branch.
4. **Generator guard, not just a fix.** `checkVarargReturnTypes` fails
   generation for any vararg return type lacking a `Variant` conversion, and
   `TestGeneratedVarargBuffersAreVariantSized` scans the committed
   `classes.gen.go` so a future regeneration cannot regress silently.

## Root cause analysis

### 1. Undersized r_return buffer for varcalls

Vararg calls use varcall (`object_method_bind_call`); the engine always
constructs a full `Variant` at `r_return`. The template passed a pointer to
the declared-type value instead.

```go
// BEFORE — engine writes 24B into an 8B Error slot (pinned stack cell)
var ret Error
retPtr := (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(&ret))
...
return ret

// AFTER — Variant-sized buffer, converted after the call
var retVariant Variant
defer retVariant.Destroy()
retPtr := (GDExtensionUninitializedVariantPtr)(unsafe.Pointer(&retVariant))
...
return Error(retVariant.ToInt64())
```

**Files:** `cmd/generate/gdclassimpl/classes.go.tmpl`,
`pkg/gdclassimpl/classes.gen.go` (generated: `Object.emit_signal`,
`Node.rpc`, `Node.rpc_id`).

```
  varcall r_return contract

  engine                     Go BEFORE                Go AFTER
  +---------+   write 24B   +---------+              +--------------+
  | Variant | ----------->  | Error 8B| --clobber--> | Variant 24B  |
  +---------+               +---------+   of 16B     |  (then cast  |
                            +---------+   adjacent    |   to Error)  |
                            | stack   |   memory     +--------------+
                            +---------+
```

### 2. Audit-count drift found mid-implementation

The planning audit had conflated class and utility varargs (27 "class"). The
census script showed the true split: 15 class, 6 built-in, 12 utility. The
broken set (`emit_signal`, `rpc`, `rpc_id`) was unaffected; design.md and
tasks.md were corrected. Lesson: generate counts from the API mechanically
before publishing them.

### 3. Wrong engine expectation caught by the live run

The new test initially expected `ERR_INVALID_PARAMETER` (31) for emitting a
nonexistent signal; the engine returns `ERR_UNAVAILABLE` (2) — `Object::emit_signalp`
rejects it before any emit. The test failing with the engine's true value was
itself evidence the conversion works: the fixed binding faithfully returns
whatever Error the engine wrote into the Variant.

**File:** `test/pkg/vararg_return_tests.go`.

## Go-side vs Godot-side

Root cause is entirely Go-side (codegen template). The Godot-side fact it
relies on — varcall always writes a `Variant` — is engine contract, matching
godot-cpp. No engine bug found.

## Remaining items (traced, out of scope)

- `Variant`-returning vararg bindings hand ownership to the Go caller with no
  auto-destroy; governed by the separate ownership capability, unchanged here.
- `openspec/config.yaml` has an unquoted `example:` rule the CLI warns about;
  unrelated to this change.
- Pre-existing `go vet` complaints in `test/concepts/*` (untouched by this
  branch).

## Key learnings

- The three broken methods were invisible to every earlier test run because
  the corruption needed a GC-timed collision (~1/15); a regression test must
  interleave emit + guard-string + GC in a long loop to have signal.
- `runtime.Pinner` keeps `retVariant` stack-resident, so the buffer's size is
  what protects neighbors — the pin does not.
- Encoding an audit as a fixture + drift test converts a one-time spreadsheet
  into a permanent API-refresh check.
