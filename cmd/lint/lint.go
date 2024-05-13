package lint

import (
	"context"
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/urfave/cli/v2"

	"github.com/somebadcode/commit-tool/linter"
	"github.com/somebadcode/commit-tool/slogctx"
	"github.com/somebadcode/commit-tool/traverser"
)

type Command struct {
	Repository    string
	Revision      plumbing.Revision
	OtherRevision plumbing.Revision
}

func (cmd *Command) Execute(ctx context.Context, args []string) error {
	var repoPath string
	if len(args) == 1 {
		repoPath = args[0]
	} else {
		wd, wdErr := os.Getwd()
		if wdErr != nil {
			return fmt.Errorf("repository path required")
		}

		repoPath = wd
	}

	trav := traverser.Traverser{
		Rev:        cmd.Revision,
		OtherRev:   cmd.OtherRevision,
		ReportFunc: traverser.SlogReporter(slogctx.L(ctx)),
		VisitFunc: linter.Linter{
			Rules: linter.Rules{
				linter.RuleConventionalCommit,
			},
		}.Lint,
	}

	var err error

	trav.Repo, err = git.PlainOpen(repoPath)
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
	}
}

func revisionFlagAction(dest *plumbing.Revision) func(*cli.Context, string) error {
	return func(c *cli.Context, s string) error {
		*dest = plumbing.Revision(s)

		return nil
	}
}
