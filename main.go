package main

import (
	"context"
	_ "embed"
	"os"
	"os/signal"
	"syscall"

	"github.com/somebadcode/commit-tool/cmd"
	_ "github.com/somebadcode/commit-tool/cmd/lint"
)

// //go:embed LICENSE
// var License string

//go:embed NOTICE
var CopyrightNotice string

func main() {
	os.Exit(run())
}

func run() int {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := cmd.RootCommand.ExecuteContext(ctx)
	if err != nil {
		return 1
	}

	return 0
}
