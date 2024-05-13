package main

import (
	_ "embed"
	"os"

	"github.com/somebadcode/commit-tool/cmd"
)

// //go:embed LICENSE
// var License string

//go:embed NOTICE
var CopyrightNotice string

func main() {
	os.Exit(cmd.Execute())
}
