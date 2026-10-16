# Retrospective: fix-ffi-loader-names

## What this was

The roadmap treated `callable_custom_get_user_data` as a missing engine
interface function and gated the C1 Go-func->Callable bridge behind a spike to
verify whether that gap blocked custom callables. This change proved the gap
was ours, not the engine's, and removed the whole bug class.

## The investigation journey

1. **Start from the known-issue list, not the assumption.** AGENTS.md tolerated
   four startup warnings as engine gaps. The explore phase asked the inverse
   question: does the engine register a similar name we are failing to ask for?
2. **Audit both inventories mechanically.** Extracted every string the
   generated loader requests:
   `grep -oP 'LoadProcAddress\("\K[^"]+' pkg/ffi/ffi.gen.go | sort` (177
   names), and every name the engine actually registers:
   `REGISTER_INTERFACE_FUNC(...)` in `core/extension/gdextension_interface.cpp`
   plus `register_interface_function("...")` in `core/extension/gdextension.cpp`
   (180 names). `comm -23 want have` narrowed "missing engine functions" from
   the whole loader to four candidates.
3. **A false negative along the way.** The first `have` set covered only
   `gdextension_interface.cpp` and falsely flagged the entire
   `classdb_register_extension_class*` family as missing. Those register from
   `gdextension.cpp` instead; widening the extraction collapsed the list to
   four. A single-file grep would have sent the investigation far off course.
4. **Match each failure to its registration.** All four existed under nearly
   identical names, and each traced to the same generator function.

## Root cause

`loadProcAddressName` in `cmd/generate/ffi/templatefunctions.go` reconstructs
the C symbol by snake-casing the Go typedef identifier and hand-patching
collisions:

```go
ret := strcase.ToSnake(typeName)          // GetUserData -> ...get_user_data
ret = strings.Replace(ret, "_2", "2", 1)  // fixups for class2/class3/class4...
```

The header's `@name` doc comment - the authoritative registration name - was
never consulted. Four drift cases:

| before | after | drift class |
|---|---|---|
| `callable_custom_get_user_data` | `callable_custom_get_userdata` | `ToSnake` splits `UserData` |
| `classdb_register_extension_class_5` | `classdb_register_extension_class5` | fixups covered `_2/_3/_4`, nobody added `_5` |
| `classdb_register_extension_class_6` | `classdb_register_extension_class6` | same |
| `get_variant_get_internal_ptr_func` | `variant_get_ptr_internal_getter` | typedef kept a legacy identifier while `@name` moved |

Before/after at the load site (`pkg/ffi/ffi.gen.go`):

```go
// before
x.CallableCustomGetUserData = (...)(LoadProcAddress("callable_custom_get_user_data"))
// after
x.CallableCustomGetUserData = (...)(LoadProcAddress("callable_custom_get_userdata"))
```

The Go-side API (`FFI.CallableCustomGetUserData`, `CallFunc_*` wrappers)
derives from the typedef identifier and did not change, so no callers moved.

## The fix

`cmd/generate/ffi/interfacenames.go` scans the bundled header text for the
`@name` tag above each `GDExtensionInterface*` typedef and maps it by typedef
identifier. The template now emits `interfaceFunctionName`, which prefers the
`@name` value and falls back to the legacy transform only for typedefs without
metadata. A coverage test pins that every interface typedef in the bundled
header carries `@name` (two exemptions: the `FunctionPtr` alias and
`GetProcAddress`, neither loaded via `LoadProcAddress`), so a future
`make update_godot_headers_from_binary` refresh that introduces an undocumented
typedef fails loudly instead of silently regressing a name.

Why text-scan over comments-in-AST: the clang AST parser drops comments, the
header is machine-generated with a stable doc convention (179 blocks scanned,
zero surprises), and a line scan keeps the AST artifact untouched.

## Godot-side vs Go-side

Everything found was Go-side (our codegen). The engine's interface registry
was complete and correct on every count. `git blame` logic applied to names:
the fixups table was correct the day it was written; the engine added
`class5`/`class6` later and nothing failed when the table stopped covering
reality.

## Learnings

- **"Known missing" is a hypothesis, not a fact.** An inventory diff of
  requested vs registered symbols answered in minutes what the spike would
  have probed speculatively.
- **Reconstructing identifiers from other identifiers is drift bait.** Any
  hand-patched name table silently rots at every upstream addition; deriving
  from upstream's published name makes refreshes self-healing.
- **Extract the "have" set from every registration site**, not the obvious
  file (`classdb_*` registers from `gdextension.cpp`).
- **The mutation test earns the guard its teeth**: stripping one `@name` line
  from a fixture must make the coverage guard name the typedef.

## C1 de-risked

The spike risk ("does the missing `callable_custom_get_user_data` block custom
callables?") is answered: `callable_custom_create2` always loaded, and
`callable_custom_get_userdata` now loads too, so `Callable::get_custom()`
parity (needed later by C2/C4 to unwrap Go callables) has its interface
function available. No blocker remains for `support-go-callable-bridge`.

## Suite result

`GODOT=/home/pcting/bin/godot make test`: gate PASS - 1332 assertions,
0 failures, shutdown marker present, no leaked engine objects, and
zero "Attempt to get non-existent interface function" lines where four
tolerated warning lines used to appear. The tolerated-warnings requirement
in `test-harness-integrity` remains satisfied (vacuously) as designed.
