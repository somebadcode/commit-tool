package main

import (
	"context"
	"debug/buildinfo"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/somebadcode/commitlint/internal/commitlinter/defaultlinter"
	"github.com/urfave/cli/v2"

	"github.com/somebadcode/commitlint/internal/commitlinter"
	"github.com/somebadcode/commitlint/internal/revisionflag"
)

const (
	StatusOK = iota
	StatusError
)

func Run(ctx context.Context, bi *buildinfo.BuildInfo) int {
	appCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	revFlag := revisionflag.New("rev", "HEAD")
	otherRevFlag := revisionflag.New("other", "ref/heads/master")

	wd, wdErr := os.Getwd()
	repoPathFlag := &cli.PathFlag{
		Name:        "path",
		Usage:       "git repository directory",
		Required:    wdErr != nil,
		Value:       wd,
		Destination: &wd,
	}

	app := &cli.App{
		Copyright: License,
		Version:   bi.Main.Version,
		Commands: []*cli.Command{
			{
				Name:     "commitlint",
				Category: "lint",
				Aliases:  []string{"lint"},
				Usage:    "lint commits in Git repository",
				Flags:    []cli.Flag{repoPathFlag, revFlag, otherRevFlag},
				Action: func(cCtx *cli.Context) error {
					repo, err := git.PlainOpen(cCtx.String("path"))
					if err != nil {
						return err
					}

					linter := commitlinter.CommitLinter{
						Repo:       repo,
						Rev:        cCtx.Generic("rev").(plumbing.Revision),
						ReportFunc: nil,
						Linter:     defaultlinter.Linter{},
					}

					return linter.Run(ctx)
				},
			},
		},
	}

	if err := app.RunContext(appCtx, os.Args); err != nil {
		return StatusError
	}

	return StatusOK
}
