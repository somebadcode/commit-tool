package lint

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/spf13/cobra"

	"github.com/somebadcode/commit-tool/cmd"
	"github.com/somebadcode/commit-tool/linter"
	"github.com/somebadcode/commit-tool/revisionflag"
	"github.com/somebadcode/commit-tool/slogctx"
	"github.com/somebadcode/commit-tool/traverser"
)

const (
	FlagRev                = "rev"
	FlagOtherRev           = "other"
	FlagAllowInitialCommit = "allow-initial-commit"
)

var command = &cobra.Command{
	Use:     "lint",
	Aliases: []string{"linter"},
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

		var err error

		trav := traverser.Traverser{
			Rev:        cmd.Flags().Lookup(FlagRev).Value.(*revisionflag.Revision).Rev,
			OtherRev:   cmd.Flags().Lookup(FlagOtherRev).Value.(*revisionflag.Revision).Rev,
			ReportFunc: traverser.SlogReporter(slogctx.L(cmd.Context())),
			VisitFunc: linter.Linter{
				Rules: linter.Rules{
					linter.RuleConventionalCommit,
				},
			}.Lint,
		}

		trav.Repo, err = git.PlainOpen(repoPath)
		if err != nil {
			return err
		}

		return trav.Run(cmd.Context())
	},
}

func init() {
	flagSet := command.Flags()
	flagSet.Var(revisionflag.New(plumbing.Revision(plumbing.HEAD)), FlagRev, "revision to start at")
	flagSet.Var(revisionflag.New(""), FlagOtherRev, "revision for which the merge base of this and rev should stop the linting at")

	cmd.AddCommand(command)
}
