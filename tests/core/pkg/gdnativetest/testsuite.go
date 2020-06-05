package gdnativetest

/*
#cgo pkg-config: --define-variable=PROJECTDIR=${SRCDIR}/../../../.. ${SRCDIR}/../../../../godot.pc
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
	os.Exit(0)
}

func runTests() {
	//defer GinkgoRecover()
	log.SetOutput(GinkgoWriter)
	RegisterFailHandler(Fail)
	RunSpecs(GinkgoT(), "Godot Integration Test Suite")
}
