# Tasks

## 1. Await state and core primitive

- [x] 1.1 Create `pkg/core/signal_await.go` with `AwaitResult`, `AwaitOpaque`, and an engine-free await-state (buffered cap-1 channel, `sync.Once` close, deliver-then-close) plus a pure-Go unit test `TestSignalAwaitState` in `pkg/core/signal_await_test.go` covering delivery, close-without-send, and double-close safety; verify with `go test ./pkg/core -run TestSignalAwaitState`
- [x] 1.2 Implement `Await(obj Object, signalName string) <-chan AwaitResult`: build a `NewGoCallable` whose callback delivers into the await-state, `Connect` with `CONNECT_ONE_SHOT`, then immediately `Destroy()` the Go Callable reference (D2); wire the first harness scenario (GDScript emits `hit(4, "critical")` → goroutine receives both args in order, channel closed); verify via a new `test/pkg/signal_await_tests.go` scenario registered like the go_callable tests and run through `GODOT=/home/pcting/bin/godot make test`
- [x] 1.3 Implement the D4 marshalling table (nil/bool/int64/float64/String/StringName/NodePath/math POD/objects/nested Array+Dictionary with depth cap/opaque fallback) used by the callback; add harness scenarios asserting Go-native payload types, object-handle identity, nested container contents, and an opaque `Callable` payload; verify through `make test`
- [x] 1.4 Implement and verify one-shot semantics (D2 connect-then-destroy already arms one-shot): harness scenario emits twice on the same signal, first channel delivers only the first emission, a fresh `Await` receives the next emission, `GoCallableRegistryCount` returns to baseline after each delivery; verify through `make test`
- [x] 1.5 Implement the misuse path (non-OK `Connect` → `log.Panic` naming class + signal, no dangling subscription) and a harness scenario invoking `Await` with an unknown signal inside a `recover()` wrapper asserting the diagnostic and unchanged registry count; verify through `make test`

## 2. Liveness paths

- [x] 2.1 Implement free-driven close (await-state close-on-free via the callable's free callback; D2/D3 guarantee it fires only when the connection dies) and a harness scenario: await a fresh Node, free the node from GDScript, receiver unblocks with `ok == false`, registry drains to baseline; verify through `make test`
- [x] 2.2 Add a shutdown-with-pending-await harness scenario: arm an `Await` on a live signal near end of run and never satisfy it; verify the run still reports `TESTS FINISHED`, zero failures, no ObjectDB leaks, and a clean main-loop shutdown marker per `test/check_test_output.sh` (same green-gate criterion the C1 "Shutdown with live and dead callables" scenario established — the script contains no registry-count probe)

## 3. Timer path and documentation

- [x] 3.1 Add the SceneTreeTimer harness scenario: `Await` on `timeout` of a 1-second timer created from GDScript (or Go singleton access) delivers within the test window; verify through `make test`
- [x] 3.2 Write the `Await` doc comment covering the `select`/timeout/context composition idiom, the `v, ok := <-ch` freed-object convention, and the documented threading hazards (payload delivery off-main, engine calls after await); verify with `gofmt -l pkg/core` clean and review of rendered `go doc` output
- [x] 3.3 Update `docs/roadmap.md` C2 status and add a short usage section (await idiom + L1 Go-closure-to-signal connect pattern) to the project docs where extension features are described; verify by diff review

## 4. Integration

- [x] 4.1 Run the full gate: `GODOT=/home/pcting/bin/godot make test` green per `test/check_test_output.sh`, plus `revive -config .revive.toml` on the touched packages (no new violation categories; the repo has no `lint` make target or CI step) and `go vet` clean on `./pkg/core ./test/pkg` (repo-wide vet noise is confined to untouched pre-existing `test/concepts` scratch code)
