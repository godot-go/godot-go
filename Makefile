.DEFAULT_GOAL := build

GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
CLANG_FORMAT?=$(shell which clang-format | which clang-format-10 | which clang-format-11 | which clang-format-12)
GODOT?=$(shell which godot)
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

.PHONY: goenv generate dumpextensionapi build clean_gdnative clean test interactivetest

goenv:
	go env

generate: clean
	go generate
	if [ ! -z "$(CLANG_FORMAT)" ]; then \
		$(CLANG_FORMAT) -i pkg/gdnative/gdnative_wrapper.gen.h; \
		$(CLANG_FORMAT) -i pkg/gdnative/gdnative_wrapper.gen.c; \
	fi

dumpextensionapi:
	cd godot_headers; \
	DISPLAY=:0 \
	$(GODOT) --dump-extension-api extension_api.json; \
	echo "**** remember to run cp \${GODOT_SRC}/core/extension/gdnative_interface.h godot_headers/godot/"
	echo "**** alternatively, you can visit https://github.com/godotengine/godot-headers for the latest stable headers ****"

build: goenv
	CGO_ENABLED=1 \
	GOOS=$(GOOS) \
	GOARCH=$(GOARCH) \
	CGO_CFLAGS='-Og -ggdb -DX86=1 -fPIC' \
	CGO_LDFLAGS='-Og -ggdb' \
	go build -gcflags=all="-N -l" -tags tools -buildmode=c-shared -x -trimpath -o "$(TEST_BINARY_PATH)" $(TEST_MAIN)

clean_gdnative:
	rm -f pkg/gdnative/*.gen.c
	rm -f pkg/gdnative/*.gen.h
	rm -f pkg/gdnative/*.gen.go
	rm -f pkg/gdextension/*.gen.c
	rm -f pkg/gdextension/*.gen.h
	rm -f pkg/gdextension/*.gen.go
	rm -f test/demo/lib/libgodotgo-*

clean: clean_gdnative

remotedebugtest:
	CI=1 \
	LOG_LEVEL=debug \
	GOTRACEBACK=crash \
	DISPLAY=:0 \
	GODEBUG=asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	gdbserver --once :55555 $(GODOT) --headless --verbose --debug --path $(CWD)/test/demo/

test:
	CI=1 \
	LOG_LEVEL=debug \
	GOTRACEBACK=crash \
	DISPLAY=:0 \
	GODEBUG=asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	$(GODOT) --headless --verbose --debug --path test/demo/

interactivetest:
	LOG_LEVEL=debug \
	GOTRACEBACK=1 \
	DISPLAY=:0 \
	GODEBUG=asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	$(GODOT) --verbose --debug --path test/demo/

opendemoeditor:
	LOG_LEVEL=debug \
	GOTRACEBACK=1 \
	DISPLAY=:0 \
	GODEBUG=asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	$(GODOT) --verbose --debug --path test/demo/ --editor
