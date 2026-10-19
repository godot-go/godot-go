.DEFAULT_GOAL := build

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
GOIMPORTS?=$(shell which goimports)
CLANG_FORMAT?=$(shell which clang-format || which clang-format-10 || which clang-format-11 || which clang-format-12)
GODOT?=$(shell which godot)
CWD=$(shell pwd)
# GOEXPERIMENT="cgocheck2=0"

OUTPUT_PATH=test/demo/lib
CGO_ENABLED=1
TEST_MAIN=test/main.go
# Extra flags appended to CGO_CFLAGS / CGO_LDFLAGS in build and build-full.
# Example (Alpine/musl C backtraces via libexecinfo):
#   make build EXTRA_CGO_CFLAGS='-DGODOT_GO_ENABLE_EXECINFO' EXTRA_CGO_LDFLAGS='-lexecinfo'
EXTRA_CGO_CFLAGS?=
EXTRA_CGO_LDFLAGS?=

ifeq ($(GOOS),windows)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-windows-$(GOARCH).dll
else ifeq ($(GOOS),darwin)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-macos-$(GOARCH).framework
else ifeq ($(GOOS),linux)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-linux-$(GOARCH).so
else
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-$(GOOS)-$(GOARCH).so
endif

# Debug info for the build target. Global C flags stay DWARF on every
# platform; Windows native backtraces go through dbghelp.dll, which reads
# only PDB/COFF, so on windows the link additionally produces a PDB pinned
# by absolute path next to the DLL (-Wl,--pdb=, GNU ld >= 2.40). The one
# package whose C frames appear in captured backtraces and compiles safely,
# pkg/log, opts into CodeView with a `#cgo windows CFLAGS: -gcodeview`
# directive rather than a global flag: dwarf2codeview in MinGW-Builds
# GCC 15.2.0 ICEs on the generated pkg/ffi TUs and on every pkg/core cgo
# preamble TU. Other C code keeps DWARF in the DLL for offline resolution.
# See openspec/changes/windows-pdb-line-info/.
DEBUG_INFO_CFLAGS=-ggdb
DEBUG_INFO_LDFLAGS=
ifeq ($(GOOS),windows)
	# binutils 2.45.1 (the runner's ld) emits a malformed DBI stream from
	# --pdb (binutils bug 34605: file-info substream not 4-byte aligned);
	# dbghelp refuses to load that PDB, so live line lookups silently fail.
	# The fix is on binutils master only, not any mingw release. lld is the
	# reference PDB writer and produces a PDB dbghelp loads, so the windows
	# test link goes through lld instead. /DEBUG must be slash-form (the
	# lld COFF spelling); --pdb= works for both linkers.
	DEBUG_INFO_LDFLAGS=-fuse-ld=lld -Wl,/DEBUG -Wl,--pdb=$(CURDIR)/$(OUTPUT_PATH)/$(patsubst %.dll,%.pdb,$(notdir $(TEST_BINARY_PATH)))
endif

.PHONY: goenv installdeps generate fmt update_godot_headers_from_binary build build-full clean_src clean remote_debug_test ci_gen_test_project_files test test_delegation_trap interactive_test open_demo_in_editor

goenv: ## Print Go environment variables
	go env

installdeps: ## Install Go dependencies (goimports)
	go install golang.org/x/tools/cmd/goimports@latest

generate: clean ## Generate Go bindings from Godot API
	go generate
	if [ ! -z "$(CLANG_FORMAT)" ]; then \
		"$(CLANG_FORMAT)" -i pkg/ffi/ffi_wrapper.gen.h; \
		"$(CLANG_FORMAT)" -i pkg/ffi/ffi_wrapper.gen.c; \
	fi
	find pkg -name *.gen.go -exec go fmt {} \;
	if [ ! -z "$(GOIMPORTS)" ]; then \
		find pkg -name *.gen.go -exec $(GOIMPORTS) -w {} \; ; \
	fi

fmt: ## Run go fmt
	go fmt ./...

update_godot_headers_from_binary: ## update godot_headers from the godot binary
	DISPLAY=:0 "$(GODOT)" --dump-extension-api --headless; \
	mv extension_api.json godot_headers/extension_api.json; \
	DISPLAY=:0 "$(GODOT)" --dump-gdextension-interface --headless; \
	mv gdextension_interface.h godot_headers/godot/

# https://medium.com/@aviad.hayumi/debug-golang-applications-with-c-c-bindings-f254b3f1259e
build: goenv ## Build GDExtension library (debug mode)
	# On windows the external linker writes the --pdb file during linking,
	# before go build creates the -o destination's parent directory; ld's
	# ENOENT there is only a warning and silently drops the PDB, so the
	# directory must exist first.
	@mkdir -p "$(OUTPUT_PATH)"
	CGO_ENABLED=1 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	GOAMD64=v3 \
	CGO_CFLAGS='-fno-omit-frame-pointer -fPIC -g $(DEBUG_INFO_CFLAGS) -O0 -I$(CWD) $(EXTRA_CGO_CFLAGS)' \
	CGO_LDFLAGS='-g3 -g -O0 $(DEBUG_INFO_LDFLAGS) $(EXTRA_CGO_LDFLAGS)' \
	GOEXPERIMENT=$(GOEXPERIMENT) \
	go build -gcflags=all="-N -l" -tags tools -buildmode=c-shared -v -x -trimpath -o "$(TEST_BINARY_PATH)" $(TEST_MAIN)

build-full: goenv ## Build GDExtension library with full debug symbols
	CGO_ENABLED=1 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	GOAMD64=v3 \
	CGO_CFLAGS='-fno-omit-frame-pointer -g3 -g -gdwarf -DX86=1 -fPIC -O0 -I$(CWD) $(EXTRA_CGO_CFLAGS)' \
	CGO_LDFLAGS='-g3 -g $(EXTRA_CGO_LDFLAGS)' \
	GOEXPERIMENT=$(GOEXPERIMENT) \
	go build -gcflags="-N -l" -ldflags="-compressdwarf=0" -tags tools -buildmode=c-shared -v -x -trimpath -o "$(TEST_BINARY_PATH)" $(TEST_MAIN)

clean_src: ## Remove generated source files
	find pkg -name *.gen.go -delete
	find pkg -name *.gen.c -delete
	find pkg -name *.gen.h -delete

clean: clean_src ## Clean generated files and binaries
	rm -f test/demo/lib/libgodotgo-*

remote_debug_test: ## Start Godot with gdbserver for remote debugging
	CI=1 \
	LOG_LEVEL=info \
	GOTRACEBACK=crash \
	GODEBUG=asyncpreemptoff=1,cgocheck=1,invalidptr=1,clobberfree=1,tracebackancestors=5 \
	gdbserver --once :55555 "$(GODOT)" --headless --verbose --debug --path test/demo/

ci_gen_test_project_files: ## Generate test project files for CI
	CI=1 \
	LOG_LEVEL=info \
	GOTRACEBACK=1 \
	GODEBUG=gctrace=1,asyncpreemptoff=1,cgocheck=1,invalidptr=1,clobberfree=1,tracebackancestors=5 \
	"$(GODOT)" --headless --verbose --path test/demo/ --editor --quit
	# hack until fix lands: https://github.com/godotengine/godot/issues/84460
	if [ ! -f "test/demo/.godot/extension_list.cfg" ]; then \
		echo 'res://example.gdextension' >> test/demo/.godot/extension_list.cfg; \
	fi

test: build ## Run headless tests: runtime suite, then the editor-plugin run (green only if both gates pass)
	@status=0; \
	bash -o pipefail -ec ' \
		CI=1 \
		LOG_LEVEL=WARN \
		GOTRACEBACK=single \
		GODEBUG=gctrace=1,asyncpreemptoff=1,cgocheck=1,invalidptr=1,clobberfree=1 \
		"$(GODOT)" --headless --verbose --path test/demo/ --quit-after 12 2>&1 | tee $(CURDIR)/test-output.log' || status=$$?; \
	bash $(CURDIR)/test/check_test_output.sh $(CURDIR)/test-output.log $$status || exit 1; \
	edstatus=0; \
	bash -o pipefail -ec ' \
		CI=1 \
		LOG_LEVEL=WARN \
		GOTRACEBACK=single \
		GODEBUG=asyncpreemptoff=1,cgocheck=1,invalidptr=1 \
		"$(GODOT)" --headless --verbose --path test/demo/ --editor --quit-after 12 2>&1 | tee $(CURDIR)/editor-output.log' || edstatus=$$?; \
	bash $(CURDIR)/test/check_editor_output.sh $(CURDIR)/editor-output.log $$edstatus

test_gate: ## Run the fixture tests for the test-output gate itself (no Godot needed)
	@bash $(CURDIR)/test/fixtures/run_fixture_tests.sh

check_type_rejection: build ## Assert the engine rejects a wrong-class typed argument (negative fixture)
	@out="$$($(GODOT) --headless --check-only --script res://type_rejection.gd --path test/demo/ 2>&1)"; \
	status=$$?; \
	if [ $$status -eq 0 ]; then \
		echo "FAIL: type_rejection.gd parsed cleanly; the advertised argument classes are no longer enforced"; \
		exit 1; \
	fi; \
	if ! printf '%s' "$$out" | grep -q 'should be "CollisionShape2D" but is "Image"'; then \
		echo "FAIL: expected a CollisionShape2D-vs-Image rejection; got:"; \
		printf '%s\n' "$$out" | tail -10; \
		exit 1; \
	fi; \
	if ! printf '%s' "$$out" | grep -q 'should be "Shape2D" but is "Image"'; then \
		echo "FAIL: expected the Ref pointee class Shape2D to be advertised; got:"; \
		printf '%s\n' "$$out" | tail -10; \
		exit 1; \
	fi; \
	echo "PASS: engine rejects wrong-class typed arguments (CollisionShape2D and Ref pointee Shape2D both enforced)"

test_delegation_trap: build ## Expect the delegating virtual repro to abort godot with the bounded-depth recursion diagnostic (non-zero exit)
	@if [ ! -x "$(GODOT)" ]; then \
		echo "FAIL: godot binary not found; set GODOT=/path/to/godot"; \
		exit 1; \
	fi
	@CI=1 \
	LOG_LEVEL=WARN \
	GOTRACEBACK=single \
	GODEBUG=asyncpreemptoff=1,cgocheck=1,invalidptr=1,clobberfree=1 \
	"$(GODOT)" --headless --verbose --path test/demo/ res://delegation_trap.tscn > $(CURDIR)/delegation-trap-output.log 2>&1; \
	status=$$?; \
	if [ $$status -eq 0 ]; then \
		echo "FAIL: expected non-zero exit from the delegation repro (trap no longer aborts)"; \
		exit 1; \
	fi; \
	if grep -q "delegation recursion confirmed" $(CURDIR)/delegation-trap-output.log; then \
		echo "PASS: bounded-depth recursion trap pinned"; \
	else \
		echo "FAIL: recursion diagnostic missing from delegation-trap-output.log"; \
		tail -30 $(CURDIR)/delegation-trap-output.log; \
		exit 1; \
	fi

interactive_test: ## Run Godot editor with debugging for interactive testing
	LOG_LEVEL=info \
	GOTRACEBACK=1 \
	GODEBUG=gctrace=1,asyncpreemptoff=1,cgocheck=1,invalidptr=1,clobberfree=1,tracebackancestors=5 \
	"$(GODOT)" --verbose --debug --path test/demo/

open_demo_in_editor: ## Open demo project in Godot editor with debugging
	DISPLAY=:0 \
	LOG_LEVEL=info \
	GOTRACEBACK=1 \
	GODEBUG=gctrace=1,asyncpreemptoff=1,cgocheck=1,invalidptr=1,clobberfree=1,tracebackancestors=5 \
	"$(GODOT)" --verbose --debug --path test/demo/ --editor

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
