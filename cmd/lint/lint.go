package lint

import (
	"context"
	"fmt"
	"github.com/somebadcode/commit-tool/slognop"
	"log/slog"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/urfave/cli/v2"

	"github.com/somebadcode/commit-tool/linter"
	"github.com/somebadcode/commit-tool/traverser"
)

type Command struct {
	RepositoryPath string
	Revision       plumbing.Revision
	OtherRevision  plumbing.Revision
	Logger         *slog.Logger
}

func (cmd *Command) Validate() error {
	if cmd.Logger == nil {
		cmd.Logger = slognop.New()
	}

	if cmd.RepositoryPath == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("repository path required: %w", err)
		}

		cmd.RepositoryPath = wd
	}

	return nil
}

func (cmd *Command) Execute(ctx context.Context) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	trav := traverser.Traverser{
		Rev:        cmd.Revision,
		OtherRev:   cmd.OtherRevision,
		ReportFunc: traverser.SlogReporter(cmd.Logger),
		VisitFunc: linter.Linter{
			Rules: linter.Rules{
				linter.RuleConventionalCommit,
			},
		}.Lint,
		Logger: cmd.Logger,
	}

	var err error
	trav.Repo, err = git.PlainOpen(cmd.RepositoryPath)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	return trav.Run(ctx)
}

func (cmd *Command) Flags() []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:     "rev",
			Category: "revision",
			Usage:    "git revision to start at",
			Value:    "HEAD",
			Action:   revisionFlagAction(&cmd.Revision),
		},
		&cli.StringFlag{
			Name:     "other",
			Category: "revision",
			Usage:    "git revision to stop linting at",
			Action:   revisionFlagAction(&cmd.Revision),
		},
		&cli.PathFlag{
			Name:     "repository",
			Category: "repository",
			DefaultText: func() string {
				s, _ := os.Getwd()
				return s
			}(),
			Usage:       "path to git repository root directory",
			Destination: &cmd.RepositoryPath,
			Aliases:     []string{"repo"},
			EnvVars:     nil,
			TakesFile:   false,
			Action:      nil,
		},
	}
}

func revisionFlagAction(dest *plumbing.Revision) func(*cli.Context, string) error {
	return func(c *cli.Context, s string) error {
		*dest = plumbing.Revision(s)

		return nil
	}
}
