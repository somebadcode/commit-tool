package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	"github.com/go-git/go-git/v5"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/somebadcode/commit-tool/internal/commitlinter"
	"github.com/somebadcode/commit-tool/internal/commitlinter/defaultlinter"
	"github.com/somebadcode/commit-tool/internal/logging"
	"github.com/somebadcode/commit-tool/internal/revisionflag"
)

const (
	StatusOK = iota
	StatusError
)

func Run(ctx context.Context) int {
	appCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer cancel()

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		panic("build information is missing")
	}

	logger, _ := logging.New(os.Stderr, zap.InfoLevel)

	wd, wdErr := os.Getwd()
	flags := []cli.Flag{
		revisionflag.New("rev", "HEAD"),
		revisionflag.New("until", ""),
		&cli.PathFlag{
			Name:        "path",
			Usage:       "git repository directory",
			Required:    wdErr != nil,
			Value:       wd,
			Destination: &wd,
		},
		&cli.BoolFlag{
			Name:        "allow-initial-commit",
			DefaultText: "disabled",
		},
	}

	app := &cli.App{
		Name:      "conventional-commits-tool",
		Copyright: CopyrightNotice,
		Version:   buildInfo.Main.Version,
		Commands: []*cli.Command{
			{
				Name:     "lint",
				Category: "lint",
				Usage:    "lint commits in Git repository",
				Flags:    flags,
				Action: func(cmdCtx *cli.Context) error {
					repo, err := git.PlainOpen(cmdCtx.String("path"))
					if err != nil {
						return err
					}

					linter := commitlinter.CommitLinter{
						Repo:     repo,
						Rev:      cmdCtx.Value("rev").(*revisionflag.Revision).Revision(),
						UntilRev: cmdCtx.Value("until").(*revisionflag.Revision).Revision(),
						ReportFunc: func(err error) {
							var lintError commitlinter.LintError
							if errors.As(err, &lintError) {
								logger.Error("bad commit message",
									zap.Stringer("hash", lintError.Hash),
									zap.Int("pos", lintError.Pos),
									zap.Error(errors.Unwrap(err)),
								)
							} else {
								logger.Error("bad commit message",
									zap.Error(err),
								)
							}
						},
						Linter: defaultlinter.New(),
					}

					// cmdCtx.Bool("allow-initial-commit")

					return linter.Run(cmdCtx.Context)
				},
			},
		},
	}

	if err := app.RunContext(appCtx, os.Args); err != nil {
		return StatusError
	}

	return StatusOK
}
