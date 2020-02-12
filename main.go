package main

import (
	"godot-go/pkg/godot"
	"log"
)

var (
	//GitTag is meant to be replaced through the linker: -ldflags -Xmain.GitTag=XXX
	GitTag = "untagged"

	//GitCommit is meant to be replaced through the linker: -ldflags -Xmain.GitCommit=XXX
	GitCommit = "HEAD"
)

func init() {
	log.Printf("Godot API Version %d", godot.GODOTAPIVERSION)
	log.Printf("Loading libopenpf2 (%s:%s)...", GitTag, GitCommit)
}

func main() {
}
