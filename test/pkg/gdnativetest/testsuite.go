package gdnativetest

/*
#cgo CFLAGS: -DX86=1 -g -fPIC -std=c99 -I${SRCDIR}/../../../godot_headers -I${SRCDIR}/../../../pkg/gdnative
#include <cgo_example.h>
#include <stdlib.h>
*/
import "C"
import (
	"os"

	"github.com/godot-go/godot-go/pkg/gdnative"
	"github.com/godot-go/godot-go/pkg/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	gdnative.RegisterInitCallback(initNativescript)
}

func initNativescript() {
	log.Debug("initNativescript called")
	// runTests()

	// TODO: replace this with a better godot way of exiting
	if v, _ := os.LookupEnv("CI"); v == "1" {
		os.Exit(0)
	}
}

func runTests() {
	isSet := func(name string) bool {
		v, _ := os.LookupEnv(name)
		return v == "1"
	}

	if isSet("TEST_USE_GINKGO_RECOVER") {
		defer GinkgoRecover()
	}

	if isSet("TEST_USE_GINKGO_WRITER") {
		log.SetWriteSyncer(GinkgoWriter)
	}

	RegisterFailHandler(Fail)
	RunSpecs(GinkgoT(), "Godot Integration Test Suite")
}
