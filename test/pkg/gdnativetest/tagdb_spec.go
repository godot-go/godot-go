package gdnativetest

/*
#include <cgo_example.h>
#include <gdnative.wrapper.gen.h>
#include <stdlib.h>
*/
import "C"
import (
	"github.com/godot-go/godot-go/pkg/gdnative"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TagDB", func() {
	It("should exist through TagDB.ClassExists", func() {
		Ω(gdnative.GetSingletonClassDB().IsClassEnabled("KinematicBody2D")).Should(BeTrue())
		Ω(gdnative.GetSingletonClassDB().CanInstance("KinematicBody2D")).Should(BeTrue())
		Ω(gdnative.GetSingletonClassDB().ClassExists("KinematicBody2D")).Should(BeTrue())
	})
})
