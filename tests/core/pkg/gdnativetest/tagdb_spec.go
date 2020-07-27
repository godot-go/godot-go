package gdnativetest

/*
#include <cgo_example.h>
#include <gdnative.wrappergen.h>
#include <stdlib.h>
*/
import "C"
import (
	"github.com/pcting/godot-go/pkg/gdnative"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TagDB", func() {
	It("should exist through TagDB.ClassExists", func() {
		str := gdnative.NewStringFromGoString("KinematicBody2D")
		Ω(gdnative.GetSingletonClassDB().IsClassEnabled(str)).Should(BeTrue())
		Ω(gdnative.GetSingletonClassDB().CanInstance(str)).Should(BeTrue())
		Ω(gdnative.GetSingletonClassDB().ClassExists(str)).Should(BeTrue())
	})
})
