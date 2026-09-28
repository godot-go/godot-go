# Tasks

## 1. Fix the decoder

- [x] 1.1 Replace the `reflect.Zero(t).Interface()` type switch in the `reflect.Interface` case with the `switch { case t.Implements(gdObjectType): … case t.Implements(refType): … default: … }` form already used at ~1076. Object logic moved into the `gdObjectType` case, Ref logic into the `refType` case, `default` still panics naming the argument index and type.
- [x] 1.2 Dereference the argument cell: `objPtr := *(*GDExtensionObjectPtr)(arg)`. See design D5. Falsified in 2.3 — without it the run segfaults.
- [x] 1.3 Null guard before `object_get_class_name`: assign `reflect.Zero(t)` and return.
- [x] 1.4 `Ref` branch decoding semantics untouched; only the redundant `if t.Implements(refType)` nesting level was removed.

## 2. Make the path testable, then test it

- [x] 2.1 `DecodePtrcallArgs` seam in `pkg/core/ptrcall_decode_seam.go`, documented with the reason: the path is otherwise reachable only through typed GDScript call sites, which a sibling change gates.
- [x] 2.2 In-engine tests in `test/pkg/ptrcall_object_decode_tests.go`, driven from `test/demo/main.gd`. Objects arrive via the varcall path (works today) and each test rebuilds the ptrcall cell by hand so the shape under test is explicit. All five green:
  - plain object — `decoded instance 27984397808`, instance ID matches
  - subclass polymorphism — `Shape2D parameter received instance … as CircleShape2D`
  - null — `null cell decoded to a nil Node without touching the engine`
  - Ref regression guard — `RefShape2D still resolves instance …`
  - undecodable — `refused loudly: unsupported interface type for argument 0: pkg.notAnObjectAtAll`
  - GDScript vars are untyped on purpose; a typed call site would be rejected at parse by the metadata defect this change is *not* about.
- [x] 2.3 Falsified both defects independently.
  - **Dispatch reverted to unreachable** (`case false && t.Implements(gdObjectType)`): all three object tests panic with `unsupported interface type for argument 0: builtin.Node` / `builtin.Shape2D`, `FAILURES: 3`, `make` exit 2. The Ref regression test still passed, confirming it was never broken.
  - **Cell dereference removed** (`objPtr := GDExtensionObjectPtr(arg)`): `handle_crash: Program crashed with signal 11`, godot exit 134, no driver summary. This confirms design D5 — the second defect alone is worse than the original loud panic, and both had to be fixed together.
- [x] 2.4 (added during implementation) The panic value now carries the argument index and type name, not just the bare message. zap puts structured fields in the log but not in the panic value, so a caller recovering from the panic could not tell what was refused. The spec's guarantee is now testable at the panic boundary rather than only by scraping log output.

## 3. Verify

- [x] 3.1 `go build ./...` clean; `go vet ./pkg/... ./test/pkg/...` clean.
- [x] 3.2 `make generate` produces **no diff in any `*.gen.*` file** — this change touches no template.
- [x] 3.3 `GODOT=/home/pcting/bin/godot make test` — **1091 assertions, 0 failures, 0 leaked instances**, exit 0.
- [x] 3.4 Sibling `fix-object-argument-type-metadata` stash applied on top cleanly, and its typed `var child: Node` plus `test_object_arg_add_child(child)` no longer panics — the run went green at 1093 assertions with typed object arguments parsing **and** decoding. Verified as part of landing that change; its own verification lives in its own tasks.
