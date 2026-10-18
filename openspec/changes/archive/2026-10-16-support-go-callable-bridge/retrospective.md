# Retrospective: support-go-callable-bridge

## What was built

`core.NewGoCallable` turns a Go closure into an engine-native `Callable` via
`callable_custom_create2`. Three callbacks (`call`, `free`, `to_string`),
`token = FFI.Library`, `cgo.Handle` registry, borrowed argument views,
assignment-semantics return, panic barrier. Every piece mirrored an existing
in-repo precedent, which is why the core compiled and passed on the first
in-engine run.

## Investigation journey

### 1. Design pre-lit by exploration

The explore phase mapped every needed mechanism to existing code
(varcall callback, classdb shims, `cgo_handle_to_ptr`, A3's Callable copy
marshalling), so implementation was assembly, not discovery. The header's
"only `call_func` and `token` are strictly required" sentence drove the
three-callback minimum.

### 2. Return semantics: caught mid-implementation

Design D5 originally said "bit-copy the closure's Variant into `r_return`"
(copying the varcall binder). While locating the borrow-echo machinery we
found `isPtrcallBorrowEcho` and its comment: a byte-identical return cannot
be distinguished between "borrowed argument passed through" (must not gain
ownership) and "owned copy" (must). The varcall binder gets away with it
because its encoder knows ownership; a user closure cannot.

Fix: copy-construct the engine slot with `variant_new_copy` (godot-cpp
assignment parity). Borrowed pass-through is then always safe; the closure
keeps ownership of its return value, matching the repo-wide copy-on-store
policy. One extra refcount op, zero heuristics.

### 3. cgo cross-package struct identity

Constructing `C.GDExtensionCallableCustomInfo2` in `pkg/core` and converting
to `ffi.GDExtensionCallableCustomInfo2` fails: cgo types from different
packages are distinct types even with identical layout. Resolution: the same
pattern `class_create_info.go` already used — a constructor inside `pkg/ffi`
(`callable_custom_info.go`) taking ffi-alias parameters. Field name trap:
v2's tenth field is `get_argument_count_func`, not the header-comment's
shorter name.

### 4. The two phantom registry entries (the real debug)

Suite went 1356/1358 green with two count failures: after dropping all
references, the registry showed one entry more than expected, and the final
global check saw nine instead of zero.

Instrumented create/free with handle-stamped logs. Verdict:
every `free_func` fired correctly; all 22 handles balanced except two, and
the two survivors were callables whose last references were held by the
**GDScript interpreter, not our registry**:

```
 GDScript call-site scratch (per call site, lives until overwritten)
 ---------------------------------------------------------------
 example.call("create_counter_callable")   -> return Variant parked
   .call()                                    in the OP_CALL site's
                                              static scratch
 live.append(cb)  x11                      -> last arg Variant parked
                                              at the append site
```

Failed-out `assert_equal(...)` keeps the last method-call return per call
site; `Array.clear()` releases the array's references, but the loop's final
argument copy stays pinned until that exact site re-executes or the script
dies. First fix attempt — a "benign replacement call"
`example.call("go_callable_registry_count")` — failed because scratches are
per-site: the replacement had its own scratch and freed nothing.

Final approach: accept the interpreter's ownership as real engine-side
references, make them deterministic in the test: an eleventh sacrificial
loop entry so the post-clear count provably equals `baseline + 1` (the
dynamic-probe scratch's pin, already inside `baseline`). The coroutine-frame
variant of the same phenomenon — locals of an awaited helper die *after* the
awaiting function resumes — killed the wrapper-level `== 0` check; the
release word moved inside the scenario frame.

Key distinction this pinned down: a registry entry alive because *the engine
holds a reference* is correct bookkeeping, not a leak. The leak gate (and
the teardown batch visible in logs) confirms the engine drops everything at
exit.

## Remaining notes

- `assert_equal`'s failure printer uses a missing `print_raw`; failures only
  surface via the summary banner (pre-existing, out of scope).
- The `SCRIPT ERROR: Invalid call. Nonexistent function 'GoCallable
  (Callable)'` wording on failed callable invocations is how GDScript
  formats a call-error with our `GDEXTENSION_CALL_ERROR_INVALID_METHOD`;
  message quality is engine-side.
- GDScript typed `Callable` locals cannot be nulled; release-by-scope-exit
  is the only pattern (cf. typed `Resource` locals in the Ref tests).

## Learnings

1. GDScript call-site scratches are real reference holders; count-based
   assertions must account for or drain them at the exact site.
2. Log the identity (handle) of every lifecycle event before reasoning about
   refcounts; the phantom hunt took one instrumented run versus several
   speculative ones.
3. Mirroring godot-cpp semantics (assignment return) removes entire hazard
   classes instead of patching them case-by-case.
