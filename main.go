package main

import (
	"context"
	_ "embed"
	"os"
)

// //go:embed LICENSE
// var License string

//go:embed NOTICE
var CopyrightNotice string

func main() {
	os.Exit(Run(context.Background()))
}
