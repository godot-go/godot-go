.DEFAULT_GOAL := build

OUTPUT_PATH=test/libs
CGO_ENABLED=1
GOOS?=$(shell go env GOOS)
GOARCH?=$(shell go env GOARCH)
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

.PHONY: goenv generate updateextensionapi build clean_gdnative clean test interactivetest

goenv:
	go env

generate: clean
	go generate
	clang-format -i pkg/gdnative/gdnative_wrapper.gen.h
	clang-format -i pkg/gdnative/gdnative_wrapper.gen.c

updateextensionapi:
	cd godot_headers; \
	godot --dump-extension-api extension_api.json; \
	echo "**** remember to copy gdnative_interface.h from godot ****"

build: goenv
	CGO_ENABLED=1 GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_CFLAGS='-Og -g' CGO_LDFLAGS='-Og -g' go build -gcflags=all="-N -l" -tags tools -buildmode=c-shared -x -trimpath \
		-o "$(TEST_BINARY_PATH)" $(TEST_MAIN)

clean_gdnative:
	rm -f pkg/gdnative/*.gen.c
	rm -f pkg/gdnative/*.gen.h
	rm -f pkg/gdnative/*.gen.go

clean: clean_gdnative

test:
	CI=1 \
	LOG_LEVEL=debug \
	GOTRACEBACK=crash \
	GODEBUG=asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	godot --headless --verbose --debug --path test/demo/ -glog=3

interactivetest:
	LOG_LEVEL=debug \
	GOTRACEBACK=1 \
	GODEBUG=asyncpreemptoff=1,cgocheck=0,invalidptr=1,clobberfree=1,tracebackancestors=0 \
	godot --verbose --debug --path test/demo/ -glog=3
