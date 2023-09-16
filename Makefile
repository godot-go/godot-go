.DEFAULT_GOAL := build

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CLANG_FORMAT?=$(shell which clang-format | which clang-format-10 | which clang-format-11 | which clang-format-12)
GODOT?=$(shell which godot)
CWD=$(shell pwd)

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

.PHONY: goenv generate update_godot_headers_from_binary build clean_src clean remote_debug_test test interactive_test open_demo_in_editor

goenv:
	go env

generate: clean
	go generate
	if [ ! -z "$(CLANG_FORMAT)" ]; then \
		$(CLANG_FORMAT) -i pkg/gdextensionffi/ffi_wrapper.gen.h; \
		$(CLANG_FORMAT) -i pkg/gdextensionffi/ffi_wrapper.gen.c; \
	fi
	go fmt pkg/gdextensionffi/*.gen.go
	go fmt pkg/gdextension/*.gen.go

update_godot_headers_from_binary: ## update godot_headers from the godot binary
	DISPLAY=:0 $(GODOT) --dump-extension-api --headless; \
	mv extension_api.json godot_headers/extension_api.json; \
	DISPLAY=:0 $(GODOT) --dump-gdextension-interface --headless; \
	mv gdextension_interface.h godot_headers/godot/

build: goenv
	CGO_ENABLED=1 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_CFLAGS='-g3 -g -gdwarf -DX86=1 -fPIC -O0' \
	CGO_LDFLAGS='-g3 -g' \
	go build -gcflags=all="-v -N -l -L -clobberdead -clobberdeadreg -dwarf -dwarflocationlists=false" -tags tools -buildmode=c-shared -x -trimpath -o "$(TEST_BINARY_PATH)" $(TEST_MAIN)

clean_src:
	rm -f pkg/gdextensionffi/*.gen.c
	rm -f pkg/gdextensionffi/*.gen.h
	rm -f pkg/gdextensionffi/*.gen.go
	rm -f pkg/gdextension/*.gen.c
	rm -f pkg/gdextension/*.gen.h
	rm -f pkg/gdextension/*.gen.go

clean: clean_src
	rm -f test/demo/lib/libgodotgo-*

remote_debug_test:
	CI=1 \
	LOG_LEVEL=info \
	GOTRACEBACK=crash \e
	GODEBUG=sbrk=1,asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=5 \
	gdbserver --once :55555 $(GODOT) --headless --verbose --debug --path test/demo/

ci_gen_test_project_files:
	CI=1 \
	LOG_LEVEL=info \
	GOTRACEBACK=1 \
	GODEBUG=sbrk=1,gctrace=1,asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=5 \
	$(GODOT) --headless --verbose --path test/demo/ --editor --quit

test:
	CI=1 \
	LOG_LEVEL=info \
	GOTRACEBACK=1 \
	GODEBUG=sbrk=1,gctrace=1,asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=5 \
	$(GODOT) --debug --headless --path test/demo/ --quit

interactive_test:
	LOG_LEVEL=info \
	GOTRACEBACK=1 \
	GODEBUG=sbrk=1,gctrace=1,asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=5 \
	$(GODOT) --verbose --debug --path test/demo/

open_demo_in_editor:
	DISPLAY=:0 \
	LOG_LEVEL=info \
	GOTRACEBACK=1 \
	GODEBUG=sbrk=1,gctrace=1,asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=5 \
	$(GODOT) --verbose --debug --path test/demo/ --editor

godot_unit_test:
	$(GODOT) --test
