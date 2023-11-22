package main

import (
	"context"
	_ "embed"
	"os"
	"runtime/debug"
)

//go:embed LICENSE.txt
var License string

func main() {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		buildInfo = &debug.BuildInfo{}
	}

	os.Exit(Run(context.Background(), buildInfo))
}
