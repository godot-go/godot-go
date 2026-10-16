# Tasks

## 1. Registration mapping (design D2, D5)

- [x] 1.1 Rewrite the `reflect.Slice` case of `ReflectTypeToGDExtensionVariantType` (`pkg/core/variant_reflect_type.go`) to map elements through the closed D2 table to `PACKED_*` variant types (replacing the element-scalar mis-mapping), with a `log.Panic` naming the element type for unmappable elements; verify `go build ./...` clean
- [x] 1.2 Add Go unit tests pinning all ten table rows (including `[]int`→PACKED_INT64_ARRAY and `[]byte`→PACKED_BYTE_ARRAY) and rejection panics for `[]bool`, `[]Vector2i`, `[]String`, `[][]int`; verify `go test ./pkg/core/ -run VariantType` green

## 2. Slice argument decoding (design D1, D4, D6)

- [x] 2.1 Create `pkg/core/slice_marshalling.go` with cell→Go decode helpers: bulk copy via `Size()` + `operator_index_const(cell, 0)` + `unsafe.Slice` with length-0 null guard for the nine plain-data kinds, per-element `operator_index_const(i)` + `ToUtf8()` for string arrays, and NIL→nil handling; verify `make build` compiles clean
- [x] 2.2 Wire the varcall argument decoder (`pkg/core/method_bind_reflect.go:214`) through the helpers using `ToPacked*K()` owned temporaries destroyed after the element copy including on failure paths; add Go echo fixtures for all ten slice kinds plus main.gd varcall assertions (typed values via `call()`, nil argument → length 0); verify `GODOT=/home/pcting/bin/godot make test` green with the new assertion counts
- [x] 2.3 Wire the ptrcall argument decoder (`pkg/core/method_bind_reflect.go:599`) to read elements through the call-scoped borrow via const operator_index (no container copy, no release); extend main.gd to drive the same fixtures via direct typed calls; verify `make test` green

## 3. Slice return encoding (design D3, D4, D6)

- [x] 3.1 Add Go-slice→cell builders to `pkg/core/slice_marshalling.go`: `NewPacked*K()` + resize + bulk write through mutable `operator_index` for plain data, `Append(String)` per element for string arrays (fall back to the `Array` bridge only if the generated `Append` signature is unusable), nil/empty → empty container; verify `make build` compiles clean
- [x] 3.2 Wire the return encoder (`pkg/core/variant_refect_value.go`) with a `reflect.Slice` case: build cell → generated `Packed*KCopyConstructor(rOut, tmp.NativeConstPtr())` → `Destroy()` temp; add main.gd round-trip assertions for all ten kinds plus nil-`[]byte` return asserting `.is_empty()` and not `null`; verify `make test` green

## 4. Fidelity and independence fixtures (spec requirements 4, 5)

- [x] 4.1 Add fidelity fixtures: UTF-8 `PackedStringArray` round trip (`"héllo"`, `"日本"`), 10,000-element `PackedByteArray` with distinct values, full-range negative `int64`s, `float32`/`float64` precision values, `Vector2`/`Vector3`/`Vector4`/`Color` component equality — driven on both call paths; verify `make test` green with exact element assertions
- [x] 4.2 Add independence fixtures: decoded `[]byte` retained and re-read on a later frame with intact contents (documented against D5a in the archived string-marshalling design), and no-write-back check where Go mutates a decoded slice and GDScript asserts the original array unchanged; verify `make test` green with zero leaked engine objects and the shutdown marker under `GODEBUG=cgocheck=1`
- [x] 4.3 Update the `docs/overview.md` "Packed Arrays" limitation to describe supported `[]T`↔`Packed*K` pairings and rejection behavior; verify the stale "not yet implemented" sentence is gone

## 5. Closeout integration

- [x] 5.1 Run the full gate: `GODOT=/home/pcting/bin/godot make test` → banner, assertion counts, zero failures, shutdown marker, no leaked or orphaned engine objects (ignore the three known `AGENTS.md` interface-function warnings); `openspec validate support-native-slice-marshalling` green
- [x] 5.2 Mark A1 `Done` in `docs/roadmap.md` (phase-table checkbox, DONE note, gantt `done` tag); verify by diff review against the A2 closeout pattern
