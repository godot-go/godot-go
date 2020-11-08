package gdnativetest

import (
	"github.com/godot-go/godot-go/pkg/gdnative"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KinematicBody2D", func() {
	Context("MoveAndCollide()", func() {
		It("returns nil for no collision", func() {
			n := gdnative.NewKinematicBody2D()

			tree := gdnative.NewSceneTree()
			scene := tree.GetCurrentScene()
			scene.AddChild(n, true)

			v := gdnative.NewVector2(1.0, 0.0)

			collision := n.MoveAndCollide(v, true, true, true)

			Ω(collision).Should(BeNil())
		})

		// TODO: help wanted
		XIt("returns a collision", func() {
		})
	})
})
