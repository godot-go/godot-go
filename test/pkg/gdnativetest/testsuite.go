package gdnativetest

/*
#cgo pkg-config: --define-variable=PROJECTDIR=${SRCDIR}/../../.. ${SRCDIR}/../../../godot.pc
#include <cgo_example.h>
#include <stdlib.h>
*/
import "C"
import (
	"os"

	"github.com/pcting/godot-go/pkg/gdnative"
	"github.com/pcting/godot-go/pkg/log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func init() {
	gdnative.RegisterInitCallback(initNativescript)
}

func initNativescript() {
	log.SetLevel(log.DebugLevel)
	log.Trace("initNativescript called")
	runTests()

	// TODO: replace this with a better godot way of exiting
	os.Exit(0)
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
		log.SetOutput(GinkgoWriter)
	}

	RegisterFailHandler(Fail)
	RunSpecs(GinkgoT(), "Godot Integration Test Suite")
}
