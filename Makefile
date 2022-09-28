.DEFAULT_GOAL := build

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CLANG_FORMAT?=$(shell which clang-format | which clang-format-10 | which clang-format-11 | which clang-format-12)
GODOT?=$(shell which godot)
# godot 4 beta1
GODOT_HEADER_COMMIT_HASH?=62e5472d8e12b6e098f95c5d9f472857d7724a04
CWD=$(shell pwd)

OUTPUT_PATH=test/demo/lib
CGO_ENABLED=1
TEST_MAIN=test/src/example.go

ifeq ($(GOOS),windows)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-windows-4.0-$(GOARCH).dll
else ifeq ($(GOOS),darwin)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-darwin-10.6-$(GOARCH).dylib
else ifeq ($(GOOS),linux)
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-linux-$(GOARCH).so
else
	TEST_BINARY_PATH=$(OUTPUT_PATH)/libgodotgo-test-$(GOOS)-$(GOARCH).so
endif

.PHONY: goenv generate update_godot_headers_from_source update_godot_headers_from_github build clean_src clean remote_debug_test test interactive_test open_demo_in_editor

goenv:
	go env

generate: clean
	go generate
	if [ ! -z "$(CLANG_FORMAT)" ]; then \
		$(CLANG_FORMAT) -i pkg/gdnative/gdnative_wrapper.gen.h; \
		$(CLANG_FORMAT) -i pkg/gdnative/gdnative_wrapper.gen.c; \
	fi

update_godot_headers_from_source: ## update godot_headers to align with the latest godot srouce code on master
	cd godot_headers; \
	DISPLAY=:0 \
	$(GODOT) --dump-extension-api extension_api.json; \
	if [ -z "${GODOT_SRC}" ]; then \
		echo "plase set GODOT_SRC to copy gdnative_interface.h"
		exit 1
	fi \
	cp ${GODOT_SRC}/core/extension/gdnative_interface.h godot_headers/godot/

update_godot_headers_from_github: ## update godot_headers to align with the latest godot 4 beta1 stable binary
	wget https://raw.githubusercontent.com/godotengine/godot-headers/${GODOT_HEADER_COMMIT_HASH}/godot/gdnative_interface.h -O godot_headers/godot/gdnative_interface.h
	wget https://raw.githubusercontent.com/godotengine/godot-headers/${GODOT_HEADER_COMMIT_HASH}/extension_api.json -O godot_headers/extension_api.json

build: goenv
	CGO_ENABLED=1 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_CFLAGS='-Og -ggdb -DX86=1 -fPIC' \
	CGO_LDFLAGS='-Og -ggdb' \
	go build -gcflags=all="-N -l" -tags tools -buildmode=c-shared -x -trimpath -o "$(TEST_BINARY_PATH)" $(TEST_MAIN)

clean_src:
	rm -f pkg/gdnative/*.gen.c
	rm -f pkg/gdnative/*.gen.h
	rm -f pkg/gdnative/*.gen.go
	rm -f pkg/gdextension/*.gen.c
	rm -f pkg/gdextension/*.gen.h
	rm -f pkg/gdextension/*.gen.go

clean: clean_src
	rm -f test/demo/lib/libgodotgo-*

remote_debug_test:
	CI=1 \
	LOG_LEVEL=debug \
	GOTRACEBACK=crash \
	DISPLAY=:0 \
	GODEBUG=asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	gdbserver --once :55555 $(GODOT) --headless --verbose --debug --path test/demo/

test:
	CI=1 \
	LOG_LEVEL=info \
	GOTRACEBACK=crash \
	DISPLAY=:0 \
	LD_DEBUG=libs \
	GODEBUG=asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	$(GODOT) --headless --verbose --path test/demo/

interactive_test:
	LOG_LEVEL=debug \
	GOTRACEBACK=1 \
	DISPLAY=:0 \
	GODEBUG=asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	$(GODOT) --verbose --debug --path test/demo/

open_demo_in_editor:
	LOG_LEVEL=debug \
	GOTRACEBACK=1 \
	DISPLAY=:0 \
	GODEBUG=asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	$(GODOT) --verbose --debug --path test/demo/ --editor
