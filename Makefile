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

ifeq ($(GOOS),windows)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-windows-$(GOARCH).dll
else ifeq ($(GOOS),darwin)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-macos-$(GOARCH).framework
else ifeq ($(GOOS),linux)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-linux-$(GOARCH).so
else
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-$(GOOS)-$(GOARCH).so
endif

.PHONY: goenv installdeps generate update_godot_headers_from_binary build build-full clean_src clean remote_debug_test ci_gen_test_project_files test interactive_test open_demo_in_editor

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

update_godot_headers_from_binary: ## update godot_headers from the godot binary
	DISPLAY=:0 "$(GODOT)" --dump-extension-api --headless; \
	mv extension_api.json godot_headers/extension_api.json; \
	DISPLAY=:0 "$(GODOT)" --dump-gdextension-interface --headless; \
	mv gdextension_interface.h godot_headers/godot/

# https://medium.com/@aviad.hayumi/debug-golang-applications-with-c-c-bindings-f254b3f1259e
build: goenv ## Build GDExtension library (debug mode)
	CGO_ENABLED=1 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_CFLAGS='-fPIC -g -ggdb -O0' \
	CGO_LDFLAGS='-g3 -g -O0' \
	GOEXPERIMENT=$(GOEXPERIMENT) \
	go build -gcflags=all="-N -l" -tags tools -buildmode=c-shared -v -x -trimpath -o "$(TEST_BINARY_PATH)" $(TEST_MAIN)

build-full: goenv ## Build GDExtension library with full debug symbols
	CGO_ENABLED=1 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_CFLAGS='-g3 -g -gdwarf -DX86=1 -fPIC -O0' \
	CGO_LDFLAGS='-g3 -g' \
	GOEXPERIMENT=$(GOEXPERIMENT) \
	go build -gcflags="-N -l" -ldflags=-compressdwarf=0 -tags tools -buildmode=c-shared -v -x -trimpath -o "$(TEST_BINARY_PATH)" $(TEST_MAIN)

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

test: ## Run headless tests
	CI=1 \
	LOG_LEVEL=WARN \
	GOTRACEBACK=single \
	GODEBUG=gctrace=1,asyncpreemptoff=1,cgocheck=1,invalidptr=1,clobberfree=1 \
	"$(GODOT)" --headless --verbose --path test/demo/ --quit

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
